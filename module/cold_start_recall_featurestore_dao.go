package module

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/goburrow/cache"
)

type ColdStartRecallFeatureStoreDao struct {
	fsClient     *fs.FSClient
	recallCount  int
	recallName   string
	table        string
	ch           chan string
	itemIds      []string
	lastScanTime time.Time // last scan data time
	fields       []string
	filterParam  *FilterParam
	cache        cache.Cache
	itemCache    cache.Cache
}

func NewColdStartRecallFeatureStoreDao(config recconf.RecallConfig) *ColdStartRecallFeatureStoreDao {
	fsclient, err := fs.GetFeatureStoreClient(config.ColdStartDaoConf.FeatureStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao := &ColdStartRecallFeatureStoreDao{
		fsClient:    fsclient,
		recallCount: config.RecallCount,
		table:       config.ColdStartDaoConf.FeatureStoreViewName,
		recallName:  config.Name,
		ch:          make(chan string, 1000),
		itemIds:     make([]string, 0, 1024),

		cache:     cache.New(cache.WithMaximumSize(500000), cache.WithExpireAfterAccess(time.Minute)),
		itemCache: cache.New(cache.WithMaximumSize(500000), cache.WithExpireAfterAccess(10*time.Minute)),
	}
	featureView := dao.fsClient.GetProject().GetFeatureView(dao.table)
	if featureView == nil {
		panic(fmt.Sprintf("featureView not found, table:%s", dao.table))
	}
	if len(config.FilterParams) > 0 {
		dao.filterParam = NewFilterParamWithConfig(config.FilterParams)
		for _, filerParam := range config.FilterParams {
			dao.fields = append(dao.fields, filerParam.Name)
		}
	}
	go dao.initItemData()
	if featureView.GetType() == "Stream" {
		go dao.loopIterateData()
	}
	return dao
}
func (d *ColdStartRecallFeatureStoreDao) initItemData() {
	featureView := d.fsClient.GetProject().GetFeatureView(d.table)
	if featureView == nil {
		log.Error(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\trecallName=%s\terror=featureView not found, table:%s", d.recallName, d.table))
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
			ids, err = featureView.ScanAndIterateData("", d.ch)
		}

		if err == nil {
			break
		}
		time.Sleep(10 * time.Second)
	}
	d.lastScanTime = time.Now()
	if err != nil {
		log.Error(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\terror=%v", err))
		return
	}

	d.itemIds = ids
	go d.fetchItemData(d.itemIds)
}
func (d *ColdStartRecallFeatureStoreDao) loopIterateData() {
	ticker := time.NewTicker(time.Minute)
	var ids []string
	appendItems := func() {
		newItemIds := make([]string, len(d.itemIds))
		copy(newItemIds, d.itemIds)
		m := make(map[string]bool)
		for _, id := range newItemIds {
			m[id] = true
		}
		for _, id := range ids {
			if _, ok := m[id]; !ok {
				newItemIds = append(newItemIds, id)
			}
		}
		d.fetchItemData(ids)
		ids = ids[:0]
		d.itemIds = newItemIds
		//d.lastScanTime = time.Now()
	}
	for id := range d.ch {
		ids = append(ids, id)
		select {
		case <-ticker.C:
			if len(ids) > 0 {
				appendItems()
			}
		default:
			if len(ids) > 1000 {
				appendItems()
			}
		}
	}
}

func (d *ColdStartRecallFeatureStoreDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	var cacheKey string
	if d.filterParam != nil {
		contextFeatures := context.GetParameter("features").(map[string]interface{})
		if data, err := json.Marshal(contextFeatures); err == nil {
			cacheKey = fmt.Sprintf("%s_%s", string(user.Id), string(data))

		}
	} else {
		cacheKey = string(user.Id)
	}
	if cacheValue, ok := d.cache.GetIfPresent(cacheKey); ok {
		if items, ok := cacheValue.([]*Item); ok {
			return items
		}
	}

	if d.filterParam == nil {
		//itemIds := d.itemIds
		itemIds := make([]string, len(d.itemIds))
		copy(itemIds, d.itemIds)
		rand.Shuffle(len(itemIds), func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})
		if len(itemIds) > d.recallCount {
			itemIds = itemIds[:d.recallCount]
		}
		for _, itemId := range itemIds {
			item := NewItem(itemId)
			item.RetrieveId = d.recallName
			ret = append(ret, item)
		}
	} else {
		itemIds := make([]string, len(d.itemIds))
		copy(itemIds, d.itemIds)
		rand.Shuffle(len(itemIds), func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})

		userFeatures := user.MakeUserFeatures2()
		for _, id := range itemIds {
			cacheValue, ok := d.itemCache.GetIfPresent(id)
			if ok {
				if r, err := d.filterParam.EvaluateByDomain(userFeatures, cacheValue.(map[string]any)); err == nil && r {
					item := NewItem(id)
					item.RetrieveId = d.recallName
					ret = append(ret, item)
					if len(ret) >= d.recallCount {
						break
					}
				}
			}
		}

	}

	go func() {
		if time.Since(d.lastScanTime) <= time.Duration(60)*time.Minute {
			return
		}
		d.lastScanTime = time.Now()

		go func() {
			featureView := d.fsClient.GetProject().GetFeatureView(d.table)
			if featureView == nil {
				log.Error(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\trecallName=%s\terror=featureView not found, name:%s", d.recallName, d.table))
				return
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
				log.Error(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\terror=%v", err))
				return
			}
			if len(ids) == 0 {
				return
			}

			d.itemIds = ids
			go d.fetchItemData(d.itemIds)
			log.Info(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\tmsg=load item\tsize=%d", len(d.itemIds)))
		}()
	}()

	if len(ret) > 0 {
		d.cache.Put(cacheKey, ret)
	}
	return

}

func (d *ColdStartRecallFeatureStoreDao) fetchItemData(itemIdList []string) {
	if d.filterParam == nil {
		return
	}

	start := time.Now()
	itemIds := make([]string, len(itemIdList))
	copy(itemIds, itemIdList)
	rand.Shuffle(len(itemIds), func(i, j int) {
		itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
	})

	featureView := d.fsClient.GetProject().GetFeatureView(d.table)
	if featureView == nil {
		return
	}
	featureEntity := d.fsClient.GetProject().GetFeatureEntity(featureView.GetFeatureEntityName())
	if featureEntity == nil {
		return
	}
	size := len(itemIds)
	var cacheSize atomic.Int32
	var wg sync.WaitGroup
	ch := make(chan int, 5)
	batchSize := 200
	for i := 0; i < size; i += batchSize {
		start := i
		end := i + batchSize
		if end > size {
			end = size
		}

		wg.Add(1)
		ch <- 1
		go func(start, end int) {
			size := 0
			defer wg.Done()
			defer func() { <-ch }()
			joinIds := make([]any, 0, end-start)
			for i := start; i < end; i++ {
				joinIds = append(joinIds, itemIds[i])
			}
			features, err := featureView.GetOnlineFeatures(joinIds, d.fields, map[string]string{})
			if err != nil {
				log.Error(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\tmsg=load item cache\tfields=%v\terror=%v", d.fields, err))
				return
			}
			for _, featureMap := range features {
				itemId := featureMap[featureEntity.FeatureEntityJoinid]
				if id := utils.ToString(itemId, ""); id != "" {
					d.itemCache.Put(id, featureMap)
					size++
				}
			}
			cacheSize.Add(int32(size))
		}(start, end)
	}
	wg.Wait()
	log.Info(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\tmsg=load item cache\tsize=%d\tcachesize=%d\tcost=%d", len(itemIds), cacheSize.Load(), utils.CostTime(start)))
}
