package bloomfilter

import (
	"fmt"
	"math"
)

type BitSetProvider interface {
	// batch set or test exist
	Set(string, [][]uint) error
	Test(string, [][]uint) ([]bool, error)
	Clear()
	Online(bool)
}

// data hash iterator with num
type HashFunc func(data [][]byte, m uint, k uint) [][]uint

type BloomFilterInterface interface {
	Add(key string, data [][]byte) error
	Exists(key string, data [][]byte) ([]bool, error)
}

type BloomFilter struct {
	BloomRotation
	m         uint
	k         uint
	bitSetMap map[string]BitSetProvider
	hash      HashFunc
	// metaInfo  BloomMeta
	// metaStore BloomMetaStore
}

func New(m uint, k uint) *BloomFilter {
	f := &BloomFilter{m: m, k: k, bitSetMap: make(map[string]BitSetProvider)}
	f.BloomRotation.bloom = f
	return f
}

func EstimateParameters(n uint, p float64) (uint, uint) {
	m := math.Ceil(float64(n) * math.Log(p) / math.Log(1.0/math.Pow(2.0, math.Ln2)))
	k := math.Ln2*m/float64(n) + 0.5

	return uint(m), uint(k)
}

func (f *BloomFilter) BitSetClear(name string) {
	if _, ok := f.bitSetMap[name]; ok {
		go func() {
			f.bitSetMap[name].Clear()
		}()
	}
}
func (f *BloomFilter) BitSetOnline(name string, online bool) {
	if _, ok := f.bitSetMap[name]; ok {
		f.bitSetMap[name].Online(online)
	}
}

func (f *BloomFilter) AddBitSetProvider(name string, bitSet BitSetProvider) {
	f.bitSetMap[name] = bitSet
}
func (f *BloomFilter) SetHashFunc(hf HashFunc) {
	f.hash = hf
}

func (f *BloomFilter) Add(key string, data [][]byte) error {
	locations := f.hash(data, f.m, f.k)
	var err error
	for _, bitSet := range f.bitSetMap {
		err1 := bitSet.Set(key, locations)
		if err1 != nil {
			err = err1
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func (f *BloomFilter) Exists(key string, data [][]byte) ([]bool, error) {
	locations := f.hash(data, f.m, f.k)
	if len(f.bitSetMap) == 1 {
		for _, bitSet := range f.bitSetMap {
			return bitSet.Test(key, locations)
		}
	}

	name := f.metaInfo.currActiveDbName
	if _, exist := f.bitSetMap[name]; !exist {
		return nil, fmt.Errorf("not found BitSetProvider, bitsetname:%s", name)
	}

	return f.bitSetMap[name].Test(key, locations)
}
