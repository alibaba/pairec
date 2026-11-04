package feature

import (
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"io"
	"net"
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
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/featurestore"
	"github.com/goburrow/cache"
)

const reloadFSName = "featureconfs_reload_test_fs"

var reloadFSOnce sync.Once
var reloadFSProject *domain.Project

func runReloadTestInSubprocess(t *testing.T) bool {
	t.Helper()
	if os.Getenv("PAIREC_FEATURE_RELOAD_SUBPROCESS") == "1" {
		return false
	}
	cmd := exec.Command(os.Args[0], "-test.run", "^"+regexp.QuoteMeta(t.Name())+"$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), "PAIREC_FEATURE_RELOAD_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()
	t.Logf("%s", output)
	if err != nil || !strings.Contains(string(output), "--- PASS: "+t.Name()) {
		t.Fatalf("isolated cache test failed: %v", err)
	}
	return true
}

type reloadTestInvalidation struct {
	delay time.Duration
	run   func()
}

var reloadTestInvalidations []reloadTestInvalidation
var reloadTestCaches map[cache.Cache]bool

func flushReloadTestInvalidations() {
	for _, pending := range reloadTestInvalidations {
		pending.run()
	}
	reloadTestInvalidations = nil
}

func prepareReloadFS(t *testing.T) *domain.Project {
	t.Helper()
	reloadFSOnce.Do(func() {
		var projectLists atomic.Int32
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch {
			case strings.HasSuffix(r.URL.Path, "/instances"):
				fmt.Fprint(w, `{"Instances":[{"InstanceId":"test"}],"TotalCount":1}`)
			case strings.HasSuffix(r.URL.Path, "/projects"):
				if projectLists.Add(1) > 1 {
					fmt.Fprint(w, `{"Projects":[],"TotalCount":0}`)
					return
				}
				fmt.Fprint(w, `{"Projects":[{"ProjectId":"1","Name":"reload_project","OfflineDatasourceId":"2","OnlineDatasourceId":"1","OfflineDatasourceType":"MaxCompute","OnlineDatasourceType":"FeatureDB"}],"TotalCount":1}`)
			case strings.Contains(r.URL.Path, "/datasources/"):
				fmt.Fprint(w, `{"Type":"FeatureDB","Name":"test","WorkspaceId":"1","Config":"{}"}`)
			case strings.HasSuffix(r.URL.Path, "/datasources"):
				fmt.Fprint(w, `{"Datasources":[],"TotalCount":0}`)
			case strings.HasSuffix(r.URL.Path, "/featureentities"):
				fmt.Fprint(w, `{"FeatureEntities":[],"TotalCount":0}`)
			default:
				fmt.Fprint(w, `{"FeatureViews":[],"Models":[],"LabelTables":[],"TotalCount":0}`)
			}
		}))
		defer server.Close()
		roots := x509.NewCertPool()
		roots.AddCert(server.Certificate())
		t.Setenv("GODEBUG", "x509usefallbackroots=1")
		x509.SetFallbackRoots(roots)
		fs.Load(&recconf.RecommendConfig{FeatureStoreConfs: map[string]recconf.FeatureStoreConfig{
			reloadFSName: {AccessId: "test", AccessKey: "test", RegionId: "test", Domain: strings.TrimPrefix(server.URL, "https://"), ProjectName: "reload_project", TestMode: true},
		}})
		client, err := fs.GetFeatureStoreClient(reloadFSName)
		if err != nil {
			t.Fatal(err)
		}
		field := reflect.ValueOf(client).Elem().FieldByName("client")
		sdk := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface().(*featurestore.FeatureStoreClient)
		sdk.Stop()
		reloadFSProject = client.GetProject()
		reloadFSProject.FeatureEntityMap["entity"] = domain.NewFeatureEntity(&api.FeatureEntity{FeatureEntityJoinid: "id", FeatureEntityName: "entity"})
	})
	return reloadFSProject
}

type reloadTestView struct {
	domain.FeatureView
	calls atomic.Int32
	fetch func([]any) []map[string]any
}

