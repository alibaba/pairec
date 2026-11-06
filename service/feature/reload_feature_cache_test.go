package feature

import (
	"bytes"
	"fmt"
	"runtime/pprof"
	"strings"
	"testing"

	reccontext "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

// goburrowProcessFrame is the stack frame of the background goroutine that
// every goburrow cache starts in cache.New and only stops in Close.
const goburrowProcessFrame = "goburrow/cache.(*localCache).processEntries("

func countGoroutineFrames(frame string) int {
	var buf bytes.Buffer
	_ = pprof.Lookup("goroutine").WriteTo(&buf, 2)
	return strings.Count(buf.String(), frame)
}

// TestReloadFeatureConfigLeaksDaoCaches verifies cache reuse and the retained
// workers of replaced DAO caches under data-only invalidation.
func TestReloadFeatureConfigLeaksDaoCaches(t *testing.T) {
	const (
		cachedFeatures = 10
		reloads        = 5
	)

	t.Run("item features", func(t *testing.T) {
		if runReloadTestInSubprocess(t) {
			return
		}
		reloadTestService(t)
		conf, view := reloadTestConfig(t, "item")
		base := countGoroutineFrames(goburrowProcessFrame)
		confs := make([]recconf.FeatureLoadConfig, cachedFeatures)
		project := prepareReloadFS(t)
		for i := range confs {
			confs[i] = conf
			name := fmt.Sprintf("%s_%d", conf.FeatureDaoConf.FeatureStoreViewName, i)
			confs[i].FeatureDaoConf.FeatureStoreViewName = name
			project.FeatureViewMap.Store(name, view)
			t.Cleanup(func() { project.FeatureViewMap.Delete(name) })
		}
		reloadScene(confs...)
		first := featureService.FeatureSceneMap["test"]
		reloadScene(confs...)
		if featureService.FeatureSceneMap["test"][0] != first[0] {
			t.Fatal("unchanged scene was rebuilt")
		}
		item := module.NewItem("same")
		first[1].LoadFeatures(module.NewUser("u"), []*module.Item{item}, reccontext.NewRecommendContext())
		if item.StringProperty("value") != "item" || view.calls.Load() != 1 {
			t.Fatal("valid DAO did not fetch item features")
		}
		for i := 1; i <= reloads; i++ {
			confs[0].FeatureDaoConf.CacheTime = 3600 + i
			reloadScene(confs...)
			fresh := featureService.FeatureSceneMap["test"]
			if fresh[1].featureCache() != first[1].featureCache() {
				t.Fatal("unchanged DAO cache was rebuilt")
			}
			item = module.NewItem("same")
			fresh[1].LoadFeatures(module.NewUser("u"), []*module.Item{item}, reccontext.NewRecommendContext())
			if item.StringProperty("value") != "item" || view.calls.Load() != 1 {
				t.Fatal("inherited item cache missed")
			}
		}
		if len(reloadTestInvalidations) != reloads {
			t.Fatal("only changed DAO caches should be scheduled for invalidation")
		}
		flushReloadTestInvalidations()
		waitReloadCacheCount(t, base+cachedFeatures+reloads)
		t.Logf("item scene: %d cached features, %d reloads, live cache goroutines=%d (old cache workers are retained)", cachedFeatures, reloads, cachedFeatures+reloads)
	})

}
