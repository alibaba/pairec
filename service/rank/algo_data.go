package rank

import (
	"reflect"
	"sync"

	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/algorithm/response"
	"github.com/alibaba/pairec/v2/module"
)

type IAlgoData interface {
	GetFeatures() interface{}
	SetError(err error)
	Error() error
	GetItems() []*module.Item
	GetAlgoResult() map[string][]response.AlgoResponse
	SetAlgoResult(algoName string, results []response.AlgoResponse)
}

type IAlgoDataGenerator interface {
	AddFeatures(item *module.Item, itemFeatures map[string]interface{}, userFeatures map[string]interface{})
	GeneratorAlgoData() IAlgoData
	// GeneratorAlgoDataDebug generator algo data with debug mode
	GeneratorAlgoDataDebug() IAlgoData
	GeneratorAlgoDataDebugWithLevel(level int) IAlgoData
	HasFeatures() bool
}

func CreateAlgoDataGenerator(processor string, contextFeatures []string) IAlgoDataGenerator {

	if processor == eas.Eas_Processor_EASYREC {
		return NewEasyrecAlgoDataGenerator(contextFeatures)
	} else {
		return NewAlgoDataGenerator()
	}
}

type AlgoDataBase struct {
	Items []*module.Item
	// RequestData []map[string]interface{}
	AlgoResult map[string][]response.AlgoResponse
	Err        error
	Mutex      sync.Mutex
}

func (a *AlgoDataBase) SetError(err error) {
	a.Err = err
}
func (a *AlgoDataBase) Error() error {
	return a.Err
}
func (a *AlgoDataBase) SetAlgoResult(algoName string, results []response.AlgoResponse) {

	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	a.AlgoResult[algoName] = results
}
func (a *AlgoDataBase) GetAlgoResult() map[string][]response.AlgoResponse {
	return a.AlgoResult
}

func (a *AlgoDataBase) GetItems() []*module.Item {
	return a.Items
}

type AlgoData struct {
	*AlgoDataBase
	RequestData []map[string]interface{}
}

func (d *AlgoData) GetFeatures() interface{} {
	return d.RequestData
}

type EasyrecAlgoData struct {
	*AlgoDataBase
	easyrecRequest *easyrec.PBRequest
}

func (d *EasyrecAlgoData) GetFeatures() interface{} {
	return d.easyrecRequest
}

type AlgoDataGenerator struct {
	requestData []map[string]interface{}
	requestItem []*module.Item
}

func NewAlgoDataGenerator() *AlgoDataGenerator {

	return &AlgoDataGenerator{
		requestData: make([]map[string]interface{}, 0, 100),
		requestItem: make([]*module.Item, 0, 100),
	}
}
func (g *AlgoDataGenerator) AddFeatures(item *module.Item, itemFeatures map[string]interface{}, userFeatures map[string]interface{}) {
	if item != nil {
		g.requestItem = append(g.requestItem, item)
	}

	features := make(map[string]interface{}, len(itemFeatures)+len(userFeatures))
	for k, v := range userFeatures {
		features[k] = v
	}
	for k, v := range itemFeatures {
		features[k] = v
	}

	g.requestData = append(g.requestData, features)
}
func (g *AlgoDataGenerator) GeneratorAlgoData() IAlgoData {
	copydata := make([]map[string]interface{}, len(g.requestData))
	copyItems := make([]*module.Item, len(g.requestItem))
	copy(copyItems, g.requestItem)
	copy(copydata, g.requestData)

	algoData := &AlgoData{
		AlgoDataBase: &AlgoDataBase{
			Items:      copyItems,
			AlgoResult: make(map[string][]response.AlgoResponse),
		},
		RequestData: copydata,
	}

	g.requestData = g.requestData[:0]
	g.requestItem = g.requestItem[:0]
	return algoData
}
func (g *AlgoDataGenerator) GeneratorAlgoDataDebug() IAlgoData {
	return g.GeneratorAlgoData()
}

func (g *AlgoDataGenerator) GeneratorAlgoDataDebugWithLevel(level int) IAlgoData {
	return g.GeneratorAlgoData()
}

func (g *AlgoDataGenerator) HasFeatures() bool {
	return len(g.requestData) > 0
}

type feature struct {
	name      string
	valueType reflect.Type
}

