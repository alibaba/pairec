package module

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

type sharedBehaviorFeatures struct {
	once     sync.Once
	features []map[string]any
	err      error
}

// writeLengthPrefixed appends one length-prefixed segment so that values
// containing the separator bytes can never collide with another key.
func writeLengthPrefixed(b *strings.Builder, value string) {
	b.WriteString(strconv.Itoa(len(value)))
	b.WriteByte(':')
	b.WriteString(value)
	b.WriteByte('\x1e')
}

// initBehaviorShareKeyPrefix precomputes the request-invariant part of the
// sharing key (feature store, view and event set) once per DAO, so the request
// path only appends the user id, view name and select fields.
func (d *RealtimeUser2ItemFeatureStoreDao) initBehaviorShareKeyPrefix() {
	var b strings.Builder
	b.Grow(128)
	b.WriteString("realtime_u2i_behavior\x1f")
	writeLengthPrefixed(&b, d.featureStoreName)
	writeLengthPrefixed(&b, d.userTriggerTable)
	b.WriteByte('\x1f')
	events := make([]string, len(d.events))
	for i, event := range d.events {
		events[i] = fmt.Sprint(event)
	}
	sort.Strings(events)
	for _, event := range events {
		writeLengthPrefixed(&b, event)
	}
	b.WriteByte('\x1f')
	d.behaviorShareKeyPrefix = b.String()
}

func (d *RealtimeUser2ItemFeatureStoreDao) behaviorShareKey(view domain.FeatureView, ids []any, fields []string) string {
	var b strings.Builder
	b.Grow(len(d.behaviorShareKeyPrefix) + 64)
	b.WriteString(d.behaviorShareKeyPrefix)
	// The view name identifies the target view; ids keep their original order
	// because the SDK builds KKV primary keys in that order.
	writeLengthPrefixed(&b, view.GetName())
	for _, id := range ids {
		writeLengthPrefixed(&b, fmt.Sprint(id))
	}
	b.WriteByte('\x1f')
	// Field counts are small, so sort on the stack to avoid an allocation.
	var fieldBuf [8]string
	selectFields := fieldBuf[:0]
	for _, field := range fields {
		selectFields = append(selectFields, field)
	}
	sort.Strings(selectFields)
	for _, field := range selectFields {
		writeLengthPrefixed(&b, field)
	}
	return b.String()
}

func (d *RealtimeUser2ItemFeatureStoreDao) getSharedBehaviorFeatures(view domain.FeatureView, ids []any, fields []string, ctx *context.RecommendContext) ([]map[string]any, error) {
	start := time.Now()
	actual, _ := ctx.LoadOrStoreContextParam(d.behaviorShareKey(view, ids, fields), &sharedBehaviorFeatures{})
	entry := actual.(*sharedBehaviorFeatures)
	from := "reuse"
	entry.once.Do(func() {
		from = "fetch"
		// panic 仍向上传递；Once 完成后等待者会收到错误而非误判为空数据。
		entry.err = errors.New("shared behavior fetch did not complete")
		entry.features, entry.err = view.GetBehaviorFeatures(ids, d.events, fields)
	})
	// 仅共享路径新增日志，不向并发共用的 ctx.Log 切片写入。
	log.Info(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemFeatureStoreDao\tname=%s\tview=%s\tbehavior_share=%s\tcount=%d\tcost=%d",
		ctx.RecommendId, d.recallName, d.userTriggerTable, from, len(entry.features), utils.CostTime(start)))
	return entry.features, entry.err
}
