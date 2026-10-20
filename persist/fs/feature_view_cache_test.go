package fs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

// fakeFeatureView only implements GetFields; other methods panic if called.
type fakeFeatureView struct {
	domain.FeatureView
	fields []api.FeatureViewFields
}

func (f *fakeFeatureView) GetFields() []api.FeatureViewFields { return f.fields }

// newTestCache builds a cache with a live context, as newFeatureViewCache does.
func newTestCache(conf recconf.FeatureViewCacheConfig) *FeatureViewCache {
	ctx, cancel := context.WithCancel(context.Background())
	return &FeatureViewCache{
		name:            "test",
		conf:            conf,
		selectFields:    []string{"*"},
		refreshInterval: time.Hour,
		ctx:             ctx,
		cancel:          cancel,
	}
}

func isClosed(c *FeatureViewCache) bool {
	return c.ctx.Err() != nil
}

type scanFn = func(ctx context.Context, project *domain.Project, featureView domain.FeatureView, fn func(id string, properties map[string]interface{})) error

// withFakeScan replaces scanAndIterateFeatures for the duration of the test.
func withFakeScan(t *testing.T, fake scanFn) {
	t.Helper()
	original := scanAndIterateFeatures
	scanAndIterateFeatures = fake
	t.Cleanup(func() { scanAndIterateFeatures = original })
}

// fakeScanRows returns a fake scan that delivers rows synchronously and records
// the scan context it was given.
func fakeScanRows(rows map[string]map[string]any, ctxs *[]context.Context) scanFn {
	return func(ctx context.Context, _ *domain.Project, _ domain.FeatureView, fn func(string, map[string]interface{})) error {
		*ctxs = append(*ctxs, ctx)
		for id, props := range rows {
			fn(id, props)
		}
		return nil
	}
}

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

func TestFeatureViewCacheCloseIsIdempotent(t *testing.T) {
	cache := newTestCache(recconf.FeatureViewCacheConfig{})

	cache.Close()
	cache.Close()

	if !isClosed(cache) {
		t.Fatal("cache context is not canceled")
	}

	// A zero-value cache must not panic on Close.
	(&FeatureViewCache{}).Close()
}

// TestScanAndLoadEmptyScanIsValidSnapshot verifies that a successful empty scan
// is applied like any other snapshot: entries written before the scan are
// pruned and the size drops, while entries streamed in after the scan started
// survive.
func TestScanAndLoadEmptyScanIsValidSnapshot(t *testing.T) {
	c := newTestCache(recconf.FeatureViewCacheConfig{})
	c.stream = true
	defer c.Close()
	c.itemCache.Store("item-1", &cacheEntry{props: map[string]any{"status": 1}, updatedAt: time.Now().Add(-time.Minute).UnixNano()})
	c.size.Store(1)
	c.ready.Store(true)

	withFakeScan(t, func(context.Context, *domain.Project, domain.FeatureView, func(string, map[string]interface{})) error {
		// Simulate the previous iterate goroutine writing during the scan.
		c.store("streamed", map[string]any{"status": 2})
		return nil
	})

	if err := c.scanAndLoad(nil, nil); err != nil {
		t.Fatalf("scanAndLoad returned error: %v", err)
	}
	if _, ok := c.Get("item-1"); ok {
		t.Fatal("empty scan should prune items written before the scan")
	}
	if _, ok := c.Get("streamed"); !ok {
		t.Fatal("item streamed in after the scan started must be preserved")
	}
	if c.Size() != 1 {
		t.Fatalf("expected size 1, got %d", c.Size())
	}
}

