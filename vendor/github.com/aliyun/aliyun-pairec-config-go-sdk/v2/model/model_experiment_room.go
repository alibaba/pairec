package model

import (
	"strings"

	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/common"
)

type ExperimentRoom struct {
	ExpRoomId      int64  `json:"exp_room_id,omitempty"`
	SceneId        int64  `json:"scene_id"`
	ExpRoomName    string `json:"exp_room_name"`
	ExpRoomInfo    string `json:"exp_room_info"`
	DebugUsers     string `json:"debug_users,omitempty"`
	DebugCrowdId   int64  `json:"debug_crowd_id,omitempty"`
	BucketCount    int32  `json:"bucket_count,omitempty"`
	ExpRoomBuckets string `json:"exp_room_buckets,omitempty"`
	BucketType     uint32 `json:"bucket_type"`
	Filter         string `json:"filter"`
	ExpRoomConfig  string `json:"exp_room_config,omitempty"`
	Environment    int32  `json:"environment"`
	//EnvironmentStr string `json:"-"`
	Type   uint32 `json:"type"`
	Status int32  `json:"status,omitempty"`

	DebugCrowdIdUsers []string        `json:"debug_crowd_id_users"`
	debugUserMap      map[string]bool `json:"-"`
	diversionBucket   DiversionBucket
	Layers            []*Layer `json:"layers"`
}

func (e *ExperimentRoom) Init() error {
	//  deal ExpRoomBuckets
	//e.diversionBucket = NewDiversionBucket(e.Type)
	if e.diversionBucket == nil {
		if e.BucketType == common.Bucket_Type_UID {
			e.diversionBucket = NewUidDiversionBucket(int(e.BucketCount), e.ExpRoomBuckets)
		} else if e.BucketType == common.Bucket_Type_UID_HASH {
			e.diversionBucket = NewUidHashDiversionBucket(int(e.BucketCount), e.ExpRoomBuckets)
		} else if e.BucketType == common.Bucket_Type_Filter {
			diversionBucket, err := NewFilterDiversionBucket(e.Filter)
			if err != nil {
				return err
			}
			e.diversionBucket = diversionBucket
		} else if e.BucketType == common.Bucket_Type_Custom {
			e.diversionBucket = NewCustomDiversionBucket()
		}

	}
	// deal DebugUsers
	e.debugUserMap = make(map[string]bool, 0)
	if e.DebugUsers != "" {
		uids := strings.Split(e.DebugUsers, ",")
		for _, uid := range uids {
			e.debugUserMap[uid] = true
		}
	}
	if len(e.DebugCrowdIdUsers) != 0 {
		for _, user := range e.DebugCrowdIdUsers {
			e.debugUserMap[user] = true
		}
	}

	return nil
}

func (e *ExperimentRoom) AddLayer(l *Layer) {
	e.Layers = append(e.Layers, l)
}

// MatchDebugUsers return true if debug_users is set and debug_users contain of uid
func (e *ExperimentRoom) MatchDebugUsers(experimentContext *ExperimentContext) bool {
	if _, found := e.debugUserMap[experimentContext.Uid]; found {
		return true
	}

	return false
}

func (e *ExperimentRoom) Match(experimentContext *ExperimentContext) bool {

	if _, found := e.debugUserMap[experimentContext.Uid]; found {
		return true
	}
	if e.diversionBucket != nil {
		return e.diversionBucket.Match(experimentContext)
	}

	return false
}
