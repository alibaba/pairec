package module

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/tablestoredb"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type UserCustomRecallTableStoreDao struct {
	tablestore  *tablestoredb.TableStore
	itemType    string
	table       string
	recallName  string
	recallCount int
}

func NewUserCustomRecallTableStoreDao(config recconf.RecallConfig) *UserCustomRecallTableStoreDao {
	dao := &UserCustomRecallTableStoreDao{
		recallCount: 1000,
	}
	tablestore, err := tablestoredb.GetTableStore(config.DaoConf.TableStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.tablestore = tablestore
	dao.table = config.DaoConf.TableStoreTableName
	dao.itemType = config.ItemType
	dao.recallName = config.Name
	return dao
}

func (d *UserCustomRecallTableStoreDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	getRowRequest := new(tablestore.GetRowRequest)
	criteria := new(tablestore.SingleRowQueryCriteria)
	putPk := new(tablestore.PrimaryKey)
	putPk.AddPrimaryKeyColumn("user_id", string(user.Id))

	criteria.PrimaryKey = putPk
	criteria.ColumnsToGet = []string{"item_ids"}
	getRowRequest.SingleRowQueryCriteria = criteria
	getRowRequest.SingleRowQueryCriteria.TableName = d.table
	getRowRequest.SingleRowQueryCriteria.MaxVersion = 1
	getResp, err := d.tablestore.Client.GetRow(getRowRequest)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCustomRecallTableStoreDao\terror=%v", context.RecommendId, err))
		return
	}

	if len(getResp.Columns) == 0 {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=UserCustomRecallTableStoreDao\tuid=%s\terr=item ids empty", context.RecommendId, user.Id))
		return
	}

	var ids string

	if str, ok := getResp.Columns[0].Value.(string); ok {
		ids = str
	}
	itemIds := strings.Split(ids, ",")
	if len(itemIds) == 0 {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=UserCustomRecallTableStoreDao\tuid=%s\terr=item ids empty", context.RecommendId, user.Id))
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
