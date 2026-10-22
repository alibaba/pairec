package module

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	reccontext "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

// SharedFeatureFeatureStoreDao is an opt-in implementation; the existing
// FeatureFeatureStoreDao and its private cache remain unchanged.
type SharedFeatureFeatureStoreDao struct {
	*FeatureBaseDao
	conf    recconf.FeatureDaoConfig
	project func() *domain.Project
}

var _ FeatureDao = (*SharedFeatureFeatureStoreDao)(nil)

func NewSharedFeatureFeatureStoreDao(conf recconf.FeatureDaoConfig) *SharedFeatureFeatureStoreDao {
	validateSharedFeatureConfig(conf)
	if _, err := fs.GetFeatureCache(conf.FeatureCacheName); err != nil {
		panic(err)
	}
	client, err := fs.GetFeatureStoreClient(conf.FeatureStoreName)
	if err != nil {
		panic(err)
	}
	return newSharedFeatureFeatureStoreDao(conf, client.GetProject)
}

func validateSharedFeatureConfig(conf recconf.FeatureDaoConfig) {
	if strings.TrimSpace(conf.FeatureCacheName) == "" {
		panic("FeatureCacheName must not be blank")
	}
	if conf.AdapterType != recconf.DataSource_Type_FeatureStore {
		panic("FeatureCacheName is only supported by the featurestore adapter")
	}
	if conf.FeatureStore == Feature_Store_User && conf.FeatureType == Feature_Type_Sequence {
		panic("shared FeatureStore sequence_feature loading is not implemented; use the view/model path")
	}
}

func newSharedFeatureFeatureStoreDao(conf recconf.FeatureDaoConfig, project func() *domain.Project) *SharedFeatureFeatureStoreDao {
	baseConf := conf
	baseConf.CacheSize = 0
	return &SharedFeatureFeatureStoreDao{FeatureBaseDao: NewFeatureBaseDao(&baseConf), conf: conf, project: project}
}

func sharedFeatureFields(fields string) []string {
	if fields == "" || fields == "*" {
		return []string{"*"}
	}
	return strings.Split(fields, ",")
}

// Only inputs that affect the actual SDK query participate in its namespace.
// JSON framing prevents delimiter collisions and SHA-256 bounds cache key size.
func sharedFeatureCacheKey(conf recconf.FeatureDaoConfig, path, key string, appendKeys []string) string {
	namespace := struct {
		Store, Path, View, Model, Entity, Scope string
		Fields                                  []string
		DlrmHSTU                                bool
	}{Store: conf.FeatureStoreName, Path: path, Scope: conf.FeatureStore}
	if path == "view" {
		namespace.View = conf.FeatureStoreViewName
		if conf.FeatureStore == Feature_Store_User {
			namespace.Fields = sharedFeatureFields(conf.UserSelectFields)
			namespace.DlrmHSTU = conf.FeatureStoreOptions.DlrmHSTU
		} else {
			namespace.Fields = sharedFeatureFields(conf.ItemSelectFields)
		}
	} else {
		namespace.Model, namespace.Entity = conf.FeatureStoreModelName, conf.FeatureStoreEntityName
		if path == "model" && conf.FeatureStore == Feature_Store_User {
			namespace.DlrmHSTU = conf.FeatureStoreOptions.DlrmHSTU
		}
	}
	data, _ := json.Marshal(namespace)
	signature := fmt.Sprintf("%x", sha256.Sum256(data))
	entry := struct {
		Namespace, Key string
		AppendKeys     []string
	}{Namespace: signature, Key: key}
	if path == "aggregated" {
		entry.AppendKeys = appendKeys
	}
	data, _ = json.Marshal(entry)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func (d *SharedFeatureFeatureStoreDao) FeatureFetch(user *User, items []*Item, ctx *reccontext.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=SharedFeatureFeatureStoreDao\terror=%v", ctx.RecommendId, err))
		}
	}()
	// Resolve exactly once per invocation. A late write must not populate a
	// newly loaded instance with data fetched under the previous policy.
	c, _ := fs.GetFeatureCache(d.conf.FeatureCacheName)
	var err error
	if d.featureStore == Feature_Store_User {
		err = d.userFeatureFetch(user, ctx, c)
	} else {
		err = d.itemsFeatureFetch(items, ctx, c)
	}
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=SharedFeatureFeatureStoreDao\terror=%v", ctx.RecommendId, err))
	}
}

