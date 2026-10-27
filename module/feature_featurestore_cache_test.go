package module

import (
	gocontext "context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	reccontext "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

type sharedFeatureParams map[string]any

func (p sharedFeatureParams) GetParameter(name string) any { return p[name] }

func sharedFeatureContext(features map[string]any) *reccontext.RecommendContext {
	ctx := reccontext.NewRecommendContext()
	ctx.Param = sharedFeatureParams{"features": features}
	return ctx
}

func sharedFeatureTestCache(t *testing.T) *fs.FeatureCache {
	t.Helper()
	fs.LoadFeatureCaches(&recconf.RecommendConfig{FeatureCacheConfs: map[string]recconf.FeatureCacheConfig{
		"test_shared": {CacheSize: 1000, CacheTime: 60},
	}})
	t.Cleanup(func() { fs.LoadFeatureCaches(&recconf.RecommendConfig{}) })
	c, err := fs.GetFeatureCache("test_shared")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func sharedFeatureTestConfig() recconf.FeatureDaoConfig {
	return recconf.FeatureDaoConfig{
		DaoConfig: recconf.DaoConfig{AdapterType: recconf.DataSource_Type_FeatureStore,
			FeatureStoreName: "fs", FeatureStoreViewName: "view"},
		FeatureStore: Feature_Store_User, FeatureKey: "user:uid", FeatureCacheName: "test_shared",
	}
}

// Embedding the SDK interface supplies its private methods; only the exercised
// methods are overridden, so the tests never access a network service.
type sharedTestFeatureView struct {
	domain.FeatureView
	calls     atomic.Int32
	fetch     func([]any, []string, domain.FeatureViewOptions) ([]map[string]any, error)
	kind      string
	aggregate func([]any) (map[string]any, error)
}

func (v *sharedTestFeatureView) GetFeatureEntityName() string { return "entity" }
func (v *sharedTestFeatureView) GetType() string {
	if v.kind != "" {
		return v.kind
	}
	return "batch"
}
func (v *sharedTestFeatureView) GetName() string                   { return "view" }
func (v *sharedTestFeatureView) Offline2Online(name string) string { return name }
func (v *sharedTestFeatureView) GetOnlineFeatures(keys []any, fields []string, alias map[string]string) ([]map[string]any, error) {
	return v.GetOnlineFeaturesWithOptions(keys, fields, alias, domain.FeatureViewOptions{})
}
func (v *sharedTestFeatureView) GetOnlineFeaturesWithOptions(keys []any, fields []string, alias map[string]string, options domain.FeatureViewOptions) ([]map[string]any, error) {
	v.calls.Add(1)
	rows, err := v.fetch(keys, fields, options)
	for _, row := range rows {
		for name, aliased := range alias {
			if value, ok := row[name]; ok {
				row[aliased] = value
				delete(row, name)
			}
		}
	}
	return rows, err
}
func (v *sharedTestFeatureView) GetOnlineAggregatedFeatures(keys []any, fields []string, alias map[string]string) (map[string]any, error) {
	return v.GetOnlineAggregatedFeaturesWithContext(gocontext.Background(), keys, fields, alias)
}
func (v *sharedTestFeatureView) GetOnlineAggregatedFeaturesWithContext(_ gocontext.Context, keys []any, _ []string, _ map[string]string) (map[string]any, error) {
	v.calls.Add(1)
	return v.aggregate(keys)
}

func sharedFeatureProject(view *sharedTestFeatureView) *domain.Project {
	project := &domain.Project{FeatureEntityMap: map[string]*domain.FeatureEntity{
		"entity": {FeatureEntity: &api.FeatureEntity{FeatureEntityJoinid: "uid"}},
	}}
	project.FeatureViewMap.Store("view", view)
	return project
}

func TestSharedFeatureViewReusesAcrossDAOs(t *testing.T) {
	sharedFeatureTestCache(t)
	view := &sharedTestFeatureView{fetch: func(keys []any, fields []string, opts domain.FeatureViewOptions) ([]map[string]any, error) {
		if !reflect.DeepEqual(fields, []string{"*"}) {
			t.Errorf("unexpected projection %v", fields)
		}
		return []map[string]any{{"uid": keys[0], "ids": []int64{1, 2}}}, nil
	}}
	project := sharedFeatureProject(view)
	conf := sharedFeatureTestConfig()
	conf.CacheSize = 999
	first := newSharedFeatureFeatureStoreDao(conf, func() *domain.Project { return project })
	conf.UserSelectFields = "*"
	conf.CacheFeaturesName = "later"
	second := newSharedFeatureFeatureStoreDao(conf, func() *domain.Project { return project })
	if first.cache != nil || second.cache != nil {
		t.Fatal("shared DAO allocated a private cache")
	}
	user := NewUser("u")
	user.AddProperty("uid", "u")
	first.FeatureFetch(user, nil, sharedFeatureContext(nil))
	user.AddProperty("ids", []int64{99})
	next := NewUser("u")
	next.AddProperty("uid", "u")
	second.FeatureFetch(next, nil, sharedFeatureContext(nil))
	if view.calls.Load() != 1 {
		t.Fatalf("expected one sequential raw read, got %d", view.calls.Load())
	}
	if got := next.GetCacheFeatures("later")["ids"]; !reflect.DeepEqual(got, []int64{1, 2}) {
		t.Fatalf("polluted cached sequence: %v", got)
	}
	if next.GetProperty("ids") != nil {
		t.Fatal("cacheFeatures consumer changed")
	}
}

func TestSharedFeatureNamespaceIsolation(t *testing.T) {
	base := sharedFeatureTestConfig()
	key := sharedFeatureCacheKey(base, "view", "u", nil)
	cases := []func(*recconf.FeatureDaoConfig){
		func(c *recconf.FeatureDaoConfig) { c.FeatureStoreName = "other" },
		func(c *recconf.FeatureDaoConfig) { c.FeatureStoreViewName = "other" },
		func(c *recconf.FeatureDaoConfig) { c.UserSelectFields = "age" },
		func(c *recconf.FeatureDaoConfig) { c.FeatureStoreOptions.DlrmHSTU = true },
		func(c *recconf.FeatureDaoConfig) { c.FeatureStore = Feature_Store_Item },
	}
	for _, change := range cases {
		conf := base
		change(&conf)
		if sharedFeatureCacheKey(conf, "view", "u", nil) == key {
			t.Fatal("query identity collided")
		}
	}
	base.FeatureStoreViewName = ""
	base.FeatureStoreModelName, base.FeatureStoreEntityName = "model", "entity"
	modelKey := sharedFeatureCacheKey(base, "model", "u", nil)
	for _, change := range []func(*recconf.FeatureDaoConfig){
		func(c *recconf.FeatureDaoConfig) { c.FeatureStoreModelName = "other" },
		func(c *recconf.FeatureDaoConfig) { c.FeatureStoreEntityName = "other" },
	} {
		conf := base
		change(&conf)
		if sharedFeatureCacheKey(conf, "model", "u", nil) == modelKey {
			t.Fatal("model identity collided")
		}
	}
	if sharedFeatureCacheKey(base, "aggregated", "u", []string{"a", "b"}) == sharedFeatureCacheKey(base, "aggregated", "u", []string{"b", "a"}) {
		t.Fatal("append key order lost")
	}
	if sharedFeatureCacheKey(base, "aggregated", "u", []string{"a"}) == modelKey {
		t.Fatal("aggregated and normal model collided")
	}
	if sharedFeatureCacheKey(base, "aggregated", "u", []string{"a"}) == sharedFeatureCacheKey(base, "aggregated", "u", []string{"a", "a"}) {
		t.Fatal("append key multiplicity lost")
	}
}

func TestSharedModelContextFeaturesAreRequestLocal(t *testing.T) {
	c := sharedFeatureTestCache(t)
	conf := sharedFeatureTestConfig()
	conf.FeatureStoreViewName = ""
	conf.FeatureStoreModelName, conf.FeatureStoreEntityName = "model", "entity"
	dao := newSharedFeatureFeatureStoreDao(conf, nil)
	for _, aggregated := range []bool{false, true} {
		for _, priority := range []int{0, 1} {
			key := strings.Join([]string{"model", string(rune('0' + priority)), map[bool]string{false: "plain", true: "aggregated"}[aggregated]}, ":")
			policy := &sharedModelPolicy{modelFields: map[string]bool{"label": true, "fallback": true, "online": true}, labelFields: map[string]bool{"label": true}, labelPriority: priority, aggregated: aggregated}
			reads := 0
			load := func() (map[string]any, error) {
				reads++
				return map[string]any{"label": "sdk", "online": "sdk-online", "ids": []int64{1}}, nil
			}
			for _, value := range []string{"request-a", "request-b"} {
				ctx := sharedFeatureContext(map[string]any{"label": value, "fallback": value, "online": "request-online"})
				user := NewUser("u")
				user.AddProperties(ctx.GetParameter("features").(map[string]any))
				if err := dao.loadSharedUser(user, ctx, c, key, policy, load); err != nil {
					t.Fatal(err)
				}
				wantLabel := "sdk"
				if priority == 1 {
					wantLabel = value
				}
				if user.GetProperty("label") != wantLabel || user.GetProperty("online") != "sdk-online" {
					t.Fatalf("incorrect request merge: %v", user.Properties)
				}
				if aggregated {
					if user.GetProperty("fallback") != nil {
						t.Fatal("aggregated path added fallback")
					}
				} else if user.GetProperty("fallback") != value {
					t.Fatal("normal path lost fallback")
				}
				user.AddProperty("ids", []int64{99})
			}
			if reads != 1 {
				t.Fatalf("expected cache hit, reads=%d", reads)
			}
			stored, _ := c.Get(key)
			if stored["label"] != "sdk" || stored["fallback"] != nil || stored["ids"].([]int64)[0] != 1 {
				t.Fatalf("request data leaked to cache: %v", stored)
			}
		}
	}
}

func TestSharedItemsPartialHitsAndNegativeCache(t *testing.T) {
	c := sharedFeatureTestCache(t)
	conf := sharedFeatureTestConfig()
	conf.FeatureStore = Feature_Store_Item
	conf.FeatureKey = "item:id"
	view := &sharedTestFeatureView{fetch: func(keys []any, _ []string, _ domain.FeatureViewOptions) ([]map[string]any, error) {
		seen := map[string]bool{}
		for _, key := range keys {
			seen[key.(string)] = true
		}
		if len(keys) != 2 || !seen["b"] || !seen["missing"] {
			t.Errorf("wrong miss keys: %v", keys)
		}
		return []map[string]any{{"uid": "b", "ids": []int64{2}}}, nil
	}}
	project := sharedFeatureProject(view)
	dao := newSharedFeatureFeatureStoreDao(conf, func() *domain.Project { return project })
	c.Put(sharedFeatureCacheKey(conf, "view", "a", nil), map[string]any{"uid": "a", "ids": []int64{1}})
	items := []*Item{NewItem("a"), NewItem("b"), NewItem("b"), NewItem("missing")}
	dao.FeatureFetch(nil, items, sharedFeatureContext(nil))
	if items[0].GetProperty("ids").([]int64)[0] != 1 || items[1].GetProperty("ids").([]int64)[0] != 2 {
		t.Fatal("item features were not delivered")
	}
	items[1].AddProperty("ids", []int64{99})
	if items[2].GetProperty("ids").([]int64)[0] != 2 {
		t.Fatal("replacing an item's property changed another item's property")
	}
	dao.FeatureFetch(nil, []*Item{NewItem("missing"), NewItem("b")}, sharedFeatureContext(nil))
	if view.calls.Load() != 1 {
		t.Fatal("positive or negative result was not shared")
	}
}

func TestSharedFeatureErrorsDoNotCache(t *testing.T) {
	c := sharedFeatureTestCache(t)
	conf := sharedFeatureTestConfig()
	dao := newSharedFeatureFeatureStoreDao(conf, nil)
	for i := 0; i < 2; i++ {
		err := dao.loadSharedUser(NewUser("u"), sharedFeatureContext(nil), c, "bad", nil, func() (map[string]any, error) { return nil, errors.New("offline") })
		if err == nil {
			t.Fatal("expected raw read error")
		}
		if _, ok := c.Get("bad"); ok {
			t.Fatal("failed read was cached")
		}
	}
	conf.FeatureStore = Feature_Store_Item
	conf.FeatureKey = "item:id"
	view := &sharedTestFeatureView{fetch: func([]any, []string, domain.FeatureViewOptions) ([]map[string]any, error) {
		return nil, errors.New("offline")
	}}
	project := sharedFeatureProject(view)
	dao = newSharedFeatureFeatureStoreDao(conf, func() *domain.Project { return project })
	for i := 0; i < 2; i++ {
		dao.FeatureFetch(nil, []*Item{NewItem("u")}, sharedFeatureContext(nil))
	}
	if view.calls.Load() != 2 {
		t.Fatal("failed item read was negative cached")
	}
}

func TestSharedFeatureUnavailableCacheStillReads(t *testing.T) {
	sharedFeatureTestCache(t)
	conf := sharedFeatureTestConfig()
	view := &sharedTestFeatureView{fetch: func([]any, []string, domain.FeatureViewOptions) ([]map[string]any, error) {
		return []map[string]any{{"age": 18}}, nil
	}}
	project := sharedFeatureProject(view)
	dao := newSharedFeatureFeatureStoreDao(conf, func() *domain.Project { return project })
	fs.LoadFeatureCaches(&recconf.RecommendConfig{})
	user := NewUser("u")
	user.AddProperty("uid", "u")
	dao.FeatureFetch(user, nil, sharedFeatureContext(nil))
	if user.GetProperty("age") != 18 || view.calls.Load() != 1 {
		t.Fatal("cache deletion skipped feature reading")
	}
}

func TestSharedFeatureConcurrentConsumers(t *testing.T) {
	sharedFeatureTestCache(t)
	view := &sharedTestFeatureView{fetch: func(keys []any, _ []string, _ domain.FeatureViewOptions) ([]map[string]any, error) {
		return []map[string]any{{"uid": keys[0], "ids": []int64{1}}}, nil
	}}
	project := sharedFeatureProject(view)
	dao := newSharedFeatureFeatureStoreDao(sharedFeatureTestConfig(), func() *domain.Project { return project })
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			user := NewUser("u")
			user.AddProperty("uid", "u")
			dao.FeatureFetch(user, nil, sharedFeatureContext(nil))
			if user.GetProperty("ids").([]int64)[0] != 1 {
				t.Error("polluted data")
			}
			user.AddProperty("ids", []int64{99})
		}()
	}
	wg.Wait()
}

