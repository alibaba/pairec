package synclog

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"fortio.org/assert"
)

func createSyncLog() (*SyncLog, error) {
	syncLog := NewSyncLog("./tmp", func(b []byte) error {
		fmt.Println(string(b), len(b))
		return nil
	})
	err := syncLog.Init()

	return syncLog, err
}
func TestOpenSyncLog(t *testing.T) {
	os.RemoveAll("./tmp")
	syncLog, err := createSyncLog()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, syncLog.metaInfo.meta.Offset, uint32(0))
	t.Log(syncLog.metaInfo.meta.Offset)
	assert.Equal(t, syncLog.logInfo.size, 0)

	err = syncLog.Close()
	assert.Equal(t, err, nil)
}

func TestWriteSyncLog(t *testing.T) {
	syncLog, err := createSyncLog()
	if err != nil {
		t.Fatal(err)
	}

	item := &SyncLogKVItem{
		Key:   []byte("Hello"),
		Value: []byte("World"),
	}

	err = syncLog.Write(item)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(syncLog.logInfo.size)
	time.Sleep(time.Second)
	assert.Equal(t, syncLog.metaInfo.meta.Offset, uint32(syncLog.logInfo.size))
	err = syncLog.Close()
	assert.Equal(t, err, nil)
}

func TestWriteManySyncLog(t *testing.T) {
	size := 0
	syncLog := NewSyncLog("./tmp", func(b []byte) error {
		//fmt.Println(string(b), len(b))
		size++
		return nil
	})
	err := syncLog.Init()
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(nil)
	for i := 0; i < 200; i++ {
		buf.WriteString("Hello World")
	}
	for i := 0; i < 1000000; i++ {
		item := &SyncLogKVItem{
			Key:   []byte("Hello"),
			Value: buf.Bytes(),
		}

		err = syncLog.Write(item)
		if err != nil {
			t.Fatal(err)
		}

	}
	t.Log(syncLog.logInfo.size)
	time.Sleep(time.Second)
	assert.Equal(t, syncLog.metaInfo.meta.Offset, uint32(syncLog.logInfo.size))
	assert.Equal(t, size, 1000000)
	err = syncLog.Close()
	assert.Equal(t, err, nil)
}
