package module

import (
	gocontext "context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/recconf"
)

const (
	Feature_Store_User    = "user"
	Feature_Store_Item    = "item"
	Feature_Type_Sequence = "sequence_feature"
	Feature_Type_RT_Cnt   = "be_rt_cnt_feature"
)

// FeatureDao is interface, it define FeatureFetch to featch feature from implement
type FeatureDao interface {
	FeatureFetch(user *User, items []*Item, context *context.RecommendContext)
}

type EmptyFeatureDao struct {
	*FeatureBaseDao
}

func NewEmptyFeatureDao(config recconf.FeatureDaoConfig) *EmptyFeatureDao {
	return &EmptyFeatureDao{
		FeatureBaseDao: NewFeatureBaseDao(&config),
	}
}
func (d *EmptyFeatureDao) FeatureFetch(user *User, items []*Item, context *context.RecommendContext) {
	if d.featureStore == Feature_Store_User {
		d.userFeatureFetch(user, context)
	}
}
func (d *EmptyFeatureDao) userFeatureFetch(user *User, context *context.RecommendContext) {
	if d.loadFromCacheFeaturesName != "" {
		ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 150*time.Millisecond)
		defer cancel()
		select {
		case <-user.FeatureAsyncLoadCh():
			user.LoadCacheFeatures(d.loadFromCacheFeaturesName)
		case <-ctx.Done():
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureDao\tcount=%d\tname=%s\tcache_names=%s\terror=%v", context.RecommendId,
				user.FeatureAsyncLoadCount(), d.loadFromCacheFeaturesName, strings.Join(user.GetCacheFeaturesNames(), ","), ctx.Err()))
			user.LoadCacheFeatures(d.loadFromCacheFeaturesName)
		}
		return
	}
}

// NewFeatureDao create FeatureDao from config
// config.AdapterType is decide the implement
func NewFeatureDao(config recconf.FeatureDaoConfig) FeatureDao {
	if config.AdapterType == recconf.DaoConf_Adapter_Redis {
		return NewFeatureRedisDao(config)
	} else if config.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewFeatureHologresDao(config)
	} else if config.AdapterType == recconf.DaoConf_Adapter_TableStore {
		return NewFeatureTablestoreDao(config)
	} else if config.AdapterType == recconf.DaoConf_Adapter_Mysql {
		return NewFeatureMysqlDao(config)
	} else if config.AdapterType == recconf.DataSource_Type_ClickHouse {
		return NewFeatureClickHouseDao(config)
	} else if config.AdapterType == recconf.DataSource_Type_FeatureStore {
		return NewFeatureFeatureStoreDao(config)
	} else if config.AdapterType == recconf.DataSource_Type_BE {
		return NewFeatureBeDao(config)
	} else if config.AdapterType == recconf.DataSource_Type_Lindorm {
		return NewFeatureLindormDao(config)
	} else if config.AdapterType == recconf.DataSource_Type_HBase_Thrift {
		return NewFeatureHBaseThriftDao(config)
	} else {
		return NewEmptyFeatureDao(config)
	}

}

type FeatureBaseDao struct {
	featureKey                string // use the value of key to featch data
	featureStore              string // user or item
	featureType               string
	cacheFeaturesName         string
	loadFromCacheFeaturesName string
	//featureAsyncLoad          bool

	// sequence feature has under attribute
	sequenceLength      int
	sequenceName        string
	sequenceEvent       string
	sequenceDelim       string
	sequenceDimFields   []string
	sequencePlayTimeMap map[string]float64
}

func NewFeatureBaseDao(config *recconf.FeatureDaoConfig) *FeatureBaseDao {
	dao := FeatureBaseDao{
		featureKey:                config.FeatureKey,
		featureStore:              config.FeatureStore,
		featureType:               config.FeatureType,
		cacheFeaturesName:         config.CacheFeaturesName,
		loadFromCacheFeaturesName: config.LoadFromCacheFeaturesName,
		//featureAsyncLoad:          config.FeatureAsyncLoad,

		sequenceLength:      config.SequenceLength,
		sequenceName:        config.SequenceName,
		sequenceEvent:       config.SequenceEvent,
		sequenceDelim:       config.SequenceDelim,
		sequencePlayTimeMap: make(map[string]float64, 0),
	}

	if config.SequencePlayTime != "" {
		playTimes := strings.Split(config.SequencePlayTime, ";")
		for _, eventTime := range playTimes {
			strs := strings.Split(eventTime, ":")
			if len(strs) == 2 {
				if t, err := strconv.ParseFloat(strs[1], 64); err == nil {
					dao.sequencePlayTimeMap[strs[0]] = t
				}
			}
		}
	}
	if config.SequenceDimFields != "" {
		dao.sequenceDimFields = strings.Split(config.SequenceDimFields, ",")
	}
	return &dao
}
