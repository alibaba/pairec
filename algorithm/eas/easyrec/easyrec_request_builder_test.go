package easyrec

import (
	"math"
	"testing"
)

// TestUnsignedFeature pins the encoding of an unsigned integer, which PBFeature
// has no native type for, across the three ranges it can fall into.
func TestUnsignedFeature(t *testing.T) {
	cases := []struct {
		name string
		in   uint64
		want func(*PBFeature) bool
	}{
		{"fits int32", 42, func(f *PBFeature) bool {
			_, ok := f.Value.(*PBFeature_IntFeature)
			return ok && f.GetIntFeature() == 42
		}},
		{"fits int64 only", 5000000000, func(f *PBFeature) bool {
			_, ok := f.Value.(*PBFeature_LongFeature)
			return ok && f.GetLongFeature() == 5000000000
		}},
		{"MaxInt64+1 becomes exact string", uint64(math.MaxInt64) + 1, func(f *PBFeature) bool {
			_, ok := f.Value.(*PBFeature_StringFeature)
			return ok && f.GetStringFeature() == "9223372036854775808"
		}},
		{"MaxUint64 becomes exact string", uint64(math.MaxUint64), func(f *PBFeature) bool {
			_, ok := f.Value.(*PBFeature_StringFeature)
			return ok && f.GetStringFeature() == "18446744073709551615"
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if !c.want(unsignedFeature(c.in)) {
				t.Fatalf("unsignedFeature(%d) = %v, unexpected encoding", c.in, unsignedFeature(c.in))
			}
		})
	}
}

// TestAddItemFeatureUnsigned verifies an unsigned value from a restored snapshot
// reaches the item feature as an exact String feature instead of collapsing to
// the empty string the old default produced.
func TestAddItemFeatureUnsigned(t *testing.T) {
	builder := NewEasyrecRequestBuilder()
	builder.AddItemFeature("f", []interface{}{
		uint64(math.MaxInt64) + 1,
		uint64(math.MaxUint64),
		uint64(7),
	})
	features := builder.EasyrecRequest().ItemFeatures["f"].Features

	// one feature per input, the positional alignment with the items is kept
	if len(features) != 3 {
		t.Fatalf("got %d features, want 3", len(features))
	}
	if got := features[0].GetStringFeature(); got != "9223372036854775808" {
		t.Fatalf("features[0] = %q, want exact 9223372036854775808", got)
	}
	if got := features[1].GetStringFeature(); got != "18446744073709551615" {
		t.Fatalf("features[1] = %q, want exact 18446744073709551615", got)
	}
	if _, ok := features[2].Value.(*PBFeature_IntFeature); !ok || features[2].GetIntFeature() != 7 {
		t.Fatalf("features[2] = %v, want IntFeature 7", features[2])
	}
}

// TestAddContextFeatureUnsigned is the context counterpart of the item test.
func TestAddContextFeatureUnsigned(t *testing.T) {
	builder := NewEasyrecRequestBuilder()
	builder.AddContextFeature("f", []interface{}{uint64(math.MaxUint64)})
	features := builder.EasyrecRequest().ContextFeatures["f"].Features

	if len(features) != 1 {
		t.Fatalf("got %d features, want 1", len(features))
	}
	if got := features[0].GetStringFeature(); got != "18446744073709551615" {
		t.Fatalf("features[0] = %q, want exact 18446744073709551615", got)
	}
}

// TestListFeatureUnsupportedKeepsEmptyString guards that the reflect fallback
// only changed the unsigned path: nil and unsupported types still become an
// empty StringFeature, and are never skipped, so the list stays aligned.
func TestListFeatureUnsupportedKeepsEmptyString(t *testing.T) {
	for _, in := range []interface{}{nil, true, struct{}{}, []byte{1}} {
		f := listFeature(in)
		if _, ok := f.Value.(*PBFeature_StringFeature); !ok || f.GetStringFeature() != "" {
			t.Fatalf("listFeature(%v) = %v, want empty StringFeature", in, f)
		}
	}
}
