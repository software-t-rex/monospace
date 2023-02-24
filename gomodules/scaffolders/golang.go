package scaffolders

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/spf13/viper"
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
		err = writeTemplateFile("main.go", filepath.Join(pwd, "main.go"))
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
	goModPrefix := viper.GetString("go_mod_prefix")
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

	//@todo propose to add to go.work

	cmd := exec.Command("go", "mod", "init", goModPrefix+"/"+pName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
