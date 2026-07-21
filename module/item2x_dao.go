package module

import (
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/goburrow/cache"
)

// Item2XDao gets item property(x) values by item ids from an item attribute table.
type Item2XDao interface {
	// ListItem2X returns itemId => raw property value, the value is not split by delimiter
	ListItem2X(itemIds []string, context *context.RecommendContext) map[string]string
}

// NewItem2XDao creates Item2XDao by the adapter type of Item2XConfig
func NewItem2XDao(config recconf.Item2XConfig) Item2XDao {
	if config.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewItem2XHologresDao(config)
	} else if config.AdapterType == recconf.DataSource_Type_FeatureStore {
		return NewItem2XFeatureStoreDao(config)
	}

	panic("Item2XDao not implement, adapterType:" + config.AdapterType)
}

// newItem2XCache creates the local cache of itemId => property value by config,
// returns nil if CacheSize is not configured
func newItem2XCache(config recconf.Item2XConfig) cache.Cache {
	if config.CacheSize <= 0 {
		return nil
	}
	cacheTime := 3600
	if config.CacheTime > 0 {
		cacheTime = config.CacheTime
	}
	return cache.New(cache.WithMaximumSize(config.CacheSize),
		cache.WithExpireAfterWrite(time.Second*time.Duration(cacheTime)))
}

// fetchItem2XFromCache fills item2XMap with cached property values and
// returns the item ids missed in cache
func fetchItem2XFromCache(c cache.Cache, itemIds []string, item2XMap map[string]string) (missedIds []string) {
	if c == nil {
		return itemIds
	}
	for _, id := range itemIds {
		if cacheValue, ok := c.GetIfPresent(id); ok {
			if xVal, ok := cacheValue.(string); ok {
				// empty string is the negative cache value, means the item has no property value
				if xVal != "" {
					item2XMap[id] = xVal
				}
				continue
			}
		}
		missedIds = append(missedIds, id)
	}
	return
}

// putItem2XToCache puts fetched property values into cache, item ids not in
// fetchedMap are cached as empty string (negative cache)
func putItem2XToCache(c cache.Cache, itemIds []string, fetchedMap map[string]string) {
	if c == nil {
		return
	}
	for _, id := range itemIds {
		c.Put(id, fetchedMap[id])
	}
}
