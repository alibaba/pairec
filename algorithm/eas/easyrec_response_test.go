package eas

import (
	"math/rand"
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
)

func TestEasyrecMutClassificationResponseFunc(t *testing.T) {

	t.Run("test signle item", func(t *testing.T) {
		pbResponse := &easyrec.PBResponse{}
		pbResponse.ItemIds = []string{"item_1"}
		pbResponse.TfOutputs = make(map[string]*easyrec.ArrayProto)
		pbResponse.TfOutputs["probs_is_complete_play"] = &easyrec.ArrayProto{
			FloatVal:   []float32{rand.Float32()},
			Dtype:      easyrec.ArrayDataType_DT_FLOAT,
			ArrayShape: &easyrec.ArrayShape{Dim: []int64{1}},
		}
		pbResponse.TfOutputs["probs_is_like_or_cmt"] = &easyrec.ArrayProto{
			FloatVal:   []float32{rand.Float32()},
			Dtype:      easyrec.ArrayDataType_DT_FLOAT,
			ArrayShape: &easyrec.ArrayShape{Dim: []int64{1}},
		}
		pbResponse.TfOutputs["probs_is_play_label"] = &easyrec.ArrayProto{
			FloatVal:   []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
			Dtype:      easyrec.ArrayDataType_DT_FLOAT,
			ArrayShape: &easyrec.ArrayShape{Dim: []int64{1, 6}},
		}
		ret, err := easyrecMutClassificationResponseFunc(pbResponse)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(ret), 1)
		assert.Equal(t, float32(ret[0].(*EasyrecClassificationResponse).mulClassifyArr["probs_is_complete_play"][0]),
			pbResponse.TfOutputs["probs_is_complete_play"].FloatVal[0])
		assert.Equal(t, len(ret[0].(*EasyrecClassificationResponse).mulClassifyArr["probs_is_play_label"]),
			int(pbResponse.TfOutputs["probs_is_play_label"].ArrayShape.Dim[1]))
	})
	t.Run("test multi item", func(t *testing.T) {
		pbResponse := &easyrec.PBResponse{}
		pbResponse.ItemIds = []string{"item_1", "item_2"}
		pbResponse.TfOutputs = make(map[string]*easyrec.ArrayProto)
		pbResponse.TfOutputs["probs_is_complete_play"] = &easyrec.ArrayProto{
			FloatVal:   []float32{rand.Float32(), 0.3},
			Dtype:      easyrec.ArrayDataType_DT_FLOAT,
			ArrayShape: &easyrec.ArrayShape{Dim: []int64{1}},
		}
		pbResponse.TfOutputs["probs_is_like_or_cmt"] = &easyrec.ArrayProto{
			FloatVal:   []float32{rand.Float32(), 0.4},
			Dtype:      easyrec.ArrayDataType_DT_FLOAT,
			ArrayShape: &easyrec.ArrayShape{Dim: []int64{1}},
		}
		pbResponse.TfOutputs["probs_is_play_label"] = &easyrec.ArrayProto{
			FloatVal:   []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.11, 0.22, 0.33, 0.44, 0.55, 0.66},
			Dtype:      easyrec.ArrayDataType_DT_FLOAT,
			ArrayShape: &easyrec.ArrayShape{Dim: []int64{1, 6}},
		}
		ret, err := easyrecMutClassificationResponseFunc(pbResponse)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(ret), 2)
		assert.Equal(t, float32(ret[0].(*EasyrecClassificationResponse).mulClassifyArr["probs_is_complete_play"][0]),
			pbResponse.TfOutputs["probs_is_complete_play"].FloatVal[0])
		assert.Equal(t, float32(ret[1].(*EasyrecClassificationResponse).mulClassifyArr["probs_is_like_or_cmt"][0]),
			pbResponse.TfOutputs["probs_is_like_or_cmt"].FloatVal[1])
		assert.Equal(t, len(ret[0].(*EasyrecClassificationResponse).mulClassifyArr["probs_is_play_label"]),
			int(pbResponse.TfOutputs["probs_is_play_label"].ArrayShape.Dim[1]))
		assert.Equal(t, len(ret[1].(*EasyrecClassificationResponse).mulClassifyArr["probs_is_play_label"]),
			int(pbResponse.TfOutputs["probs_is_play_label"].ArrayShape.Dim[1]))
		assert.Equal(t, float32(ret[1].(*EasyrecClassificationResponse).mulClassifyArr["probs_is_play_label"][5]),
			float32(0.66))
	})
}
