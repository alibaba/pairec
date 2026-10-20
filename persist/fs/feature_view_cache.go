package fs

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/datasource/featuredb/fdbserverpb"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

var (
	featureViewCacheMu sync.RWMutex
	featureViewCaches  = make(map[string]*FeatureViewCache)
	// featureViewCacheLoadMu serializes the (potentially slow) load/reload flow
	// so that startup loading and a concurrent config hot-reload cannot run
	// loadFeatureViewCachesWithFactory at the same time. It is separate from
	// featureViewCacheMu, so holding it never blocks GetFeatureViewCache readers.
	featureViewCacheLoadMu sync.Mutex
)

// notReadyRetryInterval is how often loopRefresh retries while the cache has not
// yet completed its first successful full load, bounding the degraded window.
const notReadyRetryInterval = 30 * time.Second

// minRefreshIntervalMinutes is the lower bound for the periodic full re-scan
// interval. Any configured value below it (including unset/<=0) is clamped up to
// this minimum, so a full scan never runs more often than once per hour.
const minRefreshIntervalMinutes = 60

// scanRetryTimes and scanRetryInterval bound the retries of a single full scan.
const (
	scanRetryTimes    = 5
	scanRetryInterval = 10 * time.Second
)

// scanAndIterateFeatures is the SDK full scan + incremental iterate entry. It is
// a variable so tests can replace it.
var scanAndIterateFeatures = fdbserverpb.ScanAndIterateFeatures

type featureViewCacheFactory func(string, recconf.FeatureViewCacheConfig) (*FeatureViewCache, error)

// GetFeatureViewCache returns a named FeatureViewCache instance.
func GetFeatureViewCache(name string) (*FeatureViewCache, error) {
	featureViewCacheMu.RLock()
	defer featureViewCacheMu.RUnlock()
	if c, ok := featureViewCaches[name]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("feature view cache not found, name:%s", name)
}

// LoadFeatureViewCaches initializes or reloads all FeatureViewCache instances
// from config. Must be called after fs.Load().
func LoadFeatureViewCaches(config *recconf.RecommendConfig) {
	loadFeatureViewCachesWithFactory(config, newFeatureViewCache)
}

