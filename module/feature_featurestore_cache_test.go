package module

import (
	"errors"
	"fmt"
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

type sharedFeatureTestInstance struct {
	mu         sync.Mutex
	entries    map[string]map[string]any
	definition recconf.FeatureCacheConfig
	project    func() *domain.Project
	closed     bool
}

func (c *sharedFeatureTestInstance) Config() recconf.FeatureCacheConfig { return c.definition }
func (c *sharedFeatureTestInstance) Project() *domain.Project           { return c.project() }
func (c *sharedFeatureTestInstance) Get(key string) (map[string]any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, false
	}
	v, ok := c.entries[key]
	return fs.CloneFeatureProperties(v), ok
}
func (c *sharedFeatureTestInstance) Put(key string, value map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.entries[key] = fs.CloneFeatureProperties(value)
	}
}
func (c *sharedFeatureTestInstance) PutIfAbsent(key string, value map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.entries[key]; !c.closed && !exists {
		c.entries[key] = fs.CloneFeatureProperties(value)
	}
}
func (c *sharedFeatureTestInstance) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	c.entries = nil
}

func sharedFeatureTestCache(t *testing.T) *sharedFeatureTestInstance {
	t.Helper()
	c := &sharedFeatureTestInstance{entries: make(map[string]map[string]any), definition: recconf.FeatureCacheConfig{
		FeatureStoreName: "fs", FeatureStore: "user", FeatureStoreViewName: "view", CacheSize: 1000, CacheTime: 60,
	}}
	t.Cleanup(c.Close)
	return c
}

func newSharedFeatureTestDao(conf recconf.FeatureDaoConfig, c *sharedFeatureTestInstance, project func() *domain.Project) *SharedFeatureFeatureStoreDao {
	c.project = project
	return newSharedFeatureFeatureStoreDao(conf, func(name string) (sharedFeatureCache, error) {
		if name != "test_shared" {
			return nil, fmt.Errorf("missing cache")
		}
		return c, nil
	})
}

func sharedFeatureTestConfig() recconf.FeatureDaoConfig {
	return recconf.FeatureDaoConfig{
		DaoConfig:    recconf.DaoConfig{AdapterType: recconf.DataSource_Type_FeatureStore},
		FeatureStore: Feature_Store_User, FeatureKey: "user:uid", FeatureCacheName: "test_shared",
	}
}

// Embedding the SDK interface supplies its private methods; only the exercised
// methods are overridden, so the tests never access a network service.
type sharedTestFeatureView struct {
	domain.FeatureView
	calls atomic.Int32
	fetch func([]any, []string, domain.FeatureViewOptions) ([]map[string]any, error)
}

func (v *sharedTestFeatureView) GetFeatureEntityName() string      { return "entity" }
func (v *sharedTestFeatureView) GetType() string                   { return "batch" }
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

func sharedFeatureProject(view *sharedTestFeatureView) *domain.Project {
	project := &domain.Project{FeatureEntityMap: map[string]*domain.FeatureEntity{
		"entity": {FeatureEntity: &api.FeatureEntity{FeatureEntityJoinid: "uid"}},
	}}
	project.FeatureViewMap.Store("view", view)
	return project
}

func TestSharedFeatureViewReusesAcrossDAOs(t *testing.T) {
	c := sharedFeatureTestCache(t)
	view := &sharedTestFeatureView{fetch: func(keys []any, fields []string, opts domain.FeatureViewOptions) ([]map[string]any, error) {
		if !reflect.DeepEqual(fields, []string{"*"}) {
			t.Errorf("unexpected projection %v", fields)
		}
		return []map[string]any{{"uid": keys[0], "ids": []int64{1, 2}}}, nil
	}}
	project := sharedFeatureProject(view)
	conf := sharedFeatureTestConfig()
	conf.CacheSize = 999
	first := newSharedFeatureTestDao(conf, c, func() *domain.Project { return project })
	conf.CacheFeaturesName = "later"
	second := newSharedFeatureTestDao(conf, c, func() *domain.Project { return project })
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
	if raw, ok := c.Get("u"); !ok || raw["uid"] != "u" {
		t.Fatal("shared cache must use the actual primary key")
	}
}

