package main

import (
	"fmt"
	"math"
	"strings"
)

type ValueType int

const (
	TypeInt ValueType = iota
	TypeString
	TypeBool
	TypeNil
	TypeFloat
	TypeArray
)

type Value struct {
	Kind     ValueType
	IntVal   int
	BoolVal  bool
	StrVal   string
	FloatVal float64
	ArrVal   []Value
}

func StrVal(s string) Value      { return Value{Kind: TypeString, StrVal: s} }
func BoolVal(b bool) Value       { return Value{Kind: TypeBool, BoolVal: b} }
func FloatValue(f float64) Value { return Value{Kind: TypeFloat, FloatVal: f} }

var Nil = Value{Kind: TypeNil}

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
	case TypeBool:
		if v.BoolVal {
			return "True"
		}
		return "False"
	case TypeNil:
		return "None"
	case TypeFloat:
		if math.IsInf(v.FloatVal, 1) {
			return "inf"
		}
		return fmt.Sprintf("%g", v.FloatVal)
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

func toBool(v Value) bool {
	switch v.Kind {
	case TypeInt:
		return v.IntVal != 0
	case TypeBool:
		return v.BoolVal
	case TypeNil:
		return false
	case TypeFloat:
		return v.FloatVal != 0
	case TypeString:
		return v.StrVal != ""
	case TypeArray:
		return len(v.ArrVal) > 0
	}
	return false
}

func boolToVal(b bool) Value { return BoolVal(b) }

func toFloat(v Value) float64 {
	switch v.Kind {
	case TypeFloat:
		return v.FloatVal
	case TypeInt:
		return float64(v.IntVal)
	case TypeBool:
		if v.BoolVal {
			return 1
		}
		return 0
	}
	return 0
}

func numericLess(a, b Value) bool {
	if a.Kind == TypeFloat || b.Kind == TypeFloat {
		return toFloat(a) < toFloat(b)
	}
	return toInt(a) < toInt(b)
}

func numericEqual(a, b Value) bool {
	if a.Kind == TypeNil && b.Kind == TypeNil {
		return true
	}
	if a.Kind == TypeNil || b.Kind == TypeNil {
		return false
	}
	if a.Kind == TypeFloat || b.Kind == TypeFloat {
		return toFloat(a) == toFloat(b)
	}
	if a.Kind == TypeBool && b.Kind == TypeBool {
		return a.BoolVal == b.BoolVal
	}
	return toInt(a) == toInt(b)
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
