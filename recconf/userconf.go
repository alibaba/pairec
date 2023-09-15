package recconf

import (
	"encoding/json"
	"errors"
	"fmt"
)

type parseSetting struct {
	Key string
}

type ParseOp func(setting *parseSetting)

var WithKey = func(key string) ParseOp {
	return func(setting *parseSetting) {
		setting.Key = key
	}
}

func ParseUserDefineConfs[T any](ops ...ParseOp) (T, error) {
	var setting parseSetting
	for _, op := range ops {
		op(&setting)
	}

	var t T

	if Config == nil {
		return t, errors.New(fmt.Sprintf("config has not been set"))
	}

	if setting.Key == "" {
		err := json.Unmarshal(Config.UserDefineConfs, &t)
		return t, err
	} else {
		m := make(map[string]interface{})

		err := json.Unmarshal(Config.UserDefineConfs, &m)
		if err != nil {
			return t, err
		}

		t, ok := m[setting.Key].(T)
		if !ok {
			return t, errors.New(fmt.Sprintf("value of %s can not be resolved", setting.Key))
		} else {
			return t, nil
		}
	}
}
