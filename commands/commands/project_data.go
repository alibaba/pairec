package commands

var makefileS = `BUILD=go build -mod vendor
DOCKER?=docker
SOURCE_DIR=src
BIN_NAME=${BINNAME}
REGISTRY?=registry.cn-beijing.cr.aliyuncs.com
DOCKER_TAG?=0.0.1
TEMP_DIR_SERVER:=$(shell mktemp -d)

.PHONY: setup build clean
setup:
	go mod vendor
build:
	cd ${SOURCE_DIR}; CGO_ENABLED=0 GOARCH=amd64 GOOS=linux ${BUILD} -o ${BIN_NAME} .
	cd ${SOURCE_DIR}; mv ${BIN_NAME} ../
release:
	cd ${SOURCE_DIR}; CGO_ENABLED=0 GOARCH=amd64 GOOS=linux ${BUILD} -o ${BIN_NAME} .
	cd ${SOURCE_DIR}; mv ${BIN_NAME} ${TEMP_DIR_SERVER}/appd
	cp docker/Dockerfile ${TEMP_DIR_SERVER}/
	cp conf/config.json.production ${TEMP_DIR_SERVER}/config.json
	cd ${TEMP_DIR_SERVER}  &&  ${DOCKER} build  -t ${REGISTRY}/${BIN_NAME}:${DOCKER_TAG} .
	${DOCKER} push ${REGISTRY}/${BIN_NAME}:${DOCKER_TAG}

clean:
	-rm -rf ${BIN_NAME}
`

var gomodS = `module ${BINNAME} 

go 1.19

require (
	github.com/alibaba/pairec v1.0.1
	github.com/aliyun/aliyun-pairec-config-go-sdk v1.0.1
)
`

var confS = `{
	"RunMode": "product",
	"ListenConf": {
	  "HttpAddr": "",
	  "HttpPort": 8000
	},
	"ABTestConf": {
	  "Host": "",
	  "Token": ""
	},
	"FilterConfs": [
	],
	"RecallConfs": [
		{
		  "Name":"mock_recall",
          "RecallType":"MockRecall",
          "RecallCount":200
	    }
	],
	"SortNames": {
	  "default": [
		  "ItemRankScore"
	  ]
	},
	"FilterNames": {
	  "default": [
		"UniqueFilter"
	  ]
	},
	"AlgoConfs": [
	],
	"HologresConfs": {
	},
	"KafkaConfs": {
	},
	"RedisConfs": {
	},
	"SceneConfs": {
	    "home_feed":{
			"default":{
				"RecallNames":["mock_recall"]
			}
		}
	},
	"LogConf": {
	  "RetensionDays": 3,
	  "DiskSize": 20,
	  "LogLevel": "INFO"
	},
	"RankConf": {
	},
	"FeatureConfs": {
	}
 }
`

var mainfileS = `package main

import (
	"${BINNAME}/src/controller"
	"github.com/alibaba/pairec"
)

func main() {
	pairec.AddStartHook(func() error {
		return nil
	})

	pairec.Route("/api/rec/feed", &controller.FeedController{})
	pairec.Run()
}
`

var dockerfileS = `FROM amd64/centos
LABEL pro="recommend"

RUN mkdir -p /go/config && \
    mkdir /go/log
ADD appd /go/bin/
ADD config.json /go/config/

RUN  yes | cp  -f /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime

RUN echo -e '#!/bin/sh \n\n\
/go/bin/appd --config=/go/config/config.json \
--log_dir=/go/log --alsologtostderr \
"$@"' > /usr/bin/rec_entrypoint.sh \
&& chmod +x /usr/bin/rec_entrypoint.sh

ENTRYPOINT ["/usr/bin/rec_entrypoint.sh"]
`

var controllerfileS = `package controller

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
	"github.com/alibaba/pairec/web"
)

type RecommendParam struct {
` + "    SceneId  string                 `json:\"scene_id\"`\n" +
	"    Category string                 `json:\"category\"`\n" +
	"    Uid      string                 `json:\"uid\"`  // user id \n" +
	"    Size     int                    `json:\"size\"` // get recommend items size \n" +
	"    Debug    bool                   `json:\"debug\"` \n" +
	"    Features map[string]interface{} `json:\"features\"` \n" +
	"    ItemId   string                 `json:\"item_id\"` // similarity recommendation itemid\n" +
	`}
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
	} else if name == "item_id" {
		return r.ItemId
	}

	return nil
}

type RecommendResponse struct {
	web.Response
` +
	"    Size  int             `json:\"size\"` \n" +
	"    ExperimentId  string  `json:\"experiment_id\"` \n" +
	"    Items []*ItemData     `json:\"items\"` \n" +
	`}
type ItemData struct {
` +
	"    ItemId     string `json:\"item_id\"` \n" +
	"    Score      float64 `json:\"score\"` \n" +
	"    RetrieveId string `json:\"retrieve_id\"` \n" +
	`}

func (r *RecommendResponse) ToString() string {
	j, _ := json.Marshal(r)
	return string(j)
}

type FeedController struct {
	web.Controller
	param   RecommendParam
	context *context.RecommendContext
}

func (c *FeedController) Process(w http.ResponseWriter, r *http.Request) {
	c.Start = time.Now()
	var err error
	c.RequestBody, err = ioutil.ReadAll(r.Body)
	if err != nil {
		c.SendError(w, web.ERROR_PARAMETER_CODE, "read parammeter error")
		return
	}
	if len(c.RequestBody) == 0 {
		c.SendError(w, web.ERROR_PARAMETER_CODE, "request body empty")
		return
	}
	c.RequestId = utils.UUID()
	c.LogRequestBegin(r)
	if err := c.CheckParameter(); err != nil {
		c.SendError(w, web.ERROR_PARAMETER_CODE, err.Error())
		return
	}
	c.doProcess(w, r)
	c.End = time.Now()
	c.LogRequestEnd(r)
}
func (r *FeedController) CheckParameter() error {
	if err := json.Unmarshal(r.RequestBody, &r.param); err != nil {
		return err
	}

	if len(r.param.Uid) == 0 {
		return errors.New("uid not empty")
	}
	if r.param.Size <= 0 {
		r.param.Size = 10
	}
	if r.param.SceneId == "" {
		r.param.SceneId = "default_scene"
	}
	if r.param.Category == "" {
		r.param.Category = "default"
	}

	return nil
}
func (c *FeedController) doProcess(w http.ResponseWriter, r *http.Request) {
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
			Score:   item.Score,
			RetrieveId: item.RetrieveId,
		}

		data = append(data, idata)
	}

	expId := ""
	if c.context.ExperimentResult != nil {
		expId = c.context.ExperimentResult.GetExpId()
	}

	if len(data) < c.param.Size {
		response := RecommendResponse{
			Size:  len(data),
			ExperimentId: expId,
			Items: data,
			Response: web.Response{
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
		ExperimentId: expId,
		Items: data,
		Response: web.Response{
			RequestId: c.RequestId,
			Code:      200,
			Message:   "success",
		},
	}
	io.WriteString(w, response.ToString())
}
func (c *FeedController) makeRecommendContext() {
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
`
