package recall

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	reccontext "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/featurestore"
	"github.com/goburrow/cache"
)

func i2iReloadConf() recconf.RecallConfig {
	conf := recconf.RecallConfig{Name: "i2i_reload", RecallType: "RealTimeU2IRecall", RecallCount: 5}
	conf.RealTimeUser2ItemDaoConf.I2ICacheSize = 100
	conf.RealTimeUser2ItemDaoConf.I2ICacheTime = 60
	conf.RealTimeUser2ItemDaoConf.Item2ItemFeatureViewName = "i2i"
	trigger := &conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf
	trigger.AdapterType = recconf.DataSource_Type_FeatureStore
	trigger.FeatureStoreName = "i2i_reload_fs"
	trigger.FeatureStoreViewName = "trigger"
	trigger.TriggerCount = 2
	trigger.Limit = 2
	trigger.WeightExpression = "1"
	trigger.EventWeight = "click:3"
	return conf
}

type behaviorShareConstructionView struct {
	domain.FeatureView
	calls atomic.Int32
}

func (v *behaviorShareConstructionView) GetName() string { return "trigger" }

func (v *behaviorShareConstructionView) GetBehaviorFeatures([]any, []any, []string) ([]map[string]any, error) {
	v.calls.Add(1)
	return []map[string]any{
		{"item_id": "seed", "event": "click", "timestamp": int64(100)},
		{"item_id": "other", "event": "click", "timestamp": int64(50)},
	}, nil
}

func TestRealtimeU2IBehaviorShareConstruction(t *testing.T) {
	if runI2IReloadSubprocess(t) {
		return
	}
	prepareI2IReloadSource(t)
	client, err := fs.GetFeatureStoreClient("i2i_reload_fs")
	if err != nil {
		t.Fatal(err)
	}
	view := &behaviorShareConstructionView{}
	client.GetProject().FeatureViewMap.Store("trigger", view)
	conf := i2iReloadConf()
	conf.RealTimeUser2ItemDaoConf.I2ICacheSize = 0
	conf.RealTimeUser2ItemDaoConf.Item2ItemFeatureViewName = ""
	conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.NoUsePlayTimeField = true
	conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.WeightExpression = "eventTime"
	conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.TriggerCount = 1
	user := module.NewUser("u")
	ctx := reccontext.NewRecommendContext()
	check := func(r *RealTimeU2IRecall, ctx *reccontext.RecommendContext, count int, score float64) {
		t.Helper()
		items := r.GetCandidateItems(user, ctx)
		if len(items) != count {
			t.Fatalf("召回数量=%d，期望%d", len(items), count)
		}
		for _, item := range items {
			if item.Id == "seed" && item.Score == score {
				return
			}
		}
		t.Fatalf("seed 分数错误，期望%v", score)
	}
	// 缺省开关保持原构造与直读路径。
	check(NewRealTimeU2IRecall(conf), ctx, 1, 300)
	conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.ShareBehaviorInRequest = true
	first := NewRealTimeU2IRecall(conf)
	check(first, ctx, 1, 300)
	conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.EventWeight = "click:2"
	conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.TriggerCount = 2
	encoded, err := json.Marshal(conf)
	if err != nil {
		t.Fatal(err)
	}
	var params map[string]interface{}
	if err := json.Unmarshal(encoded, &params); err != nil {
		t.Fatal(err)
	}
	cloned := first.CloneWithConfig(params).(*RealTimeU2IRecall)
	check(cloned, ctx, 2, 200)
	// 缓存继承专用构造也必须传入开关，但不共享加权后的 Item。
	check(newRealTimeU2IRecallWithCache(conf, nil, "share-test"), ctx, 2, 200)
	if view.calls.Load() != 2 {
		t.Fatalf("同参数构造/AB 路未共享：calls=%d", view.calls.Load())
	}
	conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.ShareBehaviorInRequest = false
	encoded, _ = json.Marshal(conf)
	if err := json.Unmarshal(encoded, &params); err != nil {
		t.Fatal(err)
	}
	disabled := first.CloneWithConfig(params).(*RealTimeU2IRecall)
	check(disabled, ctx, 2, 200)
	check(disabled, ctx, 2, 200)
	if view.calls.Load() != 4 {
		t.Fatal("关闭开关的 AB 路未独立读取")
	}
	check(first, reccontext.NewRecommendContext(), 1, 300)
	if view.calls.Load() != 5 {
		t.Fatal("不同请求复用了原始行为数据")
	}
}

