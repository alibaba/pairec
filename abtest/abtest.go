package abtest

import (
	"os"

	"github.com/alibaba/pairec/v2/log"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/experiments"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
)

var experimentClient *experiments.ExperimentClient

// LoadFromEnvironment create abtest instance use env, env list:
//
// ENV params list:
//
//		PAIREC_ENVIRONMENT is the environment type, valid values are: daily, prepub, product
//		REGION region of pairec console instance, like cn-beijing,cn-hangzhou
//		INSTANCE_ID id of pairec console instance
//	    AccessKey  aliyun accessKeyId
//	    AccessSecret  aliyun accessKeySecret
func LoadFromEnvironment() {
	env := os.Getenv("PAIREC_ENVIRONMENT")
	if env == "" {
		panic("env PAIREC_ENVIRONMENT empty")
	}

	region := os.Getenv("REGION")
	instanceId := os.Getenv("INSTANCE_ID")
	accessId := os.Getenv("AccessKey")
	accessSecret := os.Getenv("AccessSecret")
	if region == "" {
		panic("env REGION empty")
	}
	if instanceId == "" {
		panic("env INSTANCE_ID empty")
	}
	if accessId == "" {
		panic("env AccessKey empty")
	}
	if accessSecret == "" {
		panic("env AccessSecret empty")
	}

	l := log.ABTestLogger{}
	client, err := experiments.NewExperimentClient(instanceId, region, accessId, accessSecret, env,
		experiments.WithLogger(experiments.LoggerFunc(l.Infof)),
		experiments.WithErrorLogger(experiments.LoggerFunc(l.Errorf)),
	)

	if err != nil {
		panic(err)
	}

	experimentClient = client
}
func GetExperimentClient() *experiments.ExperimentClient {
	return experimentClient
}

func GetParams(sceneName string) model.SceneParams {
	return experimentClient.GetSceneParams(sceneName)
}
