package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/software-t-rex/monospace/gomodules/ui"
)

func PrintError(err error) {
	if err != nil {
		theme := ui.GetTheme()
		fmt.Fprintln(os.Stderr, theme.EmphasedError("Error:"), theme.Error(err.Error()))
	}
}
func PrintWarning(warning ...string) {
	if len(warning) > 0 && warning[0] != "" {
		theme := ui.GetTheme()
		fmt.Fprintln(os.Stderr, theme.EmphasedWarning("Warning:"), theme.Warning(warning...))
	}
}
func PrintSuccess(success ...string) {
	if len(success) > 0 && success[0] != "" {
		theme := ui.GetTheme()
		fmt.Fprintln(os.Stdout, theme.EmphasedSuccess("Success:"), theme.Success(success...))
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
		PrintError(fmt.Errorf("%s: %w", msg, err))
		os.Exit(1)
	}
}

func debug(outType string, withTrace bool, vals ...any) {
	var printer func(any)
	theme := ui.GetTheme()
	fName := func(pc uintptr) string {
		fns := strings.Split(runtime.FuncForPC(pc).Name(), ".")
		return fns[len(fns)-1]
	}
	if outType == "json" {
		printer = func(val any) {
			out, _ := json.MarshalIndent(val, "", "  ")
			fmt.Print(theme.Error(string(out)) + "\n")
		}
	} else if outType == "#v" {
		printer = func(val any) { fmt.Printf(theme.Error("%#v\n"), val) }
	} else {
		printer = func(val any) { fmt.Printf(theme.Error("%+v\n"), val) }
	}

	pc, filename, line, _ := runtime.Caller(2)
	fmt.Printf(theme.EmphasedInfo("[Debug start]")+theme.Info(" %s() %s:%d")+"\n", fName(pc), filename, line)
	for _, val := range vals {
		printer(val)
	}
	if withTrace {
		fmt.Println(theme.Info("Call Stack:"))
		for i := 3; ; i++ {
			pc, filename, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			fmt.Printf(theme.Info("  - %s() %s:%d\n"), fName(pc), filename, line)
		}
	}
	fmt.Println(theme.EmphasedInfo("[End Debug]"))
}
func Debug(vals ...any)               { debug("+v", false, vals...) }
func DebugWithStack(vals ...any)      { debug("+v", true, vals...) }
func DebugSharp(vals ...any)          { debug("#v", false, vals...) }
func DebugSharpWithStack(vals ...any) { debug("#v", true, vals...) }
func DebugJson(vals ...any)           { debug("json", false, vals...) }
func DebugJsonWithStack(vals ...any)  { debug("json", true, vals...) }
