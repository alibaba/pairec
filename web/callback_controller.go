package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aliyun/aliyun-pairec-config-go-sdk/model"
	"github.com/alibaba/pairec/abtest"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	plog "github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/service"
	"github.com/alibaba/pairec/utils"
)

type CallBackParam struct {
	SceneId     string                   `json:"scene_id"`
	RequestId   string                   `json:"request_id"`
	Uid         string                   `json:"uid"`
	Features    map[string]interface{}   `json:"features"`
	ItemList    []map[string]interface{} `json:"item_list"`
	RequestInfo map[string]interface{}   `json:"request_info"`
	Debug       bool                     `json:"debug"`
}

func (r *CallBackParam) GetParameter(name string) interface{} {
	if name == "uid" {
		return r.Uid
	} else if name == "scene" {
		return r.SceneId + "_callback"
	} else if name == "request_id" {
		return r.RequestId
	} else if name == "features" {
		return r.Features
	} else if name == "category" {
		return "default"
	}

	return nil
}

type CallBackResponse struct {
	Response
}

func (r *CallBackResponse) ToString() string {
	j, _ := json.Marshal(r)
	return string(j)
}

type CallBackController struct {
	Controller
	param   CallBackParam
	context *context.RecommendContext
}

func (c *CallBackController) Process(w http.ResponseWriter, r *http.Request) {
	c.Start = time.Now()
	var err error
	c.RequestBody, err = io.ReadAll(r.Body)
	if err != nil {
		c.dealResponse(w, r, ERROR_PARAMETER_CODE, "read parammeter error")
		return
	}
	if len(c.RequestBody) == 0 {
		c.dealResponse(w, r, ERROR_PARAMETER_CODE, "request body empty")
		return
	}
	if err := c.CheckParameter(); err != nil {
		c.dealResponse(w, r, ERROR_PARAMETER_CODE, err.Error())
		return
	}
	if c.param.RequestId != "" {
		c.RequestId = c.param.RequestId
	} else {
		c.RequestId = utils.UUID()
	}
	c.LogRequestBegin(r)
	c.doProcess(w, r)
	c.End = time.Now()
	c.LogRequestEnd(r)
}

func (r *CallBackController) CheckParameter() error {
	if err := json.Unmarshal(r.RequestBody, &r.param); err != nil {
		return err
	}

	if len(r.param.Uid) == 0 {
		return errors.New("uid not empty")
	}

	if len(r.param.RequestId) == 0 {
		return errors.New("request_id not empty")
	}

	if len(r.param.SceneId) == 0 {
		return errors.New("scene_id not empty")
	}

	if len(r.param.ItemList) == 0 {
		return errors.New("recommend item list not empty")
	}
	return nil
}

func (c *CallBackController) doProcess(w http.ResponseWriter, r *http.Request) {
	// write log async
	/**
	go func() {
		c.doCallbackLog()
	}()
	**/
	Send(c)
	c.dealResponse(w, r, SUCCESS_CODE, "success")
}

