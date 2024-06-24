package ast

import (
	"math"
	"reflect"
	"testing"

	"github.com/alibaba/pairec/v2/module"
	valuate "github.com/bruceding/go-antlr-valuate"
)

func TestAST(t *testing.T) {
	expression := "${ctr} + ${click} + ${price}"

	ast, err := GetExpAST(expression)
	if err != nil {
		t.Fatal(err)
	}

	item := module.NewItem("item_1")
	item.AddAlgoScore("ctr", 0.1)
	item.AddAlgoScore("click", 0.3)
	item.AddProperty("price", 0.1)
	result := ExprASTResult(ast, item)
	if result != 0.5 {
		t.Fatal("result not equal")
	}
}

func TestGoAntlrEvaluate(t *testing.T) {
	expression := "${ctr} + ${click} + ${price}"
	ast, err := valuate.NewEvaluableExpression(expression)
	if err != nil {
		t.Fatal(err)
	}

	item := module.NewItem("item_1")
	item.AddAlgoScore("ctr", 0.1)
	item.AddAlgoScore("click", 0.3)
	item.AddProperty("price", 0.1)

	m := item.GetAlgoScores()
	data := make(map[string]any)
	for k, v := range m {
		data[k] = v
	}
	data["price"] = 0.1
	result, err := ast.Evaluate(data)
	if err != nil {
		t.Fatal(err)
	}
	if r, ok := result.(float64); ok && r == 0.5 {
	} else {
		t.Fatal("result not equal")

	}
}

func TestGoAntlrEvaluate2(t *testing.T) {
	expression := "${ctr_1} + log(${click}) + exp(${price})"
	ast, err := valuate.NewEvaluableExpression(expression)
	if err != nil {
		t.Fatal(err)
	}

	item := module.NewItem("item_1")
	item.AddAlgoScore("ctr_1", 0.1)
	item.AddAlgoScore("click", 0.3)
	item.AddProperty("price", 0.1)

	m := item.GetAlgoScores()
	data := make(map[string]any)
	for k, v := range m {
		data[k] = v
	}
	data["price"] = 0.1
	result, err := ast.Evaluate(data)
	if err != nil {
		t.Fatal(err)
	}
	if r, ok := result.(float64); ok {
		t.Log("result:", r)
	} else {
		t.Fatal("result not equal")

	}
}

func TestGoAntlrEvaluate3(t *testing.T) {
	ppnet_probs_ctr := 0.11173942685127258
	ppnet_probs_cvr := 0.006906657014042139
	log_price := 1.0986122886681098
	ret := (ppnet_probs_ctr + 2*ppnet_probs_cvr) * (math.Pow(log_price, 0.1))
	t.Log(ret)
	expression := "(${ppnet_probs_ctr}+2*${ppnet_probs_cvr})*${log_price}^0.1"
	ast, err := valuate.NewEvaluableExpression(expression)
	if err != nil {
		t.Fatal(err)
	}

	item := module.NewItem("item_1")
	item.AddAlgoScore("ppnet_probs_ctr", ppnet_probs_ctr)
	item.AddAlgoScore("ppnet_probs_cvr", ppnet_probs_cvr)
	item.AddProperty("log_price", log_price)

	data := item.ExprData()
	result, err := ast.Evaluate(data)
	if err != nil {
		t.Fatal(err)
	}
	if r, ok := result.(float64); ok && r == ret {
		t.Log("result:", r)
	} else {
		t.Fatal("result not equal")

	}

	expast, err := GetExpAST(expression)
	if err != nil {
		t.Fatal(err)
	}

	result2 := ExprASTResult(expast, item)
	t.Log("result2:", result2)
	if result2 != ret {
		t.Fatal("result not equal")
	}
}

