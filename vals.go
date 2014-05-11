package tql
import (
    "errors"
    "math"
)

const (
	ValString = iota
	ValInt
	ValFloat
	ValQuoteString
	ValBool
	ValNull
	ValReference
	ValList
)

type Val struct {
	ValType int
	V     interface{}
}

var ErrTypeNotMatch = errors.New("type not match")

func (v Val) LT(v1 interface{}) (bool, error) {
    switch v.ValType {
        case ValInt:
            return v.V.(int64) < v1.(int64), nil
        case ValFloat:
            return v.V.(float64) < v1.(float64), nil
    }
    return false, ErrTypeNotMatch
}

func (v Val) GT(v1 interface{}) (bool, error) {
    switch v.ValType {
        case ValInt:
            return v.V.(int64) > v1.(int64), nil
        case ValFloat:
            return v.V.(float64) > v1.(float64), nil
    }
    return false, ErrTypeNotMatch
}

func (v Val) EQ(v1 interface{}) (bool, error) {
    switch v.ValType {
        case ValInt:
            return v.V.(int64) == v1.(int64), nil
        case ValFloat:
            return math.Abs(v.V.(float64) - v1.(float64)) < 1e-7, nil
        case ValQuoteString:
            return v.V.(string) == v1.(string), nil
    }
    return false, ErrTypeNotMatch
}

func (v Val) NOTEQ(v1 interface{}) (bool, error) {
    return true, nil
}

func (v Val) LTE(v1 interface{}) (bool, error) {
    b, err := v.GT(v1)
    return !b, err
}

func (v Val) GTE(v1 interface{}) (bool, error) {
    b, err := v.LT(v1)
    return !b, err
}

