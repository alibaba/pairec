package filter

import (
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
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

func itemStateReloadConfig() recconf.FilterConfig {
	return recconf.FilterConfig{Name: "state", FilterType: "ItemStateFilter", ItemStateCacheSize: 100, ItemStateCacheTime: 60,
		ItemStateDaoConf: recconf.ItemStateDaoConfig{DaoConfig: recconf.DaoConfig{AdapterType: recconf.DataSource_Type_FeatureStore}}}
}

func TestItemStateCacheReloadSignature(t *testing.T) {
	conf := itemStateReloadConfig()
	_, sign := itemStateCacheConfig(conf, nil)
	if sign == "" {
		t.Fatal("private state cache must have a compatibility signature")
	}
	var missing *ItemStateFilter
	if inherited, got := itemStateCacheConfig(conf, missing); inherited != nil || got != sign {
		t.Fatal("typed nil old filter must be treated as absent")
	}
	conf.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: "copy"}}
	conf.FilterParams = []recconf.FilterParamConfig{{Name: "value", Operator: "equal", Value: "allowed"}}
	_, changed := itemStateCacheConfig(conf, nil)
	if changed != sign {
		t.Fatal("conditions and conversions must not invalidate attribute-cache compatibility")
	}
	changes := map[string]func(*recconf.FilterConfig){
		"ttl":      func(c *recconf.FilterConfig) { c.ItemStateCacheTime++ },
		"size":     func(c *recconf.FilterConfig) { c.ItemStateCacheSize++ },
		"name":     func(c *recconf.FilterConfig) { c.Name += "other" },
		"fields":   func(c *recconf.FilterConfig) { c.ItemStateDaoConf.SelectFields = "other" },
		"where":    func(c *recconf.FilterConfig) { c.ItemStateDaoConf.WhereClause = "state = 1" },
		"defaults": func(c *recconf.FilterConfig) { c.ItemStateDaoConf.DefaultFieldValues = map[string]any{"state": 1} },
		"key":      func(c *recconf.FilterConfig) { c.GenerateUserDataExpr = "item.item_id" },
		"adapter":  func(c *recconf.FilterConfig) { c.ItemStateDaoConf.AdapterType = recconf.DaoConf_Adapter_Hologres },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			changedConf := conf
			change(&changedConf)
			_, got := itemStateCacheConfig(changedConf, nil)
			if got == sign {
				t.Fatal("cache policy or source change was considered compatible")
			}
		})
	}
	conf.ItemStateDaoConf.FeatureViewCacheName = "scan"
	if _, got := itemStateCacheConfig(conf, nil); got != "" {
		t.Fatal("scan cache entered private cache inheritance")
	}
	conf.ItemStateDaoConf.FeatureViewCacheName = ""
	conf.ItemStateDaoConf.AdapterType = recconf.DaoConf_Adapter_TableStore
	if _, got := itemStateCacheConfig(conf, nil); got != "" {
		t.Fatal("unrelated adapter entered private cache inheritance")
	}
}

func TestItemStateCacheReloadInvalidation(t *testing.T) {
	original := itemStateCacheAfterFunc
	defer func() { itemStateCacheAfterFunc = original }()
	var callbacks []func()
	itemStateCacheAfterFunc = func(delay time.Duration, run func()) *time.Timer {
		if delay != 30*time.Second {
			t.Error("unexpected invalidation delay")
		}
		callbacks = append(callbacks, run)
		return nil
	}
	removed := make(chan struct{})
	old := cache.New(cache.WithMaximumSize(100), cache.WithRemovalListener(func(key cache.Key, _ cache.Value) {
		if key == "old" {
			close(removed)
		}
	}))
	fresh := cache.New(cache.WithMaximumSize(100))
	defer old.Close()
	defer fresh.Close()
	old.Put("old", "value")
	fresh.Put("fresh", "value")
	invalidateItemStateCache(old, old)
	invalidateItemStateCache(nil, fresh)
	if len(callbacks) != 0 {
		t.Fatal("inherited cache was scheduled for invalidation")
	}
	invalidateItemStateCache(old, fresh)
	if len(callbacks) != 1 {
		t.Fatal("old cache invalidation was not scheduled")
	}
	if _, ok := old.GetIfPresent("old"); !ok {
		t.Fatal("old cache was cleared before the delay")
	}
	callbacks[0]()
	if _, ok := old.GetIfPresent("old"); ok {
		t.Fatal("old cache data was not invalidated")
	}
	if _, ok := fresh.GetIfPresent("fresh"); !ok {
		t.Fatal("new cache data was cleared")
	}
	select {
	case <-removed:
	case <-time.After(2 * time.Second):
		t.Fatal("invalidated data was not physically removed")
	}
	old.Put("late", "still allowed")
	if value, ok := old.GetIfPresent("late"); !ok || value != "still allowed" {
		t.Fatal("old cache was closed instead of invalidated")
	}
}

