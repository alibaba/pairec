package module

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/goburrow/cache"
)

type ItemStateFilterFeatureStoreDao struct {
	fsClient      *fs.FSClient
	table         string
	itemFieldName string
	selectFields  []string
	filterParam   *FilterParam
	itmCache      cache.Cache
}

func NewItemStateFilterFeatureStoreDao(config recconf.FilterConfig) *ItemStateFilterFeatureStoreDao {

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
			cache.WithExpireAfterAccess(time.Second*time.Duration(cacheTime)))
	}
	if len(config.FilterParams) > 0 {
		dao.filterParam = NewFilterParamWithConfig(config.FilterParams)
	}
	return dao
}

func (d *ItemStateFilterFeatureStoreDao) Filter(user *User, items []*Item) (ret []*Item) {
	fields := make(map[string]bool, len(items))
	cpuCount := utils.MaxInt(int(math.Ceil(float64(len(items))/float64(requestCount))), 1)

	requestCh := make(chan []interface{}, cpuCount)
	maps := make(map[int][]interface{}, cpuCount)
	itemMap := make(map[ItemId]*Item, len(items))
	index := 0
	for i, item := range items {
		itemId := string(item.Id)
		if d.itmCache != nil {
			if attrs, ok := d.itmCache.GetIfPresent(itemId); ok {
				properties := attrs.(map[string]interface{})
				item.AddProperties(properties)
				if d.filterParam != nil {
					result, err := d.filterParam.Evaluate(properties)
					if err == nil && result {
						fields[itemId] = true
					}
				} else {
					fields[itemId] = true
				}
				continue
			}
		}
		itemMap[item.Id] = item
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
						if item, ok := itemMap[ItemId(itemId)]; ok {
							item.AddProperties(itemFeatures)
							if d.itmCache != nil {
								d.itmCache.Put(itemId, itemFeatures)
							}
							if d.filterParam != nil {
								result, err := d.filterParam.Evaluate(itemFeatures)
								if err == nil && result {
									fieldMap[itemId] = true
								}
							} else {
								fieldMap[itemId] = true
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
		if _, ok := fields[string(item.Id)]; ok {
			ret = append(ret, item)
		}
	}
	return
}
