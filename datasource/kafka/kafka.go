package kafka

import (
	"context"
	"fmt"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	kafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
	"strings"
)

type KafkaLog struct {
}

func (l *KafkaLog) Errorf(msg string, args ...interface{}) {

	log.Error(fmt.Sprintf("msg=%s, args=%v", msg, args))
}

type KafkaProducer struct {
	BootstrapServers string
	Topic            string
	Producer         *kafka.Writer
}

var kafkaProducerInstances = make(map[string]*KafkaProducer)

func GetKafkaProducer(name string) (*KafkaProducer, error) {
	if _, ok := kafkaProducerInstances[name]; !ok {
		return nil, fmt.Errorf("KafkaProducer not found, name:%s", name)
	}

	return kafkaProducerInstances[name], nil
}

func NewKafkaProducer(bootstrapServers, topic string) *KafkaProducer {
	p := &KafkaProducer{
		BootstrapServers: bootstrapServers,
		Topic:            topic,
	}
	return p
}

func (k *KafkaProducer) Init() error {
	l := &KafkaLog{}
	w := kafka.Writer{
		Addr:        kafka.TCP(strings.Split(k.BootstrapServers, ",")...),
		Topic:       k.Topic,
		Balancer:    kafka.CRC32Balancer{},
		MaxAttempts: 3,
		Async:       true,
		BatchBytes:  1048576 * 4,
		ErrorLogger: kafka.LoggerFunc(l.Errorf),
		Compression: compress.Snappy,
	}

	k.Producer = &w
	return nil
}

func (k *KafkaProducer) SendMessage(message []byte) {
	err := k.Producer.WriteMessages(context.Background(),
		kafka.Message{
			Value: message,
		})

	if err != nil {
		log.Error(fmt.Sprintf("error=kafka write message(%v)", err))
	}
}

func (k *KafkaProducer) SendMessages(module []byte, message []byte) {
	err := k.Producer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   module,
			Value: message,
		})

	if err != nil {
		log.Error(fmt.Sprintf("error=kafka write message(%v)", err))
	}
}

func (k *KafkaProducer) Close() {
	if k.Producer != nil {
		k.Producer.Close()
	}
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.KafkaConfs {
		if _, ok := kafkaProducerInstances[name]; ok {
			continue
		}
		m := &KafkaProducer{
			BootstrapServers: conf.BootstrapServers,
			Topic:            conf.Topic,
		}
		err := m.Init()
		if err != nil {
			panic(err)
		}
		kafkaProducerInstances[name] = m
	}
}