// TestInitDataEmptyViewIsReady verifies that an initial empty scan marks the
// cache ready, so consumers filter with cache-miss semantics instead of failing
// open forever and loopRefresh does not keep rescanning every 30s.
func TestInitDataEmptyViewIsReady(t *testing.T) {
	var ctxs []context.Context
	withFakeScan(t, fakeScanRows(nil, &ctxs))

	project := &domain.Project{}
	project.FeatureViewMap.Store("view", &fakeFeatureView{})
	c := newTestCache(recconf.FeatureViewCacheConfig{})
	c.fsClient = &FSClient{project: project}
	c.viewName = "view"
	defer c.Close()

	if err := c.initData(); err != nil {
		t.Fatalf("initData returned error: %v", err)
	}
	if !c.Ready() {
		t.Fatal("successful empty initial scan should mark the cache ready")
	}
	if c.lastScanTime.IsZero() {
		t.Fatal("successful scan should advance lastScanTime")
	}
}

// TestInitDataFailureStaysNotReady verifies a failed initial scan leaves the
// cache not ready, and that Close interrupts the scan retries.
func TestInitDataFailureStaysNotReady(t *testing.T) {
	project := &domain.Project{}
	project.FeatureViewMap.Store("view", &fakeFeatureView{})
	c := newTestCache(recconf.FeatureViewCacheConfig{})
	c.fsClient = &FSClient{project: project}
	c.viewName = "view"

	attempts := 0
	withFakeScan(t, func(context.Context, *domain.Project, domain.FeatureView, func(string, map[string]interface{})) error {
		attempts++
		c.Close()
		return errors.New("scan failed")
	})

	if err := c.initData(); err == nil {
		t.Fatal("expected initData error")
	}
	if attempts != 1 {
		t.Fatalf("Close should stop retries, got %d attempts", attempts)
	}
	if c.Ready() {
		t.Fatal("failed initial scan must not mark the cache ready")
	}
}

// TestScanAndLoadReplacesContentAndStopsBatchIterate verifies a full scan loads
// all rows, prunes items absent from the snapshot, keeps the size in sync, and
// cancels the SDK iterate goroutine right away for Batch views.
func TestScanAndLoadReplacesContentAndStopsBatchIterate(t *testing.T) {
	var ctxs []context.Context
	withFakeScan(t, fakeScanRows(map[string]map[string]any{
		"a": {"item_id": "a", "status": 1},
		"b": {"item_id": "b", "status": 2},
	}, &ctxs))

	c := newTestCache(recconf.FeatureViewCacheConfig{})
	defer c.Close()
	c.itemCache.Store("stale", &cacheEntry{props: map[string]any{}, updatedAt: time.Now().Add(-time.Minute).UnixNano()})
	c.size.Store(1)

	if err := c.scanAndLoad(nil, nil); err != nil {
		t.Fatalf("scanAndLoad returned error: %v", err)
	}
	if props, ok := c.Get("b"); !ok || props["status"] != 2 {
		t.Fatalf("scanned item not cached correctly: %v", props)
	}
	if _, ok := c.Get("stale"); ok {
		t.Fatal("stale item absent from snapshot should be pruned")
	}
	if c.Size() != 2 {
		t.Fatalf("expected size 2, got %d", c.Size())
	}
	if len(ctxs) != 1 || ctxs[0].Err() == nil {
		t.Fatal("batch view iterate context should be canceled after the scan")
	}
}

// TestScanAndLoadStreamIterateLifecycle verifies that for Stream views the
// iterate context of the latest scan stays alive, a new scan cancels the
// previous one, and Close cancels the current one.
func TestScanAndLoadStreamIterateLifecycle(t *testing.T) {
	var ctxs []context.Context
	withFakeScan(t, fakeScanRows(map[string]map[string]any{"a": {"item_id": "a"}}, &ctxs))

	c := newTestCache(recconf.FeatureViewCacheConfig{})
	c.stream = true

	if err := c.scanAndLoad(nil, nil); err != nil {
		t.Fatalf("first scan returned error: %v", err)
	}
	if ctxs[0].Err() != nil {
		t.Fatal("stream iterate context should stay alive after the scan")
	}
	if err := c.scanAndLoad(nil, nil); err != nil {
		t.Fatalf("second scan returned error: %v", err)
	}
	if ctxs[0].Err() == nil {
		t.Fatal("previous stream iterate context should be canceled by a new scan")
	}
	if ctxs[1].Err() != nil {
		t.Fatal("latest stream iterate context should stay alive")
	}

	c.Close()
	if ctxs[1].Err() == nil {
		t.Fatal("Close should cancel the latest stream iterate context")
	}
}

