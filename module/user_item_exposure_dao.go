package module

import (
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/hook"
)

type GenerateItemDataFunc func(uid UID, item *Item) string

var generateItemDataFuncMap map[string]GenerateItemDataFunc

func init() {
	generateItemDataFuncMap = make(map[string]GenerateItemDataFunc)
}

func RegisterGenerateItemDataFunc(name string, f GenerateItemDataFunc) {
	generateItemDataFuncMap[name] = f
}

func getGenerateItemDataFunc(name string) GenerateItemDataFunc {
	if f, exist := generateItemDataFuncMap[name]; exist {
		return f
	} else {
		return defaultGenerateItemDataFunc
	}
}

// default function
func defaultGenerateItemDataFunc(uid UID, item *Item) string {
	return string(item.Id)
}

type User2ItemExposureDao interface {
	LogHistory(user *User, items []*Item, context *context.RecommendContext)
	FilterByHistory(uid UID, ids []*Item) (ret []*Item)
	ClearHistory(user *User, context *context.RecommendContext)
	GetExposureItemIds(user *User, context *context.RecommendContext) (ret string)
}

func NewUser2ItemExposureDao(config recconf.FilterConfig) User2ItemExposureDao {
	var dao User2ItemExposureDao
	if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Mysql {
		dao = NewUser2ItemExposureMysqlDao(config)
	} else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_TableStore {
		dao = NewUser2ItemExposureTableStoreDao(config)
	} else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		dao = NewUser2ItemExposureHologresDao(config)
	} else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Redis {
		dao = NewUser2ItemExposureRedisDao(config)
	} else if config.DaoConf.AdapterType == recconf.DataSource_Type_BE {
		dao = NewUser2ItemExposureBeDao(config)
	} else if config.DaoConf.AdapterType == recconf.Datasource_Type_Graph {
		dao = NewUser2ItemExposureGraphDao(config)
	} else {
		panic("not found User2ItemExposureDao implement")
	}

	if config.WriteLog {
		f := func(dao User2ItemExposureDao) hook.RecommendCleanHookFunc {

			return func(context *context.RecommendContext, params ...interface{}) {
				user := params[0].(*User)
				items := params[1].([]*Item)
				dao.LogHistory(user, items, context)
			}
		}(dao)
		hook.RegisterRecommendCleanHook(fmt.Sprintf("%s_write_log", config.Name), f)
	}

	if config.ClearLogIfNotEnoughScene != "" {
		f := func(dao User2ItemExposureDao) hook.RecommendCleanHookFunc {
			return func(context *context.RecommendContext, params ...interface{}) {
				user := params[0].(*User)
				items := params[1].([]*Item)
				if len(items) < context.Size {
					dao.ClearHistory(user, context)
				}
			}
		}(dao)
		hook.RegisterRecommendCleanHook(fmt.Sprintf("%s_clear_log", config.Name), f)
	}
	return dao
}
