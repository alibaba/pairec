package feature

import (
	gocontext "context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/datasource/featuredb"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

const testEmbeddingExpr = `fsEmbedding("fs", "model", user.title, 100)`

var (
	featureDBServerOnce sync.Once
	featureDBHandler    atomic.Value
)

// useFeatureDBServer points the process-wide FeatureDB client at a local server.
// The client is a singleton, so the server is shared and only its handler is replaced.
func useFeatureDBServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	featureDBServerOnce.Do(func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureDBHandler.Load().(http.HandlerFunc)(w, r)
		}))
		featuredb.InitFeatureDBClient(server.URL, "token", "", true)
	})
	featureDBHandler.Store(handler)
}

func useFakeEmbeddings(t *testing.T, fn func(ctx gocontext.Context, model *domain.LLMConfig, text string) ([][]float32, error)) {
	t.Helper()
	origin := createFSEmbeddings
	createFSEmbeddings = fn
	t.Cleanup(func() { createFSEmbeddings = origin })
}

func newTestLLMConfig(modelType string) *domain.LLMConfig {
	return domain.NewLLMConfig(&api.LLMConfig{LLMConfigId: 7, Name: "model", ModelType: modelType}, "inst", "sig")
}

func newEmbeddingNormalizer(t *testing.T, expression string) *ExprNormalizer {
	t.Helper()
	n := NewExprNormalizer(expression)
	if n.prog == nil || n.embedding == nil {
		t.Fatalf("expression %q should compile as an embedding expression, err:%v", expression, n.embedding)
	}
	return n
}

func prepareTestModel(n *ExprNormalizer, model *domain.LLMConfig) {
	n.embedding.models[[2]string{"fs", "model"}] = model
}

func exprNormalizerOf(t *testing.T, f *Feature) *ExprNormalizer {
	t.Helper()
	n, ok := f.featureTrans[0].(*featureTrans).normalizer.(*ExprNormalizer)
	if !ok {
		t.Fatalf("normalizer type error:%T", f.featureTrans[0].(*featureTrans).normalizer)
	}
	return n
}

func TestFSEmbeddingArgumentValidation(t *testing.T) {
	testcases := []struct {
		name       string
		expression string
		wantErr    string
	}{
		{"too few arguments", `fsEmbedding("fs", "model", user.title)`, "four arguments"},
		{"too many arguments", `fsEmbedding("fs", "model", user.title, 100, 1)`, "four arguments"},
		{"store is not a literal", `fsEmbedding(user.store, "model", user.title, 100)`, "non-empty string literals"},
		{"model is not a literal", `fsEmbedding("fs", user.model, user.title, 100)`, "non-empty string literals"},
		{"blank store", `fsEmbedding("  ", "model", user.title, 100)`, "non-empty string literals"},
		{"blank model", `fsEmbedding("fs", "", user.title, 100)`, "non-empty string literals"},
		{"timeout is not a literal", `fsEmbedding("fs", "model", user.title, user.timeout)`, "positive millisecond timeout"},
		{"zero timeout", `fsEmbedding("fs", "model", user.title, 0)`, "positive millisecond timeout"},
		{"negative timeout", `fsEmbedding("fs", "model", user.title, -1)`, "positive millisecond timeout"},
		{"timeout overflows time.Duration", `fsEmbedding("fs", "model", user.title, 9223372036855)`, "positive millisecond timeout"},
		{"first invalid call wins", `[fsEmbedding("fs", "model", user.title, 0), fsEmbedding("", "model", user.title, 100)]`, "positive millisecond timeout"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			n := NewExprNormalizer(tc.expression)
			if n.prog != nil {
				t.Fatal("invalid embedding expression should not compile")
			}
			if n.embedding == nil || n.embedding.err == nil || !strings.Contains(n.embedding.err.Error(), tc.wantErr) {
				t.Fatalf("embedding error mismatch, want:%q, got:%+v", tc.wantErr, n.embedding)
			}
			// a broken embedding expression yields nil, so the target feature is kept
			if result := n.Apply(map[string]any{"user": map[string]any{"title": "hello"}}); result != nil {
				t.Fatalf("result should be nil, got:%v", result)
			}
		})
	}

	t.Run("valid arguments", func(t *testing.T) {
		n := newEmbeddingNormalizer(t, `fsEmbedding("fs", "model", user.title, 100) ?? fsEmbedding("fs2", "model2", user.title, 9223372036854)`)
		if n.embedding.err != nil {
			t.Fatalf("unexpected error:%v", n.embedding.err)
		}
		want := map[[2]string]*domain.LLMConfig{{"fs", "model"}: nil, {"fs2", "model2"}: nil}
		if !reflect.DeepEqual(n.embedding.models, want) {
			t.Fatalf("models mismatch, got:%v", n.embedding.models)
		}
	})

	t.Run("plain expression is not affected", func(t *testing.T) {
		n := NewExprNormalizer(`user.age + 1`)
		if n.prog == nil || n.embedding != nil {
			t.Fatalf("plain expression should compile without embedding, embedding:%+v", n.embedding)
		}
		if result := n.Apply(map[string]any{"user": map[string]any{"age": 1}}); result != 2 {
			t.Fatalf("result error, got:%v", result)
		}
		// a failed plain expression still falls back to ""
		if result := NewExprNormalizer(`user.age +`).Apply(map[string]any{}); result != "" {
			t.Fatalf("result should be empty string, got:%v", result)
		}
	})
}

