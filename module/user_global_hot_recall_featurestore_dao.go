package module

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type UserGlobalHotRecallFeatureStoreDao struct {
	fsClient       *fs.FSClient
	itemType       string
	recallName     string
	table          string
	itemIdField    string
	itemScoreField string
	recallCount    int
}

func NewUserGlobalHotRecallFeatureStoreDao(config recconf.RecallConfig) *UserGlobalHotRecallFeatureStoreDao {
	fsclient, err := fs.GetFeatureStoreClient(config.DaoConf.FeatureStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao := &UserGlobalHotRecallFeatureStoreDao{
		recallCount:    config.RecallCount,
		fsClient:       fsclient,
		table:          config.DaoConf.FeatureStoreViewName,
		itemIdField:    config.DaoConf.ItemIdField,
		itemScoreField: config.DaoConf.ItemScoreField,
		itemType:       config.ItemType,
		recallName:     config.Name,
	}
	return dao
}

func (d *UserGlobalHotRecallFeatureStoreDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	itemField := "item_ids"
	if d.itemIdField != "" {
		itemField = d.itemIdField
	}

	featureView := d.fsClient.GetProject().GetFeatureView(d.table)
	if featureView == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserGlobalHotRecallFeatureStoreDao\terror=featureView not found, table:%s", context.RecommendId, d.table))
		return
	}

	features, err := featureView.GetOnlineFeatures([]any{"-1"}, []string{"*"}, map[string]string{})
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserGlobalHotRecallFeatureStoreDao\terror=%v", context.RecommendId, err))
		return
	}

	if len(features) == 0 {
		return
	}
	itemIdsStr := utils.ToString(features[0][itemField], "")
	if itemIdsStr == "" {
		return
	}

	itemIds := make([]string, 0, d.recallCount)
	idList := strings.Split(itemIdsStr, ",")
	for _, id := range idList {
		if id != "" {
			itemIds = append(itemIds, id)
		}
	}

	if len(itemIds) == 0 {
		return
	}

	for _, id := range itemIds {
		strs := strings.Split(id, ":")
		if len(strs) == 1 {
			// itemid
			item := NewItem(strs[0])
			item.ItemType = d.itemType
			item.RetrieveId = d.recallName
			item.Score = rand.Float64()
			ret = append(ret, item)
		} else if len(strs) == 2 {
			// itemid:RetrieveId
			item := NewItem(strs[0])
			item.ItemType = d.itemType
			if strs[1] != "" {
				item.RetrieveId = strs[1]
			} else {
				item.RetrieveId = d.recallName
			}
			item.Score = rand.Float64()
			ret = append(ret, item)
		} else if len(strs) == 3 {
			item := NewItem(strs[0])
			item.ItemType = d.itemType
			if strs[1] != "" {
				item.RetrieveId = strs[1]
			} else {
				item.RetrieveId = d.recallName
			}
			item.Score = utils.ToFloat(strs[2], float64(0))
			item.AddAlgoScore("hot_score", item.Score)
			ret = append(ret, item)
		}
	}

	if len(ret) > d.recallCount {
		sort.Sort(sort.Reverse(ItemScoreSlice(ret)))
		return ret[:d.recallCount]
	}
	return
}
