package filter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type PluginAPIFilter struct {
	url string
}

func NewPluginAPIFilter(config recconf.FilterConfig) *PluginAPIFilter {
	filter := PluginAPIFilter{}

	filter.url = config.PluginAPIFilterConf.URL

	return &filter
}
func (f *PluginAPIFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	return f.doFilter(filterData)
}

func (f *PluginAPIFilter) doFilter(filterData *FilterData) error {
	items := filterData.Data.([]*module.Item)
	var newItems []*module.Item

	reqData := module.NewPluginAPIRequest(filterData.User, items, filterData.Context)
	// 将结构体编码为 JSON 格式
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	// 创建一个 HTTP 请求
	req, err := http.NewRequest("POST", f.url, bytes.NewBuffer(jsonData))
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
			log.Warning(fmt.Sprintf("requestId=%s\tmodule=%s\tmsg=%v", filterData.Context.RecommendId, "PluginAPIFilter", respData.Msg))
		}
	}

	itemMap := make(map[string]*module.Item, len(items))
	for i, item := range items {
		itemMap[string(item.Id)] = items[i]
	}

	for _, itemId := range respData.Items {
		newItems = append(newItems, itemMap[itemId])
	}

	filterData.Data = newItems

	return nil
}