func (d *SharedFeatureFeatureStoreDao) userFeatureFetch(user *User, ctx *reccontext.RecommendContext, c *fs.FeatureCache) error {
	parts := strings.Split(d.featureKey, ":")
	if len(parts) < 2 {
		return fmt.Errorf("featureKey error(%s)", d.featureKey)
	}
	key := user.StringProperty(parts[1])
	if key == "" {
		return fmt.Errorf("property not found(%s)", parts[1])
	}
	var appendKeys []string
	if d.featureAppendKey != "" {
		parts := strings.Split(d.featureAppendKey, ":")
		if len(parts) < 2 {
			return fmt.Errorf("featureAppendKey error(%s)", d.featureAppendKey)
		}
		var err error
		appendKeys, err = user.ListStringProperty(parts[1])
		if err != nil {
			return err
		}
	}

	if d.conf.FeatureStoreViewName != "" {
		cacheKey := sharedFeatureCacheKey(d.conf, "view", key, nil)
		return d.loadSharedUser(user, ctx, c, cacheKey, nil, func() (map[string]any, error) {
			view := d.project().GetFeatureView(d.conf.FeatureStoreViewName)
			if view == nil {
				return nil, fmt.Errorf("feature view not found(%s)", d.conf.FeatureStoreViewName)
			}
			features, err := view.GetOnlineFeaturesWithOptions([]any{key}, sharedFeatureFields(d.conf.UserSelectFields), nil,
				domain.FeatureViewOptions{DlrmHSTU: d.conf.FeatureStoreOptions.DlrmHSTU})
			if err != nil {
				return nil, err
			}
			if len(features) == 0 {
				return nil, fmt.Errorf("get features empty, view:%s", d.conf.FeatureStoreViewName)
			}
			return features[0], nil
		})
	}

	project := d.project()
	model := project.GetModel(d.conf.FeatureStoreModelName)
	if model == nil {
		return fmt.Errorf("model not found(%s)", d.conf.FeatureStoreModelName)
	}
	labelTable := model.GetLabelTable()
	if labelTable == nil {
		return fmt.Errorf("label table not found(%s)", model.LabelDatasourceTable)
	}
	entity := project.GetFeatureEntity(d.conf.FeatureStoreEntityName)
	if entity == nil {
		return fmt.Errorf("feature entity not found(%s)", d.conf.FeatureStoreEntityName)
	}
	policy := &sharedModelPolicy{modelFields: make(map[string]bool), labelFields: make(map[string]bool),
		labelPriority: model.GetLabelPriorityLevel(), aggregated: len(appendKeys) > 0}
	for _, name := range labelTable.GetFeatureNames() {
		policy.labelFields[name] = true
	}
	for _, feature := range model.Features {
		name := feature.Name
		if feature.AliasName != "" {
			name = feature.AliasName
		}
		policy.modelFields[name] = true
	}
	path := "model"
	if policy.aggregated {
		path = "aggregated"
	}
	cacheKey := sharedFeatureCacheKey(d.conf, path, key, appendKeys)
	return d.loadSharedUser(user, ctx, c, cacheKey, policy, func() (map[string]any, error) {
		if policy.aggregated {
			keys := []any{key}
			for _, appendKey := range appendKeys {
				keys = append(keys, appendKey)
			}
			features, err := model.GetOnlineFeaturesWithAggregatedSequence(key, keys, d.conf.FeatureStoreEntityName)
			if err != nil {
				return nil, err
			}
			if len(features) == 0 {
				return nil, fmt.Errorf("get aggregated features empty")
			}
			return features, nil
		}
		features, err := model.GetOnlineFeaturesWithEntityWithOptions(map[string][]any{entity.FeatureEntityJoinid: {key}},
			d.conf.FeatureStoreEntityName, domain.ModelOptions{DlrmHSTU: d.conf.FeatureStoreOptions.DlrmHSTU})
		if err != nil {
			return nil, err
		}
		if len(features) == 0 {
			return nil, fmt.Errorf("get model features empty")
		}
		return features[0], nil
	})
}

type sharedModelPolicy struct {
	modelFields   map[string]bool
	labelFields   map[string]bool
	labelPriority int
	aggregated    bool
}

func (p *sharedModelPolicy) deletedFields(contextFeatures map[string]any) []string {
	var deleted []string
	if len(p.labelFields) > 0 {
		for name := range contextFeatures {
			if p.modelFields[name] && !p.labelFields[name] {
				deleted = append(deleted, name)
			}
		}
	}
	return deleted
}

func (p *sharedModelPolicy) merge(features map[string]any, contextFeatures map[string]any, deleted []string) {
	if p.labelPriority == 1 {
		for name, value := range contextFeatures {
			if p.modelFields[name] && p.labelFields[name] {
				features[name] = value
			}
		}
	}
	// The legacy aggregated path does not backfill missing model-only fields.
	if !p.aggregated {
		for _, name := range deleted {
			if _, exists := features[name]; !exists {
				features[name] = contextFeatures[name]
			}
		}
	}
}

