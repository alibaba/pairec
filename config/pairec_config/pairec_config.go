package pairec_config

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/alibaba/pairec/abtest"
	"github.com/alibaba/pairec/config"
)

var (
	Pairec_Config_Scene_Name = "pairec_config_manage"
)

func init() {
	config.Register("pairec_config", &PairecConfig{})
}

type PairecConfig struct{}

type PairecConfigContainer struct {
	data    map[string]interface{}
	rawData []byte
	sync.RWMutex
}

/**
* ParseFile load config from pairec config server
* configName the name of pairec config server parameter name
 */
func (js *PairecConfig) ParseFile(configName string) (config.Configer, error) {

	configData := abtest.GetParams(Pairec_Config_Scene_Name).GetString(configName, "")
	if configData == "" {
		// if configData empty, set the default value
		configData = "{\"ListenConf\":{\"HttpAddr\":\"\", \"HttpPort\":8000}}"
	}

	return js.ParseData([]byte(configData))
}

func (js *PairecConfig) ParseData(data []byte) (config.Configer, error) {
	x := &PairecConfigContainer{
		rawData: data,
		data:    make(map[string]interface{}),
	}
	err := json.Unmarshal(data, &x.data)
	if err != nil {
		return nil, err
	}

	return x, nil
}

func (c *PairecConfigContainer) Set(key, val string) error {
	c.Lock()
	defer c.Unlock()
	c.data[key] = val
	return nil
}
func (c *PairecConfigContainer) Get(key string) (interface{}, error) {
	c.RLock()
	defer c.RUnlock()

	val, ok := c.data[key]
	if !ok {
		return nil, fmt.Errorf("Config:not find key:%s value", key)
	}
	return val, nil
}
func (c *PairecConfigContainer) String(key string) (string, error) {
	val, err := c.Get(key)
	if err != nil {
		return "", err
	}
	if v, ok := val.(string); ok {
		return v, nil
	}
	return "", errors.New("not string value")
}
func (c *PairecConfigContainer) Int64(key string) (int64, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	if v, ok := val.(float64); ok {
		return int64(v), nil
	}
	return 0, errors.New("not int64 value")
}
func (c *PairecConfigContainer) Int(key string) (int, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	if v, ok := val.(float64); ok {
		return int(v), nil
	}
	return 0, errors.New("not int value")
}
func (c *PairecConfigContainer) Float64(key string) (float64, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0.0, err
	}
	if v, ok := val.(float64); ok {
		return v, nil
	}
	return 0.0, errors.New("not float64 value")
}
func (c *PairecConfigContainer) Bool(key string) (bool, error) {
	val, err := c.Get(key)
	if err != nil {
		return false, err
	}

	return config.ParseBool(val)
}
func (c *PairecConfigContainer) RawData() []byte {
	return c.rawData
}
