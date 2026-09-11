package recall

import (
	"context"
	"fmt"
	"strings"

	pairecctx "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/ha3engine"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type Ha3KnowledgeRecall struct {
	*BaseRecall
	knowledge *ha3KnowledgeSearcher
}

func NewHa3KnowledgeRecall(config recconf.RecallConfig) *Ha3KnowledgeRecall {
	if config.Ha3KnowledgeVectorConf == nil {
		panic("HA3_KNOWLEDGE_RECALL requires Ha3KnowledgeVectorConf")
	}
	engineName := strings.TrimSpace(config.Ha3KnowledgeVectorConf.EngineName)
	if engineName == "" {
		panic("Ha3KnowledgeVectorConf.EngineName is required for HA3_KNOWLEDGE_RECALL")
	}
	client, err := ha3engine.GetHa3EngineClient(engineName)
	if err != nil {
		panic(fmt.Errorf("get HA3 knowledge engine %q: %w", engineName, err))
	}
	return &Ha3KnowledgeRecall{
		BaseRecall: NewBaseRecall(config),
		knowledge:  newHa3KnowledgeSearcher(client, *config.Ha3KnowledgeVectorConf),
	}
}

func (r *Ha3KnowledgeRecall) GetCandidateItems(*module.User, *pairecctx.RecommendContext) []*module.Item {
	return nil
}

func (r *Ha3KnowledgeRecall) KnowledgeEnabled() bool {
	return r != nil && r.knowledge != nil
}

func (r *Ha3KnowledgeRecall) SearchKnowledge(ctx context.Context, query string) (*KnowledgeSearchResult, error) {
	if !r.KnowledgeEnabled() {
		return nil, nil
	}
	return r.knowledge.SearchKnowledge(ctx, query)
}
