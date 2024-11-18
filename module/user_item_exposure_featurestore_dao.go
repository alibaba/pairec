package module

import (
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/datasource/featuredb/fdbserverpb"
)

type User2ItemExposureFeatureStoreDao struct {
	fsClient                 *fs.FSClient
	table                    string
	timeInterval             int64 //  second
	generateItemDataFuncName string
	writeLogExcludeScenes    map[string]bool
	clearLogScene            string
}

func NewUser2ItemExposureFeatureStoreDao(config recconf.FilterConfig) *User2ItemExposureFeatureStoreDao {
	fsclient, err := fs.GetFeatureStoreClient(config.DaoConf.FeatureStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao := &User2ItemExposureFeatureStoreDao{
		timeInterval:             0,
		generateItemDataFuncName: config.GenerateItemDataFuncName,
		writeLogExcludeScenes:    make(map[string]bool),
		clearLogScene:            config.ClearLogIfNotEnoughScene,
		fsClient:                 fsclient,
	}
	dao.table = config.DaoConf.FeatureStoreViewName

	if config.TimeInterval > 0 {
		dao.timeInterval = int64(config.TimeInterval)
	}
	for _, scene := range config.WriteLogExcludeScenes {
		dao.writeLogExcludeScenes[scene] = true
	}
	return dao
}

func (d *User2ItemExposureFeatureStoreDao) LogHistory(user *User, items []*Item, context *context.RecommendContext) {
	start := time.Now()
	scene := context.GetParameter("scene").(string)
	if _, exist := d.writeLogExcludeScenes[scene]; exist {
		return
	}

	if len(items) == 0 {
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureFeatureStoreDao\terr=items empty", context.RecommendId))
		return
	}

	project := d.fsClient.GetProject()
	featureView := project.GetFeatureView(d.table)
	if featureView == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureFeatureStoreDao\terror=table not found, name:%s", context.RecommendId, d.table))
		return
	}

	request := new(fdbserverpb.BatchWriteKVReqeust)
	uid := string(user.Id)

	ttl := int64(featureView.GetTTL())

	ts := start.Unix() - ttl + d.timeInterval

	for _, item := range items {
		itemData := getGenerateItemDataFunc(d.generateItemDataFuncName)(user.Id, item)
		request.Kvs = append(request.Kvs, &fdbserverpb.KVData{
			Key:   uid,
			Value: []byte(itemData),
			Ts:    ts * 1000, // ms
		})
	}

	err := fdbserverpb.BatchWriteBloomKV(project, featureView, request)

	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureFeatureStoreDao\tuid=%s\terr=%v", context.RecommendId, user.Id, err))
		return
	}

	log.Info(fmt.Sprintf("requestId=%s\tscene=%s\tuid=%s\tmsg=log history success\tcost=%d", context.RecommendId, scene, user.Id, utils.CostTime(start)))
}
func (d *User2ItemExposureFeatureStoreDao) FilterByHistory(uid UID, items []*Item) (ret []*Item) {
	project := d.fsClient.GetProject()
	featureView := project.GetFeatureView(d.table)
	if featureView == nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureFeatureStoreDao\terror=table not found, name:%s", d.table))
		ret = items
		return
	}

	request := new(fdbserverpb.TestBloomItemsRequest)

	request.Key = string(uid)
	for _, item := range items {
		itemData := getGenerateItemDataFunc(d.generateItemDataFuncName)(uid, item)
		request.Items = append(request.Items, itemData)
	}

	tests, err := fdbserverpb.TestBloomItems(project, featureView, request)
	if err != nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureFeatureStoreDao\terr=%v", err))
		ret = items
		return
	}

	ret = make([]*Item, 0, len(items))
	for i, test := range tests {
		if !test {
			ret = append(ret, items[i])
		}
	}

	return
}

func (d *User2ItemExposureFeatureStoreDao) ClearHistory(user *User, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)
	if scene != d.clearLogScene {
		return
	}
	project := d.fsClient.GetProject()
	featureView := project.GetFeatureView(d.table)
	if featureView == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureFeatureStoreDao\terror=table not found, name:%s", context.RecommendId, d.table))
		return
	}

	err := fdbserverpb.DeleteBloomByKey(project, featureView, string(user.Id))
	if err != nil {
		context.LogError(fmt.Sprintf("delete user [%s] exposure items failed, err=%v", user.Id, err))
	}
}

func (d *User2ItemExposureFeatureStoreDao) GetExposureItemIds(user *User, context *context.RecommendContext) (ret string) {
	return
}
