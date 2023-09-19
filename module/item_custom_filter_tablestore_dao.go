package module

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/tablestoredb"
	"github.com/alibaba/pairec/v2/recconf"
)

type ItemCustomFilterTableStoreDao struct {
	table      string
	tablestore *tablestoredb.TableStore
	fiterIds   map[ItemId]bool
}

func NewItemCustomFilterTableStoreDao(config recconf.FilterConfig) *ItemCustomFilterTableStoreDao {
	dao := &ItemCustomFilterTableStoreDao{
		fiterIds: make(map[ItemId]bool),
	}
	tablestore, err := tablestoredb.GetTableStore(config.DaoConf.TableStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.tablestore = tablestore
	dao.table = config.DaoConf.TableStoreTableName

	go dao.loopLoadFiterIds()

	return dao
}
func (d *ItemCustomFilterTableStoreDao) loopLoadFiterIds() {
	for {
		ret := d.getFilterIds()
		if len(ret) > 0 {
			log.Info(fmt.Sprintf("module=ItemCustomFilterTableStoreDao\tevent=ListFilterIds\tcount=%d", len(ret)))
			d.fiterIds = ret
		}

		time.Sleep(time.Minute)
	}

}

func (d *ItemCustomFilterTableStoreDao) getFilterIds() (ret map[ItemId]bool) {
	getRangeRequest := &tablestore.GetRangeRequest{}
	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = d.table

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumnWithMinValue("item_ids")
	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumnWithMaxValue("item_ids")
	rangeRowQueryCriteria.StartPrimaryKey = startPK
	rangeRowQueryCriteria.EndPrimaryKey = endPK
	rangeRowQueryCriteria.ColumnsToGet = []string{}
	rangeRowQueryCriteria.Direction = tablestore.FORWARD
	rangeRowQueryCriteria.MaxVersion = 1
	rangeRowQueryCriteria.Limit = 1000000
	getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria

	getRangeResp, err := d.tablestore.Client.GetRange(getRangeRequest)

	ret = make(map[ItemId]bool)

	for {
		if err != nil {
			log.Error(fmt.Sprintf("module=ItemCustomFilterTableStoreDao\terr=%v", err))
			return
		}
		for _, row := range getRangeResp.Rows {
			if ids, ok := row.PrimaryKey.PrimaryKeys[0].Value.(string); ok && ids != "" {
				idList := strings.Split(ids, ",")
				for _, id := range idList {
					ret[ItemId(id)] = true
				}
			}
		}

		if getRangeResp.NextStartPrimaryKey == nil {
			break
		} else {
			getRangeRequest.RangeRowQueryCriteria.StartPrimaryKey = getRangeResp.NextStartPrimaryKey
			getRangeResp, err = d.tablestore.Client.GetRange(getRangeRequest)
		}
	}

	return
}
func (d *ItemCustomFilterTableStoreDao) GetFilterItems() (ret map[ItemId]bool) {
	if len(d.fiterIds) > 0 {
		ret = d.fiterIds
	} else {
		ret = d.getFilterIds()
	}

	return
}
