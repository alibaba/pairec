package bloomfilter

import (
	"sync"

	"github.com/gomodule/redigo/redis"
)

const redisMaxLength = 8 * 512 * 1024 * 1024

type ResetFunc func(conn redis.Conn)
type RedisBitSet struct {
	keyPrefix  string
	redisPool  *redis.Pool
	batchCount int // merge request size to invoke redis
	resetFunc  ResetFunc
	online     bool
	lock       sync.RWMutex // when clear the db, block the BF.MADD until clear db finish
}

func NewRedisBitSet(redis *redis.Pool) *RedisBitSet {
	return &RedisBitSet{redisPool: redis, batchCount: 500, online: true}
}

func (r *RedisBitSet) SetResetFunc(f ResetFunc) {
	r.resetFunc = f
}
func (r *RedisBitSet) SetBatchCount(c int) {
	r.batchCount = c
}
func (r *RedisBitSet) Set(key string, offsets [][]uint) error {
	conn := r.redisPool.Get()
	defer conn.Close()
	args := []interface{}{key}
	for _, offset := range offsets {
		for _, val := range offset {
			args = append(args, "SET", "u1", val, 1)
		}
	}

	// batch set bits
	_, err := conn.Do("BITFIELD", args...)
	return err
}

func (r *RedisBitSet) Test(key string, offsets [][]uint) ([]bool, error) {

	k := len(offsets[0])
	ret := make([]bool, len(offsets))
	// num := len(offsets)
	for i := range ret {
		ret[i] = true
	}

	count := r.batchCount

	n := 0
	keys := make([]int, 0, len(offsets))
	for j := 0; j < k; j++ {
		mapOffset := make(map[int]uint, len(keys))
		// collect offset
		for i, offset := range offsets {
			if ret[i] == true {
				mapOffset[i] = offset[j]
			}
		}

		if len(mapOffset) == 0 {
			break
		}
		keys = keys[:0]
		for k := range mapOffset {
			keys = append(keys, k)
		}

		n = len(mapOffset) / count

		var wg sync.WaitGroup
		for i := 0; i <= n; i++ {
			wg.Add(1)
			go func(num int, keys []int, mapOffset map[int]uint) {
				defer wg.Done()
				conn := r.redisPool.Get()
				defer conn.Close()
				args := make([]interface{}, 0, count*3+1)
				args = append(args, key)
				for i := num * count; i < (num+1)*count && i < len(keys); i++ {
					args = append(args, "GET", "u1", mapOffset[keys[i]])
				}
				if len(args) > 1 {
					reply, err := redis.Ints(conn.Do("BITFIELD", args...))
					// fmt.Println(num, len(keys), len(args), err)
					if err == nil {
						for ii, intVal := range reply {
							if intVal == 0 {
								ret[keys[num*count+ii]] = false
							}
						}
					}
				}

			}(i, keys, mapOffset)
		}
		wg.Wait()
	}

	return ret, nil
}

func (r *RedisBitSet) redisBloomSet(key string, data [][]byte) error {
	// empty data
	if len(data) == 0 {
		return nil
	}
	if false == r.online {
		return nil
	}
	conn := r.redisPool.Get()
	defer conn.Close()
	args := []interface{}{key}
	for _, d := range data {
		args = append(args, string(d))
	}

	// batch set bits
	_, err := conn.Do("BF.MADD", args...)
	return err
}
func (r *RedisBitSet) redisBloomTest(key string, data [][]byte) ([]bool, error) {
	conn := r.redisPool.Get()
	defer conn.Close()
	args := make([]interface{}, 0, len(data)+2)
	args = append(args, key)
	for _, d := range data {
		args = append(args, string(d))
	}
	ret := make([]bool, len(data))
	// empty data
	if len(data) == 0 {
		return ret, nil
	}
	reply, err := redis.Ints(conn.Do("BF.MEXISTS", args...))

	for i := range reply {
		if reply[i] == 1 {
			ret[i] = true
		}
	}

	return ret, err
}
func (r *RedisBitSet) Clear() {
	conn := r.redisPool.Get()
	defer conn.Close()
	if r.resetFunc != nil {
		r.resetFunc(conn)
	}

}
func (r *RedisBitSet) Online(online bool) {
	r.online = online
}
