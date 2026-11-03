package recconf

import (
	"fmt"
	"strings"
	"testing"
)

// featureCacheScopeConfig wraps a consumer config with a single user-scoped
// named feature cache so the scope constraint can be exercised in isolation.
func featureCacheScopeConfig(consumer string) string {
	return fmt.Sprintf(`{
  "FeatureCacheConfs": {
    "shared": {
      "FeatureStoreName": "fs",
      "FeatureStore": "user",
      "FeatureStoreViewName": "user_view",
      "SelectFields": "*",
      "CacheSize": 100,
      "CacheTime": 60
    }
  },
  %s
}`, consumer)
}

func recallScopeConsumer(daoScope string) string {
	return fmt.Sprintf(`"RecallConfs": [{
    "Name": "recall_scope",
    "UserFeatureConfs": [{
      "FeatureDaoConf": {"AdapterType":"featurestore","FeatureCacheName":"shared","FeatureKey":"user:uid","FeatureStore":%q}
    }]
  }]`, daoScope)
}

func generalRankScopeConsumer(daoScope string) string {
	return fmt.Sprintf(`"GeneralRankConfs": {"rank_scope": {"FeatureLoadConfs": [{
      "FeatureDaoConf": {"AdapterType":"featurestore","FeatureCacheName":"shared","FeatureKey":"user:uid","FeatureStore":%q}
    }]}}`, daoScope)
}

func sceneFeatureScopeConsumer(daoScope string) string {
	return fmt.Sprintf(`"FeatureConfs": {"scene_scope": {"FeatureLoadConfs": [{
      "FeatureDaoConf": {"AdapterType":"featurestore","FeatureCacheName":"shared","FeatureKey":"user:uid","FeatureStore":%q}
    }]}}`, daoScope)
}

func moduleIssues(t *testing.T, configData string, index ModuleIndex) []string {
	t.Helper()
	results, err := CheckRecommendConfig(configData)
	if err != nil {
		t.Fatalf("CheckRecommendConfig: %v", err)
	}
	result, ok := results[index]
	if !ok {
		t.Fatalf("no check result for %v", index)
	}
	return result.Issues
}

func hasIssueContaining(issues []string, substr string) bool {
	for _, issue := range issues {
		if strings.Contains(issue, substr) {
			return true
		}
	}
	return false
}

func TestFeatureCacheScopeStaticCheck(t *testing.T) {
	cases := []struct {
		name     string
		daoScope string
		want     bool
	}{
		{name: "mismatch", daoScope: "item", want: true},
		{name: "empty", daoScope: "", want: true},
		{name: "match", daoScope: "user", want: false},
	}
	index := ModuleIndex{Type: RecallConfig{}.ModuleType(), Name: "recall_scope"}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := featureCacheScopeConfig(recallScopeConsumer(tc.daoScope))
			issues := moduleIssues(t, data, index)
			if got := hasIssueContaining(issues, "does not match the cache scope"); got != tc.want {
				t.Fatalf("scope issue present=%v, want %v; issues=%v", got, tc.want, issues)
			}
		})
	}
}

func TestFeatureCacheScopeStaticCheckGeneralRank(t *testing.T) {
	data := featureCacheScopeConfig(generalRankScopeConsumer("item"))
	index := ModuleIndex{Type: GeneralRankConfig{}.ModuleType(), Name: "rank_scope"}
	issues := moduleIssues(t, data, index)
	if !hasIssueContaining(issues, "does not match the cache scope") {
		t.Fatalf("expected scope mismatch issue, got %v", issues)
	}
}

func TestFeatureCacheScopeStaticCheckSceneFeature(t *testing.T) {
	data := featureCacheScopeConfig(sceneFeatureScopeConsumer("item"))
	index := ModuleIndex{Type: SceneFeatureConfig{}.ModuleType(), Name: "scene_scope"}
	issues := moduleIssues(t, data, index)
	if !hasIssueContaining(issues, "does not match the cache scope") {
		t.Fatalf("expected scope mismatch issue, got %v", issues)
	}
}
