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

type UserGroupHotRecallFeatureStoreDao struct {
	fsClient    *fs.FSClient
	itemType    string
	recallName  string
	table       string
	recallCount int
	trigger     *Trigger
}

func NewUserGroupHotRecallFeatureStoreDao(config recconf.RecallConfig) *UserGroupHotRecallFeatureStoreDao {
	fsclient, err := fs.GetFeatureStoreClient(config.DaoConf.FeatureStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao := &UserGroupHotRecallFeatureStoreDao{
		recallCount: config.RecallCount,
		fsClient:    fsclient,
		table:       config.DaoConf.FeatureStoreViewName,
		itemType:    config.ItemType,
		recallName:  config.Name,
		trigger:     NewTrigger(config.Triggers),
	}
	return dao
}

func (d *UserGroupHotRecallFeatureStoreDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	featureView := d.fsClient.GetProject().GetFeatureView(d.table)
	if featureView == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserGroupHotRecallFeatureStoreDao\terror=featureView not found, table:%s", context.RecommendId, d.table))
		return
	}
	triggerId := d.trigger.GetValue(user.MakeUserFeatures2())
	if context.Debug {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=UserGroupHotRecallFeatureStoreDao\ttriggerId=%s\t", context.RecommendId, triggerId))
	}

	triggerIds := strings.Split(triggerId, TIRRGER_SPLIT)
	triggers := make([]any, 0, len(triggerIds))
	for _, trigger := range triggerIds {
		if trigger != "" {
			triggers = append(triggers, trigger)
		}
	}

	features, err := featureView.GetOnlineFeatures(triggers, []string{"item_ids"}, map[string]string{})
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserGroupHotRecallFeatureStoreDao\terror=%v", context.RecommendId, err))
		return
	}

	if len(features) == 0 {
		return
	}
	itemIds := make([]string, 0, d.recallCount)
	for _, feature := range features {
		if itemIdsStr := utils.ToString(feature["item_ids"], ""); itemIdsStr != "" {
			idList := strings.Split(itemIdsStr, ",")
			for _, id := range idList {
				if id != "" {
					itemIds = append(itemIds, id)
				}
			}

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

func (d *UserGroupHotRecallFeatureStoreDao) TriggerValue(user *User) string {
	return d.trigger.GetValue(user.MakeUserFeatures2())
}
