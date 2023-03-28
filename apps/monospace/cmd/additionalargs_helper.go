package cmd

import (
	"os"

	"github.com/software-t-rex/monospace/utils"
)

func getAdditionalArgs() []string {
	additionalArgsIndex := utils.SliceFindIndex(os.Args, "--")
	var additionalArgs []string
	if additionalArgsIndex > 0 {
		additionalArgs = os.Args[additionalArgsIndex+1:]
	}
	return additionalArgs
}

// return args passed after a double hyphen
func splitAdditionalArgs(args *[]string) []string {
	additionalArgs := getAdditionalArgs()
	length := len(additionalArgs)
	if length > 0 {
		tmp := *args
		tmp = tmp[0 : len(*args)-length]
		*args = tmp
	}
	return additionalArgs
}