func (*reloadTestView) GetFeatureEntityName() string { return "entity" }
func (*reloadTestView) Close()                       {}
func (v *reloadTestView) GetOnlineFeatures(keys []any, _ []string, _ map[string]string) ([]map[string]any, error) {
	v.calls.Add(1)
	return v.fetch(keys), nil
}
func (v *reloadTestView) GetOnlineFeaturesWithOptions(keys []any, fields []string, alias map[string]string, _ domain.FeatureViewOptions) ([]map[string]any, error) {
	return v.GetOnlineFeatures(keys, fields, alias)
}

func reloadTestService(t *testing.T) {
	t.Helper()
	old := featureService
	oldAfterFunc := featureCacheAfterFunc
	featureService = NewFeatureService()
	reloadTestCaches = make(map[cache.Cache]bool)
	reloadTestInvalidations = nil
	featureCacheAfterFunc = func(delay time.Duration, run func()) *time.Timer {
		reloadTestInvalidations = append(reloadTestInvalidations, reloadTestInvalidation{delay, run})
		return nil
	}
	t.Cleanup(func() {
		flushReloadTestInvalidations()
		for c := range reloadTestCaches {
			_ = c.Close()
		}
		reloadTestCaches = nil
		featureCacheAfterFunc = oldAfterFunc
		featureService = old
	})
}

func reloadTestConfig(t *testing.T, store string) (recconf.FeatureLoadConfig, *reloadTestView) {
	t.Helper()
	project := prepareReloadFS(t)
	name := t.Name() + "_" + store
	view := &reloadTestView{fetch: func(keys []any) []map[string]any {
		rows := make([]map[string]any, 0, len(keys))
		for _, key := range keys {
			rows = append(rows, map[string]any{"id": key, "value": store})
		}
		return rows
	}}
	project.FeatureViewMap.Store(name, view)
	t.Cleanup(func() { project.FeatureViewMap.Delete(name) })
	conf := recconf.FeatureLoadConfig{FeatureDaoConf: recconf.FeatureDaoConfig{
		DaoConfig:    recconf.DaoConfig{AdapterType: recconf.DataSource_Type_FeatureStore, FeatureStoreName: reloadFSName, FeatureStoreViewName: name},
		FeatureStore: store,
		FeatureKey:   store + ":id", CacheSize: 100, CacheTime: 60,
	}}
	return conf, view
}

func reloadScene(confs ...recconf.FeatureLoadConfig) {
	LoadFeatureConfig(&recconf.RecommendConfig{FeatureConfs: map[string]recconf.SceneFeatureConfig{
		"test": {FeatureLoadConfs: confs},
	}})
	for _, f := range featureService.FeatureSceneMap["test"] {
		if c := f.featureCache(); c != nil {
			reloadTestCaches[c] = true
		}
	}
}

func TestFeatureConfsUserCacheReload(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadTestService(t)
	conf, view := reloadTestConfig(t, "user")
	reloadScene(conf)
	old := featureService.FeatureSceneMap["test"][0]
	ctx := reccontext.NewRecommendContext()
	fetch := func(f *Feature) *module.User {
		u := module.NewUser("same")
		u.AddProperty("id", "same")
		f.LoadFeatures(u, nil, ctx)
		return u
	}
	if got := fetch(old).StringProperty("value"); got != "user" {
		t.Fatalf("backsource value=%s", got)
	}
	conf.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: "copy", FeatureStore: "user", FeatureSource: "user:value"}}
	reloadScene(conf)
	fresh := featureService.FeatureSceneMap["test"][0]
	if fresh == old || fresh.featureCache() != old.featureCache() {
		t.Fatal("conversion-only reload did not inherit")
	}
	if got := fetch(fresh).StringProperty("copy"); got != "user" {
		t.Fatalf("new conversion value=%s", got)
	}
	if view.calls.Load() != 1 {
		t.Fatal("cache hit fetched remotely")
	}
}

