package jsonutil

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

const Quotation_Str = "\""

func MarshalSliceWithByteBuffer(data interface{}, buf *bytes.Buffer) error {
	if _, ok := data.([]map[string]interface{}); !ok {
		return errors.New("data is not slice type")
	}

	buf.WriteByte('[')
	for _, dataMap := range data.([]map[string]interface{}) {
		buf.WriteByte('{')
		for k, v := range dataMap {
			buf.WriteString(Quotation_Str + k + Quotation_Str)
			buf.WriteByte(':')
			switch d := v.(type) {
			case string:
				buf.WriteString(Quotation_Str + d + Quotation_Str)
			case uint8:
				buf.WriteString(strconv.Itoa(int(d)))
			case uint32:
				buf.WriteString(strconv.Itoa(int(d)))
			case uint64:
				buf.WriteString(strconv.FormatUint(d, 10))
			case int:
				buf.WriteString(strconv.Itoa(d))
			case int8:
				buf.WriteString(strconv.Itoa(int(d)))
			case int32:
				buf.WriteString(strconv.Itoa(int(d)))
			case int64:
				buf.WriteString(strconv.FormatInt(d, 10))
			case float32:
				buf.WriteString(strconv.FormatFloat(float64(d), 'f', -1, 32))
			case float64:
				buf.WriteString(strconv.FormatFloat(d, 'f', -1, 64))
			case bool:
				buf.WriteString(strconv.FormatBool(d))
			default:
				fmt.Println("not support type")
			}

			buf.WriteByte(',')
		}

		buf.Truncate(buf.Len() - 1)
		buf.WriteByte('}')
		buf.WriteByte(',')
	}
	buf.Truncate(buf.Len() - 1)
	buf.WriteByte(']')
	return nil
}
