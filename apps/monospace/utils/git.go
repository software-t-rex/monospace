package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/software-t-rex/monospace/colors"
)

// execute a git command in current directory redirecting stdin/out/err to main process
func gitExec(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
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
	if FileExistsNoErr(".gitignore") {
		fmt.Println(".gitignore already exists, left untouched")
		return nil
	}
	fmt.Println("Add default .gitignore")
	return WriteFile(".gitignore", "node_modules\n.vscode\ndist\ncoverage\n")
}

// initialize a git repo in the given directory
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

// print the last git commit for given directory
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

type GitExternalizeOptions struct {
	InitialBranch string
	Origin        string
	PushOrigin    bool
	AllowStash    bool
}

// initialize a new git repo in subdir for a project within the parentDir keeping its history
func GitExternalize(parentDir string, subDir string, opts GitExternalizeOptions) (err error) {
	subRepoUrl := opts.Origin
	cleanExp, err := regexp.Compile("[^a-zA-Z0-9_-]+")
	if err != nil {
		return err
	}
	branchName := "externalize-" + cleanExp.ReplaceAllString(subDir, "-")

	// move to parent directory
	err = os.Chdir(parentDir)
	if err != nil {
		return err
	}

	// stash changes and restore them at the end
	fmt.Println("Stashing changes will unstash at the end")

	// check repo state and optionally stash changes if any
	if !GitIsClean(parentDir, subDir) {
		if !opts.AllowStash {
			return fmt.Errorf("'%s' is not clean, try to pass AllowStash to true", subDir)
		}
		fmt.Println("Stashing changes in", subDir)
		err = gitExec("stash", "push", "-a", "-m ", "monospace-externalizing", "--", subDir)
		defer func() {
			cmd := exec.Command("git", "stash", "list")
			res, _ := cmd.CombinedOutput()
			lines := strings.Split(string(res), "\n")
			if err != nil {
				PrintError(err)
			} else if len(lines) == 0 {
				fmt.Println("nothing to unstash")
			} else if len(lines) > 0 && strings.Contains(lines[0], "monospace-externalizing") {
				fmt.Println("Unstashing saved changes")
				gitExec("stash", "pop")
			}
		}()
	}

	// create a new branch containing the wanted files
	fmt.Println("Create subtree branch", branchName)
	err = gitExec("subtree", "split", "-P", subDir, "--branch", branchName)
	if err != nil {
		return err
	}
	// don't forget to remove the newly created branch
	defer func() {
		if err != nil {
			fmt.Println("Files from", subDir, "can be recovered from branch", branchName)
			return
		}
		if os.Chdir(parentDir) == nil {
			fmt.Println("delete temporary subtree branch", branchName)
			gitExec("branch", "-D", branchName)
		}
	}()

	// clean up the project directory
	fmt.Println("clean project directory", subDir)
	err = gitExec("rm", "-rf", subDir)
	if err != nil {
		return err
	}

	// init the new repo
	fmt.Println("init git in project directory", subDir)
	if opts.InitialBranch != "" {
		err = gitExec("init", subDir, "--initial-branch", opts.InitialBranch)
	} else {
		err = gitExec("init", subDir)
	}
	if err != nil {
		return err
	}

	// add parent as remote and merge the branch
	fmt.Println("Add parent as temporary remote and merge", branchName)
	err = gitExec("-C", subDir, "remote", "add", "remoteMonospace", parentDir)
	if err != nil {
		// @todo => perform a reset --hard to restore the original directory
		return
	}
	err = gitExec("-C", subDir, "fetch", "remoteMonospace")
	if err != nil {
		// @todo => perform a reset --hard to restore the original directory
		return
	}
	err = gitExec("-C", subDir, "merge", "--ff", "remoteMonospace/"+branchName)
	if err != nil {
		// @todo => perform a reset --hard to restore the original directory
		return
	}
	fmt.Println("Merge done, remove parent from remotes")
	err = gitExec("-C", subDir, "remote", "remove", "remoteMonospace")
	if err != nil {
		return
	}

	// add origin if any
	if subRepoUrl != "" {
		fmt.Println("add project origin", subRepoUrl)
		err = gitExec("-C", subDir, "remote", "add", "origin", subRepoUrl)
		if err != nil {
			return
		}
		if opts.PushOrigin {
			fmt.Println("push to", subRepoUrl)
			notImportantErr := gitExec("-C", subDir, "push", "-u", "origin")
			if notImportantErr != nil {
				PrintWarning("Error while pushing to origin", notImportantErr.Error())
			}
		}
	}

	// add project as external in the monospace
	fmt.Println("add project to monospace gitignore and set it as external")
	err = MonospaceAddProjectToGitignore(subDir)
	if err != nil {
		return err
	}
	if subRepoUrl != "" {
		err = MonospaceAddProject(subDir, subRepoUrl)
	} else {
		err = MonospaceAddProject(subDir, "local")
	}

	// stage modified files
	gitExec("add", ".gitignore", ".monospace.yml")

	fmt.Println("You can review the changes before committing")

	return

}
