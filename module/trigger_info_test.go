package module

import (
	"database/sql"
	"testing"

	"github.com/alibaba/pairec/v2/recconf"
)

func newTestTriggerInfo(itemId string, weight float64, propertyValues ...string) *TriggerInfo {
	info := &TriggerInfo{
		ItemId: itemId,
		Weight: weight,
	}
	for _, v := range propertyValues {
		info.propertyFieldValues = append(info.propertyFieldValues, sql.NullString{String: v, Valid: true})
	}
	return info
}

func TestAggregateTriggerPropertyWeights(t *testing.T) {
	propertyFieldMap := map[string]int{"tags": 0, "category": 1}

	t.Run("multi value split and weight sum", func(t *testing.T) {
		triggerInfos := []*TriggerInfo{
			newTestTriggerInfo("item1", 3, "tag1|tag2", "c1"),
			newTestTriggerInfo("item2", 2, "tag2|tag3", "c2"),
			newTestTriggerInfo("item3", 1, "tag3", "c3"),
		}
		ret := AggregateTriggerPropertyWeights(triggerInfos, "tags", propertyFieldMap, "|")
		if len(ret) != 3 {
			t.Fatalf("expect 3 property triggers, got %d", len(ret))
		}
		weights := make(map[string]float64, len(ret))
		for _, info := range ret {
			weights[info.ItemId] = info.Weight
		}
		if weights["tag1"] != 3 || weights["tag2"] != 5 || weights["tag3"] != 3 {
			t.Fatalf("unexpected weights: %v", weights)
		}
		// sorted by weight in descending order
		if ret[0].ItemId != "tag2" {
			t.Fatalf("expect tag2 first, got %s", ret[0].ItemId)
		}
	})

	t.Run("single value without delimiter", func(t *testing.T) {
		triggerInfos := []*TriggerInfo{
			newTestTriggerInfo("item1", 2, "tag1", "c1"),
			newTestTriggerInfo("item2", 1, "tag1", "c1"),
		}
		ret := AggregateTriggerPropertyWeights(triggerInfos, "category", propertyFieldMap, "")
		if len(ret) != 1 || ret[0].ItemId != "c1" || ret[0].Weight != 3 {
			t.Fatalf("unexpected result: %+v", ret)
		}
	})

	t.Run("empty and null values are skipped", func(t *testing.T) {
		triggerInfos := []*TriggerInfo{
			newTestTriggerInfo("item1", 2, "", "c1"),
			newTestTriggerInfo("item2", 1, "null", "c2"),
			newTestTriggerInfo("item3", 1, "NULL", "c3"),
			newTestTriggerInfo("item4", 1, "Null", "c4"),
			newTestTriggerInfo("item5", 1, "tag1|", "c5"),
		}
		ret := AggregateTriggerPropertyWeights(triggerInfos, "tags", propertyFieldMap, "|")
		if len(ret) != 1 || ret[0].ItemId != "tag1" || ret[0].Weight != 1 {
			t.Fatalf("unexpected result: %+v", ret)
		}
	})

	t.Run("xKey not in propertyFieldMap returns empty", func(t *testing.T) {
		triggerInfos := []*TriggerInfo{
			newTestTriggerInfo("item1", 2, "tag1", "c1"),
		}
		ret := AggregateTriggerPropertyWeights(triggerInfos, "not_exist", propertyFieldMap, "|")
		if len(ret) != 0 {
			t.Fatalf("expect empty result, got %+v", ret)
		}
	})
}

func TestAggregateItem2XWeights(t *testing.T) {
	triggerInfos := []*TriggerInfo{
		{ItemId: "item1", Weight: 3},
		{ItemId: "item2", Weight: 2},
		{ItemId: "item3", Weight: 1}, // not in item2XMap, skipped
	}
	item2XMap := map[string]string{
		"item1": "tag1,tag2",
		"item2": "tag2",
	}
	ret := AggregateItem2XWeights(triggerInfos, item2XMap, ",")
	if len(ret) != 2 {
		t.Fatalf("expect 2 property triggers, got %d", len(ret))
	}
	if ret[0].ItemId != "tag2" || ret[0].Weight != 5 {
		t.Fatalf("expect tag2:5 first, got %s:%f", ret[0].ItemId, ret[0].Weight)
	}
	if ret[1].ItemId != "tag1" || ret[1].Weight != 3 {
		t.Fatalf("expect tag1:3 second, got %s:%f", ret[1].ItemId, ret[1].Weight)
	}
}

func TestIsTriggerOutOfTimeWindow(t *testing.T) {
	now := int64(1700000000)

	// timeWindow not configured, no filter
	if isTriggerOutOfTimeWindow(now-999999, now, 0) {
		t.Fatal("expect no filter when timeWindow is 0")
	}
	// behavior within the time window is kept
	if isTriggerOutOfTimeWindow(now-100, now, 3600) {
		t.Fatal("expect behavior within time window to be kept")
	}
	// behavior at the window boundary is kept
	if isTriggerOutOfTimeWindow(now-3600, now, 3600) {
		t.Fatal("expect behavior at window boundary to be kept")
	}
	// behavior out of the time window is filtered
	if !isTriggerOutOfTimeWindow(now-3601, now, 3600) {
		t.Fatal("expect behavior out of time window to be filtered")
	}
	// timestamp 0 means no valid time info, keep it even when timeWindow is configured
	if isTriggerOutOfTimeWindow(0, now, 3600) {
		t.Fatal("expect behavior with zero timestamp to be kept")
	}
	// negative timestamp is also treated as no valid time info
	if isTriggerOutOfTimeWindow(-1, now, 3600) {
		t.Fatal("expect behavior with negative timestamp to be kept")
	}
}

func TestItem2XCache(t *testing.T) {
	t.Run("cache disabled when CacheSize not configured", func(t *testing.T) {
		if c := newItem2XCache(recconf.Item2XConfig{}); c != nil {
			t.Fatal("expect nil cache when CacheSize is 0")
		}
		item2XMap := make(map[string]string)
		missedIds := fetchItem2XFromCache(nil, []string{"item1", "item2"}, item2XMap)
		if len(missedIds) != 2 {
			t.Fatalf("expect all ids missed without cache, got %v", missedIds)
		}
	})

	t.Run("cache hit and negative cache", func(t *testing.T) {
		c := newItem2XCache(recconf.Item2XConfig{CacheSize: 100, CacheTime: 60})
		if c == nil {
			t.Fatal("expect cache created")
		}
		// item1 has property value, item2 has no data (negative cache)
		putItem2XToCache(c, []string{"item1", "item2"}, map[string]string{"item1": "tag1|tag2"})

		item2XMap := make(map[string]string)
		missedIds := fetchItem2XFromCache(c, []string{"item1", "item2", "item3"}, item2XMap)
		if len(missedIds) != 1 || missedIds[0] != "item3" {
			t.Fatalf("expect only item3 missed, got %v", missedIds)
		}
		if item2XMap["item1"] != "tag1|tag2" {
			t.Fatalf("expect item1 hit cache, got %v", item2XMap)
		}
		if _, exist := item2XMap["item2"]; exist {
			t.Fatal("expect item2 negative cached and not in result")
		}
	})
}
