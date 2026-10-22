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
			{Name: "item_id", Domain: "item", Operator: "equal", Type: "string", Value: "item-1"},
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
	if got := cache.data["gen_item-1"]["item_id"]; got != "gen_item-1" {
		t.Fatalf("cached item_id must remain unchanged, got %v", got)
	}
}

func TestItemStateFilterFeatureViewCacheDaoIsolatesFilterProperties(t *testing.T) {
	prog, err := expr.Compile(`"gen_" + item.item_id`, expr.AllowUndefinedVariables())
	if err != nil {
		t.Fatalf("compile expr failed: %v", err)
	}
	for _, tc := range []struct {
		name        string
		operator    string
		value       int
		cacheStatus bool
		transform   bool
		wantCount   int
	}{
		{name: "missing status must not use original item property", operator: "equal", value: 1, wantCount: 0},
		{name: "missing status remains null", operator: "is_null", wantCount: 1},
		{name: "cached status overrides original item property", operator: "equal", value: 1, cacheStatus: true, wantCount: 0},
		{name: "transformation uses complete item properties", operator: "equal", value: 2, transform: true, wantCount: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			properties := map[string]any{"item_id": "gen_item-1"}
			if tc.cacheStatus {
				properties["status"] = 0
			}
			cache := &fakeFeatureViewCache{
				joinId: "item_id",
				ready:  true,
				data:   map[string]map[string]any{"gen_item-1": properties},
			}
			dao := &ItemStateFilterFeatureViewCacheDao{
				cache:               cache,
				generateUserProgram: prog,
				filterParam: NewFilterParamWithConfig([]recconf.FilterParamConfig{
					{Name: "item_id", Domain: "item", Operator: "equal", Type: "string", Value: "item-1"},
					{Name: "status", Domain: "item", Operator: tc.operator, Type: "int", Value: tc.value},
				}),
			}
			transformed := false
			if tc.transform {
				dao.transFunc = func(_ *User, item *Item, _ *recommendcontext.RecommendContext) {
					transformed = true
					if item.StringProperty("item_id") != "item-1" || item.GetProperty("status") != 1 {
						t.Fatalf("transformation received unexpected item properties: %v", item.GetProperties())
					}
					item.AddProperty("status", 2)
				}
			}
			item := NewItemWithProperty("item-1", map[string]any{"status": 1})
			result := dao.Filter(NewUser("user-1"), []*Item{item}, recommendcontext.NewRecommendContext())
			if len(result) != tc.wantCount {
				t.Fatalf("expected %d items, got %d", tc.wantCount, len(result))
			}
			if transformed != tc.transform {
				t.Fatalf("transformation called = %v, want %v", transformed, tc.transform)
			}
			if got := item.StringProperty("item_id"); got != "item-1" {
				t.Fatalf("item_id was not restored, got %q", got)
			}
			if properties["item_id"] != "gen_item-1" {
				t.Fatalf("cached item_id was modified: %v", properties)
			}
			status, ok := properties["status"]
			if ok != tc.cacheStatus || (ok && status != 0) {
				t.Fatalf("cached status was modified: %v", properties)
			}
		})
	}
}
