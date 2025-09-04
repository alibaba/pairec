package fallback

import (
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
)

type IFallback interface {
	GetTimer() *time.Timer
	PutTimer(*time.Timer)
	CompleteItemsIfNeed() bool
	Recommend(context *context.RecommendContext) []*module.Item
}