type itemStateReloadView struct {
	domain.FeatureView
	calls atomic.Int32
}

func (*itemStateReloadView) GetFeatureEntityName() string { return "entity" }
func (*itemStateReloadView) Close()                       {}
func (v *itemStateReloadView) GetOnlineFeatures(keys []any, _ []string, _ map[string]string) ([]map[string]any, error) {
	v.calls.Add(1)
	var rows []map[string]any
	for _, key := range keys {
		if key != "missing" {
			rows = append(rows, map[string]any{"id": key, "value": "raw"})
		}
	}
	return rows, nil
}

func prepareItemStateReloadSource(t *testing.T) *itemStateReloadView {
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
			fmt.Fprint(w, `{"Projects":[{"ProjectId":"1","Name":"state_project","OfflineDatasourceId":"2","OnlineDatasourceId":"1","OfflineDatasourceType":"MaxCompute","OnlineDatasourceType":"FeatureDB"}],"TotalCount":1}`)
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
		"state_fs": {AccessId: "test", AccessKey: "test", RegionId: "test", Domain: strings.TrimPrefix(server.URL, "https://"), ProjectName: "state_project", TestMode: true},
	}})
	client, err := fs.GetFeatureStoreClient("state_fs")
	if err != nil {
		t.Fatal(err)
	}
	field := reflect.ValueOf(client).Elem().FieldByName("client")
	sdk := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface().(*featurestore.FeatureStoreClient)
	sdk.Stop()
	project := client.GetProject()
	project.FeatureEntityMap["entity"] = domain.NewFeatureEntity(&api.FeatureEntity{FeatureEntityName: "entity", FeatureEntityJoinid: "id"})
	view := &itemStateReloadView{}
	project.FeatureViewMap.Store("view", view)
	return view
}

