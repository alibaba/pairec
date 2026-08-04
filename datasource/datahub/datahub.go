package datahub

import (
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/hook"
	"github.com/alibaba/pairec/v2/utils/synclog"
	alidatahub "github.com/aliyun/aliyun-datahub-sdk-go/datahub"
)

type Datahub struct {
	accessId     string
	accessKey    string
	endpoint     string
	projectName  string
	topicName    string
	schemas      []recconf.DatahubTopicSchema
	datahubApi   alidatahub.DataHubApi
	producer     alidatahub.Producer
	recordSchema *alidatahub.RecordSchema

	name           string
	syncLog        *synclog.SyncLog
	compressorType string
}

var (
	mu               sync.RWMutex
	datahubInstances = make(map[string]*Datahub)
)

func GetDatahub(name string) (*Datahub, error) {
	mu.RLock()
	defer mu.RUnlock()
	if _, ok := datahubInstances[name]; !ok {
		return nil, fmt.Errorf("Datahub not found, name:%s", name)
	}

	return datahubInstances[name], nil
}
func RegisterDatahub(name string, dh *Datahub) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := datahubInstances[name]; !ok {
		datahubInstances[name] = dh
		dh.name = name
	}
}

func RemoveDatahub(name string) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := datahubInstances[name]; ok {
		datahubInstances[name].StopLoopListShards()
		delete(datahubInstances, name)
	}
}

func NewDatahub(accessId, accessKey, endpoint, project, topic, compressorType string, schemas []recconf.DatahubTopicSchema) *Datahub {
	p := &Datahub{
		accessId:       accessId,
		accessKey:      accessKey,
		endpoint:       endpoint,
		projectName:    project,
		topicName:      topic,
		schemas:        schemas,
		compressorType: compressorType,
	}
	return p
}

func (d *Datahub) Init() error {
	var account alidatahub.Account
	var err error
	if d.accessId == "" || d.accessKey == "" {
		account, err = NewAklessAccount()
		if err != nil {
			return err
		}
	} else {
		account = alidatahub.NewAliyunAccount(d.accessId, d.accessKey)
	}
	config := alidatahub.NewDefaultConfig()
	switch d.compressorType {
	case "lz4":
		config.CompressorType = alidatahub.LZ4
	case "zstd":
		config.CompressorType = alidatahub.ZSTD
	case "deflate":
		config.CompressorType = alidatahub.DEFLATE
	case "none":
		config.CompressorType = alidatahub.NOCOMPRESS
	default:
		config.CompressorType = alidatahub.LZ4
	}
	config.HttpClient = alidatahub.DefaultHttpClient()
	dh := alidatahub.NewClientWithConfig(d.endpoint, config, account)
	d.datahubApi = dh
	if len(d.schemas) > 0 {
		if err := d.createTopic(); err != nil {
			time.Sleep(2 * time.Second)
			if err = d.createTopic(); err != nil {
				return err
			}
		}
	}

	// use producer api to send records, shard list, retry and
	// compress type negotiation are managed by producer internally
	producerConfig := alidatahub.NewProducerConfig()
	producerConfig.Account = account
	producerConfig.Endpoint = d.endpoint
	producerConfig.Project = d.projectName
	producerConfig.Topic = d.topicName
	producer := alidatahub.NewProducer(producerConfig)
	if err = producer.Init(); err != nil {
		return err
	}
	d.producer = producer

	// cache the server side schema as fallback when GetSchema failed at runtime
	schema, err := producer.GetSchema()
	if err != nil {
		return err
	}
	d.recordSchema = schema

	dir := fmt.Sprintf("./tmp/%s/%s", d.projectName, d.topicName)
	synclog := synclog.NewSyncLog(dir, d.consumeSyncLog)
	if err := synclog.Init(); err != nil {
		log.Error(fmt.Sprintf("project=%s\ttopic=%s\terror=init sync log error(%v)", d.projectName, d.topicName, err))
		return err
	}

	d.syncLog = synclog

	return nil
}
func (d *Datahub) createTopic() error {
	getTopicResult, err := d.datahubApi.GetTopic(d.projectName, d.topicName)
	if err != nil {
		recordSchema := alidatahub.NewRecordSchema()
		for _, schema := range d.schemas {
			switch schema.Type {
			case "string":
				recordSchema.AddField(alidatahub.Field{Name: schema.Field, Type: alidatahub.STRING, AllowNull: true})
			case "integer":
				recordSchema.AddField(alidatahub.Field{Name: schema.Field, Type: alidatahub.INTEGER, AllowNull: true})
			case "bigint":
				recordSchema.AddField(alidatahub.Field{Name: schema.Field, Type: alidatahub.BIGINT, AllowNull: true})
			case "double":
				recordSchema.AddField(alidatahub.Field{Name: schema.Field, Type: alidatahub.DOUBLE, AllowNull: true})
			case "float":
				recordSchema.AddField(alidatahub.Field{Name: schema.Field, Type: alidatahub.FLOAT, AllowNull: true})
			case "timestamp":
				recordSchema.AddField(alidatahub.Field{Name: schema.Field, Type: alidatahub.TIMESTAMP, AllowNull: true})
			}
		}
		if _, err := d.datahubApi.CreateTupleTopic(d.projectName, d.topicName, fmt.Sprintf("create topic %s by pairec", d.topicName), 6, 3, recordSchema); err != nil {
			return err
		}
		d.recordSchema = recordSchema
	} else {
		d.recordSchema = getTopicResult.RecordSchema
	}
	return nil
}
func (d *Datahub) DataHubApi() alidatahub.DataHubApi {
	return d.datahubApi
}

