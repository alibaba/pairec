package module

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type UserCollaborativeFeatureStoreDao struct {
	fsClient   *fs.FSClient
	userTable  string
	itemTable  string
	itemType   string
	recallName string

	normalization bool
}

func NewUserCollaborativeFeatureStoreDao(config recconf.RecallConfig) *UserCollaborativeFeatureStoreDao {
	fsclient, err := fs.GetFeatureStoreClient(config.UserCollaborativeDaoConf.FeatureStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao := &UserCollaborativeFeatureStoreDao{
		fsClient:   fsclient,
		userTable:  config.UserCollaborativeDaoConf.User2ItemFeatureViewName,
		itemTable:  config.UserCollaborativeDaoConf.Item2ItemFeatureViewName,
		itemType:   config.ItemType,
		recallName: config.Name,
	}

	if config.UserCollaborativeDaoConf.Normalization == "on" || config.UserCollaborativeDaoConf.Normalization == "" {
		dao.normalization = true
	}
	return dao
}

func (d *UserCollaborativeFeatureStoreDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	uid := string(user.Id)
	featureView := d.fsClient.GetProject().GetFeatureView(d.userTable)
	if featureView == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeFeatureStoreDao\terror=featureView not found, table:%s", context.RecommendId, d.userTable))
		return
	}

	features, err := featureView.GetOnlineFeatures([]any{uid}, []string{"item_ids"}, map[string]string{})
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeFeatureStoreDao\terror=%v", context.RecommendId, err))
		return
	}

	if len(features) == 0 {
		return
	}
	itemIds := make([]string, 0)
	preferScoreMap := make(map[string]float64)
	itemIdsStr := features[0]["item_ids"]
	if ids := utils.ToString(itemIdsStr, ""); ids != "" {
		idList := strings.Split(ids, ",")
		for _, id := range idList {
			strs := strings.Split(id, ":")
			if strs[0] == "" {
				continue
			}
			itemIds = append(itemIds, strs[0])
			preferScoreMap[strs[0]] = 1
			if len(strs) > 1 {
				if score, err := strconv.ParseFloat(strs[1], 64); err == nil {
					preferScoreMap[strs[0]] = score
				} else {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeFeatureStoreDao\tevent=ParsePreferScore\tuid=%s\terr=%v", context.RecommendId, uid, err))
				}
			}
		}
	}

	if len(itemIds) == 0 {
		log.Info(fmt.Sprintf("module=UserCollaborativeFeatureStoreDao\tuid=%s\terr=item ids empty", uid))
		return
	}

	if len(itemIds) > 200 {
		rand.Shuffle(len(itemIds)/2, func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})

		itemIds = itemIds[:200]
	}

	cpuCount := 4
	maps := make(map[int][]interface{})
	for i, id := range itemIds {
		maps[i%cpuCount] = append(maps[i%cpuCount], id)
	}

	itemIdCh := make(chan []interface{}, cpuCount)
	for _, ids := range maps {
		itemIdCh <- ids
	}

	itemCh := make(chan []*Item, cpuCount)
	for i := 0; i < cpuCount; i++ {
		go func() {
			result := make([]*Item, 0)
		LOOP:
			for {
				select {
				case ids := <-itemIdCh:
					featureView := d.fsClient.GetProject().GetFeatureView(d.itemTable)
					if featureView == nil {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeFeatureStoreDao\terror=featureView not found, table:%s", context.RecommendId, d.userTable))
						goto LOOP
					}

					featureEntity := d.fsClient.GetProject().GetFeatureEntity(featureView.GetFeatureEntityName())
					if featureEntity == nil {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeFeatureStoreDao\terror=featureEntity not found, name:%s", context.RecommendId, featureView.GetFeatureEntityName()))
						goto LOOP
					}
					features, err := featureView.GetOnlineFeatures(ids, []string{"similar_item_ids"}, map[string]string{})
					if err != nil {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeFeatureStoreDao\terror=%v", context.RecommendId, err))
						goto LOOP
					}

					if len(features) == 0 {
						goto LOOP
					}

					for _, feature := range features {
						triggerId := utils.ToString(feature[featureEntity.FeatureEntityJoinid], "")
						ids := utils.ToString(feature["similar_item_ids"], "")
						if triggerId == "" || ids == "" {
							continue
						}

						preferScore := preferScoreMap[triggerId]

						list := strings.Split(ids, ",")
						for _, str := range list {
							strs := strings.Split(str, ":")
							if len(strs) == 2 && len(strs[0]) > 0 && strs[0] != "null" {
								item := NewItem(strs[0])
								item.RetrieveId = d.recallName
								item.ItemType = d.itemType
								if tmpScore, err := strconv.ParseFloat(strings.TrimSpace(strs[1]), 64); err == nil {
									item.Score = tmpScore * preferScore
								} else {
									item.Score = preferScore
								}

								result = append(result, item)
							}

						}

					}
				default:
					goto DONE

				}
			}
		DONE:
			itemCh <- result
		}()
	}

	ret = mergeUserCollaborativeItemsResult(itemCh, cpuCount, d.normalization)

	close(itemCh)
	close(itemIdCh)
	return
}

func (d *UserCollaborativeFeatureStoreDao) GetTriggers(user *User, context *context.RecommendContext) (itemTriggers map[string]float64) {
	itemTriggers = make(map[string]float64)
	triggerInfos := d.GetTriggerInfos(user, context)

	for _, trigger := range triggerInfos {
		itemTriggers[trigger.ItemId] = trigger.Weight
	}

	return
}

func (d *UserCollaborativeFeatureStoreDao) GetTriggerInfos(user *User, context *context.RecommendContext) (triggerInfos []*TriggerInfo) {

	return

}
