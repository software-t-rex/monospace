package scaffolders

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var pmInitArgs = map[string][]string{
	"npm":  {"init", "-y"},
	"pnpm": {"init"},
	"yarn": {"init"},
}

func getPMcmd() string {
	pmCfg := os.Getenv("MONOSPACE_JSPM")
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

func Javascript() (err error) {
	hasPackageJson := fileExistsNoErr("package.json")
	if hasPackageJson {
		fmt.Println("JS scaffolding: package.json already exists => skip")
		return
	}
	hasIndexJs := fileExistsNoErr("index.js")
	if !hasIndexJs {
		fmt.Printf("init index.js file\n")
		err = writeTemplateFile("index.js", "./index.js")
		if err != nil {
			printWarning("Error writing index.js")
			err = nil
		}
	}
	pm := getPMcmd()
	pmArgs, hasPmArgs := pmInitArgs[pm]
	if !hasPmArgs {
		printWarning(fmt.Sprintf("can't init '%s' package manager", pm))
	}
	//@todo propose to add to workspace file
	// #nosec G204 - vars don't come from user inputs
	cmd := exec.Command(pm, pmArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
