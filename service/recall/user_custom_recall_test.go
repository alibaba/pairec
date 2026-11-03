package recall

import (
	"testing"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/persist/cache"
	"github.com/alibaba/pairec/v2/recconf"
)

type userCustomRecallTestDao struct {
	calls int
	items []*module.Item
}

func (d *userCustomRecallTestDao) ListItemsByUser(_ *module.User, _ *context.RecommendContext) []*module.Item {
	d.calls++
	return d.items
}

type userCustomRecallTestPut struct {
	key   string
	value interface{}
	ttl   time.Duration
}

type userCustomRecallTestCache struct {
	cache.Cache
	puts chan userCustomRecallTestPut
}

func (c *userCustomRecallTestCache) Put(key string, value interface{}, ttl time.Duration) error {
	err := c.Cache.Put(key, value, ttl)
	c.puts <- userCustomRecallTestPut{key: key, value: value, ttl: ttl}
	return err
}

func newUserCustomRecallCacheTest(t *testing.T, cacheTime int) (*UserCustomRecall, *userCustomRecallTestDao, *userCustomRecallTestCache) {
	t.Helper()
	base := NewBaseRecall(recconf.RecallConfig{
		Name:         "U2IRecall",
		ItemType:     "video",
		CacheAdapter: "localCache",
		CacheConfig:  `{"defaultExpiration":600,"cleanupInterval":0}`,
		CachePrefix:  "u2i_",
		CacheTime:    cacheTime,
	})
	if base.cache == nil {
		t.Fatal("localCache was not initialized")
	}
	c := &userCustomRecallTestCache{Cache: base.cache, puts: make(chan userCustomRecallTestPut, 8)}
	base.cache = c
	dao := &userCustomRecallTestDao{}
	return &UserCustomRecall{BaseRecall: base, userCustomRecallDao: dao}, dao, c
}

func waitUserCustomRecallCachePut(t *testing.T, c *userCustomRecallTestCache) userCustomRecallTestPut {
	t.Helper()
	select {
	case put := <-c.puts:
		return put
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for asynchronous cache write")
		return userCustomRecallTestPut{}
	}
}

func TestUserCustomRecallCacheHitTypes(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value interface{}
	}{
		{name: "localCache_string", value: "item_1::0.9,item_2"},
		{name: "redis_bytes", value: []byte("item_1::0.9,item_2")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, dao, c := newUserCustomRecallCacheTest(t, 600)
			if err := c.Cache.Put("u2i_user_1", tc.value, time.Minute); err != nil {
				t.Fatal(err)
			}
			items := r.GetCandidateItems(module.NewUser("user_1"), context.NewRecommendContext())
			if dao.calls != 0 {
				t.Fatalf("cache hit called DAO %d times", dao.calls)
			}
			if len(items) != 2 {
				t.Fatalf("got %d cached items, want 2", len(items))
			}
			if items[0].Id != "item_1" || items[0].Score != 0.9 || items[1].Id != "item_2" || items[1].Score != 0 {
				t.Fatalf("unexpected cached IDs or scores: %v, %v", items[0], items[1])
			}
			for _, item := range items {
				if item.ItemType != "video" || item.RetrieveId != "U2IRecall" {
					t.Fatalf("unexpected cached metadata: type=%s recall=%s", item.ItemType, item.RetrieveId)
				}
			}
			select {
			case <-c.puts:
				t.Fatal("cache hit must not write the cache again")
			default:
			}
		})
	}
}

