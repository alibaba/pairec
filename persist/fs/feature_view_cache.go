package fs

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
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
			old.stop()
		}
		featureViewCaches[name] = c
		c.start()
		log.Info(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=initialized", name))
	}
	for name, c := range featureViewCaches {
		if !activeNames[name] {
			c.stop()
			delete(featureViewCaches, name)
			log.Info(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=removed (no longer in config)", name))
		}
	}
}

// FeatureViewCache is a local materialized view of a FeatureStore FeatureView.
// Data is loaded via scan on startup and refreshed periodically.
// Stream-type views also receive incremental updates via iterate.
type FeatureViewCache struct {
	name            string
	fsClient        *FSClient
	viewName        string
	joinId          string // feature entity join id field name, e.g. item_id
	selectFields    []string
	conf            recconf.FeatureViewCacheConfig
	itemCache       sync.Map     // itemId(string) -> map[string]any
	size            atomic.Int64 // current cache size
	ready           atomic.Bool  // true after the first successful full load
	ch              chan string  // iterate channel for Stream type
	lastScanTime    time.Time
	refreshInterval time.Duration
	stopCh          chan struct{} // signal to stop background goroutines
	stopOnce        sync.Once
	stream          bool
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
	return val.(map[string]any), true
}

// GetMulti returns cached properties for multiple item IDs.
// Missing items are omitted from the result.
func (c *FeatureViewCache) GetMulti(ids []string) map[string]map[string]any {
	result := make(map[string]map[string]any, len(ids))
	for _, id := range ids {
		if val, ok := c.itemCache.Load(id); ok {
			result[id] = val.(map[string]any)
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
	if refreshMinutes <= 0 {
		refreshMinutes = 60
	}

	selectFields := []string{"*"}
	if conf.SelectFields != "" && conf.SelectFields != "*" {
		fields := strings.Split(conf.SelectFields, ",")
		selectFields = make([]string, 0, len(fields))
		for _, f := range fields {
			selectFields = append(selectFields, strings.TrimSpace(f))
		}
	}

	cache := &FeatureViewCache{
		name:            name,
		fsClient:        fsclient,
		viewName:        conf.FeatureStoreViewName,
		joinId:          joinId,
		selectFields:    selectFields,
		conf:            conf,
		ch:              make(chan string, 1000),
		refreshInterval: time.Duration(refreshMinutes) * time.Minute,
		stream:          featureView.GetType() == "Stream",
		stopCh:          make(chan struct{}),
	}

	// For Stream views attach the iterate consumer BEFORE the initial scan:
	// ScanAndIterateData starts an SDK producer goroutine that blocks once c.ch's
	// buffer fills, so a consumer must already be draining it during the
	// (possibly minutes-long) initial full load. Every error return above happens
	// before the struct is built, so this cannot orphan a goroutine.
	if cache.stream {
		go cache.loopIterateData()
	}

	// Config errors (client/view missing) already failed fast above. A data-plane
	// error during the initial load is logged and the instance is still published
	// in a not-ready state so that:
	//   - Stream views keep a consumer for the SDK iterate goroutine's channel,
	//     avoiding a goroutine and FeatureDB snapshot leak;
	//   - the DAO can resolve the cache by name without panicking at startup.
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
// published. The Stream iterate consumer is started earlier in
// newFeatureViewCache so that it drains c.ch during the initial scan.
func (c *FeatureViewCache) start() {
	go c.loopRefresh()
}

func (c *FeatureViewCache) stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
}

// initData performs the initial full scan and data load.
func (c *FeatureViewCache) initData() error {
	featureView := c.fsClient.GetProject().GetFeatureView(c.viewName)
	if featureView == nil {
		return fmt.Errorf("featureView not found: %s", c.viewName)
	}

	var (
		ids []string
		err error
	)
	for i := 0; i < 5; i++ {
		if featureView.GetType() == "Batch" {
			ids, err = featureView.ScanAndIterateData("", nil)
		} else {
			ids, err = featureView.ScanAndIterateData("", c.ch)
		}
		if err == nil {
			break
		}
		log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=scan retry %d\terror=%v", c.name, i+1, err))
		time.Sleep(10 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("scan failed after retries: %w", err)
	}
	// Only advance lastScanTime after a successful scan, so a failed init never
	// suppresses the next refresh attempt.
	c.lastScanTime = time.Now()

	log.Info(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=scan completed\tids=%d", c.name, len(ids)))
	if err := c.fetchAllData(ids); err != nil {
		return err
	}
	// Only mark ready when the view actually has data; an empty view stays
	// not-ready so consumers fail open instead of filtering every item.
	if len(ids) > 0 {
		c.ready.Store(true)
	}
	return nil
}

// fetchAllData loads properties for all given IDs via GetOnlineFeatures and
// replaces the cache content. Items not in the new ID list are removed.
func (c *FeatureViewCache) fetchAllData(ids []string) error {
	if len(ids) == 0 {
		// Never wipe a populated cache on an empty scan: it is more likely a
		// transient FeatureDB glitch or an unproduced view than a legitimate
		// "all items removed" signal. Keep the previous snapshot; ready is left
		// unchanged so consumers keep serving last-known-good data (or fail open
		// if this was the initial, still-empty load).
		log.Warning(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=scan returned empty ids, keep previous cache\tsize=%d", c.name, c.size.Load()))
		return nil
	}

	start := time.Now()
	featureView := c.fsClient.GetProject().GetFeatureView(c.viewName)
	if featureView == nil {
		return fmt.Errorf("featureView not found: %s", c.viewName)
	}
	featureEntity := c.fsClient.GetProject().GetFeatureEntity(featureView.GetFeatureEntityName())
	if featureEntity == nil {
		return fmt.Errorf("featureEntity not found: %s", featureView.GetFeatureEntityName())
	}

	newIds := make(map[string]bool, len(ids))
	for _, id := range ids {
		newIds[id] = true
	}

	var cacheSize atomic.Int64
	var fetchErr error
	var fetchErrOnce sync.Once
	var wg sync.WaitGroup
	concurrencyCh := make(chan int, 5)
	batchSize := 200

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}

		wg.Add(1)
		concurrencyCh <- 1
		go func(batchIds []string) {
			defer wg.Done()
			defer func() { <-concurrencyCh }()

			joinIds := make([]any, 0, len(batchIds))
			for _, id := range batchIds {
				joinIds = append(joinIds, id)
			}
			features, err := featureView.GetOnlineFeatures(joinIds, c.selectFields, map[string]string{})
			if err != nil {
				fetchErrOnce.Do(func() {
					fetchErr = err
				})
				return
			}
			for _, featureMap := range features {
				itemId := utils.ToString(featureMap[featureEntity.FeatureEntityJoinid], "")
				if itemId != "" {
					c.itemCache.Store(itemId, featureMap)
					cacheSize.Add(1)
				}
			}
		}(ids[i:end])
	}
	wg.Wait()
	if fetchErr != nil {
		return fmt.Errorf("GetOnlineFeatures failed: %w", fetchErr)
	}

	// remove items not in the new scan result (handle deletions)
	c.itemCache.Range(func(key, _ any) bool {
		if !newIds[key.(string)] {
			c.itemCache.Delete(key)
		}
		return true
	})

	c.size.Store(cacheSize.Load())
	log.Info(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=fetchAllData completed\tsize=%d\tcost=%d",
		c.name, cacheSize.Load(), utils.CostTime(start)))
	return nil
}

// fetchBatchData loads properties for a batch of IDs and adds them to the cache
// (used for incremental updates from Stream iterate).
func (c *FeatureViewCache) fetchBatchData(ids []string) {
	if len(ids) == 0 {
		return
	}

	featureView := c.fsClient.GetProject().GetFeatureView(c.viewName)
	if featureView == nil {
		return
	}
	featureEntity := c.fsClient.GetProject().GetFeatureEntity(featureView.GetFeatureEntityName())
	if featureEntity == nil {
		return
	}

	joinIds := make([]any, 0, len(ids))
	for _, id := range ids {
		joinIds = append(joinIds, id)
	}
	features, err := featureView.GetOnlineFeatures(joinIds, c.selectFields, map[string]string{})
	if err != nil {
		log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=fetchBatchData error\terror=%v", c.name, err))
		return
	}
	for _, featureMap := range features {
		itemId := utils.ToString(featureMap[featureEntity.FeatureEntityJoinid], "")
		if itemId != "" {
			if _, loaded := c.itemCache.LoadOrStore(itemId, featureMap); loaded {
				c.itemCache.Store(itemId, featureMap)
			} else {
				c.size.Add(1)
			}
		}
	}
}

// loopIterateData consumes incremental IDs from the Stream iterate channel
// and immediately loads their latest properties into the cache.
func (c *FeatureViewCache) loopIterateData() {
	for {
		select {
		case <-c.stopCh:
			return
		case id, ok := <-c.ch:
			if !ok {
				return
			}

			ids := []string{id}
		drain:
			for {
				select {
				case <-c.stopCh:
					return
				case id, ok := <-c.ch:
					if !ok {
						return
					}
					ids = append(ids, id)
				default:
					break drain
				}
			}
			c.fetchBatchData(ids)
		}
	}
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
		case <-c.stopCh:
			return
		case <-time.After(interval):
		}

		// Once ready, skip if a scan already happened within the refresh interval.
		if c.ready.Load() && time.Since(c.lastScanTime) < c.refreshInterval {
			continue
		}

		featureView := c.fsClient.GetProject().GetFeatureView(c.viewName)
		if featureView == nil {
			log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\terror=featureView not found: %s", c.name, c.viewName))
			continue
		}

		var (
			ids []string
			err error
		)
		for i := 0; i < 5; i++ {
			ids, err = featureView.ScanAndIterateData("", nil)
			if err == nil {
				break
			}
			time.Sleep(10 * time.Second)
		}
		if err != nil {
			log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\terror=refresh scan failed: %v", c.name, err))
			continue
		}

		if err := c.fetchAllData(ids); err != nil {
			log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\terror=refresh data failed: %v", c.name, err))
			continue
		}
		// Advance lastScanTime and mark ready only after a successful, non-empty
		// refresh, so a failed or empty attempt is retried on the next tick.
		c.lastScanTime = time.Now()
		if len(ids) > 0 {
			c.ready.Store(true)
		}
		log.Info(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=refresh completed\tsize=%d", c.name, c.size.Load()))
	}
}
