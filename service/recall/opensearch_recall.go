package recall

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/opensearch"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/alibabacloud-go/tea/tea"
)

type OpenSearchRecall struct {
	*BaseRecall
	openSearchClient *opensearch.OpenSearchClient
	AppName          string
	ItemId           string
	RequestParams    map[string]any
	Params           []string
}

func NewOpenSearchRecall(config recconf.RecallConfig) *OpenSearchRecall {
	openSearchClient, err := opensearch.GetOpenSearchClient(config.OpenSearchConf.OpenSearchName)
	if err != nil {
		panic(err)
	}
	recall := OpenSearchRecall{
		BaseRecall:       NewBaseRecall(config),
		openSearchClient: openSearchClient,
		RequestParams:    config.OpenSearchConf.RequestParams,
		AppName:          config.OpenSearchConf.AppName,
		ItemId:           config.OpenSearchConf.ItemId,
		Params:           config.OpenSearchConf.Params,
	}

	return &recall
}

func (i *OpenSearchRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()
	requestParams, err := i.getRequestParams(user, context)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=OpenSearchRecall\terr=%s", context.RecommendId, err.Error()))
		return
	}
	if i.recallCount > 0 {
		requestParams["query"] = fmt.Sprintf("%s&&config=start:0,hit:%d,format:fulljson", requestParams["query"], i.recallCount)
	}
	//log requestParams
	if context.Debug {
		log.Info(fmt.Sprintf("event=OpenSearchRecall\trequest_params=%v", requestParams))
	}

	result, err := i.openSearchClient.OpenSearchClient.Request(tea.String("GET"), tea.String("/v3/openapi/apps/"+i.AppName+"/search"), requestParams, nil, nil, i.openSearchClient.Runtime)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=OpenSearchRecall\terr=%s", context.RecommendId, err.Error()))
		return
	}
	if result == nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=OpenSearchRecall\terr=empty result", context.RecommendId))
		return
	}

	if result.Status != "OK" {
		log.Error(fmt.Sprintf("requestId=%s\tevent=OpenSearchRecall\terr=opensearch invoke error(%v)", context.RecommendId, result.Errors))
		return

	}

	for _, osItem := range result.Result.Items {
		if itemId, ok := osItem.Fields[i.ItemId]; ok {
			properties := make(map[string]interface{})
			for k, v := range osItem.Fields {
				properties[k] = v
			}
			item := module.NewItemWithProperty(itemId, properties)
			item.RetrieveId = i.modelName
			if len(osItem.SortExprValues) > 0 {
				item.Score = utils.ToFloat(osItem.SortExprValues[0], 0)
			}

			ret = append(ret, item)
		}
	}

	log.Info(fmt.Sprintf("requestId=%s\tmodule=OpenSearchRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, i.modelName, len(ret), utils.CostTime(start)))
	return
}

func (i *OpenSearchRecall) getRequestParams(user *module.User, context *context.RecommendContext) (map[string]any, error) {
	requestParams := make(map[string]any, len(i.RequestParams))
	for k, v := range i.RequestParams {
		requestParams[k] = v
	}

	paramResults := make([]string, len(i.Params))
	for x, param := range i.Params {
		paramArr := strings.Split(param, ".")

		switch len(paramArr) {
		case 1:
			return requestParams, fmt.Errorf("Params(%s) type is error, its type should be a.b or a.b.c", param)
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
						return requestParams, errors.New("context.feature is null")
					}
				} else {
					return requestParams, fmt.Errorf("Params(%s) only support context.features.xxx", param)
				}
			} else {
				return requestParams, fmt.Errorf("Params(%s) only support context.features.xxx", param)
			}
		default:
			return requestParams, fmt.Errorf("Params(%s) type is error, its type should be a.b or a.b.c", param)
		}
	}

	for x, result := range paramResults {
		replaceStr := fmt.Sprintf("$%d", x+1) // query string 中是从 $1 开始
		for k, val := range requestParams {
			if str, ok := val.(string); ok {
				requestParams[k] = strings.Replace(str, replaceStr, result, 1)
			}
		}
	}
	return requestParams, nil
}
