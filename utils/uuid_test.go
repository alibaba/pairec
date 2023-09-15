package utils

import (
	"crypto/md5"
	"fmt"
	"testing"
)

func TestUUID(t *testing.T) {
	uuid := UUID()

	t.Logf("uuid:%s", uuid)
}
func TestMd5Val(t *testing.T) {
	id := "A0000059DAC1F8"

	md5 := md5.Sum([]byte(id))
	fmt.Println(md5)
	total := 0
	for _, c := range md5 {
		i := int8(c)
		if i < 0 {
			total += int(-1 * i)
		} else {
			total += int(i)
		}
	}
	prefix := 1000 + total%50

	fmt.Println(prefix)
}
