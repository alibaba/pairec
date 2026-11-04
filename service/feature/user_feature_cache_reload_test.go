package feature

import (
	"fmt"
	"sync"
	"testing"
	"time"

	reccontext "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
	"github.com/goburrow/cache"
)

func reloadUserTestService(t *testing.T) {
	t.Helper()
	reloadTestService(t)
	original := userFeatureService
	userFeatureService = NewUserFeatureService()
	t.Cleanup(func() { userFeatureService = original })
}

func reloadUserScene(scene string, confs ...recconf.FeatureLoadConfig) {
	UserLoadFeatureConfig(&recconf.RecommendConfig{UserFeatureConfs: map[string]recconf.SceneFeatureConfig{
		scene: {FeatureLoadConfs: confs},
	}})
	for _, f := range userFeatureService.FeatureSceneMap[scene] {
		if c := f.featureCache(); c != nil {
			reloadTestCaches[c] = true
		}
	}
}

func userReloadContext(scene string) *reccontext.RecommendContext {
	ctx := reccontext.NewRecommendContext()
	ctx.Param = localReloadParams{"scene": scene}
	return ctx
}

func loadUserReloadRequest(scene, key string) *module.User {
	u := module.NewUser(key)
	u.AddProperty("id", key)
	userFeatureService.LoadUserFeatures(u, userReloadContext(scene))
	return u
}

func TestUserFeatureCacheReload(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	fsConf, view := reloadTestConfig(t, "user")
	holoConf, holoDriver := newReloadHologresConfig(t)
	for _, test := range []struct {
		name  string
		conf  recconf.FeatureLoadConfig
		value string
		calls func() int32
	}{
		{"featurestore", fsConf, "user", view.calls.Load},
		{"hologres", holoConf, "hologres_remote", holoDriver.queries.Load},
	} {
		t.Run(test.name, func(t *testing.T) {
			reloadUserTestService(t)
			conf := test.conf
			reloadUserScene("test", conf)
			old := userFeatureService.FeatureSceneMap["test"][0]
			if old.daoSign == "" || old.featureCache() == nil {
				t.Fatal("UserFeatureConfs did not enter cache inheritance")
			}
			before := test.calls()
			if loadUserReloadRequest("test", "same").StringProperty("value") != test.value || test.calls() != before+1 {
				t.Fatal("user DAO did not successfully fetch data")
			}
			conf.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: "copy",
				FeatureStore: "user", FeatureSource: "user:value"}}
			for i := 0; i < 5; i++ {
				conf.Features[0].FeatureName = fmt.Sprint("copy", i)
				reloadUserScene("test", conf)
				fresh := userFeatureService.FeatureSceneMap["test"][0]
				if fresh.featureDao == old.featureDao || fresh.featureCache() != old.featureCache() {
					t.Fatal("reload must inherit only the cache, not the DAO")
				}
				if got := loadUserReloadRequest("test", "same").StringProperty(conf.Features[0].FeatureName); got != test.value {
					t.Fatal("new conversion did not run on cached user data")
				}
				if len(userFeatureService.FeatureAsyncLoadMap) != 1 || len(reloadTestInvalidations) != 0 || test.calls() != before+1 {
					t.Fatal("compatible reload retained old Feature keys, invalidated, or fetched again")
				}
			}
			reloadUserScene("other", conf)
			other := userFeatureService.FeatureSceneMap["other"][0]
			if other.featureCache() == old.featureCache() {
				t.Fatal("different user scenes shared a cache")
			}
			reloadScene(conf)
			if featureService.FeatureSceneMap["test"][0].featureCache() == old.featureCache() {
				t.Fatal("FeatureConfs and UserFeatureConfs shared a cache")
			}
			conf.FeatureDaoConf.CacheTime++
			reloadUserScene("test", conf)
			fresh := userFeatureService.FeatureSceneMap["test"][0]
			if fresh.featureCache() == old.featureCache() || len(reloadTestInvalidations) != 1 || reloadTestInvalidations[0].delay != 30*time.Second {
				t.Fatal("TTL change did not create a separate cache and schedule invalidation")
			}
			if _, ok := old.featureCache().GetIfPresent("same"); !ok {
				t.Fatal("old cache was cleared before the delay")
			}
			flushReloadTestInvalidations()
			waitReloadCacheMiss(t, old.featureCache(), "same")
			if len(userFeatureService.FeatureAsyncLoadMap) != 2 {
				t.Fatal("reload dropped another scene's keys or retained obsolete keys")
			}
		})
	}
}

