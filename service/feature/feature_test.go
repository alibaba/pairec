package feature

import (
	"fmt"
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func TestCreateNewFeatureOp(t *testing.T) {
	item1 := module.NewItem("123")
	item1.RetrieveId = "mind"
	item2 := module.NewItem("456")
	item2.RetrieveId = "realtime_retarget_click"
	var items []*module.Item
	items = append(items, item1, item2)

	t.Run("use expression normalizer", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "item",
			//FeatureSource: "item:recall_name",
			Normalizer:  "expression",
			Expression:  "recall_name in ('retarget_u2i','realtime_retarget_click')",
			FeatureName: "is_retarget",
		})

		feature := LoadWithConfig(conf)
		feature.LoadFeatures(nil, items, context.NewRecommendContext())

		for _, item := range items {
			fmt.Println(item.Id, item.GetProperties())
			if item.Id == "123" && item.GetProperty("is_retarget") != 0 {
				t.Errorf("itemid:%s create new feature fail", item.Id)
			}

			if item.Id == "456" && item.GetProperty("is_retarget") != 1 {
				t.Errorf("itemid:%s create new feature fail", item.Id)
			}
		}
	})
	t.Run("use expr normalizer", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "item",
			Normalizer:   "expr",
			Expression:   "item.recall_name in ['retarget_u2i','realtime_retarget_click']",
			FeatureName:  "is_retarget",
		})

		feature := LoadWithConfig(conf)
		feature.LoadFeatures(nil, items, context.NewRecommendContext())

		for _, item := range items {
			fmt.Println(item.Id, item.GetProperties())
			if item.Id == "123" && item.GetProperty("is_retarget") != 0 {
				t.Errorf("itemid:%s create new feature fail", item.Id)
			}

			if item.Id == "456" && item.GetProperty("is_retarget") != 1 {
				t.Errorf("itemid:%s create new feature fail", item.Id)
			}
		}
	})

}

func TestCreateNewFeatureOp2(t *testing.T) {
	item1 := module.NewItem("123")
	item1.RetrieveId = "mind"
	item1.AddProperty("ali_recall_name", "")
	item2 := module.NewItem("456")
	item2.RetrieveId = "realtime_retarget_click"
	item2.AddProperty("ali_recall_name", "dssm")
	var items []*module.Item
	items = append(items, item1, item2)

	t.Run("use expression normalizer", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "item",
			Normalizer:   "expression",
			Expression:   "getString(ali_recall_name,  recall_name)",
			FeatureName:  "retrived",
		})

		feature := LoadWithConfig(conf)
		feature.LoadFeatures(nil, items, context.NewRecommendContext())

		for _, item := range items {
			if item.Id == "123" {
				assert.Equal(t, item.StringProperty("retrived"), "mind")
			} else if item.Id == "456" {
				assert.Equal(t, item.StringProperty("retrived"), "dssm")
			}
		}
	})
	t.Run("use expr normalizer", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "item",
			Normalizer:   "expr",
			Expression:   "getString(item.ali_recall_name,  item.recall_name)",
			FeatureName:  "retrived",
		})

		feature := LoadWithConfig(conf)
		feature.LoadFeatures(nil, items, context.NewRecommendContext())

		for _, item := range items {
			if item.Id == "123" {
				assert.Equal(t, item.StringProperty("retrived"), "mind")
			} else if item.Id == "456" {
				assert.Equal(t, item.StringProperty("retrived"), "dssm")
			}
		}
	})

}

func TestCreateNewFeatureOp3(t *testing.T) {
	item1 := module.NewItem("123")
	item1.RetrieveId = "mind"
	item1.AddProperty("ali_recall_name", "")
	item1.AddProperty("index", 5)
	item2 := module.NewItem("456")
	item2.RetrieveId = "realtime_retarget_click"
	item2.AddProperty("ali_recall_name", "dssm")
	item2.AddProperty("index", 8)
	var items []*module.Item
	items = append(items, item1, item2)

	user := module.NewUser("test")
	user.AddProperty("user_index", 13)
	t.Run("use expr normalizer", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "item",
			Normalizer:   "expr",
			Expression:   "user.user_index - item.index",
			FeatureName:  "index_delta",
		})

		feature := LoadWithConfig(conf)
		feature.LoadFeatures(user, items, context.NewRecommendContext())

		for _, item := range items {
			if item.Id == "123" {
				assert.Equal(t, item.GetProperty("index_delta"), 8)
			} else if item.Id == "456" {
				assert.Equal(t, item.GetProperty("index_delta"), 5)
			}
		}
	})

}
