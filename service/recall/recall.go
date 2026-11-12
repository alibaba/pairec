package recall

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

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

// ICloneRecall is implemented by recalls that support AB experiment parameter overrides.
// Mirrors the sort.ICloneSort pattern: interface assertion + instance-local caching.
type ICloneRecall interface {
	CloneWithConfig(params map[string]interface{}) Recall
	GetRecallName() string
}

var recalls = make(map[string]Recall)
var recallSigns = make(map[string]string)
var recallsMu sync.RWMutex
var recallLoadMu sync.Mutex

func RegisterRecall(name string, recall Recall) {
	recallsMu.Lock()
	defer recallsMu.Unlock()
	recalls[name] = recall
}
func GetRecall(name string) (Recall, error) {
	recallsMu.RLock()
	defer recallsMu.RUnlock()
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
	recallLoadMu.Lock()
	defer recallLoadMu.Unlock()
	for _, conf := range config.RecallConfs {
		if conf.Ha3KnowledgeVectorConf != nil &&
			conf.RecallType != "HA3_CHAT_RECALL" &&
			conf.RecallType != "HA3_KNOWLEDGE_RECALL" {
			panic(fmt.Sprintf("Ha3KnowledgeVectorConf requires RecallType HA3_CHAT_RECALL or HA3_KNOWLEDGE_RECALL, got %q", conf.RecallType))
		}
		sign, _ := json.Marshal(&conf)
		recallsMu.RLock()
		old, ok := recalls[conf.Name]
		unchanged := ok && utils.Md5(string(sign)) == recallSigns[conf.Name]
		recallsMu.RUnlock()
		if unchanged {
			continue
		}

		cacheConf, inherited, cacheSign := localCacheConfig(conf, old)
		var recall Recall
		if conf.RecallType == "UserCollaborativeFilterRecall" {
			recall = NewUserCollaborativeFilterRecall(conf)
		} else if conf.RecallType == "UserTopicRecall" {
			recall = NewUserTopicRecall(conf)
		} else if conf.RecallType == "VectorRecall" {
			recall = NewVectorRecall(conf)
		} else if conf.RecallType == "UserCustomRecall" {
			recall = NewUserCustomRecall(cacheConf)
		} else if conf.RecallType == "HologresVectorRecall" {
			recall = NewHologresVectorRecall(conf)
		} else if conf.RecallType == "HologresVectorRecallV2" {
			recall = NewHologresVectorRecallV2(conf)
		} else if conf.RecallType == "ItemCollaborativeFilterRecall" {
			recall = NewItemCollaborativeFilterRecall(cacheConf)
		} else if conf.RecallType == "UserGroupHotRecall" {
			recall = NewUserGroupHotRecall(cacheConf)
		} else if conf.RecallType == "UserGlobalHotRecall" {
			recall = NewUserGlobalHotRecall(cacheConf)
		} else if conf.RecallType == "I2IVectorRecall" {
			recall = NewI2IVectorRecall(conf)
		} else if conf.RecallType == "ColdStartRecall" {
			recall = NewColdStartRecall(conf)
		} else if conf.RecallType == "MilvusVectorRecall" {
			//recall = NewMilvusVectorRecall(conf)
		} else if conf.RecallType == "BeRecall" {
			recall = NewBeRecall(conf)
		} else if conf.RecallType == "RealTimeU2IRecall" {
			inheritedI2I, i2iSign := i2iCacheConfig(conf, old)
			recall = newRealTimeU2IRecallWithCache(conf, inheritedI2I, i2iSign)
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
		} else if conf.RecallType == "GenRecRecall" {
			recall = NewGenRecRecall(conf)
		} else if conf.RecallType == "RecallEngineRecall" {
			recall = NewRecallEngineRecall(conf)
		} else if conf.RecallType == "HA3_CHAT_RECALL" {
			recall = NewHa3ChatRecall(conf)
		} else if conf.RecallType == "HA3_KNOWLEDGE_RECALL" {
			recall = NewHa3KnowledgeRecall(conf)
		}

		if recall == nil {
			panic(fmt.Sprintf("recall empty, name:%s", conf.Name))
		}

		applyLocalCache(recall, conf, inherited, cacheSign)
		recallsMu.Lock()
		recalls[conf.Name] = recall
		recallSigns[conf.Name] = utils.Md5(string(sign))
		recallsMu.Unlock()
		if previous, ok := old.(*RealTimeU2IRecall); ok && previous != nil && previous.i2iCacheSign != "" {
			if c := previous.i2iCache(); c != nil {
				current, ok := recall.(*RealTimeU2IRecall)
				if !ok || current.i2iCache() != c {
					i2iCacheAfterFunc(30*time.Second, c.InvalidateAll)
				}
			}
		}
	}
}

