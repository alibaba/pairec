package datahub

import (
	"encoding/base64"
	"fmt"
	"sync"
	"sync/atomic"
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
	shards       []alidatahub.ShardEntry
	index        uint64
	recordSchema *alidatahub.RecordSchema

	name           string
	syncLog        *synclog.SyncLog
	compressorType string

	active bool
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
		index:          0,
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
	} else {
		topic, err := dh.GetTopic(d.projectName, d.topicName)
		if err != nil {
			return err
		}

		d.recordSchema = topic.RecordSchema
	}
	d.active = true
	dir := fmt.Sprintf("./tmp/%s/%s", d.projectName, d.topicName)
	synclog := synclog.NewSyncLog(dir, d.consumeSyncLog)
	if err := synclog.Init(); err != nil {
		log.Error(fmt.Sprintf("project=%s\ttopic=%s\terror=init sync log error(%v)", d.projectName, d.topicName, err))
		return err
	}

	d.syncLog = synclog

	go d.loopListShards()

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
	for _, shard := range d.shards {
		ret = append(ret, shard.ShardId)
	}
	return
}

func (d *Datahub) loopListShards() error {
	i := 0
	for d.active {
		ls, err := d.datahubApi.ListShard(d.projectName, d.topicName)
		if err != nil {
			log.Error(fmt.Sprintf("project=%s\ttopic=%s\terror=get shard list failed(%v)", d.projectName, d.topicName, err))
			i++
			time.Sleep(time.Second * 10)
			if i >= 10 {
				d.Stop()
			}
			continue
		}
		var shards []alidatahub.ShardEntry
		for _, shard := range ls.Shards {
			if shard.State == alidatahub.ACTIVE {
				shards = append(shards, shard)
			}
		}
		if len(shards) > 0 {
			d.shards = shards
		}
		i = 0
		time.Sleep(time.Minute)
	}

	return nil
}

func (d *Datahub) Stop() {
	d.StopLoopListShards()
	RemoveDatahub(d.name)
}

func (d *Datahub) StopLoopListShards() {
	d.active = false
}

func (d *Datahub) SendMessage(messages []map[string]interface{}) {
	records := make([]alidatahub.IRecord, 0, len(messages))
	shards := d.shards
	for i := 0; i < 3; i++ {
		if len(shards) > 0 {
			break
		}
		shards = d.shards
		time.Sleep(time.Second)
	}
	if len(shards) == 0 {
		log.Error("topic shards empty")
		return
	}
	i := atomic.AddUint64(&d.index, 1)
	shard := shards[(i)%uint64(len(shards))]
	for _, messsage := range messages {
		record := alidatahub.NewTupleRecord(d.recordSchema)
		//record.ShardId = shard.ShardId
		for k, v := range messsage {
			record.SetValueByName(k, v)
		}

		records = append(records, record)
	}

	maxReTry := 3
	retryNum := 0
	retrySendMessage := func() {
		for _, msg := range messages {
			if err := d.syncLog.Write(NewSyncLogDatahubItem(msg)); err != nil {
				log.Error(fmt.Sprintf("project=%s\ttopic=%s\tmsg=write sync log failed(%v)", d.projectName, d.topicName, err))
			}
		}
	}
	for retryNum < maxReTry {
		_, err := d.datahubApi.PutRecordsByShard(d.projectName, d.topicName, shard.ShardId, records)
		if err != nil {
			if _, ok := err.(*alidatahub.LimitExceededError); ok {
				retryNum++
				time.Sleep(2 * time.Second)
				continue
			} else {
				log.Warning(fmt.Sprintf("project=%s\ttopic=%s\tmsg=put record failed(%v)", d.projectName, d.topicName, err))
				retrySendMessage()
				return
			}
		}
		break
	}

	if retryNum >= maxReTry {
		log.Warning(fmt.Sprintf("project=%s\ttopic=%s\tmsg=put record failed", d.projectName, d.topicName))
		retrySendMessage()
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
	records := make([]alidatahub.IRecord, 0, 1)
	shards := d.shards
	for i := 0; i < 3; i++ {
		if len(shards) > 0 {
			break
		}
		shards = d.shards
		time.Sleep(time.Second)
	}
	if len(shards) == 0 {
		return fmt.Errorf("topic shards empty")
	}
	i := atomic.AddUint64(&d.index, 1)

	shard := shards[(i)%uint64(len(shards))]
	record := alidatahub.NewTupleRecord(d.recordSchema)
	//record.ShardId = shard.ShardId
	for k, v := range message {
		record.SetValueByName(k, v)
	}

	records = append(records, record)

	maxReTry := 2
	retryNum := 0
	for retryNum < maxReTry {
		_, err := d.datahubApi.PutRecordsByShard(d.projectName, d.topicName, shard.ShardId, records)
		if err != nil {
			if _, ok := err.(*alidatahub.LimitExceededError); ok {
				log.Error("maybe qps exceed limit,retry")
				retryNum++
				time.Sleep(2 * time.Second)
				continue
			} else {
				return err
			}
		}
		break
	}

	if retryNum >= maxReTry {
		return fmt.Errorf("put record failed")
	}

	return nil
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
