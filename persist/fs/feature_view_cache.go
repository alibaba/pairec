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
)

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
	featureViewCacheMu.Lock()
	defer featureViewCacheMu.Unlock()

	// track which caches are still in config
	activeNames := make(map[string]bool, len(config.FeatureViewCacheConfs))

	for name, conf := range config.FeatureViewCacheConfs {
		activeNames[name] = true

		if existing, ok := featureViewCaches[name]; ok {
			// check if config changed
			if existing.configEqual(conf) {
				continue
			}
			// config changed, stop old and recreate
			existing.stop()
		}

		c, err := newFeatureViewCache(name, conf)
		if err != nil {
			log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\terror=init failed: %v", name, err))
			continue
		}
		featureViewCaches[name] = c
		log.Info(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=initialized", name))
	}

	// remove caches no longer in config
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
	selectFields    []string
	conf            recconf.FeatureViewCacheConfig
	itemCache       sync.Map     // itemId(string) -> map[string]any
	size            atomic.Int64 // current cache size
	ch              chan string  // iterate channel for Stream type
	lastScanTime    time.Time
	refreshInterval time.Duration
	stopCh          chan struct{} // signal to stop background goroutines
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
		selectFields:    selectFields,
		conf:            conf,
		ch:              make(chan string, 1000),
		refreshInterval: time.Duration(refreshMinutes) * time.Minute,
		stopCh:          make(chan struct{}),
	}

	go cache.initData()
	if featureView.GetType() == "Stream" {
		go cache.loopIterateData()
	}
	go cache.loopRefresh()

	return cache, nil
}

func (c *FeatureViewCache) configEqual(conf recconf.FeatureViewCacheConfig) bool {
	return c.conf.FeatureStoreName == conf.FeatureStoreName &&
		c.conf.FeatureStoreViewName == conf.FeatureStoreViewName &&
		c.conf.SelectFields == conf.SelectFields &&
		c.conf.RefreshIntervalMinutes == conf.RefreshIntervalMinutes
}

func (c *FeatureViewCache) stop() {
	close(c.stopCh)
}

// initData performs the initial full scan and data load.
func (c *FeatureViewCache) initData() {
	featureView := c.fsClient.GetProject().GetFeatureView(c.viewName)
	if featureView == nil {
		log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\terror=featureView not found: %s", c.name, c.viewName))
		return
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
	c.lastScanTime = time.Now()
	if err != nil {
		log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\terror=scan failed after retries: %v", c.name, err))
		return
	}

	log.Info(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=scan completed\tids=%d", c.name, len(ids)))
	c.fetchAllData(ids)
}

// fetchAllData loads properties for all given IDs via GetOnlineFeatures and
// replaces the cache content. Items not in the new ID list are removed.
func (c *FeatureViewCache) fetchAllData(ids []string) {
	if len(ids) == 0 {
		// clear cache if scan returned empty
		var keysToDelete []string
		c.itemCache.Range(func(key, _ any) bool {
			keysToDelete = append(keysToDelete, key.(string))
			return true
		})
		for _, k := range keysToDelete {
			c.itemCache.Delete(k)
		}
		c.size.Store(0)
		log.Warning(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=scan returned empty, cache cleared", c.name))
		return
	}

	start := time.Now()
	featureView := c.fsClient.GetProject().GetFeatureView(c.viewName)
	if featureView == nil {
		log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\terror=featureView not found: %s", c.name, c.viewName))
		return
	}
	featureEntity := c.fsClient.GetProject().GetFeatureEntity(featureView.GetFeatureEntityName())
	if featureEntity == nil {
		log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\terror=featureEntity not found: %s", c.name, featureView.GetFeatureEntityName()))
		return
	}

	newIds := make(map[string]bool, len(ids))
	for _, id := range ids {
		newIds[id] = true
	}

	var cacheSize atomic.Int64
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
				log.Error(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=GetOnlineFeatures error\terror=%v", c.name, err))
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
		select {
		case <-c.stopCh:
			return
		case <-time.After(c.refreshInterval):
		}

		if time.Since(c.lastScanTime) < c.refreshInterval {
			continue
		}
		c.lastScanTime = time.Now()

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

		c.fetchAllData(ids)
		log.Info(fmt.Sprintf("module=FeatureViewCache\tname=%s\tmsg=refresh completed\tsize=%d", c.name, c.size.Load()))
	}
}
