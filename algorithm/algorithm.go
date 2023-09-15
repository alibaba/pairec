package algorithm

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/alibaba/pairec/algorithm/eas"
	"github.com/alibaba/pairec/algorithm/faiss"
	"github.com/alibaba/pairec/algorithm/seldon"
	"github.com/alibaba/pairec/algorithm/tfserving"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
)

var algoFactory *AlgorithmFactory

func init() {
	algoFactory = NewAlgorithmFactory()
}

// type AlgoData struct {
// Data interface{}
// }

type IAlgorithm interface {
	Init(conf *recconf.AlgoConfig) error
	Run(algoData interface{}) (interface{}, error)
}

type AlgorithmFactory struct {
	algorithms       map[string]IAlgorithm
	requestDataFuncs map[string]RequestDataFunc
	mutex            sync.RWMutex
	algorithmSigns   map[string]string
}

func NewAlgorithmFactory() *AlgorithmFactory {
	factory := AlgorithmFactory{
		algorithmSigns: make(map[string]string, 2),
	}
	factory.algorithms = make(map[string]IAlgorithm)
	factory.requestDataFuncs = make(map[string]RequestDataFunc)

	return &factory
}
func (a *AlgorithmFactory) Init(algoConfs []recconf.AlgoConfig) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for _, conf := range algoConfs {
		sign, _ := json.Marshal(conf)
		if _, ok := a.algorithms[conf.Name]; ok {
			if utils.Md5(string(sign)) == a.algorithmSigns[conf.Name] {
				continue
			}
		}
		algo, err := a.initAlgo(conf)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		a.algorithms[conf.Name] = algo
		a.algorithmSigns[conf.Name] = utils.Md5(string(sign))
	}
}
func (a *AlgorithmFactory) initAlgo(conf recconf.AlgoConfig) (IAlgorithm, error) {
	var algo IAlgorithm
	if conf.Type == "EAS" {
		algo = eas.NewEasModel(conf.Name)
		err := algo.Init(&conf)
		if err != nil {
			return nil, fmt.Errorf("init algorithm error, name:%s, err:%v", conf.Name, err)
		}
	} else if conf.Type == "FAISS" {
		algo = faiss.NewFaissModel(conf.Name)
		err := algo.Init(&conf)
		if err != nil {
			return nil, fmt.Errorf("init algorithm error, name:%s, err:%v", conf.Name, err)
		}
	} else if conf.Type == "LOOKUP" {
		algo = NewLookupPolicy()
		err := algo.Init(&conf)
		if err != nil {
			return nil, fmt.Errorf("init algorithm error, name:%s, err:%v", conf.Name, err)
		}
	} else if conf.Type == "SELDON" {
		algo = new(seldon.Model)
		err := algo.Init(&conf)
		if err != nil {
			return nil, fmt.Errorf("init algorithm error, name:%s, err:%v", conf.Name, err)
		}
	} else if conf.Type == "TFSERVING" {
		algo = tfserving.NewTFservingModel(conf.Name)
		err := algo.Init(&conf)
		if err != nil {
			return nil, fmt.Errorf("init algorithm error, name:%s, err:%v", conf.Name, err)
		}

	} else {
		return nil, fmt.Errorf("algorithm type not support , type:%s", conf.Type)
	}
	return algo, nil
}
func (a *AlgorithmFactory) Run(name string, algoData interface{}) (interface{}, error) {
	a.mutex.RLock()
	algo, found := a.algorithms[name]
	f, funcFound := a.requestDataFuncs[name]
	a.mutex.RUnlock()
	if !found {
		return nil, errors.New("not found algorithm, name:" + name)
	}
	// if find request data func
	if funcFound {
		return algo.Run(f(name, algoData))
	}
	return algo.Run(algoData)
}

// init algorithm from the config, and add to the algoFactory
func Load(config *recconf.RecommendConfig) {
	algoFactory.Init(config.AlgoConfs)
}
func Run(name string, algoData interface{}) (interface{}, error) {
	return algoFactory.Run(name, algoData)
}
func AddAlgo(conf recconf.AlgoConfig) {
	algoFactory.mutex.Lock()
	defer algoFactory.mutex.Unlock()
	_, found := algoFactory.algorithms[conf.Name]
	if found {
		return
	}
	algo, err := algoFactory.initAlgo(conf)
	if err != nil {
		log.Error(err.Error())
		return
	}
	algoFactory.algorithms[conf.Name] = algo
}

func RegisterAlgorithm(name string, algo IAlgorithm) {
	algoFactory.mutex.Lock()
	defer algoFactory.mutex.Unlock()
	algoFactory.algorithms[name] = algo
}

type RequestDataFunc func(string, interface{}) interface{}

func RegistRequestDataFunc(name string, f RequestDataFunc) {
	algoFactory.requestDataFuncs[name] = f
}