func TestUserFeatureReloadFailure(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadUserTestService(t)
	conf, _ := reloadTestConfig(t, "user")
	conf.FeatureDaoConf.FeatureAsyncLoad = true
	reloadUserScene("test", conf)
	old := userFeatureService.FeatureSceneMap["test"][0]
	oldCache := old.featureCache()
	oldCache.Put("same", map[string]any{"value": "old"})
	sign := userFeatureService.FeatureSceneSigns["test"]
	waitReloadCacheCount(t, 1)
	base := countGoroutineFrames(goburrowProcessFrame)
	changed := conf
	changed.FeatureDaoConf.CacheSize++
	bad := conf
	bad.Features = []recconf.FeatureConfig{{FeatureType: "unknown"}}
	for _, invalid := range []recconf.FeatureLoadConfig{bad, {FeatureDaoConf: recconf.FeatureDaoConfig{FeatureStore: "item"}}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("expected invalid user configuration to fail")
				}
			}()
			reloadUserScene("test", conf, changed, invalid)
		}()
		if userFeatureService.FeatureSceneMap["test"][0] != old || userFeatureService.FeatureSceneSigns["test"] != sign || len(userFeatureService.FeatureAsyncLoadMap) != 1 {
			t.Fatal("failed reload changed the published scene or async keys")
		}
		if !userFeatureService.FeatureAsyncLoadMap[old] || len(reloadTestInvalidations) != 0 {
			t.Fatal("failed reload changed async behavior or scheduled invalidation")
		}
		if _, ok := oldCache.GetIfPresent("same"); !ok {
			t.Fatal("failed reload cleared an inherited cache")
		}
		waitReloadCacheCount(t, base)
	}
}

type blockingUserReloadDAO struct {
	entered chan struct{}
	release chan struct{}
	done    chan struct{}
}

func (d *blockingUserReloadDAO) FeatureFetch(*module.User, []*module.Item, *reccontext.RecommendContext) {
	close(d.entered)
	<-d.release
	close(d.done)
}

func TestUserFeatureReloadAsync(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	for _, callback := range []bool{false, true} {
		for _, custom := range []bool{false, true} {
			t.Run(fmt.Sprintf("callback=%v/custom=%v", callback, custom), func(t *testing.T) {
				reloadUserTestService(t)
				conf, _ := reloadTestConfig(t, "user")
				conf.FeatureDaoConf.FeatureAsyncLoad = true
				reloadUserScene("test", conf)
				old := userFeatureService.FeatureSceneMap["test"][0]
				d := &blockingUserReloadDAO{make(chan struct{}), make(chan struct{}), make(chan struct{})}
				old.featureDao = d
				var release sync.Once
				t.Cleanup(func() { release.Do(func() { close(d.release) }) })
				customDone := make(chan struct{})
				if custom {
					userFeatureService.AddFeatureFunc("test", func(_ *module.User, items []*module.Item, _ *reccontext.RecommendContext) []*module.Item {
						close(customDone)
						return items
					})
				}
				u := module.NewUser("same")
				returned := make(chan struct{})
				go func() {
					defer close(returned)
					if callback {
						userFeatureService.LoadUserFeaturesForCallback(u, userReloadContext("test"))
					} else {
						userFeatureService.LoadUserFeatures(u, userReloadContext("test"))
					}
				}()
				select {
				case <-d.entered:
				case <-time.After(2 * time.Second):
					t.Fatal("async DAO did not start")
				}
				select {
				case <-returned:
				case <-time.After(2 * time.Second):
					t.Fatal("async feature unexpectedly blocked the caller")
				}
				conf.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: "copy", FeatureSource: "user:value"}}
				reloadUserScene("test", conf)
				if _, exists := userFeatureService.FeatureAsyncLoadMap[old]; exists || !userFeatureService.isAsyncLoad(old) {
					t.Fatal("obsolete Feature key retained or lost its original async flag")
				}
				fresh := userFeatureService.FeatureSceneMap["test"][0]
				userFeatureService.FeatureAsyncLoadMap[fresh] = false
				if userFeatureService.isAsyncLoad(fresh) {
					t.Fatal("current map override was ignored")
				}
				if custom && !callback {
					select {
					case <-customDone:
						t.Fatal("custom feature callback ran before the async DAO finished")
					default:
					}
				}
				release.Do(func() { close(d.release) })
				<-d.done
				select {
				case <-u.FeatureAsyncLoadCh():
				case <-time.After(2 * time.Second):
					t.Fatal("async load counter did not return to zero")
				}
				if custom && !callback {
					<-customDone
				}
			})
		}
	}
}

