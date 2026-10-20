package recall

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

// HologresJinjaSqlTemplateRecall renders the configured SqlTemplate (jinja syntax, by gonja) for each request,
// then runs the rendered sql on hologres.
// Variables in the template:
//
//	context.uid / context.scene / context.category / context.size / context.features.xxx
//	user.xxx
//
// {{ bind(value) }} does not write the value into the sql, it appends the value to the query args
// and writes the placeholder $N, so the value is bound as a query parameter.
// The ItemId column is the item id, the score column (if selected) is the item score,
// all the selected columns are set as item properties.
type HologresJinjaSqlTemplateRecall struct {
	*BaseRecall
	db       *sql.DB
	ItemId   string
	template *exec.Template
}

func NewHologresJinjaSqlTemplateRecall(config recconf.RecallConfig) *HologresJinjaSqlTemplateRecall {
	hologres, err := holo.GetPostgres(config.HologresJinjaSqlTemplateRecallConf.HologresName)
	if err != nil {
		panic(err)
	}
	template, err := gonja.FromString(config.HologresJinjaSqlTemplateRecallConf.SqlTemplate)
	if err != nil {
		panic(fmt.Sprintf("recall(%s) parse SqlTemplate error: %v", config.Name, err))
	}
	recall := HologresJinjaSqlTemplateRecall{
		BaseRecall: NewBaseRecall(config),
		db:         hologres.DB,
		ItemId:     config.HologresJinjaSqlTemplateRecallConf.ItemId,
		template:   template,
	}

	return &recall
}

func (i *HologresJinjaSqlTemplateRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()
	query, args, err := i.renderSql(user, context)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=HologresJinjaSqlTemplateRecall\terr=%s", context.RecommendId, err.Error()))
		return
	}
	//log sql and args
	if context.Debug {
		log.Info(fmt.Sprintf("event=HologresJinjaSqlTemplateRecall\tsql=%s\targs=%v", query, args))
	}

	rows, err := i.db.Query(query, args...)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=HologresJinjaSqlTemplateRecall\tsql=%s\terr=%s", context.RecommendId, query, err.Error()))
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=HologresJinjaSqlTemplateRecall\terr=%s", context.RecommendId, err.Error()))
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
			log.Error(fmt.Sprintf("requestId=%s\tevent=HologresJinjaSqlTemplateRecall\terr=%s", context.RecommendId, err.Error()))
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
		log.Error(fmt.Sprintf("requestId=%s\tevent=HologresJinjaSqlTemplateRecall\terr=%s", context.RecommendId, err.Error()))
	}

	log.Info(fmt.Sprintf("requestId=%s\tmodule=HologresJinjaSqlTemplateRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, i.modelName, len(ret), utils.CostTime(start)))
	return
}

// renderSql renders the template of this request, returns the sql and the args bound by bind().
func (i *HologresJinjaSqlTemplateRecall) renderSql(user *module.User, context *context.RecommendContext) (string, []any, error) {
	var args []any
	bind := func(params *exec.VarArgs) (string, error) {
		if len(params.Args) != 1 || len(params.KwArgs) != 0 {
			return "", fmt.Errorf("bind() expects exactly one argument, got %d", len(params.Args)+len(params.KwArgs))
		}
		arg, err := toSqlArg(params.Args[0].Interface())
		if err != nil {
			return "", err
		}
		args = append(args, arg)
		return "$" + strconv.Itoa(len(args)), nil
	}
	// features is empty instead of none, so context.features.xxx is none when features is not set
	features, _ := context.GetParameter("features").(map[string]interface{})
	if features == nil {
		features = make(map[string]interface{})
	}

	data := map[string]any{
		"context": map[string]any{
			"uid":      context.GetParameter("uid"),
			"scene":    context.GetParameter("scene"),
			"category": context.GetParameter("category"),
			"size":     context.Size,
			"features": features,
		},
		"user": user.MakeUserFeatures2(),
		"bind": bind,
	}

	query, err := i.template.ExecuteToString(exec.NewContext(data))
	if err != nil {
		return "", nil, err
	}

	return query, args, nil
}

// toSqlArg converts the value to a query arg, lists are converted to the array literal like {1,2,3},
// dicts are converted to json string.
func toSqlArg(value any) (any, error) {
	if value == nil {
		return nil, nil
	}
	if _, ok := value.([]byte); ok {
		return value, nil
	}
	switch reflect.ValueOf(value).Kind() {
	case reflect.Slice, reflect.Array:
		return toArrayLiteral(reflect.ValueOf(value)), nil
	case reflect.Map, reflect.Struct:
		b, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		return string(b), nil
	default:
		return value, nil
	}
}

func toArrayLiteral(value reflect.Value) string {
	elements := make([]string, value.Len())
	for x := 0; x < value.Len(); x++ {
		element := value.Index(x)
		for element.Kind() == reflect.Interface && !element.IsNil() {
			element = element.Elem()
		}
		switch element.Kind() {
		case reflect.Interface, reflect.Pointer:
			if element.IsNil() {
				elements[x] = "NULL"
			} else {
				elements[x] = quoteArrayElement(fmt.Sprint(element.Interface()))
			}
		case reflect.Slice, reflect.Array:
			elements[x] = toArrayLiteral(element)
		case reflect.String:
			elements[x] = quoteArrayElement(element.String())
		default:
			elements[x] = fmt.Sprint(element.Interface())
		}
	}

	return "{" + strings.Join(elements, ",") + "}"
}

func quoteArrayElement(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}
