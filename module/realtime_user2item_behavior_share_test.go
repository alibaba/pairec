package module

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/Knetic/govaluate"
	reccontext "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

type behaviorShareView struct {
	domain.FeatureView
	name  string
	calls atomic.Int32
	fetch func([]any, []any, []string) ([]map[string]any, error)
}

func (v *behaviorShareView) GetName() string {
	if v.name == "" {
		return "behavior"
	}
	return v.name
}

func (v *behaviorShareView) GetBehaviorFeatures(ids, events []any, fields []string) ([]map[string]any, error) {
	v.calls.Add(1)
	return v.fetch(ids, events, fields)
}

func behaviorShareTestDao(t testing.TB, view *behaviorShareView) *RealtimeUser2ItemFeatureStoreDao {
	t.Helper()
	project := &domain.Project{}
	project.FeatureViewMap.Store("behavior", view)
	client := &fs.FSClient{}
	// 仅测试注入项目，不为生产 FSClient 增加测试接口。
	field := reflect.ValueOf(client).Elem().FieldByName("project")
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(project))
	expression, err := govaluate.NewEvaluableExpression("1")
	if err != nil {
		t.Fatal(err)
	}
	return &RealtimeUser2ItemFeatureStoreDao{
		RealtimeUser2ItemBaseDao: &RealtimeUser2ItemBaseDao{
			recallName: "test", triggerCount: 10, eventWeightMap: map[string]float64{"click": 1},
		},
		fsClient: client, featureStoreName: "test_fs", userTriggerTable: "behavior",
		itemIdFieldName: "item_id", eventFieldName: "event", timestampFieldName: "timestamp",
		weightMode: weight_mode_sum, weightEvaluableExpression: expression,
		events: []any{"click"}, shareBehaviorInRequest: true,
		behaviorShareKeyPrefix: buildTestShareKeyPrefix("test_fs", "behavior", []any{"click"}),
	}
}

// buildTestShareKeyPrefix 与生产 initBehaviorShareKeyPrefix 同口径，供直接构造 DAO 的用例使用。
func buildTestShareKeyPrefix(store, view string, events []any) string {
	dao := &RealtimeUser2ItemFeatureStoreDao{featureStoreName: store, userTriggerTable: view, events: events}
	dao.initBehaviorShareKeyPrefix()
	return dao.behaviorShareKeyPrefix
}

func TestBehaviorShareKey(t *testing.T) {
	view := &behaviorShareView{}
	dao := behaviorShareTestDao(t, view)
	dao.events = []any{"add_cart", "click"}
	dao.initBehaviorShareKeyPrefix()
	ids := []any{UID("u"), "extra"}
	fields := []string{"event", "item_id"}
	key := dao.behaviorShareKey(view, ids, fields)
	dao.events = []any{"click", "add_cart"}
	dao.initBehaviorShareKeyPrefix()
	if other := dao.behaviorShareKey(view, ids, []string{"item_id", "event"}); key != other {
		t.Fatal("事件或字段顺序影响签名")
	}
	if !reflect.DeepEqual(dao.events, []any{"click", "add_cart"}) || !reflect.DeepEqual(fields, []string{"event", "item_id"}) {
		t.Fatal("签名计算修改了原始切片")
	}
	for name, mutate := range map[string]func(*RealtimeUser2ItemFeatureStoreDao){
		"store": func(d *RealtimeUser2ItemFeatureStoreDao) {
			d.featureStoreName += "other"
			d.initBehaviorShareKeyPrefix()
		},
		"view": func(d *RealtimeUser2ItemFeatureStoreDao) {
			d.userTriggerTable += "other"
			d.initBehaviorShareKeyPrefix()
		},
		"events": func(d *RealtimeUser2ItemFeatureStoreDao) {
			d.events = []any{"click"}
			d.initBehaviorShareKeyPrefix()
		},
	} {
		t.Run(name, func(t *testing.T) {
			d := *dao
			mutate(&d)
			if d.behaviorShareKey(view, ids, fields) == key {
				t.Fatal("不同取数来源或事件未隔离")
			}
		})
	}
	for name, other := range map[string]string{
		"uid":      dao.behaviorShareKey(view, []any{"other", "extra"}, fields),
		"addition": dao.behaviorShareKey(view, []any{"u", "other"}, fields),
		"fields":   dao.behaviorShareKey(view, ids, []string{"item_id"}),
		"viewName": dao.behaviorShareKey(&behaviorShareView{name: "another_view"}, ids, fields),
	} {
		if other == key {
			t.Fatalf("%s 未隔离", name)
		}
	}
	// 视图身份只用名字：同名的新对象（项目刷新后重建）视为同一视图，合并读取。
	if shared := dao.behaviorShareKey(&behaviorShareView{}, ids, fields); shared != key {
		t.Fatal("同名视图不同对象未能共享")
	}
	if dao.behaviorShareKey(view, []any{"u\x1eextra"}, fields) == key {
		t.Fatal("分隔符导致多主键碰撞")
	}
	if dao.behaviorShareKey(view, []any{"u3:abc"}, fields) == dao.behaviorShareKey(view, []any{"u"}, []string{"abc"}) {
		t.Fatal("长度前缀未能防止跨字段碰撞")
	}
	if dao.behaviorShareKey(view, []any{123}, fields) != dao.behaviorShareKey(view, []any{"123"}, fields) {
		t.Fatal("SDK 使用相同主键字符串的值应可共享")
	}
}

