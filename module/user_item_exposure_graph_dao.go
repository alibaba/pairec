package module

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/graph"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	igraph "github.com/aliyun/aliyun-igraph-go-sdk"
)

type User2ItemExposureGraphDao struct {
	graphClient              *graph.GraphClient
	instanceId               string
	tableName                string
	userNode                 string
	itemNode                 string
	edge                     string
	maxItems                 int
	timeInterval             int //  second
	generateItemDataFuncName string
	writeLogExcludeScenes    map[string]bool
	clearLogScene            string
}

func NewUser2ItemExposureGraphDao(config recconf.FilterConfig) *User2ItemExposureGraphDao {
	graphClient, err := graph.GetGraphClient(config.DaoConf.GraphName)
	if err != nil {
		panic(err)
	}

	dao := &User2ItemExposureGraphDao{
		graphClient:              graphClient,
		instanceId:               config.DaoConf.InstanceId,
		tableName:                config.DaoConf.TableName,
		userNode:                 config.DaoConf.UserNode,
		itemNode:                 config.DaoConf.ItemNode,
		edge:                     config.DaoConf.Edge,
		maxItems:                 -1,
		timeInterval:             -1,
		generateItemDataFuncName: config.GenerateItemDataFuncName,
		writeLogExcludeScenes:    make(map[string]bool),
		clearLogScene:            config.ClearLogIfNotEnoughScene,
	}

	if config.MaxItems > 0 {
		dao.maxItems = config.MaxItems
	}

	if config.TimeInterval > 0 {
		dao.timeInterval = config.TimeInterval
	}
	for _, scene := range config.WriteLogExcludeScenes {
		dao.writeLogExcludeScenes[scene] = true
	}
	return dao
}

func (d *User2ItemExposureGraphDao) LogHistory(user *User, items []*Item, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)
	if _, exist := d.writeLogExcludeScenes[scene]; exist {
		return
	}

	if len(items) == 0 {
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureGraphDao\terr=items empty", context.RecommendId))
		return
	}

	userContent := make(map[string]string)
	itemContent := make(map[string]string)
	edgeContent := make(map[string]string)

	// 将 uid 写入 user 节点
	uid := string(user.Id)
	createTime := strconv.FormatInt(time.Now().Unix(), 10)
	userContent["uid"] = uid
	userContent["create_time"] = createTime
	userRequest := igraph.NewWriteRequest(igraph.WriteTypeAdd, d.instanceId, d.tableName, d.userNode, "uid", "", userContent)
	_, err := d.graphClient.GraphClient.Write(userRequest)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureGraphDao\tuid=%s\terr=%v", context.RecommendId, user.Id, err))
	}

	//将 item 数据写入 item 节点
	var ret string
	for _, item := range items {
		itemData := getGenerateItemDataFunc(d.generateItemDataFuncName)(user.Id, item)
		ret = ret + "," + itemData
	}
	ret = ret[1:]
	itemContent["item"] = ret
	itemContent["create_time"] = createTime

	itemRequest := igraph.NewWriteRequest(igraph.WriteTypeAdd, d.instanceId, d.tableName, d.itemNode, "item", "", itemContent)
	_, err = d.graphClient.GraphClient.Write(itemRequest)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureGraphDao\tuid=%s\terr=%v", context.RecommendId, user.Id, err))
	}

	// 将 user 和 item 写入 edge
	edgeContent["uid"] = uid
	edgeContent["item"] = ret
	edgeContent["create_time"] = createTime

	edgeRequest := igraph.NewWriteRequest(igraph.WriteTypeAdd, d.instanceId, d.tableName, d.edge, "uid", "", edgeContent)
	_, err = d.graphClient.GraphClient.Write(edgeRequest)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureGraphDao\tuid=%s\terr=%v", context.RecommendId, user.Id, err))
	}

	log.Info(fmt.Sprintf("requestId=%s\tscene=%s\tuid=%s\tmsg=log history success", context.RecommendId, scene, user.Id))
}

func (d *User2ItemExposureGraphDao) FilterByHistory(uid UID, items []*Item, context *context.RecommendContext) (ret []*Item) {
	queryString := fmt.Sprintf("g(\"%s\").V(\"%s\").hasLabel(\"%s\").outE()", d.tableName, string(uid), d.userNode)

	if d.timeInterval > 0 {
		t := time.Now().Unix() - int64(d.timeInterval)
		queryString += fmt.Sprintf(".filter(\"create_time>%d\")", t)
	}

	queryString += ".inV()"

	if d.maxItems > 0 {
		queryString += fmt.Sprintf(".limit(%d)", d.maxItems)
	}

	queryParam := make(map[string]string)
	readRequest := &igraph.ReadRequest{
		QueryString: queryString,
		QueryParams: queryParam,
	}
	resp, err := d.graphClient.GraphClient.Read(readRequest)
	if err != nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureGraphDao\tuid=%s\terr=%v", uid, err))
	}

	fiterIds := make(map[string]bool)

	results := resp.Result
	for _, res := range results {
		for _, data := range res.Data {
			itemDatas := data["item"].(string)
			ids := strings.Split(itemDatas, ",")
			for _, id := range ids {
				fiterIds[id] = true
			}
		}
	}

	for _, item := range items {
		itemData := getGenerateItemDataFunc(d.generateItemDataFuncName)(uid, item)
		if _, ok := fiterIds[itemData]; !ok {
			ret = append(ret, item)
		}
	}
	return
}

func (d *User2ItemExposureGraphDao) ClearHistory(user *User, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)
	if scene != d.clearLogScene {
		return
	}

	content := make(map[string]string)

	content["uid"] = string(user.Id)

	// 删除 edge 中的 uid
	request := igraph.NewWriteRequest(igraph.WriteTypeDelete, d.instanceId, d.tableName, d.edge, "uid", "", content)
	_, err := d.graphClient.GraphClient.Write(request)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureGraphDao\tuid=%s\terr=%v", context.RecommendId, user.Id, err))
	}
	log.Info(fmt.Sprintf("requestId=%s\tscene=%s\tuid=%s\tmsg=log history success", context.RecommendId, scene, user.Id))

}

func (d *User2ItemExposureGraphDao) GetExposureItemIds(user *User, context *context.RecommendContext) (ret string) {
	queryString := fmt.Sprintf("g(\"%s\").V(\"%s\").hasLabel(\"%s\").outE()", d.tableName, string(user.Id), d.userNode)

	if d.timeInterval > 0 {
		t := time.Now().Unix() - int64(d.timeInterval)
		queryString += fmt.Sprintf(".filter(\"create_time>%d\")", t)
	}

	queryString += ".inV()"

	if d.maxItems > 0 {
		queryString += fmt.Sprintf(".limit(%d)", d.maxItems)
	}

	queryParam := make(map[string]string)
	readRequest := &igraph.ReadRequest{
		QueryString: queryString,
		QueryParams: queryParam,
	}
	resp, err := d.graphClient.GraphClient.Read(readRequest)
	if err != nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureGraphDao\tuid=%s\terr=%v", user.Id, err))
	}

	fiterIds := make([]string, 0, 10)

	results := resp.Result
	for _, res := range results {
		for _, data := range res.Data {
			itemDatas := data["item"].(string)
			ids := strings.Split(itemDatas, ",")
			for _, id := range ids {
				fiterIds = append(fiterIds, id)
			}
		}
	}
	ret = strings.Join(fiterIds, ",")
	return
}