func (f *feature) defaultValue() interface{} {
	switch f.valueType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(0)
	case reflect.Float32, reflect.Float64:
		return float64(0)
	case reflect.String:
		return ""
	default:
		return ""
	}
}

type EasyrecAlgoDataGenerator struct {
	requestItem     []*module.Item
	contextFeatures map[string][]interface{}
	parseFeature    bool
	itemFeatures    []*feature
	userFeatures    map[string]interface{}
}

func NewEasyrecAlgoDataGenerator(contextFeatures []string) *EasyrecAlgoDataGenerator {
	generator := &EasyrecAlgoDataGenerator{
		requestItem:     make([]*module.Item, 0, 100),
		contextFeatures: make(map[string][]interface{}, 8),
		parseFeature:    false,
	}

	if len(contextFeatures) > 0 {
		for _, featureName := range contextFeatures {
			feature := &feature{
				name:      featureName,
				valueType: reflect.TypeOf(""),
			}
			generator.itemFeatures = append(generator.itemFeatures, feature)
		}

		generator.parseFeature = true
	}
	return generator
}

func (g *EasyrecAlgoDataGenerator) AddFeatures(item *module.Item, itemFeatures map[string]interface{}, userFeatures map[string]interface{}) {
	if item != nil {
		g.requestItem = append(g.requestItem, item)
	}
	if !g.parseFeature {
		for k, v := range itemFeatures {
			feature := &feature{
				name:      k,
				valueType: reflect.TypeOf(v),
			}
			g.itemFeatures = append(g.itemFeatures, feature)
		}

		g.userFeatures = userFeatures
		g.parseFeature = true
	}
	if len(g.userFeatures) == 0 {
		g.userFeatures = userFeatures
	}

	for _, f := range g.itemFeatures {
		if v, ok := itemFeatures[f.name]; ok {
			g.contextFeatures[f.name] = append(g.contextFeatures[f.name], v)
		} else {
			g.contextFeatures[f.name] = append(g.contextFeatures[f.name], f.defaultValue())
		}
	}

}

func (g *EasyrecAlgoDataGenerator) GeneratorAlgoData() IAlgoData {
	copyItems := make([]*module.Item, len(g.requestItem))
	copy(copyItems, g.requestItem)

	builder := easyrec.NewEasyrecRequestBuilder()
	for k, v := range g.userFeatures {
		builder.AddUserFeature(k, v)
	}
	for _, item := range g.requestItem {
		builder.AddItemId(string(item.Id))
	}
	for k, v := range g.contextFeatures {
		builder.AddContextFeature(k, v)
		g.contextFeatures[k] = g.contextFeatures[k][:0]
	}

	algoData := &EasyrecAlgoData{
		AlgoDataBase: &AlgoDataBase{
			Items:      copyItems,
			AlgoResult: make(map[string][]response.AlgoResponse),
		},
		easyrecRequest: builder.EasyrecRequest(),
	}

	g.requestItem = g.requestItem[:0]
	return algoData
}

func (g *EasyrecAlgoDataGenerator) GeneratorAlgoDataDebug() IAlgoData {
	return g.GeneratorAlgoDataDebugWithLevel(1)
}

func (g *EasyrecAlgoDataGenerator) GeneratorAlgoDataDebugWithLevel(level int) IAlgoData {
	copyItems := make([]*module.Item, len(g.requestItem))
	copy(copyItems, g.requestItem)

	builder := easyrec.NewEasyrecRequestBuilderDebugWithLevel(level)
	for k, v := range g.userFeatures {
		builder.AddUserFeature(k, v)
	}
	for _, item := range g.requestItem {
		builder.AddItemId(string(item.Id))
	}
	for k, v := range g.contextFeatures {
		builder.AddContextFeature(k, v)
		g.contextFeatures[k] = g.contextFeatures[k][:0]
	}

	algoData := &EasyrecAlgoData{
		AlgoDataBase: &AlgoDataBase{
			Items:      copyItems,
			AlgoResult: make(map[string][]response.AlgoResponse),
		},
		easyrecRequest: builder.EasyrecRequest(),
	}

	g.requestItem = g.requestItem[:0]
	return algoData
}

func (g *EasyrecAlgoDataGenerator) HasFeatures() bool {
	return len(g.requestItem) > 0
}