func TestBehaviorShareConcurrentAndLate(t *testing.T) {
	rows := []map[string]any{{"item_id": "item", "event": "click", "timestamp": int64(10)}}
	entered, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	view := &behaviorShareView{fetch: func([]any, []any, []string) ([]map[string]any, error) {
		close(entered)
		<-release
		return rows, nil
	}}
	dao := behaviorShareTestDao(t, view)
	ctx := reccontext.NewRecommendContext()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := dao.getSharedBehaviorFeatures(view, []any{"u"}, []string{"item_id"}, ctx)
			if err != nil || !reflect.DeepEqual(got, rows) {
				t.Errorf("共享结果错误：%v %v", got, err)
			}
		}()
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("读取未开始")
	}
	releaseOnce.Do(func() { close(release) })
	wg.Wait()
	for i := 0; i < 3; i++ {
		if _, err := dao.getSharedBehaviorFeatures(view, []any{"u"}, []string{"item_id"}, ctx); err != nil {
			t.Fatal(err)
		}
	}
	if view.calls.Load() != 1 {
		t.Fatalf("同请求读取%d次，期望1", view.calls.Load())
	}
}

func TestBehaviorShareGroupsAndRequests(t *testing.T) {
	view := &behaviorShareView{fetch: func([]any, []any, []string) ([]map[string]any, error) { return nil, nil }}
	dao := behaviorShareTestDao(t, view)
	ctx := reccontext.NewRecommendContext()
	for _, events := range [][]any{{"click"}, {"add_cart"}, {"click"}, {"add_cart"}} {
		d := *dao
		d.events = events
		d.initBehaviorShareKeyPrefix()
		if _, err := d.getSharedBehaviorFeatures(view, []any{"u"}, nil, ctx); err != nil {
			t.Fatal(err)
		}
	}
	if view.calls.Load() != 2 {
		t.Fatalf("两组空结果读取%d次，期望2", view.calls.Load())
	}
	if _, err := dao.getSharedBehaviorFeatures(view, []any{"u"}, nil, reccontext.NewRecommendContext()); err != nil {
		t.Fatal(err)
	}
	if view.calls.Load() != 3 {
		t.Fatal("跨请求复用了数据")
	}
}

