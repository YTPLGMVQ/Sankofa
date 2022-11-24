// miscelaneous utility methods
//
// # DESIGN, TACTICS AND HACKS
//
// We use:
//   - int64 for all math functions.
//     these are required for position ranks on 32-bit architectures.
//     by using them everywhere, we avoid duplicating constants and conversions.
//     CAVEAT: must convert to int for array/slice indices.
package ow

// logging

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"

	"golang.org/x/text/message"
)

func init() {
	// remove time stamp from logs
	log.SetFlags(0)
}

// print debug messages or not
var Verbose bool = true
var printer = message.NewPrinter(message.MatchLanguage("en"))

// logger with call trace and goroutine number (yes, it's possible and useful)
func Log(message ...interface{}) {
	var prefix string
	if !Verbose {
		return
	}
	// identify goroutine
	// from https://blog.sgmansfield.com/2015/12/goroutine-ids/
	stack := make([]byte, 64)
	stack = stack[:runtime.Stack(stack, false)]
	stack = bytes.TrimPrefix(stack, []byte("goroutine "))
	stack = stack[:bytes.IndexByte(stack, ' ')]
	goroutine, _ := strconv.ParseUint(string(stack), 10, 64)
	// identify caller (package.method)
	programCounter, _, _, ok := runtime.Caller(1)
	if ok {
		method := runtime.FuncForPC(programCounter).Name()
		prefix = fmt.Sprintf("%v:%v:", goroutine, method)
	}
	log.Println(prefix, message)
}

// dump stack trace, print error message and exit
//
// this is NOT Golang's panic() and there is no try-catch mechanism!
func Panic(message ...interface{}) {
	fmt.Fprintln(os.Stderr, "Guru Meditation ⇢", message)
	debug.PrintStack()
	os.Exit(42)
}

// check error and panic if needed
func Check(err error) {
	if err != nil {
		Panic(err)
	}
}