func TestFSEmbeddingCallInput(t *testing.T) {
	useFakeEmbeddings(t, func(ctx gocontext.Context, model *domain.LLMConfig, text string) ([][]float32, error) {
		t.Fatal("remote request should not be sent")
		return nil, nil
	})
	e := &fsEmbedding{models: map[[2]string]*domain.LLMConfig{{"fs", "model"}: newTestLLMConfig(domain.LLMModelTypeTextEmbedding)}}

	for _, input := range []any{nil, "", "  \t\n"} {
		if _, err := e.call("fs", "model", input, 100); !errors.Is(err, errSkipFSEmbedding) {
			t.Fatalf("input %q should be skipped, got:%v", input, err)
		}
	}
	if _, err := e.call("fs", "model", 123, 100); err == nil || !strings.Contains(err.Error(), "must be a string, got int") {
		t.Fatalf("non string input error mismatch, got:%v", err)
	}

	// empty input is skipped before the configuration error is reported
	e.err = errors.New("config error")
	if _, err := e.call("fs", "model", "", 100); !errors.Is(err, errSkipFSEmbedding) {
		t.Fatalf("empty input should be skipped, got:%v", err)
	}
	if _, err := e.call("fs", "model", "hello", 100); err != e.err {
		t.Fatalf("configuration error should be returned, got:%v", err)
	}
}

func TestFSEmbeddingCallModelMissing(t *testing.T) {
	e := &fsEmbedding{models: map[[2]string]*domain.LLMConfig{{"fs", "model"}: nil}}
	for _, key := range [][2]string{{"fs", "model"}, {"fs", "other"}} {
		if _, err := e.call(key[0], key[1], "hello", 100); err == nil || !strings.Contains(err.Error(), "is not prepared") {
			t.Fatalf("model %v should not be prepared, got:%v", key, err)
		}
	}
}

