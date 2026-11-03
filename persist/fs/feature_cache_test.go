package fs

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

type featureCacheTestView struct{ domain.FeatureView }

func (*featureCacheTestView) GetFeatureEntityName() string      { return "entity" }
func (*featureCacheTestView) Offline2Online(name string) string { return name }

func freshFeatureCaches(t *testing.T) {
	t.Helper()
	LoadFeatureCaches(&recconf.RecommendConfig{})
	project := &domain.Project{FeatureEntityMap: map[string]*domain.FeatureEntity{
		"entity":    {FeatureEntity: &api.FeatureEntity{FeatureEntityJoinid: "uid"}},
		"unrelated": {FeatureEntity: &api.FeatureEntity{FeatureEntityJoinid: "other_uid"}},
	}}
	project.FeatureViewMap.Store("view", &featureCacheTestView{})
	project.FeatureViewMap.Store("new_view", &featureCacheTestView{})
	project.ModelMap.Store("model", domain.NewModel(&api.Model{Features: []*api.ModelFeatures{{FeatureViewName: "view", Name: "age"}}}, project, domain.NewLabelTable(&api.LabelTable{})))
	project.ModelMap.Store("empty_model", domain.NewModel(&api.Model{}, project, domain.NewLabelTable(&api.LabelTable{})))
	fsInstancesMu.Lock()
	previous := fsInstances
	fsInstances = map[string]*FSClient{"test_cache_fs": {project: project}}
	fsInstancesMu.Unlock()
	t.Cleanup(func() {
		LoadFeatureCaches(&recconf.RecommendConfig{})
		fsInstancesMu.Lock()
		fsInstances = previous
		fsInstancesMu.Unlock()
	})
}

func featureCacheDefinition(size, ttl int) recconf.FeatureCacheConfig {
	return recconf.FeatureCacheConfig{FeatureStoreName: "test_cache_fs", FeatureStore: "user", FeatureStoreViewName: "view", CacheSize: size, CacheTime: ttl}
}

func featureCacheConfig(size, ttl int) *recconf.RecommendConfig {
	return &recconf.RecommendConfig{FeatureCacheConfs: map[string]recconf.FeatureCacheConfig{
		"shared": featureCacheDefinition(size, ttl),
	}}
}

