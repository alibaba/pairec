package recall

import (
	"fmt"
	"testing"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
)

func TestMultibizRecall(t *testing.T) {
	config := recconf.BeConfig{
		BeName:       "be",
		BizName:      "shihuo_common",
		BeRecallType: "multi_merge_recall",
		BeRecallParams: []recconf.BeRecallParam{
			{
				Count:        100,
				RecallType:   "x2i_recall",
				RecallName:   "global_hot",
				TriggerType:  "fixvalue",
				TriggerValue: "-1",
				Priority:     5,
				ItemIdName:   "item_id",
			},
			{
				Count:       100,
				RecallType:  "x2i_recall",
				RecallName:  "group_hot",
				TriggerType: "user",
				UserTriggers: []recconf.TriggerConfig{
					{
						TriggerKey: "sex",
					},
				},

				Priority:   5,
				ItemIdName: "item_id",
			},
		},
	}
	berecall := NewBeMultiBizRecall(nil, config, "model")

	user := module.NewUser("ee145bfad044f86a")
	user.AddProperty("sex", "male")

	ctx := context.NewRecommendContext()
	ctx.Debug = true
	multiReadRequest := berecall.buildRequest(user, ctx)
	uri := multiReadRequest.BuildUri()
	fmt.Println(uri.RequestURI())
}
