package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service"
	"github.com/alibaba/pairec/v2/utils"
)

type EmbeddingParam struct {
	RequestId    string                 `json:"request_id"`
	SceneId      string                 `json:"scene_id"`
	Uid          string                 `json:"uid"` // user id
	Debug        bool                   `json:"debug"`
	UserFeatures map[string]interface{} `json:"user_features"`
	ItemId       string                 `json:"item_id"`
	ItemFeatures map[string]interface{} `json:"item_features"`
}

func (r *EmbeddingParam) GetParameter(name string) interface{} {
	if name == "uid" {
		return r.Uid
	} else if name == "scene" {
		return r.SceneId
	} else if name == "category" {
		return "default"
	} else if name == "user_features" {
		return r.UserFeatures
	} else if name == "item_id" {
		return r.ItemId
	} else if name == "item_features" {
		return r.ItemFeatures
	}

	return nil
}

type EmbeddingResponse struct {
	Response
	Embedding []float32 `json:"embedding"`
}

func (r *EmbeddingResponse) ToString() string {
	j, _ := json.Marshal(r)
	return string(j)
}
func (r *EmbeddingResponse) ToBytes() []byte {
	j, _ := json.Marshal(r)
	return j
}

type EmbeddingController struct {
	Controller
	param   EmbeddingParam
	context *context.RecommendContext
}

func (c *EmbeddingController) Process(w http.ResponseWriter, r *http.Request) {
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
	if err := c.decodeRequestBody(); err != nil {
		c.SendError(w, ERROR_PARAMETER_CODE, err.Error())
		return
	}

	c.LogRequestBegin(r)
	if err := c.CheckParameter(); err != nil {
		c.SendError(w, ERROR_PARAMETER_CODE, err.Error())
		return
	}
	c.doProcess(w, r)
	c.End = time.Now()
	c.LogRequestEnd(r)
}
func (r *EmbeddingController) decodeRequestBody() error {
	if err := json.Unmarshal(r.RequestBody, &r.param); err != nil {
		return err
	}
	if r.param.RequestId == "" {
		r.RequestId = utils.UUID()
	} else {
		r.RequestId = r.param.RequestId
	}
	return nil
}
func (r *EmbeddingController) CheckParameter() error {

	if r.param.Uid == "" && r.param.ItemId == "" {
		return errors.New("uid or item_id not empty")
	}
	if r.param.SceneId == "" {
		return errors.New("scene_id not empty")
	}

	return nil
}
func (c *EmbeddingController) doProcess(w http.ResponseWriter, r *http.Request) {
	c.makeRecommendContext()
	embeddingService := service.NewEmbeddingService()
	embeddings, err := embeddingService.Recommend(c.context)

	if c.param.Debug {
		fmt.Printf("requestId=%s\tembeddings=%v\terror=%v\n", c.RequestId, embeddings, err)
	}
	if err != nil {
		response := EmbeddingResponse{
			Response: Response{
				RequestId: c.RequestId,
				Code:      500,
				Message:   err.Error(),
			},
		}
		w.Write(response.ToBytes())
		return
	}

	response := EmbeddingResponse{
		Embedding: embeddings,
		Response: Response{
			RequestId: c.RequestId,
			Code:      200,
			Message:   "success",
		},
	}
	w.Write(response.ToBytes())
}

func (c *EmbeddingController) makeRecommendContext() {
	c.context = context.NewRecommendContext()
	c.context.Debug = c.param.Debug
	c.context.Param = &c.param
	c.context.RecommendId = c.RequestId
	c.context.Config = recconf.Config

}