func TestSharedFeatureRejectsDuplicateSourceSettings(t *testing.T) {
	c := sharedFeatureTestCache(t)
	for _, change := range []func(*recconf.FeatureDaoConfig){
		func(c *recconf.FeatureDaoConfig) { c.FeatureStoreName = "fs" },
		func(c *recconf.FeatureDaoConfig) { c.FeatureStoreViewName = "view" },
		func(c *recconf.FeatureDaoConfig) { c.FeatureStoreModelName = "model" },
		func(c *recconf.FeatureDaoConfig) { c.FeatureStoreEntityName = "entity" },
		func(c *recconf.FeatureDaoConfig) { c.UserSelectFields = "*" },
		func(c *recconf.FeatureDaoConfig) { c.ItemSelectFields = "age" },
		func(c *recconf.FeatureDaoConfig) { c.FeatureStoreOptions.DlrmHSTU = true },
	} {
		conf := sharedFeatureTestConfig()
		change(&conf)
		func() {
			defer func() {
				if message, ok := recover().(string); !ok || !strings.Contains(message, "FeatureCacheConfs") {
					t.Error("duplicate settings accepted")
				}
			}()
			newSharedFeatureTestDao(conf, c, nil)
		}()
	}
	conf := sharedFeatureTestConfig()
	conf.FeatureStore = Feature_Store_Item
	func() {
		defer func() {
			if recover() == nil {
				t.Error("mismatched scope accepted")
			}
		}()
		newSharedFeatureTestDao(conf, c, nil)
	}()
}

