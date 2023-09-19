package module

import (
	"fmt"
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
)

type FeatureFeatureStoreDao struct {
	*FeatureBaseDao
	client             *fs.FSClient
	fsModel            string
	fsEntity           string
	userFeatureKeyName string
	itemFeatureKeyName string
	//timestampFeatureKeyName string
	//eventFeatureKeyName     string
	//playTimeFeatureKeyName  string
	//tsFeatureKeyName        string
	userSelectFields string
	itemSelectFields string
}

func NewFeatureFeatureStoreDao(config recconf.FeatureDaoConfig) *FeatureFeatureStoreDao {
	dao := &FeatureFeatureStoreDao{
		FeatureBaseDao:     NewFeatureBaseDao(&config),
		fsModel:            config.FeatureStoreModelName,
		fsEntity:           config.FeatureStoreEntityName,
		userFeatureKeyName: config.UserFeatureKeyName,
		itemFeatureKeyName: config.ItemFeatureKeyName,
		userSelectFields:   config.UserSelectFields,
		itemSelectFields:   config.ItemSelectFields,
	}
	client, err := fs.GetFeatureStoreClient(config.FeatureStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao.client = client
	return dao
}

func (d *FeatureFeatureStoreDao) FeatureFetch(user *User, items []*Item, context *context.RecommendContext) {
	if d.featureStore == Feature_Store_User && d.featureType == Feature_Type_Sequence {
		//d.userSequenceFeatureFetch(user, context)
	} else if d.featureStore == Feature_Store_User {
		d.userFeatureFetch(user, context)
	} else {
		d.itemsFeatureFetch(items, context)
	}
}

func (d *FeatureFeatureStoreDao) userFeatureFetch(user *User, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=%v", context.RecommendId, err))
			return
		}
	}()

	comms := strings.Split(d.featureKey, ":")
	if len(comms) < 2 {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\terror=featureKey error(%s)", context.RecommendId, user.Id, d.featureKey))
		return
	}

	key := user.StringProperty(comms[1])
	if key == "" {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=property not found(%s)", context.RecommendId, comms[1]))
		return
	}

	model := d.client.GetProject().GetModel(d.fsModel)
	if model == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=model not found(%s)", context.RecommendId, d.fsModel))
		return
	}

	entity := d.client.GetProject().GetFeatureEntity(d.fsEntity)
	if entity == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=feature entity not found(%s)", context.RecommendId, d.fsEntity))
		return
	}

	features, err := model.GetOnlineFeaturesWithEntity(map[string][]interface{}{entity.FeatureEntityJoinid: {key}}, d.fsEntity)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=get features error(%s)", context.RecommendId, err))
		return
	}
	if len(features) == 0 {
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=get features empty", context.RecommendId))
		return
	}

	if d.cacheFeaturesName != "" {
		user.AddCacheFeatures(d.cacheFeaturesName, features[0])
	} else {
		user.AddProperties(features[0])
	}
}

func (d *FeatureFeatureStoreDao) itemsFeatureFetch(items []*Item, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=%v", context.RecommendId, err))
			return
		}
	}()

	if len(items) == 0 {
		return
	}

	fk := d.featureKey
	if fk != "item:id" {
		comms := strings.Split(d.featureKey, ":")
		if len(comms) < 2 {
			log.Error(fmt.Sprintf("requestId=%s\tevent=itemsFeatureFetch\terror=featureKey error(%s)", context.RecommendId, d.featureKey))
			return
		}

		fk = comms[1]
	}
	var keys []interface{}
	key2Item := make(map[string]*Item, len(items))
	for _, item := range items {
		var key string
		if fk == "item:id" {
			key = string(item.Id)
		} else {
			key = item.StringProperty(fk)
		}
		keys = append(keys, key)
		key2Item[key] = item
	}
	model := d.client.GetProject().GetModel(d.fsModel)
	if model == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=model not found(%s)", context.RecommendId, d.fsModel))
		return
	}

	entity := d.client.GetProject().GetFeatureEntity(d.fsEntity)
	if entity == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=feature entity not found(%s)", context.RecommendId, d.fsEntity))
		return
	}

	features, err := model.GetOnlineFeaturesWithEntity(map[string][]interface{}{entity.FeatureEntityJoinid: keys}, d.fsEntity)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=get features error(%s)", context.RecommendId, err))
		return
	}

	for key, item := range key2Item {
		for i, featureMap := range features {
			if key == featureMap[entity.FeatureEntityJoinid] {
				item.AddProperties(featureMap)
				features[i] = features[len(features)-1]
				features = features[:len(features)-1]
			}
		}
	}

}
