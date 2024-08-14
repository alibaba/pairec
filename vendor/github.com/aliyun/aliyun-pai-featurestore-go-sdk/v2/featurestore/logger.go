package featurestore

// Logger interface API for log.Logger
type Logger interface {
	Printf(string, ...interface{})
}

// LoggerFunc is a bridge between Logger and any third party logger
// Usage:

// l := NewLogger() // some logger
// client, err := NewFeatureStoreClient(host, env,
// WithLogger(featurestore.LoggerFunc(l.Infof)),
// WithErrorLogger(featurestore.LoggerFunc(l.Errorf)))

type LoggerFunc func(string, ...interface{})

func (f LoggerFunc) Printf(msg string, args ...interface{}) { f(msg, args...) }
