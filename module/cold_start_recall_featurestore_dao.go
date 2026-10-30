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
	// itemFilterParam holds item only conditions, evaluated right after item
	// features are loaded so that items not matching never enter itemCache and
	// the request path never spends time on them.
	itemFilterParam *FilterParam
	// filteredItemIds holds the ids that passed ItemFilterParams while loading.
	// The request path samples from it instead of the full scan result, so the
	// per request cost is proportional to the filtered pool instead of the whole
	// feature view. Only maintained when itemFilterParam != nil.
	filteredItemIds atomic.Pointer[[]string]
	filteredMu      sync.Mutex // serializes the read-modify-write of filteredItemIds
	// loadMu serializes background pool updates. A full load evaluates a scan
	// snapshot and then replaces the pool, so it must not overlap with a stream
	// increment: a slow full load would otherwise publish last and drop the
	// increments that became visible while it was running.
	loadMu    sync.Mutex
	cache     cache.Cache
	itemCache cache.Cache
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
		itemCache: cache.New(cache.WithMaximumSize(500000), cache.WithExpireAfterAccess(time.Hour)),
	}
	featureView := dao.fsClient.GetProject().GetFeatureView(dao.table)
	if featureView == nil {
		panic(fmt.Sprintf("featureView not found, table:%s", dao.table))
	}
	// FilterParams and ItemFilterParams may reference the same field, keep fields
	// distinct so that GetOnlineFeatures is not asked for duplicated columns.
	fieldSet := make(map[string]bool, 8)
	addField := func(name string) {
		if fieldSet[name] {
			return
		}
		fieldSet[name] = true
		dao.fields = append(dao.fields, name)
	}
	if len(config.FilterParams) > 0 {
		dao.filterParam = NewFilterParamWithConfig(config.FilterParams)
		for _, filerParam := range config.FilterParams {
			addField(filerParam.Name)
		}
	}
	if len(config.ItemFilterParams) > 0 {
		dao.itemFilterParam = NewFilterParamWithConfig(config.ItemFilterParams)
		for _, filterParam := range config.ItemFilterParams {
			addField(filterParam.Name)
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
	d.fetchItemData(d.itemIds, true)
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
		d.fetchItemData(ids, false)
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
			clonedItems := make([]*Item, len(items))
			for i, item := range items {
				clonedItems[i] = item.DeepClone()
			}
			return clonedItems
		}
	}

	if d.filterParam == nil && d.itemFilterParam == nil {
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
		candidates := d.candidateIds()
		itemIds := make([]string, len(candidates))
		copy(itemIds, candidates)
		rand.Shuffle(len(itemIds), func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})

		if d.filterParam == nil {
			// ItemFilterParams was already settled at load time and nothing else
			// looks at the item features, so the pool is used as is.
			if len(itemIds) > d.recallCount {
				itemIds = itemIds[:d.recallCount]
			}
			for _, id := range itemIds {
				item := NewItem(id)
				item.RetrieveId = d.recallName
				ret = append(ret, item)
			}
		} else {
			// Only FilterParams looks at the user, ItemFilterParams was already
			// settled at load time.
			userFeatures := user.MakeUserFeatures2()
			for _, id := range itemIds {
				// Missing from itemCache means the features are not available, so
				// FilterParams cannot be evaluated and the item is skipped.
				cacheValue, ok := d.itemCache.GetIfPresent(id)
				if !ok {
					continue
				}
				if r, err := d.filterParam.EvaluateByDomain(userFeatures, cacheValue.(map[string]any)); err != nil || !r {
					continue
				}
				item := NewItem(id)
				item.RetrieveId = d.recallName
				ret = append(ret, item)
				if len(ret) >= d.recallCount {
					break
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
				// Preserve previous itemIds on transient errors to avoid flapping.
				log.Error(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\terror=%v", err))
				return
			}

			// Always overwrite itemIds with the latest full scan result,
			// including empty results, so that the recall pool stays in sync
			// with the underlying featureview. This prevents stale items from
			// being recalled after the source data has been cleared.
			d.itemIds = ids
			if len(ids) == 0 {
				log.Warning(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\trecallName=%s\tmsg=scan returned empty result, itemIds cleared", d.recallName))
				// Clear the filtered pool as well, otherwise cleared source data
				// would keep being recalled from the previous pool.
				if d.itemFilterParam != nil {
					d.loadMu.Lock()
					d.publishFilteredIds(nil, nil, true)
					d.loadMu.Unlock()
				}
				return
			}
			log.Info(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\tmsg=load item\tsize=%d", len(d.itemIds)))
			d.fetchItemData(d.itemIds, true)
		}()
	}()

	if len(ret) > 0 {
		d.cache.Put(cacheKey, ret)
	}
	return

}

func (d *ColdStartRecallFeatureStoreDao) fetchItemData(itemIdList []string, fullScan bool) {
	if d.filterParam == nil && d.itemFilterParam == nil {
		return
	}

	// Hold loadMu for the whole load so that loading the features and publishing
	// the result stay one step. A stream increment arriving meanwhile waits and
	// is applied after this load, which is the right order because it carries
	// fresher data than this load's scan snapshot.
	d.loadMu.Lock()
	defer d.loadMu.Unlock()

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
	// Only FilterParams reads the features back at request time. Without it the
	// features are useless once ItemFilterParams has been evaluated here, so only
	// the ids of the items that passed are kept.
	cacheFeatures := d.filterParam != nil
	var cacheSize atomic.Int32
	var wg sync.WaitGroup
	var resultMu sync.Mutex
	var passedIds, failedIds []string
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
				log.Error(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\trecallName=%s\tmsg=load item cache\tfields=%v\terror=%v", d.recallName, d.fields, err))
				resultMu.Lock()
				failedIds = append(failedIds, itemIds[start:end]...)
				resultMu.Unlock()
				return
			}
			passed := make([]string, 0, end-start)
			for _, featureMap := range features {
				itemId := featureMap[featureEntity.FeatureEntityJoinid]
				if id := utils.ToString(itemId, ""); id != "" {
					if d.itemFilterParam != nil {
						r, err := d.itemFilterParam.EvaluateByDomain(nil, featureMap)
						if err != nil {
							// Keep it anyway and let FilterParams have the final say.
							log.Error(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\trecallName=%s\tmsg=eval ItemFilterParams\titemId=%s\terror=%v", d.recallName, id, err))
						} else if !r {
							continue
						}
						passed = append(passed, id)
					}
					if cacheFeatures {
						d.itemCache.Put(id, featureMap)
					}
					size++
				}
			}
			if len(passed) > 0 {
				resultMu.Lock()
				passedIds = append(passedIds, passed...)
				resultMu.Unlock()
			}
			cacheSize.Add(int32(size))
		}(start, end)
	}
	wg.Wait()
	if d.itemFilterParam != nil {
		d.publishFilteredIds(passedIds, failedIds, fullScan)
	}
	log.Info(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\trecallName=%s\tmsg=load item cache\tsize=%d\tcachesize=%d\tpassed=%d\tfailed=%d\tcost=%d",
		d.recallName, len(itemIds), cacheSize.Load(), len(passedIds), len(failedIds), utils.CostTime(start)))
}

