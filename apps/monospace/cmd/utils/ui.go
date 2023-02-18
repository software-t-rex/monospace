package utils

import (
	"fmt"
	"monospace/monospace/cmd/colors"
	"strings"
)

func Confirm(msg string, dflt bool) bool {
	var response string
	dfltString := " [y/" + Underline("N") + "]: "
	if dflt {
		dfltString = " [" + Underline("Y") + "|n]: "
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
