package module

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/mysqldb"
	"github.com/alibaba/pairec/v2/recconf"
)

type UserTopicMysqlDao struct {
	db             *sql.DB
	userTopicTable string
	topicItemTable string
	//configs        map[string]interface{}
	itemType   string
	recallName string
	total      int
	topicNum   int
}

const (
	SQL_User_Topic_Filter = "SELECT favorite_topic FROM %s WHERE user_id='%s'"
	SQL_Topic_Item_Filter = "SELECT article_ids FROM %s WHERE topic='%s' ORDER BY publish_time DESC, value DESC LIMIT %d"
	//User_Topic_Table      = "pai_recommend_recall_v2_user_topic_list_result"
	//Topic_Item_Table      = "pai_recommend_recall_v2_topic_article_list"
)

func NewUserTopicMysqlDao(config recconf.RecallConfig) *UserTopicMysqlDao {
	dao := &UserTopicMysqlDao{
		itemType:       config.ItemType,
		recallName:     config.Name,
		total:          config.RecallCount,
		topicNum:       10,
		userTopicTable: config.UserTopicDaoConf.UserTopicTable,
		topicItemTable: config.UserTopicDaoConf.TopicItemTable,
	}
	mysql, err := mysqldb.GetMysql(config.UserTopicDaoConf.MysqlName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.db = mysql.DB
	return dao
}

type topicInfo struct {
	topic string
	value float64
	count int
}

func (d *UserTopicMysqlDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	uid := string(user.Id)
	sql := fmt.Sprintf(SQL_User_Topic_Filter, d.userTopicTable, uid)
	rows, err := d.db.Query(sql)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserTopicMysqlDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
		return
	}
	topics := make([]*topicInfo, 0)
	for rows.Next() {
		var comm string
		if err := rows.Scan(&comm); err == nil {
			idList := strings.Split(comm, ",")
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
				}
			}
		}
	}
	rows.Close()

	if len(topics) == 0 {
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
	for _, topic := range topics {
		topic.count = int((topic.value / totalValue) * float64(d.total))
		result, err := func(info *topicInfo) ([]*Item, error) {
			result := make([]*Item, 0)
			sql := fmt.Sprintf(SQL_Topic_Item_Filter, d.topicItemTable, info.topic, info.count)
			rows, err := d.db.Query(sql)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=UserTopicMysqlDao\tsql=%s\terror=%v", context.RecommendId, sql, err))
				return result, err
			}
			for rows.Next() {
				var ids string
				if err := rows.Scan(&ids); err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=UserTopicMysqlDao\terror=%v", context.RecommendId, err))
					continue
				}

				item := NewItem(ids)
				item.ItemType = d.itemType
				item.RetrieveId = d.recallName
				result = append(result, item)
			}
			rows.Close()

			return result, nil
		}(topic)

		if err == nil {
			ret = append(ret, result...)
		}
	}

	return
}
