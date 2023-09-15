package module

import (
	gocontext "context"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/datasource/hbase_thrift"
	"github.com/alibaba/pairec/datasource/hbase_thrift/gen-go/hbase"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
)

type FeatureHBaseThriftDao struct {
	*FeatureBaseDao

	table            string
	columnFamily     string
	userSelectFields string
	userColumns      []*hbase.TColumn
	itemSelectFields string
	itemColumns      []*hbase.TColumn
	hBaseName        string
}

func NewFeatureHBaseThriftDao(config recconf.FeatureDaoConfig) *FeatureHBaseThriftDao {
	dao := &FeatureHBaseThriftDao{
		FeatureBaseDao:   NewFeatureBaseDao(&config),
		table:            config.HBaseTable,
		columnFamily:     config.ColumnFamily,
		userSelectFields: config.UserSelectFields,
		itemSelectFields: config.ItemSelectFields,
		hBaseName:        config.HBaseName,
	}

	if dao.columnFamily != "" && dao.userSelectFields != "*" {
		fields := strings.Split(dao.userSelectFields, ",")
		for _, field := range fields {
			column := &hbase.TColumn{
				Family:    []byte(dao.columnFamily),
				Qualifier: []byte(field),
			}
			dao.userColumns = append(dao.userColumns, column)
		}
	}

	if dao.columnFamily != "" && dao.itemSelectFields != "*" {
		fields := strings.Split(dao.itemSelectFields, ",")
		for _, field := range fields {
			column := &hbase.TColumn{
				Family:    []byte(dao.columnFamily),
				Qualifier: []byte(field),
			}
			dao.itemColumns = append(dao.itemColumns, column)
		}
	}

	return dao
}

func (d *FeatureHBaseThriftDao) FeatureFetch(user *User, items []*Item, context *context.RecommendContext) {
	if d.featureStore == Feature_Store_User {
		d.userFeatureFetch(user, context)
	} else {
		d.itemsFeatureFetch(items, context)
	}
}

func (d *FeatureHBaseThriftDao) userFeatureFetch(user *User, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHBaseThriftDao\terror=%v", context.RecommendId, err))
			return
		}
	}()

	comms := strings.Split(d.featureKey, ":")
	if len(comms) < 2 {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\terror=featureKey error(%s)", context.RecommendId, user.Id, d.featureKey))
		return
	}

	key := user.StringProperty(comms[1])
	if key == "" {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHBaseThriftDao\terror=property not found(%s)", context.RecommendId, comms[1]))
		return
	}

	defaultCtx := gocontext.Background()
	client, _ := hbase_thrift.GetHBaseThrift(d.hBaseName)
	result, err := client.Client.Get(defaultCtx, []byte(d.table), &hbase.TGet{Row: []byte(key), Columns: d.userColumns})

	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHBaseThriftDao\terror=%v", context.RecommendId, err))
		return
	}
	defer hbase_thrift.PutHBaseThrift(d.hBaseName, client)

	properties := make(map[string]interface{}, len(result.ColumnValues))
	for _, cell := range result.ColumnValues {
		properties[string(cell.Qualifier)] = string(cell.Value)
	}

	if d.cacheFeaturesName != "" {
		user.AddCacheFeatures(d.cacheFeaturesName, properties)
	} else {
		user.AddProperties(properties)
	}
}

func (d *FeatureHBaseThriftDao) itemsFeatureFetch(items []*Item, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHBaseThriftDao\terror=%v", context.RecommendId, err))
			return
		}
	}()

	fk := d.featureKey
	if fk != "item:id" {
		comms := strings.Split(d.featureKey, ":")
		if len(comms) < 2 {
			log.Error(fmt.Sprintf("requestId=%s\tevent=itemsFeatureFetch\terror=featureKey error(%s)", context.RecommendId, d.featureKey))
			return
		}

		fk = comms[1]
	}

	requestCount := 100
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
	for i := 0; i < cpuCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case itemlist := <-requestCh:
				var keys []string
				key2Item := make(map[string]*Item, len(itemlist))
				for _, item := range itemlist {
					var key string
					if fk == "item:id" {
						key = string(item.Id)
					} else {
						key = item.StringProperty(fk)
					}
					keys = append(keys, key)
					key2Item[key] = item
				}

				var gets []*hbase.TGet
				for _, rowKey := range keys {
					gets = append(gets, &hbase.TGet{Row: []byte(rowKey), Columns: d.itemColumns})
				}
				defaultCtx := gocontext.Background()
				client, _ := hbase_thrift.GetHBaseThrift(d.hBaseName)
				result, err := client.Client.GetMultiple(defaultCtx, []byte(d.table), gets)

				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHBaseThriftDao\terror=%v", context.RecommendId, err))
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
					if nil != item && len(properties) > 0 {
						item.AddProperties(properties)
					}

				}
			default:
			}
		}()
	}
	wg.Wait()
}
