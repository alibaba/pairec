package feature

import (
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/cespare/xxhash/v2"
)

func TestComposeFeatureOp(t *testing.T) {
	item1 := module.NewItem("123")
	item1.RetrieveId = "mind"
	item1.AddProperty("category", "movie")
	item2 := module.NewItem("456")
	item2.RetrieveId = "realtime_retarget_click"
	item2.AddProperty("category", "book")
	var items []*module.Item
	items = append(items, item1, item2)

	user := module.NewUser("user1")
	user.AddProperties(map[string]interface{}{"gender": "male", "uid": "user1"})

	t.Run("test_compose_feature_op", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:   "compose_feature",
			FeatureStore:  "item",
			FeatureSource: "user:gender,item:category",
			FeatureName:   "compose_feature",
		})

		feature := LoadWithConfig(conf)
		feature.LoadFeatures(user, items, context.NewRecommendContext())

		assert.Equal(t, items[0].StringProperty("compose_feature"), "compose_feature_male_movie")
		assert.Equal(t, items[1].StringProperty("compose_feature"), "compose_feature_male_book")

	})
	t.Run("test_compose_feature_op with item id", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:   "compose_feature",
			FeatureStore:  "item",
			FeatureSource: "user:uid,item:id",
			FeatureName:   "compose_feature",
		})

		feature := LoadWithConfig(conf)
		feature.LoadFeatures(user, items, context.NewRecommendContext())

		assert.Equal(t, items[0].StringProperty("compose_feature"), "compose_feature_user1_123")
		assert.Equal(t, items[1].StringProperty("compose_feature"), "compose_feature_user1_456")
	})
	t.Run("test_compose_feature_op with change item feature name", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:   "raw_feature",
			FeatureStore:  "item",
			FeatureSource: "item:compose_feature",
			FeatureName:   "newid",
		})

		feature := LoadWithConfig(conf)
		feature.LoadFeatures(user, items, context.NewRecommendContext())

		assert.Equal(t, items[0].StringProperty("newid"), "compose_feature_user1_123")
		assert.Equal(t, items[1].StringProperty("newid"), "compose_feature_user1_456")

		assert.Equal(t, items[0].StringProperty("compose_feature"), "compose_feature_user1_123")
		assert.Equal(t, items[1].StringProperty("compose_feature"), "compose_feature_user1_456")
	})

}

func TestNewFeatureWithExpressionFeatureOp(t *testing.T) {
	item1 := module.NewItem("123")
	item1.RetrieveId = "mind"
	item1.AddProperty("category", "movie")
	item2 := module.NewItem("456")
	item2.RetrieveId = "realtime_retarget_click"
	item2.AddProperty("category", "book")
	var items []*module.Item
	items = append(items, item1, item2)

	user := module.NewUser("user1")
	user.AddProperties(map[string]interface{}{"gender": "male", "uid": "user1"})

	conf := recconf.FeatureLoadConfig{}
	conf.Features = append(conf.Features, recconf.FeatureConfig{
		FeatureType:   "compose_feature",
		FeatureStore:  "item",
		FeatureSource: "user:uid,item:id",
		FeatureName:   "compose_feature",
	})

	feature := LoadWithConfig(conf)
	feature.LoadFeatures(user, items, context.NewRecommendContext())

	assert.Equal(t, items[0].StringProperty("compose_feature"), "compose_feature_user1_123")
	assert.Equal(t, items[1].StringProperty("compose_feature"), "compose_feature_user1_456")
	t.Run("test_new_feature_with_expression", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:   "new_feature",
			FeatureStore:  "item",
			FeatureSource: "item:compose_feature",
			FeatureName:   "trim_feature",
			Normalizer:    "expression",
			Expression:    "trimPrefix(compose_feature, 'compose_feature_')",
		})

		feature := LoadWithConfig(conf)
		feature.LoadFeatures(user, items, context.NewRecommendContext())

		assert.Equal(t, items[0].StringProperty("trim_feature"), "user1_123")
		assert.Equal(t, items[1].StringProperty("trim_feature"), "user1_456")

	})
	t.Run("test_new_feature_with_expression", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:   "new_feature",
			FeatureStore:  "item",
			FeatureSource: "item:compose_feature",
			FeatureName:   "replace_feature",
			Normalizer:    "expression",
			Expression:    "replace(trimPrefix(compose_feature, 'compose_feature_'), '_', '#')",
		})

		feature := LoadWithConfig(conf)
		feature.LoadFeatures(user, items, context.NewRecommendContext())

		assert.Equal(t, items[0].StringProperty("replace_feature"), "user1#123")
		assert.Equal(t, items[1].StringProperty("replace_feature"), "user1#456")

	})
	t.Run("test_hash_expression", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:   "new_feature",
			FeatureStore:  "item",
			FeatureSource: "item:category",
			FeatureName:   "category_hash",
			Normalizer:    "expression",
			Expression:    "hash(category)",
		})
		feature := LoadWithConfig(conf)
		feature.LoadFeatures(user, items, context.NewRecommendContext())

		assert.Equal(t, items[0].GetProperty("category_hash"), xxhash.Sum64String("movie"))
		assert.Equal(t, items[1].GetProperty("category_hash"), xxhash.Sum64String("book"))
	})

}

