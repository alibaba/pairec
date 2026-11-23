package recallenginerecall

import (
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/context"
)

func TestUserRealtimeEmbeddingTriggerBuildAlgoData(t *testing.T) {
	userFeatures := map[string]interface{}{"uid": "user_1", "age": 20}

	t.Run("debug request sets debug level", func(t *testing.T) {
		trigger := &UserRealtimeEmbeddingTrigger{recallAlgoType: eas.Eas_Processor_EASYREC}
		recommendContext := context.NewRecommendContext()
		recommendContext.Debug = true
		recommendContext.RecommendId = "request_1"

		algoData := trigger.buildAlgoData(userFeatures, recommendContext)
		request, ok := algoData.GetFeatures().(*easyrec.PBRequest)
		assert.Equal(t, ok, true)
		assert.Equal(t, request.DebugLevel, int32(realtimeEmbeddingDebugLevel))
		assert.Equal(t, request.MetaData["request_id"], "request_1")
	})

	t.Run("non debug request sets no debug level", func(t *testing.T) {
		trigger := &UserRealtimeEmbeddingTrigger{recallAlgoType: eas.Eas_Processor_EASYREC}
		recommendContext := context.NewRecommendContext()
		recommendContext.RecommendId = "request_2"

		algoData := trigger.buildAlgoData(userFeatures, recommendContext)
		request, ok := algoData.GetFeatures().(*easyrec.PBRequest)
		assert.Equal(t, ok, true)
		assert.Equal(t, request.DebugLevel, int32(0))
		assert.Equal(t, len(request.MetaData), 0)
	})
}
