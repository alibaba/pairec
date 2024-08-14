package aliyun_igraph_go_sdk

import (
	"encoding/json"
	"fmt"
)

func PrintResult(i interface{}) {
	iByte, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(iByte))
}

func ToJson(i interface{}) string {
	iByte, _ := json.Marshal(i)
	return string(iByte)
}
