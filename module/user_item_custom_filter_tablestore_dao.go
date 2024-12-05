package module

import (
	"fmt"
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/tablestoredb"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
)

type User2ItemCustomFilterTableStoreDao struct {
	table      string
	tablestore *tablestoredb.TableStore
}

func NewUser2ItemCustomFilterTableStoreDao(config recconf.FilterConfig) *User2ItemCustomFilterTableStoreDao {
	dao := &User2ItemCustomFilterTableStoreDao{}
	tablestore, err := tablestoredb.GetTableStore(config.DaoConf.TableStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.tablestore = tablestore
	dao.table = config.DaoConf.TableStoreTableName
	return dao
}

func (d *User2ItemCustomFilterTableStoreDao) Filter(uid UID, items []*Item, ctx *context.RecommendContext) (ret []*Item) {
	getRowRequest := new(tablestore.GetRowRequest)
	criteria := new(tablestore.SingleRowQueryCriteria)
	putPk := new(tablestore.PrimaryKey)
	putPk.AddPrimaryKeyColumn("user_id", string(uid))

	criteria.PrimaryKey = putPk
	criteria.ColumnsToGet = []string{"item_ids"}
	getRowRequest.SingleRowQueryCriteria = criteria
	getRowRequest.SingleRowQueryCriteria.TableName = d.table
	getRowRequest.SingleRowQueryCriteria.MaxVersion = 1

	getResp, err := d.tablestore.Client.GetRow(getRowRequest)

	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemCustomFilterTableStoreDao\terr=%v", ctx.RecommendId, err))
		ret = items
		return
	}

	if len(getResp.Columns) == 0 {
		ret = items
		return
	}
	var ids string
	if str, ok := getResp.Columns[0].Value.(string); ok {
		ids = str
	}

	idList := strings.Split(ids, ",")
	if idList == nil {
		ret = items
		return
	}
	fiterIds := make(map[string]bool, len(idList))

	for _, id := range idList {
		fiterIds[id] = true
	}

	for _, item := range items {
		if _, ok := fiterIds[string(item.Id)]; !ok {
			ret = append(ret, item)
		}
	}
	return
}