func TestFSEmbeddingCallResult(t *testing.T) {
	testcases := []struct {
		name    string
		vectors [][]float32
		wantErr string
	}{
		{"single vector", [][]float32{{0.1, -0.2}}, ""},
		{"no vector", nil, "no single non-empty vector"},
		{"empty vector", [][]float32{{}}, "no single non-empty vector"},
		{"more than one vector", [][]float32{{0.1}, {0.2}}, "no single non-empty vector"},
		{"NaN", [][]float32{{0.1, float32(math.NaN())}}, "non-finite value"},
		{"positive infinity", [][]float32{{float32(math.Inf(1))}}, "non-finite value"},
		{"negative infinity", [][]float32{{0.1, float32(math.Inf(-1))}}, "non-finite value"},
	}
	e := &fsEmbedding{models: map[[2]string]*domain.LLMConfig{{"fs", "model"}: newTestLLMConfig(domain.LLMModelTypeTextEmbedding)}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			useFakeEmbeddings(t, func(ctx gocontext.Context, model *domain.LLMConfig, text string) ([][]float32, error) {
				return tc.vectors, nil
			})
			result, err := e.call("fs", "model", "hello", 100)
			if tc.wantErr == "" {
				if err != nil || !reflect.DeepEqual(result, tc.vectors[0]) {
					t.Fatalf("result mismatch, result:%v, err:%v", result, err)
				}
				return
			}
			if result != nil || err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error mismatch, want:%q, result:%v, err:%v", tc.wantErr, result, err)
			}
		})
	}
}

func TestFSEmbeddingCallRemoteError(t *testing.T) {
	e := &fsEmbedding{models: map[[2]string]*domain.LLMConfig{{"fs", "model"}: newTestLLMConfig(domain.LLMModelTypeTextEmbedding)}}

	t.Run("timeout from the timeout argument", func(t *testing.T) {
		useFakeEmbeddings(t, func(ctx gocontext.Context, model *domain.LLMConfig, text string) ([][]float32, error) {
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) > 20*time.Millisecond {
				t.Errorf("deadline should follow the timeout argument, deadline:%v", deadline)
			}
			<-ctx.Done()
			return nil, ctx.Err()
		})
		if _, err := e.call("fs", "model", "hello", 20); !errors.Is(err, gocontext.DeadlineExceeded) {
			t.Fatalf("timeout error mismatch, got:%v", err)
		}
	})

	t.Run("SDK error details are hidden", func(t *testing.T) {
		useFakeEmbeddings(t, func(ctx gocontext.Context, model *domain.LLMConfig, text string) ([][]float32, error) {
			return nil, errors.New("body: secret " + text)
		})
		_, err := e.call("fs", "model", "user-text", 100)
		if err == nil || !strings.Contains(err.Error(), "SDK request failed") || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "user-text") {
			t.Fatalf("error should be sanitized, got:%v", err)
		}
	})
}

func TestFSEmbeddingSDKRequest(t *testing.T) {
	testcases := []struct {
		modelType string
		path      string
		body      string
	}{
		{domain.LLMModelTypeTextEmbedding, "/api/v1/llm_configs/inst/7/text_embeddings", `{"input":["hello"]}`},
		{domain.LLMModelTypeMultiModalEmbedding, "/api/v1/llm_configs/inst/7/multi_modal_embeddings", `{"multi_modal_input":[{"text":"hello"}]}`},
	}
	for _, tc := range testcases {
		t.Run(tc.modelType, func(t *testing.T) {
			useFeatureDBServer(t, func(w http.ResponseWriter, r *http.Request) {
				var body any
				_ = json.NewDecoder(r.Body).Decode(&body)
				var wantBody any
				_ = json.Unmarshal([]byte(tc.body), &wantBody)
				if r.URL.Path != tc.path || !reflect.DeepEqual(body, wantBody) || r.Header.Get("Auth") != "sig" {
					t.Errorf("request mismatch, path:%s, body:%v, auth:%s", r.URL.Path, body, r.Header.Get("Auth"))
				}
				_, _ = w.Write([]byte(`{"code":"OK","data":{"embeddings":[[0.5,-1.5]]}}`))
			})
			n := newEmbeddingNormalizer(t, testEmbeddingExpr)
			prepareTestModel(n, newTestLLMConfig(tc.modelType))

			result := n.Apply(map[string]any{"user": map[string]any{"title": "hello"}})
			if !reflect.DeepEqual(result, []float32{0.5, -1.5}) {
				t.Fatalf("result mismatch, got:%T %v", result, result)
			}
		})
	}

	t.Run("http error", func(t *testing.T) {
		useFeatureDBServer(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "secret response body", http.StatusInternalServerError)
		})
		e := &fsEmbedding{models: map[[2]string]*domain.LLMConfig{{"fs", "model"}: newTestLLMConfig(domain.LLMModelTypeTextEmbedding)}}
		_, err := e.call("fs", "model", "hello", 1000)
		if err == nil || !strings.Contains(err.Error(), "SDK request failed") || strings.Contains(err.Error(), "secret") {
			t.Fatalf("error should be sanitized, got:%v", err)
		}
	})

	t.Run("server slower than the timeout", func(t *testing.T) {
		useFeatureDBServer(t, func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(time.Second):
			}
		})
		e := &fsEmbedding{models: map[[2]string]*domain.LLMConfig{{"fs", "model"}: newTestLLMConfig(domain.LLMModelTypeTextEmbedding)}}
		if _, err := e.call("fs", "model", "hello", 20); !errors.Is(err, gocontext.DeadlineExceeded) {
			t.Fatalf("timeout error mismatch, got:%v", err)
		}
	})
}