// candidateIds returns the ids the request path should sample from. Once
// ItemFilterParams has been applied while loading, only the ids that passed are
// worth looking at.
func (d *ColdStartRecallFeatureStoreDao) candidateIds() []string {
	if d.itemFilterParam == nil {
		return d.itemIds
	}
	if ids := d.filteredItemIds.Load(); ids != nil {
		return *ids
	}
	// The first load has not published a pool yet, recall nothing instead of
	// falling back to the unfiltered scan result.
	return nil
}

// publishFilteredIds updates the pool of ids that passed ItemFilterParams: a
// full scan replaces the pool, a stream increment merges into it.
func (d *ColdStartRecallFeatureStoreDao) publishFilteredIds(passed, failed []string, fullScan bool) {
	d.filteredMu.Lock()
	defer d.filteredMu.Unlock()
	old := d.filteredItemIds.Load()

	if !fullScan {
		if len(passed) == 0 {
			return
		}
		if old == nil {
			d.filteredItemIds.Store(&passed)
			return
		}
		exist := make(map[string]bool, len(*old))
		for _, id := range *old {
			exist[id] = true
		}
		merged := make([]string, len(*old), len(*old)+len(passed))
		copy(merged, *old)
		for _, id := range passed {
			if !exist[id] {
				exist[id] = true
				merged = append(merged, id)
			}
		}
		d.filteredItemIds.Store(&merged)
		return
	}

	// Ids whose feature batch failed to load keep the verdict of the previous
	// pool, otherwise a transient FeatureStore error would shrink the recall pool
	// until the next full scan. They belong to the current scan result, so this
	// cannot resurrect items already removed from the feature view.
	ids := passed
	if len(failed) > 0 && old != nil && len(*old) > 0 {
		pending := make(map[string]bool, len(failed))
		for _, id := range failed {
			pending[id] = true
		}
		ids = make([]string, 0, len(passed)+len(failed))
		ids = append(ids, passed...)
		for _, id := range *old {
			if pending[id] {
				ids = append(ids, id)
			}
		}
	}
	d.filteredItemIds.Store(&ids)
}
