package main

import (
	"fmt"
	"strings"
)

type ValueType int

const (
	TypeInt ValueType = iota
	TypeString
	TypeArray
)

type Value struct {
	Kind   ValueType
	IntVal int
	StrVal string
	ArrVal []Value
}

func StrVal(s string) Value { return Value{Kind: TypeString, StrVal: s} }

var Nil = Value{Kind: TypeInt}

func IntVal(n int) Value { return Value{Kind: TypeInt, IntVal: n} }

func ArrayVal(elems []Value) Value {
	cp := make([]Value, len(elems))
	copy(cp, elems)
	return Value{Kind: TypeArray, ArrVal: cp}
}

func (v Value) Format() string {
	switch v.Kind {
	case TypeInt:
		return fmt.Sprintf("%d", v.IntVal)
	case TypeString:
		return v.StrVal
	case TypeArray:
		parts := make([]string, len(v.ArrVal))
		for i, elem := range v.ArrVal {
			parts[i] = elem.Format()
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	return "<unknown>"
}

func toInt(v Value) int {
	if v.Kind == TypeInt {
		return v.IntVal
	}
	return 0
}

func deepCopyVars(src map[string]Value) map[string]Value {
	dst := make(map[string]Value, len(src))
	for k, v := range src {
		if v.Kind == TypeArray {
			dst[k] = ArrayVal(v.ArrVal)
		} else {
			dst[k] = v
		}
	}
	return dst
}
