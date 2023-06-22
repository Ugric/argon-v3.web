package main

import (
	"fmt"
	"os"
)

func makeGlobal() ArObject {
	var vars = anymap{}
	vars["global"] = vars
	vars["term"] = ArTerm
	vars["number"] = builtinFunc{"number", ArgonNumber}
	vars["string"] = builtinFunc{"string", ArgonString}
	vars["socket"] = builtinFunc{"boolean", ArSocket}
	vars["infinity"] = infinity
	vars["map"] = builtinFunc{"map", func(a ...any) (any, ArErr) {
		if len(a) == 0 {
			return Map(anymap{}), ArErr{}
		}
		switch x := a[0].(type) {
		case ArObject:
			if typeof(x) == "array" {
				newmap := anymap{}
				for i, v := range x.obj["__value__"].([]any) {
					v := ArValidToAny(v)
					switch y := v.(type) {
					case []any:
						if len(y) == 2 {
							if isUnhashable(y[0]) {
								return nil, ArErr{TYPE: "TypeError", message: "Cannot use unhashable value as key: " + typeof(y[0]), EXISTS: true}
							}
							key := ArValidToAny(y[0])
							newmap[key] = y[1]
							continue
						}
					}
					newmap[i] = v
				}
				return Map(newmap), ArErr{}
			} else if typeof(x) == "string" {
				newmap := anymap{}
				for i, v := range x.obj["__value__"].(string) {
					newmap[i] = ArString(string(v))
				}
				return Map(newmap), ArErr{}
			}
			return x, ArErr{}
		}
		return nil, ArErr{TYPE: "TypeError", message: "Cannot create map from '" + typeof(a[0]) + "'", EXISTS: true}
	}}
	vars["hex"] = builtinFunc{"hex", func(a ...any) (any, ArErr) {
		if len(a) != 1 {
			return nil, ArErr{TYPE: "TypeError", message: "expected 1 argument, got " + fmt.Sprint(len(a)), EXISTS: true}
		}
		a[0] = ArValidToAny(a[0])
		switch x := a[0].(type) {
		case number:
			if x.Denom().Cmp(one.Denom()) != 0 {
				return nil, ArErr{TYPE: "TypeError", message: "Cannot convert non-integer to hex", EXISTS: true}
			}
			n := x.Num().Int64()
			return ArString(fmt.Sprintf("%x", n)), ArErr{}
		}
		return nil, ArErr{TYPE: "TypeError", message: "Cannot convert '" + typeof(a[0]) + "' to hex", EXISTS: true}
	}}
	vars["throwError"] = builtinFunc{"throwError", ArThrowError}
	vars["array"] = builtinFunc{"array", func(a ...any) (any, ArErr) {
		if len(a) == 0 {
			return ArArray([]any{}), ArErr{}
		}
		switch x := a[0].(type) {
		case ArObject:
			if typeof(x) == "array" {
				return x, ArErr{}
			} else if typeof(x) == "string" {

				newarray := []any{}
				for _, v := range x.obj["__value__"].(string) {
					newarray = append(newarray, ArString(string(v)))
				}
				return ArArray(newarray), ArErr{}
			}
			newarray := []any{}
			for key, val := range x.obj {
				newarray = append(newarray, []any{key, val})
			}
			return ArArray(newarray), ArErr{}
		}
		return nil, ArErr{TYPE: "TypeError", message: "Cannot create array from '" + typeof(a[0]) + "'", EXISTS: true}
	}}
	vars["boolean"] = builtinFunc{"boolean", func(a ...any) (any, ArErr) {
		if len(a) == 0 {
			return false, ArErr{}
		}
		return anyToBool(a[0]), ArErr{}
	}}
	vars["time"] = ArTime
	vars["PI"] = PI
	vars["π"] = PI
	vars["e"] = e
	vars["ln"] = builtinFunc{"ln", ArgonLn}
	vars["log"] = builtinFunc{"log", ArgonLog}
	vars["logN"] = builtinFunc{"logN", ArgonLogN}
	vars["thread"] = builtinFunc{"thread", ArThread}
	vars["input"] = ArInput
	vars["round"] = builtinFunc{"round", func(a ...any) (any, ArErr) {
		if len(a) == 0 {
			return nil, ArErr{TYPE: "round", message: "round takes 1 argument",
				EXISTS: true}
		}
		precision := newNumber()
		if len(a) > 1 {
			switch x := a[1].(type) {
			case number:
				if !x.IsInt() {
					return nil, ArErr{TYPE: "TypeError", message: "Cannot round to '" + typeof(a[1]) + "'", EXISTS: true}
				}
				precision = x
			default:
				return nil, ArErr{TYPE: "TypeError", message: "Cannot round to '" + typeof(a[1]) + "'", EXISTS: true}
			}
		}

		switch x := a[0].(type) {
		case number:
			return round(newNumber().Set(x), int(precision.Num().Int64())), ArErr{}
		}
		return nil, ArErr{TYPE: "TypeError", message: "Cannot round '" + typeof(a[0]) + "'", EXISTS: true}
	}}
	vars["floor"] = builtinFunc{"floor", func(a ...any) (any, ArErr) {
		if len(a) == 0 {
			return nil, ArErr{TYPE: "floor", message: "floor takes 1 argument",
				EXISTS: true}
		}
		switch x := a[0].(type) {
		case number:
			return floor(x), ArErr{}
		}
		return nil, ArErr{TYPE: "TypeError", message: "Cannot floor '" + typeof(a[0]) + "'", EXISTS: true}
	}}
	vars["ceil"] = builtinFunc{"ceil", func(a ...any) (any, ArErr) {
		if len(a) == 0 {
			return nil, ArErr{TYPE: "ceil", message: "ceil takes 1 argument",
				EXISTS: true}
		}

		switch x := a[0].(type) {
		case number:
			return ceil(x), ArErr{}
		}
		return nil, ArErr{TYPE: "TypeError", message: "Cannot ceil '" + typeof(a[0]) + "'", EXISTS: true}
	}}
	vars["sqrt"] = builtinFunc{"sqrt", ArgonSqrt}
	vars["file"] = ArFile
	vars["random"] = ArRandom
	vars["json"] = ArJSON
	vars["sin"] = ArSin
	vars["arcsin"] = ArArcsin
	vars["cos"] = ArCos
	vars["arccos"] = ArArccos
	vars["tan"] = ArTan
	vars["arctan"] = ArArctan
	vars["cosec"] = ArCosec
	vars["arccosec"] = ArArccosec
	vars["sec"] = ArSec
	vars["arcsec"] = ArArcsec
	vars["cot"] = ArCot
	vars["arccot"] = ArArccot
	vars["todeg"] = ArToDeg
	vars["colour"] = ArColour
	vars["torad"] = ArToRad
	vars["abs"] = ArAbs
	vars["fraction"] = builtinFunc{"fraction", func(a ...any) (any, ArErr) {
		if len(a) == 0 {
			return nil, ArErr{TYPE: "fraction", message: "fraction takes 1 argument",
				EXISTS: true}
		}
		switch x := a[0].(type) {
		case number:
			return ArString(x.String()), ArErr{}
		case ArObject:
			if callable, ok := x.obj["__fraction__"]; ok {
				resp, err := runCall(
					call{
						callable: callable,
						args:     []any{},
					},
					stack{},
					0,
				)
				if err.EXISTS {
					return nil, err
				}
				return resp, ArErr{}
			}
		}
		return nil, ArErr{TYPE: "TypeError", message: "Cannot fraction '" + typeof(a[0]) + "'", EXISTS: true}
	}}
	vars["dir"] = builtinFunc{"dir", func(a ...any) (any, ArErr) {
		if len(a) == 0 {
			return ArArray([]any{}), ArErr{}
		}
		t := AnyToArValid(a[0])
		switch x := t.(type) {
		case ArObject:
			newarray := []any{}
			for key := range x.obj {
				newarray = append(newarray, key)
			}
			return ArArray(newarray), ArErr{}
		}
		return ArArray([]any{}), ArErr{}
	}}
	vars["subprocess"] = builtinFunc{"subprocess", ArSubprocess}
	vars["sequence"] = builtinFunc{"sequence", ArSequence}
	vars["exit"] = builtinFunc{"exit", func(a ...any) (any, ArErr) {
		if len(a) == 0 {
			os.Exit(0)
		}
		switch x := a[0].(type) {
		case number:
			os.Exit(int(floor(x).Num().Int64()))
		}
		os.Exit(0)
		return nil, ArErr{}
	}}
	vars["error"] = builtinFunc{"error", func(a ...any) (any, ArErr) {
		if len(a) < 1 || len(a) > 2 {
			return nil, ArErr{TYPE: "error", message: "error takes 1 or 2 arguments, got " + fmt.Sprint(len(a)), EXISTS: true}
		}
		if len(a) == 1 {
			a[0] = ArValidToAny(a[0])
			switch x := a[0].(type) {
			case string:
				return nil, ArErr{TYPE: "Error", message: x, EXISTS: true}
			}
		} else {
			a[0] = ArValidToAny(a[0])
			a[1] = ArValidToAny(a[1])
			switch x := a[0].(type) {
			case string:
				switch y := a[1].(type) {
				case string:
					return nil, ArErr{TYPE: x, message: y, EXISTS: true}
				}
			}
		}
		return nil, ArErr{TYPE: "TypeError", message: "Cannot create error from '" + typeof(a[0]) + "'", EXISTS: true}
	}}
	vars["chr"] = builtinFunc{"chr", func(a ...any) (any, ArErr) {
		if len(a) != 1 {
			return nil, ArErr{TYPE: "chr", message: "chr takes 1 argument, got " + fmt.Sprint(len(a)), EXISTS: true}
		}
		switch x := a[0].(type) {
		case number:
			return string([]rune{rune(floor(x).Num().Int64())}), ArErr{}
		}
		return nil, ArErr{TYPE: "TypeError", message: "Cannot convert '" + typeof(a[0]) + "' to string", EXISTS: true}
	}}
	vars["ord"] = builtinFunc{"ord", func(a ...any) (any, ArErr) {
		if len(a) != 1 {
			return nil, ArErr{TYPE: "ord", message: "ord takes 1 argument, got " + fmt.Sprint(len(a)), EXISTS: true}
		}
		a[0] = ArValidToAny(a[0])
		switch x := a[0].(type) {
		case string:
			if len(x) != 1 {
				return nil, ArErr{TYPE: "ord", message: "ord takes a string with only one character, got " + fmt.Sprint(len(a)), EXISTS: true}
			}
			return floor(newNumber().SetInt64(int64([]rune(x)[0]))), ArErr{}
		}
		return nil, ArErr{TYPE: "TypeError", message: "Cannot convert '" + typeof(a[0]) + "' to string", EXISTS: true}
	}}
	return Map(vars)
}
