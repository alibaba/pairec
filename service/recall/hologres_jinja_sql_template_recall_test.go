package recall

import (
	"reflect"
	"strings"
	"testing"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/nikolalohinski/gonja/v2"
)

type jinjaTestParam map[string]interface{}

func (p jinjaTestParam) GetParameter(name string) interface{} {
	return p[name]
}

func TestHologresJinjaSqlTemplateRecallRenderSql(t *testing.T) {
	sqlTemplate := `SELECT id AS doc_id, text_field, field1, field2
FROM documents
WHERE field1 > {{ bind(context.features.min_field1) }}::int
{% if context.features.query_text %}
  AND text_search(text_field, {{ bind(context.features.query_text) }}) > 0
{% endif %}
{% if context.features.field2 is not none %}
  AND field2 = {{ bind(context.features.field2) }}
{% endif %}
{% if context.features.query_vector %}
  AND approx_cosine_distance(vector1, {{ bind(context.features.query_vector) }}::float4[]) > 0
{% endif %}
{% if user.city %}
  AND city = {{ bind(user.city) }}
{% endif %}
ORDER BY id
LIMIT {{ bind(context.size) }}::int`
	template, err := gonja.FromString(sqlTemplate)
	if err != nil {
		t.Fatal(err)
	}
	recall := &HologresJinjaSqlTemplateRecall{template: template}

	testcases := []struct {
		name       string
		features   map[string]interface{}
		user       map[string]interface{}
		contains   []string
		notContain []string
		args       []any
	}{
		{
			name:     "all",
			features: map[string]interface{}{"min_field1": float64(2), "query_text": "test5", "field2": "a", "query_vector": []interface{}{2.8, 2.3, 2.4}},
			user:     map[string]interface{}{"city": "hz"},
			contains: []string{"field1 > $1::int", "text_search(text_field, $2)", "field2 = $3", "approx_cosine_distance(vector1, $4::float4[])", "city = $5", "LIMIT $6::int"},
			args:     []any{float64(2), "test5", "a", "{2.8,2.3,2.4}", "hz", 10},
		},
		{
			name:       "only min_field1",
			features:   map[string]interface{}{"min_field1": float64(6)},
			contains:   []string{"field1 > $1::int", "LIMIT $2::int"},
			notContain: []string{"text_search", "field2 =", "approx_cosine_distance", "city ="},
			args:       []any{float64(6), 10},
		},
		{
			name:       "empty query_text, null field2",
			features:   map[string]interface{}{"query_text": "", "field2": nil},
			contains:   []string{"field1 > $1::int", "LIMIT $2::int"},
			notContain: []string{"text_search", "field2 ="},
			args:       []any{nil, 10},
		},
		{
			name:       "no features",
			features:   nil,
			contains:   []string{"field1 > $1::int", "LIMIT $2::int"},
			notContain: []string{"text_search", "field2 ="},
			args:       []any{nil, 10},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			param := jinjaTestParam{"uid": "user_001", "scene": "home_feed", "category": "default"}
			if tc.features != nil {
				param["features"] = tc.features
			}
			ctx := context.NewRecommendContext()
			ctx.Param = param
			ctx.Size = 10
			user := module.NewUserWithContext("user_001", ctx)
			user.AddProperties(tc.user)

			query, args, err := recall.renderSql(user, ctx)
			if err != nil {
				t.Fatal(err)
			}
			for _, s := range tc.contains {
				if !strings.Contains(query, s) {
					t.Errorf("sql should contain %q, sql:\n%s", s, query)
				}
			}
			for _, s := range tc.notContain {
				if strings.Contains(query, s) {
					t.Errorf("sql should not contain %q, sql:\n%s", s, query)
				}
			}
			if !reflect.DeepEqual(args, tc.args) {
				t.Errorf("args = %#v, want %#v", args, tc.args)
			}
		})
	}
}

func TestHologresJinjaSqlTemplateRecallRenderLoopAndError(t *testing.T) {
	ctx := context.NewRecommendContext()
	ctx.Param = jinjaTestParam{"features": map[string]interface{}{"cities": []interface{}{"hz", "sh"}, "empty": []interface{}{}}}
	user := module.NewUserWithContext("user_001", ctx)

	template, err := gonja.FromString(`city IN ({% for c in context.features.cities %}{{ bind(c) }}{% if not loop.last %},{% endif %}{% endfor %}){% if context.features.empty %} AND x{% endif %}`)
	if err != nil {
		t.Fatal(err)
	}
	query, args, err := (&HologresJinjaSqlTemplateRecall{template: template}).renderSql(user, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if query != "city IN ($1,$2)" || !reflect.DeepEqual(args, []any{"hz", "sh"}) {
		t.Errorf("query = %q, args = %#v", query, args)
	}

	template, err = gonja.FromString(`{{ bind(context.uid, context.scene) }}`)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := (&HologresJinjaSqlTemplateRecall{template: template}).renderSql(user, ctx); err == nil {
		t.Error("bind() with two arguments should return error")
	}

	if _, err := gonja.FromString(`{% if x %}`); err == nil {
		t.Error("template without endif should fail to parse")
	}
}

func TestToSqlArg(t *testing.T) {
	testcases := []struct {
		value any
		want  any
	}{
		{nil, nil},
		{"a", "a"},
		{float64(1.5), float64(1.5)},
		{[]interface{}{1, 2, 3}, "{1,2,3}"},
		{[]interface{}{"a", `b"c`, nil}, `{"a","b\"c",NULL}`},
		{[]interface{}{[]interface{}{1, 2}, []interface{}{3, 4}}, "{{1,2},{3,4}}"},
		{map[string]interface{}{"k": 1}, `{"k":1}`},
	}
	for _, tc := range testcases {
		got, err := toSqlArg(tc.value)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("toSqlArg(%#v) = %#v, want %#v", tc.value, got, tc.want)
		}
	}
}
