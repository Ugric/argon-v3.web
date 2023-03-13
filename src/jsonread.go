package main

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

func convertToArgon(obj any) any {
	switch x := obj.(type) {
	case map[string]interface{}:
		newmap := ArMap{}
		for key, value := range x {
			newmap[key] = convertToArgon(value)
		}
		return newmap
	case ArArray:
		newarray := ArArray{}
		for _, value := range x {
			newarray = append(newarray, convertToArgon(value))
		}
		return newarray
	case string:
		return x
	case float64:
		return newNumber().SetFloat64(x)
	case bool:
		return x
	case nil:
		return nil
	}
	return nil
}

func parse(str string) any {
	var jsonMap any
	json.Unmarshal([]byte(str), &jsonMap)
	return convertToArgon(jsonMap)
}

func stringify(obj any) (string, error) {
	output := []string{}
	obj = classVal(obj)
	switch x := obj.(type) {
	case ArMap:
		for key, value := range x {
			str, err := stringify(value)
			if err != nil {
				return "", err
			}
			output = append(output, ""+strconv.Quote(anyToArgon(key, false, true, 3, 0, false, 0))+": "+str)
		}
		return "{" + strings.Join(output, ", ") + "}", nil
	case ArArray:
		output = append(output, "[")
		for _, value := range x {
			str, err := stringify(value)
			if err != nil {
				return "", err
			}
			output = append(output, str)
		}
		output = append(output, "]")
		return strings.Join(output, ", "), nil
	case string:
		return strconv.Quote(x), nil
	case number:
		return anyToArgon(x, true, false, 1, 0, false, 0), nil
	case bool:
		return strconv.FormatBool(x), nil
	case nil:
		return "null", nil
	}
	err := errors.New("Cannot stringify '" + typeof(obj) + "'")
	return "", err
}

var ArJSON = ArMap{
	"parse": builtinFunc{"parse", func(args ...any) (any, ArErr) {
		if len(args) == 0 {
			return ArMap{}, ArErr{TYPE: "Runtime Error", message: "parse takes 1 argument", EXISTS: true}
		}
		if typeof(args[0]) != "string" {
			return ArMap{}, ArErr{TYPE: "Runtime Error", message: "parse takes a string not a '" + typeof(args[0]) + "'", EXISTS: true}
		}
		return parse(args[0].(string)), ArErr{}
	}},
	"stringify": builtinFunc{"stringify", func(args ...any) (any, ArErr) {
		if len(args) == 0 {
			return ArMap{}, ArErr{TYPE: "Runtime Error", message: "stringify takes 1 argument", EXISTS: true}
		}
		str, err := stringify(args[0])
		if err != nil {
			return ArMap{}, ArErr{TYPE: "Runtime Error", message: err.Error(), EXISTS: true}
		}
		return str, ArErr{}
	}},
}
