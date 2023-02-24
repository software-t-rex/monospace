package scaffolders

import (
	"fmt"
	"os"
	"os/exec"
)

func Monospace() error {

	fmt.Printf("create .monospace.yml\n")
	err := writeTemplateFile("monospace", ".monospace.yml")
	if err != nil {
		return err
	}
	// check for npmRc or create it
	if !fileExistsNoErr(".npmrc") {
		fmt.Printf("create .npmrc\n")
		err = writeTemplateFile("npmrc", ".npmrc")
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
