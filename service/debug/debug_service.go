package debug

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/datahub"
	plog "github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type LogOutputer interface {
	WriteLog(log map[string]interface{})
}

type ConsoleOutput struct {
}

func (t *ConsoleOutput) WriteLog(log map[string]interface{}) {
	j, _ := json.Marshal(log)
	fmt.Println(string(j))
}

type DatahubOutput struct {
	datahub *datahub.Datahub
}

func NewDatahubOutput(config *recconf.DebugConfig) *DatahubOutput {
	datahubclient, err := datahub.GetDatahub(config.DatahubName)
	hub := DatahubOutput{}
	if err != nil {
		fmt.Println(err)
	} else {
		hub.datahub = datahubclient
	}

	return &hub
}
func (t *DatahubOutput) WriteLog(log map[string]interface{}) {
	t.datahub.SendMessage([]map[string]interface{}{log})
}

var fileOutputMux sync.Mutex

type FileOutput struct {
	path       string
	maxFileNum int
}

func NewFileOutput(config *recconf.DebugConfig) *FileOutput {
	file := FileOutput{}

	file.path = config.FilePath
	filepath.Clean(file.path)
	if !strings.HasSuffix(file.path, "/") {
		file.path += "/"
	}
	if config.MaxFileNum > 0 {
		file.maxFileNum = config.MaxFileNum
	} else {
		file.maxFileNum = 20
	}

	return &file
}

func (t *FileOutput) WriteLog(log map[string]interface{}) {
	logData, _ := json.MarshalIndent(log, "", "	")
	logData = append(logData, '\n')

	err := os.MkdirAll(t.path, 0755)
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := t.getCurrFile()
	if err != nil {
		fmt.Println(err)
		return
	}

	fileOutputMux.Lock()
	defer fileOutputMux.Unlock()
	defer f.Close()

	_, err = f.Write(logData)
	if err != nil {
		fmt.Println(err)
	}
}