func TestFeatureConfsMixedCacheReload(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadTestService(t)
	userConf, userView := reloadTestConfig(t, "user")
	itemConf, itemView := reloadTestConfig(t, "item")
	reloadScene(userConf, itemConf, itemConf)
	old := featureService.FeatureSceneMap["test"]
	userCache, itemCache, duplicateCache := old[0].featureCache(), old[1].featureCache(), old[2].featureCache()
	if userCache == itemCache || itemCache == duplicateCache {
		t.Fatal("distinct entries shared a cache")
	}
	userCache.Put("same", map[string]any{"value": "user_cached"})
	itemCache.Put("same", map[string]any{"value": "item_cached"})
	duplicateCache.Put("same", map[string]any{"value": "duplicate_cached"})
	itemConf.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: "copy", FeatureStore: "item", FeatureSource: "item:value"}}
	reloadScene(itemConf, userConf, itemConf)
	fresh := featureService.FeatureSceneMap["test"]
	if fresh[0].featureCache() != itemCache || fresh[1].featureCache() != userCache || fresh[2].featureCache() != duplicateCache {
		t.Fatal("reorder did not match caches one-to-one")
	}
	u := module.NewUser("same")
	u.AddProperty("id", "same")
	item := module.NewItem("same")
	ctx := reccontext.NewRecommendContext()
	fresh[0].LoadFeatures(u, []*module.Item{item}, ctx)
	fresh[1].LoadFeatures(u, []*module.Item{item}, ctx)
	if item.StringProperty("copy") != "item_cached" || u.StringProperty("value") != "user_cached" {
		t.Fatal("user/item cache values mixed")
	}
	if userView.calls.Load() != 0 || itemView.calls.Load() != 0 {
		t.Fatal("cached data fetched remotely")
	}
	reloadScene(userConf, itemConf)
	if _, ok := duplicateCache.GetIfPresent("same"); !ok {
		t.Fatal("removed cache was cleared before the delay")
	}
	if len(reloadTestInvalidations) != 1 || reloadTestInvalidations[0].delay != 30*time.Second {
		t.Fatal("removed cache was not scheduled exactly once after 30 seconds")
	}
	flushReloadTestInvalidations()
	waitReloadCacheMiss(t, duplicateCache, "same")
	if _, ok := itemCache.GetIfPresent("same"); !ok {
		t.Fatal("invalidation cleared an inherited cache")
	}
	for i := 0; i < 10; i++ {
		itemConf.Features[0].FeatureName = fmt.Sprint("copy", i)
		reloadScene(userConf, itemConf)
		if featureService.FeatureSceneMap["test"][1].featureCache() != itemCache {
			t.Fatal("compatible reload allocated a new cache")
		}
	}
}

func TestFeatureConfsIncompatibleCacheReload(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	changes := map[string]func(*recconf.FeatureLoadConfig){
		"ttl":     func(c *recconf.FeatureLoadConfig) { c.FeatureDaoConf.CacheTime++ },
		"size":    func(c *recconf.FeatureLoadConfig) { c.FeatureDaoConf.CacheSize++ },
		"fields":  func(c *recconf.FeatureLoadConfig) { c.FeatureDaoConf.ItemSelectFields = "value" },
		"view":    func(c *recconf.FeatureLoadConfig) { c.FeatureDaoConf.FeatureStoreViewName += "_new" },
		"key":     func(c *recconf.FeatureLoadConfig) { c.FeatureDaoConf.FeatureKey = "item:other" },
		"disable": func(c *recconf.FeatureLoadConfig) { c.FeatureDaoConf.CacheSize = 0 },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			reloadTestService(t)
			conf, _ := reloadTestConfig(t, "item")
			reloadScene(conf)
			old := featureService.FeatureSceneMap["test"][0]
			c := old.featureCache()
			c.Put("same", map[string]any{"value": "old"})
			change(&conf)
			reloadScene(conf)
			fresh := featureService.FeatureSceneMap["test"][0]
			if fresh.featureCache() == c {
				t.Fatal("incompatible reload reused cache")
			}
			if _, ok := c.GetIfPresent("same"); !ok {
				t.Fatal("old cache was cleared before the delay")
			}
			if len(reloadTestInvalidations) != 1 || reloadTestInvalidations[0].delay != 30*time.Second {
				t.Fatal("old cache was not scheduled for delayed invalidation")
			}
			flushReloadTestInvalidations()
			waitReloadCacheMiss(t, c, "same")
			if fresh.featureCache() != nil {
				if _, ok := fresh.featureCache().GetIfPresent("same"); ok {
					t.Fatal("old value entered new cache")
				}
			}
		})
	}
}

