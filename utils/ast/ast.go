package ast

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"github.com/alibaba/pairec/v2/log"
	valuate "github.com/bruceding/go-antlr-valuate"
)

// 基础表达式节点接口
type ExprAST interface {
	toStr() string
}

// 数字表达式节点
type NumberExprAST struct {
	// 具体的值
	Val float64
}

type ParameterExprAST struct {
	Val string
}

// 操作表达式节点
type BinaryExprAST struct {
	// 操作符
	Op string
	// 左右节点，可能是 数字表达式节点/操作表达式节点/nil
	Lhs,
	Rhs ExprAST
}

// 实现接口
func (n NumberExprAST) toStr() string {
	return fmt.Sprintf(
		"NumberExprAST:%s",
		strconv.FormatFloat(n.Val, 'f', 0, 64),
	)
}

// 实现接口
func (n ParameterExprAST) toStr() string {
	return fmt.Sprintf(
		"ParameterExprAST:%s",
		n.Val,
	)
}

// 实现接口
func (b BinaryExprAST) toStr() string {
	return fmt.Sprintf(
		"BinaryExprAST: (%s %s %s)",
		b.Op,
		b.Lhs.toStr(),
		b.Rhs.toStr(),
	)
}

// AST 生成器结构体
type AST struct {
	// 词法分析的结果
	Tokens []*Token
	// 源字符串
	source string
	// 当前分析器分析的 Token
	currTok *Token
	// 当前分析器的位置
	currIndex int
	// 错误收集
	Err error
}

// 定义操作符优先级，value 越高，优先级越高
var precedence = map[string]int{"+": 20, "-": 20, "*": 40, "/": 40, "%": 40, "^": 60, "#": 80}

// 语法分析器入口
func (a *AST) ParseExpression() ExprAST {
	lhs := a.parsePrimary()
	return a.parseBinOpRHS(0, lhs)
}

// 获取下一个 Token
func (a *AST) getNextToken() *Token {
	a.currIndex++
	if a.currIndex < len(a.Tokens) {
		a.currTok = a.Tokens[a.currIndex]
		return a.currTok
	}
	return nil
}

// 获取操作优先级
func (a *AST) getTokPrecedence() int {
	if p, ok := precedence[a.currTok.Tok]; ok {
		return p
	}
	return -1
}

// 解析数字，并生成一个 NumberExprAST 节点
func (a *AST) parseNumber() NumberExprAST {
	f64, err := strconv.ParseFloat(a.currTok.Tok, 64)
	if err != nil {
		a.Err = errors.New(
			fmt.Sprintf("%v\nwant '(' or '0-9' but get '%s'\n%s",
				err.Error(),
				a.currTok.Tok,
				ErrPos(a.source, a.currTok.Offset)))
		return NumberExprAST{}
	}
	n := NumberExprAST{
		Val: f64,
	}
	a.getNextToken()
	return n
}

// 解析参数
func (a *AST) parseParameter() ParameterExprAST {
	n := ParameterExprAST{
		Val: a.currTok.Tok,
	}
	a.getNextToken()
	return n
}

// 获取一个节点，返回 ExprAST
// 这里会处理所有可能出现的类型，并对相应类型做解析
func (a *AST) parsePrimary() ExprAST {
	switch a.currTok.Type {
	case Literal:
		return a.parseNumber()
	case Parameter:
		return a.parseParameter()
	case Operator:
		// 对 () 语法处理
		if a.currTok.Tok == "(" {
			a.getNextToken()
			e := a.ParseExpression()
			if e == nil {
				return nil
			}
			if a.currTok.Tok != ")" {
				a.Err = errors.New(
					fmt.Sprintf("want ')' but get %s\n%s",
						a.currTok.Tok,
						ErrPos(a.source, a.currTok.Offset)))
				return nil
			}
			a.getNextToken()
			return e
		} else {
			return a.parseNumber()
		}
	default:
		return nil
	}
}

