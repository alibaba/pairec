package bloomfilter

import (
	"fmt"
)

type redisBitSetProvider interface {
	// batch set or test exist
	redisBloomSet(string, [][]byte) error
	redisBloomTest(string, [][]byte) ([]bool, error)
	Clear()
	Online(bool)
}

type RedisBloomFilter struct {
	BloomRotation
	m         uint
	k         uint
	bitSetMap map[string]redisBitSetProvider
}

func NewRedisBloom() *RedisBloomFilter {
	f := &RedisBloomFilter{bitSetMap: make(map[string]redisBitSetProvider)}
	f.BloomRotation.bloom = f
	return f
}

func (f *RedisBloomFilter) AddBitSetProvider(name string, bitSet redisBitSetProvider) {
	f.bitSetMap[name] = bitSet
}

func (f *RedisBloomFilter) Add(key string, data [][]byte) error {
	var err error
	for _, bitSet := range f.bitSetMap {
		err1 := bitSet.redisBloomSet(key, data)
		if err1 != nil {
			err = err1
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func (f *RedisBloomFilter) Exists(key string, data [][]byte) ([]bool, error) {
	if len(f.bitSetMap) == 1 {
		for _, bitSet := range f.bitSetMap {
			return bitSet.redisBloomTest(key, data)
		}
	}

	name := f.metaInfo.currActiveDbName
	if _, exist := f.bitSetMap[name]; !exist {
		return nil, fmt.Errorf("not found BitSetProvider, bitsetname:%s", name)
	}

	return f.bitSetMap[name].redisBloomTest(key, data)
}

func (f *RedisBloomFilter) BitSetClear(name string) {
	if _, ok := f.bitSetMap[name]; ok {
		go func() {
			f.bitSetMap[name].Clear()
		}()
	}
}
func (f *RedisBloomFilter) BitSetOnline(name string, online bool) {
	if _, ok := f.bitSetMap[name]; ok {
		f.bitSetMap[name].Online(online)
	}
}
