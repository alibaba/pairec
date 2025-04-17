package module

import (
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/datasource/featuredb/fdbserverpb"
)

type User2ItemCustomFilterFeatureStoreDao struct {
	fsClient *fs.FSClient
	table    string
	//cache    cache.Cache
}

func NewUser2ItemCustomFilterFeatureStoreDao(config recconf.FilterConfig) *User2ItemCustomFilterFeatureStoreDao {
	fsclient, err := fs.GetFeatureStoreClient(config.DaoConf.FeatureStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao := &User2ItemCustomFilterFeatureStoreDao{
		fsClient: fsclient,
		table:    config.DaoConf.FeatureStoreViewName,
	}
	/**
	if config.ItemStateCacheSize > 0 {
		cacheTime := 3600
		if config.ItemStateCacheTime > 0 {
			cacheTime = config.ItemStateCacheTime
		}
		dao.cache = cache.New(cache.WithMaximumSize(config.ItemStateCacheSize),
			cache.WithExpireAfterAccess(time.Second*time.Duration(cacheTime)))
	}
			**/
	return dao
}

func (d *User2ItemCustomFilterFeatureStoreDao) Filter(uid UID, items []*Item, ctx *context.RecommendContext) (ret []*Item) {
	project := d.fsClient.GetProject()
	featureView := project.GetFeatureView(d.table)
	if featureView == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemCustomFilterFeatureStoreDao\terror=table not found, name:%s", ctx.RecommendId, d.table))
		ret = items
		return
	}
	request := new(fdbserverpb.TestBloomItemsRequest)

	request.Key = string(uid)
	for _, item := range items {
		request.Items = append(request.Items, string(item.Id))
	}

	tests, err := fdbserverpb.TestBloomItems(project, featureView, request)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemCustomFilterFeatureStoreDao\terr=%v", ctx.RecommendId, err))
		ret = items
		return
	}

	ret = make([]*Item, 0, len(items))
	for i, test := range tests {
		if !test {
			ret = append(ret, items[i])
		}
	}

	return
}
