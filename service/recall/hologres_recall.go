package recall

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

// HologresRecall runs the configured Sql on hologres, Params are resolved the same way as
// OpenSearchRecall and bound to $1..$N as query parameters.
// The ItemId column is the item id, the score column (if selected) is the item score,
// all the selected columns are set as item properties.
type HologresRecall struct {
	*BaseRecall
	db     *sql.DB
	ItemId string
	Sql    string
	Params []string
}

func NewHologresRecall(config recconf.RecallConfig) *HologresRecall {
	hologres, err := holo.GetPostgres(config.HologresRecallConf.HologresName)
	if err != nil {
		panic(err)
	}
	recall := HologresRecall{
		BaseRecall: NewBaseRecall(config),
		db:         hologres.DB,
		ItemId:     config.HologresRecallConf.ItemId,
		Sql:        config.HologresRecallConf.Sql,
		Params:     config.HologresRecallConf.Params,
	}

	return &recall
}

func (i *HologresRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()
	args, err := i.getRequestParams(user, context)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=HologresRecall\terr=%s", context.RecommendId, err.Error()))
		return
	}
	//log sql and args
	if context.Debug {
		log.Info(fmt.Sprintf("event=HologresRecall\tsql=%s\targs=%v", i.Sql, args))
	}

	rows, err := i.db.Query(i.Sql, args...)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=HologresRecall\tsql=%s\terr=%s", context.RecommendId, i.Sql, err.Error()))
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=HologresRecall\terr=%s", context.RecommendId, err.Error()))
		return
	}
	values := make([]any, len(columns))
	dest := make([]any, len(columns))
	for x := range values {
		dest[x] = &values[x]
	}

	for rows.Next() {
		if i.recallCount > 0 && len(ret) >= i.recallCount {
			break
		}
		if err := rows.Scan(dest...); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=HologresRecall\terr=%s", context.RecommendId, err.Error()))
			continue
		}

		properties := make(map[string]interface{}, len(columns))
		for x, column := range columns {
			if b, ok := values[x].([]byte); ok {
				properties[column] = string(b)
			} else {
				properties[column] = values[x]
			}
		}
		if itemId := utils.ToString(properties[i.ItemId], ""); itemId != "" {
			item := module.NewItemWithProperty(itemId, properties)
			item.RetrieveId = i.modelName
			item.Score = utils.ToFloat(properties["score"], 0)

			ret = append(ret, item)
		}
	}
	if err := rows.Err(); err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=HologresRecall\terr=%s", context.RecommendId, err.Error()))
	}

	log.Info(fmt.Sprintf("requestId=%s\tmodule=HologresRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, i.modelName, len(ret), utils.CostTime(start)))
	return
}

// getRequestParams resolves Params like OpenSearchRecall, the x-th value is bound to $x+1.
func (i *HologresRecall) getRequestParams(user *module.User, context *context.RecommendContext) ([]any, error) {
	paramResults := make([]any, len(i.Params))
	for x, param := range i.Params {
		val, err := i.getParam(param, user, context)
		if err != nil {
			return paramResults, err
		}
		paramResults[x] = val
	}

	return paramResults, nil
}

// getParam resolves one param written as user.xxx, context.xxx or context.features.xxx.
func (i *HologresRecall) getParam(param string, user *module.User, context *context.RecommendContext) (string, error) {
	paramArr := strings.Split(param, ".")

	switch len(paramArr) {
	case 2:
		if paramArr[0] == "user" {
			return user.StringProperty(paramArr[1]), nil
		} else if paramArr[0] == "context" {
			return utils.ToString(context.GetParameter(paramArr[1]), ""), nil
		}
		return "", nil
	case 3: //get value from context features
		if paramArr[0] == "context" && paramArr[1] == "features" {
			features := context.GetParameter("features")
			if features == nil {
				return "", errors.New("context.feature is null")
			}
			featureMap := features.(map[string]interface{})
			return utils.ToString(featureMap[paramArr[2]], ""), nil
		}
		return "", fmt.Errorf("Params(%s) only support context.features.xxx", param)
	default:
		return "", fmt.Errorf("Params(%s) type is error, its type should be a.b or a.b.c", param)
	}
}