func TestRealtimeU2ICacheReloadSignature(t *testing.T) {
	conf := i2iReloadConf()
	_, sign := i2iCacheConfig(conf, nil)
	if sign == "" {
		t.Fatal("FS I2I cache compatibility signature is missing")
	}
	var missing *RealTimeU2IRecall
	if inherited, got := i2iCacheConfig(conf, missing); inherited != nil || got != sign {
		t.Fatal("typed nil recall must not provide an inherited cache")
	}
	conf.CacheAdapter, conf.CacheConfig, conf.CachePrefix, conf.CacheTime = "localCache", "{}", "new_", 123
	if _, got := i2iCacheConfig(conf, nil); got != sign {
		t.Fatal("top-level recall cache policy changed the I2I signature")
	}
	changes := map[string]func(*recconf.RecallConfig){
		"ttl":    func(c *recconf.RecallConfig) { c.RealTimeUser2ItemDaoConf.I2ICacheTime++ },
		"size":   func(c *recconf.RecallConfig) { c.RealTimeUser2ItemDaoConf.I2ICacheSize++ },
		"weight": func(c *recconf.RecallConfig) { c.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.EventWeight = "click:4" },
		"view":   func(c *recconf.RecallConfig) { c.RealTimeUser2ItemDaoConf.Item2ItemFeatureViewName = "other" },
		"count":  func(c *recconf.RecallConfig) { c.RecallCount++ },
		"merge":  func(c *recconf.RecallConfig) { c.RealTimeUser2ItemDaoConf.MergeMode = "snake" },
		"name":   func(c *recconf.RecallConfig) { c.Name += "other" },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			changed := conf
			change(&changed)
			if _, got := i2iCacheConfig(changed, nil); got == sign {
				t.Fatal("I2I data or policy change was considered compatible")
			}
		})
	}
	conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.AdapterType = recconf.DaoConf_Adapter_Hologres
	if _, got := i2iCacheConfig(conf, nil); got != "" {
		t.Fatal("unmanaged adapter entered I2I inheritance")
	}
}

func runI2IReloadSubprocess(t *testing.T) bool {
	t.Helper()
	if os.Getenv("PAIREC_I2I_RELOAD_TEST") == "1" {
		return false
	}
	cmd := exec.Command(os.Args[0], "-test.run", "^"+regexp.QuoteMeta(t.Name())+"$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), "PAIREC_I2I_RELOAD_TEST=1")
	output, err := cmd.CombinedOutput()
	t.Logf("%s", output)
	if err != nil || !strings.Contains(string(output), "--- PASS: "+t.Name()) {
		t.Fatalf("isolated I2I test failed: %v", err)
	}
	return true
}

type i2iReloadView struct {
	domain.FeatureView
	calls atomic.Int32
	fetch func([]any) []map[string]any
}

func (*i2iReloadView) Close()                       {}
func (*i2iReloadView) GetFeatureEntityName() string { return "entity" }
func (*i2iReloadView) GetBehaviorFeatures([]any, []any, []string) ([]map[string]any, error) {
	return []map[string]any{
		{"item_id": "seed", "event": "click", "timestamp": time.Now().Unix(), "play_time": 1},
		{"item_id": "negative", "event": "click", "timestamp": time.Now().Unix(), "play_time": 1},
	}, nil
}
func (v *i2iReloadView) GetOnlineFeatures(keys []any, _ []string, _ map[string]string) ([]map[string]any, error) {
	v.calls.Add(1)
	return v.fetch(keys), nil
}

func prepareI2IReloadSource(t *testing.T) *i2iReloadView {
	t.Helper()
	var lists atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/instances"):
			fmt.Fprint(w, `{"Instances":[{"InstanceId":"test"}],"TotalCount":1}`)
		case strings.HasSuffix(r.URL.Path, "/projects"):
			if lists.Add(1) > 1 {
				fmt.Fprint(w, `{"Projects":[],"TotalCount":0}`)
				return
			}
			fmt.Fprint(w, `{"Projects":[{"ProjectId":"1","Name":"i2i_project","OfflineDatasourceId":"2","OnlineDatasourceId":"1","OfflineDatasourceType":"MaxCompute","OnlineDatasourceType":"FeatureDB"}],"TotalCount":1}`)
		case strings.Contains(r.URL.Path, "/datasources/"):
			fmt.Fprint(w, `{"Type":"FeatureDB","Name":"test","WorkspaceId":"1","Config":"{}"}`)
		default:
			fmt.Fprint(w, `{"Datasources":[],"FeatureEntities":[],"FeatureViews":[],"Models":[],"LabelTables":[],"TotalCount":0}`)
		}
	}))
	defer server.Close()
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	t.Setenv("GODEBUG", "x509usefallbackroots=1")
	x509.SetFallbackRoots(roots)
	fs.Load(&recconf.RecommendConfig{FeatureStoreConfs: map[string]recconf.FeatureStoreConfig{
		"i2i_reload_fs": {AccessId: "test", AccessKey: "test", RegionId: "test", Domain: strings.TrimPrefix(server.URL, "https://"), ProjectName: "i2i_project", TestMode: true},
	}})
	client, err := fs.GetFeatureStoreClient("i2i_reload_fs")
	if err != nil {
		t.Fatal(err)
	}
	field := reflect.ValueOf(client).Elem().FieldByName("client")
	sdk := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface().(*featurestore.FeatureStoreClient)
	sdk.Stop()
	project := client.GetProject()
	project.FeatureEntityMap["entity"] = domain.NewFeatureEntity(&api.FeatureEntity{FeatureEntityName: "entity", FeatureEntityJoinid: "item_id"})
	view := &i2iReloadView{fetch: func(keys []any) []map[string]any {
		var rows []map[string]any
		for _, key := range keys {
			if key == "seed" {
				rows = append(rows, map[string]any{"item_id": key, "similar_item_ids": "result:2"})
			}
		}
		return rows
	}}
	project.FeatureViewMap.Store("trigger", &i2iReloadView{})
	project.FeatureViewMap.Store("i2i", view)
	return view
}

