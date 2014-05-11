package tql

/*
  Tql, Simple SQL-Like query language

  BNF:
  select <property> [, <property> ...] | *]
    [from <from>]
    [where <condition> [and <condition> ...]]
    [order by <property> [asc|desc]]
    [limit <num> offset <num> | limit <num>, <num>]

  <condition> := <property> {< | <= | > | >= | = | != | in} <value>

  Author: c4pt0r(dongxu)  <huang@wandoujia.com>
*/

import (
	"regexp"
	"strconv"
	"strings"
    "errors"
)

var tokenize_regex = `(?:"[^"\n\r]*")+|(?:'[^'\n\r]*')+|<=|>=|!=|=|<|>|,|\*|-?\d+(?:\.\d+)?|\w+(?:\.\w+)*|(?:"[^"\s]+")+|\(|\)|\S+`

type Tql struct {
	tokens  []string
	query   string
	pos     int
    // export
	Props   []string
	From    string
	Conds   []Cond
	Limit   int64
	Offset  int64
	OrderBy string
	Order   int
}
var ErrNotImplement = errors.New("feature not implement yet")

type Cond struct {
	Identifier string
	Op         string
	Value        Val
}

var ErrUnknown = errors.New("unknown error")
var ErrFieldNotExists = errors.New("field not exists")

type Row map[string]interface{}

func (c *Cond) Match(r map[string]interface{}) (bool, error) {
    key := c.Identifier
    if _, b := r[key]; !b {
        return false, ErrFieldNotExists
    }
    target := r[key]
    v := c.Value
    op := c.Op
    switch op {
        case ">":
            return v.LT(target)
        case "<":
            return v.GT(target)
        case ">=":
            return v.LTE(target)
        case "<=":
            return v.GTE(target)
        case "=":
            return v.EQ(target)
        case "!=":
            return v.NOTEQ(target)
        case "in":
            return false, ErrNotImplement
    }
    return false, ErrUnknown
}

func NewTql(query string) *Tql {
	t := new(Tql)
	t.pos = 0
	t.query = strings.ToLower(query)
	t.Limit = -1
	t.Offset = -1
	t.Order = -1
	if re, err := regexp.Compile(tokenize_regex); err != nil {
		return nil
	} else {
		t.tokens = re.FindAllString(t.query, -1)
		if len(t.tokens) <= 0 {
			return nil
		}
	}
	if t.__Select() {
	    return t
    }
    return nil
}

func (t *Tql) Match(v Row) (bool, error) {
    b := true
    for _, c := range t.Conds {
        r, err := c.Match(v)
        if err != nil {
            return false, err
        }
        b = b && r
    }
    return b, nil
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
		t.Props = append(t.Props, t.__ExpectIdentifier())
		for t.__Consume(",") {
			t.Props = append(t.Props, t.__ExpectIdentifier())
		}
	} else {
		t.Props = append(t.Props, "*")
	}
	return t.__From()
}

func (t *Tql) __From() bool {
	if t.__Consume("from") {
		t.From = t.__ExpectIdentifier()
	} else {
		return false
	}
	return t.__Where()
}

func (t *Tql) __Where() bool {
	if t.__Consume("where") {
		return t.__ParseFilterList()
	}
	return t.__orderBy()
}

var num_regex = `(\d+)$`

func (t *Tql) __Limit() bool {
	if t.__Consume("limit") {
		_, limit := t.__ConsumeRegexp(num_regex)
		n, err := strconv.ParseInt(limit, 10, 64)
		if err != nil {
			return false
		}
		if t.__Consume(",") {
			t.Offset = n
			_, limit := t.__ConsumeRegexp(num_regex)
			n, err = strconv.ParseInt(limit, 10, 64)
			if err != nil {
				return false
			}
		}
		t.Limit = n
	}
	return t.__Offset()
}

func (t *Tql) __Offset() bool {
	if t.__Consume("offset") {
		b, offset := t.__ConsumeRegexp(num_regex)
		if !b {
			return false
		}
		n, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			return false
		}
		t.Offset = n
		return true
	}
	return true
}

var quoted_string_regex = `((?:\'[^\'\n\r]*\')+)|((?:"[^"\n\r]*")+)`

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
			if t.tokens[t.pos-1][0] == '\'' {
				val = strings.Replace(val, "''", "'", -1)
			} else {
				val = strings.Replace(val, `""`, `"`, -1)
			}
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

func (t *Tql) __ValueList() (bool, Val) {
	var vals []Val
	t.__Expect("(")
	for {
		if b, val := t.__Value(); b {
			vals = append(vals, val)
		} else {
			return false, Val{}
		}
		if !t.__Consume(",") {
			break
		}
	}
	t.__Expect(")")
	return true, Val{ValList, vals}
}

func (t *Tql) __orderBy() bool {
	if t.__Consume("order") {
		if t.__Consume("by") {
			t.OrderBy = t.__ExpectIdentifier()
			if t.__Consume("asc") {
				t.Order = 1
			} else if t.__Consume("desc") {
				t.Order = -1
			}
		} else {
			panic("parsing order error")
		}
	}
	return t.__Limit()
}

var condition_regex = `(<=|>=|!=|=|<|>|in)$`

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
	if !b && op == "in" {
		b, val = t.__ValueList()
	}
	if !b {
		return b
	}
	t.Conds = append(t.Conds, Cond{ident, op, val})
	if t.__Consume("and") {
		return t.__ParseFilterList()
	}
	return t.__orderBy()
}
