package feature

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"fortio.org/assert"
	"github.com/Knetic/govaluate"
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
	t.Run("s2CellNeighbors", func(t *testing.T) {
		normalizer := NewExpressionNormalizer("s2CellNeighbors(lat, lng)")
		result := normalizer.Apply(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})
		t.Log(result)

		assert.Equal(t, 9, len(result.([]int)))
		lat := 39.9042  // 纬度
		lng := 116.4074 // 经度

		ll := s2.LatLngFromDegrees(lat, lng)

		cellID := s2.CellIDFromLatLng(ll)

		level := 15
		cellIDAtLevel := cellID.Parent(level)

		t.Log(int(cellIDAtLevel))
		cellIds := result.([]int)
		assert.Equal(t, cellIds[len(cellIds)-1], int(cellIDAtLevel))
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
	t.Run("timestamp", func(t *testing.T) {
		// timestamp() returns unix timestamp in seconds
		normalizer := NewExpressionNormalizer("timestamp()")
		before := time.Now().Unix()
		result := normalizer.Apply(map[string]interface{}{})
		after := time.Now().Unix()
		assert.True(t, result.(float64) >= float64(before) && result.(float64) <= float64(after))

		// timestamp("ms") returns unix timestamp in milliseconds
		normalizer = NewExpressionNormalizer("timestamp('ms')")
		beforeMs := time.Now().UnixMilli()
		result = normalizer.Apply(map[string]interface{}{})
		afterMs := time.Now().UnixMilli()
		assert.True(t, result.(float64) >= float64(beforeMs) && result.(float64) <= float64(afterMs))
	})
}

