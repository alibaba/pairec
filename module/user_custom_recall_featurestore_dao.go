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

type UserCustomRecallFeatureStoreDao struct {
	fsClient    *fs.FSClient
	itemType    string
	recallName  string
	table       string
	recallCount int
}

func NewUserCustomRecallFeatureStoreDao(config recconf.RecallConfig) *UserCustomRecallFeatureStoreDao {
	fsclient, err := fs.GetFeatureStoreClient(config.DaoConf.FeatureStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao := &UserCustomRecallFeatureStoreDao{
		recallCount: config.RecallCount,
		fsClient:    fsclient,
		table:       config.DaoConf.FeatureStoreViewName,
		itemType:    config.ItemType,
		recallName:  config.Name,
	}
	return dao
}

func (d *UserCustomRecallFeatureStoreDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	uid := string(user.Id)

	featureView := d.fsClient.GetProject().GetFeatureView(d.table)
	if featureView == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCustomRecallFeatureStoreDao\terror=featureView not found, table:%s", context.RecommendId, d.table))
		return
	}

	features, err := featureView.GetOnlineFeatures([]any{uid}, []string{"*"}, map[string]string{})
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCustomRecallFeatureStoreDao\terror=%v", context.RecommendId, err))
		return
	}

	if len(features) == 0 {
		return
	}

	itemIdsStr := utils.ToString(features[0]["item_ids"], "")
	if itemIdsStr == "" {
		return
	}

	itemIds := make([]string, 0, d.recallCount)
	idList := strings.Split(itemIdsStr, ",")
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