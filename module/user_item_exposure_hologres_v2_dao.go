package module

import (
	gocontext "context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/huandu/go-sqlbuilder"
)

/** create table ddl

BEGIN;
CREATE TABLE user_exposure_history (
    serial_id serial, --就在这里加一列自增列就行
    uid text NOT NULL,
    item text NOT NULL,
    create_time integer NOT NULL
) WITH (
orientation = 'column',
storage_format = 'orc',
bitmap_columns = 'uid,item',
clustering_key = 'uid:asc,create_time:asc',
dictionary_encoding_columns = 'uid,item',
distribution_key = 'uid',
segment_key = 'create_time',
table_storage_mode = 'any',
time_to_live_in_seconds = '172800'
);
comment on table "user_exposure_history" is '曝光记录表';
COMMIT;
**/

/*
var (

	holo_exposure_insert_sql = "INSERT INTO %s (uid, item, create_time) VALUES($1, $2, $3)"

)
*/
var (
	holo_exposure_select_sql = `SELECT t.item
FROM (
    SELECT serial_id
    FROM %s 
    WHERE uid = $1 
      AND create_time >= $2  
    ORDER BY create_time DESC
    LIMIT $3 
) AS u
JOIN %s AS t
ON t.serial_id = u.serial_id;`
)

type User2ItemExposureHologresV2Dao struct {
	db                       *sql.DB
	table                    string
	maxItems                 int
	timeInterval             int //  second
	mu                       sync.Mutex
	insertStmt               *sql.Stmt
	selectStmt               *sql.Stmt
	generateItemDataFuncName string
	writeLogExcludeScenes    map[string]bool
	clearLogScene            string
	onlyLogUserExposeFlag    bool
	generateUserProgram      *vm.Program
}

func NewUser2ItemExposureHologresV2Dao(config recconf.FilterConfig) *User2ItemExposureHologresV2Dao {
	dao := &User2ItemExposureHologresV2Dao{
		maxItems:                 -1,
		timeInterval:             -1,
		generateItemDataFuncName: config.GenerateItemDataFuncName,
		writeLogExcludeScenes:    make(map[string]bool),
		clearLogScene:            config.ClearLogIfNotEnoughScene,
		onlyLogUserExposeFlag:    config.OnlyLogUserExposeFlag,
	}
	hologres, err := holo.GetPostgres(config.DaoConf.HologresName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.db = hologres.DB
	dao.table = config.DaoConf.HologresTableName
	if config.MaxItems > 0 {
		dao.maxItems = config.MaxItems
	}

	if config.TimeInterval > 0 {
		dao.timeInterval = config.TimeInterval
	} else {
		dao.timeInterval = 86400 * 90 // default 90 days for time interval
	}

	for _, scene := range config.WriteLogExcludeScenes {
		dao.writeLogExcludeScenes[scene] = true
	}
	if config.GenerateUserDataExpr != "" {
		if p, err := expr.Compile(config.GenerateUserDataExpr); err != nil {
			panic(err)
		} else {
			dao.generateUserProgram = p
		}
	}
	return dao
}

func (d *User2ItemExposureHologresV2Dao) LogHistory(user *User, items []*Item, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)
	if _, exist := d.writeLogExcludeScenes[scene]; exist {
		return
	}

	if len(items) == 0 {
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureHologresV2Dao\terr=items empty", context.RecommendId))
		return
	}

	uid := string(user.Id)
	userData, err := d.getUserData(user.Id, context)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureHologresV2Dao\tuid=%s\terr=%v", context.RecommendId, uid, err))
		return
	}
	if d.insertStmt == nil {
		d.mu.Lock()
		if d.insertStmt == nil {
			stmt, err := d.db.Prepare(fmt.Sprintf(holo_exposure_insert_sql, d.table))
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureHologresV2Dao\tuid=%s\terr=%v", context.RecommendId, uid, err))
				d.mu.Unlock()
				return
			}
			d.insertStmt = stmt
			d.mu.Unlock()
		} else {
			d.mu.Unlock()
		}
	}

	createTime := time.Now().Unix()
	var ret string
	for _, item := range items {
		itemData := getGenerateItemDataFunc(d.generateItemDataFuncName)(user.Id, item)
		ret = ret + "," + itemData
	}
	ret = ret[1:]
	_, err = d.insertStmt.Exec(userData, ret, createTime)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureHologresV2Dao\tuid=%s\terr=%v", context.RecommendId, user.Id, err))
		return
	}

	log.Info(fmt.Sprintf("requestId=%s\tscene=%s\tuid=%s\tmsg=log history success", context.RecommendId, scene, user.Id))

}
func (d *User2ItemExposureHologresV2Dao) FilterByHistory(uid UID, items []*Item, context *context.RecommendContext) (ret []*Item) {
	userData, err := d.getUserData(uid, context)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureHologresV2Dao\tuid=%s\terr=%v", context.RecommendId, uid, err))
		ret = items
		return
	}
	var (
		timeInterval int64
		maxItems     int
	)

	if d.timeInterval > 0 {
		timeInterval = time.Now().Unix() - int64(d.timeInterval)
	}

	if d.maxItems > 0 {
		maxItems = d.maxItems
	}

	if d.selectStmt == nil {
		d.mu.Lock()
		if d.selectStmt == nil {
			stmt, err := d.db.Prepare(fmt.Sprintf(holo_exposure_select_sql, d.table, d.table))
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureHologresV2Dao\tuid=%s\terr=%v", context.RecommendId, uid, err))
				ret = items
				d.mu.Unlock()
				return
			}
			d.selectStmt = stmt
			d.mu.Unlock()
		} else {
			d.mu.Unlock()
		}
	}

	ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 100*time.Millisecond)
	defer cancel()
	rows, err := d.selectStmt.QueryContext(ctx, userData, timeInterval, maxItems)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureHologresV2Dao\tuid=%s\terr=%v", context.RecommendId, uid, err))
		ret = items
		return
	}
	defer rows.Close()

	fiterIds := make(map[string]bool)
	for rows.Next() {
		var itemDatas string
		if err := rows.Scan(&itemDatas); err == nil {
			ids := strings.Split(itemDatas, ",")
			for _, id := range ids {
				fiterIds[id] = true
			}

		}
	}
	if d.onlyLogUserExposeFlag {
		for _, item := range items {
			itemData := getGenerateItemDataFunc(d.generateItemDataFuncName)(uid, item)
			if _, ok := fiterIds[itemData]; ok {
				item.AddProperty("_is_exposure_", 1)
			}
		}

		ret = items
	} else {
		for _, item := range items {
			itemData := getGenerateItemDataFunc(d.generateItemDataFuncName)(uid, item)
			if _, ok := fiterIds[itemData]; !ok {
				ret = append(ret, item)
			}
		}
	}

	return
}

