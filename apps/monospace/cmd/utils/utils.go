package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"monospace/monospace/cmd/colors"
	"os"
	"strings"
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

var underline = colors.Style(colors.Underline)

func Confirm(msg string, dflt bool) bool {
	var response string
	dfltString := " [y/" + underline("N") + "]: "
	if dflt {
		dfltString = " [" + underline("Y") + "|n]: "
	}
	fmt.Print(msg + dfltString)
	_, err := fmt.Scan(&response)
	CheckErr(err)
	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	case "":
		return dflt
	default:
		fmt.Println(colors.Style(colors.Red)("Please type (y)es or (n)o and then press enter:"))
		return Confirm(msg, dflt)
	}
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func WriteFile(filePath string, body string) error {
	bbody := []byte(body)
	return os.WriteFile(filePath, bbody, 0644)
}

func FileAppend(filePath string, appendString string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	if !strings.HasSuffix(appendString, "\n") {
		appendString = appendString + "\n"
	}
	_, err = f.WriteString(appendString)
	return err
}

func FileRemoveLine(filePath string, line string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	cleanContent := bytes.Replace(content, []byte(line+"\n"), []byte(""), -1)
	return os.WriteFile(filePath, cleanContent, 0644)
}

func MakeDir(path string) error {
	return os.MkdirAll(path, 0750)
}

func RmDir(path string) error {
	return os.RemoveAll(path)
}

func MapGetKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	i := 0
	for k := range m {
		keys = append(keys, k)
		i++
	}
	return keys
}

func Filter[T any](array []T, predicate func(T) bool) (res []T) {
	for _, val := range array {
		if predicate(val) {
			res = append(res, val)
		}
	}
	return
}

func Map[T any, V any, Mapper func(T) V](array []T, mapper Mapper) (res []V) {
	for _, val := range array {
		res = append(res, mapper(val))
	}
	return
}

func MapAndFilter[T any, V any, Mapper func(T) (V, bool)](array []T, mapper Mapper) (res []V) {
	for _, val := range array {
		newVal, keep := mapper(val)
		if keep {
			res = append(res, newVal)
		}
	}
	return
}

func PrefixPredicate(prefix string) func(string) bool {
	return func(s string) bool {
		return strings.HasPrefix(s, prefix)
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
