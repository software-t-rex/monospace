package utils

import (
	"errors"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
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
		if FileExists(filepath.Join(path, "/", DfltcfgFileName)) {
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

func MonospaceCloneRepo(projectName string, repoUrl string) (err error) {
	err = MonospaceAddProjectToGitignore(projectName)
	if err == nil {
		projectPath := filepath.Join(MonospaceGetRoot(), "/", projectName)
		_, err = git.PlainClone(projectPath, false, &git.CloneOptions{
			URL:      repoUrl,
			Progress: os.Stdout,
		})
	}
	return
}

func MonospaceInitRepo(projectName string) (err error) {
	err = MonospaceAddProjectToGitignore(projectName)
	if err == nil {
		projectPath := filepath.Join(MonospaceGetRoot(), "/", projectName)
		_, err = git.PlainInit(projectPath, false)
	}
	return
}
