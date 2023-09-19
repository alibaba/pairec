package module

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/tablestoredb"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type UserCollaborativeTableStoreDao struct {
	tablestore *tablestoredb.TableStore
	itemType   string
	userTable  string
	itemTable  string
	recallName string

	normalization bool
}

func NewUserCollaborativeTableStoreDao(config recconf.RecallConfig) *UserCollaborativeTableStoreDao {
	dao := &UserCollaborativeTableStoreDao{}
	tablestore, err := tablestoredb.GetTableStore(config.UserCollaborativeDaoConf.TableStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.tablestore = tablestore
	dao.userTable = config.UserCollaborativeDaoConf.User2ItemTable
	dao.itemTable = config.UserCollaborativeDaoConf.Item2ItemTable
	dao.itemType = config.ItemType
	dao.recallName = config.Name
	if config.UserCollaborativeDaoConf.Normalization == "on" || config.UserCollaborativeDaoConf.Normalization == "" {
		dao.normalization = true
	}
	return dao
}

func (d *UserCollaborativeTableStoreDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	getRowRequest := new(tablestore.GetRowRequest)
	criteria := new(tablestore.SingleRowQueryCriteria)
	putPk := new(tablestore.PrimaryKey)
	putPk.AddPrimaryKeyColumn("user_id", string(user.Id))

	criteria.PrimaryKey = putPk
	criteria.ColumnsToGet = []string{"item_ids"}
	getRowRequest.SingleRowQueryCriteria = criteria
	getRowRequest.SingleRowQueryCriteria.TableName = d.userTable
	getRowRequest.SingleRowQueryCriteria.MaxVersion = 1
	start := time.Now()
	getResp, err := d.tablestore.Client.GetRow(getRowRequest)
	log.Debug(fmt.Sprintf("UserCollaborativeTableStoreDao resp, cost=%d", utils.CostTime(start)))
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeTableStoreDao\terror=%v", context.RecommendId, err))
		return
	}
	if len(getResp.Columns) == 0 {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeTableStoreDao\tuid=%s\terr=item ids empty", context.RecommendId, user.Id))
		return
	}

	var ids string

	if str, ok := getResp.Columns[0].Value.(string); ok {
		ids = str
	}
	itemIds := strings.Split(ids, ",")
	if len(itemIds) == 0 {
		log.Info(fmt.Sprintf("module=UserCollaborativeTableStoreDao\tuid=%s\terr=item ids empty", user.Id))
		return
	}

	preferScoreMap := make(map[string]float64)

	cpuCount := 4
	maps := make(map[int][]string)
	for i, id := range itemIds {
		ss := strings.Split(id, ":")
		if ss[0] == "" {
			continue
		}
		preferScoreMap[ss[0]] = 1
		if len(ss) > 1 {
			if score, err := strconv.ParseFloat(ss[1], 64); err == nil {
				preferScoreMap[ss[0]] = score
			} else {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeTableStoreDao\tevent=ParsePreferScore\tuid=%s\terr=%v", context.RecommendId, user.Id, err))
			}
		}
		maps[i%cpuCount] = append(maps[i%cpuCount], ss[0])
	}

	itemIdCh := make(chan []string, cpuCount)
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
					batchGetRowRequest := new(tablestore.BatchGetRowRequest)
					multiRowQueryCriteria := new(tablestore.MultiRowQueryCriteria)
					multiRowQueryCriteria.ColumnsToGet = []string{"similar_item_ids"}
					multiRowQueryCriteria.TableName = d.itemTable
					multiRowQueryCriteria.MaxVersion = 1

					for _, id := range ids {
						putPk := new(tablestore.PrimaryKey)
						putPk.AddPrimaryKeyColumn("item_id", id)
						multiRowQueryCriteria.AddRow(putPk)
					}
					batchGetRowRequest.MultiRowQueryCriteria = []*tablestore.MultiRowQueryCriteria{multiRowQueryCriteria}

					batchGetResp, err := d.tablestore.Client.BatchGetRow(batchGetRowRequest)
					if err != nil {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeTableStoreDao\terror=%v", context.RecommendId, err))
						goto LOOP
					}

					for _, rows := range batchGetResp.TableToRowsResult {
						for _, row := range rows {
							if row.IsSucceed && len(row.Columns) > 0 {
								var preferScore float64 = 1
								pks := row.PrimaryKey.PrimaryKeys
								for _, pk := range pks {
									if pk.ColumnName == "item_id" {
										triggerId, _ := pk.Value.(string)
										preferScore = preferScoreMap[triggerId]
									}
								}

								if str, ok := row.Columns[0].Value.(string); ok {
									list := strings.Split(str, ",")
									for _, id := range list {
										strs := strings.Split(id, ":")
										if len(strs) == 1 && strs[0] != "" && strs[0] != "null" {
											item := NewItem(strs[0])
											item.RetrieveId = d.recallName
											item.ItemType = d.itemType

											result = append(result, item)

										} else if len(strs) == 2 && len(strs[0]) > 0 && strs[0] != "null" {
											item := NewItem(strs[0])
											item.RetrieveId = d.recallName
											item.ItemType = d.itemType
											if tmpScore, err := strconv.ParseFloat(strs[1], 64); err == nil {
												item.Score = tmpScore * preferScore
											} else {
												item.Score = preferScore
											}

											result = append(result, item)
										}
									}
								}

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
func (d *UserCollaborativeTableStoreDao) GetTriggers(user *User, context *context.RecommendContext) (itemTriggers map[string]float64) {
	itemTriggers = make(map[string]float64)
	triggerInfos := d.GetTriggerInfos(user, context)

	for _, trigger := range triggerInfos {
		itemTriggers[trigger.ItemId] = trigger.Weight
	}

	return
}

func (d *UserCollaborativeTableStoreDao) GetTriggerInfos(user *User, context *context.RecommendContext) (triggerInfos []*TriggerInfo) {
	getRowRequest := new(tablestore.GetRowRequest)
	criteria := new(tablestore.SingleRowQueryCriteria)
	putPk := new(tablestore.PrimaryKey)
	putPk.AddPrimaryKeyColumn("user_id", string(user.Id))

	criteria.PrimaryKey = putPk
	criteria.ColumnsToGet = []string{"item_ids"}
	getRowRequest.SingleRowQueryCriteria = criteria
	getRowRequest.SingleRowQueryCriteria.TableName = d.userTable
	getRowRequest.SingleRowQueryCriteria.MaxVersion = 1
	getResp, err := d.tablestore.Client.GetRow(getRowRequest)

	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeTableStoreDao\terror=%v", context.RecommendId, err))
		return
	}
	if len(getResp.Columns) == 0 {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeTableStoreDao\tuid=%s\terr=item ids empty", context.RecommendId, user.Id))
		return
	}

	var ids string

	if str, ok := getResp.Columns[0].Value.(string); ok {
		ids = str
	}

	itemIds := strings.Split(ids, ",")
	if len(itemIds) == 0 {
		log.Info(fmt.Sprintf("module=UserCollaborativeTableStoreDao\tuid=%s\terr=item ids empty", user.Id))
		return
	}

	for _, id := range itemIds {
		ss := strings.Split(id, ":")
		if ss[0] == "" {
			continue
		}
		trigger := &TriggerInfo{
			ItemId: ss[0],
			Weight: 1,
		}

		if len(ss) > 1 {
			if score, err := strconv.ParseFloat(ss[1], 64); err == nil {
				trigger.Weight = score
			} else {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeTableStoreDao\tevent=ParsePreferScore\tuid=%s\terr=%v", context.RecommendId, user.Id, err))
			}
		}
		triggerInfos = append(triggerInfos, trigger)
	}
	return
}
