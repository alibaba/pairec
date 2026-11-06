package feature

import (
	"fmt"
	"sync"
	"testing"
	"time"

	reccontext "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/goburrow/cache"
)

func stateReloadCache(dao module.ItemStateFilterDao) cache.Cache {
	switch d := dao.(type) {
	case *module.ItemStateFilterFeatureStoreDao:
		return d.ItemStateCache()
	case *module.ItemStateFilterHologresDao:
		return d.ItemStateCache()
	}
	return nil
}

func reloadItemState(t *testing.T, conf recconf.FilterConfig, inherited cache.Cache) module.ItemStateFilterDao {
	t.Helper()
	var trans module.FeatureTransFunc
	if len(conf.Features) > 0 {
		compiled := LoadWithConfig(recconf.FeatureLoadConfig{Features: conf.Features})
		trans = func(user *module.User, item *module.Item, ctx *reccontext.RecommendContext) {
			compiled.LoadFeatures(user, []*module.Item{item}, ctx)
		}
	}
	dao := module.NewItemStateFilterDaoWithCache(conf, trans, inherited)
	if c := stateReloadCache(dao); c != nil {
		reloadTestCaches[c] = true
	}
	return dao
}

func stateReloadData(dao module.ItemStateFilterDao, key string) []*module.Item {
	item := module.NewItem(key)
	item.AddProperty("local", "yes")
	return dao.Filter(module.NewUser("u"), []*module.Item{item}, reccontext.NewRecommendContext())
}

func TestItemStateFilterCacheReload(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	fsConf, view := reloadTestConfig(t, "item")
	view.fetch = func(keys []any) []map[string]any {
		var rows []map[string]any
		for _, key := range keys {
			if key != "missing" {
				rows = append(rows, map[string]any{"id": key, "value": "remote"})
			}
		}
		return rows
	}
	holoConf, driver := newReloadHologresConfig(t)
	for _, test := range []struct {
		name  string
		dao   recconf.DaoConfig
		value string
		calls func() int32
	}{
		{"fs", fsConf.FeatureDaoConf.DaoConfig, "remote", view.calls.Load},
		{"holo", holoConf.FeatureDaoConf.DaoConfig, "hologres_remote", driver.queries.Load},
	} {
		t.Run(test.name, func(t *testing.T) {
			reloadTestService(t)
			conf := recconf.FilterConfig{Name: "state_" + test.name, FilterType: "ItemStateFilter", ItemStateCacheSize: 100, ItemStateCacheTime: 60,
				ItemStateDaoConf: recconf.ItemStateDaoConfig{DaoConfig: test.dao, ItemFieldName: "id", SelectFields: "value", DefaultFieldValues: map[string]any{"value": "default"}}}
			old := reloadItemState(t, conf, nil)
			c := stateReloadCache(old)
			if c == nil {
				t.Fatal("private state cache was not created")
			}
			before := test.calls()
			data := stateReloadData(old, "same")
			if len(data) != 1 || test.calls() != before+1 {
				t.Fatal("state DAO did not fetch data")
			}
			conf.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: "copy", FeatureStore: "item", FeatureSource: "item:value"}}
			conf.FilterParams = []recconf.FilterParamConfig{{Name: "copy", Domain: "item", Operator: "equal", Type: "string", Value: test.value}}
			fresh := reloadItemState(t, conf, c)
			if fresh == old || stateReloadCache(fresh) != c {
				t.Fatal("condition/conversion reload did not inherit the cache")
			}
			data = stateReloadData(fresh, "same")
			if len(data) != 1 || data[0].StringProperty("copy") != test.value || test.calls() != before+1 {
				t.Fatal("new conversion did not run on inherited cache data")
			}
			// An empty conversion must not re-read all item properties for conditions.
			conf.Features = nil
			conf.FilterParams = []recconf.FilterParamConfig{{Name: "local", Domain: "item", Operator: "equal", Type: "string", Value: "yes"}}
			without := reloadItemState(t, conf, c)
			if stateReloadCache(without) != c {
				t.Fatal("removing conversions replaced the cache")
			}
			for _, key := range []string{"same", "new", "missing"} {
				if got := stateReloadData(without, key); len(got) != 0 {
					t.Fatal("nil conversion was replaced by an empty callback")
				}
			}
			conf.FilterParams = []recconf.FilterParamConfig{{Name: "value", Domain: "item", Operator: "equal", Type: "string", Value: "default"}}
			negative := reloadItemState(t, conf, c)
			beforeNegative := test.calls()
			if len(stateReloadData(negative, "missing")) != 1 || test.calls() != beforeNegative {
				t.Fatal("default/negative cache was not inherited")
			}
			conf.ItemStateCacheTime++
			replaced := reloadItemState(t, conf, nil)
			if stateReloadCache(replaced) == c {
				t.Fatal("TTL change reused the old cache")
			}
			if _, ok := c.GetIfPresent("same"); !ok {
				t.Fatal("old cache was cleared before the grace period")
			}
			conf.ItemStateDaoConf.FeatureStoreName = "missing"
			conf.ItemStateDaoConf.HologresName = "missing"
			func() {
				defer func() {
					if recover() == nil {
						t.Error("expected source construction failure")
					}
				}()
				reloadItemState(t, conf, c)
			}()
			if _, ok := c.GetIfPresent("same"); !ok {
				t.Fatal("failed state construction cleared an inherited cache")
			}
		})
	}
}

