package berecall

import (
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
)

// user exposure history filter
type User2ItemExposureFilter struct {
	user2ItemExposureDao module.User2ItemExposureDao
}

func NewUser2ItemExposureFilter(config recconf.BeFilterConfig) *User2ItemExposureFilter {
	filter := User2ItemExposureFilter{
		user2ItemExposureDao: module.NewUser2ItemExposureDao(config.FilterConfig),
	}

	return &filter
}

func (f *User2ItemExposureFilter) BuildQueryParams(user *module.User, context *context.RecommendContext) (ret map[string]string) {
	ret = map[string]string{
		"exposure_list": "",
	}

	filterIds := f.user2ItemExposureDao.GetExposureItemIds(user, context)
	if filterIds != "" {
		ret["exposure_list"] = filterIds
	}

	return
}
