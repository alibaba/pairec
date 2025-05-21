package fallback

import (
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
)

type IFallback interface {
	GetTimer() *time.Timer
	Recommend(context *context.RecommendContext) []*module.Item
}