func TestRealtimeU2ICacheReloadMain(t *testing.T) {
	if runI2IReloadSubprocess(t) {
		return
	}
	view := prepareI2IReloadSource(t)
	conf := i2iReloadConf()
	originalAfter := i2iCacheAfterFunc
	var callbacks []func()
	i2iCacheAfterFunc = func(delay time.Duration, run func()) *time.Timer {
		if delay != 30*time.Second {
			t.Error("unexpected invalidation delay")
		}
		callbacks = append(callbacks, run)
		return nil
	}
	caches := make(map[cache.Cache]bool)
	t.Cleanup(func() {
		for _, run := range callbacks {
			run()
		}
		for c := range caches {
			_ = c.Close()
		}
		i2iCacheAfterFunc = originalAfter
	})
	load := func() *RealTimeU2IRecall {
		Load(&recconf.RecommendConfig{RecallConfs: []recconf.RecallConfig{conf}})
		r, _ := GetRecall(conf.Name)
		fresh := r.(*RealTimeU2IRecall)
		if c := fresh.i2iCache(); c != nil {
			caches[c] = true
		}
		return fresh
	}
	old := load()
	if load() != old {
		t.Fatal("unchanged recall configuration rebuilt the instance")
	}
	items := old.GetCandidateItems(module.NewUser("u"), reccontext.NewRecommendContext())
	if len(items) != 1 || items[0].Id != "result" || items[0].Score != 6 || view.calls.Load() != 1 {
		t.Fatal("original I2I DAO result or weighted score changed")
	}
	c := old.i2iCache()
	if negative, ok := c.GetIfPresent("negative"); !ok || negative != "" {
		t.Fatal("original negative cache was not populated")
	}
	for i := 0; i < 5; i++ {
		conf.CacheTime = 10 + i
		fresh := load()
		if fresh == old || fresh.i2iCache() != c {
			t.Fatal("top-level TTL update did not inherit I2I data")
		}
		items = fresh.GetCandidateItems(module.NewUser("u"), reccontext.NewRecommendContext())
		if len(items) != 1 || items[0].Score != 6 || view.calls.Load() != 1 || len(callbacks) != 0 {
			t.Fatal("I2I inheritance lost weighted data, negative cache, or scheduled cleanup")
		}
	}
	for _, change := range []func(*recconf.RecallConfig){
		func(c *recconf.RecallConfig) {
			c.CacheAdapter = "localCache"
			c.CacheConfig = `{"cleanupInterval":0}`
		},
		func(c *recconf.RecallConfig) { c.CacheConfig = `{"defaultExpiration":120,"cleanupInterval":0}` },
		func(c *recconf.RecallConfig) { c.CachePrefix = "other_" },
		func(c *recconf.RecallConfig) { c.CacheTime++ },
	} {
		previous := load()
		change(&conf)
		fresh := load()
		if fresh.i2iCache() != c || fresh.cache == previous.cache || fresh.cache == nil {
			t.Fatal("top-level policy must rebuild only the top-level cache")
		}
		items := fresh.GetCandidateItems(module.NewUser("u"), reccontext.NewRecommendContext())
		if len(items) != 1 || items[0].Score != 6 || view.calls.Load() != 1 {
			t.Fatal("top-level cache policy change lost inherited I2I data")
		}
	}
	conf.RealTimeUser2ItemDaoConf.I2ICacheTime++
	fresh := load()
	if fresh.i2iCache() == c || len(callbacks) != 1 {
		t.Fatal("I2I TTL update did not create a separate cache")
	}
	if _, ok := c.GetIfPresent("seed"); !ok {
		t.Fatal("old I2I data was cleared before the delay")
	}
	sign := recallSigns[conf.Name]
	conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.WeightExpression = "invalid("
	func() {
		defer func() {
			if recover() == nil {
				t.Error("invalid DAO expression must not publish a recall")
			}
		}()
		load()
	}()
	current, _ := GetRecall(conf.Name)
	if current != fresh || recallSigns[conf.Name] != sign || len(callbacks) != 1 {
		t.Fatal("failed recall construction changed the registry or old cache")
	}
	callbacks[0]()
	callbacks = nil
	if _, ok := c.GetIfPresent("seed"); ok {
		t.Fatal("old I2I data was not invalidated")
	}
	c.Put("late", "still allowed")
	fresh.i2iCache().Put("seed", "result:6")
	conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.WeightExpression = "1"
	conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.EventWeight = "click:4"
	weighted := load()
	items = weighted.GetCandidateItems(module.NewUser("u"), reccontext.NewRecommendContext())
	if weighted.i2iCache() == fresh.i2iCache() || len(items) != 1 || items[0].Score != 8 || view.calls.Load() != 2 || len(callbacks) != 1 {
		t.Fatal("changed trigger weight reused previously weighted I2I data")
	}
	callbacks[0]()
	callbacks = nil
	conf.RealTimeUser2ItemDaoConf.I2ICacheSize = 0
	conf.CacheAdapter = ""
	disabled := load()
	if disabled.i2iCache() != nil || len(callbacks) != 1 {
		t.Fatal("disabling I2I cache did not retire the previous cache")
	}
	callbacks[0]()
	callbacks = nil
	if items := disabled.GetCandidateItems(module.NewUser("u"), reccontext.NewRecommendContext()); len(items) != 1 || items[0].Score != 8 || view.calls.Load() != 3 {
		t.Fatal("disabled I2I cache did not retain original backsource behavior")
	}
	conf.RealTimeUser2ItemDaoConf.I2ICacheSize = 100
	fresh = load()
	fresh.i2iCache().Put("seed", "result:8")
	conf.RecallType, conf.CacheAdapter = "MockRecall", ""
	Load(&recconf.RecommendConfig{RecallConfs: []recconf.RecallConfig{conf}})
	if len(callbacks) != 1 {
		t.Fatal("replacing FS U2I with another recall did not schedule old I2I cleanup")
	}
	callbacks[0]()
	callbacks = nil
	if _, ok := fresh.i2iCache().GetIfPresent("seed"); ok {
		t.Fatal("retired I2I data remained after recall type replacement")
	}
}