func TestUserCustomRecallLocalCacheMissThenHit(t *testing.T) {
	r, dao, c := newUserCustomRecallCacheTest(t, 600)
	item := module.NewItem("item_1")
	item.Score = 0.75
	dao.items = []*module.Item{item}
	ctx := context.NewRecommendContext()
	user := module.NewUser("user_1")
	first := r.GetCandidateItems(user, ctx)
	if len(first) != 1 || first[0] != item {
		t.Fatal("cache miss must return the DAO result unchanged")
	}
	put := waitUserCustomRecallCachePut(t, c)
	if put.key != "u2i_user_1" || put.value != "item_1::0.75" || put.ttl != 600*time.Second {
		t.Fatalf("unexpected cache write: %+v", put)
	}
	second := r.GetCandidateItems(user, ctx)
	if dao.calls != 1 {
		t.Fatalf("DAO called %d times, want 1 after repeat uid", dao.calls)
	}
	if len(second) != 1 || second[0].Id != item.Id || second[0].Score != item.Score || second[0] == item {
		t.Fatal("cache hit must reconstruct an independent item with the same ID and score")
	}
	r.GetCandidateItems(module.NewUser("user_2"), ctx)
	waitUserCustomRecallCachePut(t, c)
	if dao.calls != 2 {
		t.Fatalf("DAO called %d times, want 2 after another uid", dao.calls)
	}
}

func TestUserCustomRecallCacheTime(t *testing.T) {
	for _, tc := range []struct {
		name string
		secs int
		want time.Duration
	}{
		{name: "configured", secs: 600, want: 600 * time.Second},
		{name: "default", secs: 0, want: 1800 * time.Second},
		{name: "nonpositive_default", secs: -1, want: 1800 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, dao, c := newUserCustomRecallCacheTest(t, tc.secs)
			dao.items = []*module.Item{module.NewItem("item_1")}
			r.GetCandidateItems(module.NewUser("user_1"), context.NewRecommendContext())
			if put := waitUserCustomRecallCachePut(t, c); put.ttl != tc.want {
				t.Fatalf("Put TTL=%s, want %s", put.ttl, tc.want)
			}
		})
	}
}

func TestUserCustomRecallLocalCacheExpiry(t *testing.T) {
	r, dao, c := newUserCustomRecallCacheTest(t, 1)
	dao.items = []*module.Item{module.NewItem("item_1")}
	ctx := context.NewRecommendContext()
	user := module.NewUser("user_1")
	r.GetCandidateItems(user, ctx)
	put := waitUserCustomRecallCachePut(t, c)
	deadline := time.Now().Add(2 * time.Second)
	for c.Get(put.key) != nil {
		if time.Now().After(deadline) {
			t.Fatal("CacheTime=1 did not expire the entry; CacheConfig defaultExpiration must not override it")
		}
		time.Sleep(10 * time.Millisecond)
	}
	r.GetCandidateItems(user, ctx)
	waitUserCustomRecallCachePut(t, c)
	if dao.calls != 2 {
		t.Fatalf("expired entry called DAO %d times, want 2", dao.calls)
	}
}

func TestUserCustomRecallCacheFallback(t *testing.T) {
	t.Run("unsupported_value", func(t *testing.T) {
		r, dao, c := newUserCustomRecallCacheTest(t, 600)
		if err := c.Cache.Put("u2i_user_1", 123, time.Minute); err != nil {
			t.Fatal(err)
		}
		r.GetCandidateItems(module.NewUser("user_1"), context.NewRecommendContext())
		if dao.calls != 1 {
			t.Fatalf("unsupported cached type called DAO %d times, want 1", dao.calls)
		}
		select {
		case <-c.puts:
			t.Fatal("empty DAO result must retain existing no-cache behavior")
		default:
		}
	})
	t.Run("disabled", func(t *testing.T) {
		dao := &userCustomRecallTestDao{items: []*module.Item{module.NewItem("item_1")}}
		r := &UserCustomRecall{
			BaseRecall:          NewBaseRecall(recconf.RecallConfig{Name: "U2IRecall"}),
			userCustomRecallDao: dao,
		}
		for i := 0; i < 2; i++ {
			items := r.GetCandidateItems(module.NewUser("user_1"), context.NewRecommendContext())
			if len(items) != 1 || items[0] != dao.items[0] {
				t.Fatal("disabled cache must return DAO result unchanged")
			}
		}
		if dao.calls != 2 {
			t.Fatalf("disabled cache called DAO %d times, want 2", dao.calls)
		}
	})
}
