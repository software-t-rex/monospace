package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"monospace/monospace/cmd/colors"
	"os"
)

func Exit(errorMsg string) {
	PrintError(errors.New(errorMsg))
	os.Exit(1)
}

func PrintError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, colors.EmphasedError("Error:"), colors.Error(err.Error()))
	}
}

func CheckErr(err error) {
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}
}
func CheckErrWithMsg(err error, msg string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, colors.Error(msg))
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
