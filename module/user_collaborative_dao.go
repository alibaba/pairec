package module

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
)

type UserCollaborativeDao interface {
	ListItemsByUser(user *User, context *context.RecommendContext) []*Item
	GetTriggers(user *User, context *context.RecommendContext) (itemTriggers map[string]float64)
	GetTriggerInfos(user *User, context *context.RecommendContext) (triggerInfos []*TriggerInfo)
}

func NewUserCollaborativeDao(config recconf.RecallConfig) UserCollaborativeDao {
	if config.UserCollaborativeDaoConf.AdapterType == recconf.DaoConf_Adapter_Mysql {
		if config.UserCollaborativeDaoConf.Adapter == "UserCollaborativeMysqlDao" {
			return NewUserCollaborativeMysqlDao(config)
		} else if config.UserCollaborativeDaoConf.Adapter == "UserVideoCollaborativeMysqlDao" {
			return NewUserVideoCollaborativeMysqlDao(config)
		}
	} else if config.UserCollaborativeDaoConf.AdapterType == recconf.DaoConf_Adapter_TableStore {
		return NewUserCollaborativeTableStoreDao(config)
	} else if config.UserCollaborativeDaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewUserCollaborativeHologresDao(config)
	} else if config.UserCollaborativeDaoConf.AdapterType == recconf.DaoConf_Adapter_Redis {
		return NewUserCollaborativeRedisDao(config)
	}

	panic("not found UserCollaborativeDao implement")
}

func mergeUserCollaborativeItemsResult(itemCh chan []*Item, cpuCount int, normalization bool) []*Item {
	retMap := make(map[ItemId]*Item)
	var maxScore float64 = 0

	for i := 0; i < cpuCount; i++ {
		items := <-itemCh
		for _, item := range items {
			if retMap[item.Id] == nil {
				retMap[item.Id] = item
			} else {
				retMap[item.Id].Score += item.Score
			}
		}
	}

	ret := make([]*Item, 0, len(retMap))
	for _, item := range retMap {
		if item.Score > maxScore {
			maxScore = item.Score
		}
		ret = append(ret, item)
	}

	if normalization && maxScore > 0 {
		for i := range ret {
			ret[i].Score /= maxScore
		}
	}

	return ret
}
