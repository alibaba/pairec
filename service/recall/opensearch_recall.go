package recall

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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

type OpenSearchResult struct {
	Body struct {
		Status    string `json:"status"`
		RequestId string `json:"request_id"`
		Errors    []any  `json:"errors"`
		Result    struct {
			Items []struct {
				Fields map[string]string `json:"fields"`
			} `json:"items"`
		} `json:"result"`
	} `json:"body"`
}
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
			log.Info(fmt.Sprintf("requestId=%s\tmodule=OpenSearchRecall\tfrom=cache\tcount=%d\tcost=%d", context.RecommendId, len(ret), utils.CostTime(start)))
			return
		}
	}
	requestParams, err := i.getRequestParams(user, context)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=OpenSearchRecall\terr=%s", context.RecommendId, err.Error()))
		return
	}
	//log query_string
	if context.Debug {
		log.Info(fmt.Sprintf("event=OpenSearchRecall\trequest_params=%v", requestParams))
	}

	resp, err := i.openSearchClient.OpenSearchClient.Request(tea.String("GET"), tea.String("/v3/openapi/apps/"+i.AppName+"/search"), requestParams, nil, nil, i.openSearchClient.Runtime)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=OpenSearchRecall\terr=%s", context.RecommendId, err.Error()))
		return
	}
	j, _ := json.Marshal(resp)

	result := OpenSearchResult{}
	err = json.Unmarshal(j, &result)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=OpenSearchRecall\terr=%s", context.RecommendId, err.Error()))
		return
	}
	if result.Body.Status != "OK" {
		log.Error(fmt.Sprintf("requestId=%s\tevent=OpenSearchRecall\terr=opensearch invoke error(%v)", context.RecommendId, result.Body.Errors))
		return

	}

	for _, item := range result.Body.Result.Items {
		if itemId, ok := item.Fields[i.ItemId]; ok {
			item := module.NewItem(itemId)
			item.RetrieveId = i.modelName

			ret = append(ret, item)
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
				log.Error(fmt.Sprintf("requestId=%s\tmodule=OpenSearchRecall\terror=%v", context.RecommendId, err))
			}
		}()
	}

	log.Info(fmt.Sprintf("requestId=%s\tmodule=OpenSearchRecall\tcount=%d\tcost=%d", context.RecommendId, len(ret), utils.CostTime(start)))
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
