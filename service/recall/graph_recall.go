package recall

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	igraph "github.com/aliyun/aliyun-igraph-go-sdk"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/graph"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type GraphRecall struct {
	*BaseRecall
	graphClient *graph.GraphClient
	ItemId      string
	QueryString string
	Params      []string
}

func NewGraphRecall(config recconf.RecallConfig) *GraphRecall {
	graphClient, err := graph.GetGraphClient(config.GraphConf.GraphName)
	if err != nil {
		panic(err)
	}
	Graph := GraphRecall{
		BaseRecall:  NewBaseRecall(config),
		graphClient: graphClient,
		QueryString: config.GraphConf.QueryString,
		ItemId:      config.GraphConf.ItemId,
		Params:      config.GraphConf.Params,
	}

	return &Graph
}

func (i *GraphRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()
	if i.cache != nil {
		key := i.cachePrefix + string(user.Id)
		cacheRet := i.cache.Get(key)
		if itemStr, ok := cacheRet.([]uint8); ok {
			itemIds := strings.Split(string(itemStr), ",")
			for _, id := range itemIds {
				var item *module.Item
				if strings.Contains(id, ":") {
					vars := strings.Split(id, ":")
					item = module.NewItem(vars[0])
					f, _ := strconv.ParseFloat(vars[1], 64)
					item.Score = f
				} else {
					item = module.NewItem(id)
				}
				item.RetrieveId = i.modelName
				ret = append(ret, item)
			}
			log.Info(fmt.Sprintf("requestId=%s\tmodule=GraphRecall\tfrom=cache\tcount=%d\tcost=%d", context.RecommendId, len(ret), utils.CostTime(start)))
			return
		}
	}
	queryString, err := i.getQueryString(user, context)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=GraphRecall\terr=%s", context.RecommendId, err.Error()))
		return
	}
	//log query_string
	if context.Debug {
		log.Info(fmt.Sprintf("event=GraphRecall\tqueryString=%s", queryString))
	}

	queryParam := make(map[string]string)
	request := igraph.ReadRequest{
		QueryString: queryString,
		QueryParams: queryParam,
	}

	resp, err := i.graphClient.GraphClient.Read(&request)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=GraphRecall\terr=%s", context.RecommendId, err.Error()))
		return
	}

	result := resp.Result

	for _, res := range result {
		for _, data := range res.Data {
			itemId := data[i.ItemId]
			score := data["score"]
			if itemId != nil {
				item := module.NewItem(itemId.(string))
				item.RetrieveId = i.modelName

				if score != nil {
					item.Score = score.(float64)
				}

				ret = append(ret, item)
			}
		}
	}

	if i.cache != nil && len(ret) > 0 {
		go func() {
			key := i.cachePrefix + string(user.Id)
			var itemIds string
			for _, item := range ret {
				itemIds += fmt.Sprintf("%s:%v", string(item.Id), item.Score) + ","
			}
			itemIds = itemIds[:len(itemIds)-1]
			if err = i.cache.Put(key, itemIds, time.Duration(i.cacheTime)*time.Second); err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=GraphRecall\terror=%v", context.RecommendId, err))
			}
		}()
	}

	log.Info(fmt.Sprintf("requestId=%s\tmodule=GraphRecall\tcount=%d\tcost=%d", context.RecommendId, len(ret), utils.CostTime(start)))
	return
}

func (i *GraphRecall) getQueryString(user *module.User, context *context.RecommendContext) (string, error) {

	paramResults := make([]string, len(i.Params))
	for x, param := range i.Params {
		paramArr := strings.Split(param, ".")

		switch len(paramArr) {
		case 1:
			return "", fmt.Errorf("Params(%s) type is error, its type should be a.b or a.b.c", param)
		case 2:
			if paramArr[0] == "user" {
				val := user.StringProperty(paramArr[1])
				paramResults[x] = val
			} else if paramArr[0] == "context" {
				val := utils.ToString(context.GetParameter(paramArr[1]), "")
				paramResults[x] = val
			}
		case 3: //get value from context features
			if paramArr[0] == "context" {
				if paramArr[1] == "features" {
					var featureMap map[string]interface{}
					features := context.GetParameter("features")
					if features != nil {
						featureMap = features.(map[string]interface{})
						val := utils.ToString(featureMap[paramArr[2]], "")
						paramResults[x] = val
					} else {
						return "", errors.New("context.feature is null")
					}
				} else {
					return "", fmt.Errorf("Params(%s) only support context.features.xxx", param)
				}
			} else {
				return "", fmt.Errorf("Params(%s) only support context.features.xxx", param)
			}
		default:
			return "", fmt.Errorf("Params(%s) type is error, its type should be a.b or a.b.c", param)
		}
	}

	queryString := i.QueryString
	for x, result := range paramResults {
		replaceStr := fmt.Sprintf("$%d", x+1) // query string 中是从 $1 开始
		queryString = strings.Replace(queryString, replaceStr, result, 1)
	}
	return queryString, nil
}
