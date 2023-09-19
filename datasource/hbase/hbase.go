package hbase

import (
	"context"
	"fmt"

	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/hrpc"
	"github.com/alibaba/pairec/v2/recconf"
)

type HBase struct {
	Client   gohbase.Client
	ZKQuorum string
	Timeout  int
}

var hbaseInstances = make(map[string]*HBase)

func GetHBase(name string) (*HBase, error) {
	if _, ok := hbaseInstances[name]; !ok {
		return nil, fmt.Errorf("Hbase not found, name:%s", name)
	}

	return hbaseInstances[name], nil
}
func NewHBase(zkquorum string, timeout int) *HBase {
	h := &HBase{
		ZKQuorum: zkquorum,
		Timeout:  timeout,
	}
	return h
}
func (h *HBase) Init() error {
	client := gohbase.NewClient(h.ZKQuorum)
	h.Client = client

	return nil
}

func (h *HBase) Insert(table, key, columnFamily, qualifier string, value []byte) (*hrpc.Result, error) {
	values := map[string]map[string][]byte{columnFamily: map[string][]byte{qualifier: value}}
	putRequest, err := hrpc.NewPutStr(context.Background(), table, key, values)
	if err != nil {
		return nil, err
	}
	rsp, err := h.Client.Put(putRequest)

	if err != nil {
		return nil, err
	}

	return rsp, nil

}
func (h *HBase) Get(table, key, columnFamily, qualifier string) (*hrpc.Result, error) {
	family := map[string][]string{columnFamily: []string{qualifier}}
	getRequest, err := hrpc.NewGetStr(context.Background(), table, key,
		hrpc.Families(family))
	if err != nil {
		return nil, err
	}
	getRsp, err := h.Client.Get(getRequest)
	if err != nil {
		return nil, err
	}

	return getRsp, nil
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.HBaseConfs {
		if _, ok := hbaseInstances[name]; ok {
			continue
		}
		d := NewHBase(conf.ZKQuorum, 100)
		err := d.Init()
		if err != nil {
			panic(err)
		}
		hbaseInstances[name] = d
	}
}
