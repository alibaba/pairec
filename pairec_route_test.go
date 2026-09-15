package pairec

import (
	"testing"

	"github.com/alibaba/pairec/v2/web"
)

func TestRecommendRoutesUseRecommendController(t *testing.T) {
	previous := PairecApp
	PairecApp = NewApp()
	t.Cleanup(func() { PairecApp = previous })
	registerRouteInfo()

	for _, path := range []string{"/api/recommend", "/api/rec/feed"} {
		info, exists := PairecApp.Handlers.routeInfos[path]
		if !exists {
			t.Fatalf("route %s is not registered", path)
		}
		if _, ok := info.initialize().(*web.RecommendController); !ok {
			t.Fatalf("route %s does not use RecommendController", path)
		}
	}
}
