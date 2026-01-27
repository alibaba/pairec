package module

import (
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/recallengine"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	re "github.com/aliyun/aliyun-pairec-config-go-sdk/v2/recallengine"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type User2ItemExposureRecallEngineDao struct {
	reClient                 *recallengine.RecallEngineClient
	table                    string
	timeInterval             int64 //  second
	generateItemDataFuncName string
	writeLogExcludeScenes    map[string]bool
	clearLogScene            string
	onlyLogUserExposeFlag    bool
	generateUserProgram      *vm.Program
}

func NewUser2ItemExposureRecallEngineDao(config recconf.FilterConfig) *User2ItemExposureRecallEngineDao {
	reClient, err := recallengine.GetRecallEngineClient(config.DaoConf.RecallEngineName)
	if err != nil {
		panic(fmt.Errorf("get recallengine client error=%v", err))
	}
	dao := &User2ItemExposureRecallEngineDao{
		timeInterval:             0,
		generateItemDataFuncName: config.GenerateItemDataFuncName,
		writeLogExcludeScenes:    make(map[string]bool),
		clearLogScene:            config.ClearLogIfNotEnoughScene,
		reClient:                 reClient,
		onlyLogUserExposeFlag:    config.OnlyLogUserExposeFlag,
	}
	dao.table = config.DaoConf.RecallEngineTableName

	if config.TimeInterval > 0 {
		dao.timeInterval = int64(config.TimeInterval)
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

func (d *User2ItemExposureRecallEngineDao) LogHistory(user *User, items []*Item, context *context.RecommendContext) {
	start := time.Now()
	scene := context.GetParameter("scene").(string)
	if _, exist := d.writeLogExcludeScenes[scene]; exist {
		return
	}

	if len(items) == 0 {
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureRecallEngineDao\terr=items empty", context.RecommendId))
		return
	}

	userData, err := d.getUserData(user.Id, context)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureHologresDao\tuid=%s\terr=%v", context.RecommendId, userData, err))
		return
	}

	//ttl := int64(featureView.GetTTL())

	//ts := start.Unix() - ttl + d.timeInterval
	ts := time.Now().UnixMilli()

	writeRequest := re.WriteRequest{}
	for _, item := range items {
		itemData := getGenerateItemDataFunc(d.generateItemDataFuncName)(user.Id, item)
		writeRequest.Content = append(writeRequest.Content, map[string]any{
			"item_id":   itemData,
			"timestamp": ts,
			"user_id":   userData,
		})
	}

	_, err = d.reClient.GetRecallEngineClient().Write(d.reClient.InstanceId(), d.table, &writeRequest)

	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureRecallEngineDao\tuid=%s\terr=%v", context.RecommendId, user.Id, err))
		return
	}

	log.Info(fmt.Sprintf("requestId=%s\tscene=%s\tuid=%s\tmsg=log history success\tcost=%d", context.RecommendId, scene, user.Id, utils.CostTime(start)))
}
func (d *User2ItemExposureRecallEngineDao) FilterByHistory(uid UID, items []*Item, context *context.RecommendContext) (ret []*Item) {
	return items
}

func (d *User2ItemExposureRecallEngineDao) ClearHistory(user *User, context *context.RecommendContext) {
}

func (d *User2ItemExposureRecallEngineDao) GetExposureItemIds(user *User, context *context.RecommendContext) (ret string) {
	return
}
func (d *User2ItemExposureRecallEngineDao) getUserData(uid UID, context *context.RecommendContext) (string, error) {
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
