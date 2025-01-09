package recall

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/persist/cache"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type Recall interface {
	GetCandidateItems(user *module.User, context *context.RecommendContext) []*module.Item
}

var recalls = make(map[string]Recall)
var recallSigns = make(map[string]string)

func RegisterRecall(name string, recall Recall) {
	recalls[name] = recall
}
func GetRecall(name string) (Recall, error) {
	recall, ok := recalls[name]
	if !ok {
		return nil, fmt.Errorf("recall:not found, recall name:%s", name)
	}

	if recall == nil {
		return nil, fmt.Errorf("recall is nil, recall name:%s", name)
	}

	return recall, nil
}
func Load(config *recconf.RecommendConfig) {
	for _, conf := range config.RecallConfs {
		if _, ok := recalls[conf.Name]; ok {
			sign, _ := json.Marshal(&conf)
			if utils.Md5(string(sign)) == recallSigns[conf.Name] {
				continue
			}
		}

		var recall Recall
		if conf.RecallType == "UserCollaborativeFilterRecall" {
			recall = NewUserCollaborativeFilterRecall(conf)
		} else if conf.RecallType == "UserTopicRecall" {
			recall = NewUserTopicRecall(conf)
		} else if conf.RecallType == "VectorRecall" {
			recall = NewVectorRecall(conf)
		} else if conf.RecallType == "UserCustomRecall" {
			recall = NewUserCustomRecall(conf)
		} else if conf.RecallType == "HologresVectorRecall" {
			recall = NewHologresVectorRecall(conf)
		} else if conf.RecallType == "ItemCollaborativeFilterRecall" {
			recall = NewItemCollaborativeFilterRecall(conf)
		} else if conf.RecallType == "UserGroupHotRecall" {
			recall = NewUserGroupHotRecall(conf)
		} else if conf.RecallType == "UserGlobalHotRecall" {
			recall = NewUserGlobalHotRecall(conf)
		} else if conf.RecallType == "I2IVectorRecall" {
			recall = NewI2IVectorRecall(conf)
		} else if conf.RecallType == "ColdStartRecall" {
			recall = NewColdStartRecall(conf)
		} else if conf.RecallType == "MilvusVectorRecall" {
			//recall = NewMilvusVectorRecall(conf)
		} else if conf.RecallType == "BeRecall" {
			recall = NewBeRecall(conf)
		} else if conf.RecallType == "RealTimeU2IRecall" {
			recall = NewRealTimeU2IRecall(conf)
		} else if conf.RecallType == "OnlineHologresVectorRecall" {
			recall = NewOnlineHologresVectorRecall(conf)
		} else if conf.RecallType == "GraphRecall" {
			recall = NewGraphRecall(conf)
		} else if conf.RecallType == "MockRecall" {
			recall = NewMockRecall(conf)
		} else if conf.RecallType == "OpenSearchRecall" {
			recall = NewOpenSearchRecall(conf)
		} else if conf.RecallType == "OnlineVectorRecall" {
			recall = NewOnlineVectorRecall(conf)
		}

		if recall == nil {
			panic(fmt.Sprintf("recall empty, name:%s", conf.Name))
		}

		RegisterRecall(conf.Name, recall)
		sign, _ := json.Marshal(&conf)
		recallSigns[conf.Name] = utils.Md5(string(sign))

	}

}

type BaseRecall struct {
	modelName   string
	cache       cache.Cache
	cachePrefix string
	cacheTime   int
	itemType    string
	recallCount int
	recallAlgo  string
}

func NewBaseRecall(config recconf.RecallConfig) *BaseRecall {
	recall := &BaseRecall{
		modelName:   config.Name,
		itemType:    config.ItemType,
		recallCount: config.RecallCount,
		recallAlgo:  config.RecallAlgo,
	}
	if len(config.CacheAdapter) > 0 {
		cache, err := cache.NewCache(config.CacheAdapter,
			config.CacheConfig)
		if err != nil {
			log.Error(fmt.Sprintf("error=%v", err))
		} else {
			recall.cache = cache
			recall.cachePrefix = config.CachePrefix
			recall.cacheTime = 1800
			if config.CacheTime > 0 {
				recall.cacheTime = config.CacheTime
			}
		}
	}

	return recall
}
