package scaffolders

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
)

var pmInitArgs = map[string][]string{
	"npm":  {"init", "-y"},
	"pnpm": {"init"},
	"yarn": {"init"},
}

func getPMcmd() string {
	pmCfg := viper.GetString("js_package_manager")
	if pmCfg == "" {
		return "pnpm"
	} else if strings.Contains(pmCfg, "pnpm") {
		return "pnpm"
	} else if strings.Contains(pmCfg, "yarn") {
		return "yarn"
	} else if strings.Contains(pmCfg, "npm") {
		return "npm"
	}
	return "pnpm"
}

func Javascript() error {
	pm := getPMcmd()
	cmd := exec.Command(pm, pmInitArgs[pm]...)

	fmt.Printf("init index.js file\n")
	err := writeTemplateFile("index.js", "./index.js")
	if err != nil {
		printWarning("Error writing main.go")
		err = nil
	}

	//@todo propose to add to workspace file

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
