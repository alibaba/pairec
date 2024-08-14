package be

import (
	"fmt"
	"hash/fnv"
	"net/url"
	"strconv"
	"strings"
)

type WriteType string

const (
	WriteTypeAdd    WriteType = "ADD"
	WriteTypeDelete WriteType = "DELETE"
)

type WriteRequest struct {
	WriteType   WriteType           `json:"write_type"`
	TableName   string              `json:"table_name"`
	Contents    []map[string]string `json:"contents"`
	PrimaryKey  string              `json:"primary_key"`
	QueryParams map[string]string   `json:"query_params"`
}

func NewWriteRequest(writeType WriteType, tableName string, primaryKey string, contents []map[string]string) *WriteRequest {
	return &WriteRequest{WriteType: writeType,
		TableName:   tableName,
		PrimaryKey:  primaryKey,
		Contents:    contents,
		QueryParams: map[string]string{},
	}
}

func (r *WriteRequest) Validate() error {
	if r.TableName == "" {
		return InvalidParamsError{"Table name not set"}
	}
	if len(r.Contents) == 0 {
		return InvalidParamsError{"Empty contents"}
	}
	if r.PrimaryKey == "" {
		return InvalidParamsError{"Partition key not set"}
	}
	primaryKeyExist := false
	for _, content := range r.Contents {
		for k, v := range content {
			if k == "" || v == "" {
				return InvalidParamsError{fmt.Sprintf("Key or value is empty for kv pair[%s][%s]", k, v)}
			}
			if k == r.PrimaryKey {
				primaryKeyExist = true
			}
		}
	}
	if !primaryKeyExist {
		return InvalidParamsError{fmt.Sprintf("Partition key[%s] not exist in contents", r.PrimaryKey)}
	}
	return nil
}

func (r *WriteRequest) AddQueryParam(key string, value string) *WriteRequest {
	r.QueryParams[key] = value
	return r
}

func (r *WriteRequest) SetQueryParams(params map[string]string) *WriteRequest {
	r.QueryParams = params
	return r
}

func (r *WriteRequest) BuildUri(index int) url.URL {
	uri := url.URL{Path: "sendmsg"}

	var separator byte = 31
	var lineBreak byte = '\n'

	var contentBuilder strings.Builder
	var primaryKeyValue string
	var content = r.Contents[index]
	for k, v := range content {
		if r.PrimaryKey == k {
			primaryKeyValue = v
			continue
		}
		contentBuilder.WriteString(k)
		contentBuilder.WriteString("=")
		contentBuilder.WriteString(v)
		contentBuilder.WriteByte(separator)
		contentBuilder.WriteByte(lineBreak)
	}

	var builder strings.Builder
	builder.WriteString("CMD=")
	builder.WriteString(strings.ToLower(string(r.WriteType)))
	builder.WriteByte(separator)
	builder.WriteByte(lineBreak)
	builder.WriteString(r.PrimaryKey + "=" + primaryKeyValue)
	builder.WriteByte(separator)
	builder.WriteByte(lineBreak)
	builder.WriteString(contentBuilder.String())

	h := fnv.New64a()
	_, err := h.Write([]byte(primaryKeyValue))
	if err != nil {
		//TODO add log
	}
	hashValue := h.Sum64()

	query := uri.Query()
	query.Add("table", r.TableName)
	query.Add("h", strconv.Itoa(int(hashValue)))
	query.Add("msg", builder.String())

	rawQuery := query.Encode()
	for k, v := range r.QueryParams {
		rawQuery = rawQuery + "&" + k + "=" + v
	}
	uri.RawQuery = rawQuery
	return uri
}
