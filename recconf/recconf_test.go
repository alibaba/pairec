package recconf

import (
	"fmt"
	"sync"
	"testing"
)

func TestCopyConfig(t *testing.T) {

	src := &RecommendConfig{
		RunMode: "dev",
		ListenConf: ListenConfig{
			HttpAddr: "",
			HttpPort: 80,
		},
		//		SortNames:   []string{"item_score"},
		//	    FilterNames: []string{"item_exposure_filter"},
	}

	dst := &RecommendConfig{
		RunMode: "dev",
		ListenConf: ListenConfig{
			HttpAddr: "",
			HttpPort: 80,
		},
	}

	fmt.Println(dst)
	CopyConfig(src, dst, func(name string) bool {
		return name == "SortNames"
	})
	fmt.Println(dst)
	CopyConfig(src, dst)
	fmt.Println(dst)
}

type MyConf struct {
	TestConf struct {
		K string
	}
}

func TestParseUserDefineConfs(t *testing.T) {
	Config = &RecommendConfig{
		UserDefineConfs: []byte("{\"k\": \"v\", \"TestConf\": {\"K\": \"V1\"}}"),
	}

	v, err := ParseUserDefineConfs[string](WithKey("k"))
	if err != nil {
		t.Error(err)
	} else if v != "v" {
		t.Errorf("expect v, got %s", v)
	}

	c, err := ParseUserDefineConfs[MyConf]()
	if err != nil {
		t.Error(err)
	} else if c.TestConf.K != "V1" {
		t.Errorf("expect V1, got %s", c.TestConf.K)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	ch := Subscribe()
	go func() {
		<-ch

		c, err = ParseUserDefineConfs[MyConf]()
		if err != nil {
			t.Error(err)
		} else if c.TestConf.K != "V2" {
			t.Errorf("expect V2, got %s", c.TestConf.K)
		}

		wg.Done()
	}()

	UpdateConf(&RecommendConfig{
		UserDefineConfs: []byte("{\"k\": \"v\", \"TestConf\": {\"K\": \"V2\"}}"),
	})

	wg.Wait()
}