func TestFeatureDaoSharedModeValidationAndLegacyFactory(t *testing.T) {
	for _, change := range []func(*recconf.FeatureDaoConfig){
		func(c *recconf.FeatureDaoConfig) { c.AdapterType = recconf.DaoConf_Adapter_Redis },
		func(c *recconf.FeatureDaoConfig) { c.FeatureType = Feature_Type_Sequence },
		func(c *recconf.FeatureDaoConfig) { c.FeatureCacheName = "missing" },
	} {
		conf := sharedFeatureTestConfig()
		change(&conf)
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid shared config did not panic")
				}
			}()
			NewFeatureDao(conf)
		}()
	}
	legacy := NewFeatureDao(recconf.FeatureDaoConfig{FeatureStore: Feature_Store_User, CacheSize: 10})
	old, ok := legacy.(*EmptyFeatureDao)
	if !ok || old.cache == nil {
		t.Fatal("legacy factory or private cache changed")
	}
	defer old.cache.Close()
}

func sharedFeatureModelProject(view *sharedTestFeatureView) *domain.Project {
	project := sharedFeatureProject(view)
	label := domain.NewLabelTable(&api.LabelTable{Fields: []*api.LabelTableField{{Name: "label_alias"}}})
	model := domain.NewModel(&api.Model{Name: "model", LabelPriorityLevel: 1, Features: []*api.ModelFeatures{
		{FeatureViewName: "view", Name: "label", AliasName: "label_alias"},
		{FeatureViewName: "view", Name: "online"},
		{FeatureViewName: "view", Name: "fallback"},
	}}, project, label)
	project.ModelMap.Store("model", model)
	return project
}

