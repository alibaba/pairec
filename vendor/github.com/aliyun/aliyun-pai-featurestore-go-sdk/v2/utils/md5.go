package utils

import (
	"crypto/md5"
	"encoding/hex"
	"hash/fnv"
	"io"
)

func Md5(msg string) string {
	h := md5.New()
	io.WriteString(h, msg)
	return hex.EncodeToString(h.Sum(nil))
}

func HashValue(hashKey string) uint64 {
	md5 := md5.Sum([]byte(hashKey))
	hash := fnv.New64()
	hash.Write(md5[:])

	return hash.Sum64()
}
