package sls

import (
	"fmt"
	"time"

	alisls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/producer"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type SlsClient struct {
	accessKeyId     string
	accessKeySecret string
	endPoint        string
	projectName     string
	logstoreName    string
	clientIP        string

	Client   alisls.ClientInterface
	Producer *producer.Producer

	callback *Callback
}

var slsclientInstances = make(map[string]*SlsClient)

func GetSlsClient(name string) (*SlsClient, error) {

	client, ok := slsclientInstances[name]
	if !ok {
		return nil, fmt.Errorf("slsclient not found, name:%s", name)
	}
	return client, nil

}
func NewSlsClient(accessKeyId, accessKeySecret, endpoint, projectName, logstoreName string) *SlsClient {
	s := &SlsClient{
		accessKeyId:     accessKeyId,
		accessKeySecret: accessKeySecret,
		endPoint:        endpoint,
		projectName:     projectName,
		logstoreName:    logstoreName,
	}

	return s
}
func (s *SlsClient) Init() {
	s.Client = alisls.CreateNormalInterface(s.endPoint, s.accessKeyId, s.accessKeySecret, "")

	producerConfig := producer.GetDefaultProducerConfig()
	producerConfig.Endpoint = s.endPoint
	producerConfig.AccessKeyID = s.accessKeyId
	producerConfig.AccessKeySecret = s.accessKeySecret
	producerConfig.MaxIoWorkerCount = 8
	producerConfig.MaxBatchSize = 4 * 1024 * 1024
	s.Producer = producer.InitProducer(producerConfig)
	s.Producer.Start()
	s.callback = &Callback{}

	if clientIP, err := utils.GetClientIp(); err == nil {
		s.clientIP = clientIP
	}
}

func (s *SlsClient) SendLog(data map[string]string) {
	log := producer.GenerateLog(uint32(time.Now().Unix()), data)
	err := s.Producer.SendLogWithCallBack(s.projectName, s.logstoreName, "", s.clientIP, log, s.callback)
	if err != nil {
		fmt.Println(err)
	}
}
func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.SlsConfs {
		if _, ok := slsclientInstances[name]; ok {
			continue
		}
		s := NewSlsClient(conf.AccessKeyId, conf.AccessKeySecret, conf.EndPoint, conf.ProjectName, conf.LogstoreName)
		s.Init()
		slsclientInstances[name] = s
	}
}

type Callback struct {
}

func (callback *Callback) Success(result *producer.Result) {
}

func (callback *Callback) Fail(result *producer.Result) {
	fmt.Printf("requestId=%s\tmodule=sls\tmessage=%s\tcode=%s", result.GetRequestId(), result.GetErrorMessage(), result.GetErrorCode())
}
