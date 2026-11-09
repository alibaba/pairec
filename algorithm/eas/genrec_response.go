package eas

// GenRecResponse is the JSON response of the EAS GenRec /generate API. Only
// the fields consumed by the recall are declared; the rest of the payload is
// ignored during unmarshalling.
type GenRecResponse struct {
	OutputIds []int64       `json:"output_ids"`
	MetaInfo  GenRecMetaInfo `json:"meta_info"`
}

type GenRecMetaInfo struct {
	BeamResults []GenRecBeamResult `json:"beam_results"`
}

// GenRecBeamResult is one SID candidate. ItemIds holds the items mapped to the
// SID sequence; one SID may map to multiple items sharing the same score.
type GenRecBeamResult struct {
	OutputIds []int64  `json:"output_ids"`
	ItemIds   []string `json:"item_ids"`
	MetaInfo  struct {
		// SequenceScore is the length-normalized log probability of the SID
		// sequence in (-inf, 0]. Larger is better; it is not a CTR.
		SequenceScore float64 `json:"sequence_score"`
	} `json:"meta_info"`
}