// doCallbackLog log user features and fg features(invoke eas if have)
func (c *CallBackController) doCallbackLog() {
	c.makeCallBackContext()

	userId := module.UID(c.param.Uid)
	user := module.NewUserWithContext(userId, c.context)
	callBackService := service.NewCallBackService()
	callBackService.User = user
	callBackService.LoadUserFeatures(c.context)

	var items []*module.Item
	for _, info := range c.param.ItemList {
		itemId := fmt.Sprintf("%v", info["item_id"])
		item := module.NewItem(itemId)
		items = append(items, item)

		for k, v := range info {
			if k == "item_id" {
				continue
			}
			item.AddProperty(k, v)
		}
	}
	// CallBackProcessFunc process
	if f, ok := callBackProcessFuncMap[c.param.SceneId]; ok {
		f(user, items, c.context)
	}

	// load characteristics
	callBackService.Items = items
	// invoke recall
	callBackService.Recommend(c.context)
	callBackService.LoadFeatures(c.context)

	// model rank
	callBackService.Rank(c.context)

	currTime := time.Now()
	log := make(map[string]interface{})
	requestInfo := make(map[string]interface{})
	log["request_id"] = c.param.RequestId
	log["scene"] = c.param.SceneId
	log["request_time"] = currTime.Unix()
	userFeaturesData, _ := json.Marshal(callBackService.User.MakeUserFeatures())
	log["user_features"] = string(userFeaturesData)
	log["user_id"] = string(callBackService.User.Id)
	for k, v := range c.param.RequestInfo {
		requestInfo[k] = v
	}
	requestInfoData, _ := json.Marshal(requestInfo)
	log["request_info"] = string(requestInfoData)

	// first write user log
	log["module"] = "user"
	if c.param.Debug {
		msg, _ := json.Marshal(log)
		info := fmt.Sprintf("requestId=%s\tmsg=%s", c.RequestId, string(msg))
		plog.Info(info)
	}

	if err := callBackService.RecordLogList(c.context, []map[string]interface{}{log}); err != nil {
		plog.Error(fmt.Sprintf("requestId=%s\tevent=RecordLogList\terror=%v", c.RequestId, err))
		return
	}

	// write item feature
	var messages []map[string]interface{}
	i := 0
	for _, item := range callBackService.Items {
		log := make(map[string]interface{})
		log["request_id"] = c.param.RequestId
		log["scene"] = c.param.SceneId
		log["request_time"] = currTime.Unix()
		log["module"] = "item"

		log["item_id"] = string(item.Id)
		log["user_id"] = string(callBackService.User.Id)
		log["raw_features"] = ""
		if str, ok := item.Properties["raw_features"]; ok {
			log["raw_features"] = str
		}
		//log["raw_features"] = item.StringProperty("raw_features")
		log["generate_features"] = ""
		if buf, ok := item.Properties["generate_features"]; ok && buf != nil {
			log["generate_features"] = buf.(*bytes.Buffer).String()
		}
		//log["generate_features"] = item.StringProperty("generate_features")
		//log["context_features"] = item.StringProperty("context_features")
		log["context_features"] = ""
		if str, ok := item.Properties["context_features"]; ok {
			log["context_features"] = str
		}
		//log["context_features"] = item.StringProperty("context_features")
		delete(item.Properties, "raw_features")
		delete(item.Properties, "generate_features")
		delete(item.Properties, "context_features")
		itemFeatures := item.GetFeatures()
		itemFeaturesData, _ := json.Marshal(itemFeatures)
		log["item_features"] = string(itemFeaturesData)

		if c.param.Debug {
			msg, _ := json.Marshal(log)
			info := fmt.Sprintf("requestId=%s\tmsg=%s", c.RequestId, string(msg))
			plog.Info(info)
		}

		messages = append(messages, log)
		i++
		if i%10 == 0 {
			if err := callBackService.RecordLogList(c.context, messages); err != nil {
				plog.Error(fmt.Sprintf("requestId=%s\tevent=RecordLogList\terror=%v", c.RequestId, err))
				return
			}
			messages = messages[:0]
		}
	}
	if len(messages) > 0 {
		if err := callBackService.RecordLogList(c.context, messages); err != nil {
			plog.Error(fmt.Sprintf("requestId=%s\tevent=RecordLogList\terror=%v", c.RequestId, err))
			return
		}
	}

}

func (c *CallBackController) dealResponse(w http.ResponseWriter, r *http.Request, code int, msg string) {
	response := CallBackResponse{
		Response: Response{
			RequestId: c.RequestId,
			Code:      code,
			Message:   msg,
		},
	}

	io.WriteString(w, response.ToString())
}

func (c *CallBackController) makeCallBackContext() {
	c.context = context.NewRecommendContext()
	c.context.Param = &c.param
	c.context.Config = recconf.Config
	c.context.RecommendId = c.RequestId
	c.context.Debug = c.param.Debug

	abcontext := model.ExperimentContext{
		Uid:          c.param.Uid,
		RequestId:    c.RequestId,
		FilterParams: map[string]interface{}{},
	}

	sceneId := c.param.SceneId + "_callback"
	if abtest.GetExperimentClient() != nil {
		c.context.ExperimentResult = abtest.GetExperimentClient().MatchExperiment(sceneId, &abcontext)
		log.Info(c.context.ExperimentResult.Info())
	}
}
