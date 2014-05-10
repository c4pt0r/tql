package tql

import (
	"testing"
    "log"
)

func TestTokenize(t *testing.T) {
    tql := NewTql("select * from hello")
    if tql != nil {
        for _, token := range tql.tokens {
            log.Println(token)
        }
    }
}

func TestParseSelect(t *testing.T) {
    log.Println("parsing select")
    tql := NewTql("select a,b,c,d,e from hello")
    if tql != nil {
        for _, prop := range tql.props {
            log.Println(prop)
        }
    }
}


func TestParseWhere(t *testing.T) {
    log.Println("parsing select")
    tql := NewTql("select a,b,c,d,e from hello where a > 100 and b < 10 and c = 'Hello World'")
    if tql != nil {
        for _, cond := range tql.conds {
            log.Println(cond)
        }
    }
}
