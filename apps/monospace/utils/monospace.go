package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/software-t-rex/go-jobExecutor/v2"
	"github.com/software-t-rex/monospace/app"
)

var monospaceRoot string = ""

func MonospaceGetRoot() string {
	if monospaceRoot != "" {
		return monospaceRoot
	}
	return MonospaceGetRootNoCache()
}

func MonospaceGetRootNoCache() string {
	path, err := os.Getwd()
	CheckErr(err)
	monospaceRoot = MonospaceGetRootForPath(path)
	return monospaceRoot
}

func MonospaceGetRootForPath(absPath string) string {
	absPath = filepath.ToSlash(absPath)

	// @todo check this work on windows before release to public
	for absPath != "" && absPath != "." {
		if FileExistsNoErr(filepath.Join(absPath, "/", app.DfltcfgFilePath)) {
			return absPath
		}
		// go up one dir
		absPath = filepath.Clean(filepath.Join(absPath, "../"))
		if absPath == "/home" || absPath == "." || absPath == "/" {
			break
		}
	}

	return ""
}

func MonospaceChdir() error {
	root := MonospaceGetRoot()
	if root == "" {
		return errors.New("can't find monospace root dir")
	}
	return os.Chdir(root)
}

func MonospaceGetConfigPath() string {
	return filepath.Join(MonospaceGetRoot(), "/", app.DfltcfgFilePath)
}

func MonospaceHasProject(projectName string) bool {
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

func MonospaceAddProjectToGitignore(projectName string) error {
	return FileAppend(filepath.Join(MonospaceGetRoot(), "/.gitignore"), projectName)
}

/* exit on error */
func MonospaceClone(destDirectory string, repoUrl string) {
	if CheckErrOrReturn(FileExists(destDirectory)) {
		PrintError(errors.New("path already exists"))
		os.Exit(1)
	}
	fmt.Println(Info("Cloning root repository..."))
	err := GitClone(repoUrl, destDirectory)
	CheckErr(err)
	fmt.Println(Success("Cloning done."))
	// move to the monorepo root
	monospaceRoot = destDirectory
	CheckErr(os.Chdir(destDirectory))
	if !FileExistsNoErr(app.DfltcfgFilePath) {
		fmt.Println(`This doesn't seem to be a monospace project yet
To turn the cloned repository into a monospace you can run this command:
cd ` + destDirectory + ` && monospace init`)
	}

	// read the config file
	config := CheckErrOrReturn(app.ConfigRead(app.DfltcfgFilePath))
	if config.Projects == nil || len(config.Projects) < 1 {
		fmt.Println("No external projects found in the config file")
		fmt.Println(Success("Terminated with success"))
		return
	}
	// proceed to clone external projects
	projects := ProjectsAsStructs(config.Projects)
	externals := SliceFilter(projects, func(p Project) bool { return p.Kind == External })
	if len(externals) == 0 {
		fmt.Println(Success("Terminated with success"))
		return
	}
	fmt.Println(Info("Cloning externals projects..."))
	jobExecutor := jobExecutor.NewExecutor().WithOngoingStatusOutput()
	for _, project := range externals {
		jobExecutor.AddNamedJobCmd("clone "+project.Name, exec.Command("git", "clone", project.RepoUrl, project.Name))
	}
	errs := jobExecutor.Execute()
	fmt.Println(Success("Cloning done"))
	if len(errs) > 0 {
		fmt.Println(ErrorStyle("Terminated with errors" + errs.String()))
	} else {
		fmt.Println(Success("Terminated with success"))
	}
}

func MonospaceInitRepo(projectName string) (err error) {
	err = MonospaceAddProjectToGitignore(projectName)
	if err == nil {
		projectPath := filepath.Join(MonospaceGetRoot(), "/", projectName)
		hasGitIgnore := FileExistsNoErr(filepath.Join(projectPath, ".gitignore"))
		err = GitInit(projectPath, !hasGitIgnore)
	}
	return
}
