package jsonutil

import (
	"bytes"
	"fmt"
	"testing"
)

func TestMarshalSliceWithByteBuffer(t *testing.T) {

	data := []map[string]interface{}{
		{
			"key1": "value1",
			"key2": 0.4,
			"key3": int32(1),
			"key4": int64(-43),
			"key5": uint64(43),
			"key6": int8(4),
			"key7": float32(0),
			"key8": float64(0),
		},
		{
			"key1": "value1",
			"key2": 0.8908,
			"key3": 1,
		},
	}

	var buf bytes.Buffer

	err := MarshalSliceWithByteBuffer(data, &buf)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(buf.String())
}
func TestMarshalSliceWithByteBufferE(t *testing.T) {

	data := []map[string]interface{}{}

	var buf bytes.Buffer

	err := MarshalSliceWithByteBuffer(data, &buf)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(buf.String())
}
