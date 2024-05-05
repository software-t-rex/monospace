package utils

import (
	"fmt"
	"strings"
)

// This is a onliner If
func If[T any](b bool, t, f T) T {
	if b {
		return t
	}
	return f
}

func Indent(s string, indentation string) string {
	return fmt.Sprintf("%s%s\n", indentation, strings.Replace(s, "\n", "\n"+indentation, -1))
}
