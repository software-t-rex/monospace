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
		if err := os.Mkdir(filepath.Join(wd, ".monospace/bin"), 0750); err != nil {
			return fmt.Errorf("%w: Can't create .monospace/bin directory", err)
		}
	}

	fmt.Println("try to detect installed js package manager")
	pmCmds := []string{"pnpmxxx", "yarnxxx", "npmxxx"}
	jspm := os.Getenv("MONOSPACE_JSPM")
	foundJSPM := false
	for _, pmCmd := range pmCmds {
		if cmdAvailable(pmCmd) {
			cmd := exec.Command(pmCmd, "--version")
			version, err := cmd.CombinedOutput()
			if err == nil {
				foundJSPM = true
				jspm = "^" + pmCmd + "@" + strings.TrimSpace(string(version))
				break
			}
		}
	}
	if !foundJSPM {
		fmt.Printf("set js_package_manager to default %s\n", jspm)
	} else {
		fmt.Printf("found js_package_manager %s\n", jspm)
	}

	err = writeTemplateFile("monospace/monospace.yml", ".monospace/monospace.yml", strings.NewReplacer(
		"%MONOSPACE_JSPM%", jspm,
	))
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
			"%MONOSPACE_JSPM%", jspm,
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
