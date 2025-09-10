package feature

import (
	"fmt"
	"math"
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func TestCreateConstValueNormalizer(t *testing.T) {

	user := module.NewUser("123")
	conf := recconf.FeatureLoadConfig{}
	conf.Features = append(conf.Features, recconf.FeatureConfig{
		FeatureType:  "new_feature",
		FeatureStore: "user",
		Normalizer:   "hour_in_day",
		FeatureName:  "hour",
	}, recconf.FeatureConfig{
		FeatureType:  "new_feature",
		FeatureStore: "user",
		Normalizer:   "const_value",
		FeatureValue: "test",
		FeatureName:  "cus_fea",
	})

	feature := LoadWithConfig(conf)
	feature.LoadFeatures(user, nil, context.NewRecommendContext())

	fmt.Println(user.Properties)
	if user.StringProperty("cus_fea") != "test" {
		t.Error("const value featrue erro")
	}

}

func TestExpressionNormalizer(t *testing.T) {
	normalizer := NewExpressionNormalizer("recall_name in ('retarget_u2i','realtime_retarget_click')")
	result := normalizer.Apply(map[string]interface{}{"recall_name": "mind"})
	if val, ok := result.(bool); !ok || val {
		t.Fatalf("result error, type:%T, value:%v", result, result)
	}
	t.Log(result)
	result = normalizer.Apply(map[string]interface{}{"recall_name": "retarget_u2i"})
	if val, ok := result.(bool); !ok || !val {
		t.Fatalf("result error, type:%T, value:%v", result, result)
	}
	t.Log(result)
	normalizer = NewExpressionNormalizer("getString(a, b)")
	str := normalizer.Apply(map[string]interface{}{"a": "mind", "b": "other"})
	t.Log(str)
	normalizer = NewExpressionNormalizer("a")
	str = normalizer.Apply(map[string]interface{}{})
	t.Log(str)
}

func TestCreateMonthNormalizer(t *testing.T) {

	user := module.NewUser("123")
	conf := recconf.FeatureLoadConfig{}
	conf.Features = append(conf.Features,
		recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "user",
			Normalizer:   "month",
			FeatureName:  "month",
		},
		recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "user",
			Normalizer:   "week",
			FeatureName:  "week",
		},
		recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "user",
			Normalizer:   "hour_in_day",
			FeatureName:  "hour",
		},
		recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "user",
			Normalizer:   "weekday",
			FeatureName:  "weekday",
		},
	)

	feature := LoadWithConfig(conf)
	feature.LoadFeatures(user, nil, context.NewRecommendContext())

	t.Log(user.Properties)

}

func TestExpressionFunctionNormalizer(t *testing.T) {
	t.Run("max function", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("a > b ? a : b")
		result := normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, float64(10))
		// use max func
		normalizer = NewExpressionNormalizer("max(a, b)")
		result = normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, float64(10))

	})
	t.Run("min function", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("a > b ? b : a")
		result := normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, float64(8))

		// use min func
		normalizer = NewExpressionNormalizer("min(a, b)")
		result = normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, float64(8))
	})
	t.Run("log function", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("log(a)")
		result := normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, math.Log(10))
	})
	t.Run("log10 function", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("log10(a)")
		result := normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, math.Log10(10))
	})
	t.Run("log2 function", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("log2(a)")
		result := normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, math.Log2(10))
	})
	t.Run("pow function", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("pow(a, b)")
		result := normalizer.Apply(map[string]interface{}{"a": 2, "b": 3})
		assert.Equal(t, result, float64(8))
	})
}
