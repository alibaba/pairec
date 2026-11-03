package fs

import (
	"container/list"
	"fmt"
	"maps"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

var (
	featureCachesMu     sync.RWMutex
	featureCaches       = make(map[string]*FeatureCache)
	featureCachesLoadMu sync.Mutex
)

// FeatureCache is a synchronous, bounded, process-local LRU. Writes and
// conditional insertion are immediately visible, even at full capacity.
type FeatureCache struct {
	mu      sync.Mutex
	entries map[string]*list.Element
	recency list.List
	conf    recconf.FeatureCacheConfig
	client  *FSClient
	closed  bool
}

// Config returns the immutable definition belonging to this cache generation.
func (c *FeatureCache) Config() recconf.FeatureCacheConfig { return c.conf }

// Project resolves SDK metadata through the existing client, without keeping
// view or model objects across requests. Close does not close the shared client.
func (c *FeatureCache) Project() *domain.Project { return c.client.GetProject() }

func (c *FeatureCache) configMatches(conf recconf.FeatureCacheConfig) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.closed && c.conf == conf
}

func normalizeFeatureCacheConfig(conf recconf.FeatureCacheConfig) recconf.FeatureCacheConfig {
	if conf.CacheTime == 0 {
		conf.CacheTime = 3600
	}
	if conf.FeatureStoreViewName != "" && conf.SelectFields == "" {
		conf.SelectFields = "*"
	}
	return conf
}

type featureCacheEntry struct {
	key        string
	properties map[string]any
	expiresAt  time.Time
}

type featureCacheFactory func(string, recconf.FeatureCacheConfig) (*FeatureCache, error)

func newFeatureCache(name string, conf recconf.FeatureCacheConfig) (*FeatureCache, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("feature cache name must not be blank")
	}
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("feature cache %q: %w", name, err)
	}
	conf = normalizeFeatureCacheConfig(conf)
	client, err := GetFeatureStoreClient(conf.FeatureStoreName)
	if err != nil {
		return nil, err
	}
	project := client.GetProject()
	if project == nil {
		return nil, fmt.Errorf("feature store project is unavailable, name:%s", conf.FeatureStoreName)
	}
	if conf.FeatureStoreViewName != "" {
		view := project.GetFeatureView(conf.FeatureStoreViewName)
		if view == nil {
			return nil, fmt.Errorf("feature view not found(%s)", conf.FeatureStoreViewName)
		}
		if project.GetFeatureEntity(view.GetFeatureEntityName()) == nil {
			return nil, fmt.Errorf("feature entity not found(%s)", view.GetFeatureEntityName())
		}
	} else {
		model := project.GetModel(conf.FeatureStoreModelName)
		if model == nil {
			return nil, fmt.Errorf("model not found(%s)", conf.FeatureStoreModelName)
		}
		if project.GetFeatureEntity(conf.FeatureStoreEntityName) == nil {
			return nil, fmt.Errorf("feature entity not found(%s)", conf.FeatureStoreEntityName)
		}
		// The SDK accepts only entities participating in the model's views,
		// not every entity in its project. Validate before publishing a source.
		entityInModel := false
		for _, feature := range model.Features {
			view := project.GetFeatureView(feature.FeatureViewName)
			if view == nil {
				return nil, fmt.Errorf("model feature view not found(%s)", feature.FeatureViewName)
			}
			if view.GetFeatureEntityName() == conf.FeatureStoreEntityName {
				entityInModel = true
				break
			}
		}
		if !entityInModel {
			return nil, fmt.Errorf("feature entity %s does not belong to model %s", conf.FeatureStoreEntityName, conf.FeatureStoreModelName)
		}
		if conf.FeatureStore == "user" && model.GetLabelTable() == nil {
			return nil, fmt.Errorf("label table not found(%s)", conf.FeatureStoreModelName)
		}
	}
	return &FeatureCache{conf: conf, client: client, entries: make(map[string]*list.Element)}, nil
}

// LoadFeatureCaches publishes validated source definitions and policies before
// feature DAOs are built. Invalid definitions leave the registry intact.
func LoadFeatureCaches(config *recconf.RecommendConfig) {
	if err := loadFeatureCachesWithFactory(config, newFeatureCache); err != nil {
		panic(err)
	}
}

