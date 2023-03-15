package main

import (
	"fmt"
	"strings"
)

type ArMap = map[any]any
type ArArray = []any

var mapGetCompile = makeRegex(`(.|\n)+\.([a-zA-Z_]|(\p{L}\p{M}*))([a-zA-Z0-9_]|(\p{L}\p{M}*))*( *)`)
var indexGetCompile = makeRegex(`(.|\n)+\[(.|\n)+\]( *)`)

type ArMapGet struct {
	VAL   any
	args  ArArray
	index bool
	line  int
	code  string
	path  string
}

func mapGet(r ArMapGet, stack stack, stacklevel int) (any, ArErr) {
	resp, err := runVal(r.VAL, stack, stacklevel+1)
	if err.EXISTS {
		return nil, err
	}
	switch m := resp.(type) {
	case ArMap:
		if len(r.args) > 1 {
			return nil, ArErr{
				"IndexError",
				"index not found",
				r.line,
				r.path,
				r.code,
				true,
			}
		}
		key, err := runVal(r.args[0], stack, stacklevel+1)
		if err.EXISTS {
			return nil, err
		}
		if isUnhashable(key) {
			return nil, ArErr{
				"TypeError",
				"unhashable type: '" + typeof(key) + "'",
				r.line,
				r.path,
				r.code,
				true,
			}
		}
		if _, ok := m[key]; !ok {
			return nil, ArErr{
				"KeyError",
				"key '" + fmt.Sprint(key) + "' not found",
				r.line,
				r.path,
				r.code,
				true,
			}
		}
		return m[key], ArErr{}

	case ArArray:
		return getFromArArray(m, r, stack, stacklevel)
	case string:
		return getFromString(m, r, stack, stacklevel)
	}

	key, err := runVal(r.args[0], stack, stacklevel+1)
	if err.EXISTS {
		return nil, err
	}
	return nil, ArErr{
		"TypeError",
		"cannot read " + anyToArgon(key, true, true, 3, 0, false, 0) + " from type '" + typeof(resp) + "'",
		r.line,
		r.path,
		r.code,
		true,
	}
}

func classVal(r any) any {
	if _, ok := r.(ArMap); ok {
		if _, ok := r.(ArMap)["__value__"]; ok {
			return r.(ArMap)["__value__"]
		}
	}
	return r
}

func isMapGet(code UNPARSEcode) bool {
	return mapGetCompile.MatchString(code.code)
}

func mapGetParse(code UNPARSEcode, index int, codelines []UNPARSEcode) (ArMapGet, bool, ArErr, int) {
	trim := strings.TrimSpace(code.code)
	split := strings.Split(trim, ".")
	start := strings.Join(split[:len(split)-1], ".")
	key := split[len(split)-1]
	resp, worked, err, i := translateVal(UNPARSEcode{code: start, realcode: code.realcode, line: code.line, path: code.path}, index, codelines, 0)
	if !worked {
		return ArMapGet{}, false, err, i
	}
	return ArMapGet{resp, ArArray{key}, false, code.line, code.realcode, code.path}, true, ArErr{}, 1
}

func isIndexGet(code UNPARSEcode) bool {
	return indexGetCompile.MatchString(code.code)
}

func indexGetParse(code UNPARSEcode, index int, codelines []UNPARSEcode) (ArMapGet, bool, ArErr, int) {
	trim := strings.TrimSpace(code.code)
	trim = trim[:len(trim)-1]
	split := strings.Split(trim, "[")
	for i := 1; i < len(split); i++ {
		ti := strings.Join(split[:i], "[")
		innerbrackets := strings.Join(split[i:], "[")
		args, success, argserr := getValuesFromLetter(innerbrackets, ":", index, codelines, true)
		if !success {
			if i == len(split)-1 {
				return ArMapGet{}, false, argserr, 1
			}
			continue
		}
		fmt.Println(args)
		if len(args) > 3 {
			return ArMapGet{}, false, ArErr{
				"SyntaxError",
				"too many arguments for index get",
				code.line,
				code.path,
				code.realcode,
				true,
			}, 1
		}
		tival, worked, err, i := translateVal(UNPARSEcode{code: ti, realcode: code.realcode, line: code.line, path: code.path}, index, codelines, 0)
		if !worked {
			if i == len(split)-1 {
				return ArMapGet{}, false, err, i
			}
			continue
		}
		return ArMapGet{tival, args, true, code.line, code.realcode, code.path}, true, ArErr{}, 1
	}
	return ArMapGet{}, false, ArErr{
		"SyntaxError",
		"invalid index get",
		code.line,
		code.path,
		code.realcode,
		true,
	}, 1
}