// localCacheBase limits inheritance to the supported A recall implementations.
func localCacheBase(r Recall) *BaseRecall {
	if utils.IsNil(r) {
		return nil
	}
	switch r := r.(type) {
	case *ItemCollaborativeFilterRecall:
		return r.BaseRecall
	case *UserGlobalHotRecall:
		return r.BaseRecall
	case *UserGroupHotRecall:
		return r.BaseRecall
	case *UserCustomRecall:
		return r.BaseRecall
	}
	return nil
}

func localCacheConfig(conf recconf.RecallConfig, old Recall) (recconf.RecallConfig, cache.Cache, string) {
	if conf.CacheAdapter != "localCache" {
		return conf, nil, ""
	}
	switch conf.RecallType {
	case "ItemCollaborativeFilterRecall", "UserGlobalHotRecall", "UserGroupHotRecall", "UserCustomRecall":
	default:
		return conf, nil, ""
	}
	withoutTTL := conf
	withoutTTL.CacheTime = 0
	encoded, err := json.Marshal(withoutTTL)
	if err != nil {
		return conf, nil, ""
	}
	cacheSign := utils.Md5(string(encoded))
	base := localCacheBase(old)
	if base != nil && base.cacheSign == cacheSign {
		if _, ok := base.cache.(*cache.LocalCache); ok {
			// Avoid constructing a throwaway cache before assigning the old one.
			conf.CacheAdapter = ""
			return conf, base.cache, cacheSign
		}
	}
	return conf, nil, cacheSign
}

func applyLocalCache(r Recall, conf recconf.RecallConfig, inherited cache.Cache, sign string) {
	if base := localCacheBase(r); base != nil {
		base.cacheSign = sign
		if inherited != nil {
			base.cache = inherited
			base.cachePrefix = conf.CachePrefix
			base.cacheTime = 1800
			if conf.CacheTime > 0 {
				base.cacheTime = conf.CacheTime
			}
		}
	}
}

type BaseRecall struct {
	modelName   string
	cache       cache.Cache
	cachePrefix string
	cacheTime   int
	cacheSign   string
	itemType    string
	recallCount int
	recallAlgo  string

	// cloneInstances caches recall instances cloned with AB params, keyed by md5 of the AB config.
	cloneInstances map[string]Recall
	cloneMu        sync.Mutex
}

func NewBaseRecall(config recconf.RecallConfig) *BaseRecall {
	recall := &BaseRecall{
		modelName:      config.Name,
		itemType:       config.ItemType,
		recallCount:    config.RecallCount,
		recallAlgo:     config.RecallAlgo,
		cloneInstances: make(map[string]Recall),
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

// GetRecallName returns the configured recall name.
func (r *BaseRecall) GetRecallName() string {
	return r.modelName
}

// cloneWithBuilder implements the common clone-by-AB-config flow for all recall types.
// It computes a cache key from params JSON (Go json.Marshal sorts map keys, so the
// key is deterministic), checks the instance-local cache, and on miss deserializes
// params into a RecallConfig and invokes builder to construct a new instance.
// If anything fails (marshal error / panic in builder), nil is returned and the caller
// should fall back to the original recall instance.
func (r *BaseRecall) cloneWithBuilder(params map[string]interface{}, builder func(recconf.RecallConfig) Recall) Recall {
	// Compute cache key directly from params (hot path: 1 marshal + 1 md5)
	j, err := json.Marshal(params)
	if err != nil {
		log.Error(fmt.Sprintf("event=CloneWithConfig\trecall=%s\terror=%v", r.modelName, err))
		return nil
	}
	key := utils.Md5(string(j))

	// Fast path: check cache
	r.cloneMu.Lock()
	if inst, ok := r.cloneInstances[key]; ok {
		r.cloneMu.Unlock()
		return inst
	}
	r.cloneMu.Unlock()

	// Cache miss: deserialize params into RecallConfig
	cfg := recconf.RecallConfig{}
	if err := json.Unmarshal(j, &cfg); err != nil {
		log.Error(fmt.Sprintf("event=CloneWithConfig\trecall=%s\terror=%v", r.modelName, err))
		return nil
	}
	cfg.Name = r.modelName

	// Build new instance with panic protection
	var newRecall Recall
	func() {
		defer func() {
			if e := recover(); e != nil {
				log.Error(fmt.Sprintf("event=CloneWithConfig\trecall=%s\tpanic=%v\tstack=%s",
					r.modelName, e, strings.ReplaceAll(string(debug.Stack()), "\n", "\t")))
				newRecall = nil
			}
		}()
		newRecall = builder(cfg)
	}()

	if newRecall == nil {
		return nil
	}

	// Double-check and store
	r.cloneMu.Lock()
	if inst, ok := r.cloneInstances[key]; ok {
		r.cloneMu.Unlock()
		return inst
	}
	r.cloneInstances[key] = newRecall
	r.cloneMu.Unlock()
	log.Info(fmt.Sprintf("event=CloneWithConfig\trecall=%s\tkey=%s\tregister new clone", r.modelName, key))
	return newRecall
}