func mustFeatureCache(t *testing.T, name string) *FeatureCache {
	t.Helper()
	c, err := GetFeatureCache(name)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestFeatureCacheReloadLifecycle(t *testing.T) {
	freshFeatureCaches(t)
	LoadFeatureCaches(featureCacheConfig(10, 0))
	old := mustFeatureCache(t, "shared")
	old.Put("u", map[string]any{"age": 18})
	LoadFeatureCaches(featureCacheConfig(10, 3600))
	if current := mustFeatureCache(t, "shared"); current != old {
		t.Fatal("equivalent effective TTL must preserve the instance")
	}
	if _, ok := old.Get("u"); !ok {
		t.Fatal("unchanged configuration discarded cached data")
	}
	for _, conf := range []*recconf.RecommendConfig{featureCacheConfig(20, 3600), featureCacheConfig(20, 60)} {
		LoadFeatureCaches(conf)
		current := mustFeatureCache(t, "shared")
		if current == old {
			t.Fatal("changed policy must replace the instance")
		}
		old.Put("late", map[string]any{"v": 1})
		if _, ok := old.Get("late"); ok {
			t.Fatal("retired cache accepted a late write")
		}
		if _, ok := current.Get("late"); ok {
			t.Fatal("late write reached the replacement")
		}
		old.Close()
		old = current
	}
	LoadFeatureCaches(&recconf.RecommendConfig{})
	if _, err := GetFeatureCache("shared"); err == nil {
		t.Fatal("removed cache is still registered")
	}
	old.Put("x", map[string]any{"v": 2})
	if _, ok := old.Get("x"); ok {
		t.Fatal("removed instance is not closed")
	}
}

func TestFeatureCacheBuildFailureIsAtomic(t *testing.T) {
	freshFeatureCaches(t)
	LoadFeatureCaches(featureCacheConfig(10, 10))
	old := mustFeatureCache(t, "shared")
	old.Put("x", map[string]any{"v": 1})
	conf := featureCacheConfig(20, 20)
	conf.FeatureCacheConfs["new"] = featureCacheDefinition(10, 0)
	var created []*FeatureCache
	err := loadFeatureCachesWithFactory(conf, func(name string, conf recconf.FeatureCacheConfig) (*FeatureCache, error) {
		if len(created) == 1 {
			return nil, errors.New("build failed")
		}
		c, err := newFeatureCache(name, conf)
		created = append(created, c)
		return c, err
	})
	if err == nil {
		t.Fatal("expected factory error")
	}
	if mustFeatureCache(t, "shared") != old {
		t.Fatal("failed load changed the registry")
	}
	if _, ok := old.Get("x"); !ok {
		t.Fatal("failed load closed the old instance")
	}
	if _, err := GetFeatureCache("new"); err == nil {
		t.Fatal("failed load published a partial registry")
	}
	for _, c := range created {
		c.Put("x", map[string]any{"v": 1})
		if _, ok := c.Get("x"); ok {
			t.Fatal("unpublished instance was not closed")
		}
	}
}

func TestFeatureCacheInvalidConfigDoesNotMutateRegistry(t *testing.T) {
	freshFeatureCaches(t)
	LoadFeatureCaches(featureCacheConfig(10, 10))
	old := mustFeatureCache(t, "shared")
	cases := []recconf.FeatureCacheConfig{{}, {CacheSize: -1}, {CacheSize: 10, CacheTime: -1}}
	for _, conf := range cases {
		config := &recconf.RecommendConfig{FeatureCacheConfs: map[string]recconf.FeatureCacheConfig{"shared": conf}}
		if err := loadFeatureCachesWithFactory(config, newFeatureCache); err == nil {
			t.Fatalf("invalid policy accepted: %+v", conf)
		}
		if mustFeatureCache(t, "shared") != old {
			t.Fatal("invalid load replaced the old instance")
		}
	}
	config := &recconf.RecommendConfig{FeatureCacheConfs: map[string]recconf.FeatureCacheConfig{" ": {CacheSize: 10}}}
	if err := loadFeatureCachesWithFactory(config, newFeatureCache); err == nil {
		t.Fatal("blank cache name accepted")
	}
}

func TestFeatureCacheIsolatesTopLevelProperties(t *testing.T) {
	freshFeatureCaches(t)
	LoadFeatureCaches(featureCacheConfig(10, 10))
	c := mustFeatureCache(t, "shared")
	var nilSlice []int64
	var nilMap map[string]float32
	input := map[string]any{
		"large":     uint64(18446744073709551615),
		"ids":       []int64{9007199254740993},
		"embedding": []float32{1, 2},
		"weights":   map[int64][]float64{3: {4, 5}},
		"nested":    []any{map[string]any{"value": []string{"a"}}},
		"nilSlice":  nilSlice, "nilMap": nilMap, "nil": nil,
	}
	expected := CloneFeatureProperties(input)
	c.Put("u", input)
	input["ids"] = []int64{1}
	input["weights"] = map[int64][]float64{3: {99}}
	input["nested"] = []any{"changed"}
	input["added"] = true
	got, ok := c.Get("u")
	if !ok || !reflect.DeepEqual(got, expected) {
		t.Fatalf("write did not isolate values: %#v", got)
	}
	got["embedding"] = []float32{9}
	got["weights"] = map[int64][]float64{3: {10}}
	got["nested"] = []any{"read"}
	delete(got, "ids")
	next, _ := c.Get("u")
	if !reflect.DeepEqual(next, expected) {
		t.Fatalf("read did not isolate values: %#v", next)
	}
}

func TestCloneFeaturePropertiesSharesReadOnlyNestedValues(t *testing.T) {
	ids := []int64{9007199254740993}
	embedding := []float32{1, 2}
	weights := map[string]float64{"a": 0.5}
	input := map[string]any{"ids": ids, "embedding": embedding, "weights": weights}
	cloned := CloneFeatureProperties(input)
	if &cloned["ids"].([]int64)[0] != &ids[0] || &cloned["embedding"].([]float32)[0] != &embedding[0] {
		t.Fatal("read-only sequences and embeddings must not be copied")
	}
	if reflect.ValueOf(cloned["weights"]).Pointer() != reflect.ValueOf(weights).Pointer() {
		t.Fatal("read-only nested maps must not be copied")
	}
	cloned["ids"] = []int64{7}
	delete(cloned, "embedding")
	if input["ids"].([]int64)[0] != 9007199254740993 || len(input["embedding"].([]float32)) != 2 {
		t.Fatal("top-level replacement changed the source map")
	}
}

func TestCloneFeaturePropertiesNilAndEmpty(t *testing.T) {
	if CloneFeatureProperties(nil) != nil {
		t.Fatal("nil map must remain nil")
	}
	input := map[string]any{}
	cloned := CloneFeatureProperties(input)
	if cloned == nil {
		t.Fatal("empty map must remain non-nil")
	}
	cloned["added"] = 1
	if len(input) != 0 {
		t.Fatal("empty map was not cloned")
	}
}

func BenchmarkCloneFeatureProperties(b *testing.B) {
	for _, size := range []int{128, 16384} {
		name := "small"
		if size == 16384 {
			name = "large"
		}
		b.Run(name, func(b *testing.B) {
			properties := map[string]any{
				"uid": "u", "embedding": make([]float32, size), "sequence": make([]int64, size),
			}
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					cloned := CloneFeatureProperties(properties)
					if len(cloned) != 3 {
						b.Error("invalid clone")
					}
				}
			})
		})
	}
}

