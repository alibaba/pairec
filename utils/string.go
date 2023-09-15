package utils

import (
	"reflect"
	"unsafe"
)

// from fasthttp
func Byte2string(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func String2byte(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}
