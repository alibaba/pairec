package ast

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// 字面量，e.g. 50
	Literal = iota
	// 操作符, e.g. + - * /
	Operator

	// 参数变量
	Parameter
)

type Token struct {
	// 原始字符
	Tok string
	// 类型，有 Literal、Operator 两种
	Type int

	Offset int
}

// 定义一个结构体，描述一个词法分析器
type Parser struct {
	// 输入的字符串
	Source string
	// 扫描器当前所在的字符
	ch byte
	// 扫描器当前所在的位置
	offset int
	// 扫描过程出现的错误收集
	err error
}

// 逐个字符扫描，得到一串 Token 序列
func (p *Parser) parse() []*Token {
	toks := make([]*Token, 0)
	// 一直获取下一个 Token
	for {
		tok := p.nextTok()
		if tok == nil {
			// 已经到达末尾或者出现错误时，停止
			break
		}
		// 收集 Token
		toks = append(toks, tok)
	}
	return toks
}

// 获取下一个 Token
func (p *Parser) nextTok() *Token {
	// 已经到达末尾或者出现错误
	if p.offset >= len(p.Source) || p.err != nil {
		return nil
	}
	var err error
	// 跳过所有无意义的空白符
	for p.isWhitespace(p.ch) && err == nil {
		err = p.nextCh()
	}
	start := p.offset
	var tok *Token
	switch p.ch {
	// 操作符
	case
		'#',
		'(',
		')',
		'+',
		'-',
		'*',
		'/',
		'^',
		'%':
		tok = &Token{
			Tok:  string(p.ch),
			Type: Operator,
		}
		tok.Offset = start
		// 前进到下一个字符
		err = p.nextCh()

		// 字面量(数字)
	case
		'0',
		'1',
		'2',
		'3',
		'4',
		'5',
		'6',
		'7',
		'8',
		'9':
		for p.isDigitNum(p.ch) && p.nextCh() == nil {
		}
		tok = &Token{
			Tok:  strings.ReplaceAll(p.Source[start:p.offset], "_", ""),
			Type: Literal,
		}
		tok.Offset = start

	case '$':
		if p.nextCh() == nil && p.ch == '{' {
			for p.ch != '}' && p.nextCh() == nil {
			}
			tok = &Token{
				Tok:  string(p.Source[start+2 : p.offset]),
				Type: Parameter,
			}
			tok.Offset = start + 2
			// get '}' next char
			p.nextCh()
		}
		// 捕获错误
	default:
		if p.ch != ' ' {
			s := fmt.Sprintf("symbol error: unkown '%v', pos [%v:]\n%s",
				string(p.ch),
				start,
				ErrPos(p.Source, start))
			p.err = errors.New(s)
		}
	}
	return tok
}

// 前进到下一个字符
func (p *Parser) nextCh() error {
	p.offset++
	if p.offset < len(p.Source) {
		p.ch = p.Source[p.offset]
		return nil
	}
	// 到达字符串末尾
	return errors.New("EOF")
}

// 空白符
func (p *Parser) isWhitespace(c byte) bool {
	return c == ' ' ||
		c == '\t' ||
		c == '\n' ||
		c == '\v' ||
		c == '\f' ||
		c == '\r'
}

// 数字
func (p *Parser) isDigitNum(c byte) bool {
	return '0' <= c && c <= '9' || c == '.' || c == '_' || c == 'e'
}

// 对错误包装，进行可视化展示
func ErrPos(s string, pos int) string {
	r := strings.Repeat("-", len(s)) + "\n"
	s += "\n"
	for i := 0; i < pos; i++ {
		s += " "
	}
	s += "^\n"
	return r + s + r
}

// 封装词法分析过程，直接调用该函数即可解析字符串为[]Token
func Parse(s string) ([]*Token, error) {
	// 初始化 Parser
	p := &Parser{
		Source: s,
		err:    nil,
		ch:     s[0],
	}
	// 调用 parse 方法
	toks := p.parse()
	if p.err != nil {
		return nil, p.err
	}
	return toks, nil
}
