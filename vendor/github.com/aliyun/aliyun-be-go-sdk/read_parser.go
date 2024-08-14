package be

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-be-go-sdk/be_fb"
	flatbuffers "github.com/google/flatbuffers/go"
)

type ReadParser interface {
	parse(buf []byte, result *ReadResult) error
}

var (
	defaultJsonReadParser = JsonReadParser{}
	defaultFbReadParser   = FbReadParser{}
)

type JsonReadParser struct {
}

func (p *JsonReadParser) parse(buf []byte, readResult *ReadResult) error {

	if jErr := json.Unmarshal(buf, readResult); jErr != nil {
		fmt.Println(jErr)
		return jErr
	}
	return nil
}

type FbReadParser struct {
}

func (p *FbReadParser) parse(buf []byte, readResult *ReadResult) (err error) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("failed to parse result using fb:", e)
			err = errors.New("failed to parse result using fb")
		}
	}()

	result := be_fb.GetRootAsResult(buf, 0)
	records := result.Records(nil)
	if records == nil {
		return errors.New("failed to parse result using fb")
	}

	readResult.MatchItems = MatchItem{}

	fieldNames := make([]string, records.FieldNameLength())
	for i := 0; i < records.FieldNameLength(); i++ {
		fieldNames[i] = string(records.FieldName(i))
	}
	readResult.ErrorCode = int(result.ErrorCode())
	readResult.ErrorMessage = string(result.ErrorMessage())
	traceInfoBytes := result.TraceInfo()
	if traceInfoBytes != nil {
		var traceInfo map[string]interface{}
		err := json.Unmarshal(traceInfoBytes, &traceInfo)
		if err != nil {
			return err
		}
		readResult.TraceInfo = traceInfo
	}
	readResult.MatchItems.FieldNames = fieldNames
	readResult.MatchItems.FieldValues = make([][]interface{}, records.DocCount())

	for i := 0; i < records.RecordColumnsLength(); i++ {
		column := new(be_fb.FieldValueColumnTable)
		records.RecordColumns(column, i)

		fieldValueColumnUnion := new(flatbuffers.Table)
		if column.FieldValueColumn(fieldValueColumnUnion) {
			unionType := column.FieldValueColumnType()

			switch unionType {
			case be_fb.FieldValueColumnInt8ValueColumn:
				fieldValueColumn := new(be_fb.Int8ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], int64(fieldValueColumn.Value(j)))
				}
			case be_fb.FieldValueColumnUInt8ValueColumn:
				fieldValueColumn := new(be_fb.UInt8ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], int64(fieldValueColumn.Value(j)))
				}
			case be_fb.FieldValueColumnInt16ValueColumn:
				fieldValueColumn := new(be_fb.Int16ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], int64(fieldValueColumn.Value(j)))
				}
			case be_fb.FieldValueColumnUInt16ValueColumn:
				fieldValueColumn := new(be_fb.UInt16ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], int64(fieldValueColumn.Value(j)))
				}
			case be_fb.FieldValueColumnInt32ValueColumn:
				fieldValueColumn := new(be_fb.Int32ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], int64(fieldValueColumn.Value(j)))
				}
			case be_fb.FieldValueColumnUInt32ValueColumn:
				fieldValueColumn := new(be_fb.UInt32ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], int64(fieldValueColumn.Value(j)))
				}
			case be_fb.FieldValueColumnInt64ValueColumn:
				fieldValueColumn := new(be_fb.Int64ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], int64(fieldValueColumn.Value(j)))
				}
			case be_fb.FieldValueColumnUInt64ValueColumn:
				fieldValueColumn := new(be_fb.UInt64ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], int64(fieldValueColumn.Value(j)))
				}
			case be_fb.FieldValueColumnFloatValueColumn:
				fieldValueColumn := new(be_fb.FloatValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], float64(fieldValueColumn.Value(j)))
				}
			case be_fb.FieldValueColumnDoubleValueColumn:
				fieldValueColumn := new(be_fb.DoubleValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], fieldValueColumn.Value(j))
				}
			case be_fb.FieldValueColumnStringValueColumn:
				fieldValueColumn := new(be_fb.StringValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], string(fieldValueColumn.Value(j)))
				}
			case be_fb.FieldValueColumnMultiInt8ValueColumn:
				fieldValueColumn := new(be_fb.MultiInt8ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					multiValue := new(be_fb.MultiInt8Value)
					fieldValueColumn.Value(multiValue, j)
					multiValueSlice := make([]int64, multiValue.ValueLength())
					for k := 0; k < multiValue.ValueLength(); k++ {
						multiValueSlice[k] = int64(multiValue.Value(k))
					}
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], multiValueSlice)
				}
			case be_fb.FieldValueColumnMultiUInt8ValueColumn:
				fieldValueColumn := new(be_fb.MultiUInt8ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					multiValue := new(be_fb.MultiUInt8Value)
					fieldValueColumn.Value(multiValue, j)
					multiValueSlice := make([]int64, multiValue.ValueLength())
					for k := 0; k < multiValue.ValueLength(); k++ {
						multiValueSlice[k] = int64(multiValue.Value(k))
					}
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], multiValueSlice)
				}
			case be_fb.FieldValueColumnMultiInt16ValueColumn:
				fieldValueColumn := new(be_fb.MultiInt16ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					multiValue := new(be_fb.MultiInt16Value)
					fieldValueColumn.Value(multiValue, j)
					multiValueSlice := make([]int64, multiValue.ValueLength())
					for k := 0; k < multiValue.ValueLength(); k++ {
						multiValueSlice[k] = int64(multiValue.Value(k))
					}
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], multiValueSlice)
				}
			case be_fb.FieldValueColumnMultiUInt16ValueColumn:
				fieldValueColumn := new(be_fb.MultiUInt16ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					multiValue := new(be_fb.MultiUInt16Value)
					fieldValueColumn.Value(multiValue, j)
					multiValueSlice := make([]int64, multiValue.ValueLength())
					for k := 0; k < multiValue.ValueLength(); k++ {
						multiValueSlice[k] = int64(multiValue.Value(k))
					}
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], multiValueSlice)
				}
			case be_fb.FieldValueColumnMultiInt32ValueColumn:
				fieldValueColumn := new(be_fb.MultiInt32ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					multiValue := new(be_fb.MultiInt32Value)
					fieldValueColumn.Value(multiValue, j)
					multiValueSlice := make([]int64, multiValue.ValueLength())
					for k := 0; k < multiValue.ValueLength(); k++ {
						multiValueSlice[k] = int64(multiValue.Value(k))
					}
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], multiValueSlice)
				}
			case be_fb.FieldValueColumnMultiUInt32ValueColumn:
				fieldValueColumn := new(be_fb.MultiUInt32ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					multiValue := new(be_fb.MultiUInt32Value)
					fieldValueColumn.Value(multiValue, j)
					multiValueSlice := make([]int64, multiValue.ValueLength())
					for k := 0; k < multiValue.ValueLength(); k++ {
						multiValueSlice[k] = int64(multiValue.Value(k))
					}
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], multiValueSlice)
				}
			case be_fb.FieldValueColumnMultiInt64ValueColumn:
				fieldValueColumn := new(be_fb.MultiInt64ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					multiValue := new(be_fb.MultiInt64Value)
					fieldValueColumn.Value(multiValue, j)
					multiValueSlice := make([]int64, multiValue.ValueLength())
					for k := 0; k < multiValue.ValueLength(); k++ {
						multiValueSlice[k] = int64(multiValue.Value(k))
					}
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], multiValueSlice)
				}
			case be_fb.FieldValueColumnMultiUInt64ValueColumn:
				fieldValueColumn := new(be_fb.MultiUInt64ValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					multiValue := new(be_fb.MultiUInt64Value)
					fieldValueColumn.Value(multiValue, j)
					multiValueSlice := make([]int64, multiValue.ValueLength())
					for k := 0; k < multiValue.ValueLength(); k++ {
						multiValueSlice[k] = int64(multiValue.Value(k))
					}
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], multiValueSlice)
				}
			case be_fb.FieldValueColumnMultiFloatValueColumn:
				fieldValueColumn := new(be_fb.MultiFloatValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					multiValue := new(be_fb.MultiFloatValue)
					fieldValueColumn.Value(multiValue, j)
					multiValueSlice := make([]float64, multiValue.ValueLength())
					for k := 0; k < multiValue.ValueLength(); k++ {
						multiValueSlice[k] = float64(multiValue.Value(k))
					}
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], multiValueSlice)
				}
			case be_fb.FieldValueColumnMultiDoubleValueColumn:
				fieldValueColumn := new(be_fb.MultiDoubleValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					multiValue := new(be_fb.MultiDoubleValue)
					fieldValueColumn.Value(multiValue, j)
					multiValueSlice := make([]float64, multiValue.ValueLength())
					for k := 0; k < multiValue.ValueLength(); k++ {
						multiValueSlice[k] = multiValue.Value(k)
					}
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], multiValueSlice)
				}
			case be_fb.FieldValueColumnMultiStringValueColumn:
				fieldValueColumn := new(be_fb.MultiStringValueColumn)
				fieldValueColumn.Init(fieldValueColumnUnion.Bytes, fieldValueColumnUnion.Pos)
				for j := 0; j < fieldValueColumn.ValueLength(); j++ {
					multiValue := new(be_fb.MultiStringValue)
					fieldValueColumn.Value(multiValue, j)
					multiValueSlice := make([]string, multiValue.ValueLength())
					for k := 0; k < multiValue.ValueLength(); k++ {
						multiValueSlice[k] = string(multiValue.Value(k))
					}
					readResult.MatchItems.FieldValues[j] =
						append(readResult.MatchItems.FieldValues[j], multiValueSlice)
				}
			}
		}
	}
	return nil
}
