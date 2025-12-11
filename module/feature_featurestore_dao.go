package module

import (
	"fmt"
	"strings"
	"sync"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

type FeatureFeatureStoreDao struct {
	*FeatureBaseDao
	client           *fs.FSClient
	fsModel          string
	fsEntity         string
	fsViewName       string
	userSelectFields string
	itemSelectFields string
	fieldsMap        sync.Map
}

func NewFeatureFeatureStoreDao(config recconf.FeatureDaoConfig) *FeatureFeatureStoreDao {
	dao := &FeatureFeatureStoreDao{
		FeatureBaseDao:   NewFeatureBaseDao(&config),
		fsModel:          config.FeatureStoreModelName,
		fsEntity:         config.FeatureStoreEntityName,
		fsViewName:       config.FeatureStoreViewName,
		userSelectFields: config.UserSelectFields,
		itemSelectFields: config.ItemSelectFields,
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
	appendKeys := make([]string, 0)
	log.Info(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\tmsg=featureAppendKey(%s)", context.RecommendId, d.featureAppendKey))
	if d.featureAppendKey != "" {
		appendComms := strings.Split(d.featureAppendKey, ":")
		if len(appendComms) < 2 {
			log.Error(fmt.Sprintf("requestId=%s\tuid=%s\terror=featureAppendKey error(%s)", context.RecommendId, user.Id, d.featureAppendKey))
			return
		}
		err := error(nil)
		appendKeys, err = user.ListStringProperty(appendComms[1])
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=property error(%s):%s", context.RecommendId, appendComms[1], err))
			return
		}
		log.Info(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\tmsg=appendKeys(%v)", context.RecommendId, appendKeys))
	}
	// hit user cache
	if d.cache != nil {
		if cacheValue, ok := d.cache.GetIfPresent(key); ok {
			if d.cacheFeaturesName != "" {
				user.AddCacheFeatures(d.cacheFeaturesName, cacheValue.(map[string]interface{}))
			} else {
				user.AddProperties(cacheValue.(map[string]interface{}))
			}
			if context.Debug {
				log.Info(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\tmsg=hit cache(%s)", context.RecommendId, key))
			}
			return
		}
	}

	if d.fsViewName != "" {
		d.doUserFeatureFetchWithFeatureView(user, context, key)
	} else if len(appendKeys) > 0 {
		d.doUserFeatureFetchWithEntityAppendKeys(user, context, key, appendKeys)
	} else {
		d.doUserFeatureFetchWithEntity(user, context, key)
	}
}

func (d *FeatureFeatureStoreDao) doUserFeatureFetchWithEntity(user *User, context *context.RecommendContext, key string) {
	model := d.client.GetProject().GetModel(d.fsModel)
	if model == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=model not found(%s)", context.RecommendId, d.fsModel))
		return
	}

	var (
		labelFieldMap map[string]bool
		modelFieldMap map[string]bool
	)
	labelTable := model.GetLabelTable()
	if labelTable == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=label table not found(%s)", context.RecommendId, model.LabelDatasourceTable))
		return
	}

	if labelFields, ok := d.fieldsMap.Load(labelTable.Name); ok {
		labelFieldMap = labelFields.(map[string]bool)
	} else {
		fNames := labelTable.GetFeatureNames()
		labelFieldMap = make(map[string]bool, len(fNames))
		for _, f := range fNames {
			labelFieldMap[f] = true
		}
		d.fieldsMap.Store(labelTable.Name, labelFieldMap)
	}
	if modelFields, ok := d.fieldsMap.Load(model.Name); ok {
		modelFieldMap = modelFields.(map[string]bool)
	} else {
		modelFeatures := model.Features

		modelFieldMap = make(map[string]bool, len(modelFeatures))
		for _, f := range modelFeatures {
			if f.AliasName != "" {
				modelFieldMap[f.AliasName] = true
			} else {
				modelFieldMap[f.Name] = true
			}
		}
		d.fieldsMap.Store(model.Name, modelFieldMap)
	}
	if len(labelFieldMap) > 0 {
		contextFeatures := context.GetParameter("features")
		if contextFeatures != nil {
			var deleteProperties []string
			if ctxFeatures, ok := contextFeatures.(map[string]any); ok {
				for k := range ctxFeatures {
					_, labelOK := labelFieldMap[k]
					_, modelOK := modelFieldMap[k]
					if modelOK && !labelOK {
						deleteProperties = append(deleteProperties, k)
					}
				}
			}
			user.DeleteProperties(deleteProperties)
		}
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

	if model.GetLabelPriorityLevel() == 1 {
		contextFeatures := context.GetParameter("features")
		if contextFeatures != nil {
			if ctxFeatures, ok := contextFeatures.(map[string]any); ok {
				for k, v := range ctxFeatures {
					_, labelOK := labelFieldMap[k]
					_, modelOK := modelFieldMap[k]
					if modelOK && labelOK {
						features[0][k] = v
					}
				}
			}
		}
	}

	if d.cacheFeaturesName != "" {
		user.AddCacheFeatures(d.cacheFeaturesName, features[0])
	} else {
		user.AddProperties(features[0])
	}
	if d.cache != nil {
		d.cache.Put(key, features[0])
	}
}

func (d *FeatureFeatureStoreDao) doUserFeatureFetchWithEntityAppendKeys(user *User, context *context.RecommendContext, key string, appendKeys []string) {
	model := d.client.GetProject().GetModel(d.fsModel)
	if model == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=model not found(%s)", context.RecommendId, d.fsModel))
		return
	}

	var (
		labelFieldMap map[string]bool
		modelFieldMap map[string]bool
	)
	labelTable := model.GetLabelTable()
	if labelTable == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=label table not found(%s)", context.RecommendId, model.LabelDatasourceTable))
		return
	}

	if labelFields, ok := d.fieldsMap.Load(labelTable.Name); ok {
		labelFieldMap = labelFields.(map[string]bool)
	} else {
		fNames := labelTable.GetFeatureNames()
		labelFieldMap = make(map[string]bool, len(fNames))
		for _, f := range fNames {
			labelFieldMap[f] = true
		}
		d.fieldsMap.Store(labelTable.Name, labelFieldMap)
	}
	if modelFields, ok := d.fieldsMap.Load(model.Name); ok {
		modelFieldMap = modelFields.(map[string]bool)
	} else {
		modelFeatures := model.Features

		modelFieldMap = make(map[string]bool, len(modelFeatures))
		for _, f := range modelFeatures {
			if f.AliasName != "" {
				modelFieldMap[f.AliasName] = true
			} else {
				modelFieldMap[f.Name] = true
			}
		}
		d.fieldsMap.Store(model.Name, modelFieldMap)
	}
	if len(labelFieldMap) > 0 {
		contextFeatures := context.GetParameter("features")
		if contextFeatures != nil {
			var deleteProperties []string
			if ctxFeatures, ok := contextFeatures.(map[string]any); ok {
				for k := range ctxFeatures {
					_, labelOK := labelFieldMap[k]
					_, modelOK := modelFieldMap[k]
					if modelOK && !labelOK {
						deleteProperties = append(deleteProperties, k)
					}
				}
			}
			user.DeleteProperties(deleteProperties)
		}
	}

	entity := d.client.GetProject().GetFeatureEntity(d.fsEntity)
	if entity == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=feature entity not found(%s)", context.RecommendId, d.fsEntity))
		return
	}

	appendSequenceKeys := make([]interface{}, 0, len(appendKeys)+1)
	appendSequenceKeys = append(appendSequenceKeys, key)
	for _, appendKey := range appendKeys {
		appendSequenceKeys = append(appendSequenceKeys, appendKey)
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\tmsg=get online features with aggregated sequence(%v)", context.RecommendId, appendSequenceKeys))
	features, err := model.GetOnlineFeaturesWithAggregatedSequence(key, appendSequenceKeys, d.fsEntity)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=get features error(%s)", context.RecommendId, err))
		return
	}
	if len(features) == 0 {
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=get features empty", context.RecommendId))
		return
	}

	if model.GetLabelPriorityLevel() == 1 {
		contextFeatures := context.GetParameter("features")
		if contextFeatures != nil {
			if ctxFeatures, ok := contextFeatures.(map[string]any); ok {
				for k, v := range ctxFeatures {
					_, labelOK := labelFieldMap[k]
					_, modelOK := modelFieldMap[k]
					if modelOK && labelOK {
						features[k] = v
					}
				}
			}
		}
	}

	if d.cacheFeaturesName != "" {
		user.AddCacheFeatures(d.cacheFeaturesName, features)
	} else {
		user.AddProperties(features)
	}
	if d.cache != nil {
		d.cache.Put(key, features)
	}
}

