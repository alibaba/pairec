package module

import (
	"testing"

	recommendcontext "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/expr-lang/expr"
)

// fakeFeatureViewCache is a test double for fs.FeatureViewCache's read side.
type fakeFeatureViewCache struct {
	joinId string
	ready  bool
	data   map[string]map[string]any
}

func (f *fakeFeatureViewCache) Get(id string) (map[string]any, bool) {
	v, ok := f.data[id]
	return v, ok
}

func (f *fakeFeatureViewCache) JoinId() string { return f.joinId }

func (f *fakeFeatureViewCache) Ready() bool { return f.ready }

func TestItemStateFilterFeatureViewCacheDaoEvaluatesCacheMiss(t *testing.T) {
	dao := &ItemStateFilterFeatureViewCacheDao{
		cache: &fakeFeatureViewCache{ready: true},
		filterParam: NewFilterParamWithConfig([]recconf.FilterParamConfig{
			{
				Name:     "status",
				Domain:   "item",
				Operator: "equal",
				Type:     "int",
				Value:    1,
			},
		}),
	}

	items := []*Item{NewItem("item-1")}
	result := dao.Filter(NewUser("user-1"), items, recommendcontext.NewRecommendContext())
	if len(result) != 0 {
		t.Fatalf("expected cache miss item to be filtered, got %d items", len(result))
	}
}

func TestItemStateFilterFeatureViewCacheDaoTransformsCacheMiss(t *testing.T) {
	transformed := false
	dao := &ItemStateFilterFeatureViewCacheDao{
		cache: &fakeFeatureViewCache{ready: true},
		filterParam: NewFilterParamWithConfig([]recconf.FilterParamConfig{
			{
				Name:     "status",
				Domain:   "item",
				Operator: "equal",
				Type:     "int",
				Value:    1,
			},
		}),
		transFunc: func(_ *User, item *Item, _ *recommendcontext.RecommendContext) {
			transformed = true
			item.AddProperty("status", 1)
		},
	}

	items := []*Item{NewItem("item-1")}
	result := dao.Filter(NewUser("user-1"), items, recommendcontext.NewRecommendContext())
	if !transformed {
		t.Fatal("expected cache miss item to run feature transformation")
	}
	if len(result) != 1 {
		t.Fatalf("expected transformed cache miss item to pass filter, got %d items", len(result))
	}
}

// TestItemStateFilterFeatureViewCacheDaoFailOpenWhenNotReady verifies W2/W3:
// while the cache has not completed its first full load, the filter fails open
// (returns all items unfiltered) instead of emptying the result set.
func TestItemStateFilterFeatureViewCacheDaoFailOpenWhenNotReady(t *testing.T) {
	dao := &ItemStateFilterFeatureViewCacheDao{
		cache: &fakeFeatureViewCache{ready: false},
		filterParam: NewFilterParamWithConfig([]recconf.FilterParamConfig{
			{
				Name:     "status",
				Domain:   "item",
				Operator: "equal",
				Type:     "int",
				Value:    1,
			},
		}),
	}

	items := []*Item{NewItem("item-1"), NewItem("item-2")}
	result := dao.Filter(NewUser("user-1"), items, recommendcontext.NewRecommendContext())
	if len(result) != 2 {
		t.Fatalf("expected not-ready cache to fail open and keep all %d items, got %d", len(items), len(result))
	}
}

// TestItemStateFilterFeatureViewCacheDaoRestoresGeneratedItemId verifies C5:
// when GenerateUserDataExpr derives a query key and the entity join id is
// item_id, the cached join id value (the generated key) must not leak into the
// item's item_id property; it is restored to the real item id, matching
// ItemStateFilterFeatureStoreDao.
func TestItemStateFilterFeatureViewCacheDaoRestoresGeneratedItemId(t *testing.T) {
	// The view is keyed by the generated key "gen_<itemId>"; its join id field
	// (item_id) therefore holds the generated key, not the real item id.
	cache := &fakeFeatureViewCache{
		joinId: "item_id",
		ready:  true,
		data: map[string]map[string]any{
			"gen_item-1": {"item_id": "gen_item-1", "status": 1},
		},
	}
	prog, err := expr.Compile(`"gen_" + item.item_id`, expr.AllowUndefinedVariables())
	if err != nil {
		t.Fatalf("compile expr failed: %v", err)
	}
	dao := &ItemStateFilterFeatureViewCacheDao{
		cache:               cache,
		generateUserProgram: prog,
		filterParam: NewFilterParamWithConfig([]recconf.FilterParamConfig{
			{Name: "status", Domain: "item", Operator: "equal", Type: "int", Value: 1},
		}),
	}

	item := NewItem("item-1")
	result := dao.Filter(NewUser("user-1"), []*Item{item}, recommendcontext.NewRecommendContext())
	if len(result) != 1 {
		t.Fatalf("expected item to pass filter, got %d items", len(result))
	}
	if got := item.StringProperty("item_id"); got != "item-1" {
		t.Fatalf("expected item_id restored to real id %q, got %q", "item-1", got)
	}
}