func TestGoAntlrEvaluate4(t *testing.T) {
	cdn_probs_ctr := 0.004212516359984875
	cdn_probs_cvr := 0.0014530600747093558
	log_price := 0.6931471805599453
	ret := (cdn_probs_ctr + 2*cdn_probs_cvr) * (math.Pow(log_price, 0.1))
	expression := "(${cdn_probs_ctr}+2*${cdn_probs_cvr})*${log_price}^0.1"
	ast, err := valuate.NewEvaluableExpression(expression)
	if err != nil {
		t.Fatal(err)
	}

	item := module.NewItem("item_1")
	item.AddAlgoScore("cdn_probs_ctr", cdn_probs_ctr)
	item.AddAlgoScore("cdn_probs_cvr", cdn_probs_cvr)
	item.AddProperty("log_price", log_price)

	data := item.ExprData()
	result, err := ast.Evaluate(data)
	if err != nil {
		t.Fatal(err)
	}
	if r, ok := result.(float64); ok && r == ret {
	} else {
		t.Fatal("result not equal")

	}
	expast, err := GetExpAST(expression)
	if err != nil {
		t.Fatal(err)
	}

	result2 := ExprASTResult(expast, item)
	t.Log("result2:", result2)
	if result2 != ret {
		t.Fatal("result not equal")
	}
}

func BenchmarkGoAntlrEvaluate4(b *testing.B) {
	cdn_probs_ctr := 0.004212516359984875
	cdn_probs_cvr := 0.0014530600747093558
	log_price := 0.6931471805599453
	expression := "(${cdn_probs_ctr}+2*${cdn_probs_cvr})*${log_price}^0.1"
	ast, err := valuate.NewEvaluableExpression(expression)
	if err != nil {
		b.Fatal(err)
	}

	item := module.NewItem("item_1")
	item.AddAlgoScore("cdn_probs_ctr", cdn_probs_ctr)
	item.AddAlgoScore("cdn_probs_cvr", cdn_probs_cvr)
	item.AddProperty("log_price", log_price)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := item.ExprData()
		ast.Evaluate(data)

	}
}

func BenchmarkNoramlAST(b *testing.B) {
	cdn_probs_ctr := 0.004212516359984875
	cdn_probs_cvr := 0.0014530600747093558
	log_price := 0.6931471805599453
	expression := "(${cdn_probs_ctr}+2*${cdn_probs_cvr})*${log_price}^0.1"

	item := module.NewItem("item_1")
	item.AddAlgoScore("cdn_probs_ctr", cdn_probs_ctr)
	item.AddAlgoScore("cdn_probs_cvr", cdn_probs_cvr)
	item.AddProperty("log_price", log_price)
	expast, err := GetExpAST(expression)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExprASTResult(expast, item)
	}
}

func TestASTWithType(t *testing.T) {
	expression := "${ctr} + ${click} + ${price}"

	astType := ""
	ast, err := GetExpASTWithType(expression, astType)
	if err != nil {
		t.Fatal(err)
	}

	item := module.NewItem("item_1")
	item.AddAlgoScore("ctr", 0.1)
	item.AddAlgoScore("click", 0.3)
	item.AddProperty("price", 0.1)
	result := ExprASTResultWithType(ast, item, astType)
	if result != 0.5 {
		t.Fatal("result not equal")
	}
	astType = "antlr"
	ast, err = GetExpASTWithType(expression, astType)
	if err != nil {
		t.Fatal(err)
	}

	result = ExprASTResultWithType(ast, item, astType)
	if result != 0.5 {
		t.Fatalf("result not equal, result:%v\n", result)
	}
	cdn_probs_ctr := 0.004212516359984875
	cdn_probs_cvr := 0.0014530600747093558
	log_price := 0.6931471805599453
	item.AddAlgoScore("cdn_probs_ctr", cdn_probs_ctr)
	item.AddAlgoScore("cdn_probs_cvr", cdn_probs_cvr)
	item.AddProperty("log_price", log_price)

	astType = ""
	expression = "(${cdn_probs_ctr}+2*${cdn_probs_cvr})*${log_price}^0.1"
	ast, err = GetExpASTWithType(expression, astType)
	if err != nil {
		t.Fatal(err)
	}
	result1 := ExprASTResultWithType(ast, item, astType)

	astType = "antlr"
	ast, err = GetExpASTWithType(expression, astType)
	if err != nil {
		t.Fatal(err)
	}
	result2 := ExprASTResultWithType(ast, item, astType)

	if !reflect.DeepEqual(result1, result2) {
		t.Fatalf("result not equal, result1:%v, result2:%v\n", result1, result2)
	}
}