func (d *FeatureFeatureStoreDao) doUserFeatureFetchWithFeatureView(user *User, context *context.RecommendContext, key string) {
	featureView := d.client.GetProject().GetFeatureView(d.fsViewName)
	if featureView == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=feature view not found(%s)", context.RecommendId, d.fsViewName))
		return
	}

	var featuresNames []string
	if d.userSelectFields == "" || d.userSelectFields == "*" {
		featuresNames = []string{"*"}
	} else {
		featuresNames = strings.Split(d.userSelectFields, ",")
	}
	features, err := featureView.GetOnlineFeatures([]any{key}, featuresNames, nil)
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
	if d.cache != nil {
		d.cache.Put(key, features[0])
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
	keysMap := make(map[string]bool)
	key2Item := make(map[string][]*Item, len(items))
	for _, item := range items {
		var key string
		if fk == "item:id" {
			key = string(item.Id)
		} else {
			key = item.StringProperty(fk)
		}
		if d.cache != nil {
			if cacheValue, ok := d.cache.GetIfPresent(key); ok {
				item.AddProperties(cacheValue.(map[string]any))
				if context.Debug {
					item.AddProperty("__debug_cache_hit__", true)
				}
				continue
			}
		}

		keysMap[key] = true
		key2Item[key] = append(key2Item[key], item)
	}
	for k := range keysMap {
		keys = append(keys, k)
	}

	if len(keys) == 0 {
		return
	}

	var (
		entityJoinId string
		features     []map[string]any
		err          error
	)
	if d.fsViewName != "" {
		featureView := d.client.GetProject().GetFeatureView(d.fsViewName)
		if featureView == nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=feature view not found(%s)", context.RecommendId, d.fsViewName))
			return
		}
		features, err = d.doItemsFeatureFetchWithFeatureView(featureView, keys)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=%v", context.RecommendId, err))
			return
		}

		entityName := featureView.GetFeatureEntityName()
		entity := d.client.GetProject().GetFeatureEntity(entityName)
		if entity == nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=feature entity not found(%s)", context.RecommendId, entityName))
			return
		}

		entityJoinId = entity.FeatureEntityJoinid

	} else {
		entity := d.client.GetProject().GetFeatureEntity(d.fsEntity)
		if entity == nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=feature entity not found(%s)", context.RecommendId, d.fsEntity))
			return
		}
		features, err = d.doItemsFeatureFetchWithEntity(entity, keys)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureFeatureStoreDao\terror=%v", context.RecommendId, err))
			return
		}

		entityJoinId = entity.FeatureEntityJoinid

	}

	for key, itemlist := range key2Item {
		for i, featureMap := range features {
			if key == featureMap[entityJoinId] {
				for _, item := range itemlist {
					item.AddProperties(featureMap)
					if d.cache != nil {
						d.cache.Put(key, featureMap)
					}
				}
				features[i] = features[len(features)-1]
				features = features[:len(features)-1]
			}
		}
	}

}

func (d *FeatureFeatureStoreDao) doItemsFeatureFetchWithEntity(entity *domain.FeatureEntity, keys []any) ([]map[string]any, error) {
	model := d.client.GetProject().GetModel(d.fsModel)
	if model == nil {
		return nil, fmt.Errorf("model not found(%s)", d.fsModel)
	}

	features, err := model.GetOnlineFeaturesWithEntity(map[string][]interface{}{entity.FeatureEntityJoinid: keys}, d.fsEntity)
	if err != nil {
		return nil, err
	}

	return features, nil
}

func (d *FeatureFeatureStoreDao) doItemsFeatureFetchWithFeatureView(featureView domain.FeatureView, keys []any) ([]map[string]any, error) {
	var featuresNames []string
	if d.itemSelectFields == "" || d.itemSelectFields == "*" {
		featuresNames = []string{"*"}
	} else {
		featuresNames = strings.Split(d.itemSelectFields, ",")
	}

	features, err := featureView.GetOnlineFeatures(keys, featuresNames, nil)
	if err != nil {
		return nil, err
	}

	return features, nil
}
