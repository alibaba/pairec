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
		// an integer keeps an integer type, only a real decimal becomes a float64
		assert.Equal(t, user.GetProperty("age"), 30)
		assert.Equal(t, user.GetProperty("score"), float64(1.5))
		assert.Equal(t, user.GetProperty("tags"), []string{"a", "b"})
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

	t.Run("expand user snapshot to every item", func(t *testing.T) {
		user := module.NewUser("user1")
		user.AddProperty("user_features", `{"gender":"male","age":30}`)
		items := []*module.Item{module.NewItem("1"), module.NewItem("2"), module.NewItem("3")}

		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:         "expand_json_feature",
			FeatureStore:        "item",
			FeatureSource:       "user:user_features",
			RemoveFeatureSource: true,
		})

		feature := LoadWithConfig(conf)
		feature.LoadFeatures(user, items, context.NewRecommendContext())

		// the user property is shared by every item, so all of them are expanded
		// and RemoveFeatureSource does not drop it after the first item
		for _, item := range items {
			assert.Equal(t, item.StringProperty("gender"), "male")
			assert.Equal(t, item.GetProperty("age"), 30)
		}
		assert.Equal(t, user.StringProperty("user_features"), `{"gender":"male","age":30}`)
	})

	t.Run("keep the precision of an id beyond the float64 mantissa", func(t *testing.T) {
		user := module.NewUser("user1")
		user.AddProperty("user_features", `{"uid":9007199254740993,"big_id":1234567890123456789}`)

		feature := LoadWithConfig(userConf(true))
		feature.LoadFeatures(user, nil, context.NewRecommendContext())

		// both values are above 2^53, a float64 round trip would alter them
		assert.Equal(t, user.GetProperty("uid"), 9007199254740993)
		assert.Equal(t, user.GetProperty("big_id"), 1234567890123456789)
	})

	t.Run("restore list features by their json element type", func(t *testing.T) {
		user := module.NewUser("user1")
		user.AddProperty("user_features", `{"tags":["a","b"],"ids":["101","102"],"seq":[101,102],"scores":[1.5,2.5],"nested":[["a"],["b","c"]],"nested_num":[[1.5],[2.5]],"mixed":[1,"a"],"empty":[]}`)

		feature := LoadWithConfig(userConf(true))
		feature.LoadFeatures(user, nil, context.NewRecommendContext())

		// a quoted element stays a string even when it holds only digits, which is
		// what keeps a []string snapshot from turning into a numeric list
		assert.Equal(t, user.GetProperty("tags"), []string{"a", "b"})
		assert.Equal(t, user.GetProperty("ids"), []string{"101", "102"})
		// a bare number becomes a typed numeric list the request builder carries
		assert.Equal(t, user.GetProperty("seq"), []int{101, 102})
		assert.Equal(t, user.GetProperty("scores"), []float64{1.5, 2.5})
		assert.Equal(t, user.GetProperty("nested"), [][]string{{"a"}, {"b", "c"}})
		assert.Equal(t, user.GetProperty("nested_num"), [][]float64{{1.5}, {2.5}})
		// the builder has no type for these two, only the elements are restored
		assert.Equal(t, user.GetProperty("mixed"), []interface{}{1, "a"})
		assert.Equal(t, user.GetProperty("empty"), []interface{}{})
	})

	t.Run("restore map features by their json value type", func(t *testing.T) {
		user := module.NewUser("user1")
		user.AddProperty("user_features", `{"cate":{"a":"x","b":"y"},"cnt":{"a":1,"b":9007199254740993},"rate":{"a":1.5},"mixed":{"a":1,"b":"x"},"empty":{},"deep":{"a":{"b":1}}}`)

		feature := LoadWithConfig(userConf(true))
		feature.LoadFeatures(user, nil, context.NewRecommendContext())

		// the json object keys are restored as strings, the value type decides the map type
		assert.Equal(t, user.GetProperty("cate"), map[string]string{"a": "x", "b": "y"})
		// int64 and not int, the builder casts a map[string]int down to int32 unchecked
		assert.Equal(t, user.GetProperty("cnt"), map[string]int64{"a": 1, "b": 9007199254740993})
		assert.Equal(t, user.GetProperty("rate"), map[string]float64{"a": 1.5})
		// the builder has no type for these, only the values are restored
		assert.Equal(t, user.GetProperty("mixed"), map[string]interface{}{"a": 1, "b": "x"})
		assert.Equal(t, user.GetProperty("empty"), map[string]interface{}{})
		assert.Equal(t, user.GetProperty("deep"), map[string]interface{}{"a": map[string]int64{"b": 1}})
	})
}