func TestExprFunctionNormalizer(t *testing.T) {
	t.Run("max function", func(t *testing.T) {
		normalizer := NewExprNormalizer("a > b ? a : b")
		result := normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, int(10))
		// use max func
		normalizer = NewExprNormalizer("max(a, b)")
		result = normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, float64(10))

	})
	t.Run("min function", func(t *testing.T) {
		normalizer := NewExprNormalizer("a > b ? b : a")
		result := normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, int(8))

		// use min func
		normalizer = NewExprNormalizer("min(a, b)")
		result = normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, float64(8))
	})
	t.Run("log function", func(t *testing.T) {
		normalizer := NewExprNormalizer("log(a)")
		result := normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, math.Log(10))
	})
	t.Run("log10 function", func(t *testing.T) {
		normalizer := NewExprNormalizer("log10(a)")
		result := normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, math.Log10(10))
	})
	t.Run("log2 function", func(t *testing.T) {
		normalizer := NewExprNormalizer("log2(a)")
		result := normalizer.Apply(map[string]interface{}{"a": 10, "b": 8})
		assert.Equal(t, result, math.Log2(10))
	})
	t.Run("pow function", func(t *testing.T) {
		normalizer := NewExprNormalizer("pow(a, b)")
		result := normalizer.Apply(map[string]interface{}{"a": 2, "b": 3})
		assert.Equal(t, result, float64(8))
	})
	t.Run("string contact", func(t *testing.T) {
		normalizer := NewExprNormalizer("'hello ' + a")
		result := normalizer.Apply(map[string]interface{}{"a": "world", "b": 3})
		assert.Equal(t, result, "hello world")

		key := result.(string)
		normalizer = NewExprNormalizer("int(hash32(key)) % 100")
		result = normalizer.Apply(map[string]interface{}{"key": key})
		assert.Equal(t, result, int(murmur3.Sum32([]byte(key))%100))
	})
	t.Run("s2CellID", func(t *testing.T) {
		normalizer := NewExprNormalizer("s2CellID(lat, lng)")
		result := normalizer.Apply(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})

		lat := 39.9042  // 纬度
		lng := 116.4074 // 经度

		ll := s2.LatLngFromDegrees(lat, lng)

		cellID := s2.CellIDFromLatLng(ll)

		level := 15
		cellIDAtLevel := cellID.Parent(level)

		assert.Equal(t, utils.ToInt(result, 1), utils.ToInt(uint64(cellIDAtLevel), 0))

		normalizer = NewExprNormalizer("s2CellID(lat, lng, 20)")
		result = normalizer.Apply(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})

		level = 20
		cellIDAtLevel = cellID.Parent(level)

		assert.Equal(t, utils.ToInt(result, 1), utils.ToInt(uint64(cellIDAtLevel), 0))
	})
	t.Run("s2CellNeighbors", func(t *testing.T) {
		normalizer := NewExprNormalizer("s2CellNeighbors(lat, lng)")
		result := normalizer.Apply(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})
		t.Log(result)

		assert.Equal(t, 9, len(result.([]int)))
		lat := 39.9042  // 纬度
		lng := 116.4074 // 经度

		ll := s2.LatLngFromDegrees(lat, lng)

		cellID := s2.CellIDFromLatLng(ll)

		level := 15
		cellIDAtLevel := cellID.Parent(level)

		t.Log(int(cellIDAtLevel))
		cellIds := result.([]int)
		assert.Equal(t, cellIds[len(cellIds)-1], int(cellIDAtLevel))
	})
	t.Run("geoHash", func(t *testing.T) {
		normalizer := NewExprNormalizer("geoHash(lat, lng)")
		result := normalizer.Apply(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})

		lat := 39.9042  // 纬度
		lng := 116.4074 // 经度

		hashResult := geohash.EncodeWithPrecision(lat, lng, 6)
		assert.Equal(t, result, hashResult)

		normalizer = NewExprNormalizer("geoHash(lat, lng, 12)")
		result = normalizer.Apply(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})
		hashResult = geohash.EncodeWithPrecision(lat, lng, 12)
		assert.Equal(t, result, hashResult)
	})
	t.Run("geoHashWithNeighbors", func(t *testing.T) {
		normalizer := NewExprNormalizer("geoHashWithNeighbors(lat, lng)")
		result := normalizer.Apply(map[string]interface{}{"lat": 39.9042, "lng": 116.4074})

		lat := 39.9042  // 纬度
		lng := 116.4074 // 经度

		hashResult := geohash.EncodeWithPrecision(lat, lng, 6)
		assert.Equal(t, result.([]string)[8], hashResult)

		neighbors := geohash.Neighbors(hashResult)
		assert.Equal(t, result.([]string)[:8], neighbors)
	})
	t.Run("haversine", func(t *testing.T) {
		normalizer := NewExprNormalizer("haversine(lng1, lat1, lng2, lat2)")
		result := normalizer.Apply(map[string]interface{}{"lat1": 39.9042, "lng1": 116.4074, "lat2": 31.2304, "lng2": 121.4737})

		assert.Equal(t, utils.ToInt(result, 0), 1067)
	})
	t.Run("sphereDistance", func(t *testing.T) {
		normalizer := NewExprNormalizer("sphereDistance(lng1, lat1, lng2, lat2)")
		result := normalizer.Apply(map[string]interface{}{"lat1": 39.9042, "lng1": 116.4074, "lat2": 31.2304, "lng2": 121.4737})

		assert.Equal(t, utils.ToInt(result, 0), 1067)
	})
	t.Run("maxIndex && maxValue", func(t *testing.T) {
		normalizer := NewExprNormalizer("maxIndex(arr)")
		result := normalizer.Apply(map[string]interface{}{"arr": []float64{0.1, 0.2, 0.3, 0.4}})

		assert.Equal(t, utils.ToInt(result, 0), 3)

		normalizer = NewExprNormalizer("maxValue(arr)")
		result = normalizer.Apply(map[string]interface{}{"arr": []float64{0.1, 0.2, 0.3, 0.4}})

		assert.Equal(t, result.(float64), float64(0.4))
	})
	t.Run("timestamp", func(t *testing.T) {
		// timestamp() returns unix timestamp in seconds
		normalizer := NewExprNormalizer("timestamp()")
		before := time.Now().Unix()
		result := normalizer.Apply(map[string]interface{}{})
		after := time.Now().Unix()
		assert.True(t, result.(float64) >= float64(before) && result.(float64) <= float64(after))

		// timestamp("ms") returns unix timestamp in milliseconds
		normalizer = NewExprNormalizer("timestamp('ms')")
		beforeMs := time.Now().UnixMilli()
		result = normalizer.Apply(map[string]interface{}{})
		afterMs := time.Now().UnixMilli()
		assert.True(t, result.(float64) >= float64(beforeMs) && result.(float64) <= float64(afterMs))
	})
}

// TestDescribeParamValue verifies that problematic values (nil/NaN/Inf) are
// rendered with clear markers, while normal values print as-is.
func TestDescribeParamValue(t *testing.T) {
	assert.Equal(t, describeParamValue(nil), "<nil>")
	assert.Equal(t, describeParamValue(math.NaN()), "<NaN>")
	assert.Equal(t, describeParamValue(math.Inf(1)), "<+Inf>")
	assert.Equal(t, describeParamValue(math.Inf(-1)), "<-Inf>")
	assert.Equal(t, describeParamValue(float32(math.Inf(1))), "<+Inf>")
	assert.Equal(t, describeParamValue(3.14), "3.14")
	assert.Equal(t, describeParamValue("abc"), "abc")
	assert.Equal(t, describeParamValue(10), "10")
}

