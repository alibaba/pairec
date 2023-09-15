package types

type Index uint64

func (i Index) Uint64() uint64 {
	return uint64(i)
}

func FromUint64(i uint64) Index {
	return Index(i)
}

func FromUint64Slice(ii []uint64) []Index {
	ret := make([]Index, 0, len(ii))
	for _, i := range ii {
		ret = append(ret, FromUint64(i))
	}
	return ret
}