func TestExpandJsonFeatureOp(t *testing.T) {
	userConf := func(remove bool) recconf.FeatureLoadConfig {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:         "expand_json_feature",
			FeatureStore:        "user",
			FeatureSource:       "user:user_features",
			RemoveFeatureSource: remove,
		})
		return conf
	}

	t.Run("expand user snapshot and remove source", func(t *testing.T) {
		user := module.NewUser("user1")
		user.AddProperties(map[string]interface{}{
			"user_features": `{"gender":"male","age":30,"score":1.5,"tags":["a","b"]}`,
		})

		feature := LoadWithConfig(userConf(true))
		feature.LoadFeatures(user, nil, context.NewRecommendContext())

		assert.Equal(t, user.StringProperty("gender"), "male")
		// every json number is restored as float64
		assert.Equal(t, user.GetProperty("age"), float64(30))
		assert.Equal(t, user.GetProperty("score"), float64(1.5))
		assert.Equal(t, user.GetProperty("tags"), []interface{}{"a", "b"})
		// RemoveFeatureSource drops the json column itself
		assert.Equal(t, user.GetProperty("user_features"), nil)
	})

	t.Run("keep source when RemoveFeatureSource is false", func(t *testing.T) {
		user := module.NewUser("user1")
		user.AddProperties(map[string]interface{}{
			"user_features": `{"gender":"female"}`,
		})

		feature := LoadWithConfig(userConf(false))
		feature.LoadFeatures(user, nil, context.NewRecommendContext())

		assert.Equal(t, user.StringProperty("gender"), "female")
		assert.Equal(t, user.StringProperty("user_features"), `{"gender":"female"}`)
	})

	t.Run("expanded value overrides existing property", func(t *testing.T) {
		user := module.NewUser("user1")
		user.AddProperties(map[string]interface{}{
			"gender":        "unknown",
			"user_features": `{"gender":"male"}`,
		})

		feature := LoadWithConfig(userConf(true))
		feature.LoadFeatures(user, nil, context.NewRecommendContext())

		assert.Equal(t, user.StringProperty("gender"), "male")
	})

	t.Run("invalid json leaves properties untouched", func(t *testing.T) {
		user := module.NewUser("user1")
		user.AddProperties(map[string]interface{}{
			"user_features": `not a json`,
		})

		feature := LoadWithConfig(userConf(true))
		feature.LoadFeatures(user, nil, context.NewRecommendContext())

		// source is kept so the failure stays diagnosable
		assert.Equal(t, user.StringProperty("user_features"), `not a json`)
	})

	t.Run("missing source is a no-op", func(t *testing.T) {
		user := module.NewUser("user1")
		user.AddProperties(map[string]interface{}{"gender": "male"})

		feature := LoadWithConfig(userConf(true))
		feature.LoadFeatures(user, nil, context.NewRecommendContext())

		assert.Equal(t, user.StringProperty("gender"), "male")
		assert.Equal(t, user.GetProperty("user_features"), nil)
	})

	t.Run("expand item snapshot", func(t *testing.T) {
		item := module.NewItem("123")
		item.AddProperty("item_features", `{"category":"movie","price":9.9}`)
		user := module.NewUser("user1")

		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:         "expand_json_feature",
			FeatureStore:        "item",
			FeatureSource:       "item:item_features",
			RemoveFeatureSource: true,
		})

		feature := LoadWithConfig(conf)
		feature.LoadFeatures(user, []*module.Item{item}, context.NewRecommendContext())

		assert.Equal(t, item.StringProperty("category"), "movie")
		assert.Equal(t, item.GetProperty("price"), float64(9.9))
		assert.Equal(t, item.GetProperty("item_features"), nil)
	})
}
