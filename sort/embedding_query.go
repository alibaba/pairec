package sort

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/huandu/go-sqlbuilder"
)

// embeddingDatasource encapsulates the shared logic for loading item embedding
// vectors from either a hologres table (SQL) or a featurestore feature view.
// DPPSort and SSDSort embed it so both reuse exactly the same query behavior;
// logModule is the only per-sorter difference and keeps each sorter's log prefix.
type embeddingDatasource struct {
	logModule      string
	fsClient       *fs.FSClient
	fsViewName     string
	db             *sql.DB
	tableName      string
	keyField       string
	embeddingField string
	embSeparator   string
}

// hasEmbeddingTable reports whether an embedding datasource (hologres table or
// featurestore feature view) is configured.
func (d embeddingDatasource) hasEmbeddingTable() bool {
	if d.fsClient != nil {
		return len(d.fsViewName) > 0
	}
	return len(d.tableName) > 0
}

// embeddingSource returns the configured embedding datasource name for logging.
func (d embeddingDatasource) embeddingSource() string {
	if d.fsClient != nil {
		return d.fsViewName
	}
	return d.tableName
}

// queryEmbeddings fetches the raw (un-normalized) embedding vectors of the absent
// items from the configured datasource, keyed by item id. sourceSuffix is the
// runtime suffix resolved from the AB param named by TableSuffixParam: it is
// appended to the hologres table name or, for featurestore, to the feature view
// name, enabling per-experiment/version source switching. An empty suffix keeps
// the configured base name.
func (d embeddingDatasource) queryEmbeddings(ctx *context.RecommendContext, absentItemIds []interface{}, sourceSuffix string) (map[string][]float64, error) {
	if d.fsClient != nil {
		return d.queryEmbeddingsFromFeatureStore(ctx, absentItemIds, sourceSuffix)
	}
	return d.queryEmbeddingsFromHologres(ctx, absentItemIds, sourceSuffix)
}

func (d embeddingDatasource) queryEmbeddingsFromHologres(ctx *context.RecommendContext, absentItemIds []interface{}, tableSuffix string) (map[string][]float64, error) {
	table := d.tableName + tableSuffix
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(d.keyField, d.embeddingField)
	builder.From(table)
	builder.Where(builder.In(d.keyField, absentItemIds...))

	sqlQuery, args := builder.Build()
	ctx.LogDebug("module=" + d.logModule + "\tsqlquery=" + sqlQuery)
	rows, err := d.db.Query(sqlQuery, args...)
	if err != nil {
		ctx.LogError(fmt.Sprintf("module=%s\terror=%v", d.logModule, err))
		return nil, err
	}
	defer rows.Close()
	result := make(map[string][]float64, len(absentItemIds))
	itemID := &sql.NullString{}
	itemEmb := &sql.NullString{}
	for rows.Next() {
		if err := rows.Scan(itemID, itemEmb); err != nil {
			ctx.LogError(fmt.Sprintf("module=Scan %s\terror=%v\tProductID=%s",
				d.logModule, err, itemID.String))
			continue
		}
		result[itemID.String] = d.parseEmbedding(ctx, itemEmb.String)
	}
	// rows.Next() returns false both at the normal end of rows and on a
	// mid-iteration failure (network reset, statement timeout, driver error,
	// context cancellation); the latter is only surfaced via rows.Err(). Without
	// this check a partial result would be silently treated as complete, so the
	// un-fetched items would be counted as missing and random-filled downstream,
	// corrupting the ranking. Return the error so the caller falls back to the
	// original order instead.
	if err := rows.Err(); err != nil {
		ctx.LogError(fmt.Sprintf("module=%s\trows error=%v", d.logModule, err))
		return nil, err
	}
	return result, nil
}

