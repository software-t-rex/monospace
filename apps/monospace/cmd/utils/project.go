package utils

import (
	"errors"
	"fmt"
	"monospace/monospace/cmd/colors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

type ProjectKind int

const (
	Local    ProjectKind = iota
	Internal ProjectKind = iota
	External ProjectKind = iota
)

func (p ProjectKind) String() string {
	return [...]string{"local", "internal", "external"}[p]
}
func (p ProjectKind) EnumIndex() int {
	return int(p)
}

// const ProjectKind = Struct{
// 	local: KIND_LOCAL,
// 	internal: KIND_INTERNAL,
// 	external: KIND_EXTERNAL
// }

// func (p ProjectKind) Valid() bool {
// 	switch p {
// 	case "internal", "external", "local":
// 		return true
// 	default:
// 		return false
// 	}
// }

type Project struct {
	// Name is relative path of the project in the monospace
	Name string
	// can be either empty "", "detached" or a git repository url
	RepoUrl string
	// wether the projet is managed by the monospace repository or not
	Kind ProjectKind
}

var styles = map[ProjectKind](func(s string) string){
	Internal: colors.Style(colors.Green),
	External: colors.Style(colors.Blue),
	Local:    colors.Style(colors.Red),
}

func (p Project) StyledString() string {
	return styles[p.Kind](p.Name)
}

var projectsMapCache map[string]string

func refreshProjectsMap() map[string]string {
	projectsMapCache = viper.GetViper().GetStringMapString("projects")
	return projectsMapCache
}
func getCachedProjectsMap() map[string]string {
	if len(projectsMapCache) == 0 {
		refreshProjectsMap()
	}
	return projectsMapCache
}

/* check project name is valid */
func ProjectIsValidName(name string) bool {
	// check if the project name is containing only letters, numbers, and underscores, slashes and hyphens
	match, _ := regexp.MatchString("^[a-zA-Z_][a-zA-Z0-9_-]*(\\/[a-zA-Z_][a-zA-Z0-9_-]*)*$", name)
	return match
}

/* check project is a monospace passenger */
func ProjectExists(name string) (exists bool) {
	projects := getCachedProjectsMap()
	_, exists = projects[name]
	return exists
}

/* return all project names declared in the .monospace.yml */
func ProjectsGetAllNameOnly() (res []string) {
	projectsMap := getCachedProjectsMap()
	res = MapGetKeys(projectsMap)
	sort.Strings(res)
	return
}

func ProjectsGetAll() []Project {
	projectsMap := getCachedProjectsMap()
	var res []Project
	for name, repo := range projectsMap {
		res = append(res, ProjectAsStruct(name, repo))
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return res
}

/* return all project names declared in the .monospace.yml that match the given prefix */
func ProjectsGetByPrefix(prefix string, noPrefix bool) (res []string) {
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	projects := ProjectsGetAllNameOnly()
	hasPrefix := PrefixPredicate(prefix)
	if !noPrefix {
		res = Filter(projects, hasPrefix)
	} else {
		res = MapAndFilter(projects, func(p string) (string, bool) {
			if hasPrefix(p) {
				p, _ = strings.CutPrefix(p, prefix)
				return p, true
			}
			return "", false
		})
	}
	return
}

func ProjectAsStruct(name string, repoUrl string) (res Project) {
	res.Name = name
	res.RepoUrl = repoUrl
	if repoUrl == "" || repoUrl == "internal" {
		res.Kind = Internal
	} else if repoUrl == "local" {
		res.Kind = Local
	} else {
		res.Kind = External
	}
	return
}

func ProjectGetByName(name string) (Project, error) {
	projects := getCachedProjectsMap()
	repoUrl, exists := projects[name]
	var err error
	if !exists {
		err = errors.New("Unknown project '" + name + "'")
	}
	return ProjectAsStruct(name, repoUrl), err
}

func ProjectGetPath(projectName string) string {
	return filepath.Join(MonospaceGetRoot(), filepath.Clean(projectName))
}

func ProjectCreateDirectory(projectName string) error {
	return os.MkdirAll(ProjectGetPath(projectName), 0750)
}

func ProjectChdir(projectName string) error {
	return os.Chdir(ProjectGetPath(projectName))
}

/* exit on error */
func ProjectCreate(projectName string, repoUrl string, skipPmTasks bool) {
	name := projectName
	skipPM := skipPmTasks

	if ProjectExists(name) {
		Exit("project already exists")
	} else if !ProjectIsValidName(name) {
		Exit("invalid project name")
	}

	project := Project{
		Name:    name,
		RepoUrl: repoUrl,
		Kind:    Internal,
	}

	if project.RepoUrl == Local.String() {
		project.Kind = Local
	} else if project.RepoUrl == Internal.String() {
		project.Kind = Internal
	} else if project.RepoUrl != "" {
		project.Kind = External
	}

	// set cwd to monospace directory
	CheckErr(MonospaceChdir())
	CheckErrWithMsg(ProjectCreateDirectory(project.Name), "Error while creating package")
	if project.Kind == External {
		CheckErr(MonospaceAddProject(project.Name, project.RepoUrl))
	} else {
		CheckErr(MonospaceAddProject(project.Name, project.Kind.String()))
	}

	// move to new package directory
	CheckErr(ProjectChdir(project.Name))
	if !skipPM {
		fmt.Println("Initialize package manager")
		CheckErr(PMinit())
	}
	if project.Kind == Local {
		fmt.Println("Initialize local repository")
		CheckErr(MonospaceInitRepo(project.Name))
		CheckErr(WriteTemplateGitinore("./"))
	} else if project.Kind == Internal {
		CheckErr(WriteTemplateGitinore("./"))
	} else if project.Kind == External {
		CheckErr(MonospaceCloneRepo(project.Name, project.RepoUrl))
	}

	fmt.Println(colors.Success("project successfully added to your monospace"))
}

/* exit on error */
func ProjectRemove(projectName string, rmdir bool, withConfirm bool) {
	project, err := ProjectGetByName(projectName)
	CheckErr(err)
	fmt.Println("Remove from monospace.yml")
	projects := viper.GetStringMap("projects")
	projects[projectName] = nil
	viper.Set("projects", projects)
	err = viper.WriteConfig()
	CheckErr(err)

	rootDir := MonospaceGetRoot()
	printSuccess := func() { fmt.Println(colors.Success("Project " + projectName + " successfully removed")) }

	if project.Kind != Internal {
		fmt.Println("Remove from gitignore")
		err = FileRemoveLine(filepath.Join(rootDir, ".gitignore"), projectName)
		CheckErr(err)
	}

	if !rmdir {
		printSuccess()
		fmt.Println("You should now delete the project directory.")
	} else {
		if !withConfirm || Confirm("Do you want to delete "+project.Name, false) {
			CheckErr(RmDir(filepath.Join(rootDir, project.Name)))
		}
		printSuccess()
	}
}
