package module

import (
	gocontext "context"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/alibaba/pairec/v2/datasource/hbase_thrift"
	"github.com/alibaba/pairec/v2/datasource/hbase_thrift/gen-go/hbase"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type ItemStateFilterHBaseThriftDao struct {
	hBaseName     string
	table         string
	itemFieldName string
	selectFields  string
	columnFamily  string
	filterParam   *FilterParam
	columns       []*hbase.TColumn
}

func NewItemStateFilterHBaseThriftDao(config recconf.FilterConfig) *ItemStateFilterHBaseThriftDao {
	dao := &ItemStateFilterHBaseThriftDao{
		hBaseName:     config.ItemStateDaoConf.HBaseName,
		table:         config.ItemStateDaoConf.HBaseTable,
		columnFamily:  config.ItemStateDaoConf.ColumnFamily,
		itemFieldName: config.ItemStateDaoConf.ItemFieldName,
		selectFields:  config.ItemStateDaoConf.SelectFields,
	}
	if len(config.FilterParams) > 0 {
		dao.filterParam = NewFilterParamWithConfig(config.FilterParams)
	}

	if dao.columnFamily != "" && dao.selectFields != "*" {
		fields := strings.Split(dao.selectFields, ",")
		for _, field := range fields {
			column := &hbase.TColumn{
				Family:    []byte(dao.columnFamily),
				Qualifier: []byte(field),
			}
			dao.columns = append(dao.columns, column)
		}
	}

	return dao
}

func (d *ItemStateFilterHBaseThriftDao) Filter(user *User, items []*Item) (ret []*Item) {
	requestCount := 500
	fields := make(map[string]bool, len(items))
	cpuCount := utils.MaxInt(int(math.Ceil(float64(len(items))/float64(requestCount))), 1)
	requestCh := make(chan []*Item, cpuCount)
	defer close(requestCh)

	if cpuCount == 1 {
		requestCh <- items
	} else {
		maps := make(map[int][]*Item)
		index := 0
		for i, item := range items {
			maps[index%cpuCount] = append(maps[index%cpuCount], item)
			if (i+1)%requestCount == 0 {
				index++
			}
		}

		for _, itemlist := range maps {
			requestCh <- itemlist
		}

	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	mergeFunc := func(maps map[string]bool) {
		mu.Lock()
		for k, v := range maps {
			if v {
				fields[k] = v
			}
		}
		mu.Unlock()
	}

	userProperties := user.MakeUserFeatures2()
	for i := 0; i < cpuCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case itemlist := <-requestCh:
				fieldMap := make(map[string]bool, len(itemlist))
				key2Item := make(map[string]*Item, len(itemlist))
				var keys []string
				for _, item := range itemlist {
					key := string(item.Id)
					keys = append(keys, key)
					key2Item[key] = item
					fieldMap[string(item.Id)] = true
				}

				var gets []*hbase.TGet
				for _, rowKey := range keys {
					gets = append(gets, &hbase.TGet{Row: []byte(rowKey), Columns: d.columns})
				}
				defaultCtx := gocontext.Background()
				client, _ := hbase_thrift.GetHBaseThrift(d.hBaseName)
				result, err := client.Client.GetMultiple(defaultCtx, []byte(d.table), gets)

				if err != nil {
					log.Error(fmt.Sprintf("module=ItemStateFilterHBaseThriftDao\terror=%v", err))
					// if error , not filter item
					for _, item := range itemlist {
						fieldMap[string(item.Id)] = true
					}
					mergeFunc(fieldMap)
					return
				}
				defer hbase_thrift.PutHBaseThrift(d.hBaseName, client)

				for _, row := range result {
					var item *Item
					properties := make(map[string]interface{}, len(row.ColumnValues))
					item = key2Item[string(row.Row)]
					for _, cell := range row.ColumnValues {
						properties[string(cell.Qualifier)] = string(cell.Value)
					}
					if item != nil {
						if d.filterParam != nil {
							result, err := d.filterParam.EvaluateByDomain(userProperties, properties)
							if err == nil && !result {
								fieldMap[string(item.Id)] = false
							}
						} else {
							fieldMap[string(item.Id)] = true
						}

					}
					if nil != item && len(properties) > 0 && fieldMap[string(item.Id)] {
						item.AddProperties(properties)
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