func TestRealtimeU2ICacheReloadIndependentConstructors(t *testing.T) {
	if runI2IReloadSubprocess(t) {
		return
	}
	prepareI2IReloadSource(t)
	conf := i2iReloadConf()
	first, second := NewRealTimeU2IRecall(conf), NewRealTimeU2IRecall(conf)
	t.Cleanup(func() { _ = first.i2iCache().Close() })
	t.Cleanup(func() { _ = second.i2iCache().Close() })
	if first.i2iCacheSign != "" || first.i2iCache() == second.i2iCache() {
		t.Fatal("ordinary constructors entered managed cache inheritance")
	}
	encoded, _ := json.Marshal(conf)
	var params map[string]any
	_ = json.Unmarshal(encoded, &params)
	cloned := first.CloneWithConfig(params).(*RealTimeU2IRecall)
	if cloned == first || cloned.i2iCacheSign != "" || cloned.i2iCache() == first.i2iCache() {
		t.Fatal("AB clones unexpectedly shared the main I2I cache")
	}
	t.Cleanup(func() { _ = cloned.i2iCache().Close() })
	first.i2iCache().Put("seed", "ordinary:1")
	RegisterRecall(conf.Name, first)
	Load(&recconf.RecommendConfig{RecallConfs: []recconf.RecallConfig{conf}})
	managed, _ := GetRecall(conf.Name)
	t.Cleanup(func() { _ = managed.(*RealTimeU2IRecall).i2iCache().Close() })
	if managed.(*RealTimeU2IRecall).i2iCache() == first.i2iCache() {
		t.Fatal("load took over an ordinary constructor's cache")
	}
	if value, ok := first.i2iCache().GetIfPresent("seed"); !ok || value != "ordinary:1" {
		t.Fatal("load invalidated an unmanaged ordinary cache")
	}
}

