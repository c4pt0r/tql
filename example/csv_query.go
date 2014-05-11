package main
/*
    Simple CSV Query Tool
    Usage:
    csv_query "<tql query string>"

    e.g.
    go run csv_query.go "select * from test.csv where field1 > 10 and field2 = 21"
*/

import (
	"encoding/csv"
	"github.com/c4pt0r/tql"
	"log"
	"os"
    "strconv"
)

func fileExists(fn string) bool {
	if _, err := os.Stat(fn); err == nil {
		return true
	}
	return false
}


func NewRow(fields []string, values []string) tql.Row {
    if len(fields) != len(values) {
        log.Fatal("invalid csv")
    }
    r := make(map[string]interface{})
    for i := 0 ; i < len(fields); i++ {
        s := values[i]
        n, err := strconv.ParseInt(s, 10, 64)
        if err == nil {
            r[fields[i]] = n
            continue
        }
        v, err := strconv.ParseFloat(s, 64)
        if err == nil {
            r[fields[i]] = v
            continue
        }
        r[fields[i]] = s
    }
    return r
}

func process(reader *csv.Reader) []tql.Row {
    var rows []tql.Row
    fields, err := reader.Read()
    if err != nil {
        log.Fatal(err)
    }
    for {
        values, err := reader.Read()
        if err != nil {
            break
        }
        r := NewRow(fields, values)
        rows = append(rows, r)
    }
	return rows
}

func doQuery(rows []tql.Row, t *tql.Tql) []tql.Row {
    var res []tql.Row
    for _, r := range rows {
        if b, err := t.Match(r); err != nil {
            log.Fatal(err)
        } else if b {
            res = append(res, r)
        }
    }
    return res
}

func main() {
	query := os.Args[1]
	log.Println("query:", query)

	t := tql.NewTql(query)
	if t == nil {
		log.Fatal("query invalid")
	}
	if !fileExists(t.From) {
		log.Fatal("file not exists")
	}

	if fp, err := os.Open(t.From); err == nil {
		csvReader := csv.NewReader(fp)
		if csvReader != nil {
            rows := process(csvReader)
            res := doQuery(rows, t)
            log.Println(res)
		}
	}

}