func TestPrepareFSEmbeddings(t *testing.T) {
	t.Run("embedding requires new_feature", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:  "raw_feature",
			FeatureStore: "user",
			FeatureName:  "title_emb",
			Normalizer:   "expr",
			Expression:   testEmbeddingExpr,
		})
		n := exprNormalizerOf(t, LoadWithConfig(conf))
		if n.embedding.err == nil || !strings.Contains(n.embedding.err.Error(), "requires new_feature") {
			t.Fatalf("error mismatch, got:%v", n.embedding.err)
		}
		if result := n.Apply(map[string]any{"user": map[string]any{"title": "hello"}}); result != nil {
			t.Fatalf("result should be nil, got:%v", result)
		}
	})

	t.Run("missing feature store client keeps the model unprepared", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "user",
			FeatureName:  "title_emb",
			Normalizer:   "expr",
			Expression:   `fsEmbedding("fs_embedding_test_missing", "model", user.title, 100)`,
		})
		n := exprNormalizerOf(t, LoadWithConfig(conf))
		key := [2]string{"fs_embedding_test_missing", "model"}
		if model, ok := n.embedding.models[key]; !ok || model != nil || n.embedding.err != nil {
			t.Fatalf("model should stay unprepared, models:%v, err:%v", n.embedding.models, n.embedding.err)
		}
	})

	t.Run("broken and plain expressions are skipped", func(t *testing.T) {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:  "raw_feature",
			FeatureStore: "user",
			FeatureName:  "broken",
			Normalizer:   "expr",
			Expression:   `fsEmbedding("fs", "model", user.title, 0)`,
		}, recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "user",
			FeatureName:  "plain",
			Normalizer:   "expr",
			Expression:   `user.age + 1`,
		})
		f := LoadWithConfig(conf)
		n := exprNormalizerOf(t, f)
		// compile errors are kept instead of being replaced by the new_feature check
		if n.embedding.err == nil || !strings.Contains(n.embedding.err.Error(), "positive millisecond timeout") {
			t.Fatalf("error mismatch, got:%v", n.embedding.err)
		}
		if plain := f.featureTrans[1].(*featureTrans).normalizer.(*ExprNormalizer); plain.embedding != nil {
			t.Fatalf("plain expression should not use embedding, got:%+v", plain.embedding)
		}
	})
}

