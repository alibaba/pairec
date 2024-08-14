package aliyun_igraph_go_sdk

import (
	"fmt"
	"net/url"
)

type WriteType string

const (
	WriteTypeAdd    WriteType = "ADD"
	WriteTypeDelete WriteType = "DELETE"
)

type WriteRequest struct {
	WriteType    WriteType         `json:"write_type"`
	InstanceName string            `json:"instance_name"`
	GraphName    string            `json:"graph_name"`
	LabelName    string            `json:"label_name"`
	Contents     map[string]string `json:"contents"`
	PrimaryKey   string            `json:"primary_key"`
	SecondaryKey string            `json:"secondary_key"`
	QueryParams  map[string]string `json:"query_params"`
}

func NewWriteRequest(writeType WriteType, instanceName string, graphName string, labelName string, primaryKey string, secondaryKey string, contents map[string]string) *WriteRequest {
	return &WriteRequest{WriteType: writeType,
		InstanceName: instanceName,
		GraphName:    graphName,
		LabelName:    labelName,
		PrimaryKey:   primaryKey,
		SecondaryKey: secondaryKey,
		Contents:     contents,
		QueryParams:  map[string]string{},
	}
}

func (r *WriteRequest) AddContent(key string, value string) *WriteRequest {
	r.Contents[key] = value
	return r
}

func (r *WriteRequest) Validate() error {
	if r.InstanceName == "" {
		return InvalidParamsError{"Instance name not set"}
	}

	if r.GraphName == "" {
		return InvalidParamsError{"Graph name not set"}
	}
	if r.LabelName == "" {
		return InvalidParamsError{"Label name not set"}
	}
	if len(r.Contents) == 0 {
		return InvalidParamsError{"Empty contents"}
	}
	if r.PrimaryKey == "" {
		return InvalidParamsError{"Partition key not set"}
	}
	primaryKeyExist := false
	for k, v := range r.Contents {
		if k == "" || v == "" {
			return InvalidParamsError{fmt.Sprintf("Key or value is empty for kv pair[%s][%s]", k, v)}
		}
		if k == r.PrimaryKey {
			primaryKeyExist = true
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

func (r *WriteRequest) BuildUri() url.URL {
	uri := url.URL{Path: "update"}

	var primaryKeyValue string
	var secondaryKeyValue string

	var writeType string
	if r.WriteType == WriteTypeAdd {
		writeType = "1"
	} else if r.WriteType == WriteTypeDelete {
		writeType = "2"
	}

	query := uri.Query()
	query.Add("table", r.GraphName+"_"+r.LabelName)
	query.Add("type", writeType)
	for k, v := range r.Contents {
		if r.PrimaryKey == k {
			primaryKeyValue = v
			query.Add("pkey", primaryKeyValue)
			continue
		}
		if r.SecondaryKey == k {
			secondaryKeyValue = v
			query.Add("skey", secondaryKeyValue)
			continue
		}
		query.Add(k, v)
	}

	rawQuery := query.Encode()
	for k, v := range r.QueryParams {
		rawQuery = rawQuery + "&" + k + "=" + v
	}
	uri.RawQuery = rawQuery
	return uri
}