// 循环获取操作符的优先级，将高优先级的递归成较深的节点
// 这是生成正确的 AST 结构最重要的一个算法，一定要仔细阅读、理解
func (a *AST) parseBinOpRHS(execPrec int, lhs ExprAST) ExprAST {
	for {
		tokPrec := a.getTokPrecedence()
		if tokPrec < execPrec {
			return lhs
		}
		binOp := a.currTok.Tok
		if a.getNextToken() == nil {
			return lhs
		}
		rhs := a.parsePrimary()
		if rhs == nil {
			return nil
		}
		nextPrec := a.getTokPrecedence()
		if tokPrec < nextPrec {
			// 递归，将当前优先级+1
			rhs = a.parseBinOpRHS(tokPrec+1, rhs)
			if rhs == nil {
				return nil
			}
		}
		lhs = BinaryExprAST{
			Op:  binOp,
			Lhs: lhs,
			Rhs: rhs,
		}
	}
}

// 生成一个 AST 结构指针
func NewAST(toks []*Token, s string) *AST {
	a := &AST{
		Tokens: toks,
		source: s,
	}
	if a.Tokens == nil || len(a.Tokens) == 0 {
		a.Err = errors.New("empty token")
	} else {
		a.currIndex = 0
		a.currTok = a.Tokens[0]
	}
	return a
}

/**
// 一个典型的后序遍历求解算法
func ExprASTResult(expr ExprAST, exprDatas ...ParameterExprData) float64 {
	// 左右值
	var l, r float64
	switch expr.(type) {
	// 传入的根节点是 BinaryExprAST
	case BinaryExprAST:
		ast := expr.(BinaryExprAST)
		// 递归左节点
		l = ExprASTResult(ast.Lhs, exprDatas...)
		// 递归右节点
		r = ExprASTResult(ast.Rhs, exprDatas...)
		// 现在 l,r 都有具体的值了，可以根据运算符运算
		switch ast.Op {
		case "#":
			if l != 0.0 {
				return l
			} else {
				return r
			}
		case "^":
			return math.Pow(l, r)
		case "+":
			return l + r
		case "-":
			return l - r
		case "*":
			return l * r
		case "/":
			if r == 0 {
				panic(errors.New(
					fmt.Sprintf("violation of arithmetic specification: a division by zero in ExprASTResult: [%g/%g]",
						l,
						r)))
			}
			return l / r
		case "%":
			return float64(int(l) % int(r))
		default:

		}
	// 传入的根节点是 NumberExprAST,无需做任何事情，直接返回 Val 值
	case NumberExprAST:
		return expr.(NumberExprAST).Val
	case ParameterExprAST:
		val := expr.(ParameterExprAST).Val
		for _, data := range exprDatas {
			if f, err := data.FloatExprData(val); err == nil {
				return f
			}
		}
	}

	return 0.0
}
**/

// should use sync map
var caches = make(map[string]ExprAST)
var mutex sync.RWMutex

type exprAST struct {
	expression *valuate.EvaluableExpression
}

func (e *exprAST) Evaluate(data map[string]any) (result float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error(fmt.Sprintf("error=%v, stack=%s", err, strings.ReplaceAll(stack, "\n", "\t")))
			result = float64(0)
			err = nil
		}
	}()
	ret, err1 := e.expression.Evaluate(data)
	if err1 != nil {
		err = err1
		return
	}
	if r, ok := ret.(float64); ok {
		result = r
		return
	} else {
		result = float64(0)
		err = fmt.Errorf("expression invoke result:%v", ret)
		return
	}
}
func (e *exprAST) toStr() string {
	return ""
}

func GetExpAST(source string) (ExprAST, error) {
	if source == "" {
		return nil, nil
	}
	var exprAst ExprAST
	mutex.RLock()
	exprAst, ok := caches[source]
	mutex.RUnlock()
	if !ok {
		expression, err := valuate.NewEvaluableExpression(source)
		if err != nil {
			return nil, err
		}
		exprAst = &exprAST{
			expression: expression,
		}
		mutex.Lock()
		caches[source] = exprAst
		mutex.Unlock()
	}

	return exprAst, nil
}

func ExprASTResult(expr ExprAST, exprDatas ParameterExprData) float64 {
	switch ast := expr.(type) {
	// 传入的根节点是 BinaryExprAST
	case *exprAST:
		data := exprDatas.ExprData()
		result, err := ast.Evaluate(data)
		if err != nil {
			log.Error(fmt.Sprintf("expression invoke error:%v", err))
			return float64(0)
		}
		return result
	}

	return float64(0)
}
