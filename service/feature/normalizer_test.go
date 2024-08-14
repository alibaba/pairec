package feature

import (
	"fmt"
	"testing"

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