func TestRealtimeU2ICacheReloadConcurrent(t *testing.T) {
	if runI2IReloadSubprocess(t) {
		return
	}
	prepareI2IReloadSource(t)
	conf := i2iReloadConf()
	Load(&recconf.RecommendConfig{RecallConfs: []recconf.RecallConfig{conf}})
	r, _ := GetRecall(conf.Name)
	c := r.(*RealTimeU2IRecall).i2iCache()
	t.Cleanup(func() { _ = c.Close() })
	c.Put("seed", "result:6")
	c.Put("negative", "")
	var wg sync.WaitGroup
	for worker := 0; worker < 3; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				if worker == 0 {
					updated := conf
					updated.CacheTime = i + 1
					Load(&recconf.RecommendConfig{RecallConfs: []recconf.RecallConfig{updated}})
				} else {
					current, _ := GetRecall(conf.Name)
					items := current.GetCandidateItems(module.NewUser("u"), reccontext.NewRecommendContext())
					if len(items) != 1 || items[0].Score != 6 {
						t.Error("concurrent recall lost I2I data")
					}
				}
			}
		}(worker)
	}
	wg.Wait()
	current, _ := GetRecall(conf.Name)
	if current.(*RealTimeU2IRecall).i2iCache() != c {
		t.Fatal("concurrent compatible reload changed cache instances")
	}
}

func TestRealtimeU2ICacheReloadLatePut(t *testing.T) {
	if runI2IReloadSubprocess(t) {
		return
	}
	view := prepareI2IReloadSource(t)
	conf := i2iReloadConf()
	originalAfter := i2iCacheAfterFunc
	var callbacks []func()
	i2iCacheAfterFunc = func(delay time.Duration, run func()) *time.Timer {
		if delay != 30*time.Second {
			t.Error("unexpected retirement delay")
		}
		callbacks = append(callbacks, run)
		return nil
	}
	caches := make(map[cache.Cache]bool)
	t.Cleanup(func() {
		for _, run := range callbacks {
			run()
		}
		for c := range caches {
			_ = c.Close()
		}
		i2iCacheAfterFunc = originalAfter
	})
	Load(&recconf.RecommendConfig{RecallConfs: []recconf.RecallConfig{conf}})
	r, _ := GetRecall(conf.Name)
	old := r.(*RealTimeU2IRecall)
	caches[old.i2iCache()] = true
	removed := make(chan struct{})
	c := cache.New(cache.WithMaximumSize(100), cache.WithExpireAfterWrite(time.Minute),
		cache.WithRemovalListener(func(key cache.Key, _ cache.Value) {
			if key == "sentinel" {
				close(removed)
			}
		}))
	caches[c] = true
	old.user2ItemDao = module.NewRealTimeUser2ItemDaoWithCache(conf, c)
	c.Put("sentinel", "marker:1")
	entered, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	view.fetch = func([]any) []map[string]any {
		close(entered)
		<-release
		return []map[string]any{{"item_id": "seed", "similar_item_ids": "old_result:2"}}
	}
	done := make(chan []*module.Item, 1)
	go func() {
		done <- old.GetCandidateItems(module.NewUser("u"), reccontext.NewRecommendContext())
	}()
	<-entered
	conf.RealTimeUser2ItemDaoConf.I2ICacheTime++
	Load(&recconf.RecommendConfig{RecallConfs: []recconf.RecallConfig{conf}})
	current, _ := GetRecall(conf.Name)
	fresh := current.(*RealTimeU2IRecall)
	caches[fresh.i2iCache()] = true
	if len(callbacks) != 1 {
		t.Fatal("old I2I cache cleanup was not scheduled")
	}
	callbacks[0]()
	callbacks = nil
	select {
	case <-removed:
	case <-time.After(2 * time.Second):
		t.Fatal("old I2I attributes were not physically removed")
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case items := <-done:
		if len(items) != 1 || items[0].Id != "old_result" || items[0].Score != 6 {
			t.Fatal("old U2I request could not finish with its original weighted score")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("old U2I request did not finish")
	}
	if value, ok := c.GetIfPresent("seed"); !ok || value != "old_result:6" {
		t.Fatal("old I2I cache must still accept late Put")
	}
	if _, ok := fresh.i2iCache().GetIfPresent("seed"); ok {
		t.Fatal("old request repopulated the new I2I cache")
	}
}
