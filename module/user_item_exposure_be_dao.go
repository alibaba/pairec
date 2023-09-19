package module

import (
	"fmt"
	"strconv"
	"time"

	be "github.com/aliyun/aliyun-be-go-sdk"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/beengine"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
)

type User2ItemExposureBeDao struct {
	beClient                 *be.Client
	table                    string
	userIdName               string
	itemIdName               string
	generateItemDataFuncName string
	writeLogExcludeScenes    map[string]bool
	clearLogScene            string
}

func NewUser2ItemExposureBeDao(config recconf.FilterConfig) *User2ItemExposureBeDao {
	dao := &User2ItemExposureBeDao{
		generateItemDataFuncName: config.GenerateItemDataFuncName,
		writeLogExcludeScenes:    make(map[string]bool),
		userIdName:               config.DaoConf.BeExposureUserIdName,
		itemIdName:               config.DaoConf.BeExposureItemIdName,
		clearLogScene:            config.ClearLogIfNotEnoughScene,
	}
	client, err := beengine.GetBeClient(config.DaoConf.BeName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.beClient = client.BeClient
	dao.table = config.DaoConf.BeTableName

	for _, scene := range config.WriteLogExcludeScenes {
		dao.writeLogExcludeScenes[scene] = true
	}
	return dao
}

func (d *User2ItemExposureBeDao) LogHistory(user *User, items []*Item, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)
	if _, exist := d.writeLogExcludeScenes[scene]; exist {
		return
	}

	if len(items) == 0 {
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureBeDao\terr=items empty", context.RecommendId))
		return
	}

	uid := string(user.Id)

	createTime := time.Now().Unix()
	var contents []map[string]string
	for _, item := range items {
		itemData := getGenerateItemDataFunc(d.generateItemDataFuncName)(user.Id, item)
		contents = append(contents, map[string]string{
			d.userIdName: uid,
			d.itemIdName: itemData,
			"event_time": strconv.FormatInt(createTime, 10),
		})
	}

	addWriteRequest := be.NewWriteRequest(be.WriteTypeAdd, d.table, d.userIdName, contents)
	_, err := d.beClient.Write(*addWriteRequest)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureBeDao\tuid=%s\terr=%v", context.RecommendId, user.Id, err))
	}

	log.Info(fmt.Sprintf("requestId=%s\tscene=%s\tuid=%s\tmsg=log history success", context.RecommendId, scene, user.Id))

}

// FilterByHistory filter user expose items.
// BeEngine already filter itmes, so here no need to filter
func (d *User2ItemExposureBeDao) FilterByHistory(uid UID, items []*Item) (ret []*Item) {
	ret = items
	return
}

func (d *User2ItemExposureBeDao) ClearHistory(user *User, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)
	if scene != d.clearLogScene {
		return
	}

	var contents []map[string]string
	contents = append(contents, map[string]string{
		d.userIdName: string(user.Id),
	})

	deleteReq := be.NewWriteRequest(be.WriteTypeDelete, d.table, d.userIdName, contents)

	_, err := d.beClient.Write(*deleteReq)
	if err != nil {
		context.LogError(fmt.Sprintf("delete user [%s] exposure items failed with err: %v", user.Id, err))
	}
}
func (d *User2ItemExposureBeDao) GetExposureItemIds(user *User, context *context.RecommendContext) (ret string) {
	return
}
