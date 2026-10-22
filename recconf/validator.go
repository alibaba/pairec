package recconf

import (
	"fmt"
	"math"
	"time"
)

type Validator interface {
	Validate() error
}

func (conf FeatureCacheConfig) Validate() error {
	if conf.CacheSize <= 0 {
		return fmt.Errorf("FeatureCache CacheSize must be positive")
	}
	if conf.CacheTime < 0 || int64(conf.CacheTime) > math.MaxInt64/int64(time.Second) {
		return fmt.Errorf("FeatureCache CacheTime is outside the supported duration range")
	}
	return nil
}
