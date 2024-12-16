package sort

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type PluginAPISort struct {
	name string
	url  string
}

func (s *PluginAPISort) Sort(sortData *SortData) error {
	if _, ok := sortData.Data.([]*module.Item); !ok {
		return errors.New("sort data type error")
	}

	return s.doSort(sortData)
}

func (s *PluginAPISort) doSort(sortData *SortData) error {
	start := time.Now()
	items := sortData.Data.([]*module.Item)
	var newItems []*module.Item

	reqData := module.NewPluginAPIRequest(sortData.User, items, sortData.Context)
	// 将结构体编码为 JSON 格式
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	// 创建一个 HTTP 请求
	req, err := http.NewRequest("POST", s.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: received non-200 response code: %v", resp.StatusCode)
	}

	// 读取和处理响应
	var respData module.PluginAPIFilterResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	if respData.Code != 200 {
		if respData.Code >= 400 {
			return fmt.Errorf("error: received non-200 business code: %d, msg: %s", respData.Code, respData.Msg)
		} else {
			log.Warning(fmt.Sprintf("requestId=%s\tmodule=%s\tmsg=%v", sortData.Context.RecommendId, "PluginAPISort", respData.Msg))
		}
	}

	itemMap := make(map[string]*module.Item, len(items))
	for i, item := range items {
		itemMap[string(item.Id)] = items[i]
	}

	for _, itemId := range respData.Items {
		newItems = append(newItems, itemMap[itemId])
	}

	sortData.Data = newItems
	sortInfoLogWithName(sortData, "PluginAPISort", s.name, len(items), start)
	return nil
}

func NewPluginAPISort(config recconf.SortConfig) *PluginAPISort {
	p := PluginAPISort{}
	p.name = config.Name
	p.url = config.PluginAPIConfig.URL

	return &p
}
