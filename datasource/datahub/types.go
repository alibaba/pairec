package datahub

import (
	"bytes"
	"encoding/binary"
)

type DataType int8

const (
	INT DataType = iota + 1
	INT64
	FLOAT32
	FLOAT64
	STRING
	BOOL
)

type SyncLogDatahubItem struct {
	data map[string]any
}

func NewSyncLogDatahubItem(data map[string]any) *SyncLogDatahubItem {
	return &SyncLogDatahubItem{
		data: data,
	}
}

func (s *SyncLogDatahubItem) Format() []byte {
	buf := bytes.NewBuffer(nil)

	binary.Write(buf, binary.LittleEndian, uint16(len(s.data)))
	for k, v := range s.data {
		// wirte k size
		binary.Write(buf, binary.LittleEndian, uint8(len(k)))
		buf.WriteString(k)
		switch val := v.(type) {
		case int:
			binary.Write(buf, binary.LittleEndian, INT)
			binary.Write(buf, binary.LittleEndian, int64(val))
		case int64:
			binary.Write(buf, binary.LittleEndian, INT64)
			binary.Write(buf, binary.LittleEndian, val)
		case float32:
			binary.Write(buf, binary.LittleEndian, FLOAT32)
			binary.Write(buf, binary.LittleEndian, val)
		case float64:
			binary.Write(buf, binary.LittleEndian, FLOAT64)
			binary.Write(buf, binary.LittleEndian, val)
		case bool:
			binary.Write(buf, binary.LittleEndian, BOOL)
			binary.Write(buf, binary.LittleEndian, val)
		case string:
			binary.Write(buf, binary.LittleEndian, STRING)
			binary.Write(buf, binary.LittleEndian, uint32(len(val)))
			buf.WriteString(val)
		default:
		}
	}

	return buf.Bytes()
}
func (s *SyncLogDatahubItem) Parse(data []byte) error {

	reader := bytes.NewReader(data)
	var mapSize uint16
	if err := binary.Read(reader, binary.LittleEndian, &mapSize); err != nil {
		return err
	}

	s.data = make(map[string]any, mapSize)

	for i := 0; i < int(mapSize); i++ {
		var (
			keySize uint8
			t       DataType
			key     string
		)

		if err := binary.Read(reader, binary.LittleEndian, &keySize); err != nil {
			return err
		}

		keyBytes := make([]byte, keySize)
		if _, err := reader.Read(keyBytes); err != nil {
			return err
		}
		key = string(keyBytes)

		// read value type
		if err := binary.Read(reader, binary.LittleEndian, &t); err != nil {
			return err
		}

		switch t {
		case INT:
			var val int64
			if err := binary.Read(reader, binary.LittleEndian, &val); err != nil {
				return err
			}
			s.data[key] = int(val)
		case INT64:
			var val int64
			if err := binary.Read(reader, binary.LittleEndian, &val); err != nil {
				return err
			}
			s.data[key] = val
		case FLOAT32:
			var val float32
			if err := binary.Read(reader, binary.LittleEndian, &val); err != nil {
				return err
			}
			s.data[key] = val
		case FLOAT64:
			var val float64
			if err := binary.Read(reader, binary.LittleEndian, &val); err != nil {
				return err
			}
			s.data[key] = val
		case BOOL:
			var val bool
			if err := binary.Read(reader, binary.LittleEndian, &val); err != nil {
				return err
			}
			s.data[key] = val
		case STRING:
			var valSize uint32
			if err := binary.Read(reader, binary.LittleEndian, &valSize); err != nil {
				return err
			}
			if valSize == 0 {
				s.data[key] = ""
			} else {
				valueBytes := make([]byte, valSize)

				if _, err := reader.Read(valueBytes); err != nil {
					return err
				}
				s.data[key] = string(valueBytes)
			}
		}
	}

	return nil
}
