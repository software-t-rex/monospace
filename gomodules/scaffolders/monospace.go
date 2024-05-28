package scaffolders

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/software-t-rex/monospace/gomodules/ui"
)

func Monospace() error {
	theme := ui.GetTheme()
	noInteractive := os.Getenv("MONOSPACE_NO_INTERACTIVE")
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	monoName := filepath.Base(wd)

	jspm := os.Getenv("MONOSPACE_JSPM")
	willJS := Confirm("Will you use javascript in this monospace ?", true)
	if willJS {
		fmt.Println("try to detect installed js package manager")
		pmCmds := []string{"pnpm", "yarn", "npm"}
		foundJSPM := detectJSPM(pmCmds)
		if len(foundJSPM) == 0 {
			fmt.Println(theme.Warning("no package manager found on your system will use default " + jspm))
		} else if len(foundJSPM) == 1 {
			jspm = foundJSPM[0].GetConfigVersionString()
			fmt.Printf("found one package manager set js_package_manager to %s\n", jspm)
		} else if noInteractive != "" {
			// select first option
			fmt.Printf("multiple package managers found, select %s\n", foundJSPM[0].cmd)
			jspm = foundJSPM[0].GetConfigVersionString()
		} else {
			// propose selection to user
			jspmOptions := []ui.SelectOption[string]{}
			for _, jspm := range foundJSPM {
				jspmOptions = append(jspmOptions, ui.SelectOption[string]{
					Label: jspm.cmd,
					Value: jspm.GetConfigVersionString(),
				})
			}
			SelectedJspm, errSelect := ui.NewSelect("multiple package managers found, select the one to use :", jspmOptions).Run()
			if errSelect == nil {
				jspm = SelectedJspm
			} else {
				fmt.Printf(theme.Warning("Error while selecting package manager: %w\n"), errSelect)
			}

			fmt.Printf("set js_package_manager to %s\n", jspm)
		}
	}

	fmt.Println("create .monospace/monospace.yml")
	if !IsDirNoErr(filepath.Join(wd, ".monospace")) {
		if err := os.Mkdir(filepath.Join(wd, ".monospace"), 0750); err != nil {
			return fmt.Errorf("%w: Can't create .monospace directory", err)
		}
		if err := os.Mkdir(filepath.Join(wd, ".monospace/bin"), 0750); err != nil {
			return fmt.Errorf("%w: Can't create .monospace/bin directory", err)
		}
	}
	err = writeTemplateFile("monospace/monospace.yml", ".monospace/monospace.yml", strings.NewReplacer(
		"%MONOSPACE_JSPM%", jspm,
	))
	if err != nil {
		return fmt.Errorf("%w: Can't create .monospace/monospace.yml file", err)
	}
	fmt.Printf("monospace.yml created %s\n", theme.SuccessIndicator())

	if willJS && strings.Contains(jspm, "pnpm") && !fileExistsNoErr("pnpm-workspace.yaml") && Confirm("Do you want to create a pnpm-workspace.yaml file?", true) {
		fmt.Println("create pnpm-workspace.yaml")
		err = writeTemplateFile("monospace/pnpm-workspace.yaml", "pnpm-workspace.yaml", nil)
		if err != nil {
			err = nil
			fmt.Printf(theme.Warning("Can't create pnpm-workspace.yaml file: %w\n"), err)
		} else {
			fmt.Printf("pnpm-workspace.yaml created %s\n", theme.SuccessIndicator())
		}
	}
	if willJS && !fileExistsNoErr("package.json") && Confirm("Do you want to create a package.json file?", true) {
		fmt.Printf("create package.json\n")
		err = writeTemplateFile("monospace/monospace-package.json", "package.json", strings.NewReplacer(
			"%MONOSPACE_NAME%", monoName,
			"%MONOSPACE_JSPM%", jspm,
		))
		if err != nil {
			err = nil
			fmt.Printf(theme.Warning("Can't create package.json file: %w\n"), err)
		} else {
			fmt.Printf("package.json created %s\n", theme.SuccessIndicator())
		}
	}

	// check for npmRc or create it
	if willJS && !fileExistsNoErr(".npmrc") {
		fmt.Printf("create .npmrc with recommanded settings\n")
		err = writeTemplateFile("monospace/npmrc", ".npmrc", nil)
		if err != nil {
			err = nil
			fmt.Printf(theme.Warning("Can't create .npmrc file: %w\n"), err)
		} else {
			fmt.Printf(".npmrc created %s\n", theme.SuccessIndicator())
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
			if err != nil {
				err = nil
				fmt.Printf(theme.Warning("Can't create go.work file: %w\n"), err)
			} else {
				fmt.Printf("go.work created %s\n", theme.SuccessIndicator())
			}
		}
	}

	// add hooks directory
	if Confirm("Do you want to install some git hooks in .monospace/githooks ?", true) {
		fmt.Println("adding some convenience git hooks in .monospace/githooks")
		if err := os.Mkdir(filepath.Join(wd, ".monospace/githooks"), 0750); err == nil {
			writeExecutableTemplateFile("monospace/githooks/check", ".monospace/githooks/post-merge", nil)
			writeExecutableTemplateFile("monospace/githooks/check", ".monospace/githooks/post-checkout", nil)
		}
		if err != nil {
			err = nil
			fmt.Printf(theme.Warning("Can't create .monospace/githooks directory: %w\n"), err)
		} else {
			fmt.Printf(".monospace/githooks created %s\n", theme.SuccessIndicator())
		}
	}
	return err
}