func (d *Datahub) Shards() (ret []string) {
	if d.producer != nil {
		ret = d.producer.GetActiveShards()
	}
	return
}

func (d *Datahub) Stop() {
	d.StopLoopListShards()
	RemoveDatahub(d.name)
}

// StopLoopListShards is kept for compatibility, shard list is managed by producer now
func (d *Datahub) StopLoopListShards() {
	if d.producer != nil {
		d.producer.Close()
	}
}

// getRecordSchema returns the latest topic schema, fall back to the schema cached at Init when fetch failed
func (d *Datahub) getRecordSchema() *alidatahub.RecordSchema {
	if d.producer == nil {
		return d.recordSchema
	}
	if schema, err := d.producer.GetSchema(); err == nil {
		return schema
	} else {
		log.Warning(fmt.Sprintf("project=%s\ttopic=%s\tmsg=get schema failed(%v)", d.projectName, d.topicName, err))
	}
	return d.recordSchema
}

// writeSyncLog persists messages to the local sync log for later replay
func (d *Datahub) writeSyncLog(messages []map[string]interface{}) {
	if d.syncLog == nil {
		log.Error(fmt.Sprintf("project=%s\ttopic=%s\tmsg=sync log not initialized, drop %d messages", d.projectName, d.topicName, len(messages)))
		return
	}
	for _, msg := range messages {
		if err := d.syncLog.Write(NewSyncLogDatahubItem(msg)); err != nil {
			log.Error(fmt.Sprintf("project=%s\ttopic=%s\tmsg=write sync log failed(%v)", d.projectName, d.topicName, err))
		}
	}
}

func (d *Datahub) SendMessage(messages []map[string]interface{}) {
	if len(messages) == 0 {
		return
	}
	// Init may have failed and left the instance half initialized, degrade instead of panic
	if d.producer == nil {
		log.Error(fmt.Sprintf("project=%s\ttopic=%s\tmsg=producer not initialized, drop %d messages", d.projectName, d.topicName, len(messages)))
		return
	}
	records := make([]alidatahub.IRecord, 0, len(messages))
	schema := d.getRecordSchema()
	for _, messsage := range messages {
		record := alidatahub.NewTupleRecord(schema)
		for k, v := range messsage {
			if err := record.SetValueByName(k, v); err != nil {
				log.Warning(fmt.Sprintf("project=%s\ttopic=%s\tfield=%s\tmsg=set record value failed(%v)", d.projectName, d.topicName, k, err))
			}
		}

		records = append(records, record)
	}

	// producer picks an active shard and retries retryable errors(network/limit exceeded) internally
	if _, err := d.producer.Send(records); err != nil {
		log.Warning(fmt.Sprintf("project=%s\ttopic=%s\tmsg=put record failed(%v)", d.projectName, d.topicName, err))
		d.writeSyncLog(messages)
	}
}
func (d *Datahub) consumeSyncLog(data []byte) error {
	datahubItem := NewSyncLogDatahubItem(nil)
	if err := datahubItem.Parse(data); err != nil {
		log.Error(fmt.Sprintf("parse datahub item failed(%v), data(%s), len:%d,project:%s, topic:%s", err, base64.StdEncoding.EncodeToString(data), len(data), d.projectName, d.topicName))
		return nil
	}

	err := d.doSendSingleMessage(datahubItem.data)
	if err != nil {
		log.Error(fmt.Sprintf("project=%s\ttopic=%s\tmsg=put record failed(%v)", d.projectName, d.topicName, err))
	}

	return nil
}

func (d *Datahub) doSendSingleMessage(message map[string]interface{}) error {
	if d.producer == nil {
		return fmt.Errorf("producer not initialized")
	}
	record := alidatahub.NewTupleRecord(d.getRecordSchema())
	for k, v := range message {
		if err := record.SetValueByName(k, v); err != nil {
			log.Warning(fmt.Sprintf("project=%s\ttopic=%s\tfield=%s\tmsg=set record value failed(%v)", d.projectName, d.topicName, k, err))
		}
	}

	_, err := d.producer.Send([]alidatahub.IRecord{record})
	return err
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.DatahubConfs {
		if _, ok := datahubInstances[name]; ok {
			continue
		}
		m := NewDatahub(conf.AccessId, conf.AccessKey, conf.Endpoint, conf.ProjectName, conf.TopicName, conf.CompressorType, conf.Schemas)

		err := m.Init()
		if err != nil {
			panic(err)
		}
		datahubInstances[name] = m
	}
}

type FeatureLogDatahubFunc func(*Datahub, *module.User, []*module.Item, *context.RecommendContext)

func FeatureLogToDatahub(datahubName string, f FeatureLogDatahubFunc) {
	dh, err := GetDatahub(datahubName)
	if err != nil {
		panic(fmt.Sprintf("get datahub error, :%v", err))
	}
	hook.AddRecommendCleanHook(func(datahub *Datahub, f FeatureLogDatahubFunc) hook.RecommendCleanHookFunc {

		return func(context *context.RecommendContext, params ...interface{}) {
			user := params[0].(*module.User)
			items := params[1].([]*module.Item)
			f(datahub, user, items, context)
		}
	}(dh, f))
}
