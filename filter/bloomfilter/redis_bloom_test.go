package bloomfilter

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
)

func TestRedisBloom(t *testing.T) {
	pool := &redis.Pool{
		MaxIdle:     30,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", "127.0.0.1:6379") },
	}
	m, k := EstimateParameters(1000000, 0.05)
	fmt.Printf("m:%d, k:%d\n", m, k)
	redisBitSet := NewRedisBitSet(pool)
	// redisBitSet.SetBatchCount(2000)

	bloom := NewRedisBloom()
	bloom.AddBitSetProvider("default", redisBitSet)

	count := 3000
	base := 10000000
	testStrings := [][]byte{}
	for i := base; i < base+count; i++ {
		testStrings = append(testStrings, []byte(strconv.Itoa(i)))
	}
	err := bloom.Add("mykey1", testStrings)
	if err != nil {
		t.Error(err)
	}

	start := time.Now().UnixNano()
	result, err := bloom.Exists("mykey1", testStrings)
	fmt.Println(result, err)

	testStrings = testStrings[:0]
	fmt.Println(count)
	for i := base * 2; i < base*2+count; i++ {
		testStrings = append(testStrings, []byte(strconv.Itoa(i)))
	}

	start1 := time.Now().UnixNano()
	result, err = bloom.Exists("mykey1", testStrings)
	fmt.Println("time1", (time.Now().UnixNano()-start1)/1e6)
	fmt.Println(result, err)
	fmt.Println("time", (time.Now().UnixNano()-start)/1e6)
}
