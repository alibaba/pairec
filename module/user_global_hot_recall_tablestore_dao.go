package module

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/persist/tablestoredb"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
)

type UserGlobalHotRecallTableStoreDao struct {
	tablestore  *tablestoredb.TableStore
	itemType    string
	recallName  string
	table       string
	recallCount int
}

func NewUserGlobalHotRecallTableStoreDao(config recconf.RecallConfig) *UserGlobalHotRecallTableStoreDao {
	tablestore, err := tablestoredb.GetTableStore(config.DaoConf.TableStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao := &UserGlobalHotRecallTableStoreDao{
		recallCount: config.RecallCount,
		tablestore:  tablestore,
		table:       config.DaoConf.TableStoreTableName,
		itemType:    config.ItemType,
		recallName:  config.Name,
	}
	return dao
}

func (d *UserGlobalHotRecallTableStoreDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	getRowRequest := new(tablestore.GetRowRequest)
	criteria := new(tablestore.SingleRowQueryCriteria)
	putPk := new(tablestore.PrimaryKey)
	putPk.AddPrimaryKeyColumn("trigger_id", "-1")

	criteria.PrimaryKey = putPk
	criteria.ColumnsToGet = []string{"item_ids"}
	getRowRequest.SingleRowQueryCriteria = criteria
	getRowRequest.SingleRowQueryCriteria.TableName = d.table
	getRowRequest.SingleRowQueryCriteria.MaxVersion = 1
	getResp, err := d.tablestore.Client.GetRow(getRowRequest)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserGlobalHotRecallTableStoreDao\terror=%v", context.RecommendId, err))
		return
	}

	if len(getResp.Columns) == 0 {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=UserGlobalHotRecallTableStoreDao\tuid=%s\terr=item ids empty", context.RecommendId, user.Id))
		return
	}

	var ids string

	if str, ok := getResp.Columns[0].Value.(string); ok {
		ids = str
	}

	itemIds := strings.Split(ids, ",")
	if len(itemIds) == 0 {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=UserGlobalHotRecallTableStoreDao\tuid=%s\terr=item ids empty", context.RecommendId, user.Id))
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