func TestFeatureCacheNegativeDoesNotOverwritePositive(t *testing.T) {
	freshFeatureCaches(t)
	LoadFeatureCaches(featureCacheConfig(10, 10))
	c := mustFeatureCache(t, "shared")
	c.Put("u", map[string]any{"value": 1})
	c.PutIfAbsent("u", map[string]any{})
	got, _ := c.Get("u")
	if got["value"] != 1 {
		t.Fatal("negative entry overwrote positive data")
	}
	c.PutIfAbsent("missing", map[string]any{})
	if got, ok := c.Get("missing"); !ok || len(got) != 0 {
		t.Fatal("empty result was not cached")
	}
}

func TestFeatureCacheReadersNotBlockedDuringBuild(t *testing.T) {
	freshFeatureCaches(t)
	LoadFeatureCaches(featureCacheConfig(10, 10))
	started, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		done <- loadFeatureCachesWithFactory(featureCacheConfig(20, 10), func(name string, conf recconf.FeatureCacheConfig) (*FeatureCache, error) {
			close(started)
			<-release
			return newFeatureCache(name, conf)
		})
	}()
	<-started
	read := make(chan error, 1)
	go func() { _, err := GetFeatureCache("shared"); read <- err }()
	select {
	case err := <-read:
		if err != nil {
			t.Error(err)
		}
	case <-time.After(2 * time.Second):
		t.Error("registry readers blocked during build")
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestFeatureCacheConcurrentReloadAndClose(t *testing.T) {
	freshFeatureCaches(t)
	LoadFeatureCaches(featureCacheConfig(100, 10))
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 100; j++ {
				if c, err := GetFeatureCache("shared"); err == nil {
					c.Put("u", map[string]any{"ids": []int64{1, 2}})
					c.PutIfAbsent("u", map[string]any{})
					c.Get("u")
				}
			}
		}()
	}
	close(start)
	for i := 0; i < 20; i++ {
		LoadFeatureCaches(featureCacheConfig(100+i, 10))
	}
	wg.Wait()
	c := mustFeatureCache(t, "shared")
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Close(); c.Get("u"); c.Put("u", map[string]any{}) }()
	}
	wg.Wait()
}