func (d embeddingDatasource) queryEmbeddingsFromFeatureStore(ctx *context.RecommendContext, absentItemIds []interface{}, viewSuffix string) (map[string][]float64, error) {
	// The feature view can be switched at runtime by appending the AB-driven
	// suffix (same mechanism as the hologres table suffix); an empty suffix keeps
	// the configured base view. GetFeatureView only resolves views already loaded
	// in the project, so an unknown suffix surfaces as "featureView not found".
	viewName := d.fsViewName + viewSuffix
	featureView := d.fsClient.GetProject().GetFeatureView(viewName)
	if featureView == nil {
		err := fmt.Errorf("featureView not found, view:%s", viewName)
		ctx.LogError(fmt.Sprintf("module=%s\terror=%v", d.logModule, err))
		return nil, err
	}
	features, err := featureView.GetOnlineFeatures(absentItemIds, []string{d.embeddingField}, nil)
	if err != nil {
		ctx.LogError(fmt.Sprintf("module=%s\terror=featurestore error(%v)", d.logModule, err))
		return nil, err
	}
	featureEntity := d.fsClient.GetProject().GetFeatureEntity(featureView.GetFeatureEntityName())
	if featureEntity == nil {
		err := fmt.Errorf("featureEntity not found, name:%s", featureView.GetFeatureEntityName())
		ctx.LogError(fmt.Sprintf("module=%s\terror=%v", d.logModule, err))
		return nil, err
	}
	result := make(map[string][]float64, len(features))
	for _, featureMap := range features {
		itemId := utils.ToString(featureMap[featureEntity.FeatureEntityJoinid], "")
		if itemId == "" {
			continue
		}
		raw, ok := featureMap[d.embeddingField]
		if !ok || raw == nil {
			continue
		}
		result[itemId] = d.parseEmbedding(ctx, raw)
	}
	return result, nil
}

// parseEmbedding converts a raw embedding value into a float64 vector. It is
// compatible with both the string format ("1.0,2.0", "{1.0,2.0}", "[1.0,2.0]")
// used by sql datasources and the native array formats returned by featurestore
// ([]float32, []float64). Any other type is treated as unsupported and returns
// nil so the caller routes the item to the missing-embedding fallback.
// The returned slice reserves one extra capacity slot so that the later
// append(emb, 1) while building the kernel matrix does not trigger reallocation.
// Parse failures are logged through the request-scoped ctx logger, prefixed with
// the sorter's logModule, so DPP and SSD failures can be told apart and
// correlated to a specific request.
func (d embeddingDatasource) parseEmbedding(ctx *context.RecommendContext, raw interface{}) []float64 {
	if raw == nil {
		return nil
	}
	separator := d.embSeparator
	if separator == "" {
		separator = ","
	}
	switch v := raw.(type) {
	case []float64:
		vec := make([]float64, len(v), len(v)+1)
		copy(vec, v)
		return vec
	case []float32:
		vec := make([]float64, len(v), len(v)+1)
		for i, f := range v {
			vec[i] = float64(f)
		}
		return vec
	case string:
		// FeatureStore surfaces a NULL feature value as the literal string "null"
		// (case-insensitive); module/item2x_featurestore_dao.go and
		// module/trigger_info.go skip it the same way. Treat it as a missing
		// embedding (nil) instead of parsing it into a fabricated zero vector that
		// would normalize to NaN. The hologres path already maps a NULL column to
		// "" which the empty-content check below handles.
		if strings.EqualFold(strings.TrimSpace(v), "null") {
			return nil
		}
		content := strings.Trim(v, "{}[]")
		if strings.TrimSpace(content) == "" {
			return nil
		}
		elements := strings.Split(content, separator)
		vec := make([]float64, len(elements), len(elements)+1)
		for i, e := range elements {
			val, err := strconv.ParseFloat(strings.TrimSpace(e), 64)
			// Reject NaN/Inf too: ParseFloat accepts them without error, and a
			// single invalid element makes the whole vector unusable, so return
			// nil to route it to the missing-embedding fallback instead of
			// fabricating a 0/NaN vector that would normalize to NaN and pollute
			// the shared cache and kernel matrix.
			if err != nil || math.IsNaN(val) || math.IsInf(val, 0) {
				ctx.LogError(fmt.Sprintf("module=%s\tparse embedding value failed\tvalue=%s\terr=%v", d.logModule, e, err))
				return nil
			}
			vec[i] = val
		}
		return vec
	default:
		// Unsupported type: return nil so the caller treats it as a missing
		// embedding and routes it to the fallback path, instead of fabricating a
		// zero vector that would normalize to NaN and silently corrupt the ranking.
		return nil
	}
}