func TestUserFeatureReloadLatePut(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadUserTestService(t)
	conf, view := reloadTestConfig(t, "user")
	conf.FeatureDaoConf.FeatureAsyncLoad = true
	entered, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	view.fetch = func(keys []any) []map[string]any {
		close(entered)
		<-release
		return []map[string]any{{"id": keys[0], "value": "old_remote"}}
	}
	reloadUserScene("test", conf)
	old := userFeatureService.FeatureSceneMap["test"][0]
	removed := make(chan struct{})
	c := cache.New(cache.WithMaximumSize(100), cache.WithRemovalListener(func(key cache.Key, _ cache.Value) {
		if key == "sentinel" {
			close(removed)
		}
	}))
	reloadTestCaches[c] = true
	old.featureDao = module.NewFeatureDaoWithCache(conf.FeatureDaoConf, c)
	c.Put("sentinel", "old")
	u := loadUserReloadRequest("test", "same")
	<-entered
	conf.FeatureDaoConf.CacheTime++
	reloadUserScene("test", conf)
	flushReloadTestInvalidations()
	select {
	case <-removed:
	case <-time.After(2 * time.Second):
		t.Fatal("old user cache was not cleared")
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case <-u.FeatureAsyncLoadCh():
	case <-time.After(2 * time.Second):
		t.Fatal("old user feature could not finish after invalidation")
	}
	if u.StringProperty("value") != "old_remote" {
		t.Fatal("old request lost the fetched feature")
	}
	if _, ok := c.GetIfPresent("same"); !ok {
		t.Fatal("late Put must remain allowed on the old cache")
	}
	if _, ok := userFeatureService.FeatureSceneMap["test"][0].featureCache().GetIfPresent("same"); ok {
		t.Fatal("old request wrote to the new cache")
	}
}

func TestUserFeatureReloadMatching(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadUserTestService(t)
	conf, _ := reloadTestConfig(t, "user")
	other := conf
	other.FeatureDaoConf.UserSelectFields = "id,value"
	reloadUserScene("test", conf, other, conf)
	old, _, _ := userFeatureService.getFeatures("test")
	first, second, third := old[0].featureCache(), old[1].featureCache(), old[2].featureCache()
	if first == second || first == third || second == third {
		t.Fatal("duplicate DAO configs must keep separate cache instances")
	}
	first.Put("same", map[string]any{"value": "first"})
	third.Put("same", map[string]any{"value": "third"})
	reloadUserScene("test", other, conf, conf)
	fresh, _, _ := userFeatureService.getFeatures("test")
	if fresh[0].featureCache() != second || fresh[1].featureCache() != first || fresh[2].featureCache() != third {
		t.Fatal("user cache matching relied on list indices or merged duplicates")
	}
	reloadUserScene("test", other, conf)
	if len(reloadTestInvalidations) != 1 || len(userFeatureService.FeatureAsyncLoadMap) != 2 {
		t.Fatal("removed user entry must lose its key and invalidate only its own cache")
	}
	flushReloadTestInvalidations()
	waitReloadCacheMiss(t, third, "same")
	if _, ok := first.GetIfPresent("same"); !ok {
		t.Fatal("removing a duplicate invalidated its retained sibling")
	}
}

func TestUserFeatureReloadOldListFlags(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadUserTestService(t)
	conf, _ := reloadTestConfig(t, "user")
	conf.FeatureDaoConf.FeatureAsyncLoad = true
	reloadUserScene("test", conf)
	old, _, _ := userFeatureService.getFeatures("test")
	conf.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: "copy", FeatureSource: "user:value"}}
	reloadUserScene("test", conf)
	if _, exists := userFeatureService.FeatureAsyncLoadMap[old[0]]; exists || !userFeatureService.isAsyncLoad(old[0]) {
		t.Fatal("an already acquired list lost its configured async mode after key cleanup")
	}
	current := userFeatureService.FeatureSceneMap["test"][0]
	userFeatureService.SetFeatures("alias", []*Feature{current})
	reloadUserScene("test")
	if _, exists := userFeatureService.FeatureAsyncLoadMap[current]; !exists {
		t.Fatal("cleanup removed a Feature still used by another current list")
	}
	if !userFeatureService.isAsyncLoad(current) {
		t.Fatal("retained Feature changed async mode")
	}
}

type userReloadFunctionDAO func(*module.User, *reccontext.RecommendContext)

func (f userReloadFunctionDAO) FeatureFetch(u *module.User, _ []*module.Item, ctx *reccontext.RecommendContext) {
	f(u, ctx)
}

