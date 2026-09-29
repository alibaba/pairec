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

// ItemStateFilterFeatureViewCacheDao is an ItemStateFilterDao implementation
// that reads item properties from a pre-populated FeatureViewCache (full-scan
// local cache) instead of making per-request GetOnlineFeatures calls.
type ItemStateFilterFeatureViewCacheDao struct {
	cache               *fs.FeatureViewCache
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
	userFeatures := user.MakeUserFeatures2()
	var itemIdGenMap map[string]string
	if d.generateUserProgram != nil {
		if m, err := generateItemKeyData(userFeatures, items, d.generateUserProgram); err == nil {
			itemIdGenMap = m
		}
	}

	for _, item := range items {
		itemId := getItemKeyData(itemIdGenMap, item)

		properties, found := d.cache.Get(itemId)
		if !found {
			if len(d.defaultFieldValues) > 0 {
				// negative cache: use default values for items not in FeatureView
				properties = d.defaultFieldValues
			} else {
				// item not in cache and no defaults: retain item (do not filter)
				ret = append(ret, item)
				continue
			}
		}

		item.AddProperties(properties)
		if d.transFunc != nil {
			d.transFunc(user, item, ctx)
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
