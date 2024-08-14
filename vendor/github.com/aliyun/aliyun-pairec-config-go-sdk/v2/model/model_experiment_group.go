package model

import (
	"strings"
)

type ExperimentGroup struct {
	ExpGroupId               int64  `json:"exp_group_id,omitempty"`
	LayerId                  int64  `json:"layer_id"`
	ExpRoomId                int64  `json:"exp_room_id,omitempty"`
	SceneId                  int64  `json:"scene_id,omitempty"`
	ExpGroupName             string `json:"exp_group_name"`
	ExpGroupInfo             string `json:"exp_group_info"`
	DebugUsers               string `json:"debug_users,omitempty"`
	DebugCrowdId             int64  `json:"debug_crowd_id,omitempty"`
	Owner                    string `json:"owner"`
	Filter                   string `json:"filter,omitempty"`
	DistributionType         int    `json:"distribution_type,omitempty"`
	DistributionTimeDuration int    `json:"distribution_time_duration,omitempty"`
	CrowdId                  int64  `json:"crowd_id,omitempty"`
	ExpGroupConfig           string `json:"exp_group_config,omitempty"`
	ReserveBuckets           string `json:"reserve_buckets,omitempty"`
	Status                   int32  `json:"status,omitempty"`

	Experiments     []*Experiment   `json:"experiments"`
	debugUserMap    map[string]bool `json:"-"`
	DebugCrowdUsers []string        `json:"debug_crowd_users"`
	diversionBucket DiversionBucket
	CrowdUsers      []string
	crowdUserMap    map[string]struct{}
}

func (e *ExperimentGroup) Init() error {
	if e.Filter != "" {
		diversionBucket, err := NewFilterDiversionBucket(e.Filter)
		if err != nil {
			return err
		}

		e.diversionBucket = diversionBucket
	}
	// deal DebugUsers
	e.debugUserMap = make(map[string]bool, 0)
	if e.DebugUsers != "" {
		uids := strings.Split(e.DebugUsers, ",")
		for _, uid := range uids {
			e.debugUserMap[uid] = true
		}
	}
	if len(e.DebugCrowdUsers) != 0 {
		for _, uid := range e.DebugCrowdUsers {
			e.debugUserMap[uid] = true
		}
	}

	e.crowdUserMap = make(map[string]struct{}, len(e.CrowdUsers))
	for _, uid := range e.CrowdUsers {
		e.crowdUserMap[uid] = struct{}{}
	}

	return nil
}
func (e *ExperimentGroup) AddExperiment(experiment *Experiment) {
	e.Experiments = append(e.Experiments, experiment)
}

func (e *ExperimentGroup) Match(experimentContext *ExperimentContext) bool {

	if e.DebugUsers == "" && e.Filter == "" && e.CrowdId == 0 {
		return true
	}

	if e.Filter == "" && e.CrowdId != 0 {
		if _, found := e.crowdUserMap[experimentContext.Uid]; found {
			return true
		}
	}

	if _, found := e.debugUserMap[experimentContext.Uid]; found {
		return true
	}

	if e.diversionBucket != nil {
		return e.diversionBucket.Match(experimentContext)
	}

	return false
}
