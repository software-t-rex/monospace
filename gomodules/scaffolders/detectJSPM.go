package scaffolders

import (
	"fmt"
	"os/exec"
	"strings"
)

type jspm struct {
	cmd     string
	version string
}

func (j jspm) GetConfigVersionString() string {
	return fmt.Sprintf("^%s@%s", j.cmd, j.version)
}

func detectJSPM(pmCmds []string) []jspm {
	availableJSPM := []jspm{}
	for _, pmCmd := range pmCmds {
		if cmdAvailable(pmCmd) {
			getVersionCmd := exec.Command(pmCmd, "--version")
			version, err := getVersionCmd.CombinedOutput()
			if err == nil {
				availableJSPM = append(availableJSPM, jspm{cmd: pmCmd, version: strings.TrimSpace(string(version))})
			}
		}
	}
	return availableJSPM
}
