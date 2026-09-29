package module

import (
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// featureViewCacheReader is the read-side of fs.FeatureViewCache consumed by the
// DAO. Declared as an interface so the DAO can be tested with a fake cache.
type featureViewCacheReader interface {
	Get(id string) (map[string]any, bool)
	JoinId() string
	Ready() bool
}

// ItemStateFilterFeatureViewCacheDao is an ItemStateFilterDao implementation
// that reads item properties from a pre-populated FeatureViewCache (full-scan
// local cache) instead of making per-request GetOnlineFeatures calls.
type ItemStateFilterFeatureViewCacheDao struct {
	cacheName           string
	cache               featureViewCacheReader
	filterParam         *FilterParam
	defaultFieldValues  map[string]any
	generateUserProgram *vm.Program
	transFunc           FeatureTransFunc
}

func NewItemStateFilterFeatureViewCacheDao(config recconf.FilterConfig, transFunc FeatureTransFunc) *ItemStateFilterFeatureViewCacheDao {
	cache, err := fs.GetFeatureViewCache(config.ItemStateDaoConf.FeatureViewCacheName)
	if err != nil {
		panic(fmt.Sprintf("module=ItemStateFilterFeatureViewCacheDao\terror=%v", err))
	}

	dao := &ItemStateFilterFeatureViewCacheDao{
		cacheName:          config.ItemStateDaoConf.FeatureViewCacheName,
		cache:              cache,
		defaultFieldValues: config.ItemStateDaoConf.DefaultFieldValues,
		transFunc:          transFunc,
	}
	if len(config.FilterParams) > 0 {
		dao.filterParam = NewFilterParamWithConfig(config.FilterParams)
	}
	if config.GenerateUserDataExpr != "" {
		if p, err := expr.Compile(config.GenerateUserDataExpr, expr.AllowUndefinedVariables()); err != nil {
			panic(err)
		} else {
			dao.generateUserProgram = p
		}
	}
	return dao
}

func (d *ItemStateFilterFeatureViewCacheDao) Filter(user *User, items []*Item, ctx *context.RecommendContext) (ret []*Item) {
	cache := d.cache
	if d.cacheName != "" {
		latestCache, err := fs.GetFeatureViewCache(d.cacheName)
		if err != nil {
			// Fail open: the cache is unavailable (e.g. removed from config), so skip
			// filtering rather than emptying the result set. The degraded state is
			// already logged at the cache layer, so only trace it per request here.
			ctx.LogDebug(fmt.Sprintf("module=ItemStateFilterFeatureViewCacheDao\tmsg=cache unavailable, fail-open (skip filtering)\terror=%v", err))
			return items
		}
		cache = latestCache
	}
	// Fail open while the cache has not completed its first full load (startup
	// degraded / FeatureDB outage); loopRefresh retries in the background.
	if !cache.Ready() {
		ctx.LogDebug("module=ItemStateFilterFeatureViewCacheDao\tmsg=cache not ready, fail-open (skip filtering)")
		return items
	}
	joinId := cache.JoinId()

	userFeatures := user.MakeUserFeatures2()
	var itemIdGenMap map[string]string
	if d.generateUserProgram != nil {
		if m, err := generateItemKeyData(userFeatures, items, d.generateUserProgram); err == nil {
			itemIdGenMap = m
		}
	}

	for _, item := range items {
		itemId := getItemKeyData(itemIdGenMap, item)

		properties, found := cache.Get(itemId)
		if !found {
			properties = d.defaultFieldValues
			if properties == nil {
				properties = map[string]any{}
			}
		}

		item.AddProperties(properties)
		// Consistent with ItemStateFilterFeatureStoreDao: when a generated query
		// key is used and the entity join id is item_id, the cached join id value
		// equals the generated key, so AddProperties would overwrite the item's
		// real item_id. Restore it so transFunc, FilterParams and downstream
		// stages observe the same value as the FeatureStore adapter.
		restoreItemId := found && d.generateUserProgram != nil && joinId == "item_id"
		if restoreItemId {
			item.AddProperty(joinId, string(item.Id))
		}

		if d.transFunc != nil {
			d.transFunc(user, item, ctx)
			properties = item.GetProperties()
		} else if restoreItemId {
			properties = item.GetProperties()
		}

		if d.filterParam != nil {
			result, err := d.filterParam.EvaluateByDomain(userFeatures, properties)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=ItemStateFilterFeatureViewCacheDao\tevent=EvaluateFilterParam\terror=%v", ctx.RecommendId, err))
			} else if result {
				ret = append(ret, item)
			}
		} else {
			ret = append(ret, item)
		}
	}
	return
}
