package utils

import (
	"fmt"
	"strings"
)

func Confirm(msg string, dflt bool) bool {
	var response string
	dfltString := " [y/" + Underline("N") + "]: "
	if dflt {
		dfltString = " [" + Underline("Y") + "|n]: "
	}
	fmt.Print(msg + dfltString)
	_, err := fmt.Scanln(&response)
	if err != nil && err.Error() != "unexpected newline" {
		CheckErr(err)
	}
	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	case "":
		return dflt
	default:
		fmt.Println(Red("Please type (y)es or (n)o and then press enter:"))
		return Confirm(msg, dflt)
	}
}
