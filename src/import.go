package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall/js"
)

var imported = make(map[string]ArObject)
var importing = make(map[string]bool)

const modules_folder = "argon_modules"

func FileExists(filename string) bool {
	if info, err := os.Stat(filename); err == nil && !info.IsDir() {
		return true
	}
	return false
}

func readCode(code string, path string) []UNPARSEcode {
	split := strings.Split(code, "\n")
	output := []UNPARSEcode{}
	line := 1
	for _, text := range split {
		output = append(output, UNPARSEcode{text, text, line, path})
		line++
	}
	return output
}

func importMod(path string, origin string, main bool, global ArObject) (ArObject, ArErr) {
	p := path
	realpath := p
	result, err := await(js.Global().Call("fetch", path))
	if err != nil {
		return ArObject{}, ArErr{TYPE: "Import Error", message: "Could not fetch: " + path, EXISTS: true}
	}
	text, err := await(result[0].Call("text"))
	if err != nil {
		return ArObject{}, ArErr{TYPE: "Import Error", message: "Could not read text: " + path, EXISTS: true}
	}

	if importing[p] {
		return ArObject{}, ArErr{TYPE: "Import Error", message: "Circular import: " + path, EXISTS: true}
	} else if _, ok := imported[p]; ok {
		return imported[p], ArErr{}
	}
	importing[p] = true
	codelines := readCode(text[0].String(), realpath)
	translated, translationerr := translate(codelines)

	if translationerr.EXISTS {
		return ArObject{}, translationerr
	}
	ArgsArArray := []any{}
	withoutarfile := []string{}
	if len(Args) > 1 {
		withoutarfile = Args[1:]
	}
	for _, arg := range withoutarfile {
		ArgsArArray = append(ArgsArArray, arg)
	}
	local := newscope()
	localvars := Map(anymap{
		"program": Map(anymap{
			"args":   ArArray(ArgsArArray),
			"origin": origin,
			"import": builtinFunc{"import", func(args ...any) (any, ArErr) {
				if len(args) != 1 {
					return nil, ArErr{"Import Error", "Invalid number of arguments", 0, realpath, "", true}
				}
				if _, ok := args[0].(string); !ok {
					return nil, ArErr{"Import Error", "Invalid argument type", 0, realpath, "", true}
				}
				return importMod(args[0].(string), filepath.Dir(filepath.ToSlash(p)), false, global)
			}},
			"cwd": ArString(""),
			"exc": ArString(""),
			"file": Map(anymap{
				"name": filepath.Base(p),
				"path": p,
			}),
			"main": main,
		}),
	})
	_, runimeErr := ThrowOnNonLoop(run(translated, stack{global, localvars, local}))
	importing[p] = false
	if runimeErr.EXISTS {
		return ArObject{}, runimeErr
	}
	imported[p] = local
	return local, ArErr{}
}
