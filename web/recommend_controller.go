package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aliyun/aliyun-pairec-config-go-sdk/model"
	"github.com/alibaba/pairec/abtest"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/service"
	"github.com/alibaba/pairec/utils"
)

const (
	Default_Size int = 10
)

type RecommendParam struct {
	SceneId  string                 `json:"scene_id"`
	Category string                 `json:"category"`
	Uid      string                 `json:"uid"`  // user id
	Size     int                    `json:"size"` // get recommend items size
	Debug    bool                   `json:"debug"`
	Features map[string]interface{} `json:"features"`
}

func (r *RecommendParam) GetParameter(name string) interface{} {
	if name == "uid" {
		return r.Uid
	} else if name == "scene" {
		return r.SceneId
	} else if name == "category" {
		if r.Category != "" {
			return r.Category
		}
		return "default"
	} else if name == "features" {
		return r.Features
	}

	return nil
}

type RecommendResponse struct {
	Response
	Size  int         `json:"size"`
	Items []*ItemData `json:"items"`
}
type ItemData struct {
	ItemId     string `json:"item_id"`
	ItemType   string `json:"item_type"`
	RetrieveId string `json:"retrieve_id"`
}

func (r *RecommendResponse) ToString() string {
	j, _ := json.Marshal(r)
	return string(j)
}

type RecommendController struct {
	Controller
	param   RecommendParam
	context *context.RecommendContext
}

func (c *RecommendController) Process(w http.ResponseWriter, r *http.Request) {
	c.Start = time.Now()
	var err error
	c.RequestBody, err = ioutil.ReadAll(r.Body)
	if err != nil {
		c.SendError(w, ERROR_PARAMETER_CODE, "read parammeter error")
		return
	}
	if len(c.RequestBody) == 0 {
		c.SendError(w, ERROR_PARAMETER_CODE, "request body empty")
		return
	}
	c.RequestId = utils.UUID()
	c.LogRequestBegin(r)
	if err := c.CheckParameter(); err != nil {
		c.SendError(w, ERROR_PARAMETER_CODE, err.Error())
		return
	}
	c.doProcess(w, r)
	c.End = time.Now()
	c.LogRequestEnd(r)
}
func (r *RecommendController) CheckParameter() error {
	if err := json.Unmarshal(r.RequestBody, &r.param); err != nil {
		return err
	}

	if len(r.param.Uid) == 0 {
		return errors.New("uid not empty")
	}
	if r.param.Size <= 0 {
		r.param.Size = Default_Size
	}
	if r.param.SceneId == "" {
		r.param.SceneId = "default_scene"
	}
	if r.param.Category == "" {
		r.param.Category = "default"
	}

	return nil
}
func (c *RecommendController) doProcess(w http.ResponseWriter, r *http.Request) {
	c.makeRecommendContext()
	userRecommendService := service.NewUserRecommendService()
	items := userRecommendService.Recommend(c.context)
	data := make([]*ItemData, 0)
	for _, item := range items {
		if c.param.Debug {
			fmt.Println(item)
		}

		idata := &ItemData{
			ItemId:     string(item.Id),
			ItemType:   item.ItemType,
			RetrieveId: item.RetrieveId,
		}

		data = append(data, idata)
	}

	if len(data) < c.param.Size {
		response := RecommendResponse{
			Size:  len(data),
			Items: data,
			Response: Response{
				RequestId: c.RequestId,
				Code:      299,
				Message:   "items size not enough",
			},
		}
		io.WriteString(w, response.ToString())
		return
	}

	response := RecommendResponse{
		Size:  len(data),
		Items: data,
		Response: Response{
			RequestId: c.RequestId,
			Code:      200,
			Message:   "success",
		},
	}
	io.WriteString(w, response.ToString())
}
func (c *RecommendController) makeRecommendContext() {
	c.context = context.NewRecommendContext()
	c.context.Size = c.param.Size
	c.context.Debug = c.param.Debug
	c.context.Param = &c.param
	c.context.RecommendId = c.RequestId
	c.context.Config = recconf.Config

	abcontext := model.ExperimentContext{
		Uid:          c.param.Uid,
		RequestId:    c.RequestId,
		FilterParams: map[string]interface{}{},
	}

	if abtest.GetExperimentClient() != nil {
		c.context.ExperimentResult = abtest.GetExperimentClient().MatchExperiment(c.param.SceneId, &abcontext)
		log.Info(c.context.ExperimentResult.Info())
	}
}
