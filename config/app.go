package config

import "sync"

var AppConfig *AppConf

func init() {
	AppConfig = &AppConf{}
}

type AppConf struct {
	WarmUpData bool
	Mu         sync.Mutex
	Once       sync.Once
}