func TestUserFeatureCallbackSyncOrder(t *testing.T) {
	reloadUserTestService(t)
	entered, release, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	first := &Feature{featureDao: userReloadFunctionDAO(func(*module.User, *reccontext.RecommendContext) {
		close(entered)
		<-release
	})}
	second := &Feature{featureDao: userReloadFunctionDAO(func(*module.User, *reccontext.RecommendContext) {
		close(finished)
	})}
	userFeatureService.SetFeatures("test", []*Feature{first, second})
	u := module.NewUser("same")
	userFeatureService.LoadUserFeaturesForCallback(u, userReloadContext("test"))
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("callback background sequence did not start")
	}
	select {
	case <-finished:
		t.Fatal("callback synchronous features no longer execute sequentially")
	default:
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("callback sequence did not finish")
	}
}

func TestUserFeatureReloadConcurrent(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadUserTestService(t)
	conf, _ := reloadTestConfig(t, "user")
	conf.FeatureDaoConf.FeatureAsyncLoad = true
	reloadUserScene("test", conf)
	c := userFeatureService.FeatureSceneMap["test"][0].featureCache()
	c.Put("same", map[string]any{"value": "cached"})
	var wg sync.WaitGroup
	for worker := 0; worker < 3; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				if worker == 0 {
					updated := conf
					updated.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: fmt.Sprint("copy", i), FeatureSource: "user:value"}}
					reloadUserScene("test", updated)
					continue
				}
				u := module.NewUser("same")
				u.AddProperty("id", "same")
				if worker == 1 {
					userFeatureService.LoadUserFeatures(u, userReloadContext("test"))
				} else {
					userFeatureService.LoadUserFeaturesForCallback(u, userReloadContext("test"))
				}
				select {
				case <-u.FeatureAsyncLoadCh():
				case <-time.After(2 * time.Second):
					t.Error("concurrent async load did not finish")
					return
				}
				if u.StringProperty("value") != "cached" {
					t.Error("concurrent user request lost cached data")
				}
			}
		}(worker)
	}
	wg.Wait()
	if len(userFeatureService.FeatureAsyncLoadMap) != 1 || userFeatureService.FeatureSceneMap["test"][0].featureCache() != c || len(reloadTestInvalidations) != 0 {
		t.Fatal("compatible concurrent reload accumulated keys or changed the cache")
	}
}

func TestUserFeatureReloadStages(t *testing.T) {
	reloadUserTestService(t)
	ctx := userReloadContext("test")
	ctx.ExperimentResult = &model.ExperimentResult{SceneName: "test"}
	ctx.ExperimentResult.GetExperimentParams().AddParams(map[string]any{
		"user_features.multistage.on":           true,
		"user_features.stage.base.scene.name":   "base",
		"user_features.stage.recall.scene.name": "recall",
		"user_features.stage.rank.scene.name":   "rank",
	})
	var mu sync.Mutex
	var loaded []string
	for _, scene := range []string{"base", "recall", "rank"} {
		f := &Feature{featureDao: userReloadFunctionDAO(func(*module.User, *reccontext.RecommendContext) {
			mu.Lock()
			loaded = append(loaded, scene)
			mu.Unlock()
		})}
		userFeatureService.SetFeatures(scene, []*Feature{f})
	}
	u := module.NewUser("same")
	userFeatureService.LoadUserFeatures(u, ctx)
	if fmt.Sprint(loaded) != "[base recall rank]" || u.FeatureAsyncLoadCount() != 0 {
		t.Fatal("stage ordering or load counter semantics changed")
	}
}

func TestUserFeatureSyncWait(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(fmt.Sprint(custom), func(t *testing.T) {
			reloadUserTestService(t)
			d := &blockingUserReloadDAO{make(chan struct{}), make(chan struct{}), make(chan struct{})}
			var release sync.Once
			t.Cleanup(func() { release.Do(func() { close(d.release) }) })
			userFeatureService.SetFeatures("test", []*Feature{{featureDao: d}})
			customDone := make(chan struct{})
			if custom {
				userFeatureService.AddFeatureFunc("test", func(_ *module.User, items []*module.Item, _ *reccontext.RecommendContext) []*module.Item {
					close(customDone)
					return items
				})
			}
			returned := make(chan struct{})
			go func() {
				defer close(returned)
				userFeatureService.LoadUserFeatures(module.NewUser("same"), userReloadContext("test"))
			}()
			<-d.entered
			select {
			case <-returned:
				t.Fatal("synchronous feature stopped blocking the caller")
			case <-time.After(10 * time.Millisecond):
			}
			release.Do(func() { close(d.release) })
			select {
			case <-returned:
			case <-time.After(2 * time.Second):
				t.Fatal("synchronous feature did not finish")
			}
			if custom {
				<-customDone
			}
		})
	}
}
