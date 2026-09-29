package fs

import (
	"errors"
	"testing"
	"time"

	"github.com/alibaba/pairec/v2/recconf"
)

// withFreshCacheRegistry swaps the global registry with an empty map for the
// duration of the test and restores the original afterwards.
func withFreshCacheRegistry(t *testing.T) {
	t.Helper()
	featureViewCacheMu.Lock()
	original := featureViewCaches
	featureViewCaches = make(map[string]*FeatureViewCache)
	featureViewCacheMu.Unlock()
	t.Cleanup(func() {
		featureViewCacheMu.Lock()
		featureViewCaches = original
		featureViewCacheMu.Unlock()
	})
}

func TestFeatureViewCacheStopIsIdempotent(t *testing.T) {
	cache := &FeatureViewCache{stopCh: make(chan struct{})}

	cache.stop()
	cache.stop()

	select {
	case <-cache.stopCh:
	default:
		t.Fatal("stop channel is not closed")
	}
}

// TestFetchAllDataKeepsPreviousOnEmptyScan verifies that an empty scan result
// does not wipe a populated cache (which, combined with fail-close filtering,
// would empty the result set). The previous snapshot and ready state are kept.
func TestFetchAllDataKeepsPreviousOnEmptyScan(t *testing.T) {
	c := &FeatureViewCache{name: "test", stopCh: make(chan struct{})}
	c.itemCache.Store("item-1", &cacheEntry{props: map[string]any{"status": 1}, updatedAt: time.Now().UnixNano()})
	c.size.Store(1)
	c.ready.Store(true)

	if err := c.fetchAllData(nil, time.Now().UnixNano()); err != nil {
		t.Fatalf("fetchAllData(nil) returned error: %v", err)
	}
	if _, ok := c.Get("item-1"); !ok {
		t.Fatal("empty scan wiped previously cached item")
	}
	if c.Size() != 1 {
		t.Fatalf("expected size preserved as 1, got %d", c.Size())
	}
	if !c.Ready() {
		t.Fatal("empty scan should not clear ready state")
	}
}

// TestPruneStalePreservesStreamAddedItems verifies the delete-pass fix: entries
// written by the Stream consumer at/after the scan start must survive the prune
// even though they are absent from the scan snapshot, while genuinely stale
// entries (written before the scan, absent from the snapshot) are removed.
func TestPruneStalePreservesStreamAddedItems(t *testing.T) {
	c := &FeatureViewCache{name: "test"}
	scanStart := time.Now().UnixNano()

	// Stale: written before the scan, absent from the snapshot -> must be pruned.
	c.itemCache.Store("stale", &cacheEntry{props: map[string]any{"v": 1}, updatedAt: scanStart - int64(time.Minute)})
	// Present in the snapshot -> kept regardless of age.
	c.itemCache.Store("kept", &cacheEntry{props: map[string]any{"v": 2}, updatedAt: scanStart - int64(time.Minute)})
	// Streamed in after the scan snapshot, absent from newIds -> must be kept.
	c.itemCache.Store("streamed", &cacheEntry{props: map[string]any{"v": 3}, updatedAt: scanStart + int64(time.Second)})

	c.pruneStale(map[string]bool{"kept": true}, scanStart)

	if _, ok := c.Get("stale"); ok {
		t.Fatal("stale item absent from snapshot should be pruned")
	}
	if _, ok := c.Get("kept"); !ok {
		t.Fatal("item present in snapshot should be kept")
	}
	if _, ok := c.Get("streamed"); !ok {
		t.Fatal("stream-added item written after the snapshot must be preserved")
	}
}