func TestBehaviorShareErrorsAndPanic(t *testing.T) {
	for _, panics := range []bool{false, true} {
		t.Run(fmt.Sprint(panics), func(t *testing.T) {
			want := errors.New("fetch failed")
			view := &behaviorShareView{fetch: func([]any, []any, []string) ([]map[string]any, error) {
				if panics {
					panic(want)
				}
				return nil, want
			}}
			dao := behaviorShareTestDao(t, view)
			ctx := reccontext.NewRecommendContext()
			func() {
				defer func() {
					if p := recover(); panics && p != want {
						t.Errorf("未保留原 panic：%v", p)
					}
				}()
				_, err := dao.getSharedBehaviorFeatures(view, []any{"u"}, nil, ctx)
				if panics || !errors.Is(err, want) {
					t.Errorf("首个读取结果不正确：%v", err)
				}
			}()
			finished := make(chan error, 1)
			go func() {
				_, err := dao.getSharedBehaviorFeatures(view, []any{"u"}, nil, ctx)
				finished <- err
			}()
			select {
			case err := <-finished:
				if err == nil || (!panics && err != want) {
					t.Errorf("等待者未收到共享错误：%v", err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("panic/error 后等待者未返回")
			}
			if view.calls.Load() != 1 {
				t.Fatal("失败或 panic 后重复读取")
			}
		})
	}
}

func TestBehaviorShareDifferentKeyDoesNotWait(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	view := &behaviorShareView{fetch: func(ids []any, _ []any, _ []string) ([]map[string]any, error) {
		if ids[0] == "slow" {
			close(entered)
			<-release
		}
		return nil, nil
	}}
	dao := behaviorShareTestDao(t, view)
	ctx := reccontext.NewRecommendContext()
	done := make(chan struct{})
	go func() {
		dao.getSharedBehaviorFeatures(view, []any{"slow"}, nil, ctx)
		close(done)
	}()
	<-entered
	fast := make(chan struct{})
	go func() {
		dao.getSharedBehaviorFeatures(view, []any{"fast"}, nil, ctx)
		close(fast)
	}()
	select {
	case <-fast:
	case <-time.After(2 * time.Second):
		t.Fatal("不同 key 被慢读取阻塞")
	}
	once.Do(func() { close(release) })
	<-done
}

func TestBehaviorShareTriggerCalculation(t *testing.T) {
	rows := []map[string]any{
		{"item_id": "a", "event": "click", "timestamp": int64(10)},
		{"item_id": "a", "event": "add_cart", "timestamp": int64(20)},
		{"item_id": "b", "event": "click", "timestamp": int64(40)},
		{"item_id": "c", "event": "add_cart", "timestamp": int64(30)},
	}
	view := &behaviorShareView{fetch: func([]any, []any, []string) ([]map[string]any, error) { return rows, nil }}
	first := behaviorShareTestDao(t, view)
	second := behaviorShareTestDao(t, view)
	first.events, second.events = []any{"click", "add_cart"}, []any{"add_cart", "click"}
	first.eventWeightMap = map[string]float64{"click": 1, "add_cart": 2}
	second.eventWeightMap = map[string]float64{"click": 5, "add_cart": 1}
	first.triggerCount, second.triggerCount = 2, 3
	first.weightMode, second.weightMode = weight_mode_sum, weight_mode_max
	first.events, second.events = []any{"click", "add_cart"}, []any{"add_cart", "click"}
	first.initBehaviorShareKeyPrefix()
	second.initBehaviorShareKeyPrefix()
	for _, d := range []*RealtimeUser2ItemFeatureStoreDao{first, second} {
		expression := "eventTime"
		if d == second {
			expression = "eventTime * 2"
		}
		var err error
		d.weightEvaluableExpression, err = govaluate.NewEvaluableExpression(expression)
		if err != nil {
			t.Fatal(err)
		}
	}
	user := NewUser("u")
	want := make([][]*TriggerInfo, 2)
	for i, d := range []*RealtimeUser2ItemFeatureStoreDao{first, second} {
		d.shareBehaviorInRequest = false
		want[i] = d.GetTriggerInfos(user, reccontext.NewRecommendContext())
		d.shareBehaviorInRequest = true
	}
	ctx := reccontext.NewRecommendContext()
	for i, d := range []*RealtimeUser2ItemFeatureStoreDao{first, second} {
		got := d.GetTriggerInfos(user, ctx)
		if !reflect.DeepEqual(got, want[i]) {
			t.Fatalf("第%d路计算改变：%+v vs %+v", i, got, want[i])
		}
	}
	if view.calls.Load() != 3 || len(want[0]) != 2 || len(want[1]) != 3 {
		t.Fatalf("读取次数或截断错误：calls=%d", view.calls.Load())
	}
	if want[0][0].ItemId != "c" || want[0][0].Weight != 60 || want[1][0].ItemId != "b" || want[1][0].Weight != 400 {
		t.Fatal("不同权重/聚合的期望结果错误")
	}
	if rows[0]["item_id"] != "a" || len(rows[0]) != 3 || rows[1]["timestamp"] != int64(20) {
		t.Fatal("原始行为数据被修改")
	}
}

func TestBehaviorShareAdditionUidAndFields(t *testing.T) {
	view := &behaviorShareView{fetch: func(ids, _ []any, fields []string) ([]map[string]any, error) {
		if len(ids) != 2 || ids[0] != UID("u") || ids[1] != "second" {
			t.Errorf("AdditionUid 实际入参错误：%v", ids)
		}
		return nil, nil
	}}
	first := behaviorShareTestDao(t, view)
	first.additionUid = "uid2"
	user := NewUser("u")
	user.AddProperty("uid2", "second")
	ctx := reccontext.NewRecommendContext()
	first.GetTriggerInfos(user, ctx)
	first.GetTriggerInfos(user, ctx)
	second := *first
	second.propertyFields = []string{"page"}
	second.GetTriggerInfos(user, ctx)
	second.hasPlayTimeField = true
	second.playtimeFieldName = "playtime"
	second.GetTriggerInfos(user, ctx)
	if view.calls.Load() != 3 {
		t.Fatal("不同 PropertyFields/playtime 未独立读取")
	}
}

func TestBehaviorShareDisabled(t *testing.T) {
	view := &behaviorShareView{fetch: func([]any, []any, []string) ([]map[string]any, error) { return nil, nil }}
	dao := behaviorShareTestDao(t, view)
	ctx := reccontext.NewRecommendContext()
	for i := 0; i < 2; i++ {
		dao.GetTriggerInfos(NewUser("u"), ctx)
	}
	dao.shareBehaviorInRequest = false
	for i := 0; i < 2; i++ {
		dao.GetTriggerInfos(NewUser("u"), ctx)
	}
	if view.calls.Load() != 3 {
		t.Fatal("关闭路复用了共享结果")
	}
}

func BenchmarkBehaviorShareCompleted(b *testing.B) {
	view := &behaviorShareView{fetch: func([]any, []any, []string) ([]map[string]any, error) { return nil, nil }}
	dao := behaviorShareTestDao(b, view)
	ctx := reccontext.NewRecommendContext()
	ids, fields := []any{"user_0900601"}, []string{"item_id", "event", "timestamp"}
	dao.getSharedBehaviorFeatures(view, ids, fields, ctx)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dao.getSharedBehaviorFeatures(view, ids, fields, ctx)
	}
}
