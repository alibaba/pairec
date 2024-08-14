package module

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/tablestoredb"
	"github.com/alibaba/pairec/v2/recconf"
)

type User2ItemExposureTableStoreDao struct {
	tablestore               *tablestoredb.TableStore
	table                    string
	maxItems                 int32
	timeInterval             int //  second
	generateItemDataFuncName string
	writeLogExcludeScenes    map[string]bool
	clearLogScene            string
}

func NewUser2ItemExposureTableStoreDao(config recconf.FilterConfig) *User2ItemExposureTableStoreDao {
	dao := &User2ItemExposureTableStoreDao{
		maxItems:                 -1,
		timeInterval:             -1,
		generateItemDataFuncName: config.GenerateItemDataFuncName,
		writeLogExcludeScenes:    make(map[string]bool),
		clearLogScene:            config.ClearLogIfNotEnoughScene,
	}
	tablestore, err := tablestoredb.GetTableStore(config.DaoConf.TableStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.tablestore = tablestore
	dao.table = config.DaoConf.TableStoreTableName
	if config.MaxItems > 0 {
		dao.maxItems = int32(config.MaxItems)
	}

	if config.TimeInterval > 0 {
		dao.timeInterval = config.TimeInterval
	}

	for _, scene := range config.WriteLogExcludeScenes {
		dao.writeLogExcludeScenes[scene] = true
	}

	return dao
}

func (d *User2ItemExposureTableStoreDao) LogHistory(user *User, items []*Item, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)
	if _, exist := d.writeLogExcludeScenes[scene]; exist {
		return
	}

	uid := string(user.Id)
	idList := make([]string, 0)
	for _, item := range items {
		itemData := getGenerateItemDataFunc(d.generateItemDataFuncName)(user.Id, item)
		idList = append(idList, itemData)
	}

	putRowRequest := new(tablestore.PutRowRequest)
	putRowChange := new(tablestore.PutRowChange)
	putRowChange.TableName = d.table
	putPk := new(tablestore.PrimaryKey)
	putPk.AddPrimaryKeyColumn("user_id", uid)
	putPk.AddPrimaryKeyColumnWithAutoIncrement("auto_id")

	putRowChange.PrimaryKey = putPk
	putRowChange.AddColumn("item_ids", strings.Join(idList, ","))
	putRowChange.SetCondition(tablestore.RowExistenceExpectation_IGNORE)
	putRowRequest.PutRowChange = putRowChange
	_, err := d.tablestore.Client.PutRow(putRowRequest)

	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureTableStoreDao\tuid=%s\terr=%v", context.RecommendId, user.Id, err))

	}
	log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tmsg=log history success", context.RecommendId, user.Id))

}
func (d *User2ItemExposureTableStoreDao) FilterByHistory(uid UID, items []*Item) (ret []*Item) {
	getRangeRequest := &tablestore.GetRangeRequest{}
	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = d.table

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumn("user_id", string(uid))
	startPK.AddPrimaryKeyColumnWithMaxValue("auto_id")
	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumn("user_id", string(uid))
	endPK.AddPrimaryKeyColumnWithMinValue("auto_id")
	rangeRowQueryCriteria.StartPrimaryKey = startPK
	rangeRowQueryCriteria.EndPrimaryKey = endPK
	rangeRowQueryCriteria.ColumnsToGet = []string{"item_ids"}
	rangeRowQueryCriteria.Direction = tablestore.BACKWARD
	rangeRowQueryCriteria.MaxVersion = 1
	if d.maxItems > 0 {
		rangeRowQueryCriteria.Limit = d.maxItems
	}
	if d.timeInterval > 0 {
		t := time.Now().Unix() - int64(d.timeInterval)
		trange := tablestore.TimeRange{
			Start: t * 1000,
			End:   (time.Now().Unix() + 60) * 1000,
		}
		rangeRowQueryCriteria.TimeRange = &trange
	}

	getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria

	getRangeResp, err := d.tablestore.Client.GetRange(getRangeRequest)
	if err != nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureTableStoreDao\tuid=%s\terr=%v", uid, err))
		ret = items
		return
	}

	fiterIds := make(map[string]bool)
	for _, row := range getRangeResp.Rows {
		if len(row.Columns) > 0 {
			if id, ok := row.Columns[0].Value.(string); ok && id != "" {
				idList := strings.Split(id, ",")
				for _, id := range idList {
					fiterIds[id] = true
				}
			}
		}
	}

	for _, item := range items {
		itemData := getGenerateItemDataFunc(d.generateItemDataFuncName)(uid, item)
		if _, ok := fiterIds[itemData]; !ok {
			ret = append(ret, item)
		}
	}
	return
}