func TestFeatureConfsFailedReloadKeepsOldCaches(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadTestService(t)
	conf, _ := reloadTestConfig(t, "user")
	reloadScene(conf)
	old := featureService.FeatureSceneMap["test"][0]
	oldCache := old.featureCache()
	oldCache.Put("same", map[string]any{"value": "old"})
	oldSign := featureService.FeatureSceneSigns["test"]
	base := countGoroutineFrames(goburrowProcessFrame)
	changed := conf
	changed.FeatureDaoConf.CacheTime++
	bad := conf
	bad.Features = []recconf.FeatureConfig{{FeatureType: "unknown"}}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected construction failure")
			}
		}()
		reloadScene(conf, changed, bad)
	}()
	if featureService.FeatureSceneMap["test"][0] != old || featureService.FeatureSceneSigns["test"] != oldSign {
		t.Fatal("failed reload changed the published scene")
	}
	if _, ok := oldCache.GetIfPresent("same"); !ok {
		t.Fatal("failed reload closed inherited cache")
	}
	waitReloadCacheCount(t, base)
	if len(reloadTestInvalidations) != 0 {
		t.Fatal("failed scene scheduled old-cache invalidation")
	}
	bad.FeatureDaoConf.FeatureStoreName = "missing_reload_test_client"
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected missing client failure")
			}
		}()
		reloadScene(changed, bad)
	}()
	waitReloadCacheCount(t, base)
}

func waitReloadCacheMiss(t *testing.T, c cache.Cache, key string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := c.GetIfPresent(key); !ok {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("old cache data was not invalidated")
}

func waitReloadCacheCount(t *testing.T, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for countGoroutineFrames(goburrowProcessFrame) != want && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := countGoroutineFrames(goburrowProcessFrame); got != want {
		t.Fatalf("cache goroutines=%d, want %d", got, want)
	}
}

func TestFeatureConfsOldRequestLatePut(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadTestService(t)
	conf, view := reloadTestConfig(t, "user")
	entered, release := make(chan struct{}), make(chan struct{})
	view.fetch = func(keys []any) []map[string]any {
		close(entered)
		<-release
		return []map[string]any{{"id": keys[0], "value": "old_remote"}}
	}
	reloadScene(conf)
	old := featureService.FeatureSceneMap["test"][0]
	removed := make(chan struct{})
	c := cache.New(cache.WithMaximumSize(100), cache.WithRemovalListener(func(key cache.Key, _ cache.Value) {
		if key == "sentinel" {
			close(removed)
		}
	}))
	reloadTestCaches[c] = true
	old.featureDao = module.NewFeatureDaoWithCache(conf.FeatureDaoConf, c)
	c.Put("sentinel", "old")
	u := module.NewUser("same")
	u.AddProperty("id", "same")
	done := make(chan struct{})
	go func() { defer close(done); old.LoadFeatures(u, nil, reccontext.NewRecommendContext()) }()
	<-entered
	conf.FeatureDaoConf.CacheTime++
	reloadScene(conf)
	flushReloadTestInvalidations()
	select {
	case <-removed:
	case <-time.After(2 * time.Second):
		t.Fatal("old cache data was not physically removed")
	}
	close(release)
	<-done
	if u.StringProperty("value") != "old_remote" {
		t.Fatal("old request could not finish")
	}
	if _, ok := c.GetIfPresent("same"); !ok {
		t.Fatal("old request must retain the original cache Put behavior")
	}
	if _, ok := featureService.FeatureSceneMap["test"][0].featureCache().GetIfPresent("same"); ok {
		t.Fatal("old request populated new cache")
	}
}

func TestFeatureConfsConcurrentLoad(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadTestService(t)
	conf, _ := reloadTestConfig(t, "user")
	reloadScene(conf)
	c := featureService.FeatureSceneMap["test"][0].featureCache()
	c.Put("same", map[string]any{"value": "user"})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if i == 0 {
					updated := conf
					updated.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: fmt.Sprint("copy", j), FeatureStore: "user", FeatureSource: "user:value"}}
					reloadScene(updated)
				} else {
					u := module.NewUser("same")
					u.AddProperty("id", "same")
					ctx := reccontext.NewRecommendContext()
					ctx.Param = localReloadParams{"scene": "test"}
					if i == 1 {
						featureService.LoadFeatures(u, nil, ctx)
					} else {
						featureService.LoadFeaturesForGeneralRank(u, nil, ctx, "")
						featureService.LoadFeaturesForPipelineRank(u, nil, ctx, "")
					}
				}
			}
		}(i)
	}
	wg.Wait()
}

