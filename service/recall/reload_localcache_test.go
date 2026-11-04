package recall

import (
	"bytes"
	"fmt"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	reccontext "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	pcache "github.com/alibaba/pairec/v2/persist/cache"
	"github.com/alibaba/pairec/v2/recconf"
)

const goCacheJanitorFrame = "go-cache.(*janitor).Run("

func countGoroutineFrames(frame string) int {
	var buf bytes.Buffer
	_ = pprof.Lookup("goroutine").WriteTo(&buf, 2)
	return strings.Count(buf.String(), frame)
}
func heapAllocAfterGC() uint64 {
	runtime.GC()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}
func settle(cond func() bool) {
	for i := 0; i < 50; i++ {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
		if cond() {
			return
		}
	}
}

type localCacheTestParams map[string]any

func (p localCacheTestParams) GetParameter(name string) any { return p[name] }

type localCacheTestDao struct{ calls atomic.Int64 }

func (d *localCacheTestDao) items() []*module.Item {
	d.calls.Add(1)
	item := module.NewItem("remote")
	item.Score = 3
	return []*module.Item{item}
}
func (d *localCacheTestDao) ListItemsByItem(*module.User, *reccontext.RecommendContext) []*module.Item {
	return d.items()
}
func (d *localCacheTestDao) ListItemsByMultiItemIds(*module.User, *reccontext.RecommendContext, []any) map[string][]*module.Item {
	return map[string][]*module.Item{"seed": d.items()}
}
func (d *localCacheTestDao) ListItemsByUser(*module.User, *reccontext.RecommendContext) []*module.Item {
	return d.items()
}
func (d *localCacheTestDao) TriggerValue(*module.User) string { return "group" }

func localCacheTestConf(name, typ string) recconf.RecallConfig {
	return recconf.RecallConfig{Name: name, RecallType: typ, RecallCount: 5,
		CacheAdapter: "localCache", CachePrefix: name + "_", CacheTime: 7200,
		CacheConfig: `{"defaultExpiration":3600,"cleanupInterval":3600}`}
}

// Test the same cache preparation/assignment used by Load, with real A request
// implementations and a deterministic DAO instead of an external data source.
func localCacheTestRecall(conf recconf.RecallConfig, old Recall, dao *localCacheTestDao) Recall {
	adjusted, inherited, sign := localCacheConfig(conf, old)
	base := NewBaseRecall(adjusted)
	var r Recall
	switch conf.RecallType {
	case "ItemCollaborativeFilterRecall":
		r = &ItemCollaborativeFilterRecall{BaseRecall: base, itemCollaborativeDao: dao}
	case "UserGlobalHotRecall":
		r = &UserGlobalHotRecall{BaseRecall: base, userGroupHotRecallDao: dao}
	case "UserGroupHotRecall":
		r = &UserGroupHotRecall{BaseRecall: base, userGroupHotRecallDao: dao}
	}
	applyLocalCache(r, conf, inherited, sign)
	return r
}
func localCacheTestContext() *reccontext.RecommendContext {
	ctx := reccontext.NewRecommendContext()
	ctx.Param = localCacheTestParams{"item_id": "seed"}
	return ctx
}
func localCacheTestKey(conf recconf.RecallConfig) (string, string) {
	switch conf.RecallType {
	case "UserGlobalHotRecall":
		return conf.Name, "cached::2"
	case "UserGroupHotRecall":
		return conf.CachePrefix + "group", "cached::2"
	default:
		return "seed", "cached:2"
	}
}

