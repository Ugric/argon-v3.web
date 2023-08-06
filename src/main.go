package main

import (
	"fmt"
	"os"
	"syscall/js"
)

// args without the program path
var Args = os.Args[1:]

type stack = []ArObject

const VERSION = "3.0.0"

// Example struct
type Person struct {
	Name string
	Age  int
}

func newscope() ArObject {
	return Map(anymap{})
}

func await(awaitable js.Value) ([]js.Value, []js.Value) {
	then := make(chan []js.Value)
	defer close(then)
	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		then <- args
		return nil
	})
	defer thenFunc.Release()

	catch := make(chan []js.Value)
	defer close(catch)
	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		catch <- args
		return nil
	})
	defer catchFunc.Release()

	awaitable.Call("then", thenFunc).Call("catch", catchFunc)

	select {
	case result := <-then:
		return result, nil
	case err := <-catch:
		return nil, err
	}
}

func main() {
	debugInit()
	c := make(chan bool)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("There was a fundamental error in argon v3 that caused it to crash.")
			fmt.Println()
			if fork {
				fmt.Println("This is a fork of Open-Argon. Please report this to the fork's maintainer.")
				fmt.Println("Fork repo:", forkrepo)
				fmt.Println("Fork issue page:", forkissuesPage)
				fmt.Println()
				fmt.Println("website:", website)
				fmt.Println("docs:", docs)
				fmt.Println()
				if fork {
					fmt.Println("This is a fork of Open-Argon. Please report this to the fork's maintainer.")
					fmt.Println("Fork repo:", forkrepo)
					fmt.Println("Fork issue page:", forkissuesPage)
					fmt.Println()
				} else {
					fmt.Println("Please report this to the Open-Argon team.")
					fmt.Println("Main repo:", mainrepo)
					fmt.Println("Issue page:", mainissuesPage)
					fmt.Println()
				}
				fmt.Println("please include the following information:")
				fmt.Println("panic:", r)
				os.Exit(1)
			}
		}
	}()
	initRandom()
	garbageCollect()
	global := makeGlobal()
	obj := js.Global().Get("Object").New()
	obj.Set("import", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			importMod(args[0].String(), "", false, global)
		}()

		return nil
	}))
	js.Global().Set("Ar", obj)
	<-c
}
