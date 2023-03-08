package scaffolders

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Monospace() error {
	fmt.Printf("create .monospace.yml\n")
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	monoName := filepath.Base(wd)
	// @todo handle some configs like gomodule prefix, and prefered package manager
	err = writeTemplateFile("monospace.yml", ".monospace.yml", nil)
	if err != nil {
		return err
	}

	// @TODO detect package manager and use either package.json or pnpm-workspace.yaml
	if !fileExistsNoErr("package.json") {
		fmt.Printf("create package.json\n")
		err = writeTemplateFile("monospace-package.json", "package.json", strings.NewReplacer(
			"%MONOSPACE_NAME%", monoName,
		))
		if err != nil {
			return err
		}
	}

	// check for npmRc or create it
	if !fileExistsNoErr(".npmrc") {
		fmt.Printf("create .npmrc\n")
		err = writeTemplateFile("npmrc", ".npmrc", nil)
		if err != nil {
			return err
		}
	}

	if !fileExistsNoErr("go.work") {
		if cmdAvailable("go") {
			fmt.Printf("init go.work\n")
			cmd := exec.Command("go", "work", "init")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
		}
	}

	return err
}
