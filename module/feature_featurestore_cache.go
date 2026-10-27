package module

import (
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
	cacheName string
	initial   sharedFeatureCache
	resolve   func(string) (sharedFeatureCache, error)
}

var _ FeatureDao = (*SharedFeatureFeatureStoreDao)(nil)

// The instance supplies both cache operations and its immutable source snapshot.
// The narrow interface also permits isolated DAO tests without SDK servers.
type sharedFeatureCache interface {
	Config() recconf.FeatureCacheConfig
	Project() *domain.Project
	Get(string) (map[string]any, bool)
	Put(string, map[string]any)
	PutIfAbsent(string, map[string]any)
}

func NewSharedFeatureFeatureStoreDao(conf recconf.FeatureDaoConfig) *SharedFeatureFeatureStoreDao {
	return newSharedFeatureFeatureStoreDao(conf, func(name string) (sharedFeatureCache, error) {
		c, err := fs.GetFeatureCache(name)
		if err != nil {
			return nil, err
		}
		return c, nil
	})
}

func validateSharedFeatureConfig(conf recconf.FeatureDaoConfig) {
	if strings.TrimSpace(conf.FeatureCacheName) == "" {
		panic("FeatureCacheName must not be blank")
	}
	if conf.AdapterType != recconf.DataSource_Type_FeatureStore {
		panic("FeatureCacheName is only supported by the featurestore adapter")
	}
	if conf.FeatureAppendKey != "" {
		panic("FeatureAppendKey cannot be used with FeatureCacheName")
	}
	if conf.FeatureStore != Feature_Store_User && conf.FeatureStore != Feature_Store_Item {
		panic("shared FeatureStore must be user or item")
	}
	if conf.FeatureStore == Feature_Store_User && conf.FeatureType == Feature_Type_Sequence {
		panic("shared sequence_feature loading is not implemented; use the view/model path")
	}
	if conf.FeatureStoreName != "" || conf.FeatureStoreViewName != "" || conf.FeatureStoreModelName != "" || conf.FeatureStoreEntityName != "" || conf.UserSelectFields != "" || conf.ItemSelectFields != "" || conf.FeatureStoreOptions != (recconf.FeatureStoreOptionsConfig{}) {
		panic("shared data source, fields and options must be configured in FeatureCacheConfs")
	}
}

func newSharedFeatureFeatureStoreDao(conf recconf.FeatureDaoConfig, resolve func(string) (sharedFeatureCache, error)) *SharedFeatureFeatureStoreDao {
	validateSharedFeatureConfig(conf)
	c, err := resolve(conf.FeatureCacheName)
	if err != nil {
		panic(err)
	}
	if c.Config().FeatureStore != conf.FeatureStore {
		panic("FeatureStore does not match the named cache definition")
	}
	baseConf := conf
	baseConf.CacheSize = 0
	return &SharedFeatureFeatureStoreDao{FeatureBaseDao: NewFeatureBaseDao(&baseConf), cacheName: conf.FeatureCacheName, initial: c, resolve: resolve}
}

func sharedFeatureFields(fields string) []string {
	if fields == "" || fields == "*" {
		return []string{"*"}
	}
	return strings.Split(fields, ",")
}

func (d *SharedFeatureFeatureStoreDao) FeatureFetch(user *User, items []*Item, ctx *reccontext.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=SharedFeatureFeatureStoreDao\terror=%v", ctx.RecommendId, err))
		}
	}()
	// Resolve exactly once per invocation. A late write must not populate a
	// newly loaded instance with data fetched under the previous policy.
	c, resolveErr := d.resolve(d.cacheName)
	source := c
	if resolveErr != nil {
		c = nil
		source = d.initial // Direct read for an active DAO whose name was removed.
	}
	conf := source.Config()
	if conf.FeatureStore != d.featureStore {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=SharedFeatureFeatureStoreDao\terror=cache scope changed", ctx.RecommendId))
		return
	}
	var err error
	if d.featureStore == Feature_Store_User {
		err = d.userFeatureFetch(user, ctx, c, conf, source.Project)
	} else {
		err = d.itemsFeatureFetch(items, ctx, c, conf, source.Project)
	}
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=SharedFeatureFeatureStoreDao\terror=%v", ctx.RecommendId, err))
	}
}

