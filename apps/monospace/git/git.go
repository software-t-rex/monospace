package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/software-t-rex/monospace/utils"
)

// execute a git command in current directory redirecting stdin/out/err to main process
func gitExec(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// exec git command in given directory prepending args with "-C <directory>"
// if directory is empty or . or ./ then execute in current directory
// return combinedOutput
func gitExecOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	res, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(res)), err
}

// exec git command in given directory prepending args with "-C <directory>"
// if directory is empty or . or ./ then execute in current directory
func ExecDir(directory string, args ...string) error {
	if directory == "" || directory == "./" || directory == "." {
		return gitExec(args...)
	}
	return gitExec(append([]string{"-C", directory}, args...)...)
}

// check the current git dir is clean
func IsClean(repoDir string, subDir string) bool {
	args := []string{}
	if repoDir != "" {
		args = append(args, "-C", repoDir)
	}
	args = append(args, "status", "--porcelain", "--ignored")
	if subDir != "" {
		args = append(args, "--", subDir)
	}
	/* #nosec G204 - only directories come from the outside */
	cmd := exec.Command("git", args...)
	res, err := cmd.CombinedOutput()
	if err == nil && strings.TrimSpace(string(res)) == "" {
		return true
	}
	return false
}

// clone given repo to destPath directory
func Clone(repoUrl string, destPath string) error {
	return gitExec("clone", repoUrl, destPath)
}

// add default .gitignore to current directory
func AddGitIgnoreFile() error {
	if utils.FileExistsNoErr(".gitignore") {
		fmt.Println(".gitignore already exists, left untouched")
		return nil
	}
	fmt.Println("Add default .gitignore")
	return utils.WriteFile(".gitignore", "node_modules\n.vscode\ndist\ncoverage\n")
}

// initialize a git repo in the given directory
func Init(directory string, addIgnoreFile bool) (err error) {
	if utils.FileExistsNoErr(".git") {
		fmt.Println("git init: git already initialized => skip")
	} else {
		err = gitExec("init", directory)
		if err != nil {
			return err
		}
	}
	if addIgnoreFile {
		return AddGitIgnoreFile()
	}
	return err
}

/*/ print the last git commit for given directory
func HistoryLastCommit(directory string) (res string, err error) {
	var args []string
	if colors.ColorEnabled() {
		args = append(args, "-c", "color.ui=always")
	}
	args = append(args,
		"log",
		"--pretty=format:'%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%cr) %C(bold blue)<%an>%Creset'",
		"--abbrev-commit",
		"HEAD^..HEAD",
	)
	/* #nosec G204 - cParam is not a user input * /
	cmd := exec.Command("git", args...)
	cmd.Dir = directory
	var resBytes []byte
	resBytes, err = cmd.CombinedOutput()
	res = string(resBytes)
	return res, err
}*/

func IsRepoRootDir(directory string) bool {
	return utils.FileExistsNoErr(directory + "/.git")
}