// This local PostgreSQL protocol stub keeps the original driver and registry intact.
type reloadHologresDriver struct {
	queries  atomic.Int32
	prepared atomic.Int32
}

func writeReloadPG(c net.Conn, kind byte, payload []byte) {
	message := make([]byte, 5, 5+len(payload))
	message[0] = kind
	binary.BigEndian.PutUint32(message[1:], uint32(4+len(payload)))
	_, _ = c.Write(append(message, payload...))
}

func reloadPGColumns() []byte {
	columns := []byte{0, 2}
	for _, name := range []string{"id", "value"} {
		columns = append(columns, []byte(name)...)
		columns = append(columns, 0)
		columns = append(columns, 0, 0, 0, 0, 0, 0, 0, 0, 0, 25, 255, 255, 255, 255, 255, 255, 0, 0)
	}
	return columns
}

func (d *reloadHologresDriver) serve(c net.Conn) {
	var length [4]byte
	if _, err := io.ReadFull(c, length[:]); err != nil {
		return
	}
	startup := make([]byte, int(binary.BigEndian.Uint32(length[:]))-4)
	if _, err := io.ReadFull(c, startup); err != nil {
		return
	}
	writeReloadPG(c, 'R', []byte{0, 0, 0, 0})
	writeReloadPG(c, 'S', []byte("server_version\x0014.0\x00"))
	writeReloadPG(c, 'S', []byte("client_encoding\x00UTF8\x00"))
	writeReloadPG(c, 'Z', []byte{'I'})
	var kind [1]byte
	var query string
	var params []string
	for {
		if _, err := io.ReadFull(c, kind[:]); err != nil {
			return
		}
		if _, err := io.ReadFull(c, length[:]); err != nil {
			return
		}
		payload := make([]byte, int(binary.BigEndian.Uint32(length[:]))-4)
		if _, err := io.ReadFull(c, payload); err != nil {
			return
		}
		switch kind[0] {
		case 'P':
			query = strings.Split(string(payload), "\x00")[1]
			if strings.HasPrefix(query, "SELECT") {
				d.prepared.Add(1)
			}
			writeReloadPG(c, '1', nil)
		case 'D':
			count := strings.Count(query, "$")
			types := make([]byte, 2+4*count)
			binary.BigEndian.PutUint16(types, uint16(count))
			for i := 0; i < count; i++ {
				binary.BigEndian.PutUint32(types[2+4*i:], 25)
			}
			writeReloadPG(c, 't', types)
			if strings.HasPrefix(query, "SELECT") {
				writeReloadPG(c, 'T', reloadPGColumns())
			} else {
				writeReloadPG(c, 'n', nil)
			}
		case 'B':
			p := payload
			for i := 0; i < 2; i++ {
				p = p[strings.IndexByte(string(p), 0)+1:]
			}
			formats := int(binary.BigEndian.Uint16(p))
			p = p[2+2*formats:]
			count := int(binary.BigEndian.Uint16(p))
			p = p[2:]
			params = nil
			for i := 0; i < count; i++ {
				n := int(int32(binary.BigEndian.Uint32(p)))
				p = p[4:]
				if n >= 0 {
					params = append(params, string(p[:n]))
					p = p[n:]
				}
			}
			writeReloadPG(c, '2', nil)
		case 'E':
			if strings.HasPrefix(query, "SELECT") {
				d.queries.Add(1)
				for _, key := range params {
					if key == "-1" || key == "missing" {
						continue
					}
					row := []byte{0, 2}
					for _, value := range []string{key, "hologres_remote"} {
						var size [4]byte
						binary.BigEndian.PutUint32(size[:], uint32(len(value)))
						row = append(row, size[:]...)
						row = append(row, []byte(value)...)
					}
					writeReloadPG(c, 'D', row)
				}
				writeReloadPG(c, 'C', []byte("SELECT 1\x00"))
			} else {
				writeReloadPG(c, 'C', []byte("SET\x00"))
			}
		case 'S':
			writeReloadPG(c, 'Z', []byte{'I'})
		case 'Q':
			writeReloadPG(c, 'I', nil)
			writeReloadPG(c, 'Z', []byte{'I'})
		case 'C':
			writeReloadPG(c, '3', nil)
		case 'X':
			return
		}
	}
}

