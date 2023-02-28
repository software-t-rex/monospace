package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/software-t-rex/go-jobExecutor"
	"github.com/spf13/viper"
)

const DfltcfgFileName string = ".monospace.yml"

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
		if FileExistsNoErr(filepath.Join(absPath, "/", DfltcfgFileName)) {
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
	return filepath.Join(MonospaceGetRoot(), "/", DfltcfgFileName)
}

func MonospaceHasProject(projectName string) (ok bool) {
	projects := viper.GetStringMap("projects")
	_, ok = projects[projectName]
	return
}

func MonospaceAddProject(projectName string, repoUrl string) error {
	projects := viper.GetStringMap("projects")
	projects[projectName] = repoUrl
	viper.Set("projects", projects)
	return viper.WriteConfig()
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
	os.Chdir(destDirectory)
	// read the config file
	if CheckErrOrReturn(FileExists(DfltcfgFileName)) {
		// proceed to clone external projects
		configParser := viper.New()
		configParser.SetConfigFile(DfltcfgFileName)
		CheckErr(configParser.ReadInConfig())
		projects := ProjectsAsStructs(configParser.GetStringMapString("projects"))
		externals := Filter(projects, func(p Project) bool { return p.Kind == External })
		if len(externals) == 0 {
			fmt.Println(Success("Terminated with success"))
			return
		}
		fmt.Println(Info("Cloning externals projects..."))
		jobExecutor := jobExecutor.NewExecutor().WithProgressOutput()
		for _, project := range externals {
			jobExecutor.AddJobCmd("git", "clone", project.RepoUrl, project.Name)
		}
		errs := jobExecutor.Execute()
		fmt.Println(Success("Cloning done."))
		if len(errs) > 0 {
			fmt.Println(ErrorStyle("Terminated with errors" + errs.String()))
		} else {
			fmt.Println(Success("Terminated with success"))
		}
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
