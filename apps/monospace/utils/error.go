package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func PrintError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, EmphasedError("Error:"), ErrorStyle(err.Error()))
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

func Dump(o ...any) {
	for _, val := range o {
		fmt.Printf("%+v\n", val)
		fmt.Printf("%#v\n", val)
		out, _ := json.MarshalIndent(&val, "", "  ")
		fmt.Print(string(out) + "\n")
	}
}
