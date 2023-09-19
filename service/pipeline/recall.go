package pipeline

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/recall"
	"github.com/alibaba/pairec/v2/utils"
)

var mutex sync.Mutex

type RecallService struct {
	pipelineName string
	recallNames  []string
}

func NewRecallService(config *recconf.PipelineConfig) *RecallService {
	service := RecallService{
		pipelineName: config.Name,
		recallNames:  config.RecallNames,
	}

	return &service
}
func getRecallFromABConfig(recallConfig map[string]interface{}, name, recallNewName string) recall.Recall {

	oldRecall, err := recall.GetRecall(name)
	if err != nil {
		return nil
	}

	mutex.Lock()
	if newRecall, err := recall.GetRecall(recallNewName); err == nil {
		mutex.Unlock()
		return newRecall
	}

	if m := reflect.ValueOf(oldRecall).MethodByName("CloneWithConfig"); m.IsValid() {
		if callValues := m.Call([]reflect.Value{reflect.ValueOf(recallConfig)}); len(callValues) == 1 {
			i := callValues[0].Interface()
			if newRecall, ok := i.(recall.Recall); ok {
				recall.RegisterRecall(recallNewName, newRecall)
				log.Info("register recall :" + recallNewName)
				mutex.Unlock()
				return newRecall
			}
		}
	}

	mutex.Unlock()
	return nil
}
func (s *RecallService) GetItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	var recalls []recall.Recall
	var recallNames []string
	found := false
	if context.ExperimentResult != nil {
		names := context.ExperimentResult.GetExperimentParams().Get("pipelines."+s.pipelineName+".RecallNames", nil)
		if names != nil {
			found = true
			if values, ok := names.([]interface{}); ok {
				for _, v := range values {
					if r, ok := v.(string); ok {
						recallNames = append(recallNames, r)
					}
				}
			}
		}
	}

	if !found {
		recallNames = s.recallNames

	}

	for _, name := range recallNames {
		if context.ExperimentResult != nil {
			recallConfig := context.ExperimentResult.GetExperimentParams().Get("recall."+name, nil)
			if recallConfig == nil {
				if recall, err := recall.GetRecall(name); err == nil {
					recalls = append(recalls, recall)
				}

			} else {
				d, _ := json.Marshal(recallConfig)
				recallName := name + "#" + utils.Md5(string(d))
				// find new recall by the new recall name
				if recall, err := recall.GetRecall(recallName); err == nil {
					recalls = append(recalls, recall)
				} else {
					if params, ok := recallConfig.(map[string]interface{}); ok {
						if r := getRecallFromABConfig(params, name, recallName); r != nil {
							recalls = append(recalls, r)
						}
					}
				}

			}
		} else {
			// not find abtest config
			if recall, err := recall.GetRecall(name); err == nil {
				recalls = append(recalls, recall)
			}
		}
	}

	ch := make(chan []*module.Item, len(recalls))

	for i := 0; i < len(recalls); i++ {
		go func(ch chan<- []*module.Item, recall recall.Recall) {
			// when recall is panic, can recover it
			defer func() {
				if err := recover(); err != nil {
					stack := string(debug.Stack())
					log.Error(fmt.Sprintf("error=%v, stack=%s", err, strings.ReplaceAll(stack, "\n", "\t")))

					var tmp []*module.Item
					ch <- tmp
				}
			}()

			items := recall.GetCandidateItems(user, context)
			ch <- items
		}(ch, recalls[i])
	}
	for i := 0; i < len(recalls); i++ {
		items := <-ch
		ret = append(ret, items...)
	}
	close(ch)
	return
}
