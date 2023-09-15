package config

import "fmt"

type Configer interface {
	Set(key, val string) error
	Get(key string) (interface{}, error)
	Int(key string) (int, error)
	Int64(key string) (int64, error)
	Bool(key string) (bool, error)
	String(key string) (string, error)
	Float64(key string) (float64, error)
	RawData() []byte
}
type Config interface {
	ParseFile(file string) (Configer, error)
}

var adapters = make(map[string]Config)

func Register(adapterName string, adapter Config) {
	adapters[adapterName] = adapter
}

func NewConfig(adapterName, file string) (Configer, error) {
	adapter, ok := adapters[adapterName]
	if !ok {
		return nil, fmt.Errorf("Config:%s not exist", adapterName)
	}

	return adapter.ParseFile(file)
}
