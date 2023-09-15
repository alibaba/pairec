package log

import (
	"fmt"

	"github.com/golang/glog"
)

func Debug(msg string) {
	glog.V(1).Info(fmt.Sprintf("[DEBUG]\t%s", msg))
}
func Info(msg string) {
	glog.InfoDepth(1, fmt.Sprintf("[INFO]\t%s", msg))
	if slsclient != nil {
		sendLogToSLS("INFO", msg)
	}
}
func Warning(msg string) {
	glog.InfoDepth(1, fmt.Sprintf("[WARNING]\t%s", msg))
	if slsclient != nil {
		sendLogToSLS("WARNING", msg)
	}
}
func Error(msg string) {
	glog.InfoDepth(1, fmt.Sprintf("[ERROR]\t%s", msg))
	if slsclient != nil {
		sendLogToSLS("ERROR", msg)
	}
}
func Flush() {
	glog.Flush()
}

type ABTestLogger struct{}

func (l ABTestLogger) Infof(msg string, args ...interface{}) {
	Info(fmt.Sprintf(msg, args...))
}
func (l ABTestLogger) Errorf(msg string, args ...interface{}) {
	Error(fmt.Sprintf(msg, args...))
}

type FeatureStoreLogger struct{}

func (l FeatureStoreLogger) Infof(msg string, args ...interface{}) {
	Info(fmt.Sprintf(msg, args...))
}
func (l FeatureStoreLogger) Errorf(msg string, args ...interface{}) {
	Error(fmt.Sprintf(msg, args...))
}
