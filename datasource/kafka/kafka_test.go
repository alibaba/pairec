package kafka

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	_ "go.uber.org/automaxprocs"
	"runtime"
	"strings"
	"testing"
)

func TestDebugLog(*testing.T) {
	fmt.Println("real GOMAXPROCS:", runtime.GOMAXPROCS(0))

	bootstrapServers :=
		"alikafka-post-cn-7pp2mdw5b005-1.alikafka.aliyuncs.com:9093," +
			"alikafka-post-cn-7pp2mdw5b005-2.alikafka.aliyuncs.com:9093," +
			"alikafka-post-cn-7pp2mdw5b005-3.alikafka.aliyuncs.com:9093"
	//alikafka-post-cn-7pp2mdw5b005-1.alikafka.aliyuncs.com:9093,
	//alikafka-post-cn-7pp2mdw5b005-2.alikafka.aliyuncs.com:9093,
	//alikafka-post-cn-7pp2mdw5b005-3.alikafka.aliyuncs.com:9093
	write := &kafka.Writer{
		Addr:     kafka.TCP(strings.Split(bootstrapServers, ",")...),
		Topic:    "debug_log",
		Balancer: &kafka.CRC32Balancer{},
	}
	err := write.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte("HomePage"),
			Value: []byte("module rank"),
		})
	if err != nil {
		fmt.Printf("【err】= %s", err)
	}

	if err := write.Close(); err != nil {
		fmt.Println(err)
	}
}
