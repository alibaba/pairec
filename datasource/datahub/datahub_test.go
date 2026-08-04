package datahub

import (
	"testing"
)

// Init may fail and leave producer/syncLog nil, sending message on such instance
// should degrade with an error log instead of panic, because feature log is sent
// from a bare goroutine where panic can not be recovered by the upper layer.
func TestSendMessageOnUninitializedDatahub(t *testing.T) {
	d := NewDatahub("ak", "sk", "endpoint", "project", "topic", "lz4", nil)

	d.SendMessage([]map[string]interface{}{{"uid": "100"}})

	if err := d.doSendSingleMessage(map[string]interface{}{"uid": "100"}); err == nil {
		t.Fatal("doSendSingleMessage should return error when producer is nil")
	}

	if schema := d.getRecordSchema(); schema != nil {
		t.Fatalf("getRecordSchema should return nil schema, got %v", schema)
	}

	if shards := d.Shards(); len(shards) != 0 {
		t.Fatalf("Shards should be empty, got %v", shards)
	}

	d.StopLoopListShards()
}