func newReloadHologresConfig(t *testing.T) (recconf.FeatureLoadConfig, *reloadHologresDriver) {
	t.Helper()
	d := &reloadHologresDriver{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var connections sync.Map
	var wg sync.WaitGroup
	acceptDone := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(acceptDone)
		for {
			c, err := listener.Accept()
			if err != nil {
				return
			}
			connections.Store(c, true)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer c.Close()
				d.serve(c)
			}()
		}
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		<-acceptDone
		connections.Range(func(key, _ any) bool { _ = key.(net.Conn).Close(); return true })
		wg.Wait()
	})
	address := listener.Addr().(*net.TCPAddr)
	dsn := fmt.Sprintf("host=127.0.0.1 port=%d user=test dbname=test sslmode=disable", address.Port)
	holo.Load(&recconf.RecommendConfig{HologresConfs: map[string]recconf.HologresConfig{
		"reload_holo": {DSN: dsn},
	}})
	postgres, err := holo.GetPostgres("reload_holo")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = postgres.DB.Close() })
	return recconf.FeatureLoadConfig{FeatureDaoConf: recconf.FeatureDaoConfig{
		DaoConfig: recconf.DaoConfig{AdapterType: recconf.DaoConf_Adapter_Hologres,
			HologresName: "reload_holo", HologresTableName: "features"},
		FeatureStore: "user", FeatureKey: "user:id", UserFeatureKeyName: "id",
		ItemFeatureKeyName: "id", UserSelectFields: "id,value", ItemSelectFields: "value",
		CacheSize: 100, CacheTime: 60,
	}}, d
}

