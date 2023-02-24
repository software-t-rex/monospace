package utils

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/software-t-rex/monospace/colors"
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

func GitAddGitIgnoreFile() error {
	if FileExistsNoErr(".gitignore") {
		fmt.Println(".gitignore already exists, left untouched")
		return nil
	}
	fmt.Println("Add default .gitignore")
	return WriteFile(".gitignore", "node_modules\n.vscode\ndist\ncoverage\n")
}

func GitInit(directory string, addIgnoreFile bool) (err error) {
	if FileExistsNoErr(".git") {
		fmt.Println("git init: git already initialized => skip")
	} else {
		err = gitExec("init", directory)
		if err != nil {
			return err
		}
	}
	if addIgnoreFile {
		return GitAddGitIgnoreFile()
	}
	return err
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