func TestItemStateFilterReloadRegistry(t *testing.T) {
	if os.Getenv("PAIREC_STATE_RELOAD_TEST") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestItemStateFilterReloadRegistry$", "-test.v")
		cmd.Env = append(os.Environ(), "PAIREC_STATE_RELOAD_TEST=1")
		output, err := cmd.CombinedOutput()
		t.Logf("%s", output)
		if err != nil || !strings.Contains(string(output), "--- PASS: TestItemStateFilterReloadRegistry") {
			t.Fatalf("isolated state test failed: %v", err)
		}
		return
	}
	view := prepareItemStateReloadSource(t)
	original := itemStateCacheAfterFunc
	var callbacks []func()
	itemStateCacheAfterFunc = func(delay time.Duration, run func()) *time.Timer {
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
		itemStateCacheAfterFunc = original
	})
	conf := itemStateReloadConfig()
	conf.ItemStateDaoConf.FeatureStoreName = "state_fs"
	conf.ItemStateDaoConf.FeatureStoreViewName = "view"
	conf.ItemStateDaoConf.ItemFieldName = "id"
	conf.ItemStateDaoConf.DefaultFieldValues = map[string]any{"value": "default"}
	load := func() *ItemStateFilter {
		RegisterFilterWithConfig(&recconf.RecommendConfig{FilterConfs: []recconf.FilterConfig{conf}})
		f, _ := GetFilter(conf.Name)
		s := f.(*ItemStateFilter)
		if c := s.itemStateCache(); c != nil {
			caches[c] = true
		}
		return s
	}
	check := func(f *ItemStateFilter, key string) int {
		item := module.NewItem(key)
		item.AddProperty("local", "yes")
		data := &FilterData{User: module.NewUser("u"), Data: []*module.Item{item}, Context: reccontext.NewRecommendContext()}
		if err := f.Filter(data); err != nil {
			t.Fatal(err)
		}
		return len(data.Data.([]*module.Item))
	}
	var missing *ItemStateFilter
	RegisterFilter(conf.Name, missing)
	old := load()
	if load() != old {
		t.Fatal("unchanged config rebuilt the filter")
	}
	if check(old, "same") != 1 || view.calls.Load() != 1 {
		t.Fatal("state filter did not successfully fetch properties")
	}
	conf.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: "copy", FeatureStore: "item", FeatureSource: "item:value"}}
	conf.FilterParams = []recconf.FilterParamConfig{{Name: "copy", Domain: "item", Operator: "equal", Type: "string", Value: "raw"}}
	fresh := load()
	if fresh == old || fresh.itemStateCache() != old.itemStateCache() || len(callbacks) != 0 {
		t.Fatal("conversion reload did not inherit attributes")
	}
	if check(fresh, "same") != 1 || view.calls.Load() != 1 {
		t.Fatal("new conditions/conversion did not use inherited data")
	}
	conf.Features = nil
	conf.FilterParams = []recconf.FilterParamConfig{{Name: "local", Domain: "item", Operator: "equal", Type: "string", Value: "yes"}}
	without := load()
	if without.itemStateCache() != old.itemStateCache() {
		t.Fatal("conversion removal replaced the cache")
	}
	for _, key := range []string{"same", "new", "missing"} {
		if check(without, key) != 0 {
			t.Fatal("nil conversion became an empty closure")
		}
	}
	conf.ItemStateCacheTime++
	replaced := load()
	if replaced.itemStateCache() == old.itemStateCache() || len(callbacks) != 1 {
		t.Fatal("policy change did not schedule old data cleanup")
	}
	sign := filterSigns[conf.Name]
	conf.GenerateUserDataExpr = "invalid("
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected invalid generation expression")
			}
		}()
		load()
	}()
	current, _ := GetFilter(conf.Name)
	if current != replaced || filterSigns[conf.Name] != sign || len(callbacks) != 1 {
		t.Fatal("failed construction changed registry or scheduled cleanup")
	}
	oldContext := reccontext.NewRecommendContext()
	oldContext.Param = stateReloadParams{"scene": "state_scene"}
	Load(&recconf.RecommendConfig{FilterNames: map[string][]string{"state_scene": {conf.Name}}})
	currentData := &FilterData{User: module.NewUser("u"), Data: []*module.Item{module.NewItem("same")}, Context: oldContext}
	filterService.Filter(currentData, "")
	conf.GenerateUserDataExpr = ""
	first, second := NewItemStateFilter(conf), NewItemStateFilter(conf)
	caches[first.itemStateCache()], caches[second.itemStateCache()] = true, true
	if first.cacheSign != "" || first.itemStateCache() == second.itemStateCache() {
		t.Fatal("ordinary constructors started sharing reload caches")
	}
	name := conf.Name
	var wg sync.WaitGroup
	for worker := 0; worker < 3; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				if worker == 0 {
					conf.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: fmt.Sprint("copy", i), FeatureStore: "item", FeatureSource: "item:value"}}
					load()
				} else {
					f, err := GetFilter(name)
					if err != nil {
						t.Error(err)
						return
					}
					check(f.(*ItemStateFilter), "same")
				}
			}
		}(worker)
	}
	wg.Wait()
	if len(callbacks) != 1 {
		t.Fatal("compatible concurrent replacements scheduled extra invalidation")
	}
	callbacks[0]()
	callbacks = nil
	if _, ok := old.itemStateCache().GetIfPresent("same"); ok {
		t.Fatal("old attribute data was not invalidated")
	}
}

type stateReloadParams map[string]any

func (p stateReloadParams) GetParameter(name string) any { return p[name] }
