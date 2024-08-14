package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

func init() {
	Register("json", &JsonConfig{})
}

type JsonConfig struct{}

type JsonConfigContainer struct {
	data    map[string]interface{}
	rawData []byte
	sync.RWMutex
}

func (js *JsonConfig) ParseFile(filename string) (Configer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return js.ParseData(content)
}

func (js *JsonConfig) ParseData(data []byte) (Configer, error) {
	x := &JsonConfigContainer{
		rawData: data,
		data:    make(map[string]interface{}),
	}
	err := json.Unmarshal(data, &x.data)
	if err != nil {
		return nil, err
	}

	return x, nil
}

func (c *JsonConfigContainer) Set(key, val string) error {
	c.Lock()
	defer c.Unlock()
	c.data[key] = val
	return nil
}
func (c *JsonConfigContainer) Get(key string) (interface{}, error) {
	c.Lock()
	defer c.Unlock()
	val, ok := c.data[key]
	if !ok {
		return nil, fmt.Errorf("Config:not find key:%s value", key)
	}
	return val, nil
}
func (c *JsonConfigContainer) String(key string) (string, error) {
	val, err := c.Get(key)
	if err != nil {
		return "", err
	}
	if v, ok := val.(string); ok {
		return v, nil
	}
	return "", errors.New("not string value")
}
func (c *JsonConfigContainer) Int64(key string) (int64, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	if v, ok := val.(float64); ok {
		return int64(v), nil
	}
	return 0, errors.New("not int64 value")
}
func (c *JsonConfigContainer) Int(key string) (int, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	if v, ok := val.(float64); ok {
		return int(v), nil
	}
	return 0, errors.New("not int value")
}
func (c *JsonConfigContainer) Float64(key string) (float64, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0.0, err
	}
	if v, ok := val.(float64); ok {
		return v, nil
	}
	return 0.0, errors.New("not float64 value")
}
func (c *JsonConfigContainer) Bool(key string) (bool, error) {
	val, err := c.Get(key)
	if err != nil {
		return false, err
	}

	return ParseBool(val)
}
func ParseBool(val interface{}) (value bool, err error) {
	if val != nil {
		switch v := val.(type) {
		case bool:
			return v, nil
		case string:
			switch v {
			case "1", "t", "T", "true", "TRUE", "True", "YES", "yes", "Yes", "Y", "y", "ON", "on", "On":
				return true, nil
			case "0", "f", "F", "false", "FALSE", "False", "NO", "no", "No", "N", "n", "OFF", "off", "Off":
				return false, nil
			}
		case int8, int32, int64:
			strV := fmt.Sprintf("%d", v)
			if strV == "1" {
				return true, nil
			} else if strV == "0" {
				return false, nil
			}
		case float64:
			if v == 1.0 {
				return true, nil
			} else if v == 0.0 {
				return false, nil
			}
		}
		return false, fmt.Errorf("parsing %q: invalid syntax", val)
	}
	return false, fmt.Errorf("parsing <nil>: invalid syntax")
}
func (c *JsonConfigContainer) RawData() []byte {
	return c.rawData
}
