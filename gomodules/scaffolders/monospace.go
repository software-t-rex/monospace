package scaffolders

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Monospace() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	monoName := filepath.Base(wd)
	// @todo handle some configs like gomodule prefix, and prefered package manager
	fmt.Printf("create .monospace/monospace.yml\n")
	if !IsDirNoErr(filepath.Join(wd, ".monospace")) {
		if err := os.Mkdir(filepath.Join(wd, ".monospace"), 0750); err != nil {
			return fmt.Errorf("%w: Can't create .monospace directory", err)
		}
	}
	err = writeTemplateFile("monospace.yml", ".monospace/monospace.yml", nil)
	if err != nil {
		return err
	}

	if !fileExistsNoErr("pnpm-workspace.yaml") && Confirm("Do you want to create a pnpm-workspace.yaml file?", true) {
		fmt.Println("create pnpm-workspace.yaml")
		err = writeTemplateFile("pnpm-workspace.yaml", "pnpm-workspace.yaml", nil)
		if err != nil {
			return err
		}
	}
	if !fileExistsNoErr("package.json") && Confirm("Do you want to create a package.json file?", true) {
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
		if cmdAvailable("go") && Confirm("Would you like to create a go.work file ?", true) {
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
