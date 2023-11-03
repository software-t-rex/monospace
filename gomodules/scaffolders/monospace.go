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
	// @todo handle some configs like gomodule prefix, and preferred package manager
	fmt.Println("create .monospace/monospace.yml")
	if !IsDirNoErr(filepath.Join(wd, ".monospace")) {
		if err := os.Mkdir(filepath.Join(wd, ".monospace"), 0750); err != nil {
			return fmt.Errorf("%w: Can't create .monospace directory", err)
		}
	}
	err = writeTemplateFile("monospace/monospace.yml", ".monospace/monospace.yml", nil)
	if err != nil {
		return err
	}

	if !fileExistsNoErr("pnpm-workspace.yaml") && Confirm("Do you want to create a pnpm-workspace.yaml file?", true) {
		fmt.Println("create pnpm-workspace.yaml")
		err = writeTemplateFile("monospace/pnpm-workspace.yaml", "pnpm-workspace.yaml", nil)
		if err != nil {
			return err
		}
	}
	if !fileExistsNoErr("package.json") && Confirm("Do you want to create a package.json file?", true) {
		fmt.Printf("create package.json\n")
		err = writeTemplateFile("monospace/monospace-package.json", "package.json", strings.NewReplacer(
			"%MONOSPACE_NAME%", monoName,
		))
		if err != nil {
			return err
		}
	}

	// check for npmRc or create it
	if !fileExistsNoErr(".npmrc") {
		fmt.Printf("create .npmrc with recommanded settings\n")
		err = writeTemplateFile("monospace/npmrc", ".npmrc", nil)
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

	// add hooks directory
	if Confirm("Do you want to install some git hooks in .monospace/githooks ?", true) {
		fmt.Println("adding some convenience git hooks in .monospace/githooks")
		if err := os.Mkdir(filepath.Join(wd, ".monospace/githooks"), 0750); err == nil {
			writeExecutableTemplateFile("monospace/githooks/check", ".monospace/githooks/post-merge", nil)
			writeExecutableTemplateFile("monospace/githooks/check", ".monospace/githooks/post-checkout", nil)
		}
	}

	return err
}
