package ast

import (
	"math"
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
}