func (d *User2ItemExposureTableStoreDao) ClearHistory(user *User, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)
	if scene != d.clearLogScene {
		return
	}
	getRangeRequest := &tablestore.GetRangeRequest{}
	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = d.table

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumn("user_id", string(user.Id))
	startPK.AddPrimaryKeyColumnWithMinValue("auto_id")
	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumn("user_id", string(user.Id))
	endPK.AddPrimaryKeyColumnWithMaxValue("auto_id")
	rangeRowQueryCriteria.StartPrimaryKey = startPK
	rangeRowQueryCriteria.EndPrimaryKey = endPK
	rangeRowQueryCriteria.Direction = tablestore.FORWARD
	rangeRowQueryCriteria.MaxVersion = 1

	getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria

	getRangeResp, err := d.tablestore.Client.GetRange(getRangeRequest)

	batchWriteRowRequest := &tablestore.BatchWriteRowRequest{}
	for {
		if err != nil {
			context.LogError(fmt.Sprintf("module=User2ItemExposureTableStoreDao\tuid=%s\terr=%v", user.Id, err))
			break
		}
		for _, row := range getRangeResp.Rows {
			deleteRowChange := new(tablestore.DeleteRowChange)
			deleteRowChange.TableName = d.table
			deleteRowChange.PrimaryKey = row.PrimaryKey
			deleteRowChange.SetCondition(tablestore.RowExistenceExpectation_EXPECT_EXIST)
			batchWriteRowRequest.AddRowChange(deleteRowChange)
		}
		if getRangeResp.NextStartPrimaryKey == nil {
			break
		} else {
			getRangeRequest.RangeRowQueryCriteria.StartPrimaryKey = getRangeResp.NextStartPrimaryKey
			getRangeResp, err = d.tablestore.Client.GetRange(getRangeRequest)
		}
	}

	_, err = d.tablestore.Client.BatchWriteRow(batchWriteRowRequest)
	if err != nil {
		context.LogError(fmt.Sprintf("delete user [%s] exposure items failed with error: %v", user.Id, err))
	} else {
		context.LogInfo(fmt.Sprintf("delete user [%s] exposure items", user.Id))
	}
}
func (d *User2ItemExposureTableStoreDao) GetExposureItemIds(user *User, context *context.RecommendContext) (ret string) {
	uid := string(user.Id)

	getRangeRequest := &tablestore.GetRangeRequest{}
	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = d.table

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumn("user_id", string(uid))
	startPK.AddPrimaryKeyColumnWithMaxValue("auto_id")
	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumn("user_id", string(uid))
	endPK.AddPrimaryKeyColumnWithMinValue("auto_id")
	rangeRowQueryCriteria.StartPrimaryKey = startPK
	rangeRowQueryCriteria.EndPrimaryKey = endPK
	rangeRowQueryCriteria.ColumnsToGet = []string{"item_ids"}
	rangeRowQueryCriteria.Direction = tablestore.BACKWARD
	rangeRowQueryCriteria.MaxVersion = 1
	if d.maxItems > 0 {
		rangeRowQueryCriteria.Limit = d.maxItems
	}
	if d.timeInterval > 0 {
		t := time.Now().Unix() - int64(d.timeInterval)
		trange := tablestore.TimeRange{
			Start: t * 1000,
			End:   (time.Now().Unix() + 60) * 1000,
		}
		rangeRowQueryCriteria.TimeRange = &trange
	}

	getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria

	getRangeResp, err := d.tablestore.Client.GetRange(getRangeRequest)
	if err != nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureTableStoreDao\tuid=%s\terr=%v", uid, err))
		return
	}

	fiterIds := make([]string, 0, 10)
	for _, row := range getRangeResp.Rows {
		if len(row.Columns) > 0 {
			if id, ok := row.Columns[0].Value.(string); ok && id != "" {
				idList := strings.Split(id, ",")
				for _, id := range idList {
					fiterIds = append(fiterIds, id)
				}
			}
		}
	}

	ret = strings.Join(fiterIds, ",")

	return
}