func TestFeatureCacheExpires(t *testing.T) {
	freshFeatureCaches(t)
	c, err := newFeatureCache("ttl", featureCacheDefinition(10, 1))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.Put("u", map[string]any{"v": 1})
	if _, ok := c.Get("u"); !ok {
		t.Fatal("immediate miss")
	}
	time.Sleep(1100 * time.Millisecond)
	if _, ok := c.Get("u"); ok {
		t.Fatal("expired entry was served")
	}
}

func TestFeatureCacheDefinitionReload(t *testing.T) {
	freshFeatureCaches(t)
	config := featureCacheConfig(10, 60)
	LoadFeatureCaches(config)
	old := mustFeatureCache(t, "shared")
	old.Put("u", map[string]any{"age": 18})
	conf := config.FeatureCacheConfs["shared"]
	conf.SelectFields = "*"
	config.FeatureCacheConfs["shared"] = conf
	LoadFeatureCaches(config)
	if mustFeatureCache(t, "shared") != old {
		t.Fatal("equivalent definition discarded data")
	}
	conf.FeatureStoreViewName = "new_view"
	config.FeatureCacheConfs["shared"] = conf
	LoadFeatureCaches(config)
	current := mustFeatureCache(t, "shared")
	if current == old {
		t.Fatal("changed definition reused old cache")
	}
	if _, ok := current.Get("u"); ok {
		t.Fatal("new source contains old data")
	}
	if old.Config().FeatureStoreViewName != "view" || current.Config().FeatureStoreViewName != "new_view" {
		t.Fatal("generation definitions are not immutable")
	}
	old.Put("late", map[string]any{"age": 99})
	if _, ok := current.Get("late"); ok {
		t.Fatal("late write polluted replacement")
	}
	conf.FeatureStore = "item"
	config.FeatureCacheConfs["shared"] = conf
	if err := loadFeatureCachesWithFactory(config, newFeatureCache); err == nil {
		t.Fatal("user/item scope changed for an existing name")
	}
	if mustFeatureCache(t, "shared") != current {
		t.Fatal("rejected definition mutated registry")
	}
}

func TestFeatureCacheChangedReadDefinitionResetsEntries(t *testing.T) {
	for _, change := range []func(*recconf.FeatureCacheConfig){
		func(c *recconf.FeatureCacheConfig) { c.FeatureStoreViewName = "new_view" },
		func(c *recconf.FeatureCacheConfig) { c.SelectFields = "age" },
		func(c *recconf.FeatureCacheConfig) { c.FeatureStoreOptions.DlrmHSTU = true },
		func(c *recconf.FeatureCacheConfig) {
			c.FeatureStoreViewName = ""
			c.FeatureStoreModelName = "model"
			c.FeatureStoreEntityName = "entity"
		},
	} {
		func() {
			freshFeatureCaches(t)
			config := featureCacheConfig(10, 60)
			LoadFeatureCaches(config)
			old := mustFeatureCache(t, "shared")
			old.Put("u", map[string]any{"age": 18})
			conf := config.FeatureCacheConfs["shared"]
			change(&conf)
			config.FeatureCacheConfs["shared"] = conf
			LoadFeatureCaches(config)
			current := mustFeatureCache(t, "shared")
			if current == old {
				t.Fatal("changed read definition did not rebuild cache")
			}
			if _, ok := current.Get("u"); ok {
				t.Fatal("changed read definition reused cached data")
			}
		}()
	}
}

func TestFeatureCacheConfigurationSnapshotIsIndependent(t *testing.T) {
	freshFeatureCaches(t)
	c, err := newFeatureCache("snapshot", featureCacheDefinition(10, 60))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	conf := c.Config()
	conf.FeatureStoreViewName = "changed"
	if c.Config().FeatureStoreViewName != "view" || c.Project() == nil {
		t.Fatal("configuration snapshot changed the instance")
	}
}