func TestReloadLocalCacheInheritance(t *testing.T) {
	for _, typ := range []string{"ItemCollaborativeFilterRecall", "UserGlobalHotRecall", "UserGroupHotRecall"} {
		t.Run(typ, func(t *testing.T) {
			var allocations int
			pcache.Register("localCache", func() pcache.Cache { allocations++; return pcache.NewLocalCache() })
			t.Cleanup(func() { pcache.Register("localCache", pcache.NewLocalCache) })
			conf := localCacheTestConf(t.Name(), typ)
			dao := &localCacheTestDao{}
			old := localCacheTestRecall(conf, nil, dao)
			base := localCacheBase(old)
			key, value := localCacheTestKey(conf)
			_ = base.cache.Put(key, value, time.Hour)
			for i := 0; i < 5; i++ {
				conf.CacheTime = 3600 + i
				fresh := localCacheTestRecall(conf, old, dao)
				if fresh == old || localCacheBase(fresh).cache != base.cache || localCacheBase(fresh).cacheTime != conf.CacheTime {
					t.Fatal("TTL-only reload must assign the old cache to the new recall")
				}
				got := fresh.GetCandidateItems(module.NewUser("u"), localCacheTestContext())
				if len(got) != 1 || got[0].Id != "cached" || got[0].Score != 2 || got[0].RetrieveId != conf.Name {
					t.Fatal("inherited cache no longer gives the same candidates")
				}
				old = fresh
			}
			if allocations != 1 || dao.calls.Load() != 0 || base.cacheTime != 7200 {
				t.Fatal("reload allocated a throwaway cache, fetched remotely, or modified the old instance")
			}
		})
	}
}

func TestReloadLocalCacheIncompatible(t *testing.T) {
	changes := map[string]func(*recconf.RecallConfig){
		"count":   func(c *recconf.RecallConfig) { c.RecallCount++ },
		"view":    func(c *recconf.RecallConfig) { c.ItemCollaborativeDaoConf.FeatureStoreViewName = "other" },
		"prefix":  func(c *recconf.RecallConfig) { c.CachePrefix = "other" },
		"config":  func(c *recconf.RecallConfig) { c.CacheConfig = `{"cleanupInterval":1}` },
		"disable": func(c *recconf.RecallConfig) { c.CacheAdapter = "" },
		"type":    func(c *recconf.RecallConfig) { c.RecallType = "UserGlobalHotRecall" },
		"name":    func(c *recconf.RecallConfig) { c.Name += "_other" },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			conf := localCacheTestConf(t.Name(), "ItemCollaborativeFilterRecall")
			dao := &localCacheTestDao{}
			old := localCacheTestRecall(conf, nil, dao)
			base := localCacheBase(old)
			_ = base.cache.Put("seed", "old:1", time.Hour)
			change(&conf)
			fresh := localCacheBase(localCacheTestRecall(conf, old, dao))
			if fresh.cache == base.cache || base.cache.Get("seed") != "old:1" {
				t.Fatal("incompatible reload reused or cleared the old cache")
			}
			if fresh.cache != nil && fresh.cache.Get("seed") != nil {
				t.Fatal("old data entered the new cache")
			}
		})
	}
}

func TestReloadLocalCacheActualPutTTL(t *testing.T) {
	for _, typ := range []string{"ItemCollaborativeFilterRecall", "UserGlobalHotRecall", "UserGroupHotRecall"} {
		t.Run(typ, func(t *testing.T) {
			conf := localCacheTestConf(t.Name(), typ)
			dao := &localCacheTestDao{}
			old := localCacheTestRecall(conf, nil, dao)
			conf.CacheTime = 60
			fresh := localCacheTestRecall(conf, old, dao)
			oldBase, freshBase := localCacheBase(old), localCacheBase(fresh)
			// Instrument after inheritance, before request goroutines begin.
			oldSpy := &realTimeU2IRecallTestCache{Cache: oldBase.cache, puts: make(chan realTimeU2IRecallTestPut, 1)}
			newSpy := &realTimeU2IRecallTestCache{Cache: freshBase.cache, puts: make(chan realTimeU2IRecallTestPut, 1)}
			oldBase.cache, freshBase.cache = oldSpy, newSpy
			fresh.GetCandidateItems(module.NewUser("u"), localCacheTestContext())
			put := waitRealTimeU2IRecallCachePut(t, newSpy)
			if put.ttl != time.Minute {
				t.Fatalf("new TTL=%v", put.ttl)
			}
			_ = newSpy.Delete(put.key)
			old.GetCandidateItems(module.NewUser("u"), localCacheTestContext())
			if put = waitRealTimeU2IRecallCachePut(t, oldSpy); put.ttl != 2*time.Hour {
				t.Fatalf("old TTL=%v", put.ttl)
			}
		})
	}
}

