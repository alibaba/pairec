package abtest

import (
	"os"

	"github.com/alibaba/pairec/v2/log"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/experiments"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
)

var experimentClient *experiments.ExperimentClient

// Load abtest config from  config instance
/**
func Load(config *recconf.RecommendConfig) {
	if experimentClient == nil && config.ABTestConf.Host != "" {
		env := config.RunMode
		if os.Getenv("PAIREC_ENVIRONMENT") != "" {
			env = os.Getenv("PAIREC_ENVIRONMENT")
		}

		l := log.ABTestLogger{}
		client, err := experiments.NewExperimentClient(config.ABTestConf.Host, env,
			experiments.WithLogger(experiments.LoggerFunc(l.Infof)),
			experiments.WithErrorLogger(experiments.LoggerFunc(l.Errorf)),
			experiments.WithToken(config.ABTestConf.Token),
		)

		if err != nil {
			panic(err)
		}

		experimentClient = client
	}
}
**/

// LoadFromEnvironment create abtest instance use env, env list:
//
//	PAIREC_ENVIRONMENT is the environment type, valid values are: daily, prepub, product
//	ABTEST_HOST abtest host address
//	ABTEST_TOKEN abtest token, if abtest server deploy on eas , must set it
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
		panic("env ACCESS_ID empty")
	}
	if accessSecret == "" {
		panic("env ACCESS_SECRET empty")
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