// TestDescribeExprParams verifies that only variables referenced by the
// expression are described, and that nil/NaN/Inf/missing are flagged clearly.
func TestDescribeExprParams(t *testing.T) {
	expr, err := govaluate.NewEvaluableExpression("-price + score * discount")
	if err != nil {
		t.Fatalf("build expression error: %v", err)
	}

	params := map[string]interface{}{
		"price":     nil,
		"score":     math.NaN(),
		"discount":  math.Inf(1),
		"unrelated": 12345, // not referenced by expression -> must not appear
	}

	desc := describeExprParams(expr.Vars(), params)
	t.Logf("params desc: %s", desc)

	assert.True(t, strings.Contains(desc, "price=<nil>"))
	assert.True(t, strings.Contains(desc, "score=<NaN>"))
	assert.True(t, strings.Contains(desc, "discount=<+Inf>"))
	assert.True(t, !strings.Contains(desc, "unrelated"))

	// missing key should be flagged as <missing>
	desc = describeExprParams(expr.Vars(), map[string]interface{}{"price": 1.0, "score": 2.0})
	t.Logf("params desc(missing): %s", desc)
	assert.True(t, strings.Contains(desc, "discount=<missing>"))
}

// TestExpressionNormalizerErrorLog exercises the real Apply path with a nil
// param so that the evaluation fails and the enriched error log is emitted.
// It also confirms Apply degrades gracefully (returns "") without panic.
func TestExpressionNormalizerErrorLog(t *testing.T) {
	normalizer := NewExpressionNormalizer("-price + score")
	// price is nil -> triggers: Value '<nil>' cannot be used with the modifier '-'
	result := normalizer.Apply(map[string]interface{}{"price": nil, "score": 1.0})
	assert.Equal(t, result, "")
}

// TestExprReferencedVars verifies that only the variable identifiers actually
// referenced by the expression are extracted, excluding function callees.
func TestExprReferencedVars(t *testing.T) {
	normalizer := NewExprNormalizer("max(price, score) + discount")
	if normalizer.prog == nil {
		t.Fatal("expression should compile")
	}

	vars := map[string]bool{}
	for _, v := range normalizer.vars {
		vars[v] = true
	}
	t.Logf("referenced vars: %v", normalizer.vars)

	// variables must be collected
	assert.True(t, vars["price"])
	assert.True(t, vars["score"])
	assert.True(t, vars["discount"])
	// function callee must NOT be treated as a variable
	assert.True(t, !vars["max"])
}

// TestExprNormalizerErrorLogOnlyReferencedParams confirms that the runtime error
// log describes only the params referenced by the expression, not unrelated
// (potentially sensitive) params such as user_id.
func TestExprNormalizerErrorLogOnlyReferencedParams(t *testing.T) {
	normalizer := NewExprNormalizer("a + b")
	params := map[string]interface{}{
		"a":       "not_a_number", // triggers runtime error
		"b":       1,
		"user_id": "sensitive-12345", // not referenced -> must not be described
	}

	desc := describeExprParams(normalizer.vars, params)
	t.Logf("params desc: %s", desc)
	assert.True(t, strings.Contains(desc, "a=not_a_number"))
	assert.True(t, strings.Contains(desc, "b=1"))
	assert.True(t, !strings.Contains(desc, "user_id"))
	assert.True(t, !strings.Contains(desc, "sensitive-12345"))

	// Apply still degrades gracefully to "" without panic.
	result := normalizer.Apply(params)
	assert.Equal(t, result, "")
}

// TestExprNormalizerErrorLog exercises the real Apply path with a param that
// makes the expression fail at runtime, so the enriched error log (expression
// + params) is emitted. It also confirms Apply degrades gracefully (returns "")
// without panic.
func TestExprNormalizerErrorLog(t *testing.T) {
	// applying string to arithmetic triggers a runtime error inside expr.Run
	normalizer := NewExprNormalizer("a + b")
	result := normalizer.Apply(map[string]interface{}{"a": "not_a_number", "b": 1})
	assert.Equal(t, result, "")
}

// TestExprNormalizerCompileError verifies that an invalid expression fails to
// compile, leaving prog nil, and Apply returns "" instead of panicking.
func TestExprNormalizerCompileError(t *testing.T) {
	normalizer := NewExprNormalizer("a +") // syntactically invalid
	assert.Equal(t, normalizer.prog == nil, true)
	result := normalizer.Apply(map[string]interface{}{"a": 1})
	assert.Equal(t, result, "")
}
