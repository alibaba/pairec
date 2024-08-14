package netutil

import (
	"fmt"
	"testing"
	"time"
)

func TestResolve(t *testing.T) {
	h := newHostName("p.qpic.cn")
	h.lookupHost()
	fmt.Println(h)
	addr, err := h.getSingleAddress()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(addr)
	addr, _ = h.getSingleAddress()
	fmt.Println(addr)
}

func TestDSNResolve(t *testing.T) {
	ip, err := GetAddrByHost("p.qpic.cn")
	time.Sleep(time.Second * 2)
	ip, err = GetAddrByHost("p.qpic.cn")
	fmt.Println(ip, err)
}
