package feature_log

import (
	"fmt"
	"sync"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

// featureViewByScene caches the resolved FeatureDB feature view per scene, so
// the request path no longer re-resolves client -> project -> view on every
// call, and a misconfiguration surfaces once at load time instead of silently
// dropping every request's training data.
var (
	featureViewMu      sync.RWMutex
	featureViewByScene = make(map[string]domain.FeatureView)
)

// Load resolves and caches the feature view of every featurestore feature-log
// scene. It must run after fs.Load, at both startup and config reload. A scene
// that cannot be resolved is logged prominently and left uncached, so its
// FeatureLog becomes a cheap no-op instead of a per-request error; a feature-log
// misconfiguration never blocks serving. The map is rebuilt on every call, so a
// scene removed from the config disappears and a project reloaded by fs.Load is
// picked up fresh rather than staying stale.
func Load(config *recconf.RecommendConfig) {
	resolved := make(map[string]domain.FeatureView)
	for scene, conf := range config.FeatureLogConfs {
		if conf.OutputType != "featurestore" {
			continue
		}
		view, err := resolveFeatureView(conf)
		if err != nil {
			log.Error(fmt.Sprintf("module=FeatureLog\tscene=%s\tevent=Load\terr=%v", scene, err))
			continue
		}
		resolved[scene] = view
	}

	featureViewMu.Lock()
	featureViewByScene = resolved
	featureViewMu.Unlock()
}

// resolveFeatureView resolves the client -> project -> feature view chain of a
// single featurestore feature-log config, reporting which field is bad so the
// load-time error is actionable.
func resolveFeatureView(conf recconf.FeatureLogConfig) (domain.FeatureView, error) {
	fsClient, err := fs.GetFeatureStoreClient(conf.FeatureStoreName)
	if err != nil {
		return nil, fmt.Errorf("feature store client not resolved, FeatureStoreName=%s: %w", conf.FeatureStoreName, err)
	}
	view := fsClient.GetProject().GetFeatureView(conf.FeatureStoreViewName)
	if view == nil {
		return nil, fmt.Errorf("feature view not found, FeatureStoreName=%s, FeatureStoreViewName=%s",
			conf.FeatureStoreName, conf.FeatureStoreViewName)
	}
	return view, nil
}

// getFeatureView returns the feature view resolved at load time for a scene.
func getFeatureView(scene string) (domain.FeatureView, bool) {
	featureViewMu.RLock()
	view, ok := featureViewByScene[scene]
	featureViewMu.RUnlock()
	return view, ok
}
