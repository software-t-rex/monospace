package mono

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/software-t-rex/go-jobExecutor/v2"
	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/git"
	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
)

var monospaceRoot string = ""

func SpaceGetRoot() string {
	if monospaceRoot != "" {
		return monospaceRoot
	}
	return SpaceGetRootNoCache()
}

// update internal cache or the monospace root and returns it
// an empty strings signifies that root could not be found
func SpaceGetRootNoCache() string {
	path, err := os.Getwd()
	utils.CheckErr(err)
	monospaceRoot = SpaceGetRootForPath(path)
	return monospaceRoot
}

func SpaceGetRootForPath(absPath string) string {
	absPath = filepath.ToSlash(absPath)
	for absPath != "" && absPath != "." {
		if utils.FileExistsNoErr(filepath.Join(absPath, app.DfltcfgFilePath)) {
			return absPath
		}
		// go up one dir
		lastVisted := absPath
		absPath = filepath.Clean(filepath.Join(absPath, "../"))
		if lastVisted == absPath { // we are at the root
			break
		}
	}

	return ""
}

// change the current working directory to the monospace root
func SpaceChdir() error {
	root := SpaceGetRoot()
	if root == "" {
		return errors.New("can't find monospace root dir")
	}
	return os.Chdir(root)
}

func SpaceGetConfigPath() string {
	return filepath.Join(SpaceGetRoot(), "/", app.DfltcfgFilePath)
}

func SpaceHasProject(projectName string) bool {
	config, _ := app.ConfigGet()
	if config != nil {
		projects := config.Projects
		repo, ok := projects[projectName]
		if ok && repo != "" {
			return true
		}
	}
	return false
}

func SpaceAddProjectToGitignore(projectName string) error {
	return utils.FileAppend(filepath.Join(SpaceGetRoot(), "/.gitignore"), projectName)
}

/* exit on error */
func SpaceClone(destDirectory string, repoUrl string) {
	if utils.CheckErrOrReturn(utils.FileExists(destDirectory)) {
		utils.Exit("path already exists")
	}
	destDirectory, err := filepath.Abs(destDirectory)
	if err != nil {
		utils.Exit(err.Error())
	}
	theme := ui.GetTheme()
	fmt.Println(theme.Info("Cloning root repository..."))
	err = git.Clone(repoUrl, destDirectory)
	utils.CheckErr(err)
	fmt.Println(theme.Success("Cloning done."))
	// move to the monorepo root
	monospaceRoot = destDirectory
	utils.CheckErr(os.Chdir(destDirectory))
	if !utils.FileExistsNoErr(app.DfltcfgFilePath) {
		fmt.Println(`This doesn't seem to be a monospace project yet
To turn the cloned repository into a monospace you can run this command:
cd ` + destDirectory + ` && monospace init`)
	}

	// if githooks were installed set git hooks path to .monospace/githooks
	if utils.FileExistsNoErr(filepath.Join(monospaceRoot, app.DfltHooksDir)) {
		fmt.Printf(theme.Info("found githooks directory, set git core.hookspath to %s\n"), app.DfltHooksDir)
		utils.CheckErr(git.HooksPathSet(monospaceRoot, app.DfltHooksDir))
	}

	// read the config file
	config := utils.CheckErrOrReturn(app.ConfigRead(app.DfltcfgFilePath))
	if config.Projects == nil || len(config.Projects) < 1 {
		fmt.Println("No external projects found in the config file")
		fmt.Println(theme.Success("Terminated with success"))
		return
	}
	// proceed to clone external projects
	projects := ProjectsAsStructs(config.Projects)
	externals := utils.SliceFilter(projects, func(p Project) bool { return p.Kind == External })
	if len(externals) == 0 {
		fmt.Println(theme.Success("Terminated with success"))
		return
	}
	fmt.Println(theme.Info("Cloning externals projects..."))
	jobExecutor := jobExecutor.NewExecutor().WithOngoingStatusOutput()
	for _, project := range externals {
		jobExecutor.AddNamedJobCmd("clone "+project.Name, exec.Command("git", "clone", project.RepoUrl, project.Path()))
	}
	errs := jobExecutor.Execute()
	fmt.Println(theme.Success("Cloning done"))
	if len(errs) > 0 {
		fmt.Println(theme.Error("Terminated with errors" + errs.String()))
	} else {
		fmt.Println(theme.Success("Terminated with success"))
	}
}

func SpaceInitRepo(projectName string) (err error) {
	err = SpaceAddProjectToGitignore(projectName)
	if err == nil {
		projectPath := ProjectGetPath(projectName)
		hasGitIgnore := utils.FileExistsNoErr(filepath.Join(projectPath, ".gitignore"))
		err = git.Init(projectPath, !hasGitIgnore)
	}
	return
}
