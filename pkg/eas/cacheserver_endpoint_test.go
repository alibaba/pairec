package eas

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestCacheServerEndpoint_Sync(t *testing.T) {
	domain := "http://pai-eas-internet.cn-shanghai.aliyuncs.com"
	serviceName := "network_test"
	endpoint := newCacheServerEndpoint(domain, serviceName)
	endpoint.Sync()
	if len(endpoint.endpoints) == 0 {
		t.Fatalf("cache server sync empty endpoint list")
	}
	for name, weight := range endpoint.endpoints {
		names := strings.Split(name, ":")
		if len(names) != 2 {
			t.Fatalf("bad format of host name")
		}
		port, err := strconv.ParseInt(names[1], 10, 32)
		if err != nil {
			t.Fatalf("bad format of host port")
		}
		if port < 50000 || port > 60000 {
			t.Fatalf("unexpected port")
		}
		if weight != 100 {
			t.Fatalf("unexpected weight")
		}
	}
	fmt.Printf("%v\n", endpoint.endpoints)
}
