package berecall

import (
	gocontext "context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/service/rank"

	"github.com/alibaba/pairec/context"
	plog "github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/service/feature"
	"github.com/alibaba/pairec/utils"
)

type UserEmbeddingDssmO2OTrigger struct {
	useCacheFeatures             bool
	seqDelimiter                 string
	multiValueDelimiter          string
	BizName                      string
	RecallName                   string
	BeName                       string
	o2oType                      string
	features                     []*feature.Feature
	featureConsistencyJobService *rank.FeatureConsistencyJobService
}

func NewUserEmbeddingDssmO2OTrigger(config *recconf.UserEmbeddingO2OTriggerConfig) *UserEmbeddingDssmO2OTrigger {
	trigger := &UserEmbeddingDssmO2OTrigger{
		seqDelimiter:                 ";",
		multiValueDelimiter:          "\u001D",
		featureConsistencyJobService: new(rank.FeatureConsistencyJobService),
		BeName:                       config.BeName,
		RecallName:                   config.RecallName,
		BizName:                      config.BizName,
		o2oType:                      rank.DssmO2o,
	}
	var features []*feature.Feature
	for _, conf := range config.UserFeatureConfs {
		if conf.FeatureDaoConf.LoadFromCacheFeaturesName != "" {
			trigger.useCacheFeatures = true
		}
		f := feature.LoadWithConfig(conf)
		features = append(features, f)
	}

	trigger.features = features
	if config.SeqDelimiter != "" {
		trigger.seqDelimiter = config.SeqDelimiter
	}
	if config.MultiValueDelimiter != "" {
		trigger.multiValueDelimiter = config.MultiValueDelimiter
	}

	return trigger
}
func NewUserEmbeddingMindO2OTrigger(config *recconf.UserEmbeddingO2OTriggerConfig) *UserEmbeddingDssmO2OTrigger {
	trigger := NewUserEmbeddingDssmO2OTrigger(config)
	trigger.o2oType = rank.MindO2o
	return trigger
}
func (t *UserEmbeddingDssmO2OTrigger) loadUserFeatures(user *module.User, context *context.RecommendContext) {
	var wg sync.WaitGroup
	for _, fea := range t.features {
		wg.Add(1)
		go func(fea *feature.Feature) {
			defer wg.Done()
			fea.LoadFeatures(user, nil, context)
		}(fea)
	}

	wg.Wait()

}
func (t *UserEmbeddingDssmO2OTrigger) GetTriggerKey(u *module.User, context *context.RecommendContext) *TriggerResult {
	//start := time.Now()
	if t.useCacheFeatures {
		ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 100*time.Millisecond)
		defer cancel()
		select {
		case <-u.FeatureAsyncLoadCh():
		case <-ctx.Done():
			plog.Error(fmt.Sprintf("requestId=%s\tmodule=UserEmbeddingDssmO2OTrigger\terror=%v", context.RecommendId, ctx.Err()))
		}
	}

	user := u.Clone()
	t.loadUserFeatures(user, context)

	userFeatures := user.MakeUserFeatures2()
	features := make(map[string]any, len(userFeatures))
	for k, v := range userFeatures {
		switch val := v.(type) {
		case string:
			if strings.Contains(val, t.seqDelimiter) {
				features[fmt.Sprintf("user:%s", k)] = strings.Split(val, t.seqDelimiter)
			} else if strings.Contains(val, t.multiValueDelimiter) {
				features[fmt.Sprintf("user:%s", k)] = strings.Split(val, t.multiValueDelimiter)
			} else {
				features[fmt.Sprintf("user:%s", k)] = []string{val}
			}
		case []string:
			features[fmt.Sprintf("user:%s", k)] = val
		case []any:
			features[fmt.Sprintf("user:%s", k)] = val
		default:
			if valStr := utils.ToString(val, ""); valStr != "" {
				features[fmt.Sprintf("user:%s", k)] = []string{valStr}
			}
		}
	}

	j, _ := json.Marshal(features)

	qinfo := base64.StdEncoding.EncodeToString(j)
	if t.BizName != "" && t.RecallName != "" && t.BeName != "" {
		go t.featureConsistencyJobService.LogRecallResult(user, nil, context, t.o2oType, "", qinfo, "", "", t.BeName, t.BizName, t.RecallName)
	}
	triggerResult := &TriggerResult{
		TriggerItem: qinfo, // user feature base64
	}
	return triggerResult
}
