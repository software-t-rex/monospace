package scaffolders

import (
	"embed"
	"fmt"
	"os"
	"os/exec"

	"github.com/software-t-rex/monospace/colors"
)

//go:embed templates/*
var templateFS embed.FS

func cmdAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

var warningStyle = colors.Style(colors.Yellow)

func printWarning(msg string) {
	fmt.Println(warningStyle(msg))
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, err
	} else if os.IsNotExist(err) { // not exists is not an error in this context
		return false, nil
	}
	return false, err
}

func fileExistsNoErr(filePath string) bool {
	res, err := fileExists(filePath)
	if err != nil {
		return false
	}
	return res
}

func writeTemplateFile(src string, dest string) error {
	templateStr, err := templateFS.ReadFile("templates/" + src + ".tpl")
	if err != nil {
		return err
	}
	// #nosec G306 - we want the group access
	err = os.WriteFile(dest, []byte(templateStr), 0640)
	return err
}