// TestScanAndLoadAppliesIncrementalUpdates verifies that rows delivered by the
// iterate goroutine after the scan returns are written to the cache.
func TestScanAndLoadAppliesIncrementalUpdates(t *testing.T) {
	incremental := make(chan struct{})
	delivered := make(chan struct{})
	withFakeScan(t, func(ctx context.Context, _ *domain.Project, _ domain.FeatureView, fn func(string, map[string]interface{})) error {
		fn("a", map[string]any{"item_id": "a", "v": 1})
		go func() {
			defer close(delivered)
			<-incremental
			if ctx.Err() != nil {
				return
			}
			fn("a", map[string]any{"item_id": "a", "v": 2})
			fn("b", map[string]any{"item_id": "b", "v": 3})
		}()
		return nil
	})

	c := newTestCache(recconf.FeatureViewCacheConfig{})
	c.stream = true
	defer c.Close()

	if err := c.scanAndLoad(nil, nil); err != nil {
		t.Fatalf("scanAndLoad returned error: %v", err)
	}
	close(incremental)
	<-delivered

	if props, _ := c.Get("a"); props["v"] != 2 {
		t.Fatalf("incremental update not applied, got %v", props)
	}
	if _, ok := c.Get("b"); !ok {
		t.Fatal("incrementally added item missing")
	}
	if c.Size() != 2 {
		t.Fatalf("expected size 2, got %d", c.Size())
	}
}

// TestScanAndLoadErrorCancelsScanContext verifies a failed scan returns the
// error, leaves the cache untouched and releases its context.
func TestScanAndLoadErrorCancelsScanContext(t *testing.T) {
	var scanCtx context.Context
	withFakeScan(t, func(ctx context.Context, _ *domain.Project, _ domain.FeatureView, _ func(string, map[string]interface{})) error {
		scanCtx = ctx
		return errors.New("scan failed")
	})

	c := newTestCache(recconf.FeatureViewCacheConfig{})
	c.stream = true
	defer c.Close()

	if err := c.scanAndLoad(nil, nil); err == nil {
		t.Fatal("expected scan error")
	}
	if scanCtx.Err() == nil {
		t.Fatal("failed scan context should be canceled")
	}
	if c.iterCancel != nil {
		t.Fatal("failed scan must not replace the stream iterate cancel")
	}
}

func TestSelectProperties(t *testing.T) {
	props := map[string]any{"item_id": "a", "status": 1, "other": 2}

	c := &FeatureViewCache{joinId: "item_id", selectFields: []string{"*"}}
	if got := c.selectProperties(props); len(got) != 3 {
		t.Fatalf("select * should keep all fields, got %v", got)
	}

	c.selectFields = []string{"status", "missing"}
	got := c.selectProperties(props)
	if len(got) != 2 || got["item_id"] != "a" || got["status"] != 1 {
		t.Fatalf("expected join id and selected fields only, got %v", got)
	}
}

