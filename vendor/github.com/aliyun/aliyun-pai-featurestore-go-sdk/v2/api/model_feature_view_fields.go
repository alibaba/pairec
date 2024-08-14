package api

type FeatureViewFields struct {
	Name         string `json:"name,omitempty"`
	Type         int32  `json:"type,omitempty"`
	IsPartition  bool   `json:"is_partition,omitempty"`
	IsPrimaryKey bool   `json:"is_primary_key,omitempty"`
	IsEventTime  bool   `json:"is_event_time,omitempty"`
	Position     int    `json:"position,omitempty"`
}
