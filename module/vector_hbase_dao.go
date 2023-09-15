package module

import (
	"fmt"
	"strings"

	"github.com/alibaba/pairec/datasource/hbase"
	"github.com/alibaba/pairec/recconf"
)

// VectorHBaseDao is hbase implement of vector dao
type VectorHBaseDao struct {
	client       *hbase.HBase
	prefix       string
	table        string
	columnFamily string
	qualifier    string
	// defaultKey string
}
// NewVectorHBaseDao create new VectorHBaseDao 
func NewVectorHBaseDao(config recconf.RecallConfig) *VectorHBaseDao {
	dao := &VectorHBaseDao{
		prefix:       config.DaoConf.HBasePrefix,
		table:        config.DaoConf.HBaseTable,
		columnFamily: config.DaoConf.ColumnFamily,
		qualifier:    config.DaoConf.Qualifier,
		// defaultKey: config.DaoConf.RedisDefaultKey,
	}
	client, err := hbase.GetHBase(config.DaoConf.HBaseName)
	if err != nil {
		panic(err)
	}
	dao.client = client
	return dao
}
// VectorString get data from hbase by id key 
// returns vector of string 
func (d *VectorHBaseDao) VectorString(id string) (string, error) {
	// key := fmt.Sprintf("UI2V_%s", user.Id)
	key := d.prefix + string(id)
	resp, err := d.client.Get(d.table, key, d.columnFamily, d.qualifier)
	if err != nil {
		return "", err
	}

	if len(resp.Cells) == 0 {
		return "", VectoryEmptyError
	}
	value := string(resp.Cells[0].Value)
	value = strings.Trim(value, "[]")

	values := strings.Split(value, ",")
	var list []string

	i := 1
	// same as redis format 1:xx 2:xx 3:xx
	// same as libsvm format
	for _, val := range values {
		list = append(list, fmt.Sprintf("%d:%s", i, strings.Trim(val, " ")))
		i++
	}
	return strings.Join(list, " "), nil
}
