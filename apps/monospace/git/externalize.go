package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/utils"
)

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
