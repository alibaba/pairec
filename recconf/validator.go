package recconf

import (
	"fmt"
	"math"
	"strings"
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
	if strings.TrimSpace(conf.FeatureStoreName) == "" {
		return fmt.Errorf("FeatureCache FeatureStoreName is required")
	}
	if conf.FeatureStore != "user" && conf.FeatureStore != "item" {
		return fmt.Errorf("FeatureCache FeatureStore must be user or item")
	}
	if conf.FeatureStoreViewName != "" {
		if strings.TrimSpace(conf.FeatureStoreViewName) == "" || conf.FeatureStoreModelName != "" || conf.FeatureStoreEntityName != "" {
			return fmt.Errorf("FeatureCache view and model/entity sources are mutually exclusive")
		}
	} else {
		if strings.TrimSpace(conf.FeatureStoreModelName) == "" || strings.TrimSpace(conf.FeatureStoreEntityName) == "" {
			return fmt.Errorf("FeatureCache requires a view or a model with an entity")
		}
		if conf.SelectFields != "" {
			return fmt.Errorf("FeatureCache SelectFields is only supported for a view")
		}
	}
	if conf.FeatureStore == "item" && conf.FeatureStoreOptions.DlrmHSTU {
		return fmt.Errorf("FeatureCache DlrmHSTU is only supported for user features")
	}
	return nil
}
