package feature

import (
	"fmt"
	"math"
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/golang/geo/s2"
	"github.com/mmcloughlin/geohash"
	"github.com/spaolacci/murmur3"
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
	user.AddProperties(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})
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
		recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "user",
			Normalizer:   "expression",
			FeatureName:  "cellID",
			Expression:   "s2CellID(lat, lng)",
		},
		recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "user",
			Normalizer:   "expression",
			FeatureName:  "geoHash",
			Expression:   "geoHash(lat, lng)",
		},
	)

	feature := LoadWithConfig(conf)
	feature.LoadFeatures(user, nil, context.NewRecommendContext())

	assert.Equal(t, user.GetProperty("cellID"), 3886697436164390912)
	assert.Equal(t, user.StringProperty("geoHash"), "wx4g0b")
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
	t.Run("string contact", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("'hello ' + a")
		result := normalizer.Apply(map[string]interface{}{"a": "world", "b": 3})
		assert.Equal(t, result, "hello world")

		key := result.(string)
		normalizer = NewExpressionNormalizer("hash32(key) % 100")
		result = normalizer.Apply(map[string]interface{}{"key": key})
		assert.Equal(t, result, float64(murmur3.Sum32([]byte(key))%100))
	})
	t.Run("s2CellID", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("s2CellID(lat, lng)")
		result := normalizer.Apply(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})

		lat := 39.9042  // 纬度
		lng := 116.4074 // 经度

		ll := s2.LatLngFromDegrees(lat, lng)

		cellID := s2.CellIDFromLatLng(ll)

		level := 15
		cellIDAtLevel := cellID.Parent(level)

		assert.Equal(t, utils.ToInt(result, 1), utils.ToInt(uint64(cellIDAtLevel), 0))

		normalizer = NewExpressionNormalizer("s2CellID(lat, lng, 20)")
		result = normalizer.Apply(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})

		level = 20
		cellIDAtLevel = cellID.Parent(level)

		assert.Equal(t, utils.ToInt(result, 1), utils.ToInt(uint64(cellIDAtLevel), 0))
	})
	t.Run("geoHash", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("geoHash(lat, lng)")
		result := normalizer.Apply(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})

		lat := 39.9042  // 纬度
		lng := 116.4074 // 经度

		hashResult := geohash.EncodeWithPrecision(lat, lng, 6)
		assert.Equal(t, result, hashResult)

		normalizer = NewExpressionNormalizer("geoHash(lat, lng, 12)")
		result = normalizer.Apply(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})
		hashResult = geohash.EncodeWithPrecision(lat, lng, 12)
		assert.Equal(t, result, hashResult)
	})
	t.Run("geoHashWithNeighbors", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("geoHashWithNeighbors(lat, lng)")
		result := normalizer.Apply(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})

		lat := 39.9042  // 纬度
		lng := 116.4074 // 经度

		hashResult := geohash.EncodeWithPrecision(lat, lng, 6)
		assert.Equal(t, result.([]string)[8], hashResult)

		neighbors := geohash.Neighbors(hashResult)
		assert.Equal(t, result.([]string)[:8], neighbors)
	})
	t.Run("haversine", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("haversine(lng1, lat1, lng2, lat2)")
		result := normalizer.Apply(map[string]interface{}{"lat1": 39.9042, "lng1": 116.4074, "lat2": 31.2304, "lng2": 121.4737})

		assert.Equal(t, utils.ToInt(result, 0), 1067)
	})
	t.Run("sphereDistance", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("sphereDistance(lng1, lat1, lng2, lat2)")
		result := normalizer.Apply(map[string]interface{}{"lat1": 39.9042, "lng1": 116.4074, "lat2": 31.2304, "lng2": 121.4737})

		assert.Equal(t, utils.ToInt(result, 0), 1067)
	})
}
