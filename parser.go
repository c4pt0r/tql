package tql

/*
  Tql, Simple SQL-Like query language

  DNF:
  select <property> [, <property> ...] | *]
    [from <from>]
    [where <condition> [and <condition> ...]]
*/

import (
	"regexp"
	"strconv"
	"strings"
)

var tokenize_regex = `(?:'[^'\n\r]*')+|<=|>=|!=|=|<|>|,|\*|-?\d+(?:\.\d+)?|\w+(?:\.\w+)*|(?:"[^"\s]+")+|\(|\)|\S+`

type Tql struct {
	tokens []string
	query  string
	pos    int
	props  []string
	from   string
	conds  []Cond
}

const (
	ValString = iota
	ValInt
	ValFloat
	ValQuoteString
	ValBool
	ValNull
	ValReference
)

type Val struct {
	valType int
	val     interface{}
}

type Cond struct {
	identifier string
	op         string
	val        Val
}

func NewTql(query string) *Tql {
	t := new(Tql)
	t.pos = 0
	t.query = strings.ToLower(query)
	if re, err := regexp.Compile(tokenize_regex); err != nil {
		return nil
	} else {
		t.tokens = re.FindAllString(t.query, -1)
		if len(t.tokens) <= 0 {
			return nil
		}
	}
	t.__Select()
	return t
}

func (t *Tql) __Expect(expect string) {
	if t.__Consume(expect) == false {
		panic("token error")
	}
}

func (t *Tql) __Consume(expect string) bool {
	if t.pos < len(t.tokens) {
		if t.tokens[t.pos] == expect {
			t.pos += 1
			return true
		}
	}
	return false
}

func (t *Tql) __ConsumeRegexp(regex string) (bool, string) {
	if t.pos < len(t.tokens) {
		token := t.tokens[t.pos]
		re, _ := regexp.Compile(regex)
		if re.MatchString(token) {
			t.pos += 1
			return true, token
		}
	}
	return false, ""
}

// consume a identifier an return
var identifier_regex = `(\w+(?:\.\w+)*)$`

func (t *Tql) __Identifier() (bool, string) {
	if b, ident := t.__ConsumeRegexp(identifier_regex); b {
		return true, ident
	}
	return false, ""
}

func (t *Tql) __ExpectIdentifier() string {
	if b, ident := t.__Identifier(); b {
		return ident
	}
	panic("identifier error")
}

func (t *Tql) __Select() bool {
	t.__Expect("select")
	if !t.__Consume("*") {
		t.props = append(t.props, t.__ExpectIdentifier())
		for t.__Consume(",") {
			// TODO add prop
			t.props = append(t.props, t.__ExpectIdentifier())
		}
	} else {
		t.props = append(t.props, "*")
	}
	return t.__From()
}

func (t *Tql) __From() bool {
	if t.__Consume("from") {
		t.from = t.__ExpectIdentifier()
	} else {
		return false
	}
	return t.__Where()
}

func (t *Tql) __Where() bool {
	if t.__Consume("where") {
		return t.__ParseFilterList()
	}
	return false
}

var quoted_string_regex = `((?:\'[^\'\n\r]*\')+)`

func (t *Tql) __Value() (bool, Val) {
	if t.pos < len(t.tokens) {
		token := t.tokens[t.pos]
		// try int
		i, err := strconv.ParseInt(token, 10, 64)
		if err == nil {
			t.pos += 1
			return true, Val{ValInt, i}
		}
		// try float
		f, err := strconv.ParseFloat(token, 64)
		if err == nil {
			t.pos += 1
			return true, Val{ValFloat, f}
		}
		// try quote string
		b, val := t.__ConsumeRegexp(quoted_string_regex)
		if b {
			return true, Val{ValQuoteString, val}
		}
		// try bool
		b = t.__Consume("true")
		if b {
			return true, Val{ValBool, true}
		}
		b = t.__Consume("false")
		if b {
			return true, Val{ValBool, false}
		}
		// try null
		b = t.__Consume("null")
		if b {
			return true, Val{ValNull, nil}
		}
	}
	return false, Val{}
}

var condition_regex = `(<=|>=|!=|=|<|>)$`

func (t *Tql) __ParseFilterList() bool {
	b, ident := t.__Identifier()
	if !b {
		return b
	}
	b, op := t.__ConsumeRegexp(condition_regex)
	if !b {
		return b
	}
	b, val := t.__Value()
	if !b {
		return b
	}
	t.conds = append(t.conds, Cond{ident, op, val})
	if t.__Consume("and") {
		return t.__ParseFilterList()
	}
	return true
}
