package log

import (
	"strings"

	"github.com/alibaba/pairec/v2/datasource/sls"
)

var (
	slsclient *sls.SlsClient
)

func RegisterSlsClient(client *sls.SlsClient) {
	slsclient = client
}

type LogInfo struct {
	Level   string
	Message string
}

func sendLogToSLS(level string, message string) {

	strs := strings.Split(message, "\t")

	logMap := map[string]string{
		"level": level,
	}

	for _, str := range strs {
		params := strings.Split(str, "=")
		if len(params) == 2 {
			logMap[params[0]] = params[1]
		}
	}

	if len(logMap) >= 2 {
		slsclient.SendLog(logMap)
	}
}

/**
func writeLogToSLS(logCh <-chan *LogInfo) {
	for {
		select {
		case logInfo := <-logCh:

		default:
			time.Sleep(10*second)
		}
	}
}
**/