func TestCreateNewFeatureOpKeepsValueWithoutEmbedding(t *testing.T) {
	userConf := func(expression string) recconf.FeatureLoadConfig {
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "user",
			FeatureName:  "title_emb",
			Normalizer:   "expr",
			Expression:   expression,
		})
		return conf
	}
	vector := []float32{0.1, 0.2}

	t.Run("user scope", func(t *testing.T) {
		testcases := []struct {
			name     string
			title    any
			prepared bool
			vectors  [][]float32
			err      error
			want     any
		}{
			{"embedding succeeds", "hello", true, [][]float32{vector}, nil, vector},
			{"missing input", nil, true, [][]float32{vector}, nil, "old"},
			{"blank input", "  ", true, [][]float32{vector}, nil, "old"},
			{"model is not prepared", "hello", false, [][]float32{vector}, nil, "old"},
			{"empty vector", "hello", true, [][]float32{{}}, nil, "old"},
			{"non-finite value", "hello", true, [][]float32{{float32(math.NaN())}}, nil, "old"},
			{"remote error", "hello", true, nil, errors.New("remote error"), "old"},
		}
		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				useFakeEmbeddings(t, func(ctx gocontext.Context, model *domain.LLMConfig, text string) ([][]float32, error) {
					return tc.vectors, tc.err
				})
				f := LoadWithConfig(userConf(testEmbeddingExpr))
				if tc.prepared {
					prepareTestModel(exprNormalizerOf(t, f), newTestLLMConfig(domain.LLMModelTypeTextEmbedding))
				}
				user := module.NewUser("u1")
				user.AddProperty("title_emb", "old")
				if tc.title != nil {
					user.AddProperty("title", tc.title)
				}
				f.LoadFeatures(user, nil, context.NewRecommendContext())
				if got := user.GetProperty("title_emb"); !reflect.DeepEqual(got, tc.want) {
					t.Fatalf("title_emb mismatch, want:%v, got:%v", tc.want, got)
				}
			})
		}
	})

	t.Run("user scope does not create the feature on skip", func(t *testing.T) {
		f := LoadWithConfig(userConf(testEmbeddingExpr))
		user := module.NewUser("u1")
		f.LoadFeatures(user, nil, context.NewRecommendContext())
		if _, ok := user.Properties["title_emb"]; ok {
			t.Fatalf("title_emb should not be created, got:%v", user.Properties["title_emb"])
		}
	})

	t.Run("broken embedding expression keeps the user value", func(t *testing.T) {
		f := LoadWithConfig(userConf(`fsEmbedding("fs", "model", user.title, 0)`))
		user := module.NewUser("u1")
		user.AddProperty("title", "hello")
		user.AddProperty("title_emb", "old")
		f.LoadFeatures(user, nil, context.NewRecommendContext())
		if got := user.GetProperty("title_emb"); got != "old" {
			t.Fatalf("title_emb should be kept, got:%v", got)
		}
	})

	t.Run("item scope", func(t *testing.T) {
		useFakeEmbeddings(t, func(ctx gocontext.Context, model *domain.LLMConfig, text string) ([][]float32, error) {
			if text == "fail" {
				return nil, errors.New("remote error")
			}
			return [][]float32{vector}, nil
		})
		conf := recconf.FeatureLoadConfig{}
		conf.Features = append(conf.Features, recconf.FeatureConfig{
			FeatureType:  "new_feature",
			FeatureStore: "item",
			FeatureName:  "title_emb",
			Normalizer:   "expr",
			Expression:   `fsEmbedding("fs", "model", item.title, 100)`,
		})
		f := LoadWithConfig(conf)
		prepareTestModel(exprNormalizerOf(t, f), newTestLLMConfig(domain.LLMModelTypeTextEmbedding))

		items := make([]*module.Item, 0, 3)
		for i, title := range []any{"hello", nil, "fail"} {
			item := module.NewItem(string(rune('a' + i)))
			item.AddProperty("title_emb", "old")
			if title != nil {
				item.AddProperty("title", title)
			}
			items = append(items, item)
		}
		f.LoadFeatures(module.NewUser("u1"), items, context.NewRecommendContext())

		for i, want := range []any{vector, "old", "old"} {
			if got := items[i].GetProperty("title_emb"); !reflect.DeepEqual(got, want) {
				t.Fatalf("item %d title_emb mismatch, want:%v, got:%v", i, want, got)
			}
		}
	})
}
