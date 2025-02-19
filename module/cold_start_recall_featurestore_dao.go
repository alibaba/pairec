package module

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type ColdStartRecallFeatureStoreDao struct {
	fsClient     *fs.FSClient
	recallCount  int
	timeInterval int
	recallName   string
	table        string
	whereClause  string
	ch           chan string
	itemIds      []string
	lastScanTime time.Time // last scan data time
}

func NewColdStartRecallFeatureStoreDao(config recconf.RecallConfig) *ColdStartRecallFeatureStoreDao {
	fsclient, err := fs.GetFeatureStoreClient(config.ColdStartDaoConf.FeatureStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao := &ColdStartRecallFeatureStoreDao{
		fsClient:     fsclient,
		recallCount:  config.RecallCount,
		table:        config.ColdStartDaoConf.FeatureStoreViewName,
		recallName:   config.Name,
		timeInterval: config.ColdStartDaoConf.TimeInterval,
		whereClause:  config.ColdStartDaoConf.WhereClause,
		ch:           make(chan string, 1000),
		itemIds:      make([]string, 0, 1024),
	}
	featureView := dao.fsClient.GetProject().GetFeatureView(dao.table)
	if featureView == nil {
		panic(fmt.Sprintf("featureView not found, table:%s", dao.table))
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
	where := d.whereClause
	createTime := time.Now().Add(time.Duration(-1*d.timeInterval) * time.Second)
	where = strings.ReplaceAll(where, "${time}", utils.ToString(createTime.Unix(), "0"))
	var (
		ids []string
		err error
	)
	if featureView.GetType() == "Batch" {
		ids, err = featureView.ScanAndIterateData(where, nil)
	} else {
		ids, err = featureView.ScanAndIterateData(where, d.ch)
	}
	d.lastScanTime = time.Now()
	if err != nil {
		log.Error(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\terror=%v", err))
		return
	}

	d.itemIds = ids
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
		ids = ids[:0]
		d.itemIds = newItemIds
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
	featureView := d.fsClient.GetProject().GetFeatureView(d.table)
	if featureView == nil {
		log.Error(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\trecallName=%s\terror=featureView not found, table:%s", d.recallName, d.table))
		return
	}

	for _, itemId := range d.itemIds {
		item := NewItem(itemId)
		item.RetrieveId = d.recallName
		ret = append(ret, item)
	}

	go func() {
		if time.Since(d.lastScanTime) <= time.Duration(30)*time.Minute {
			return
		}
		d.lastScanTime = time.Now()
		where := d.whereClause
		createTime := time.Now().Add(time.Duration(-1*d.timeInterval) * time.Second)
		where = strings.ReplaceAll(where, "${time}", utils.ToString(createTime.Unix(), "0"))
		ids, err := featureView.ScanAndIterateData(where, nil)
		if err != nil {
			log.Error(fmt.Sprintf("module=ColdStartRecallFeatureStoreDao\terror=%v", err))
			return
		}

		d.itemIds = ids

	}()

	rand.Shuffle(len(ret), func(i, j int) {
		ret[i], ret[j] = ret[j], ret[i]
	})
	if len(ret) > d.recallCount {
		ret = ret[:d.recallCount]
	}
	return

}