func TestSharedFeatureModelSDKAndContextMerge(t *testing.T) {
	sharedFeatureTestCache(t)
	view := &sharedTestFeatureView{fetch: func(keys []any, _ []string, options domain.FeatureViewOptions) ([]map[string]any, error) {
		if !options.DlrmHSTU {
			t.Error("model options were not forwarded")
		}
		return []map[string]any{{"uid": keys[0], "label": "sdk", "online": "sdk-online"}}, nil
	}}
	project := sharedFeatureModelProject(view)
	conf := sharedFeatureTestConfig()
	conf.FeatureStoreViewName = ""
	conf.FeatureStoreModelName, conf.FeatureStoreEntityName = "model", "entity"
	conf.FeatureStoreOptions.DlrmHSTU = true
	for _, value := range []string{"a", "b"} {
		dao := newSharedFeatureFeatureStoreDao(conf, func() *domain.Project { return project })
		ctx := sharedFeatureContext(map[string]any{"label_alias": value, "fallback": value})
		user := NewUser("u")
		user.AddProperty("uid", "u")
		user.AddProperties(ctx.GetParameter("features").(map[string]any))
		dao.FeatureFetch(user, nil, ctx)
		if user.GetProperty("label_alias") != value || user.GetProperty("fallback") != value || user.GetProperty("online") != "sdk-online" {
			t.Fatalf("model fetch or merge failed: %v", user.Properties)
		}
	}
	if view.calls.Load() != 1 {
		t.Fatal("model SDK result was not shared")
	}
}

