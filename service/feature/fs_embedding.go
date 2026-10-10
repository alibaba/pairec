package feature

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
	"github.com/expr-lang/expr/ast"
)

var errSkipFSEmbedding = errors.New("empty embedding input")

// Models are prepared before publishing the feature configuration, then read only.
type fsEmbedding struct {
	models map[[2]string]*domain.LLMConfig
	err    error
}

func (e *fsEmbedding) Visit(node *ast.Node) {
	call, ok := (*node).(*ast.CallNode)
	if !ok {
		return
	}
	callee, ok := call.Callee.(*ast.IdentifierNode)
	if !ok || callee.Value != "fsEmbedding" {
		return
	}
	if e.models == nil {
		e.models = make(map[[2]string]*domain.LLMConfig)
	}
	if e.err != nil {
		return
	}
	if len(call.Arguments) != 4 {
		e.err = errors.New("fsEmbedding requires four arguments")
		return
	}
	store, storeOK := call.Arguments[0].(*ast.StringNode)
	model, modelOK := call.Arguments[1].(*ast.StringNode)
	timeout, timeoutOK := call.Arguments[3].(*ast.IntegerNode)
	if !storeOK || !modelOK || strings.TrimSpace(store.Value) == "" || strings.TrimSpace(model.Value) == "" {
		e.err = errors.New("fsEmbedding requires non-empty string literals for the feature store and model names")
		return
	}
	if !timeoutOK || timeout.Value <= 0 || int64(timeout.Value) > math.MaxInt64/int64(time.Millisecond) {
		e.err = errors.New("fsEmbedding requires a positive millisecond timeout literal within time.Duration range")
		return
	}
	e.models[[2]string{store.Value, model.Value}] = nil
}

// Only this configuration-loading path may perform model metadata requests.
func prepareFSEmbeddings(f *Feature, conf recconf.FeatureLoadConfig) {
	for _, trans := range f.featureTrans {
		ft, ok := trans.(*featureTrans)
		if !ok {
			continue
		}
		n, ok := ft.normalizer.(*ExprNormalizer)
		if !ok || n.embedding == nil || n.prog == nil {
			continue
		}
		e := n.embedding
		if _, ok := ft.featureOp.(CreateNewFeatureOp); !ok || ft.featureStore != SOURCE_USER || conf.FeatureDaoConf.FeatureAsyncLoad {
			e.err = errors.New("fsEmbedding requires a synchronous user new_feature")
			log.Error(fmt.Sprintf("event=FSEmbedding\tfeature=%s\terror=%v", ft.featureName, e.err))
			continue
		}
		for key := range e.models {
			client, err := fs.GetFeatureStoreClient(key[0])
			if err != nil {
				log.Error(fmt.Sprintf("event=FSEmbedding\tfeature=%s\terror=%v", ft.featureName, err))
				continue
			}
			model, err := client.GetLLMConfig(key[1])
			if err != nil || model == nil || model.LLMConfig == nil {
				log.Error(fmt.Sprintf("event=FSEmbedding\tfeature=%s\tstore=%s\tmodel=%s\terror=load model failed", ft.featureName, key[0], key[1]))
				continue
			}
			if model.ModelType != domain.LLMModelTypeTextEmbedding && model.ModelType != domain.LLMModelTypeMultiModalEmbedding {
				log.Error(fmt.Sprintf("event=FSEmbedding\tfeature=%s\tstore=%s\tmodel=%s\terror=unsupported model type", ft.featureName, key[0], key[1]))
				continue
			}
			e.models[key] = model
		}
	}
}

func (e *fsEmbedding) call(args ...any) (any, error) {
	if args[2] == nil {
		return nil, errSkipFSEmbedding
	}
	text, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("fsEmbedding input must be a string, got %T", args[2])
	}
	if strings.TrimSpace(text) == "" {
		return nil, errSkipFSEmbedding
	}
	if e.err != nil {
		return nil, e.err
	}
	key := [2]string{args[0].(string), args[1].(string)}
	model := e.models[key]
	if model == nil {
		return nil, fmt.Errorf("fsEmbedding model %q/%q is not prepared", key[0], key[1])
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(args[3].(int))*time.Millisecond)
	defer cancel()
	var vectors [][]float32
	var err error
	if model.ModelType == domain.LLMModelTypeTextEmbedding {
		vectors, err = model.CreateTextEmbeddings(ctx, []string{text})
	} else {
		vectors, err = model.CreateMultiModalEmbeddings(ctx, []domain.MultiModalContent{{Text: text}})
	}
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("fsEmbedding model %q/%q: %w", key[0], key[1], ctx.Err())
		}
		// SDK errors can contain the response body; do not log user text or credentials.
		return nil, fmt.Errorf("fsEmbedding model %q/%q: SDK request failed", key[0], key[1])
	}
	if len(vectors) != 1 || len(vectors[0]) == 0 {
		return nil, fmt.Errorf("fsEmbedding model %q/%q returned no single non-empty vector", key[0], key[1])
	}
	for _, v := range vectors[0] {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return nil, fmt.Errorf("fsEmbedding model %q/%q returned a non-finite value", key[0], key[1])
		}
	}
	return vectors[0], nil
}
