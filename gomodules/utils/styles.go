package utils

import (
	"fmt"
	"strings"
)

func Indent(s string, indentation string) string {
	return fmt.Sprintf("%s%s\n", indentation, strings.Replace(s, "\n", "\n"+indentation, -1))
}
