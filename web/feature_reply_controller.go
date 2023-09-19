package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/model"
	"github.com/alibaba/pairec/v2/abtest"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service"
)

type DataSourceInfo struct {
	Type       string `json:"type"`
	AccessId   string `json:"access_id"`
	AccessKey  string `json:"access_key"`
	Endpoint   string `json:"endpoint"`
	Project    string `json:"project"`
	Topic      string `json:"topic"`
	VpcAddress string `json:"vpc_address"`
	Token      string `json:"token"`
	JobId      uint32 `json:"job_id"`
}
type FeatureReplyData struct {
	Scene        string   `json:"scene"`
	RequestId    string   `json:"request_id"`
	UserId       string   `json:"user_id"` // user id
	UserFeatures string   `json:"user_features"`
	ItemIds      []string `json:"item_ids"` // item ids
	ItemFeatures []string `json:"item_features"`
}
type FeatureReplyParam struct {
	DataSource  DataSourceInfo   `json:"datasource"`
	FeatureData FeatureReplyData `json:"items"`
}

func (r *FeatureReplyParam) GetParameter(name string) interface{} {
	if name == "uid" {
		return r.FeatureData.UserId
	} else if name == "scene" {
		return r.FeatureData.Scene
	} else if name == "datasource_type" {
		return r.DataSource.Type
	} else if name == "access_id" {
		return r.DataSource.AccessId
	} else if name == "access_key" {
		return r.DataSource.AccessKey
	} else if name == "endpoint" {
		return r.DataSource.Endpoint
	} else if name == "project" {
		return r.DataSource.Project
	} else if name == "topic" {
		return r.DataSource.Topic
	} else if name == "token" {
		return r.DataSource.Token
	} else if name == "vpc_address" {
		return r.DataSource.VpcAddress
	} else if name == "job_id" {
		return r.DataSource.JobId
	}

	return nil
}

type FeatureReplyResponse struct {
	Response
}

func (r *FeatureReplyResponse) ToString() string {
	j, _ := json.Marshal(r)
	return string(j)
}

type FeatureReplyController struct {
	Controller
	param   FeatureReplyParam
	context *context.RecommendContext
}

func (c *FeatureReplyController) Process(w http.ResponseWriter, r *http.Request) {
	c.Start = time.Now()
	var err error
	c.RequestBody, err = io.ReadAll(r.Body)
	if err != nil {
		c.SendError(w, ERROR_PARAMETER_CODE, "read parammeter error")
		return
	}
	if len(c.RequestBody) == 0 {
		c.SendError(w, ERROR_PARAMETER_CODE, "request body empty")
		return
	}
	if err := c.CheckParameter(); err != nil {
		c.SendError(w, ERROR_PARAMETER_CODE, err.Error())
		return
	}
	c.LogRequestBegin(r)
	c.doProcess(w, r)
	c.End = time.Now()
	c.LogRequestEnd(r)
}
func (r *FeatureReplyController) CheckParameter() error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(r.RequestBody, &r.param); err != nil {
		return err
	}

	if len(r.param.FeatureData.UserId) == 0 {
		return errors.New("user_id not empty")
	}
	if len(r.param.FeatureData.ItemIds) == 0 {
		return errors.New("item_ids not empty")
	}
	if len(r.param.FeatureData.ItemIds) != len(r.param.FeatureData.ItemFeatures) {
		return errors.New("item_ids length is not equal item_features")
	}

	if r.param.FeatureData.RequestId == "" {
		return errors.New("request_id not empty")
	}

	r.RequestId = r.param.FeatureData.RequestId
	return nil
}
func (c *FeatureReplyController) doProcess(w http.ResponseWriter, r *http.Request) {
	c.makeRecommendContext()
	featureReplyService := service.NewFeatureReplyService()

	featureReplyService.FeatureReply(c.param.FeatureData.UserFeatures, c.param.FeatureData.ItemFeatures,
		c.param.FeatureData.ItemIds, c.context)

	response := FeatureReplyResponse{
		Response: Response{
			RequestId: c.RequestId,
			Code:      200,
			Message:   "success",
		},
	}

	io.WriteString(w, response.ToString())
}

func (c *FeatureReplyController) makeRecommendContext() {
	c.context = context.NewRecommendContext()
	c.context.Param = &c.param
	c.context.RecommendId = c.RequestId
	c.context.Config = recconf.Config

	abcontext := model.ExperimentContext{
		Uid:          c.param.FeatureData.UserId,
		RequestId:    c.RequestId,
		FilterParams: map[string]interface{}{},
	}

	if abtest.GetExperimentClient() != nil {
		c.context.ExperimentResult = abtest.GetExperimentClient().MatchExperiment(c.param.FeatureData.Scene, &abcontext)
		log.Info(c.context.ExperimentResult.Info())
	}
}