func loadFeatureCachesWithFactory(config *recconf.RecommendConfig, factory featureCacheFactory) error {
	featureCachesLoadMu.Lock()
	defer featureCachesLoadMu.Unlock()

	policies := make(map[string]recconf.FeatureCacheConfig, len(config.FeatureCacheConfs))
	for name, conf := range config.FeatureCacheConfs {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("feature cache name must not be blank")
		}
		if err := conf.Validate(); err != nil {
			return fmt.Errorf("feature cache %q: %w", name, err)
		}
		policies[name] = normalizeFeatureCacheConfig(conf)
	}

	featureCachesMu.RLock()
	previous := make(map[string]*FeatureCache, len(featureCaches))
	for name, c := range featureCaches {
		previous[name] = c
	}
	featureCachesMu.RUnlock()

	for name, conf := range policies {
		if old := previous[name]; old != nil && old.conf.FeatureStore != conf.FeatureStore {
			return fmt.Errorf("feature cache %q cannot change user/item scope; use a new cache name", name)
		}
	}

	next := make(map[string]*FeatureCache, len(policies))
	created := make([]*FeatureCache, 0, len(policies))
	for name, conf := range policies {
		if old := previous[name]; old != nil && old.configMatches(conf) {
			next[name] = old
			continue
		}
		c, err := factory(name, conf)
		if err != nil || c == nil {
			if c != nil {
				c.Close()
			}
			for _, unpublished := range created {
				unpublished.Close()
			}
			if err == nil {
				err = fmt.Errorf("factory returned a nil cache")
			}
			return fmt.Errorf("feature cache %q: %w", name, err)
		}
		next[name] = c
		created = append(created, c)
	}

	featureCachesMu.Lock()
	featureCaches = next
	featureCachesMu.Unlock()
	// Never hold the registry lock while waiting for an instance to close.
	for name, old := range previous {
		if next[name] != old {
			old.Close()
		}
	}
	return nil
}

// GetFeatureCache resolves the current instance, including after a hot reload.
func GetFeatureCache(name string) (*FeatureCache, error) {
	featureCachesMu.RLock()
	defer featureCachesMu.RUnlock()
	if c := featureCaches[name]; c != nil {
		return c, nil
	}
	return nil, fmt.Errorf("feature cache not found, name:%s", name)
}

// Get returns a copy of the top-level map; nested values must remain read-only.
// A retired instance behaves like a miss.
func (c *FeatureCache) Get(key string) (map[string]any, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, false
	}
	entry := c.lookup(key)
	if entry == nil {
		return nil, false
	}
	c.recency.MoveToFront(c.entries[key])
	return CloneFeatureProperties(entry.properties), true
}

// Put retains its own top-level map. Callers must not mutate nested values.
func (c *FeatureCache) Put(key string, value map[string]any) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.put(key, value)
	}
}

// PutIfAbsent prevents a concurrent negative result from replacing a hit.
func (c *FeatureCache) PutIfAbsent(key string, value map[string]any) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed && c.lookup(key) == nil {
		c.put(key, value)
	}
}

// lookup and put run only with the instance lock held. TTL is not extended
// by cache hits or skipped conditional writes.
func (c *FeatureCache) lookup(key string) *featureCacheEntry {
	if element := c.entries[key]; element != nil {
		entry := element.Value.(*featureCacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry
		}
		c.recency.Remove(element)
		delete(c.entries, key)
	}
	return nil
}

func (c *FeatureCache) put(key string, value map[string]any) {
	entry := &featureCacheEntry{key: key, properties: CloneFeatureProperties(value),
		expiresAt: time.Now().Add(time.Duration(c.conf.CacheTime) * time.Second)}
	if element := c.entries[key]; element != nil {
		element.Value = entry
		c.recency.MoveToFront(element)
		return
	}
	if len(c.entries) >= c.conf.CacheSize {
		oldest := c.recency.Back()
		delete(c.entries, oldest.Value.(*featureCacheEntry).key)
		c.recency.Remove(oldest)
	}
	c.entries[key] = c.recency.PushFront(entry)
}

// Close is idempotent. Late request writes are ignored, not redirected to a
// replacement instance whose data must start with a fresh policy.
func (c *FeatureCache) Close() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.closed = true
		c.entries = nil
		c.recency.Init()
	}
}

// CloneFeatureProperties copies the top-level map. Nested maps and slices are
// shared read-only; consumers may replace properties but must not mutate values
// in place. Concrete value types and typed nil values are preserved.
func CloneFeatureProperties(properties map[string]any) map[string]any {
	return maps.Clone(properties)
}