// TestScanAndLoadRenamesPrimaryKeyToJoinId verifies that, as GetOnlineFeatures
// does, a primary key named differently from the entity join id is stored under
// the join id, for both select-all and SelectFields projections, and for rows
// from the full scan as well as the incremental iterate.
func TestScanAndLoadRenamesPrimaryKeyToJoinId(t *testing.T) {
	featureView := &fakeFeatureView{fields: []api.FeatureViewFields{
		{Name: "pk", IsPrimaryKey: true},
		{Name: "status"},
		{Name: "other"},
	}}

	for _, tc := range []struct {
		name         string
		selectFields []string
		wantLen      int
	}{
		{name: "select all", selectFields: []string{"*"}, wantLen: 3},
		{name: "select fields", selectFields: []string{"status"}, wantLen: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			delivered := make(chan struct{})
			withFakeScan(t, func(ctx context.Context, _ *domain.Project, _ domain.FeatureView, fn func(string, map[string]interface{})) error {
				fn("a", map[string]any{"pk": "a", "status": 1, "other": 2})
				go func() {
					defer close(delivered)
					fn("b", map[string]any{"pk": "b", "status": 3, "other": 4})
				}()
				return nil
			})

			c := newTestCache(recconf.FeatureViewCacheConfig{})
			c.joinId = "item_id"
			c.selectFields = tc.selectFields
			c.stream = true
			defer c.Close()

			if err := c.scanAndLoad(nil, featureView); err != nil {
				t.Fatalf("scanAndLoad returned error: %v", err)
			}
			<-delivered

			for _, id := range []string{"a", "b"} {
				props, ok := c.Get(id)
				if !ok {
					t.Fatalf("item %s missing", id)
				}
				if props["item_id"] != id {
					t.Fatalf("item %s: join id not set, got %v", id, props)
				}
				if _, ok := props["pk"]; ok {
					t.Fatalf("item %s: primary key field should be renamed, got %v", id, props)
				}
				if len(props) != tc.wantLen {
					t.Fatalf("item %s: expected %d fields, got %v", id, tc.wantLen, props)
				}
			}
		})
	}
}

func TestNormalizeJoinId(t *testing.T) {
	// Primary key already named as the join id: kept as is.
	c := &FeatureViewCache{joinId: "item_id"}
	got := c.normalizeJoinId("a", "item_id", map[string]any{"item_id": "a", "status": 1})
	if len(got) != 2 || got["item_id"] != "a" {
		t.Fatalf("same-name primary key should be kept, got %v", got)
	}

	// Unknown primary key name: join id is added, nothing is deleted.
	got = c.normalizeJoinId("a", "", map[string]any{"pk": "a"})
	if got["item_id"] != "a" || got["pk"] != "a" {
		t.Fatalf("unknown primary key should only add join id, got %v", got)
	}

	// No join id (entity not found): properties unchanged.
	c = &FeatureViewCache{}
	got = c.normalizeJoinId("a", "pk", map[string]any{"pk": "a"})
	if len(got) != 1 || got["pk"] != "a" {
		t.Fatalf("empty join id should leave properties unchanged, got %v", got)
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
	oldCache := newTestCache(recconf.FeatureViewCacheConfig{FeatureStoreViewName: "old_view", RefreshIntervalMinutes: 60})
	featureViewCaches[name] = oldCache
	defer oldCache.Close()

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
	if isClosed(oldCache) {
		t.Fatal("existing cache was stopped after replacement failure")
	}
}

func TestLoadFeatureViewCachesReplacesThenStopsExisting(t *testing.T) {
	withFreshCacheRegistry(t)

	name := "item_status_cache"
	oldCache := newTestCache(recconf.FeatureViewCacheConfig{FeatureStoreViewName: "old_view"})
	featureViewCaches[name] = oldCache

	newCache := newTestCache(recconf.FeatureViewCacheConfig{FeatureStoreViewName: "new_view"})
	defer newCache.Close()

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
	if !isClosed(oldCache) {
		t.Fatal("existing cache was not stopped after replacement")
	}
}

// TestLoadFeatureViewCachesDoesNotBlockReadersDuringBuild verifies the expensive
// factory (initData) runs outside the write lock: a concurrent GetFeatureViewCache
// reader must not be blocked while a new cache is being built.
func TestLoadFeatureViewCachesDoesNotBlockReadersDuringBuild(t *testing.T) {
	withFreshCacheRegistry(t)

	name := "item_status_cache"
	existing := newTestCache(recconf.FeatureViewCacheConfig{FeatureStoreViewName: "old_view"})
	featureViewCaches[name] = existing
	defer existing.Close()

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
			return newTestCache(recconf.FeatureViewCacheConfig{FeatureStoreViewName: "new_view"}), nil
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
	if c, err := GetFeatureViewCache(name); err == nil {
		c.Close()
	}
}
