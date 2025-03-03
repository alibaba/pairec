package module

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type ItemCollaborativeFeatureStoreDao struct {
	fsClient    *fs.FSClient
	itemType    string
	recallName  string
	table       string
	recallCount int
}

func NewItemCollaborativeFeatureStoreDao(config recconf.RecallConfig) *ItemCollaborativeFeatureStoreDao {
	fsclient, err := fs.GetFeatureStoreClient(config.ItemCollaborativeDaoConf.FeatureStoreName)
	if err != nil {
		panic(fmt.Sprintf("error=%v", err))
	}

	dao := &ItemCollaborativeFeatureStoreDao{
		fsClient:    fsclient,
		table:       config.ItemCollaborativeDaoConf.FeatureStoreViewName,
		itemType:    config.ItemType,
		recallName:  config.Name,
		recallCount: config.RecallCount,
	}
	return dao
}

func (d *ItemCollaborativeFeatureStoreDao) ListItemsByItem(user *User, context *context.RecommendContext) (ret []*Item) {
	// context get recommend item id
	item_id := utils.ToString(context.GetParameter("item_id"), "")
	if item_id == "" {
		return
	}
	featureView := d.fsClient.GetProject().GetFeatureView(d.table)
	if featureView == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeFeatureStoreDao\trecallName=%s\terror=featureView not found, table:%s", context.RecommendId, d.recallName, d.table))
		return
	}

	features, err := featureView.GetOnlineFeatures([]any{item_id}, []string{"*"}, map[string]string{})
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeFeatureStoreDao\trecallName=%s\terror=%v", context.RecommendId, d.recallName, err))
		return
	}
	if len(features) == 0 {
		return
	}

	itemIds := make([]string, 0, d.recallCount)
	var ids string
	if itemIds, exist := features[0]["item_ids"]; exist {
		ids = utils.ToString(itemIds, "")
	} else if itemIds, exist := features[0]["similar_item_ids"]; exist {
		ids = utils.ToString(itemIds, "")
	}
	idList := strings.Split(ids, ",")
	for _, id := range idList {
		if len(id) > 0 {
			itemIds = append(itemIds, id)
		}
	}

	if len(itemIds) == 0 {
		return
	}

	if len(itemIds) > d.recallCount {
		rand.Shuffle(len(itemIds)/2, func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})

		itemIds = itemIds[:d.recallCount]
	}

	for _, id := range itemIds {
		strs := strings.Split(id, ":")
		if len(strs) == 1 {
			// itemid
			item := NewItem(strs[0])
			item.ItemType = d.itemType
			item.RetrieveId = d.recallName
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
			ret = append(ret, item)
		}
	}

	return
}