func (d *SharedFeatureFeatureStoreDao) userFeatureFetch(user *User, ctx *reccontext.RecommendContext, c sharedFeatureCache, conf recconf.FeatureCacheConfig, getProject func() *domain.Project) error {
	parts := strings.Split(d.featureKey, ":")
	if len(parts) < 2 {
		return fmt.Errorf("featureKey error(%s)", d.featureKey)
	}
	key := user.StringProperty(parts[1])
	if key == "" {
		return fmt.Errorf("property not found(%s)", parts[1])
	}
	if conf.FeatureStoreViewName != "" {
		return d.loadSharedUser(user, ctx, c, key, nil, func() (map[string]any, error) {
			view := getProject().GetFeatureView(conf.FeatureStoreViewName)
			if view == nil {
				return nil, fmt.Errorf("feature view not found(%s)", conf.FeatureStoreViewName)
			}
			features, err := view.GetOnlineFeaturesWithOptions([]any{key}, sharedFeatureFields(conf.SelectFields), nil,
				domain.FeatureViewOptions{DlrmHSTU: conf.FeatureStoreOptions.DlrmHSTU})
			if err != nil {
				return nil, err
			}
			if len(features) == 0 {
				return nil, fmt.Errorf("get features empty, view:%s", conf.FeatureStoreViewName)
			}
			return features[0], nil
		})
	}

	project := getProject()
	model := project.GetModel(conf.FeatureStoreModelName)
	if model == nil {
		return fmt.Errorf("model not found(%s)", conf.FeatureStoreModelName)
	}
	labelTable := model.GetLabelTable()
	if labelTable == nil {
		return fmt.Errorf("label table not found(%s)", model.LabelDatasourceTable)
	}
	entity := project.GetFeatureEntity(conf.FeatureStoreEntityName)
	if entity == nil {
		return fmt.Errorf("feature entity not found(%s)", conf.FeatureStoreEntityName)
	}
	policy := &sharedModelPolicy{modelFields: make(map[string]bool), labelFields: make(map[string]bool),
		labelPriority: model.GetLabelPriorityLevel()}
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
	return d.loadSharedUser(user, ctx, c, key, policy, func() (map[string]any, error) {
		features, err := model.GetOnlineFeaturesWithEntityWithOptions(map[string][]any{entity.FeatureEntityJoinid: {key}},
			conf.FeatureStoreEntityName, domain.ModelOptions{DlrmHSTU: conf.FeatureStoreOptions.DlrmHSTU})
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
	for _, name := range deleted {
		if _, exists := features[name]; !exists {
			features[name] = contextFeatures[name]
		}
	}
}

// loadSharedUser keeps request-specific deletion and merging outside the cache.
func (d *SharedFeatureFeatureStoreDao) loadSharedUser(user *User, ctx *reccontext.RecommendContext, c sharedFeatureCache,
	key string, policy *sharedModelPolicy, load func() (map[string]any, error)) error {
	contextFeatures, _ := ctx.GetParameter("features").(map[string]any)
	var deleted []string
	if policy != nil {
		deleted = policy.deletedFields(contextFeatures)
		user.DeleteProperties(deleted)
	}
	var features map[string]any
	var hit bool
	if c != nil {
		features, hit = c.Get(key)
	}
	if !hit {
		raw, err := load()
		if err != nil {
			return err
		}
		if raw == nil {
			return fmt.Errorf("get features empty")
		}
		if c != nil {
			c.Put(key, raw)
		}
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

func (d *SharedFeatureFeatureStoreDao) itemsFeatureFetch(items []*Item, ctx *reccontext.RecommendContext, c sharedFeatureCache, conf recconf.FeatureCacheConfig, getProject func() *domain.Project) error {
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
	if conf.FeatureStoreViewName != "" {
		path = "view"
	}
	key2Items := make(map[string][]*Item, len(items))
	var keys []any
	for _, item := range items {
		key := string(item.Id)
		if field != "item:id" {
			key = item.StringProperty(field)
		}
		if c != nil {
			if features, hit := c.Get(key); hit {
				item.AddProperties(features)
				if ctx.Debug {
					item.AddProperty("__debug_cache_hit__", true)
				}
				continue
			}
		}
		if _, exists := key2Items[key]; !exists {
			keys = append(keys, key)
		}
		key2Items[key] = append(key2Items[key], item)
	}
	if len(keys) == 0 {
		return nil
	}

	project := getProject()
	var features []map[string]any
	var joinID string
	var err error
	if path == "view" {
		view := project.GetFeatureView(conf.FeatureStoreViewName)
		if view == nil {
			return fmt.Errorf("feature view not found(%s)", conf.FeatureStoreViewName)
		}
		features, err = view.GetOnlineFeatures(keys, sharedFeatureFields(conf.SelectFields), nil)
		if err != nil {
			return err
		}
		entity := project.GetFeatureEntity(view.GetFeatureEntityName())
		if entity == nil {
			return fmt.Errorf("feature entity not found(%s)", view.GetFeatureEntityName())
		}
		joinID = entity.FeatureEntityJoinid
	} else {
		entity := project.GetFeatureEntity(conf.FeatureStoreEntityName)
		if entity == nil {
			return fmt.Errorf("feature entity not found(%s)", conf.FeatureStoreEntityName)
		}
		model := project.GetModel(conf.FeatureStoreModelName)
		if model == nil {
			return fmt.Errorf("model not found(%s)", conf.FeatureStoreModelName)
		}
		features, err = model.GetOnlineFeaturesWithEntity(map[string][]any{entity.FeatureEntityJoinid: keys}, conf.FeatureStoreEntityName)
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
		if c != nil {
			c.Put(key, row)
		}
		for _, item := range consumers {
			item.AddProperties(fs.CloneFeatureProperties(row))
		}
	}
	for key := range key2Items {
		if !found[key] && c != nil {
			c.PutIfAbsent(key, map[string]any{})
		}
	}
	return nil
}