func TestLoadFeatureViewCachesKeepsExistingWhenReplacementFails(t *testing.T) {
	withFreshCacheRegistry(t)

	name := "item_status_cache"
	oldCache := &FeatureViewCache{
		conf:   recconf.FeatureViewCacheConfig{FeatureStoreViewName: "old_view", RefreshIntervalMinutes: 60},
		stopCh: make(chan struct{}),
	}
	featureViewCaches[name] = oldCache

	config := &recconf.RecommendConfig{
		FeatureViewCacheConfs: map[string]recconf.FeatureViewCacheConfig{
			name: {FeatureStoreViewName: "new_view", RefreshIntervalMinutes: 60},
		},
	}
	loadFeatureViewCachesWithFactory(config, func(string, recconf.FeatureViewCacheConfig) (*FeatureViewCache, error) {
		return nil, errors.New("create failed")
	})

	if featureViewCaches[name] != oldCache {
		t.Fatal("existing cache was replaced after replacement failure")
	}
	select {
	case <-oldCache.stopCh:
		t.Fatal("existing cache was stopped after replacement failure")
	default:
	}
}

func TestLoadFeatureViewCachesReplacesThenStopsExisting(t *testing.T) {
	withFreshCacheRegistry(t)

	name := "item_status_cache"
	oldCache := &FeatureViewCache{
		conf:   recconf.FeatureViewCacheConfig{FeatureStoreViewName: "old_view"},
		stopCh: make(chan struct{}),
	}
	featureViewCaches[name] = oldCache

	newCache := &FeatureViewCache{
		conf:            recconf.FeatureViewCacheConfig{FeatureStoreViewName: "new_view"},
		stopCh:          make(chan struct{}),
		refreshInterval: time.Hour,
	}
	defer newCache.stop()

	config := &recconf.RecommendConfig{
		FeatureViewCacheConfs: map[string]recconf.FeatureViewCacheConfig{
			name: {FeatureStoreViewName: "new_view"},
		},
	}
	loadFeatureViewCachesWithFactory(config, func(string, recconf.FeatureViewCacheConfig) (*FeatureViewCache, error) {
		return newCache, nil
	})

	if featureViewCaches[name] != newCache {
		t.Fatal("new cache was not published")
	}
	select {
	case <-oldCache.stopCh:
	default:
		t.Fatal("existing cache was not stopped after replacement")
	}
}

// TestLoadFeatureViewCachesDoesNotBlockReadersDuringBuild verifies the expensive
// factory (initData) runs outside the write lock: a concurrent GetFeatureViewCache
// reader must not be blocked while a new cache is being built.
func TestLoadFeatureViewCachesDoesNotBlockReadersDuringBuild(t *testing.T) {
	withFreshCacheRegistry(t)

	name := "item_status_cache"
	existing := &FeatureViewCache{
		conf:   recconf.FeatureViewCacheConfig{FeatureStoreViewName: "old_view"},
		stopCh: make(chan struct{}),
	}
	featureViewCaches[name] = existing
	defer existing.stop()

	buildStarted := make(chan struct{})
	releaseBuild := make(chan struct{})
	config := &recconf.RecommendConfig{
		FeatureViewCacheConfs: map[string]recconf.FeatureViewCacheConfig{
			name: {FeatureStoreViewName: "new_view"},
		},
	}

	loadDone := make(chan struct{})
	go func() {
		defer close(loadDone)
		loadFeatureViewCachesWithFactory(config, func(string, recconf.FeatureViewCacheConfig) (*FeatureViewCache, error) {
			close(buildStarted)
			<-releaseBuild // simulate a slow initData (full scan)
			c := &FeatureViewCache{
				conf:            recconf.FeatureViewCacheConfig{FeatureStoreViewName: "new_view"},
				stopCh:          make(chan struct{}),
				refreshInterval: time.Hour,
			}
			return c, nil
		})
	}()

	<-buildStarted
	// While the factory is blocked, a reader must still be served promptly.
	readDone := make(chan error, 1)
	go func() {
		_, err := GetFeatureViewCache(name)
		readDone <- err
	}()
	select {
	case err := <-readDone:
		if err != nil {
			t.Fatalf("reader got error while build in progress: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("GetFeatureViewCache blocked while factory was building (write lock held during initData)")
	}

	close(releaseBuild)
	<-loadDone
}
