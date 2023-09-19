package web

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service"
	"github.com/alibaba/pairec/v2/utils"
)

type UserRecallParam struct {
	SceneId  string `json:"scene_id"`
	Category string `json:"category"`
	Uid      string `json:"uid"`  // user id
	Size     int    `json:"size"` // get recommend items size
}

func (r *UserRecallParam) GetParameter(name string) interface{} {
	if name == "uid" {
		return r.Uid
	} else if name == "scene" {
		return r.SceneId
	} else if name == "category" {
		return r.Category
	}

	return nil
}

type UserRecallResponse struct {
	Response
	Size   int                   `json:"size"`
	Items  []*UserRecallItemData `json:"items"`
	Errors []error               `json:"errors"`
}
type UserRecallItemData struct {
	ItemId string `json:"item_id"`
	// ItemType   string `json:"item_type"`
	RetrieveId string `json:"retrieve_id"`
}

func (r *UserRecallResponse) ToString() string {
	j, _ := json.Marshal(r)
	return string(j)
}

type UserRecallController struct {
	Controller
	param   UserRecallParam
	context *context.RecommendContext
}

func (c *UserRecallController) Process(w http.ResponseWriter, r *http.Request) {
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
func (r *UserRecallController) CheckParameter() error {
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
		r.param.Category = "default_category"
	}

	return nil
}
func (c *UserRecallController) doProcess(w http.ResponseWriter, r *http.Request) {
	c.makeRecommendContext()
	userRecallService := service.NewUserRecallService()
	items := userRecallService.Recommend(c.context)
	data := make([]*UserRecallItemData, 0)
	for _, item := range items {
		idata := &UserRecallItemData{
			ItemId: string(item.Id),
			// ItemType:   item.ItemType,
			RetrieveId: item.RetrieveId,
		}

		data = append(data, idata)
	}

	errs := make([]error, 0)
	if len(data) < c.param.Size {
		response := UserRecallResponse{
			Size:   len(data),
			Items:  data,
			Errors: errs,
			Response: Response{
				RequestId: c.RequestId,
				Code:      299,
				Message:   "items size not enough",
			},
		}
		io.WriteString(w, response.ToString())
		return
	}

	response := UserRecallResponse{
		Size:   len(data),
		Items:  data,
		Errors: errs,
		Response: Response{
			RequestId: c.RequestId,
			Code:      200,
			Message:   "success",
		},
	}
	io.WriteString(w, response.ToString())
}
func (c *UserRecallController) makeRecommendContext() {
	c.context = context.NewRecommendContext()
	c.context.Size = c.param.Size
	c.context.Param = &c.param
	c.context.RecommendId = c.RequestId
	c.context.Config = recconf.Config
}