func TestFeatureConfsHologresCacheReload(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	conf, driver := newReloadHologresConfig(t)
	for _, store := range []string{"user", "item"} {
		t.Run(store, func(t *testing.T) {
			reloadTestService(t)
			conf := conf
			conf.FeatureDaoConf.FeatureStore = store
			conf.FeatureDaoConf.FeatureKey = store + ":id"
			reloadScene(conf)
			old := featureService.FeatureSceneMap["test"][0]
			if old.daoSign == "" || old.featureCache() == nil {
				t.Fatal("Hologres was not included in cache inheritance")
			}
			fetch := func(f *Feature, key string) string {
				u := module.NewUser(key)
				u.AddProperty("id", key)
				item := module.NewItem(key)
				f.LoadFeatures(u, []*module.Item{item}, reccontext.NewRecommendContext())
				if store == "user" {
					return u.StringProperty("value")
				}
				return item.StringProperty("value")
			}
			before := driver.queries.Load()
			if fetch(old, "same") != "hologres_remote" || driver.queries.Load() != before+1 {
				t.Fatal("original Hologres DAO did not successfully fetch data")
			}
			conf.Features = []recconf.FeatureConfig{{FeatureType: "raw_feature", FeatureName: "copy",
				FeatureStore: store, FeatureSource: store + ":value"}}
			reloadScene(conf)
			fresh := featureService.FeatureSceneMap["test"][0]
			if fresh.featureDao == old.featureDao || fresh.featureCache() != old.featureCache() {
				t.Fatal("reload must rebuild the DAO and inherit only its cache")
			}
			if fetch(fresh, "same") != "hologres_remote" || driver.queries.Load() != before+1 {
				t.Fatal("inherited Hologres cache fetched data again")
			}
			u := module.NewUser("same")
			u.AddProperty("id", "same")
			item := module.NewItem("same")
			fresh.LoadFeatures(u, []*module.Item{item}, reccontext.NewRecommendContext())
			copyValue := u.StringProperty("copy")
			if store == "item" {
				copyValue = item.StringProperty("copy")
			}
			if copyValue != "hologres_remote" {
				t.Fatal("new conversion did not run on inherited cache data")
			}
			prepared := driver.prepared.Load()
			if fetch(fresh, "new") != "hologres_remote" || driver.prepared.Load() != prepared+1 {
				t.Fatal("new DAO did not initialize its own SQL statement")
			}
			if store == "item" {
				beforeMissing := driver.queries.Load()
				fetch(fresh, "missing")
				deadline := time.Now().Add(2 * time.Second)
				for {
					if value, ok := fresh.featureCache().GetIfPresent("missing"); ok && len(value.(map[string]any)) == 0 {
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("original Hologres negative cache was not populated")
					}
					time.Sleep(time.Millisecond)
				}
				conf.Features[0].FeatureName = "copy2"
				reloadScene(conf)
				fetch(featureService.FeatureSceneMap["test"][0], "missing")
				if driver.queries.Load() != beforeMissing+1 {
					t.Fatal("inherited negative cache fetched data again")
				}
			}
			if len(reloadTestInvalidations) != 0 {
				t.Fatal("compatible reload scheduled cache invalidation")
			}
			first, second := LoadWithConfig(conf), LoadWithConfig(conf)
			for _, f := range []*Feature{first, second} {
				c := f.featureCache()
				t.Cleanup(func() { _ = c.Close() })
			}
			if first.daoSign != "" || first.featureCache() == second.featureCache() {
				t.Fatal("ordinary Hologres constructors must remain independent")
			}
		})
	}
	t.Run("mixed", func(t *testing.T) {
		reloadTestService(t)
		user, item := conf, conf
		item.FeatureDaoConf.FeatureStore = "item"
		item.FeatureDaoConf.FeatureKey = "item:id"
		reloadScene(user, item)
		old := featureService.FeatureSceneMap["test"]
		if old[0].featureCache() == old[1].featureCache() {
			t.Fatal("user and item Hologres entries shared a cache")
		}
		old[0].featureCache().Put("same", map[string]any{"value": "user"})
		old[1].featureCache().Put("same", map[string]any{"value": "item"})
		reloadScene(item, user)
		fresh := featureService.FeatureSceneMap["test"]
		if fresh[0].featureCache() != old[1].featureCache() || fresh[1].featureCache() != old[0].featureCache() {
			t.Fatal("Hologres reorder mixed user and item caches")
		}
	})
}

func TestFeatureConfsHologresCacheIsolation(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	conf, _ := newReloadHologresConfig(t)
	changes := map[string]func(*recconf.FeatureLoadConfig){
		"ttl":        func(c *recconf.FeatureLoadConfig) { c.FeatureDaoConf.CacheTime++ },
		"size":       func(c *recconf.FeatureLoadConfig) { c.FeatureDaoConf.CacheSize++ },
		"table":      func(c *recconf.FeatureLoadConfig) { c.FeatureDaoConf.HologresTableName = "other" },
		"connection": func(c *recconf.FeatureLoadConfig) { c.FeatureDaoConf.HologresName = "missing" },
		"fields":     func(c *recconf.FeatureLoadConfig) { c.FeatureDaoConf.UserSelectFields = "value" },
		"key":        func(c *recconf.FeatureLoadConfig) { c.FeatureDaoConf.FeatureKey = "user:other" },
		"adapter": func(c *recconf.FeatureLoadConfig) {
			fsConf, _ := reloadTestConfig(t, "user")
			*c = fsConf
		},
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			reloadTestService(t)
			conf := conf
			reloadScene(conf)
			old := featureService.FeatureSceneMap["test"][0]
			c := old.featureCache()
			if c == nil {
				t.Fatal("Hologres cache is not exposed for inheritance")
			}
			c.Put("same", map[string]any{"value": "old"})
			change(&conf)
			if name == "connection" {
				func() {
					defer func() {
						if recover() == nil {
							t.Error("expected a missing connection failure")
						}
					}()
					reloadScene(conf)
				}()
				if featureService.FeatureSceneMap["test"][0] != old || len(reloadTestInvalidations) != 0 {
					t.Fatal("failed construction changed the old scene")
				}
				return
			}
			reloadScene(conf)
			fresh := featureService.FeatureSceneMap["test"][0]
			if fresh.featureCache() == c || len(reloadTestInvalidations) != 1 || reloadTestInvalidations[0].delay != 30*time.Second {
				t.Fatal("changed DAO must get a separate cache and delay old-cache invalidation")
			}
			if _, ok := c.GetIfPresent("same"); !ok {
				t.Fatal("old data was cleared before the delay")
			}
			flushReloadTestInvalidations()
			waitReloadCacheMiss(t, c, "same")
			if _, ok := fresh.featureCache().GetIfPresent("same"); ok {
				t.Fatal("new cache contains old data")
			}
		})
	}
}

