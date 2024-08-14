package model

import (
	"strings"
)

type Experiment struct {
	ExperimentId      int64  `json:"experiment_id,omitempty"`
	ExpGroupId        int64  `json:"exp_group_id"`
	LayerId           int64  `json:"layer_id,omitempty"`
	ExpRoomId         int64  `json:"exp_room_id,omitempty"`
	SceneId           int64  `json:"scene_id,omitempty"`
	ExperimentName    string `json:"experiment_name"`
	ExperimentInfo    string `json:"experiment_info"`
	Type              uint32 `json:"type"`
	ExperimentFlow    uint32 `json:"experiment_flow,omitempty"`
	ExperimentBuckets string `json:"experiment_buckets,omitempty"`
	DebugUsers        string `json:"debug_users,omitempty"`
	DebugCrowdId      int64  `json:"debug_crowd_id,omitempty"`
	ExperimentConfig  string `json:"experiment_config,omitempty"`
	Status            int32  `json:"status,omitempty"`

	DebugCrowdUsers []string        `json:"debug_crowd_users"`
	debugUserMap    map[string]bool `json:"-"`
	diversionBucket DiversionBucket `json:"-"`
}

// Init is a function of init experiment data
func (e *Experiment) Init() error {
	// deal DebugUsers
	e.debugUserMap = make(map[string]bool, 0)
	if e.DebugUsers != "" {
		uids := strings.Split(e.DebugUsers, ",")
		for _, uid := range uids {
			e.debugUserMap[uid] = true
		}
	}
	if len(e.DebugCrowdUsers) > 0 {
		for _, uid := range e.DebugCrowdUsers {
			e.debugUserMap[uid] = true
		}
	}

	if e.ExperimentFlow > 0 && e.ExperimentFlow < 100 {
		e.diversionBucket = NewUidDiversionBucket(100, e.ExperimentBuckets)
	}

	return nil
}

// MatchDebugUsers return true if debug_users is set and debug_users contain of uid
func (e *Experiment) MatchDebugUsers(experimentContext *ExperimentContext) bool {
	if _, found := e.debugUserMap[experimentContext.Uid]; found {
		return true
	}

	return false
}

func (e *Experiment) Match(experimentContext *ExperimentContext) bool {

	if e.ExperimentFlow == 0 {
		return false
	}
	if e.ExperimentFlow == 100 {
		return true
	}

	if _, found := e.debugUserMap[experimentContext.Uid]; found {
		return true
	}

	if e.diversionBucket != nil {
		return e.diversionBucket.Match(&ExperimentContext{Uid: experimentContext.ExperimentHashString()})
	}

	return false
}

func (e *Experiment) Clone() *Experiment {
	exp := Experiment{
		ExperimentId:     e.ExperimentId,
		ExpGroupId:       e.ExpGroupId,
		ExpRoomId:        e.ExpRoomId,
		SceneId:          e.SceneId,
		LayerId:          e.LayerId,
		Status:           e.Status,
		ExperimentConfig: e.ExperimentConfig,
	}

	return &exp
}