func TestSharedModelContextFeaturesAreRequestLocal(t *testing.T) {
	c := sharedFeatureTestCache(t)
	conf := sharedFeatureTestConfig()
	c.definition.FeatureStoreViewName = ""
	c.definition.FeatureStoreModelName, c.definition.FeatureStoreEntityName = "model", "entity"
	dao := newSharedFeatureTestDao(conf, c, nil)
	for _, priority := range []int{0, 1} {
		key := string(rune('0' + priority))
		policy := &sharedModelPolicy{modelFields: map[string]bool{"label": true, "fallback": true, "online": true}, labelFields: map[string]bool{"label": true}, labelPriority: priority}
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
			if user.GetProperty("fallback") != value {
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

func TestSharedItemsPartialHitsAndNegativeCache(t *testing.T) {
	c := sharedFeatureTestCache(t)
	conf := sharedFeatureTestConfig()
	conf.FeatureStore = Feature_Store_Item
	c.definition.FeatureStore = "item"
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
	dao := newSharedFeatureTestDao(conf, c, func() *domain.Project { return project })
	c.Put("a", map[string]any{"uid": "a", "ids": []int64{1}})
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
	dao := newSharedFeatureTestDao(conf, c, nil)
	for i := 0; i < 2; i++ {
		err := dao.loadSharedUser(NewUser("u"), sharedFeatureContext(nil), c, "bad", nil, func() (map[string]any, error) { return nil, errors.New("offline") })
		if err == nil {
			t.Fatal("expected raw read error")
		}
		if _, ok := c.Get("bad"); ok {
			t.Fatal("failed read was cached")
		}
	}
	c = sharedFeatureTestCache(t)
	conf.FeatureStore = Feature_Store_Item
	c.definition.FeatureStore = "item"
	conf.FeatureKey = "item:id"
	view := &sharedTestFeatureView{fetch: func([]any, []string, domain.FeatureViewOptions) ([]map[string]any, error) {
		return nil, errors.New("offline")
	}}
	project := sharedFeatureProject(view)
	dao = newSharedFeatureTestDao(conf, c, func() *domain.Project { return project })
	for i := 0; i < 2; i++ {
		dao.FeatureFetch(nil, []*Item{NewItem("u")}, sharedFeatureContext(nil))
	}
	if view.calls.Load() != 2 {
		t.Fatal("failed item read was negative cached")
	}
}

func TestSharedFeatureUnavailableCacheStillReads(t *testing.T) {
	c := sharedFeatureTestCache(t)
	conf := sharedFeatureTestConfig()
	view := &sharedTestFeatureView{fetch: func([]any, []string, domain.FeatureViewOptions) ([]map[string]any, error) {
		return []map[string]any{{"age": 18}}, nil
	}}
	project := sharedFeatureProject(view)
	dao := newSharedFeatureTestDao(conf, c, func() *domain.Project { return project })
	c.Close()
	dao.resolve = func(string) (sharedFeatureCache, error) { return nil, fmt.Errorf("removed") }
	user := NewUser("u")
	user.AddProperty("uid", "u")
	dao.FeatureFetch(user, nil, sharedFeatureContext(nil))
	if user.GetProperty("age") != 18 || view.calls.Load() != 1 {
		t.Fatal("cache deletion skipped feature reading")
	}
}

func TestSharedFeatureConcurrentConsumers(t *testing.T) {
	c := sharedFeatureTestCache(t)
	view := &sharedTestFeatureView{fetch: func(keys []any, _ []string, _ domain.FeatureViewOptions) ([]map[string]any, error) {
		return []map[string]any{{"uid": keys[0], "ids": []int64{1}}}, nil
	}}
	project := sharedFeatureProject(view)
	dao := newSharedFeatureTestDao(sharedFeatureTestConfig(), c, func() *domain.Project { return project })
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
	c := sharedFeatureTestCache(t)
	view := &sharedTestFeatureView{fetch: func(keys []any, _ []string, options domain.FeatureViewOptions) ([]map[string]any, error) {
		if !options.DlrmHSTU {
			t.Error("model options were not forwarded")
		}
		return []map[string]any{{"uid": keys[0], "label": "sdk", "online": "sdk-online"}}, nil
	}}
	project := sharedFeatureModelProject(view)
	conf := sharedFeatureTestConfig()
	c.definition.FeatureStoreViewName = ""
	c.definition.FeatureStoreModelName, c.definition.FeatureStoreEntityName = "model", "entity"
	c.definition.FeatureStoreOptions.DlrmHSTU = true
	for _, value := range []string{"a", "b"} {
		dao := newSharedFeatureTestDao(conf, c, func() *domain.Project { return project })
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

func TestSharedFeatureRejectsAppendKeyConfiguration(t *testing.T) {
	for _, scope := range []string{Feature_Store_User, Feature_Store_Item} {
		for _, path := range []string{"view", "model"} {
			for _, appendKey := range []string{"user:extra", "invalid", " "} {
				conf := sharedFeatureTestConfig()
				conf.FeatureStore = scope
				conf.FeatureAppendKey = appendKey
				if path == "model" {
					conf.FeatureStoreViewName = ""
					conf.FeatureStoreModelName, conf.FeatureStoreEntityName = "model", "entity"
				}
				for _, construct := range []func(){
					func() { validateSharedFeatureConfig(conf) },
					func() { NewFeatureDao(conf) },
				} {
					func() {
						defer func() {
							message, ok := recover().(string)
							if !ok || !strings.Contains(message, "FeatureAppendKey cannot be used with FeatureCacheName") {
								t.Errorf("expected append key conflict before reading cache or FS data, got %q", message)
							}
						}()
						construct()
					}()
				}
			}
		}
	}
}

func TestFeatureAppendKeyWithoutSharedCacheUsesLegacyPath(t *testing.T) {
	dao := NewFeatureDao(recconf.FeatureDaoConfig{FeatureStore: Feature_Store_User, FeatureAppendKey: "user:extra"})
	legacy, ok := dao.(*EmptyFeatureDao)
	if !ok || legacy.featureAppendKey != "user:extra" {
		t.Fatal("append key without a shared cache must keep the legacy path")
	}
}

func TestSharedItemsModelSDKAndPropertyKeys(t *testing.T) {
	c := sharedFeatureTestCache(t)
	view := &sharedTestFeatureView{fetch: func(keys []any, _ []string, _ domain.FeatureViewOptions) ([]map[string]any, error) {
		return []map[string]any{{"uid": keys[0], "online": "sdk-online"}}, nil
	}}
	project := sharedFeatureModelProject(view)
	conf := sharedFeatureTestConfig()
	c.definition.FeatureStoreViewName = ""
	conf.FeatureStore = Feature_Store_Item
	c.definition.FeatureStore = "item"
	c.definition.FeatureStoreModelName, c.definition.FeatureStoreEntityName = "model", "entity"
	conf.FeatureKey = "item:lookup"
	dao := newSharedFeatureTestDao(conf, c, func() *domain.Project { return project })
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
	old := sharedFeatureTestCache(t)
	started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	view := &sharedTestFeatureView{fetch: func([]any, []string, domain.FeatureViewOptions) ([]map[string]any, error) {
		close(started)
		<-release
		return []map[string]any{{"value": "old-read"}}, nil
	}}
	project := sharedFeatureProject(view)
	old.project = func() *domain.Project { return project }
	current := old
	dao := newSharedFeatureFeatureStoreDao(sharedFeatureTestConfig(), func(string) (sharedFeatureCache, error) { return current, nil })
	user := NewUser("u")
	user.AddProperty("uid", "u")
	go func() { defer close(done); dao.FeatureFetch(user, nil, sharedFeatureContext(nil)) }()
	<-started
	current = sharedFeatureTestCache(t)
	current.definition.FeatureStoreViewName = "new_view"
	old.Close()
	close(release)
	<-done
	if user.GetProperty("value") != "old-read" {
		t.Fatal("reload discarded active request data")
	}
	if _, ok := current.Get("u"); ok {
		t.Fatal("old read backfilled new generation")
	}
}

func TestSharedFeatureExistingDAOUsesUpdatedDefinition(t *testing.T) {
	old := sharedFeatureTestCache(t)
	oldView := &sharedTestFeatureView{fetch: func([]any, []string, domain.FeatureViewOptions) ([]map[string]any, error) {
		return []map[string]any{{"value": "old"}}, nil
	}}
	old.project = func() *domain.Project { return sharedFeatureProject(oldView) }
	current := old
	dao := newSharedFeatureFeatureStoreDao(sharedFeatureTestConfig(), func(string) (sharedFeatureCache, error) { return current, nil })
	user := NewUser("u")
	user.AddProperty("uid", "u")
	dao.FeatureFetch(user, nil, sharedFeatureContext(nil))
	if user.GetProperty("value") != "old" {
		t.Fatal("initial definition was not used")
	}
	newView := &sharedTestFeatureView{fetch: func(_ []any, fields []string, _ domain.FeatureViewOptions) ([]map[string]any, error) {
		if !reflect.DeepEqual(fields, []string{"value"}) {
			t.Errorf("projection did not come from cache definition: %v", fields)
		}
		return []map[string]any{{"value": "new"}}, nil
	}}
	project := sharedFeatureProject(newView)
	project.FeatureViewMap.Store("new_view", newView)
	current = sharedFeatureTestCache(t)
	current.definition.FeatureStoreViewName = "new_view"
	current.definition.FeatureStoreName = "other_fs"
	current.definition.SelectFields = "value"
	current.project = func() *domain.Project { return project }
	old.Close()
	for i := 0; i < 2; i++ {
		next := NewUser("u")
		next.AddProperty("uid", "u")
		dao.FeatureFetch(next, nil, sharedFeatureContext(nil))
		if next.GetProperty("value") != "new" {
			t.Fatal("existing DAO did not use current cache definition")
		}
	}
	if oldView.calls.Load() != 1 || newView.calls.Load() != 1 {
		t.Fatal("source or cache generation mismatch")
	}
}

func TestSharedFeatureEmptyReadIsNotCached(t *testing.T) {
	c := sharedFeatureTestCache(t)
	view := &sharedTestFeatureView{fetch: func([]any, []string, domain.FeatureViewOptions) ([]map[string]any, error) { return nil, nil }}
	project := sharedFeatureProject(view)
	conf := sharedFeatureTestConfig()
	dao := newSharedFeatureTestDao(conf, c, func() *domain.Project { return project })
	for i := 0; i < 2; i++ {
		user := NewUser("u")
		user.AddProperty("uid", "u")
		dao.FeatureFetch(user, nil, sharedFeatureContext(nil))
	}
	if view.calls.Load() != 2 {
		t.Fatal("empty user result unexpectedly cached")
	}
	if _, ok := c.Get("u"); ok {
		t.Fatal("empty user result cached")
	}
}
