package recall

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/faiss/pai_web"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type VectorRecall struct {
	*BaseRecall
	dao module.VectorDao
}

func NewVectorRecall(config recconf.RecallConfig) *VectorRecall {
	recall := &VectorRecall{
		BaseRecall: NewBaseRecall(config),
		dao:        module.NewVectorDao(config),
	}
	return recall
}

func (r *VectorRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()

	if r.cache != nil {
		key := r.cachePrefix + string(user.Id)
		cacheRet := r.cache.Get(key)
		if itemStr, ok := cacheRet.([]uint8); ok {
			itemIds := strings.Split(string(itemStr), ",")
			for _, id := range itemIds {
				var item *module.Item
				if strings.Contains(id, ":") {
					vars := strings.Split(id, ":")
					item = module.NewItem(vars[0])
					f, _ := strconv.ParseFloat(vars[2], 64)
					// item.AddProperty(vars[1], f)
					item.Score = f
				} else {
					item = module.NewItem(id)
				}
				item.RetrieveId = r.modelName
				item.ItemType = r.itemType
				ret = append(ret, item)
			}
			log.Info(fmt.Sprintf("requestId=%s\tmodule=VectorRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
			return
		}
	}
	value, err := r.dao.VectorString(string(user.Id))
	if err != nil {
		if errors.Is(err, module.VectoryEmptyError) {
			log.Info(fmt.Sprintf("requestId=%s\tmodule=VectorRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
		} else {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=VectorRecall\tname=%s\tcount=%d\terr=%v\tcost=%d", context.RecommendId, r.modelName, len(ret), err, utils.CostTime(start)))

		}
		return
	}

	request := pai_web.VectorRequest{K: uint32(r.recallCount)}
	request.Vector = make([]float32, 0)
	vectors := strings.Split(value, " ")
	for _, vc := range vectors {
		if !strings.Contains(vc, ":") {
			continue
		}
		vals := strings.Split(vc, ":")
		if len(vals) == 2 {
			value, _ := strconv.ParseFloat(vals[1], 32)
			request.Vector = append(request.Vector, float32(value))
		}
	}
	if len(request.Vector) == 0 {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=VectorRecall\terror=user Vector empty", context.RecommendId))
		return
	}

	result, err := algorithm.Run(r.recallAlgo, &request)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=VectorRecall\terror=%v", context.RecommendId, err))
		return
	}
	reply := result.(*pai_web.VectorReply)
	for i, id := range reply.Labels {
		item := module.NewItem(id)
		item.RetrieveId = r.modelName
		item.ItemType = r.itemType
		// item.AddProperty(r.modelName, reply.Scores[i])
		item.Score = float64(reply.Scores[i])

		ret = append(ret, item)
	}
	if r.cache != nil && len(ret) > 0 {
		go func() {
			key := r.cachePrefix + string(user.Id)
			var itemIds string
			for _, item := range ret {
				itemIds += fmt.Sprintf("%s:%s:%v", string(item.Id), r.modelName, item.Score) + ","
			}
			itemIds = itemIds[:len(itemIds)-1]
			cacheTime := r.cacheTime
			if cacheTime == 0 {
				cacheTime = 1800
			}
			if err := r.cache.Put(key, itemIds, time.Duration(cacheTime)*time.Second); err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=VectorRecall\terror=%v",
					context.RecommendId, err))
			}
		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=VectorRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