func TestReloadLocalCacheExistingExpiryAndConcurrentReads(t *testing.T) {
	conf := localCacheTestConf(t.Name(), "ItemCollaborativeFilterRecall")
	dao := &localCacheTestDao{}
	old := localCacheTestRecall(conf, nil, dao)
	base := localCacheBase(old)
	_ = base.cache.Put("expiry", "old:1", 30*time.Millisecond)
	_ = base.cache.Put("seed", "cached:2", time.Hour)
	conf.CacheTime = 60
	fresh := localCacheTestRecall(conf, old, dao)
	var wg sync.WaitGroup
	for _, r := range []Recall{old, fresh} {
		wg.Add(1)
		go func(r Recall) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				r.GetCandidateItems(module.NewUser("u"), localCacheTestContext())
			}
		}(r)
	}
	wg.Wait()
	deadline := time.Now().Add(time.Second)
	for base.cache.Get("expiry") != nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if localCacheBase(fresh).cache.Get("expiry") != nil {
		t.Fatal("reload extended an old entry's expiry")
	}
}

func TestReloadLocalCacheIndependentConstruction(t *testing.T) {
	conf := localCacheTestConf(t.Name(), "ItemCollaborativeFilterRecall")
	first, second := NewBaseRecall(conf), NewBaseRecall(conf)
	if first.cache == second.cache {
		t.Fatal("ordinary constructors must keep independent caches")
	}
}

// Keep the original Load/GC reproduction: MockRecall has no A inheritance,
// so each reload replaces its cache and the old go-cache finalizer can run.
func TestReloadLocalCacheRecallReleasesOldCache(t *testing.T) {
	const name = "reload_gc_reproduction"
	for i := 0; i < 3; i++ {
		heapAllocAfterGC()
		time.Sleep(10 * time.Millisecond)
	}
	baseHeap := heapAllocAfterGC()
	baseJanitors := countGoroutineFrames(goCacheJanitorFrame)
	t.Cleanup(func() {
		recallsMu.Lock()
		delete(recalls, name)
		delete(recallSigns, name)
		recallsMu.Unlock()
	})
	load := func(ttl int) {
		Load(&recconf.RecommendConfig{RecallConfs: []recconf.RecallConfig{{Name: name, RecallType: "MockRecall",
			CacheAdapter: "localCache", CacheConfig: `{"cleanupInterval":3600}`, CacheTime: ttl}}})
	}
	fill := func() {
		r, _ := GetRecall(name)
		for i := 0; i < 100000; i++ {
			_ = r.(*MockRecall).cache.Put(fmt.Sprint(i), fmt.Sprintf("%d:1,%d:2,%d:3", i, i+1, i+2), time.Hour)
		}
	}
	load(7200)
	fill()
	oneCache := int64(heapAllocAfterGC()) - int64(baseHeap)
	for i := 0; i < 5; i++ {
		load(3600 + i)
		fill()
	}
	settle(func() bool { return countGoroutineFrames(goCacheJanitorFrame)-baseJanitors == 1 })
	janitors := countGoroutineFrames(goCacheJanitorFrame) - baseJanitors
	retained := int64(heapAllocAfterGC()) - int64(baseHeap)
	t.Logf("reloads=5 oneCache=%.1fMB retained=%.1fMB janitors=%d", float64(oneCache)/1e6, float64(retained)/1e6, janitors)
	if janitors != 1 || retained > 2*oneCache {
		t.Fatal("old go-cache instances were not reclaimed")
	}
}
