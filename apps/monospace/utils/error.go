package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func PrintError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, EmphasedError("Error:"), ErrorStyle(err.Error()))
	}
}
func PrintWarning(warning ...string) {
	if len(warning) > 0 && warning[0] != "" {
		fmt.Fprintln(os.Stderr, EmphasedWarning("Warning:"), Warning(warning...))
	}
}

func Exit(errorMsg string) {
	if errorMsg != "" {
		PrintError(errors.New(errorMsg))
	}
	os.Exit(1)
}

func CheckErr(err error) {
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}
}

func CheckErrOrReturn[T any](value T, err error) T {
	CheckErr(err)
	return value
}

func CheckErrWithMsg(err error, msg string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, ErrorStyle(msg))
		CheckErr(err)
	}
}

func debug(outType string, withTrace bool, vals ...any) {
	var printer func(any)
	fName := func(pc uintptr) string {
		fns := strings.Split(runtime.FuncForPC(pc).Name(), ".")
		return fns[len(fns)-1]
	}
	if outType == "json" {
		printer = func(val any) {
			out, _ := json.MarshalIndent(val, "", "  ")
			fmt.Print(ErrorStyle(string(out)) + "\n")
		}
	} else if outType == "#v" {
		printer = func(val any) { fmt.Printf(ErrorStyle("%#v\n"), val) }
	} else {
		printer = func(val any) { fmt.Printf(ErrorStyle("%+v\n"), val) }
	}

	pc, filename, line, _ := runtime.Caller(2)
	fmt.Printf(EmphasedInfo("[Debug start]")+Info(" %s() %s:%d")+"\n", fName(pc), filename, line)
	for _, val := range vals {
		printer(val)
	}
	if withTrace {
		fmt.Println(Info("Call Stack:"))
		for i := 3; ; i++ {
			pc, filename, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			fmt.Printf(Info("  - %s() %s:%d\n"), fName(pc), filename, line)
		}
	}
	fmt.Println(EmphasedInfo("[End Debug]"))
}
func Debug(vals ...any)               { debug("+v", false, vals...) }
func DebugWithStack(vals ...any)      { debug("+v", true, vals...) }
func DebugSharp(vals ...any)          { debug("#v", false, vals...) }
func DebugSharpWithStack(vals ...any) { debug("#v", true, vals...) }
func DebugJson(vals ...any)           { debug("json", false, vals...) }
func DebugJsonWithStack(vals ...any)  { debug("json", true, vals...) }
