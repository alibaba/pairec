package service

import (
	"fmt"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/recall"
)

type Category struct {
	CategoryName string
	recalls      []recall.Recall
	recallNames  []string
}

func NewCategory(name string) *Category {
	c := Category{CategoryName: name}
	c.recalls = make([]recall.Recall, 0)
	c.recallNames = make([]string, 0)
	return &c
}

func (c *Category) Init(config recconf.CategoryConfig) {
	var (
		recalls     []recall.Recall
		recallNames []string
	)
	for _, recallName := range config.RecallNames {
		recall, err := recall.GetRecall(recallName)
		if err != nil {
			log.Error(fmt.Sprintf("module=category init\terror=%v", err))
			continue
		}

		recalls = append(recalls, recall)
		recallNames = append(recallNames, recallName)
	}

	c.recalls = recalls
	c.recallNames = recallNames
}

func (c *Category) GetRecalls() []recall.Recall {
	return c.recalls
}

func (c *Category) GetRecallNames() []string {
	return c.recallNames
}