func (t *FileOutput) getCurrFile() (*os.File, error) {
	files, err := t.getSortedFiles()
	if err != nil {
		return nil, err
	}
	var f *os.File
	if len(files) == 0 {
		f, err = t.newFile()
		if err != nil {
			return nil, err
		}
	} else {
		f, err = os.OpenFile(filepath.Join(t.path, files[len(files)-1].Name()), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		fileInfo, err := f.Stat()
		if err != nil {
			return nil, err
		}
		if fileInfo.Size() > 1024*1024*1024 {
			f.Close()

			if len(files) >= t.maxFileNum {
				for i := 0; i < len(files)-t.maxFileNum+1; i++ {
					err := os.Remove(filepath.Join(t.path, files[i].Name()))
					if err != nil {
						return nil, err
					}
				}
			}

			f, err = t.newFile()
			if err != nil {
				return nil, err
			}
		} else {
			if len(files) > t.maxFileNum {
				for i := 0; i < len(files)-t.maxFileNum; i++ {
					err := os.Remove(filepath.Join(t.path, files[i].Name()))
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return f, nil
}

func (t *FileOutput) newFile() (*os.File, error) {
	datetime := time.Now().Format("20060102150405")
	f, err := os.OpenFile(t.path+"pairec.debug."+datetime+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (t *FileOutput) getSortedFiles() ([]os.FileInfo, error) {
	files := []os.FileInfo{}
	err := filepath.Walk(t.path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && path != t.path {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			if strings.HasPrefix(info.Name(), "pairec.debug.") && strings.HasSuffix(info.Name(), ".log") {
				files = append(files, info)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool {
		return t.getFileTimestamp(files[i]) < t.getFileTimestamp(files[j])
	})
	return files, nil
}

func (t *FileOutput) getFileTimestamp(file os.FileInfo) int64 {
	timestampStr := strings.TrimSuffix(strings.TrimPrefix(file.Name(), "pairec.debug."), ".log")
	timestamp, _ := strconv.ParseInt(timestampStr, 10, 64)
	return timestamp
}

type EmptyOutput struct {
}

func (t *EmptyOutput) WriteLog(log map[string]interface{}) {
}

type DebugService struct {
	logFlag     bool
	logOutputer LogOutputer
	requestTime int64
}

func NewDebugService(user *module.User, context *context.RecommendContext) *DebugService {
	service := DebugService{
		logFlag: false,
	}
	var debugConfig recconf.DebugConfig

	found := false
	if context.ExperimentResult != nil {
		data := context.ExperimentResult.GetExperimentParams().GetString("debugConfig", "")
		if data != "" {
			if err := json.Unmarshal([]byte(data), &debugConfig); err == nil {
				found = true
			}
		}
	}

	if !found {
		scene := context.GetParameter("scene").(string)
		if config, ok := context.Config.DebugConfs[scene]; ok {
			found = true
			debugConfig = config
		}
	}

	if !found {
		// not found debug config, not set logflag
		return &service
	}

	if debugConfig.Rate == 100 {
		service.logFlag = true
	} else {
		if len(debugConfig.DebugUsers) > 0 {
			for _, uid := range debugConfig.DebugUsers {
				if uid == string(user.Id) {
					service.logFlag = true
					break
				}
			}
		}

		if !service.logFlag {
			if rand.Intn(100) < debugConfig.Rate {
				service.logFlag = true
			}
		}
	}

	if service.logFlag {
		service.requestTime = time.Now().Unix()
		service.init(&debugConfig)
	}
	return &service
}
func (d *DebugService) init(config *recconf.DebugConfig) {

	switch config.OutputType {
	case "console":
		d.logOutputer = new(ConsoleOutput)

	case "datahub":
		d.logOutputer = NewDatahubOutput(config)

	case "file":
		d.logOutputer = NewFileOutput(config)

	default:
		d.logOutputer = new(EmptyOutput)
	}
}
func (d *DebugService) WriteRecallLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	if d.logFlag {
		triggerMap := make(map[module.ItemId]string, len(items))
		newItems := make([]*module.Item, 0, len(items))
		for _, item := range items {
			triggerMap[item.Id] = item.StringProperty("trigger_id")
			newItem := module.NewItem(string(item.Id))
			newItem.RetrieveId = item.RetrieveId
			newItem.Score = item.Score
			newItems = append(newItems, newItem)
		}
		go d.doWriteRecallLog(user, newItems, context, triggerMap)
	}
}

func (d *DebugService) WriteFilterLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	if d.logFlag {
		go d.doWriteFilterLog(user, items, context)
	}
}

func (d *DebugService) WriteGeneralLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	if d.logFlag {
		newItems := make([]*module.Item, len(items))
		copy(newItems, items)
		go d.doWriteGeneralLog(user, newItems, context)
	}
}

func (d *DebugService) WriteRankLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	if d.logFlag {
		go d.doWriteRankLog(user, items, context)
	}
}

func (d *DebugService) WriteSortLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	if d.logFlag {
		go d.doWriteSortLog(user, items, context)
	}
}

func (d *DebugService) WriteRecommendLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	if d.logFlag {
		go d.doWriteRecommendLog(user, items, context)
	}
}

func (d *DebugService) doWriteRecallLog(user *module.User, items []*module.Item, context *context.RecommendContext, triggerMap map[module.ItemId]string) {
	log := make(map[string]interface{})

	log["request_id"] = context.RecommendId
	log["module"] = "recall"
	log["scene_id"] = context.GetParameter("scene")
	if context.ExperimentResult != nil {
		log["exp_id"] = context.ExperimentResult.GetExpId()
	}
	log["request_time"] = d.requestTime
	log["uid"] = string(user.Id)
	var itemLogInfos []string

	itemsMap := make(map[string][]*module.Item)
	for _, item := range items {
		itemsMap[item.GetRecallName()] = append(itemsMap[item.GetRecallName()], item)
	}

	for name, itemList := range itemsMap {
		log["retrieveid"] = name
		for _, item := range itemList {
			itemLogInfos = append(itemLogInfos, fmt.Sprintf("%s:%f:%s", item.Id, item.Score, triggerMap[item.Id]))
		}
		log["items"] = strings.Join(itemLogInfos, ",")
		d.logOutputer.WriteLog(log)

		itemLogInfos = itemLogInfos[:0]
	}

}

func (d *DebugService) doWriteFilterLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	log := make(map[string]interface{})

	log["request_id"] = context.RecommendId
	log["module"] = "filter"
	log["scene_id"] = context.GetParameter("scene")
	if context.ExperimentResult != nil {
		log["exp_id"] = context.ExperimentResult.GetExpId()
	}
	log["request_time"] = d.requestTime
	log["uid"] = string(user.Id)
	var itemLogInfos []string
	itemsMap := make(map[string][]*module.Item)
	for _, item := range items {
		itemsMap[item.GetRecallName()] = append(itemsMap[item.GetRecallName()], item)
	}

	for name, itemList := range itemsMap {
		log["retrieveid"] = name
		for _, item := range itemList {
			itemLogInfos = append(itemLogInfos, fmt.Sprintf("%s:%f", item.Id, item.Score))
		}
		log["items"] = strings.Join(itemLogInfos, ",")
		d.logOutputer.WriteLog(log)

		itemLogInfos = itemLogInfos[:0]
	}

}

func (d *DebugService) doWriteGeneralLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())
			plog.Error(fmt.Sprintf("error=%v, stack=%s", err, strings.ReplaceAll(stack, "\n", "\t")))
		}
	}()
	logItems := make([]*module.Item, len(items))
	copy(logItems, items)

	log := make(map[string]interface{})

	log["request_id"] = context.RecommendId
	log["module"] = "general_rank"
	log["scene_id"] = context.GetParameter("scene")
	if context.ExperimentResult != nil {
		log["exp_id"] = context.ExperimentResult.GetExpId()
	}
	log["request_time"] = d.requestTime
	log["uid"] = string(user.Id)
	var itemLogInfos []string
	itemsMap := make(map[string][]*module.Item)
	for _, item := range logItems {
		itemsMap[item.GetRecallName()] = append(itemsMap[item.GetRecallName()], item)
	}

	for name, itemList := range itemsMap {
		log["retrieveid"] = name
		for _, item := range itemList {
			if item != nil {
				if b, err := json.Marshal(item.CloneAlgoScores()); err == nil {
					itemLogInfos = append(itemLogInfos, fmt.Sprintf("%s:%f:%s", item.Id, item.Score, string(b)))
				} else {
					itemLogInfos = append(itemLogInfos, fmt.Sprintf("%s:%f", item.Id, item.Score))
				}
			}
		}
		log["items"] = strings.Join(itemLogInfos, ",")
		d.logOutputer.WriteLog(log)
		itemLogInfos = itemLogInfos[:0]
	}

}