// loadSharedUser keeps request-specific deletion and merging outside the cache.
func (d *SharedFeatureFeatureStoreDao) loadSharedUser(user *User, ctx *reccontext.RecommendContext, c *fs.FeatureCache,
	key string, policy *sharedModelPolicy, load func() (map[string]any, error)) error {
	contextFeatures, _ := ctx.GetParameter("features").(map[string]any)
	var deleted []string
	if policy != nil {
		deleted = policy.deletedFields(contextFeatures)
		user.DeleteProperties(deleted)
	}
	features, hit := c.Get(key)
	if !hit {
		raw, err := load()
		if err != nil {
			return err
		}
		if raw == nil {
			return fmt.Errorf("get features empty")
		}
		c.Put(key, raw)
		features = fs.CloneFeatureProperties(raw)
	}
	if policy != nil {
		policy.merge(features, contextFeatures, deleted)
	}
	if d.cacheFeaturesName != "" {
		user.AddCacheFeatures(d.cacheFeaturesName, features)
	} else {
		user.AddProperties(features)
	}
	if hit && ctx.Debug {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=SharedFeatureFeatureStoreDao\tmsg=hit cache(%s)", ctx.RecommendId, key))
	}
	return nil
}

func (d *SharedFeatureFeatureStoreDao) itemsFeatureFetch(items []*Item, ctx *reccontext.RecommendContext, c *fs.FeatureCache) error {
	if len(items) == 0 {
		return nil
	}
	field := d.featureKey
	if field != "item:id" {
		parts := strings.Split(field, ":")
		if len(parts) < 2 {
			return fmt.Errorf("featureKey error(%s)", field)
		}
		field = parts[1]
	}
	path := "model"
	if d.conf.FeatureStoreViewName != "" {
		path = "view"
	}
	key2Items := make(map[string][]*Item, len(items))
	var keys []any
	for _, item := range items {
		key := string(item.Id)
		if field != "item:id" {
			key = item.StringProperty(field)
		}
		if features, hit := c.Get(sharedFeatureCacheKey(d.conf, path, key, nil)); hit {
			item.AddProperties(features)
			if ctx.Debug {
				item.AddProperty("__debug_cache_hit__", true)
			}
			continue
		}
		if _, exists := key2Items[key]; !exists {
			keys = append(keys, key)
		}
		key2Items[key] = append(key2Items[key], item)
	}
	if len(keys) == 0 {
		return nil
	}

	project := d.project()
	var features []map[string]any
	var joinID string
	var err error
	if path == "view" {
		view := project.GetFeatureView(d.conf.FeatureStoreViewName)
		if view == nil {
			return fmt.Errorf("feature view not found(%s)", d.conf.FeatureStoreViewName)
		}
		features, err = view.GetOnlineFeatures(keys, sharedFeatureFields(d.conf.ItemSelectFields), nil)
		if err != nil {
			return err
		}
		entity := project.GetFeatureEntity(view.GetFeatureEntityName())
		if entity == nil {
			return fmt.Errorf("feature entity not found(%s)", view.GetFeatureEntityName())
		}
		joinID = entity.FeatureEntityJoinid
	} else {
		entity := project.GetFeatureEntity(d.conf.FeatureStoreEntityName)
		if entity == nil {
			return fmt.Errorf("feature entity not found(%s)", d.conf.FeatureStoreEntityName)
		}
		model := project.GetModel(d.conf.FeatureStoreModelName)
		if model == nil {
			return fmt.Errorf("model not found(%s)", d.conf.FeatureStoreModelName)
		}
		features, err = model.GetOnlineFeaturesWithEntity(map[string][]any{entity.FeatureEntityJoinid: keys}, d.conf.FeatureStoreEntityName)
		if err != nil {
			return err
		}
		joinID = entity.FeatureEntityJoinid
	}
	// Match only returned join IDs, without converting their types or merging
	// pre-existing item attributes into the cached SDK rows.
	found := make(map[string]bool, len(features))
	for _, row := range features {
		key, ok := row[joinID].(string)
		if !ok {
			continue
		}
		consumers := key2Items[key]
		if len(consumers) == 0 {
			continue
		}
		found[key] = true
		c.Put(sharedFeatureCacheKey(d.conf, path, key, nil), row)
		for _, item := range consumers {
			item.AddProperties(fs.CloneFeatureProperties(row))
		}
	}
	for key := range key2Items {
		if !found[key] {
			c.PutIfAbsent(sharedFeatureCacheKey(d.conf, path, key, nil), map[string]any{})
		}
	}
	return nil
}