func (d *User2ItemExposureHologresV2Dao) ClearHistory(user *User, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)
	if scene != d.clearLogScene {
		return
	}
	userData, err := d.getUserData(user.Id, context)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureHologresV2Dao\tuid=%s\terr=%v", context.RecommendId, userData, err))
		return
	}
	db := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	db.DeleteFrom(d.table)
	db.Where(db.Equal("uid", userData))

	deleteSql, args := db.Build()
	_, err = d.db.Exec(deleteSql, args...)
	if err != nil {
		context.LogError(fmt.Sprintf("delete user [%s] exposure items from holo failed, err=%v", user.Id, err))
	}
}

func (d *User2ItemExposureHologresV2Dao) GetExposureItemIds(user *User, context *context.RecommendContext) (ret string) {
	uid := string(user.Id)
	userData, err := d.getUserData(user.Id, context)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureHologresV2Dao\tuid=%s\terr=%v", context.RecommendId, uid, err))
		return
	}
	var (
		timeInterval int64
		maxItems     int
	)

	if d.timeInterval > 0 {
		timeInterval = time.Now().Unix() - int64(d.timeInterval)
	}

	if d.maxItems > 0 {
		maxItems = d.maxItems
	}
	if d.selectStmt == nil {
		d.mu.Lock()
		if d.selectStmt == nil {
			stmt, err := d.db.Prepare(fmt.Sprintf(holo_exposure_select_sql, d.table, d.table))
			if err != nil {
				log.Error(fmt.Sprintf("module=User2ItemExposureHologresV2Dao\tuid=%s\terr=%v", uid, err))
				d.mu.Unlock()
				return
			}
			d.selectStmt = stmt
			d.mu.Unlock()
		} else {
			d.mu.Unlock()
		}
	}

	ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 100*time.Millisecond)
	defer cancel()
	rows, err := d.selectStmt.QueryContext(ctx, userData, timeInterval, maxItems)
	if err != nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureHologresV2Dao\tuid=%s\terr=%v", uid, err))
		return
	}
	defer rows.Close()

	fiterIds := make([]string, 0, 10)
	for rows.Next() {
		var itemDatas string
		if err := rows.Scan(&itemDatas); err == nil {
			ids := strings.Split(itemDatas, ",")
			fiterIds = append(fiterIds, ids...)
		}
	}

	ret = strings.Join(fiterIds, ",")
	return
}
func (d *User2ItemExposureHologresV2Dao) getUserData(uid UID, context *context.RecommendContext) (string, error) {
	userData := string(uid)
	if d.generateUserProgram != nil {
		m := map[string]any{
			"uid": string(uid),
			"context": map[string]any{
				"item_id":  utils.ToString(context.GetParameter("item_id"), ""),
				"features": context.GetParameter("features"),
			},
			"sprintf": fmt.Sprintf,
		}
		if output, err := expr.Run(d.generateUserProgram, m); err != nil {
			return "", err
		} else {
			if str := utils.ToString(output, ""); str != "" {
				userData = str
			} else {
				return "", fmt.Errorf("output error(%v)", output)
			}

		}
	}

	return userData, nil

}
