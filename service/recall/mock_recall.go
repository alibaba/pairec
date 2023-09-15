package recall

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
)

type MockRecall struct {
	*BaseRecall
}

func NewMockRecall(config recconf.RecallConfig) *MockRecall {
	recall := &MockRecall{
		BaseRecall: NewBaseRecall(config),
	}
	return recall
}

func (r *MockRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()

	for len(ret) < r.recallCount {
		id := rand.Uint32()
		item := module.NewItem(strconv.Itoa(int(id)))
		item.RetrieveId = r.modelName
		item.Score = rand.Float64()

		ret = append(ret, item)
	}

	log.Info(fmt.Sprintf("requestId=%s\tmodule=MockRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