func isUnhashable(val any) bool {
	keytype := typeof(val)
	return keytype == "array" || keytype == "map"
}

func getFromArArray(m []any, r ArMapGet, stack stack, stacklevel int) (ArArray, ArErr) {
	var (
		start int = 0
		end   any = nil
		step  int = 1
	)
	{
		startval, err := runVal(r.args[0], stack, stacklevel+1)
		if err.EXISTS {
			return nil, err
		}
		if startval == nil {
			start = 0
		} else if typeof(startval) != "number" && !startval.(number).IsInt() {
			return nil, ArErr{
				"TypeError",
				"slice index must be an integer",
				r.line,
				r.path,
				r.code,
				true,
			}
		} else {
			start = int(startval.(number).Num().Int64())
		}
	}
	if len(r.args) > 1 {
		endval, err := runVal(r.args[1], stack, stacklevel+1)
		if err.EXISTS {
			return nil, err
		}
		if endval == nil {
			end = len(m)
		} else if typeof(endval) != "number" && !endval.(number).IsInt() {
			return nil, ArErr{
				"TypeError",
				"slice ending index must be an integer",
				r.line,
				r.path,
				r.code,
				true,
			}
		} else {
			end = int(endval.(number).Num().Int64())
		}
	}
	if len(r.args) > 2 {
		stepval, err := runVal(r.args[2], stack, stacklevel+1)
		if err.EXISTS {
			return nil, err
		}
		if stepval == nil {
			step = 1
		} else if typeof(stepval) != "number" && !stepval.(number).IsInt() {
			return nil, ArErr{
				"TypeError",
				"slice step must be an integer",
				r.line,
				r.path,
				r.code,
				true,
			}
		} else {
			step = int(stepval.(number).Num().Int64())
		}
	}
	if start < 0 {
		start = len(m) + start
	}
	if _, ok := end.(int); ok && end.(int) < 0 {
		end = len(m) + end.(int)
	}

	fmt.Println(start, end, step)
	if end == nil {
		return ArArray{m[start]}, ArErr{}
	} else if step == 1 {
		return m[start:end.(int)], ArErr{}
	} else {
		output := ArArray{}
		if step > 0 {
			for i := start; i < end.(int); i += step {
				output = append(output, m[i])
			}
		} else {
			for i := end.(int) - 1; i >= start; i += step {
				output = append(output, m[i])
			}
		}
		return (output), ArErr{}
	}
}

func getFromString(m string, r ArMapGet, stack stack, stacklevel int) (string, ArErr) {
	var (
		start int = 0
		end   any = nil
		step  int = 1
	)
	{
		startval, err := runVal(r.args[0], stack, stacklevel+1)
		if err.EXISTS {
			return "", err
		}
		if startval == nil {
			start = 0
		} else if typeof(startval) != "number" && !startval.(number).IsInt() {
			return "", ArErr{
				"TypeError",
				"slice index must be an integer",
				r.line,
				r.path,
				r.code,
				true,
			}
		} else {
			start = int(startval.(number).Num().Int64())
		}
	}
	if len(r.args) > 1 {
		endval, err := runVal(r.args[1], stack, stacklevel+1)
		if err.EXISTS {
			return "", err
		}
		if endval == nil {
			end = len(m)
		} else if typeof(endval) != "number" && !endval.(number).IsInt() {
			return "", ArErr{
				"TypeError",
				"slice ending index must be an integer",
				r.line,
				r.path,
				r.code,
				true,
			}
		} else {
			end = int(endval.(number).Num().Int64())
		}
	}
	if len(r.args) > 2 {
		stepval, err := runVal(r.args[2], stack, stacklevel+1)
		if err.EXISTS {
			return "", err
		}
		if stepval == nil {
			step = 1
		} else if typeof(stepval) != "number" && !stepval.(number).IsInt() {
			return "", ArErr{
				"TypeError",
				"slice step must be an integer",
				r.line,
				r.path,
				r.code,
				true,
			}
		} else {
			step = int(stepval.(number).Num().Int64())
		}
	}
	if start < 0 {
		start = len(m) + start
	}
	if _, ok := end.(int); ok && end.(int) < 0 {
		end = len(m) + end.(int)
	}

	fmt.Println(start, end, step)
	if end == nil {
		return string(m[start]), ArErr{}
	} else if step == 1 {
		return m[start:end.(int)], ArErr{}
	} else {
		output := []byte{}
		if step > 0 {
			for i := start; i < end.(int); i += step {
				output = append(output, m[i])
			}
		} else {
			for i := end.(int) - 1; i >= start; i += step {
				output = append(output, m[i])
			}
		}
		return string(output), ArErr{}
	}
}
