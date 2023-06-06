package main

import (
	_ "embed"
	"fmt"
	"log"
	"strings"
)

//go:embed test.json
var testjson []byte

func main() {
	p := &parser{data: testjson}
	tokens, err := p.parse()
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range tokens {
		fmt.Println(t)
	}
}

type tokenType int

const (
	tokenUnknown tokenType = iota
	tokenChar
	tokenKey
	tokenValue
)

type valueType int

const (
	valueUnknown valueType = iota
	valueString
	valueFloat
)

type token struct {
	typ       tokenType
	char      byte
	value     []byte
	valueType valueType
}

func (t token) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	switch t.typ {
	case tokenUnknown:
		sb.WriteString("TOKEN_UNKOWN")
	case tokenChar:
		sb.WriteString(fmt.Sprintf("TOKEN_CHAR '%c'", t.char))
	case tokenKey:
		sb.WriteString("TOKEN_KEY")
	case tokenValue:
		sb.WriteString("TOKEN_VALUE")
	}
	switch t.valueType {
	case valueString, valueFloat:
		sb.WriteString(" " + string(t.value))
	}
	sb.WriteString("}")
	return sb.String()
}

type parser struct {
	data []byte
	cur  int
}

func (p *parser) parse() ([]token, error) {
	var tokens []token
	nextIsVal := false
	for {
		if p.cur >= len(p.data) {
			break
		}
		ch := p.eat()
		switch ch {
		case '[', ']', '{', '}', ',', ':':
			tokens = append(tokens, token{typ: tokenChar, char: ch})
			nextIsVal = false
		case '"':
			val := p.eatString()
			if !nextIsVal {
				tokens = append(tokens, token{typ: tokenKey, value: val, valueType: valueString})
				nextIsVal = true
			} else {
				tokens = append(tokens, token{typ: tokenValue, value: val, valueType: valueString})
				nextIsVal = false
			}
		case ' ', '\t', '\n':
			// expected spaces
		default:
			if ch >= '0' && ch <= '9' {
				val := append([]byte{ch}, p.eatFloat()...)
				tokens = append(tokens, token{typ: tokenValue, value: val, valueType: valueFloat})
				nextIsVal = false
			} else {
				fmt.Println(tokens)
				return nil, fmt.Errorf("unexpected character at pos %d: %c", p.cur-1, ch)
			}
		}

	}
	return tokens, nil
}

func (p *parser) eatString() []byte {
	start := p.cur
	for {
		if ch := p.eat(); ch != '"' {
			continue
		}
		return p.data[start : p.cur-1]
	}
	return nil
}

func (p *parser) eatFloat() []byte {
	start := p.cur
	for {
		ch := p.eat()
		if isNum(ch) || ch == '.' {
			continue
		}
		return p.data[start : p.cur-1]
	}
	return nil
}

func (p *parser) eat() byte {
	ch := p.data[p.cur]
	p.cur += 1
	return ch
}

func isNum(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
