package scaffolders

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"
)

func Golang() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = writeTemplateFile("main.go", filepath.Join(pwd, "main.go"))
	if err != nil {
		printWarning("Error writing main.go")
		err = nil
	}

	if !cmdAvailable("go") {
		printWarning("can't find 'go' command you will need to init the project manually")
		return nil
	}

	pName := filepath.Base(pwd)
	goModPrefix := viper.GetString("go_mod_prefix")
	if goModPrefix == "" {
		goModPrefix = pName
	}

	//@todo propose to add to go.work

	cmd := exec.Command("go", "mod", "init", goModPrefix+"/"+pName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
