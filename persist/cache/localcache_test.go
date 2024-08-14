package cache

import (
	"testing"
	"time"
)

func TestLocalCache(t *testing.T) {
	tests := []struct {
		Key    string
		Value  string
		Expire time.Duration
		Want   bool
	}{
		{
			"foo",
			"var",
			10 * time.Second,
			true,
		},
	}

	config := "{\"defaultExpiration\":1800, \"cleanupInterval\":1800}"
	cache, err := NewCache("localCache", config)
	if err != nil {
		t.Error(err)
	}
	for _, test := range tests {
		cache.Put(test.Key, test.Value, test.Expire)
		v := cache.DefaultGet(test.Key, "")
		if val, ok := v.(string); !ok {
			t.Errorf("get cache error, key=%s", test.Key)
		} else if val != test.Value {
			t.Errorf("get cache error, key=%s, want=%v, get=%v", test.Key, test.Want, false)
		}
	}
	for _, test := range tests {
		cache.Put(test.Key, test.Value, test.Expire)
		time.Sleep(test.Expire + time.Second)
		v := cache.DefaultGet(test.Key, "")
		if val, ok := v.(string); !ok {
			t.Errorf("get cache error, key=%s", test.Key)
		} else if val == test.Value {
			t.Errorf("get cache error, key=%s, want nothing, get=%s", test.Key, val)
		}
	}
}
