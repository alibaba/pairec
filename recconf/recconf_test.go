package recconf

import (
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
	for _, conf := range []FeatureCacheConfig{{CacheSize: 1}, {CacheSize: 10, CacheTime: 600}} {
		if err := conf.Validate(); err != nil {
			t.Fatalf("valid cache policy rejected: %v", err)
		}
	}
	for _, conf := range []FeatureCacheConfig{{}, {CacheSize: -1}, {CacheSize: 1, CacheTime: -1}} {
		if err := conf.Validate(); err == nil {
			t.Fatalf("invalid cache policy accepted: %+v", conf)
		}
	}
}

func TestFeatureCacheRequirements(t *testing.T) {
	conf := SceneFeatureConfig{FeatureLoadConfs: []FeatureLoadConfig{{FeatureDaoConf: FeatureDaoConfig{
		DaoConfig:        DaoConfig{AdapterType: DataSource_Type_FeatureStore, FeatureStoreName: "fs"},
		FeatureCacheName: "shared",
	}}}}
	modules := (RecommendConfig{FeatureStoreConfs: map[string]FeatureStoreConfig{"fs": {}}}).GetModules()
	if err := conf.Requirements().Check(modules); err == nil {
		t.Fatal("missing shared cache reference was accepted")
	}
	modules = (RecommendConfig{
		FeatureStoreConfs: map[string]FeatureStoreConfig{"fs": {}},
		FeatureCacheConfs: map[string]FeatureCacheConfig{"shared": {CacheSize: 10}},
	}).GetModules()
	if err := conf.Requirements().Check(modules); err != nil {
		t.Fatal(err)
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