// loadFeatureViewCachesWithFactory builds new caches OUTSIDE the write lock
// (initData performs a full scan and can take minutes) and only holds the lock
// for the atomic swap, so request-path readers via GetFeatureViewCache are never
// blocked during a reload. On build failure the existing instance is kept
// running; a new instance is published only after it is fully loaded.
func loadFeatureViewCachesWithFactory(config *recconf.RecommendConfig, factory featureViewCacheFactory) {
	// Serialize the whole load flow; this lock is independent of the map lock so
	// readers are never blocked while a slow initData runs.
	featureViewCacheLoadMu.Lock()
	defer featureViewCacheLoadMu.Unlock()

	// Phase 1: snapshot what needs (re)creating under a short read lock.
	type pending struct {
		name string
		conf recconf.FeatureViewCacheConfig
	}
	featureViewCacheMu.RLock()
	toCreate := make([]pending, 0, len(config.FeatureViewCacheConfs))
	activeNames := make(map[string]bool, len(config.FeatureViewCacheConfs))
	for name, conf := range config.FeatureViewCacheConfs {
		activeNames[name] = true
		if existing, ok := featureViewCaches[name]; ok && existing.configEqual(conf) {
			continue
		}
		toCreate = append(toCreate, pending{name: name, conf: conf})
	}
	featureViewCacheMu.RUnlock()

	// Phase 2: build new instances outside any lock (expensive initData).
	created := make(map[string]*FeatureViewCache, len(toCreate))
	for _, p := range toCreate {
		c, err := factory(p.name, p.conf)
		if err != nil {
			log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\terror=init failed: %v", p.name, err))
			continue
		}
		created[p.name] = c
	}

	// Phase 3: short write lock to atomically swap in new instances and drop
	// caches no longer present in config.
	featureViewCacheMu.Lock()
	defer featureViewCacheMu.Unlock()
	for name, c := range created {
		if old, ok := featureViewCaches[name]; ok {
			old.Close()
		}
		featureViewCaches[name] = c
		c.start()
		log.Info(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=initialized", name))
	}
	for name, c := range featureViewCaches {
		if !activeNames[name] {
			c.Close()
			delete(featureViewCaches, name)
			log.Info(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=removed (no longer in config)", name))
		}
	}
}

// cacheEntry wraps cached properties with the wall-clock time (unix nano) at
// which they were written, so the periodic full-refresh prune can distinguish
// items removed from the view (written before the scan snapshot) from items
// streamed in after the snapshot, which must not be pruned.
type cacheEntry struct {
	props     map[string]any
	updatedAt int64
}

// FeatureViewCache is a local materialized view of a FeatureStore FeatureView.
// Data is loaded via ScanAndIterateFeatures on startup and refreshed periodically.
// Stream-type views also receive incremental updates from the SDK iterate
// goroutine of the latest successful scan.
type FeatureViewCache struct {
	name            string
	fsClient        *FSClient
	viewName        string
	joinId          string // feature entity join id field name, e.g. item_id
	selectFields    []string
	conf            recconf.FeatureViewCacheConfig
	itemCache       sync.Map     // itemId(string) -> *cacheEntry
	size            atomic.Int64 // current cache size
	ready           atomic.Bool  // true after the first successful full load
	lastScanTime    time.Time
	refreshInterval time.Duration
	// ctx is the parent of every scan/iterate context; Close cancels it, which
	// stops the refresh loop, pending scan retries and all SDK iterate goroutines.
	ctx    context.Context
	cancel context.CancelFunc
	// iterCancel stops the SDK iterate goroutine of the latest successful scan.
	// Only accessed by initData and then loopRefresh, which run sequentially.
	iterCancel context.CancelFunc
	stream     bool
}

// JoinId returns the feature entity join id field name of the cached view.
func (c *FeatureViewCache) JoinId() string {
	return c.joinId
}

// Ready reports whether the cache has completed its first successful full load.
// Consumers should fail open (skip filtering) while it returns false.
func (c *FeatureViewCache) Ready() bool {
	return c.ready.Load()
}

// Get returns cached properties for a single item ID.
func (c *FeatureViewCache) Get(id string) (map[string]any, bool) {
	val, ok := c.itemCache.Load(id)
	if !ok {
		return nil, false
	}
	return val.(*cacheEntry).props, true
}

// GetMulti returns cached properties for multiple item IDs.
// Missing items are omitted from the result.
func (c *FeatureViewCache) GetMulti(ids []string) map[string]map[string]any {
	result := make(map[string]map[string]any, len(ids))
	for _, id := range ids {
		if val, ok := c.itemCache.Load(id); ok {
			result[id] = val.(*cacheEntry).props
		}
	}
	return result
}

// Keys returns all cached item IDs (snapshot).
func (c *FeatureViewCache) Keys() []string {
	keys := make([]string, 0, int(c.size.Load()))
	c.itemCache.Range(func(key, _ any) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}

// Size returns current number of cached items.
func (c *FeatureViewCache) Size() int {
	return int(c.size.Load())
}

func newFeatureViewCache(name string, conf recconf.FeatureViewCacheConfig) (*FeatureViewCache, error) {
	fsclient, err := GetFeatureStoreClient(conf.FeatureStoreName)
	if err != nil {
		return nil, fmt.Errorf("get feature store client failed, name:%s, err:%v", conf.FeatureStoreName, err)
	}

	featureView := fsclient.GetProject().GetFeatureView(conf.FeatureStoreViewName)
	if featureView == nil {
		return nil, fmt.Errorf("feature view not found, name:%s", conf.FeatureStoreViewName)
	}

	joinId := ""
	if featureEntity := fsclient.GetProject().GetFeatureEntity(featureView.GetFeatureEntityName()); featureEntity != nil {
		joinId = featureEntity.FeatureEntityJoinid
	}

	refreshMinutes := conf.RefreshIntervalMinutes
	if refreshMinutes < minRefreshIntervalMinutes {
		refreshMinutes = minRefreshIntervalMinutes
	}

	selectFields := []string{"*"}
	if conf.SelectFields != "" && conf.SelectFields != "*" {
		fields := strings.Split(conf.SelectFields, ",")
		selectFields = make([]string, 0, len(fields))
		for _, f := range fields {
			selectFields = append(selectFields, strings.TrimSpace(f))
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cache := &FeatureViewCache{
		name:            name,
		fsClient:        fsclient,
		viewName:        conf.FeatureStoreViewName,
		joinId:          joinId,
		selectFields:    selectFields,
		conf:            conf,
		refreshInterval: time.Duration(refreshMinutes) * time.Minute,
		stream:          featureView.GetType() == "Stream",
		ctx:             ctx,
		cancel:          cancel,
	}

	// Config errors (client/view missing) already failed fast above. A data-plane
	// error during the initial load is logged and the instance is still published
	// in a not-ready state so that the DAO can resolve the cache by name without
	// panicking at startup.
	// loopRefresh retries at notReadyRetryInterval until the first success, and
	// consumers fail open while Ready() is false.
	if err := cache.initData(); err != nil {
		log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=initial load failed, serving degraded and retrying in background\terror=%v", name, err))
	}
	return cache, nil
}

func (c *FeatureViewCache) configEqual(conf recconf.FeatureViewCacheConfig) bool {
	return c.conf.FeatureStoreName == conf.FeatureStoreName &&
		c.conf.FeatureStoreViewName == conf.FeatureStoreViewName &&
		c.conf.SelectFields == conf.SelectFields &&
		c.conf.RefreshIntervalMinutes == conf.RefreshIntervalMinutes
}

// start launches the periodic refresh loop; it is called after the instance is
// published.
func (c *FeatureViewCache) start() {
	go c.loopRefresh()
}

// Close cancels the cache context, stopping the refresh loop, any pending scan
// retry and the SDK iterate goroutines. It is idempotent; cached data stays
// readable but is no longer updated.
func (c *FeatureViewCache) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

// initData performs the initial full scan and data load.
func (c *FeatureViewCache) initData() error {
	n, err := c.scanWithRetry()
	if err != nil {
		return err
	}
	// Only advance lastScanTime after a successful scan, so a failed init never
	// suppresses the next refresh attempt.
	c.lastScanTime = time.Now()
	// Only mark ready when the view actually has data; an empty view stays
	// not-ready so consumers fail open instead of filtering every item.
	if n > 0 {
		c.ready.Store(true)
	}
	return nil
}

// scanWithRetry runs scanAndLoad up to scanRetryTimes times and returns the
// number of scanned rows. It gives up as soon as the cache is closed.
func (c *FeatureViewCache) scanWithRetry() (int, error) {
	var err error
	for i := 0; i < scanRetryTimes; i++ {
		project := c.fsClient.GetProject()
		featureView := project.GetFeatureView(c.viewName)
		if featureView == nil {
			return 0, fmt.Errorf("featureView not found: %s", c.viewName)
		}
		var n int
		if n, err = c.scanAndLoad(project, featureView); err == nil {
			return n, nil
		}
		if ctxErr := c.ctx.Err(); ctxErr != nil {
			return 0, ctxErr
		}
		log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=scan retry %d\terror=%v", c.name, i+1, err))
		if i == scanRetryTimes-1 {
			break
		}
		timer := time.NewTimer(scanRetryInterval)
		select {
		case <-c.ctx.Done():
			timer.Stop()
			return 0, c.ctx.Err()
		case <-timer.C:
		}
	}
	return 0, fmt.Errorf("scan failed after retries: %w", err)
}

// scanAndLoad runs one ScanAndIterateFeatures full scan, writes every scanned
// row into the cache and prunes items absent from the scan. The SDK keeps an
// iterate goroutine running after the scan: for Stream views it applies
// incremental updates and replaces the previous scan's iterate goroutine, for
// Batch views it is stopped right away. It returns the number of scanned rows.
func (c *FeatureViewCache) scanAndLoad(project *domain.Project, featureView domain.FeatureView) (int, error) {
	start := time.Now()
	// Entries written at/after scanStart (by this scan or by a still running
	// iterate goroutine) are newer than the snapshot and must survive the prune.
	scanStart := start.UnixNano()
	var (
		mu       sync.Mutex
		scanDone bool
		newIds   = make(map[string]bool)
	)
	scanCtx, scanCancel := context.WithCancel(c.ctx)
	primaryKey := primaryKeyName(featureView)
	err := scanAndIterateFeatures(scanCtx, project, featureView, func(id string, properties map[string]interface{}) {
		c.store(id, c.normalizeJoinId(id, primaryKey, properties))
		// The same callback is invoked later by the iterate goroutine; only rows
		// delivered by the synchronous full scan belong to the snapshot.
		mu.Lock()
		if !scanDone {
			newIds[id] = true
		}
		mu.Unlock()
	})
	mu.Lock()
	scanDone = true
	mu.Unlock()
	if err != nil {
		scanCancel()
		return 0, err
	}
	if c.stream {
		if c.iterCancel != nil {
			c.iterCancel()
		}
		c.iterCancel = scanCancel
	} else {
		scanCancel()
	}

	if len(newIds) == 0 {
		// Never wipe a populated cache on an empty scan: it is more likely a
		// transient FeatureDB glitch or an unproduced view than a legitimate
		// "all items removed" signal. Keep the previous snapshot; ready is left
		// unchanged so consumers keep serving last-known-good data (or fail open
		// if this was the initial, still-empty load).
		log.Warning(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=scan returned empty ids, keep previous cache\tsize=%d", c.name, c.size.Load()))
		return 0, nil
	}

	c.pruneStale(newIds, scanStart)
	log.Info(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=scan completed\tids=%d\tsize=%d\tcost=%d",
		c.name, len(newIds), c.size.Load(), utils.CostTime(start)))
	return len(newIds), nil
}

// primaryKeyName returns the primary key field name of the feature view, or ""
// when it cannot be determined.
func primaryKeyName(featureView domain.FeatureView) string {
	if featureView == nil {
		return ""
	}
	for _, field := range featureView.GetFields() {
		if field.IsPrimaryKey {
			return field.Name
		}
	}
	return ""
}

// normalizeJoinId stores the primary key value under the entity join id and
// drops the primary key field when its name differs, matching the properties
// returned by GetOnlineFeatures.
func (c *FeatureViewCache) normalizeJoinId(id, primaryKey string, properties map[string]any) map[string]any {
	if c.joinId == "" {
		return properties
	}
	properties[c.joinId] = id
	if primaryKey != "" && primaryKey != c.joinId {
		delete(properties, primaryKey)
	}
	return properties
}

// store writes one row into the cache. It is called by both the full scan and
// the Stream incremental iterate.
func (c *FeatureViewCache) store(id string, properties map[string]any) {
	entry := &cacheEntry{props: c.selectProperties(properties), updatedAt: time.Now().UnixNano()}
	if _, loaded := c.itemCache.Swap(id, entry); !loaded {
		c.size.Add(1)
	}
}

// selectProperties keeps only the configured select fields plus the join id,
// or all fields when SelectFields is "*".
func (c *FeatureViewCache) selectProperties(properties map[string]any) map[string]any {
	if len(c.selectFields) == 0 || (len(c.selectFields) == 1 && c.selectFields[0] == "*") {
		return properties
	}
	selected := make(map[string]any, len(c.selectFields)+1)
	if v, ok := properties[c.joinId]; ok {
		selected[c.joinId] = v
	}
	for _, f := range c.selectFields {
		if v, ok := properties[f]; ok {
			selected[f] = v
		}
	}
	return selected
}

// pruneStale deletes cache entries that are absent from the latest scan snapshot
// (newIds) AND were last written before the scan started. Entries written by the
// Stream iterate consumer at/after scanStartNano are updates newer than the
// snapshot, so they are preserved; pruning them would silently drop freshly
// streamed items until the next full refresh.
func (c *FeatureViewCache) pruneStale(newIds map[string]bool, scanStartNano int64) {
	c.itemCache.Range(func(key, val any) bool {
		if newIds[key.(string)] {
			return true
		}
		if entry, ok := val.(*cacheEntry); ok && entry.updatedAt >= scanStartNano {
			return true
		}
		// CompareAndDelete keeps an entry concurrently replaced by the iterate
		// goroutine after the check above.
		if c.itemCache.CompareAndDelete(key, val) {
			c.size.Add(-1)
		}
		return true
	})
}

// loopRefresh periodically re-scans the FeatureView and replaces cache content.
func (c *FeatureViewCache) loopRefresh() {
	for {
		// Retry quickly until the first successful load, then refresh on schedule.
		interval := c.refreshInterval
		if !c.ready.Load() {
			interval = notReadyRetryInterval
		}
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(interval):
		}

		// Once ready, skip if a scan already happened within the refresh interval.
		if c.ready.Load() && time.Since(c.lastScanTime) < c.refreshInterval {
			continue
		}

		n, err := c.scanWithRetry()
		if err != nil {
			if c.ctx.Err() != nil {
				return
			}
			log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\terror=refresh scan failed: %v", c.name, err))
			continue
		}
		// Advance lastScanTime and mark ready only after a successful, non-empty
		// refresh, so a failed or empty attempt is retried on the next tick.
		c.lastScanTime = time.Now()
		if n > 0 {
			c.ready.Store(true)
		}
		log.Info(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=refresh completed\tsize=%d", c.name, c.size.Load()))
	}
}
