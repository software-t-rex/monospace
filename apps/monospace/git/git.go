package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/gomodules/colors"
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
func GitIsClean(repoDir string, subDir string) bool {
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
func GitClone(repoUrl string, destPath string) error {
	return gitExec("clone", repoUrl, destPath)
}

// add default .gitignore to current directory
func GitAddGitIgnoreFile() error {
	if utils.FileExistsNoErr(".gitignore") {
		fmt.Println(".gitignore already exists, left untouched")
		return nil
	}
	fmt.Println("Add default .gitignore")
	return utils.WriteFile(".gitignore", "node_modules\n.vscode\ndist\ncoverage\n")
}

// initialize a git repo in the given directory
func GitInit(directory string, addIgnoreFile bool) (err error) {
	if utils.FileExistsNoErr(".git") {
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

// print the last git commit for given directory
func GitHistoryLastCommit(directory string) (res string, err error) {
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
	/* #nosec G204 - cParam is not a user input */
	cmd := exec.Command("git", args...)
	cmd.Dir = directory
	var resBytes []byte
	resBytes, err = cmd.CombinedOutput()
	res = string(resBytes)
	return res, err
}

func GitGetOrigin(directory string) (string, error) {
	cmd := exec.Command("git", "-C", directory, "remote", "get-url", "origin")
	var errMsg bytes.Buffer
	cmd.Stderr = &errMsg
	origin, err := cmd.Output()
	if err == nil && string(origin) != "" {
		return strings.TrimSpace(string(origin)), nil
	}
	if errMsg.Len() > 0 {
		return "", fmt.Errorf(strings.TrimSpace(errMsg.String()))
	}
	return "", err
}
func GitHasOrigin(directory string) (bool, error) {
	cmd := exec.Command("git", "-C", directory, "remote", "show")
	remotes, err := cmd.Output()
	if err != nil {
		return false, err
	}
	if utils.SliceContains(strings.Split(string(remotes), "\n"), "origin") {
		return true, nil
	}
	return false, nil
}
func GitSetOrigin(directory string, origin string) error {
	has, err := GitHasOrigin(directory)
	if err != nil {
		return err
	}
	if has {
		return gitExec("-C", directory, "remote", "set-url", "origin", origin)
	}
	return gitExec("-C", directory, "remote", "add", "origin", origin)
}

func GitRemoveOrigin(directory string) error {
	return gitExec("-C", directory, "remote", "remove", "origin")
}

func GitConfigSet(directory string, key string, value string) error {
	return gitExec("-C", directory, "config", key, value)
}

func GitIsRepoRootDir(directory string) bool {
	return utils.FileExistsNoErr(directory + "/.git")
}

type GitExternalizeOptions struct {
	InitialBranch string
	Origin        string
	PushOrigin    bool
	AllowStash    bool
}

// @todo Add some user confirm like commit change at the end
// @todo handle rollback in case it doesn't go well

// initialize a new git repo in subdir for a project within the monoRootDir keeping its history
// can exit on cleanup
func GitExternalize(monoRootDir string, subDir string, opts GitExternalizeOptions) (err error) {
	cleanExp, err := regexp.Compile("[^a-zA-Z0-9_-]+")
	if err != nil {
		return err
	}
	tmpBranchName := "externalize-" + cleanExp.ReplaceAllString(subDir, "-")

	// move to parent directory
	err = os.Chdir(monoRootDir)
	if err != nil {
		return err
	}

	// check repo state and optionally stash changes if any
	if !GitIsClean(monoRootDir, subDir) {
		if !opts.AllowStash {
			return fmt.Errorf("'%s' is not clean, either clean the dir or run this command in interactive mode", subDir)
		}
		// stash changes and restore them at the end
		fmt.Printf(utils.Bold("Stashing changes in %s, will unstash at the end\n"), subDir)
		err = gitExec("stash", "push", "-a", "-m ", "monospace-externalizing", "--", subDir)
		defer func() {
			cmd := exec.Command("git", "stash", "list")
			res, _ := cmd.CombinedOutput()
			lines := strings.Split(string(res), "\n")
			if err != nil {
				utils.PrintError(err)
			} else if len(lines) == 0 {
				fmt.Println("nothing to unstash")
			} else if len(lines) > 0 && strings.Contains(lines[0], "monospace-externalizing") {
				fmt.Println(utils.Bold("Unstashing saved changes"))
				utils.CheckErr(gitExec("stash", "pop"))
			}
		}()
	}

	// create a new branch containing the wanted files
	fmt.Println(utils.Bold("Create subtree branch", tmpBranchName))
	err = gitExec("subtree", "split", "-P", subDir, "--branch", tmpBranchName)
	if err != nil {
		return err
	}
	// don't forget to remove the newly created branch
	defer func() {
		if err != nil {
			fmt.Println(utils.Info("Files from", subDir, "can be recovered from branch", tmpBranchName))
			return
		}
		if os.Chdir(monoRootDir) == nil {
			fmt.Println(utils.Bold("delete temporary subtree branch", tmpBranchName))
			utils.CheckErr(gitExec("branch", "-D", tmpBranchName))
		}
	}()

	// clean up the project directory
	fmt.Println(utils.Bold("clean project directory", subDir))
	err = gitExec("rm", "-rf", subDir)
	if err != nil {
		return err
	}

	// init the new repo
	fmt.Println(utils.Bold("init git in project directory", subDir))
	if opts.InitialBranch != "" {
		err = gitExec("init", subDir, "--initial-branch", opts.InitialBranch)
	} else {
		err = gitExec("init", subDir)
	}
	if err != nil {
		return err
	}

	// add parent as remote and merge the branch
	fmt.Println(utils.Bold("Add parent as temporary remote and merge", tmpBranchName))
	err = gitExec("-C", subDir, "remote", "add", "remoteMonospace", monoRootDir)
	if err != nil {
		// @todo => perform a reset --hard to restore the original directory
		return
	}
	err = gitExec("-C", subDir, "fetch", "--no-tags", "remoteMonospace", tmpBranchName)
	if err != nil {
		// @todo => perform a reset --hard to restore the original directory
		return
	}
	err = gitExec("-C", subDir, "merge", "--ff", "remoteMonospace/"+tmpBranchName)
	if err != nil {
		// @todo => perform a reset --hard to restore the original directory
		return
	}
	fmt.Println(utils.Bold("Merge done, remove parent from remotes"))
	err = gitExec("-C", subDir, "remote", "remove", "remoteMonospace")
	if err != nil {
		return
	}

	// add origin if any
	if opts.Origin != "" {
		fmt.Println(utils.Bold("add project origin", opts.Origin))
		err = gitExec("-C", subDir, "remote", "add", "origin", opts.Origin)
		if err != nil {
			return
		}
		if opts.PushOrigin {
			fmt.Println(utils.Bold("push to", opts.Origin))
			notImportantErr := gitExec("-C", subDir, "push", "-u", "origin")
			if notImportantErr != nil {
				utils.PrintWarning("Error while pushing to origin", notImportantErr.Error())
			}
		}
	}

	// add project as external in the monospace
	fmt.Println(utils.Bold("add project to monospace gitignore and set it as external"))
	err = utils.FileAppend(filepath.Join(monoRootDir, "/.gitignore"), subDir)
	if err != nil {
		return err
	}
	if opts.Origin != "" {
		err = app.ConfigAddOrUpdateProject(subDir, opts.Origin, true)
	} else {
		err = app.ConfigAddOrUpdateProject(subDir, "local", true)
	}
	if err != nil {
		return err
	}

	// stage modified files
	fmt.Println(utils.Bold("Stage changed files"))
	err = gitExec("add", ".gitignore", app.DfltcfgFilePath)
	if err == nil {
		fmt.Println("You can review the changes before committing")
	}

	return

}
