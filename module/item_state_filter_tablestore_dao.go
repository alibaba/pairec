package module

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/tablestoredb"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type ItemStateFilterTablestoreDao struct {
	tablestore    *tablestoredb.TableStore
	table         string
	whereClause   string
	itemFieldName string
	selectFields  string
	filterParam   *FilterParam
	mu            sync.RWMutex
}

func NewItemStateFilterTablestoreDao(config recconf.FilterConfig) *ItemStateFilterTablestoreDao {
	tablestore, err := tablestoredb.GetTableStore(config.ItemStateDaoConf.TableStoreName)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	dao := &ItemStateFilterTablestoreDao{
		tablestore:    tablestore,
		table:         config.ItemStateDaoConf.TableStoreTableName,
		itemFieldName: config.ItemStateDaoConf.ItemFieldName,
		selectFields:  config.ItemStateDaoConf.SelectFields,
	}
	if len(config.FilterParams) > 0 {
		dao.filterParam = NewFilterParamWithConfig(config.FilterParams)
	}

	return dao
}

func (d *ItemStateFilterTablestoreDao) Filter(user *User, items []*Item) (ret []*Item) {
	requestCount := 100
	fields := make(map[string]bool, len(items))
	cpuCount := utils.MaxInt(int(math.Ceil(float64(len(items))/float64(requestCount))), 1)
	requestCh := make(chan []interface{}, cpuCount)
	maps := make(map[int][]interface{}, cpuCount)

	index := 0
	for i, item := range items {
		maps[index%cpuCount] = append(maps[index%cpuCount], string(item.Id))
		if (i+1)%requestCount == 0 {
			index++
		}
	}

	defer close(requestCh)
	for _, idlist := range maps {
		requestCh <- idlist
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	mergeFunc := func(maps map[string]bool) {
		mu.Lock()
		for k, v := range maps {
			fields[k] = v
		}
		mu.Unlock()
	}
	for i := 0; i < cpuCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case idlist := <-requestCh:
				fieldMap := make(map[string]bool, len(idlist))
				batchGetRowRequest := new(tablestore.BatchGetRowRequest)
				multiRowQueryCriteria := new(tablestore.MultiRowQueryCriteria)
				if d.selectFields != "" {
					multiRowQueryCriteria.ColumnsToGet = strings.Split(d.selectFields, ",")
				}
				multiRowQueryCriteria.TableName = d.table
				multiRowQueryCriteria.MaxVersion = 1

				for _, id := range idlist {
					putPk := new(tablestore.PrimaryKey)
					putPk.AddPrimaryKeyColumn(d.itemFieldName, id)
					multiRowQueryCriteria.AddRow(putPk)
				}
				batchGetRowRequest.MultiRowQueryCriteria = []*tablestore.MultiRowQueryCriteria{multiRowQueryCriteria}

				batchGetResp, err := d.tablestore.Client.BatchGetRow(batchGetRowRequest)
				if err != nil {
					log.Error(fmt.Sprintf("module=ItemStateFilterTablestoreDao\terror=tablestore error(%v)", err))
					for _, id := range idlist {
						fieldMap[id.(string)] = true
					}
					mergeFunc(fieldMap)
					return
				}

				for _, rows := range batchGetResp.TableToRowsResult {
					for _, row := range rows {
						if row.IsSucceed {
							if row.PrimaryKey.PrimaryKeys != nil {
								id := utils.ToString(row.PrimaryKey.PrimaryKeys[0].Value, "")
								if id != "" {
									properties := make(map[string]interface{}, len(row.Columns))
									for _, column := range row.Columns {
										properties[column.ColumnName] = column.Value
									}
									if d.filterParam != nil {
										result, err := d.filterParam.Evaluate(properties)
										if err == nil && result == true {
											fieldMap[id] = true
										}
									} else {
										fieldMap[id] = true
									}

								}

							}

						}
					}
				}
				mergeFunc(fieldMap)
			default:
			}
		}()
	}

	wg.Wait()

	for _, item := range items {
		if _, ok := fields[string(item.Id)]; ok {
			ret = append(ret, item)
		}
	}
	return
}
