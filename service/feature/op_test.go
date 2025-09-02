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