func TestSharedFeatureAggregatedSDKAppendKeys(t *testing.T) {
	sharedFeatureTestCache(t)
	view := &sharedTestFeatureView{kind: "Sequence", aggregate: func(keys []any) (map[string]any, error) {
		if len(keys) < 2 || keys[0] != "u" {
			t.Errorf("invalid aggregate keys %v", keys)
		}
		return map[string]any{"label_alias": "sdk", "online": "sdk-online", "last": keys[len(keys)-1]}, nil
	}}
	project := sharedFeatureModelProject(view)
	conf := sharedFeatureTestConfig()
	conf.FeatureStoreViewName = ""
	conf.FeatureStoreModelName, conf.FeatureStoreEntityName = "model", "entity"
	conf.FeatureAppendKey = "user:extra"
	dao := newSharedFeatureFeatureStoreDao(conf, func() *domain.Project { return project })
	for _, extra := range []string{"a", "b", "a"} {
		user := NewUser("u")
		user.AddProperties(map[string]any{"uid": "u", "extra": []string{extra}, "fallback": "request"})
		dao.FeatureFetch(user, nil, sharedFeatureContext(map[string]any{"label_alias": extra, "fallback": "request"}))
		if user.GetProperty("last") != extra || user.GetProperty("label_alias") != extra || user.GetProperty("fallback") != nil {
			t.Fatalf("aggregate path failed: %v", user.Properties)
		}
	}
	if view.calls.Load() != 2 {
		t.Fatalf("appendKeys isolation failed, reads=%d", view.calls.Load())
	}
}