func TestFeatureCacheInvalidMetadataDoesNotPublish(t *testing.T) {
	freshFeatureCaches(t)
	LoadFeatureCaches(featureCacheConfig(10, 60))
	old := mustFeatureCache(t, "shared")
	for _, change := range []func(*recconf.FeatureCacheConfig){
		func(c *recconf.FeatureCacheConfig) { c.FeatureStoreName = "missing" },
		func(c *recconf.FeatureCacheConfig) {
			c.FeatureStoreViewName = ""
			c.FeatureStoreModelName = "model"
			c.FeatureStoreEntityName = "missing"
		},
	} {
		conf := featureCacheDefinition(10, 60)
		change(&conf)
		config := featureCacheConfig(10, 60)
		config.FeatureCacheConfs["shared"] = conf
		if err := loadFeatureCachesWithFactory(config, newFeatureCache); err == nil {
			t.Fatal("invalid metadata was accepted")
		}
		if mustFeatureCache(t, "shared") != old {
			t.Fatal("invalid metadata changed registry")
		}
	}
}

func TestFeatureCacheRejectsEntityOutsideModel(t *testing.T) {
	freshFeatureCaches(t)
	config := featureCacheConfig(10, 60)
	LoadFeatureCaches(config)
	old := mustFeatureCache(t, "shared")
	old.Put("u", map[string]any{"age": 18})
	for _, scope := range []string{"user", "item"} {
		for _, source := range []struct{ model, entity string }{
			{"model", "unrelated"}, {"empty_model", "entity"},
		} {
			conf := featureCacheDefinition(10, 60)
			conf.FeatureStore = scope
			conf.FeatureStoreViewName = ""
			conf.FeatureStoreModelName, conf.FeatureStoreEntityName = source.model, source.entity
			c, err := newFeatureCache("bad_source", conf)
			if err == nil {
				c.Close()
				t.Errorf("accepted entity %s outside model %s", source.entity, source.model)
			}
			if scope == "user" {
				config.FeatureCacheConfs["shared"] = conf
				if err := loadFeatureCachesWithFactory(config, newFeatureCache); err == nil {
					t.Error("invalid model/entity source was published")
				}
				if mustFeatureCache(t, "shared") != old {
					t.Fatal("failed source validation replaced old cache")
				}
				if data, ok := old.Get("u"); !ok || data["age"] != 18 {
					t.Fatal("failed source validation closed old cache")
				}
			}
		}
	}
	conf := featureCacheDefinition(10, 60)
	conf.FeatureStoreViewName = ""
	conf.FeatureStoreModelName, conf.FeatureStoreEntityName = "model", "entity"
	c, err := newFeatureCache("valid_model", conf)
	if err != nil {
		t.Fatalf("valid model/entity rejected: %v", err)
	}
	c.Close()
}

func TestFeatureCachePositiveIsVisibleAtCapacity(t *testing.T) {
	freshFeatureCaches(t)
	c, err := newFeatureCache("full", featureCacheDefinition(1, 60))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	for i := 0; i < 100; i++ {
		c.Put("fill", map[string]any{"v": 0})
		c.Put("positive", map[string]any{"v": i})
		c.PutIfAbsent("positive", map[string]any{})
		got, ok := c.Get("positive")
		if !ok || got["v"] != i {
			t.Fatalf("positive result is not synchronously visible at capacity: hit=%v data=%v", ok, got)
		}
	}
}

func TestFeatureCacheCapacityAndLRU(t *testing.T) {
	freshFeatureCaches(t)
	c, err := newFeatureCache("bounded", featureCacheDefinition(2, 60))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.Put("a", map[string]any{"v": 1})
	c.Put("b", map[string]any{"v": 2})
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected a hit")
	}
	c.Put("c", map[string]any{"v": 3})
	if _, ok := c.Get("b"); ok {
		t.Fatal("least recently used entry was not evicted")
	}
	for _, key := range []string{"a", "c"} {
		if _, ok := c.Get(key); !ok {
			t.Fatalf("active entry %s was evicted", key)
		}
	}
}
