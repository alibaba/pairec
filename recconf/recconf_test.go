package recconf

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
)

func TestCopyConfig(t *testing.T) {

	src := &RecommendConfig{
		RunMode: "dev",
		ListenConf: ListenConfig{
			HttpAddr: "",
			HttpPort: 80,
		},
		//		SortNames:   []string{"item_score"},
		//	    FilterNames: []string{"item_exposure_filter"},
	}

	dst := &RecommendConfig{
		RunMode: "dev",
		ListenConf: ListenConfig{
			HttpAddr: "",
			HttpPort: 80,
		},
	}

	fmt.Println(dst)
	CopyConfig(src, dst, func(name string) bool {
		return name == "SortNames"
	})
	fmt.Println(dst)
	CopyConfig(src, dst)
	fmt.Println(dst)
}

type MyConf struct {
	TestConf struct {
		K string
	}
}

func TestParseUserDefineConfs(t *testing.T) {
	Config = &RecommendConfig{
		UserDefineConfs: []byte("{\"k\": \"v\", \"TestConf\": {\"K\": \"V1\"}}"),
	}

	v, err := ParseUserDefineConfs[string](WithKey("k"))
	if err != nil {
		t.Error(err)
	} else if v != "v" {
		t.Errorf("expect v, got %s", v)
	}

	c, err := ParseUserDefineConfs[MyConf]()
	if err != nil {
		t.Error(err)
	} else if c.TestConf.K != "V1" {
		t.Errorf("expect V1, got %s", c.TestConf.K)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	ch := Subscribe()
	go func() {
		<-ch

		c, err = ParseUserDefineConfs[MyConf]()
		if err != nil {
			t.Error(err)
		} else if c.TestConf.K != "V2" {
			t.Errorf("expect V2, got %s", c.TestConf.K)
		}

		wg.Done()
	}()

	UpdateConf(&RecommendConfig{
		UserDefineConfs: []byte("{\"k\": \"v\", \"TestConf\": {\"K\": \"V2\"}}"),
	})

	wg.Wait()
}

func TestFeatureCacheConfigValidation(t *testing.T) {
	view := FeatureCacheConfig{FeatureStoreName: "fs", FeatureStore: "user", FeatureStoreViewName: "view", CacheSize: 10}
	model := FeatureCacheConfig{FeatureStoreName: "fs", FeatureStore: "user", FeatureStoreModelName: "model", FeatureStoreEntityName: "entity", CacheSize: 10}
	for _, conf := range []FeatureCacheConfig{view, model} {
		if err := conf.Validate(); err != nil {
			t.Fatalf("valid definition rejected: %v", err)
		}
	}
	for _, change := range []func(*FeatureCacheConfig){
		func(c *FeatureCacheConfig) { c.FeatureStoreName = "" },
		func(c *FeatureCacheConfig) { c.FeatureStore = "" },
		func(c *FeatureCacheConfig) { c.FeatureStore = "other" },
		func(c *FeatureCacheConfig) { c.FeatureStoreViewName = "" },
		func(c *FeatureCacheConfig) { c.FeatureStoreModelName = "model" },
		func(c *FeatureCacheConfig) { c.FeatureStoreEntityName = "entity" },
		func(c *FeatureCacheConfig) { c.CacheSize = 0 },
		func(c *FeatureCacheConfig) { c.CacheTime = -1 },
		func(c *FeatureCacheConfig) { c.FeatureStore = "item"; c.FeatureStoreOptions.DlrmHSTU = true },
	} {
		conf := view
		change(&conf)
		if err := conf.Validate(); err == nil {
			t.Fatalf("invalid definition accepted: %+v", conf)
		}
	}
	for _, change := range []func(*FeatureCacheConfig){
		func(c *FeatureCacheConfig) { c.FeatureStoreEntityName = "" },
		func(c *FeatureCacheConfig) { c.SelectFields = "age" },
		func(c *FeatureCacheConfig) { c.SelectFields = "*" },
	} {
		conf := model
		change(&conf)
		if err := conf.Validate(); err == nil {
			t.Fatalf("invalid model definition accepted: %+v", conf)
		}
	}
}

func TestFeatureCacheDefinitionJSON(t *testing.T) {
	var conf RecommendConfig
	err := json.Unmarshal([]byte(`{"FeatureCacheConfs":{"shared":{"FeatureStoreName":"fs","FeatureStore":"user","FeatureStoreViewName":"view","SelectFields":"age","FeatureStoreOptions":{"DlrmHSTU":true},"CacheSize":10,"CacheTime":600}}}`), &conf)
	if err != nil {
		t.Fatal(err)
	}
	cache := conf.FeatureCacheConfs["shared"]
	if cache.FeatureStoreName != "fs" || cache.FeatureStoreViewName != "view" || cache.SelectFields != "age" || !cache.FeatureStoreOptions.DlrmHSTU {
		t.Fatalf("cache source was not parsed: %+v", cache)
	}
}

func TestFeatureCacheRequirements(t *testing.T) {
	conf := SceneFeatureConfig{FeatureLoadConfs: []FeatureLoadConfig{{FeatureDaoConf: FeatureDaoConfig{
		DaoConfig:        DaoConfig{AdapterType: DataSource_Type_FeatureStore},
		FeatureCacheName: "shared",
		FeatureStore:     "user",
	}}}}
	modules := (RecommendConfig{FeatureStoreConfs: map[string]FeatureStoreConfig{"fs": {}}}).GetModules()
	if err := conf.Requirements().Check(modules); err == nil {
		t.Fatal("missing shared cache reference was accepted")
	}
	modules = (RecommendConfig{
		FeatureStoreConfs: map[string]FeatureStoreConfig{"fs": {}},
		FeatureCacheConfs: map[string]FeatureCacheConfig{"shared": {FeatureStoreName: "fs", FeatureStore: "user", FeatureStoreViewName: "view", CacheSize: 10}},
	}).GetModules()
	if err := conf.Requirements().Check(modules); err != nil {
		t.Fatal(err)
	}
	cache := FeatureCacheConfig{FeatureStoreName: "fs", FeatureStore: "user", FeatureStoreViewName: "view", CacheSize: 10}
	if err := cache.Requirements().Check(modules); err != nil {
		t.Fatal(err)
	}
	delete(modules, ModuleIndex{Type: FeatureStoreConfig{}.ModuleType(), Name: "fs"})
	if err := cache.Requirements().Check(modules); err == nil {
		t.Fatal("cache data-source dependency was not checked")
	}
	results, err := CheckRecommendConfig(`{"FeatureCacheConfs":{"shared":{"CacheSize":0}}}`)
	if err != nil {
		t.Fatal(err)
	}
	index := ModuleIndex{Type: FeatureCacheConfig{}.ModuleType(), Name: "shared"}
	if results[index] == nil || len(results[index].Issues) == 0 {
		t.Fatal("invalid named cache is not reported by config checking")
	}
}