func TestSharedItemsModelSDKAndPropertyKeys(t *testing.T) {
	sharedFeatureTestCache(t)
	view := &sharedTestFeatureView{fetch: func(keys []any, _ []string, _ domain.FeatureViewOptions) ([]map[string]any, error) {
		return []map[string]any{{"uid": keys[0], "online": "sdk-online"}}, nil
	}}
	project := sharedFeatureModelProject(view)
	conf := sharedFeatureTestConfig()
	conf.FeatureStoreViewName = ""
	conf.FeatureStore = Feature_Store_Item
	conf.FeatureStoreModelName, conf.FeatureStoreEntityName = "model", "entity"
	conf.FeatureKey = "item:lookup"
	dao := newSharedFeatureFeatureStoreDao(conf, func() *domain.Project { return project })
	for i := 0; i < 2; i++ {
		item := NewItem("item")
		item.AddProperty("lookup", "key")
		item.AddProperty("original", "not-cached")
		dao.FeatureFetch(nil, []*Item{item}, sharedFeatureContext(nil))
		if item.GetProperty("online") != "sdk-online" {
			t.Fatal("model item result not delivered")
		}
	}
	if view.calls.Load() != 1 {
		t.Fatal("model item result not shared")
	}
}

func TestSharedFeatureHotReloadDiscardsLateBackfill(t *testing.T) {
	sharedFeatureTestCache(t)
	started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	view := &sharedTestFeatureView{fetch: func([]any, []string, domain.FeatureViewOptions) ([]map[string]any, error) {
		close(started)
		<-release
		return []map[string]any{{"value": "old-read"}}, nil
	}}
	project := sharedFeatureProject(view)
	conf := sharedFeatureTestConfig()
	dao := newSharedFeatureFeatureStoreDao(conf, func() *domain.Project { return project })
	user := NewUser("u")
	user.AddProperty("uid", "u")
	go func() { defer close(done); dao.FeatureFetch(user, nil, sharedFeatureContext(nil)) }()
	<-started
	fs.LoadFeatureCaches(&recconf.RecommendConfig{FeatureCacheConfs: map[string]recconf.FeatureCacheConfig{"test_shared": {CacheSize: 2000, CacheTime: 60}}})
	close(release)
	<-done
	if user.GetProperty("value") != "old-read" {
		t.Fatal("reload discarded the active request result")
	}
	c, _ := fs.GetFeatureCache("test_shared")
	if _, ok := c.Get(sharedFeatureCacheKey(conf, "view", "u", nil)); ok {
		t.Fatal("late backfill polluted the replacement")
	}
}

func TestSharedFeatureEmptyReadIsNotCached(t *testing.T) {
	c := sharedFeatureTestCache(t)
	view := &sharedTestFeatureView{fetch: func([]any, []string, domain.FeatureViewOptions) ([]map[string]any, error) { return nil, nil }}
	project := sharedFeatureProject(view)
	conf := sharedFeatureTestConfig()
	dao := newSharedFeatureFeatureStoreDao(conf, func() *domain.Project { return project })
	for i := 0; i < 2; i++ {
		user := NewUser("u")
		user.AddProperty("uid", "u")
		dao.FeatureFetch(user, nil, sharedFeatureContext(nil))
	}
	if view.calls.Load() != 2 {
		t.Fatal("empty user result unexpectedly cached")
	}
	if _, ok := c.Get(sharedFeatureCacheKey(conf, "view", "u", nil)); ok {
		t.Fatal("empty user result cached")
	}
}
