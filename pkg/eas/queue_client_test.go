package eas

import (
	"context"
	"fmt"
	"github.com/alibaba/pairec/pkg/eas/types"
	"strconv"
	"testing"
	"time"
)

const (
	QueueEndpoint  = "1828488879222746.cn-shanghai.pai-eas.aliyuncs.com"
	InputQueueName = "test_group.qservice"
	SinkQueueName  = "test_group.qservice/sink"
	QueueToken     = ""
)

type QueueClientTestCase struct {
	inputQueue *QueueClient
	sinkQueue  *QueueClient
}

var testCase *QueueClientTestCase

func assertEqual(t *testing.T, a, b interface{}) {
	if a != b {
		t.Fatalf("%v != %v", a, b)
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func getQueueClient(t *testing.T) *QueueClientTestCase {
	if testCase == nil {
		testCase = &QueueClientTestCase{}
		var err error
		testCase.inputQueue, err = NewQueueClient(QueueEndpoint, InputQueueName, QueueToken)
		assertNoError(t, err)
		testCase.sinkQueue, err = NewQueueClient(QueueEndpoint, SinkQueueName, QueueToken)
		assertNoError(t, err)
	}
	return testCase
}

func (c *QueueClientTestCase) truncate(t *testing.T) {
	attrs, err := c.inputQueue.Attributes()
	assertNoError(t, err)
	if index, ok := attrs["stream.lastEntry"]; ok {
		idx, _ := strconv.ParseUint(index, 10, 64)
		c.inputQueue.Truncate(context.Background(), idx+1)
	}

	attrs, err = c.sinkQueue.Attributes()
	assertNoError(t, err)
	if index, ok := attrs["stream.lastEntry"]; ok {
		idx, _ := strconv.ParseUint(index, 10, 64)
		c.sinkQueue.Truncate(context.Background(), idx+1)
	}

}

func TestTruncate(t *testing.T) {
	c := getQueueClient(t)

	c.truncate(t)

	latestIndex := uint64(0)
	for i := 0; i < 10; i++ {
		index, _, err := c.sinkQueue.Put(context.Background(), []byte("abc"), types.Tags{})
		assertNoError(t, err)
		latestIndex = index
	}

	c.sinkQueue.Truncate(context.Background(), latestIndex+1)

	attrs, err := c.sinkQueue.Attributes()
	assertNoError(t, err)

	assertEqual(t, attrs["stream.length"], "0")
}

func TestQueueGetByRequestId(t *testing.T) {
	c := getQueueClient(t)

	c.truncate(t)

	_, requestId, err := c.sinkQueue.Put(context.Background(), []byte("abc"), types.Tags{})
	assertNoError(t, err)

	list, err := c.sinkQueue.GetByRequestId(context.Background(), requestId)
	assertNoError(t, err)

	assertEqual(t, len(list), 1)
	assertEqual(t, string(list[0].Data), "abc")
}

func TestQueueGetByIndex(t *testing.T) {
	c := getQueueClient(t)

	c.truncate(t)

	index, _, err := c.sinkQueue.Put(context.Background(), []byte("abc"), types.Tags{})
	assertNoError(t, err)

	list, err := c.sinkQueue.GetByIndex(context.Background(), index)
	assertNoError(t, err)

	assertEqual(t, len(list), 1)
	assertEqual(t, string(list[0].Data), "abc")
}

func TestWatchWithAutoCommit(t *testing.T) {
	c := getQueueClient(t)

	c.truncate(t)

	for i := 0; i < 10; i++ {
		_, _, err := c.sinkQueue.Put(context.Background(), []byte(strconv.Itoa(i)), types.Tags{})
		assertNoError(t, err)
	}

	watcher, err := c.sinkQueue.Watch(context.Background(), 0, 5, false, true)
	assertNoError(t, err)

	for i := 0; i < 10; i++ {
		df := <-watcher.FrameChan()
		assertEqual(t, string(df.Data), strconv.Itoa(i))
	}

	watcher.Close()

	time.Sleep(2 * time.Second)

	attrs, err := c.sinkQueue.Attributes()
	assertNoError(t, err)
	assertEqual(t, attrs["stream.length"], "0")
}

func TestWatchWithManualCommit(t *testing.T) {
	c := getQueueClient(t)

	c.truncate(t)

	for i := 0; i < 10; i++ {
		_, _, err := c.sinkQueue.Put(context.Background(), []byte(strconv.Itoa(i)), types.Tags{})
		assertNoError(t, err)
	}

	watcher, err := c.sinkQueue.Watch(context.Background(), 0, 5, false, false)
	assertNoError(t, err)

	for i := 0; i < 10; i++ {
		df := <-watcher.FrameChan()
		err := c.sinkQueue.Commit(context.Background(), df.Index.Uint64())
		assertNoError(t, err)
		assertEqual(t, string(df.Data), strconv.Itoa(i))
	}

	watcher.Close()

	time.Sleep(2 * time.Second)

	attrs, err := c.sinkQueue.Attributes()
	assertNoError(t, err)
	assertEqual(t, attrs["stream.length"], "0")
}

func TestWatchWithReconnect(t *testing.T) {
	c := getQueueClient(t)

	c.truncate(t)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		i := 0
		for {
			select {
			case <-time.NewTicker(time.Microsecond * 1).C:
				_, _, err := c.sinkQueue.Put(context.Background(), []byte(strconv.Itoa(i)), types.Tags{})
				assertNoError(t, err)
				i += 1
			case <-ctx.Done():
				break
			}
		}
	}()

	watcher, err := c.sinkQueue.Watch(context.Background(), 0, 5, false, false)
	assertNoError(t, err)

	for i := 0; i < 100; i++ {
		df, ok := <-watcher.FrameChan()
		assertEqual(t, ok, true)
		err := c.sinkQueue.Commit(context.Background(), df.Index.Uint64())
		if err != nil {
			fmt.Printf("commit id: %v failed: %v", df.Index, err)
		}
		assertNoError(t, err)
		assertEqual(t, string(df.Data), strconv.Itoa(i))
	}

	watcher.Close()
	cancel()
}
