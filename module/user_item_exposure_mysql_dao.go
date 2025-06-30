package module

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/mysqldb"
	"github.com/alibaba/pairec/v2/recconf"
)

type User2ItemExposureMysqlDao struct {
	db            *sql.DB
	table         string
	configs       map[string]interface{}
	hashValue     int
	tablePrefix   string
	clearLogScene string
}

const (
	SQL_Filter_History = "SELECT article_id FROM %s WHERE uid=? "
	SQL_Insert_History = "INSERT INTO %s (`article_id`, `uid`, `add_time`) VALUES ('%s', '%s', %d)"
	SQL_Delete_History = "DELETE FROM %s WHERE `uid`= '%s';"
)

func NewUser2ItemExposureMysqlDao(config recconf.FilterConfig) *User2ItemExposureMysqlDao {
	dao := &User2ItemExposureMysqlDao{
		clearLogScene: config.ClearLogIfNotEnoughScene,
	}
	mysql, err := mysqldb.GetMysql(config.DaoConf.MysqlName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.db = mysql.DB
	dao.table = config.DaoConf.MysqlTable
	if len(config.DaoConf.Config) > 0 {
		if err := json.Unmarshal([]byte(config.DaoConf.Config), &dao.configs); err != nil {
			log.Error(fmt.Sprintf("error=%v", err))
			return nil
		}
		if _, ok := dao.configs["table_prefix"]; ok {
			dao.tablePrefix = dao.configs["table_prefix"].(string)
		}
		if _, ok := dao.configs["hash_value"]; ok {
			dao.hashValue = int(dao.configs["hash_value"].(float64))
		}
	}
	return dao
}

func (d *User2ItemExposureMysqlDao) getTableName(uid string) string {
	if d.tablePrefix == "" {
		return d.table
	} else {
		id, _ := strconv.Atoi(uid)
		return d.tablePrefix + strconv.Itoa(id%d.hashValue)
	}
}
func (d *User2ItemExposureMysqlDao) LogHistory(user *User, items []*Item, context *context.RecommendContext) {
	addTime := time.Now().Unix()
	uid := string(user.Id)
	var sql string
	for i := 0; i < len(items); i++ {
		articleId := string(items[i].Id)
		sql += fmt.Sprintf(SQL_Insert_History, d.getTableName(uid), articleId, uid, addTime) + ";"
	}
	if sql == "" {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureMysqlDao\terror=insert sql empty", context.RecommendId))
		return
	}

	_, err := d.db.Exec(sql)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureMysqlDao\tuid=%s\terr=%v", context.RecommendId, user.Id, err))
		// if batch sql error, split sql exec
		for i := 0; i < len(items); i++ {
			articleId := string(items[i].Id)
			sql = fmt.Sprintf(SQL_Insert_History, d.table, articleId, uid, addTime)
			d.db.Exec(sql)
		}
		return
	}
	log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tmsg=log history success", context.RecommendId, user.Id))

}
func (d *User2ItemExposureMysqlDao) FilterByHistory(uid UID, items []*Item, context *context.RecommendContext) (ret []*Item) {
	query := fmt.Sprintf(SQL_Filter_History, d.getTableName(string(uid)))
	stmt, err := d.db.Prepare(query)
	if err != nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureMysqlDao\tuid=%s\terr=%v", uid, err))
		ret = items
		return
	}

	defer stmt.Close()
	rows, err := stmt.Query(string(uid))
	if err != nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureMysqlDao\tuid=%s\terr=%v", uid, err))
		ret = items
		return
	}
	fiterIds := make(map[string]bool)
	defer rows.Close()
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err == nil {
			fiterIds[id] = true
		}
	}

	for _, item := range items {
		if _, ok := fiterIds[string(item.Id)]; !ok {
			ret = append(ret, item)
		}
	}
	return
}

func (d *User2ItemExposureMysqlDao) ClearHistory(user *User, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)
	if scene != d.clearLogScene {
		return
	}
	uid := string(user.Id)
	sql := fmt.Sprintf(SQL_Delete_History, d.getTableName(uid), uid)
	_, err := d.db.Exec(sql)
	if err != nil {
		context.LogError(fmt.Sprintf("delete user [%s] exposure items failed with err: %v", user.Id, err))
	}
}
func (d *User2ItemExposureMysqlDao) GetExposureItemIds(user *User, context *context.RecommendContext) (ret string) {
	query := fmt.Sprintf(SQL_Filter_History, d.getTableName(string(user.Id)))
	stmt, err := d.db.Prepare(query)
	if err != nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureMysqlDao\tuid=%s\terr=%v", user.Id, err))
		return
	}

	defer stmt.Close()
	rows, err := stmt.Query(string(user.Id))
	if err != nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureMysqlDao\tuid=%s\terr=%v", user.Id, err))
		return
	}

	fiterIds := make([]string, 0, 10)
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			fiterIds = append(fiterIds, id)
		}
	}

	ret = strings.Join(fiterIds, ",")

	return
}
