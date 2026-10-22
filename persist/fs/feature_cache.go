package fs

import (
	"container/list"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/recconf"
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
	closed  bool
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
	if conf.CacheTime == 0 {
		conf.CacheTime = 3600
	}
	return &FeatureCache{conf: conf, entries: make(map[string]*list.Element)}, nil
}

// LoadFeatureCaches publishes named policies before feature DAOs are built.
// Invalid policies leave the current registry intact.
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
		if conf.CacheTime == 0 {
			conf.CacheTime = 3600
		}
		policies[name] = conf
	}

	featureCachesMu.RLock()
	previous := make(map[string]*FeatureCache, len(featureCaches))
	for name, c := range featureCaches {
		previous[name] = c
	}
	featureCachesMu.RUnlock()

	next := make(map[string]*FeatureCache, len(policies))
	created := make([]*FeatureCache, 0, len(policies))
	for name, conf := range policies {
		if old := previous[name]; old != nil && old.conf == conf {
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

// Get returns a request-owned copy. A retired instance behaves like a miss.
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

// Put retains a cache-owned copy, never a request's mutable feature values.
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

// CloneFeatureProperties recursively copies the maps and slices returned by
// the SDK, preserving concrete numeric types, map keys and typed nil values.
func CloneFeatureProperties(properties map[string]any) map[string]any {
	if properties == nil {
		return nil
	}
	return cloneFeatureValue(reflect.ValueOf(properties)).Interface().(map[string]any)
}

func cloneFeatureValue(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		copy := reflect.New(v.Type()).Elem()
		copy.Set(cloneFeatureValue(v.Elem()))
		return copy
	case reflect.Map:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		copy := reflect.MakeMapWithSize(v.Type(), v.Len())
		iter := v.MapRange()
		for iter.Next() {
			copy.SetMapIndex(iter.Key(), cloneFeatureValue(iter.Value()))
		}
		return copy
	case reflect.Slice:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		copy := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			copy.Index(i).Set(cloneFeatureValue(v.Index(i)))
		}
		return copy
	case reflect.Array:
		copy := reflect.New(v.Type()).Elem()
		for i := 0; i < v.Len(); i++ {
			copy.Index(i).Set(cloneFeatureValue(v.Index(i)))
		}
		return copy
	default:
		return v
	}
}
