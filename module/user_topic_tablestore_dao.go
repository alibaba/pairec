package module

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore/search"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/tablestoredb"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type UserTopicTableStoreDao struct {
	tablestore     *tablestoredb.TableStore
	userTopicTable string
	topicItemTable string
	//configs        map[string]interface{}
	itemType    string
	recallName  string
	recallCount int
	topicNum    int
	indexName   string
}

func NewUserTopicTableStoreDao(config recconf.RecallConfig) *UserTopicTableStoreDao {
	dao := &UserTopicTableStoreDao{
		itemType:       config.ItemType,
		recallName:     config.Name,
		recallCount:    config.RecallCount,
		topicNum:       5,
		userTopicTable: config.UserTopicDaoConf.UserTopicTable,
		topicItemTable: config.UserTopicDaoConf.TopicItemTable,
		indexName:      config.UserTopicDaoConf.IndexName,
	}
	tablestore, err := tablestoredb.GetTableStore(config.UserTopicDaoConf.TableStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.tablestore = tablestore
	return dao
}

func (d *UserTopicTableStoreDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	// uid := string(user.Id)
	getRowRequest := new(tablestore.GetRowRequest)
	criteria := new(tablestore.SingleRowQueryCriteria)
	putPk := new(tablestore.PrimaryKey)
	putPk.AddPrimaryKeyColumn("user_id", string(user.Id))

	criteria.PrimaryKey = putPk
	criteria.ColumnsToGet = []string{"favorite_topic"}
	getRowRequest.SingleRowQueryCriteria = criteria
	getRowRequest.SingleRowQueryCriteria.TableName = d.userTopicTable
	getRowRequest.SingleRowQueryCriteria.MaxVersion = 1
	getResp, err := d.tablestore.Client.GetRow(getRowRequest)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserTopicTableStoreDao\terror=%v", context.RecommendId, err))
		return
	}

	if len(getResp.Columns) == 0 {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=UserTopicTableStoreDao\tuid=%s\terr=topic empty", context.RecommendId, user.Id))
		return
	}

	topics := make([]*topicInfo, 0, d.topicNum)

	if str, ok := getResp.Columns[0].Value.(string); ok {
		idList := strings.Split(str, ",")
		for _, id := range idList {
			tv := strings.Split(id, ":")
			if len(tv) == 2 {
				f, err := strconv.ParseFloat(tv[1], 64)
				if err != nil || f == float64(0) {
					f = float64(0.5)
				}

				info := topicInfo{
					topic: tv[0],
					value: f,
				}
				topics = append(topics, &info)
			} else if len(tv) == 1 {
				info := topicInfo{
					topic: tv[0],
					value: float64(0.5),
				}
				topics = append(topics, &info)
			}
		}
	}

	if len(topics) == 0 {
		log.Info(fmt.Sprintf("module=UserTopicTableStoreDao\tuid=%s\terr=topic empty", user.Id))
		return
	}

	if len(topics) > d.topicNum {
		rand.Shuffle(len(topics), func(i, j int) {
			topics[i], topics[j] = topics[j], topics[i]
		})

		topics = topics[:d.topicNum]
	}
	totalValue := float64(0)
	for _, topic := range topics {
		totalValue += topic.value
	}

	var wg sync.WaitGroup
	for _, topic := range topics {
		topic.count = int((topic.value / totalValue) * float64(d.recallCount))
		wg.Add(1)
		if d.indexName == "" {
			go func(info *topicInfo) ([]*Item, error) {
				defer wg.Done()
				result := make([]*Item, 0)

				getRangeRequest := &tablestore.GetRangeRequest{}

				rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
				rangeRowQueryCriteria.TableName = d.topicItemTable

				startPK := new(tablestore.PrimaryKey)
				startPK.AddPrimaryKeyColumn("topic", info.topic)
				startPK.AddPrimaryKeyColumnWithMinValue("item_id")
				endPK := new(tablestore.PrimaryKey)
				endPK.AddPrimaryKeyColumn("topic", info.topic)
				endPK.AddPrimaryKeyColumnWithMaxValue("item_id")
				rangeRowQueryCriteria.StartPrimaryKey = startPK
				rangeRowQueryCriteria.EndPrimaryKey = endPK
				rangeRowQueryCriteria.ColumnsToGet = []string{"value"}
				rangeRowQueryCriteria.Direction = tablestore.FORWARD
				rangeRowQueryCriteria.MaxVersion = 1
				rangeRowQueryCriteria.Limit = int32(info.count)
				getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria

				getRangeResp, err := d.tablestore.Client.GetRange(getRangeRequest)
				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=UserTopicTableStoreDao\terror=%v", context.RecommendId, err))
					return result, err
				}
				for _, row := range getRangeResp.Rows {
					item := NewItem(row.PrimaryKey.PrimaryKeys[1].Value.(string))
					item.ItemType = d.itemType
					item.RetrieveId = d.recallName
					if len(row.Columns) == 1 {
						item.Score = utils.ToFloat(row.Columns[0].Value, 0)
					}
					result = append(result, item)
				}

				ret = append(ret, result...)
				return result, nil
			}(topic)

		} else {
			go func(info *topicInfo) {
				defer wg.Done()
				result := make([]*Item, 0)

				searchRequest := &tablestore.SearchRequest{}

				searchRequest.TableName = d.topicItemTable
				searchRequest.IndexName = d.indexName
				rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
				rangeRowQueryCriteria.TableName = d.topicItemTable

				pk := new(tablestore.PrimaryKey)
				pk.AddPrimaryKeyColumn("topic", info.topic)

				searchRequest.RoutingValues = append(searchRequest.RoutingValues, pk)
				searchQuery := search.NewSearchQuery()
				searchQuery.SetSort(&search.Sort{
					Sorters: []search.Sorter{
						&search.FieldSort{
							FieldName: "value",
							Order:     search.SortOrder_DESC.Enum(),
						},
					},
				})
				searchQuery.SetLimit(int32(info.count))
				searchRequest.SetSearchQuery(searchQuery)
				searchRequest.ColumnsToGet = &tablestore.ColumnsToGet{Columns: []string{"value"}}

				searchResp, err := d.tablestore.Client.Search(searchRequest)
				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=UserTopicTableStoreDao\terror=%v", context.RecommendId, err))
					return
				}
				for _, row := range searchResp.Rows {
					item := NewItem(row.PrimaryKey.PrimaryKeys[1].Value.(string))
					item.ItemType = d.itemType
					item.RetrieveId = d.recallName
					if len(row.Columns) == 1 {
						item.Score = utils.ToFloat(row.Columns[0].Value, 0)
					}
					result = append(result, item)
				}

				ret = append(ret, result...)
			}(topic)

		}
	}

	wg.Wait()
	return
}
