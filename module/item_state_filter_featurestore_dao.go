package module

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/goburrow/cache"
)

type ItemStateFilterFeatureStoreDao struct {
	fsClient            *fs.FSClient
	table               string
	itemFieldName       string
	selectFields        []string
	filterParam         *FilterParam
	itmCache            cache.Cache
	defaultFieldValues  map[string]any
	generateUserProgram *vm.Program
	transFunc           FeatureTransFunc
}

func NewItemStateFilterFeatureStoreDao(config recconf.FilterConfig, transFunc FeatureTransFunc) *ItemStateFilterFeatureStoreDao {

	fsclient, err := fs.GetFeatureStoreClient(config.ItemStateDaoConf.FeatureStoreName)
	if err != nil {
		panic(fmt.Sprintf("error=%v", err))
	}

	dao := &ItemStateFilterFeatureStoreDao{
		fsClient:      fsclient,
		table:         config.ItemStateDaoConf.FeatureStoreViewName,
		itemFieldName: config.ItemStateDaoConf.ItemFieldName,
		selectFields:  []string{"*"},
		//selectFields:  config.ItemStateDaoConf.SelectFields,
		defaultFieldValues: config.ItemStateDaoConf.DefaultFieldValues,
		transFunc:          transFunc,
	}
	if config.ItemStateDaoConf.SelectFields != "" {
		fields := strings.Split(config.ItemStateDaoConf.SelectFields, ",")
		dao.selectFields = make([]string, 0, len(fields))
		for _, field := range fields {
			dao.selectFields = append(dao.selectFields, strings.TrimSpace(field))
		}
	}

	if config.ItemStateCacheSize > 0 {
		cacheTime := 3600
		if config.ItemStateCacheTime > 0 {
			cacheTime = config.ItemStateCacheTime
		}
		dao.itmCache = cache.New(cache.WithMaximumSize(config.ItemStateCacheSize),
			cache.WithExpireAfterWrite(time.Second*time.Duration(cacheTime)))
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

func (d *ItemStateFilterFeatureStoreDao) Filter(user *User, items []*Item, ctx *context.RecommendContext) (ret []*Item) {
	fields := make(map[string]bool, len(items))
	cpuCount := utils.MaxInt(int(math.Ceil(float64(len(items))/float64(requestCount))), 1)

	requestCh := make(chan []interface{}, cpuCount)
	maps := make(map[int][]interface{}, cpuCount)
	itemMap := make(map[string]*Item, len(items))
	var itemIdGenMap map[string]string
	index := 0
	userFeatures := user.MakeUserFeatures2()
	if d.generateUserProgram != nil {
		if m, err := generateItemKeyData(userFeatures, items, d.generateUserProgram); err == nil {
			itemIdGenMap = m
		}
	}
	for i, item := range items {
		itemId := getItemKeyData(itemIdGenMap, item)
		if d.itmCache != nil {
			if attrs, ok := d.itmCache.GetIfPresent(itemId); ok {
				properties := attrs.(map[string]interface{})
				item.AddProperties(properties)
				if d.transFunc != nil {
					d.transFunc(user, item, ctx)
					properties = item.GetProperties()
				}
				if d.filterParam != nil {
					result, err := d.filterParam.EvaluateByDomain(userFeatures, properties)
					if err == nil && result {
						fields[itemId] = true
					}
				} else {
					fields[itemId] = true
				}
				continue
			}
		}
		itemMap[itemId] = item
		maps[index%cpuCount] = append(maps[index%cpuCount], itemId)
		if (i+1)%requestCount == 0 {
			index++
		}
	}

	defer close(requestCh)
	for _, idlist := range maps {
		requestCh <- idlist
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	mergeFunc := func(maps map[string]bool) {
		mu.Lock()
		for k, v := range maps {
			fields[k] = v
		}
		mu.Unlock()
	}
	for i := 0; i < cpuCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case idlist := <-requestCh:
				fieldMap := make(map[string]bool, len(idlist))
				addPropertyMap := make(map[string]struct{}, len(idlist))

				featureView := d.fsClient.GetProject().GetFeatureView(d.table)
				if featureView == nil {
					log.Error(fmt.Sprintf("module=ItemStateFilterFeatureStoreDao\terror=featureView not found, table:%s", d.table))
					return
				}
				features, err := featureView.GetOnlineFeatures(idlist, d.selectFields, map[string]string{})
				if err != nil {
					// if error , not filter item
					log.Error(fmt.Sprintf("module=ItemStateFilterFeatureStoreDao\terror=%v", err))
					for _, id := range idlist {
						fieldMap[id.(string)] = true
					}
					mergeFunc(fieldMap)
					return
				}
				featureEntity := d.fsClient.GetProject().GetFeatureEntity(featureView.GetFeatureEntityName())
				if featureEntity == nil {
					log.Error(fmt.Sprintf("module=ItemStateFilterFeatureStoreDao\terror=featureEntity not found, name:%s", featureView.GetFeatureEntityName()))
					return
				}
				for _, itemFeatures := range features {
					itemId := utils.ToString(itemFeatures[featureEntity.FeatureEntityJoinid], "")
					if itemId != "" {
						if item, ok := itemMap[itemId]; ok {
							if d.generateUserProgram != nil && featureEntity.FeatureEntityJoinid == "item_id" { // if featureEntity.FeatureEntityJoinid is item_id
								itemFeatures[featureEntity.FeatureEntityJoinid] = string(item.Id)
							}
							item.AddProperties(itemFeatures)
							addPropertyMap[itemId] = struct{}{}
							if d.itmCache != nil {
								d.itmCache.Put(itemId, itemFeatures)
							}
							if d.transFunc != nil {
								d.transFunc(user, item, ctx)
								itemFeatures = item.GetProperties()
							}
							if d.filterParam != nil {
								result, err := d.filterParam.EvaluateByDomain(userFeatures, itemFeatures)
								if err == nil && result {
									fieldMap[itemId] = true
								}
							} else {
								fieldMap[itemId] = true
							}
						}
					}
				}
				if len(d.defaultFieldValues) > 0 {
					for _, id := range idlist {
						itemId := id.(string)
						if _, ok := addPropertyMap[itemId]; !ok {
							if item, ok := itemMap[itemId]; ok {
								item.AddProperties(d.defaultFieldValues)
								if d.itmCache != nil {
									d.itmCache.Put(itemId, d.defaultFieldValues)
								}
								properties := d.defaultFieldValues
								if d.transFunc != nil {
									d.transFunc(user, item, ctx)
									properties = item.GetProperties()
								}

								if d.filterParam != nil {
									result, err := d.filterParam.EvaluateByDomain(userFeatures, properties)
									if err == nil && result {
										fieldMap[itemId] = true
									}
								} else {
									fieldMap[itemId] = true
								}
							}
						}
					}
				}
				mergeFunc(fieldMap)
			default:
			}
		}()
	}

	wg.Wait()

	for _, item := range items {
		itemId := getItemKeyData(itemIdGenMap, item)
		if _, ok := fields[itemId]; ok {
			ret = append(ret, item)
		}
	}
	return
}

func generateItemKeyData(userFeatures map[string]any, items []*Item, p *vm.Program) (map[string]string, error) {
	m := make(map[string]string, len(items))
	if p != nil {
		params := map[string]any{
			"user": userFeatures,
		}

		for _, item := range items {
			itemFeatures := item.GetProperties()
			itemFeatures["recall_name"] = item.RetrieveId
			itemFeatures["item_id"] = string(item.Id)
			params["item"] = itemFeatures
			if output, err := expr.Run(p, params); err != nil {
				log.Error(fmt.Sprintf("module=ItemStateFilterDao\terror=generate item key data failed, params:%v, err:%v", params, err))
			} else {
				if str := utils.ToString(output, ""); str != "" {
					m[string(item.Id)] = str
				} else {
					log.Error(fmt.Sprintf("module=ItemStateFilterDao\terror=output error(%v), output:%v ", err, output))
				}

			}
		}
	}

	return m, nil

}

func getItemKeyData(itemIdMap map[string]string, item *Item) string {
	if len(itemIdMap) == 0 {
		return string(item.Id)
	}
	if data, ok := itemIdMap[string(item.Id)]; ok {
		return data
	}
	return string(item.Id)
}
