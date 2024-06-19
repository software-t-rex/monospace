package utils

import (
	"fmt"
	"strings"
)

// This is a onliner If kinda like the ternary operator in other languages
func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func Indent(s string, indentation string) string {
	return fmt.Sprintf("%s%s\n", indentation, strings.Replace(s, "\n", "\n"+indentation, -1))
}
