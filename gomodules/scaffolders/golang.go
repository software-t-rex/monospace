package scaffolders

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

func Golang() error {
	hasGoMod := fileExistsNoErr("go.mod")
	if hasGoMod {
		fmt.Println("Go scaffolding: go.mod already exists => skip")
		return nil
	}
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !fileExistsNoErr("main.go") {
		err = writeTemplateFile("golang/main.go", filepath.Join(pwd, "main.go"), nil)
		if err != nil {
			printWarning("Error writing main.go")
			err = nil
		}
	}

	if !cmdAvailable("go") {
		printWarning("can't find 'go' command you will need to init the project manually")
		return nil
	}

	pName := filepath.Base(pwd)
	goModPrefix := os.Getenv("MONOSPACE_GOPREFIX")
	if goModPrefix == "" {
		currentUser, err := user.Current()
		var username string
		if err != nil {
			username = "USERNAME"
		} else {
			username = currentUser.Username
		}
		goModPrefix = fmt.Sprintf("host.local/%s", username)
	}
	// @todo check for valid go module prefix
	moduleName := fmt.Sprintf("%s/%s", goModPrefix, pName)

	// #nosec G204 - module name should be verified
	cmd := exec.Command("go", "mod", "init", moduleName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	monoRoot := os.Getenv("MONOSPACE_ROOT")
	projectPath := os.Getenv("MONOSPACE_PROJECT_PATH")

	if monoRoot == "" || projectPath == "" {
		return err
	}
	if Confirm("Do you want to add project to monospace go.work?", true) {
		cmd2 := exec.Command("go", "work", "edit", "-use", "./"+projectPath)
		cmd2.Dir = monoRoot
		cmd2.Stdin = os.Stdin
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		return cmd2.Run()
	}
	return nil
}
