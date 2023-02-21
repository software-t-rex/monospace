package utils

import (
	"monospace/colors"
	"os"
	"os/exec"
)

func gitExec(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func GitClone(repoUrl string, destPath string) error {
	return gitExec("clone", repoUrl, destPath)
}

func GitInit(directory string) error {
	return gitExec("init", directory)
}

func GitHistoryLastCommit(directory string) (res string, err error) {
	var cParam string
	if colors.ColorEnabled() {
		cParam = "--c color.ui=always"
	}
	cmd := exec.Command(
		"git",
		"log",
		cParam,
		"--pretty=format:'%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%cr) %C(bold blue)<%an>%Creset'",
		"--abbrev-commit",
		"HEAD^..HEAD",
	)
	cmd.Dir = directory
	var resBytes []byte
	resBytes, err = cmd.CombinedOutput()
	res = string(resBytes)
	return res, err
}
