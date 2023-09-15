package datahub

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	alidatahub "github.com/aliyun/aliyun-datahub-sdk-go/datahub"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/service/hook"
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

	active bool
	name   string
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

func NewDatahub(accessId, accessKey, endpoint, project, topic string, schemas []recconf.DatahubTopicSchema) *Datahub {
	p := &Datahub{
		accessId:    accessId,
		accessKey:   accessKey,
		endpoint:    endpoint,
		projectName: project,
		topicName:   topic,
		index:       0,
		schemas:     schemas,
	}
	return p
}

func (d *Datahub) Init() error {
	account := alidatahub.NewAliyunAccount(d.accessId, d.accessKey)
	config := alidatahub.NewDefaultConfig()
	config.CompressorType = alidatahub.DEFLATE
	config.EnableBinary = false
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
		if _, err := d.datahubApi.CreateTupleTopic(d.projectName, d.topicName, fmt.Sprintf("create topic %s", d.topicName), 3, 3, recordSchema); err != nil {
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
			log.Error(fmt.Sprintf("error=get shard list failed(%v)", err))
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
	for _, messsage := range messages {
		i := atomic.AddUint64(&d.index, 1)
		shard := shards[(i)%uint64(len(shards))]
		record := alidatahub.NewTupleRecord(d.recordSchema, 0)
		record.ShardId = shard.ShardId
		for k, v := range messsage {
			record.SetValueByName(k, v)
		}

		records = append(records, record)
	}

	maxReTry := 3
	retryNum := 0
	for retryNum < maxReTry {
		result, err := d.datahubApi.PutRecords(d.projectName, d.topicName, records)
		if err != nil {
			if _, ok := err.(*alidatahub.LimitExceededError); ok {
				retryNum++
				time.Sleep(2 * time.Second)
				continue
			} else {
				log.Error(fmt.Sprintf("put record failed(%v)", err))
				return
			}
		}
		if len(result.FailedRecords) > 0 {
			log.Error(fmt.Sprintf("put successful num is %d, put records failed num is %d,msg=%s, code=%s\n", len(records)-result.FailedRecordCount, result.FailedRecordCount, result.FailedRecords[0].ErrorMessage, result.FailedRecords[0].ErrorCode))
		}
		break
	}

}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.DatahubConfs {
		if _, ok := datahubInstances[name]; ok {
			continue
		}
		m := NewDatahub(conf.AccessId, conf.AccessKey, conf.Endpoint, conf.ProjectName, conf.TopicName, conf.Schemas)

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
