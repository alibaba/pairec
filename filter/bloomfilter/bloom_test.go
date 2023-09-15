package bloomfilter

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/spaolacci/murmur3"
)

var (
	seeds = []uint32{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97, 101, 103, 107, 109, 113, 127, 131, 137, 139, 149, 151, 157, 163, 167, 173, 179, 181, 191, 193, 197, 199}
)

func hash(data [][]byte, m uint, k uint) (ret [][]uint) {
	locations := make([][]uint, 0, len(data))
	for _, d := range data {
		ret := make([]uint, 0, k)
		for i := uint(0); i < k; i++ {
			mmh3 := murmur3.New32WithSeed(seeds[i])
			mmh3.Write(d)
			ret = append(ret, uint(mmh3.Sum32())%m)
		}
		locations = append(locations, ret)
	}
	return locations
}
func TestBitSet(t *testing.T) {
	pool := &redis.Pool{
		MaxIdle:     30,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", "127.0.0.1:6379") },
	}
	m, k := EstimateParameters(1000000, 0.05)
	fmt.Printf("m:%d, k:%d\n", m, k)
	redisBitSet := NewRedisBitSet(pool)
	redisBitSet.SetBatchCount(5000)

	bloom := New(m, k)
	bloom.AddBitSetProvider("default", redisBitSet)

	bloom.SetHashFunc(hash)

	count := 3000
	base := 10000000
	testStrings := [][]byte{}
	for i := base; i < base+count; i++ {
		testStrings = append(testStrings, []byte(strconv.Itoa(i)))
	}
	err := bloom.Add("mykey", testStrings)
	if err != nil {
		t.Error(err)
	}

	start := time.Now().UnixNano()
	result, err := bloom.Exists("mykey", testStrings)
	fmt.Println(result, err)

	testStrings = testStrings[:0]
	fmt.Println(count)
	for i := base * 2; i < base*2+count; i++ {
		testStrings = append(testStrings, []byte(strconv.Itoa(i)))
	}

	start1 := time.Now().UnixNano()
	result, err = bloom.Exists("mykey", testStrings)
	fmt.Println("time1", (time.Now().UnixNano()-start1)/1e6)
	fmt.Println(result, err)
	fmt.Println("time", (time.Now().UnixNano()-start)/1e6)
}
func redisPool(db int) *redis.Pool {
	pool := &redis.Pool{
		MaxIdle:     30,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				return conn, err
			}

			_, err = conn.Do("SELECT", db)
			return conn, err
		},
	}

	return pool
}
func TestRotationDB(t *testing.T) {
	redis1 := redisPool(1)
	redis2 := redisPool(2)
	m, k := EstimateParameters(1000000, 0.05)
	fmt.Printf("m:%d, k:%d\n", m, k)
	redisBitSet1 := NewRedisBitSet(redis1)
	redisBitSet1.SetBatchCount(5000)

	redisBitSet2 := NewRedisBitSet(redis2)
	redisBitSet2.SetBatchCount(5000)

	bloom := New(m, k)
	bloom.AddBitSetProvider("db1", redisBitSet1)
	bloom.AddBitSetProvider("db2", redisBitSet2)

	bloom.SetHashFunc(hash)
	bloom.SetBloomMetaStore(&ReidsBloomMetaStore{pool: redisPool(0)})

	bloom.StartRotation("db1", []string{"db1", "db2"}, 80, false)
	count := 3000
	base := 10000000
	testStrings := [][]byte{}
	for i := base; i < base+count; i++ {
		testStrings = append(testStrings, []byte(strconv.Itoa(i)))
	}
	err := bloom.Add("mykey", testStrings)
	if err != nil {
		t.Error(err)
	}

	for {
		time.Sleep(time.Second * 30)
		result, err := bloom.Exists("mykey", testStrings)
		fmt.Println(result, err)
	}
}

func redisReset(conn redis.Conn) {
	conn.Do("FLUSHDB")
	for i := 0; i < 200; i++ {
		args := []interface{}{}
		keyName := "user_suffix_" + strconv.Itoa(i)
		fmt.Println(keyName)
		args = append(args, keyName, 0.05, 1250000)
		conn.Do("bf.reserve", args...)
	}
}
func TestRedistBloomRotationDB(t *testing.T) {
	redis1 := redisPool(1)
	redis2 := redisPool(2)
	redis3 := redisPool(3)
	m, k := EstimateParameters(1000000, 0.05)
	fmt.Printf("m:%d, k:%d\n", m, k)
	redisBitSet1 := NewRedisBitSet(redis1)
	redisBitSet1.SetResetFunc(redisReset)

	redisBitSet2 := NewRedisBitSet(redis2)
	redisBitSet2.SetResetFunc(redisReset)

	redisBitSet3 := NewRedisBitSet(redis3)
	redisBitSet3.SetResetFunc(redisReset)

	bloom := NewRedisBloom()
	bloom.AddBitSetProvider("db1", redisBitSet1)
	bloom.AddBitSetProvider("db2", redisBitSet2)
	bloom.AddBitSetProvider("db3", redisBitSet3)

	bloom.SetBloomMetaStore(&ReidsBloomMetaStore{pool: redisPool(0)})

	bloom.StartRotation("db1", []string{"db1", "db2", "db3"}, 80, false)
	count := 3000
	base := 10000000
	testStrings := [][]byte{}
	for i := base; i < base+count; i++ {
		testStrings = append(testStrings, []byte(strconv.Itoa(i)))
	}
	err := bloom.Add("mykey", testStrings)
	if err != nil {
		t.Error(err)
	}

	for {
		time.Sleep(time.Second * 30)
		result, err := bloom.Exists("mykey", testStrings)
		fmt.Println(result, err)
	}
}
