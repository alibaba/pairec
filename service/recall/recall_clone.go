package recall

// CloneWithConfig methods enable AB experiment parameter overrides for recall
// instances. Mirroring the sort.ICloneSort pattern, every supported recall
// implements ICloneRecall (CloneWithConfig + GetRecallName).
//
// AB params are deserialized into a fresh RecallConfig (same semantics as
// sort.CloneWithConfig). The AB configuration must provide all required fields;
// only the Name field is preserved from the original recall.
// The config produces an md5 cache key, so identical AB configs reuse
// the same cloned recall instance (cached inside the original instance).

import (
	"github.com/alibaba/pairec/v2/recconf"
)

// UserCollaborativeFilterRecall

func (r *UserCollaborativeFilterRecall) CloneWithConfig(params map[string]interface{}) Recall {
	if newR := r.cloneWithBuilder(params, func(cfg recconf.RecallConfig) Recall {
		return NewUserCollaborativeFilterRecall(cfg)
	}); newR != nil {
		return newR
	}
	return r
}

// RealTimeU2IRecall

func (r *RealTimeU2IRecall) CloneWithConfig(params map[string]interface{}) Recall {
	if newR := r.cloneWithBuilder(params, func(cfg recconf.RecallConfig) Recall {
		return NewRealTimeU2IRecall(cfg)
	}); newR != nil {
		return newR
	}
	return r
}

// HologresVectorRecall — not supported for AB cloning because its constructor
// starts a background goroutine (partition polling) that cannot be stopped,
// which would cause goroutine leaks on each clone.

// OnlineHologresVectorRecall

func (r *OnlineHologresVectorRecall) CloneWithConfig(params map[string]interface{}) Recall {
	if newR := r.cloneWithBuilder(params, func(cfg recconf.RecallConfig) Recall {
		return NewOnlineHologresVectorRecall(cfg)
	}); newR != nil {
		return newR
	}
	return r
}

// OnlineVectorRecall

func (r *OnlineVectorRecall) CloneWithConfig(params map[string]interface{}) Recall {
	if newR := r.cloneWithBuilder(params, func(cfg recconf.RecallConfig) Recall {
		return NewOnlineVectorRecall(cfg)
	}); newR != nil {
		return newR
	}
	return r
}

// UserCustomRecall

func (r *UserCustomRecall) CloneWithConfig(params map[string]interface{}) Recall {
	if newR := r.cloneWithBuilder(params, func(cfg recconf.RecallConfig) Recall {
		return NewUserCustomRecall(cfg)
	}); newR != nil {
		return newR
	}
	return r
}

// GraphRecall

func (r *GraphRecall) CloneWithConfig(params map[string]interface{}) Recall {
	if newR := r.cloneWithBuilder(params, func(cfg recconf.RecallConfig) Recall {
		return NewGraphRecall(cfg)
	}); newR != nil {
		return newR
	}
	return r
}

// UserGroupHotRecall

func (r *UserGroupHotRecall) CloneWithConfig(params map[string]interface{}) Recall {
	if newR := r.cloneWithBuilder(params, func(cfg recconf.RecallConfig) Recall {
		return NewUserGroupHotRecall(cfg)
	}); newR != nil {
		return newR
	}
	return r
}

// UserGlobalHotRecall

func (r *UserGlobalHotRecall) CloneWithConfig(params map[string]interface{}) Recall {
	if newR := r.cloneWithBuilder(params, func(cfg recconf.RecallConfig) Recall {
		return NewUserGlobalHotRecall(cfg)
	}); newR != nil {
		return newR
	}
	return r
}

// ColdStartRecall — not supported for AB cloning because its constructor
// starts a background goroutine (LoopLoadItems) that cannot be stopped,
// which would cause goroutine leaks on each clone.

// ContextItemRecall

func (r *ContextItemRecall) CloneWithConfig(params map[string]interface{}) Recall {
	if newR := r.cloneWithBuilder(params, func(cfg recconf.RecallConfig) Recall {
		return NewContextItemRecall(cfg)
	}); newR != nil {
		return newR
	}
	return r
}

// ItemCollaborativeFilterRecall

func (r *ItemCollaborativeFilterRecall) CloneWithConfig(params map[string]interface{}) Recall {
	if newR := r.cloneWithBuilder(params, func(cfg recconf.RecallConfig) Recall {
		return NewItemCollaborativeFilterRecall(cfg)
	}); newR != nil {
		return newR
	}
	return r
}

// OpenSearchRecall

func (r *OpenSearchRecall) CloneWithConfig(params map[string]interface{}) Recall {
	if newR := r.cloneWithBuilder(params, func(cfg recconf.RecallConfig) Recall {
		return NewOpenSearchRecall(cfg)
	}); newR != nil {
		return newR
	}
	return r
}
