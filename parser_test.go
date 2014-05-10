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


func TestParse(t *testing.T) {
    log.Println("parsing select")
    tql := NewTql(`select a,b,c from hello
                    where a > 100 and b < 10 and c = "He""ll''o World" and d in (1,2,3,4,5)
                    order by d
                    limit 1 offset 10`)
    if tql != nil {
        for _, token := range tql.tokens {
            log.Println(token)
        }
        for _, cond := range tql.conds {
            log.Println(cond)
        }

        log.Println("limit", tql.limit)
        log.Println("offset", tql.offset)
        log.Println("order by", tql.orderBy , tql.order)
    }
}
