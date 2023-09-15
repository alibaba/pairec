package module

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/persist/mysqldb"
	"github.com/alibaba/pairec/recconf"
)

type UserCustomRecallMysqlDao struct {
	db          *sql.DB
	itemType    string
	recallName  string
	table       string
	recallCount int
}

var (
	sql_select = "SELECT item_ids FROM %s WHERE user_id='%s'"
)

func NewUserCusteomRecallMysqlDao(config recconf.RecallConfig) *UserCustomRecallMysqlDao {
	dao := &UserCustomRecallMysqlDao{
		recallCount: 1000,
	}
	mysql, err := mysqldb.GetMysql(config.DaoConf.MysqlName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.db = mysql.DB
	dao.table = config.DaoConf.MysqlTable
	dao.itemType = config.ItemType
	dao.recallName = config.Name
	return dao
}

func (d *UserCustomRecallMysqlDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	uid := string(user.Id)
	sql := fmt.Sprintf(sql_select, d.table, uid)
	rows, err := d.db.Query(sql)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCustomRecallMysqlDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
		return
	}
	itemIds := make([]string, 0, d.recallCount)
	for rows.Next() {
		var ids string
		if err := rows.Scan(&ids); err == nil {
			idList := strings.Split(ids, ",")
			for _, id := range idList {
				if len(id) > 0 {
					itemIds = append(itemIds, id)
				}
			}
		}
	}
	rows.Close()

	if len(itemIds) == 0 {
		log.Info(fmt.Sprintf("module=UserCustomRecallMysqlDao\tuid=%s\terr=item ids empty", uid))
		return
	}

	if len(itemIds) > d.recallCount {
		rand.Shuffle(len(itemIds)/2, func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})

		itemIds = itemIds[:d.recallCount]
	}

	for _, id := range itemIds {
		item := &Item{
			Id:         ItemId(id),
			ItemType:   d.itemType,
			RetrieveId: d.recallName,
			Properties: make(map[string]interface{}),
		}

		ret = append(ret, item)
	}

	return
}
