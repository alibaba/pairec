package aishopping

import (
	recallsvc "github.com/alibaba/pairec/v2/service/recall"
	"github.com/alibaba/pairec/v2/service/shoppingknowledge"
)

type knowledgeEvidence = shoppingknowledge.Evidence

func newKnowledgeEvidence(result *recallsvc.KnowledgeSearchResult) *knowledgeEvidence {
	return shoppingknowledge.NewEvidence(result)
}
