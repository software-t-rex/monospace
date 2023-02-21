package utils

import (
	"errors"
	"fmt"
	"monospace/parallel"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const DfltcfgFileName string = ".monospace.yml"

var monospaceRoot string = ""

func MonospaceGetRoot() string {
	if monospaceRoot != "" {
		return monospaceRoot
	}
	path, err := os.Getwd()
	path = filepath.ToSlash(path)
	foundConfigDir := false
	CheckErr(err)

	for path != "" && path != "." {
		configExists, _ := FileExists(filepath.Join(path, "/", DfltcfgFileName))
		if configExists {
			foundConfigDir = true
			break
		}
		// go up one dir
		path = filepath.Clean(filepath.Join(path, "/../"))
		if path == "/home" || path == "." || path == "/" {
			break
		}
	}
	if !foundConfigDir {
		return ""
	}
	monospaceRoot = path
	return path
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
	fmt.Println("Cloning root repository...")
	err := GitClone(repoUrl, destDirectory)
	CheckErr(err)
	fmt.Println(Green("Cloning done."))
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
		var cmds [][]string
		for _, project := range externals {
			cmds = append(cmds, []string{"git", "clone", project.RepoUrl, project.Name})
		}
		fmt.Println("Cloning externals projects...")
		errs := MapAndFilter(
			parallel.Exec(cmds...),
			func(val error) (string, bool) {
				if val == nil {
					return "", false
				}
				return val.Error(), true
			},
		)
		fmt.Println(Green("Cloning done."))
		if len(errs) > 0 {
			fmt.Println(Red("Terminated with errors" + strings.Join(errs, "\n")))
		} else {
			fmt.Println(Green("Terminated with success"))
		}

	}

}

func MonospaceInitRepo(projectName string) (err error) {
	err = MonospaceAddProjectToGitignore(projectName)
	if err == nil {
		projectPath := filepath.Join(MonospaceGetRoot(), "/", projectName)
		err = GitInit(projectPath)
	}
	return
}
