package rank

import (
	"math"
	"testing"

	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/module"
)

func TestResolveDebugLevel(t *testing.T) {
	cases := []struct {
		name         string
		requestLevel int
		want         int
	}{
		{"unset keeps default", 0, DefaultDebugLevel},
		{"negative keeps default", -1, DefaultDebugLevel},
		{"features without prediction", 1, 1},
		{"features and tensors with prediction", 2, 2},
		{"output features without prediction", 3, 3},
		{"save prediction request", DefaultDebugLevel, DefaultDebugLevel},
		{"vector recall debug", 102, 102},
		{"per stage prediction time", 903, 903},
		{"missing feature fields log", 904, 904},
		{"undocumented level keeps default", 6, DefaultDebugLevel},
		{"unknown level keeps default", 4, DefaultDebugLevel},
		{"max int32 keeps default", math.MaxInt32, DefaultDebugLevel},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ResolveDebugLevel(c.requestLevel); got != c.want {
				t.Fatalf("ResolveDebugLevel(%d) = %d, want %d", c.requestLevel, got, c.want)
			}
		})
	}
}

func TestEasyrecAlgoDataGeneratorDebugLevel(t *testing.T) {
	newGenerator := func() *EasyrecAlgoDataGenerator {
		generator := NewEasyrecAlgoDataGenerator(nil)
		generator.SetItemFeatures(nil)
		generator.AddFeatures(module.NewItem("item_1"), map[string]interface{}{"price": 1.5},
			map[string]interface{}{"uid": "user_1"})
		return generator
	}
	meta := map[string]string{"request_id": "request_1"}

	t.Run("normal request keeps debug level and meta unset", func(t *testing.T) {
		request := newGenerator().GeneratorAlgoData().GetFeatures().(*easyrec.PBRequest)
		if request.DebugLevel != 0 {
			t.Fatalf("debug_level = %d, want 0", request.DebugLevel)
		}
		if len(request.MetaData) != 0 {
			t.Fatalf("meta_data = %v, want empty", request.MetaData)
		}
	})

	t.Run("debug request carries request level and request id", func(t *testing.T) {
		level := ResolveDebugLevel(2)
		request := newGenerator().GeneratorAlgoDataDebugWithLevel(level, meta).GetFeatures().(*easyrec.PBRequest)
		if request.DebugLevel != int32(2) {
			t.Fatalf("debug_level = %d, want 2", request.DebugLevel)
		}
		if request.MetaData["request_id"] != "request_1" {
			t.Fatalf("meta_data = %v, want request_id=request_1", request.MetaData)
		}
	})

	t.Run("debug request with undocumented level keeps default", func(t *testing.T) {
		level := ResolveDebugLevel(6)
		request := newGenerator().GeneratorAlgoDataDebugWithLevel(level, meta).GetFeatures().(*easyrec.PBRequest)
		if request.DebugLevel != int32(DefaultDebugLevel) {
			t.Fatalf("debug_level = %d, want %d", request.DebugLevel, DefaultDebugLevel)
		}
	})

	t.Run("debug request without request level keeps default", func(t *testing.T) {
		level := ResolveDebugLevel(0)
		request := newGenerator().GeneratorAlgoDataDebugWithLevel(level, meta).GetFeatures().(*easyrec.PBRequest)
		if request.DebugLevel != int32(DefaultDebugLevel) {
			t.Fatalf("debug_level = %d, want %d", request.DebugLevel, DefaultDebugLevel)
		}
	})
}