func (d *DebugService) doWriteRankLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	log := make(map[string]interface{})

	log["request_id"] = context.RecommendId
	log["module"] = "rank"
	log["scene_id"] = context.GetParameter("scene")
	if context.ExperimentResult != nil {
		log["exp_id"] = context.ExperimentResult.GetExpId()
	}
	log["request_time"] = d.requestTime
	log["uid"] = string(user.Id)
	var itemLogInfos []string
	itemsMap := make(map[string][]*module.Item)
	//gosort.Sort(gosort.Reverse(sort.ItemScoreSlice(items)))
	for _, item := range items {
		itemsMap[item.GetRecallName()] = append(itemsMap[item.GetRecallName()], item)
	}

	for name, itemList := range itemsMap {
		log["retrieveid"] = name
		for _, item := range itemList {
			if b, err := json.Marshal(item.CloneAlgoScores()); err == nil {
				itemLogInfos = append(itemLogInfos, fmt.Sprintf("%s:%f:%s", item.Id, item.Score, string(b)))
			} else {
				itemLogInfos = append(itemLogInfos, fmt.Sprintf("%s:%f", item.Id, item.Score))
			}
		}
		log["items"] = strings.Join(itemLogInfos, ",")
		d.logOutputer.WriteLog(log)
		itemLogInfos = itemLogInfos[:0]
	}

}

func (d *DebugService) doWriteSortLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	log := make(map[string]interface{})

	log["request_id"] = context.RecommendId
	log["module"] = "sort"
	log["scene_id"] = context.GetParameter("scene")
	if context.ExperimentResult != nil {
		log["exp_id"] = context.ExperimentResult.GetExpId()
	}
	log["request_time"] = d.requestTime
	log["uid"] = string(user.Id)
	var itemLogInfos []string
	itemsMap := make(map[string][]*module.Item)
	for _, item := range items {
		itemsMap[item.GetRecallName()] = append(itemsMap[item.GetRecallName()], item)
	}

	for name, itemList := range itemsMap {
		log["retrieveid"] = name
		for _, item := range itemList {
			if b, err := json.Marshal(item.CloneAlgoScores()); err == nil {
				itemLogInfos = append(itemLogInfos, fmt.Sprintf("%s:%f:%s", item.Id, item.Score, string(b)))
			} else {
				itemLogInfos = append(itemLogInfos, fmt.Sprintf("%s:%f", item.Id, item.Score))
			}
		}
		log["items"] = strings.Join(itemLogInfos, ",")
		d.logOutputer.WriteLog(log)
		itemLogInfos = itemLogInfos[:0]
	}
}

func (d *DebugService) doWriteRecommendLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	log := make(map[string]interface{})

	log["request_id"] = context.RecommendId
	log["module"] = "recommend"
	log["scene_id"] = context.GetParameter("scene")
	if context.ExperimentResult != nil {
		log["exp_id"] = context.ExperimentResult.GetExpId()
	}
	log["request_time"] = d.requestTime
	log["uid"] = string(user.Id)
	var itemLogInfos []string
	itemsMap := make(map[string][]*module.Item)
	for _, item := range items {
		itemsMap[item.GetRecallName()] = append(itemsMap[item.GetRecallName()], item)
	}

	for name, itemList := range itemsMap {
		log["retrieveid"] = name
		for _, item := range itemList {
			if b, err := json.Marshal(item.CloneAlgoScores()); err == nil {
				itemLogInfos = append(itemLogInfos, fmt.Sprintf("%s:%f:%s", item.Id, item.Score, string(b)))
			} else {
				itemLogInfos = append(itemLogInfos, fmt.Sprintf("%s:%f", item.Id, item.Score))
			}
		}
		log["items"] = strings.Join(itemLogInfos, ",")
		d.logOutputer.WriteLog(log)
		itemLogInfos = itemLogInfos[:0]
	}
}