func TestItemStateFilterCacheReloadConcurrent(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadTestService(t)
	fsConf, _ := reloadTestConfig(t, "item")
	conf := recconf.FilterConfig{Name: "concurrent_state", FilterType: "ItemStateFilter", ItemStateCacheSize: 100,
		ItemStateDaoConf: recconf.ItemStateDaoConfig{DaoConfig: fsConf.FeatureDaoConf.DaoConfig, ItemFieldName: "id"}}
	f := reloadItemState(t, conf, nil)
	c := stateReloadCache(f)
	c.Put("same", map[string]any{"value": "cached"})
	var wg sync.WaitGroup
	for worker := 0; worker < 3; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				if worker == 0 {
					updated := conf
					updated.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: fmt.Sprint("copy", i), FeatureStore: "item", FeatureSource: "item:value"}}
					if stateReloadCache(reloadItemState(t, updated, c)) != c {
						t.Error("compatible DAO construction changed the cache")
					}
				} else {
					if len(stateReloadData(f, "same")) != 1 {
						t.Error("concurrent filtering lost cached data")
					}
				}
			}
		}(worker)
	}
	wg.Wait()
	if stateReloadCache(f) != c {
		t.Fatal("concurrent construction changed the original cache")
	}
}

func TestItemStateFilterDaoReloadInjection(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	fsConf, _ := reloadTestConfig(t, "item")
	holoConf, _ := newReloadHologresConfig(t)
	for _, daoConf := range []recconf.DaoConfig{fsConf.FeatureDaoConf.DaoConfig, holoConf.FeatureDaoConf.DaoConfig} {
		conf := recconf.FilterConfig{ItemStateCacheSize: 100, ItemStateCacheTime: 60,
			ItemStateDaoConf: recconf.ItemStateDaoConfig{DaoConfig: daoConf, ItemFieldName: "id"}}
		c := cache.New(cache.WithMaximumSize(100))
		t.Cleanup(func() { _ = c.Close() })
		dao := module.NewItemStateFilterDaoWithCache(conf, nil, c)
		switch d := dao.(type) {
		case *module.ItemStateFilterFeatureStoreDao:
			if d.ItemStateCache() != c {
				t.Fatal("FS injection created a different cache")
			}
		case *module.ItemStateFilterHologresDao:
			if d.ItemStateCache() != c {
				t.Fatal("Holo injection created a different cache")
			}
		}
		conf.GenerateUserDataExpr = "invalid("
		func() {
			defer func() {
				if recover() == nil {
					t.Error("expected invalid generation expression")
				}
			}()
			module.NewItemStateFilterDaoWithCache(conf, nil, c)
		}()
		c.Put("late", "allowed")
		if _, ok := c.GetIfPresent("late"); !ok {
			t.Fatal("failed DAO construction closed an inherited cache")
		}
	}
}

func TestItemStateDaoReloadLatePut(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadTestService(t)
	fsConf, view := reloadTestConfig(t, "item")
	entered, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	view.fetch = func(keys []any) []map[string]any {
		close(entered)
		<-release
		return []map[string]any{{"id": keys[0], "value": "old_remote"}}
	}
	conf := recconf.FilterConfig{ItemStateCacheSize: 100, ItemStateCacheTime: 60,
		ItemStateDaoConf: recconf.ItemStateDaoConfig{DaoConfig: fsConf.FeatureDaoConf.DaoConfig, ItemFieldName: "id"}}
	removed := make(chan struct{})
	c := cache.New(cache.WithMaximumSize(100), cache.WithRemovalListener(func(key cache.Key, _ cache.Value) {
		if key == "sentinel" {
			close(removed)
		}
	}))
	reloadTestCaches[c] = true
	old := module.NewItemStateFilterDaoWithCache(conf, nil, c)
	c.Put("sentinel", map[string]any{})
	item := module.NewItem("same")
	done := make(chan []*module.Item, 1)
	go func() {
		done <- old.Filter(module.NewUser("u"), []*module.Item{item}, reccontext.NewRecommendContext())
	}()
	<-entered
	conf.ItemStateCacheTime++
	fresh := reloadItemState(t, conf, nil)
	c.InvalidateAll()
	select {
	case <-removed:
	case <-time.After(2 * time.Second):
		t.Fatal("old state attributes were not physically removed")
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case items := <-done:
		if len(items) != 1 || items[0].StringProperty("value") != "old_remote" {
			t.Fatal("old filtering request could not finish after data invalidation")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("old filtering request did not finish")
	}
	if _, ok := c.GetIfPresent("same"); !ok {
		t.Fatal("old state cache must allow late Put after invalidation")
	}
	if _, ok := stateReloadCache(fresh).GetIfPresent("same"); ok {
		t.Fatal("old request repopulated the new cache")
	}
}