type localReloadParams map[string]any

func (p localReloadParams) GetParameter(name string) any { return p[name] }

func TestFeatureConfsOtherPathsUnchanged(t *testing.T) {
	if runReloadTestInSubprocess(t) {
		return
	}
	reloadTestService(t)
	conf, _ := reloadTestConfig(t, "user")
	first, second := LoadWithConfig(conf), LoadWithConfig(conf)
	if first.daoSign != "" || first.featureCache() == second.featureCache() {
		t.Fatal("ordinary constructors joined reload management")
	}
	for _, f := range []*Feature{first, second} {
		c := f.featureCache()
		t.Cleanup(func() { _ = c.Close() })
	}
	userConf := &recconf.RecommendConfig{UserFeatureConfs: map[string]recconf.SceneFeatureConfig{"unrelated": {FeatureLoadConfs: []recconf.FeatureLoadConfig{conf}}}}
	original := userFeatureService
	userFeatureService = NewUserFeatureService()
	t.Cleanup(func() { userFeatureService = original })
	UserLoadFeatureConfig(userConf)
	userFeature := userFeatureService.FeatureSceneMap["unrelated"][0]
	if userFeature.daoSign != "" || userFeature.featureCache() == nil {
		t.Fatal("B1 changed UserFeatureConfs")
	}
	userCache := userFeature.featureCache()
	t.Cleanup(func() { _ = userCache.Close() })
}

type featureCacheInvalidationProbe struct {
	cache.Cache
	done chan struct{}
}

func (c *featureCacheInvalidationProbe) InvalidateAll() {
	c.Cache.InvalidateAll()
	close(c.done)
}

func TestFeatureCacheInvalidateAllTimer(t *testing.T) {
	const entries = 128
	removed := make(chan struct{})
	var removals atomic.Int32
	c := &featureCacheInvalidationProbe{Cache: cache.New(cache.WithMaximumSize(entries*2),
		cache.WithRemovalListener(func(_ cache.Key, _ cache.Value) {
			if removals.Add(1) == entries {
				close(removed)
			}
		})), done: make(chan struct{})}
	t.Cleanup(func() { _ = c.Close() })
	for i := 0; i < entries; i++ {
		c.Put(fmt.Sprint(i), "value")
	}
	if _, ok := c.GetIfPresent("0"); !ok {
		t.Fatal("initial cache value is missing")
	}
	timer := featureCacheAfterFunc(20*time.Millisecond, c.InvalidateAll)
	t.Cleanup(func() {
		if !timer.Stop() {
			<-c.done
		}
	})
	select {
	case <-c.done:
	case <-time.After(2 * time.Second):
		t.Fatal("invalidation timer did not fire")
	}
	select {
	case <-removed:
	case <-time.After(2 * time.Second):
		t.Fatal("invalidation did not physically remove the value")
	}
	for i := 0; i < entries; i++ {
		waitReloadCacheMiss(t, c, fmt.Sprint(i))
	}
	c.Put("after", "still usable")
	if value, ok := c.GetIfPresent("after"); !ok || value != "still usable" {
		t.Fatal("data invalidation must preserve the original cache operations")
	}
}

func TestFeatureCacheConcurrentInvalidation(t *testing.T) {
	c := cache.New(cache.WithMaximumSize(100))
	t.Cleanup(func() { _ = c.Close() })
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				c.Put("shared", "value")
				c.GetIfPresent("shared")
				c.InvalidateAll()
			}
		}()
	}
	wg.Wait()
}
