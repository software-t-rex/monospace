package mono

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/git"
	"github.com/software-t-rex/monospace/gomodules/scaffolders"
	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
	"github.com/software-t-rex/packageJson"
)

var RootProject Project

func init() {
	RootProject = Project{
		Name:    "root",
		RepoUrl: "internal",
		Kind:    Root,
	}
}

type ProjectKind int

const (
	Root ProjectKind = iota - 1
	Local
	Internal
	External
)

func (p ProjectKind) String() string {
	if p == -1 {
		return "root"
	}
	return [...]string{"local", "internal", "external"}[p]
}
func (p ProjectKind) EnumIndex() int {
	return int(p)
}

type Project struct {
	// Name is relative path of the project in the monospace
	Name string
	// can be either empty "", "detached" or a git repository url (can it still be detached ?)
	RepoUrl string
	// wether the projet is managed by the monospace repository or not
	Kind ProjectKind
	// which type of project this is (eg: jg, go ...)
}

var styles = map[ProjectKind](func(s ...string) string){
	Root:     ui.NewStyler(ui.Yellow.Foreground()),
	Internal: ui.NewStyler(ui.Green.Foreground()),
	External: ui.NewStyler(ui.Blue.Foreground()),
	Local:    ui.NewStyler(ui.Red.Foreground()),
}

func (p Project) String() string {
	return p.Name
}
func (p Project) StyledString() string {
	kind := p.Kind
	return styles[kind](p.Name)
}

func (p Project) Path() string {
	return ProjectGetPath(p.Name)
}

func (p Project) IsGit() bool {
	if p.IsInternal() || p.RepoUrl == "" {
		return false
	}
	return git.IsRepoRootDir(p.Path())
}
func (p Project) IsRoot() bool     { return p.Kind == Root }
func (p Project) IsInternal() bool { return p.Kind == Internal }
func (p Project) IsExternal() bool { return p.Kind == External }
func (p Project) IsLocal() bool    { return p.Kind == Local }
func (p Project) HasPackageJson() bool {
	return utils.FileExistsNoErr(filepath.Join(p.Path(), "package.json"))
}
func (p Project) GetPackageJson() (*packageJson.PackageJSON, error) {
	return packageJson.Read(filepath.Join(p.Path(), "package.json"))
}

func getProjectsMap() (map[string]string, error) {
	config, err := app.ConfigGet()
	if err != nil {
		return nil, err
	}
	return config.Projects, nil
}

/* check project name is valid */
func ProjectIsValidName(name string) bool {
	if name == "root" { // reserved name
		return false
	}
	// check if the project name is containing only letters, numbers, underscores, slashes and hyphens
	match, _ := regexp.MatchString("^[a-zA-Z_][a-zA-Z0-9_.-]*(\\/[a-zA-Z_][a-zA-Z0-9_.-]*)*$", name)
	return match
}

/* check project is a monospace passenger */
func ProjectExists(name string) (exists bool) {
	projects, _ := getProjectsMap()
	if projects != nil {
		_, exists = projects[name]
	}
	return exists
}

// return project aliases
func ProjectsGetAliasesNameOnly() (res []string) {
	config, _ := app.ConfigGet()
	if config.Aliases != nil {
		return utils.MapGetKeys(config.Aliases)
	}
	return []string{}
}

/* return all project names declared in the monospace.yml */
func ProjectsGetAllNameOnly() (res []string) {
	projectsMap, _ := getProjectsMap()
	if projectsMap != nil {
		res = utils.MapGetKeys(projectsMap)
		sort.Strings(res)
	}
	return
}

func ProjectsAsStructs(projectsMap map[string]string) []Project {
	var res []Project
	for name, repo := range projectsMap {
		res = append(res, ProjectAsStruct(name, repo))
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return res
}

func ProjectsGetAll() []Project {
	projectsMap, _ := getProjectsMap()
	return ProjectsAsStructs(projectsMap)
}

/* return all project names declared in the monospace.yml that match the given prefix */
func ProjectsGetByPrefix(prefix string, noPrefix bool) (res []string) {
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	projects := ProjectsGetAllNameOnly()
	hasPrefix := utils.PrefixPredicate(prefix)
	if !noPrefix {
		res = utils.SliceFilter(projects, hasPrefix)
	} else {
		res = utils.SliceMapAndFilter(projects, func(p string) (string, bool) {
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
	// determine project Kind
	if repoUrl == "" || repoUrl == "internal" {
		res.Kind = Internal
	} else if repoUrl == "local" {
		res.Kind = Local
	} else {
		res.Kind = External
	}
	return
}

func ProjectDetectMainLang(name string) string {
	projectPath := ProjectGetPath(name)
	if utils.FileExistsNoErr(filepath.Join(projectPath, "go.mod")) {
		return "golang"
	} else if utils.FileExistsNoErr(filepath.Join(projectPath, "package.json")) {
		return "js"
	}
	return "unknown"
}

func ProjectGetByAlias(alias string) (Project, error) {
	config, err := app.ConfigGet()
	if err != nil {
		return Project{}, err
	}
	if config.Aliases == nil || config.Aliases[alias] == "" {
		return Project{}, errors.New("alias not found: " + alias)
	}
	return ProjectGetByName(config.Aliases[alias])
}

func ProjectGetByName(name string) (Project, error) {
	config, err := app.ConfigGet()
	var p Project
	if err != nil {
		return p, err
	}

	// check alias first
	if !strings.Contains(name, "/") && config.Aliases != nil && config.Aliases[name] != "" {
		name = config.Aliases[name]
	}
	projects := config.Projects

	repoUrl, exists := projects[name]
	if !exists {
		err = errors.New("Unknown project '" + name + "'")
	}
	return ProjectAsStruct(name, repoUrl), err
}

func ProjectGetPath(projectName string) string {
	if projectName == "root" {
		return SpaceGetRoot()
	}
	return filepath.Join(SpaceGetRoot(), filepath.Clean(projectName))
}

/* exit on error */
func ProjectCreate(projectName string, repoUrl string, projectType string) {
	// check name is Ok
	if ProjectExists(projectName) {
		utils.Exit("project already exists")
	} else if !ProjectIsValidName(projectName) {
		utils.Exit("invalid project name")
	}

	projectPath := ProjectGetPath(projectName)
	dirExists, err := utils.IsDir(projectPath)
	utils.CheckErr(err)

	project := Project{
		Name:    projectName,
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

	// set some env variables
	utils.CheckErr(app.PopulateEnv(map[string]string{"PROJECT_PATH": projectName}))

	// set cwd to monospace directory
	utils.CheckErr(SpaceChdir())

	// create dir if not exists
	if !dirExists {
		utils.CheckErrWithMsg(os.MkdirAll(project.Path(), 0750), "Error while creating project directory")
	}

	// add to .monopace.yml
	if project.Kind == External {
		utils.CheckErr(app.ConfigAddProject(project.Name, project.RepoUrl, true))
	} else {
		utils.CheckErr(app.ConfigAddProject(project.Name, project.Kind.String(), true))
	}

	// move to new package directory
	utils.CheckErr(os.Chdir(projectPath))

	// strategy for different kind of projects
	switch project.Kind {
	case Local:
		fmt.Println("Initialize local repository")
		utils.CheckErr(SpaceInitRepo(project.Name)) // will add gitignore
	case Internal:
		utils.CheckErr(git.AddGitIgnoreFile())
	case External:
		fmt.Println("Clone repository")
		utils.CheckErr(ProjectCloneRepo(project))
	default:
		utils.Exit("unknown project kind must be local, internal or external")
	}

	// no need to init projects for external projects
	if projectType != "" && project.Kind != External {
		switch projectType {
		case "js":
			fmt.Println("Initialize package manager")
			utils.CheckErr(scaffolders.Javascript())
		case "go":
			fmt.Println("Initialize go module")
			utils.CheckErr(scaffolders.Golang())
		default:
			utils.PrintWarning("Unknown project type '" + projectType + "' => ignored")
		}
	}

	fmt.Println(ui.GetTheme().Success("project successfully added to your monospace"))
}

func ProjectCloneRepo(project Project) (err error) {
	err = SpaceAddProjectToGitignore(project.Name)
	if err == nil {
		err = git.Clone(project.RepoUrl, project.Path())
	}
	return
}

func ProjectRemoveFromGitignore(project Project, silent bool) (err error) {
	if project.Kind != Internal {
		if !silent {
			fmt.Println("Remove from gitignore")
		}
		err = utils.FileRemoveLine(filepath.Join(SpaceGetRoot(), ".gitignore"), project.Name)
	}
	return
}

/* exit on error */
func ProjectRemove(projectName string, rmdir bool, withConfirm bool) {
	project := utils.CheckErrOrReturn(ProjectGetByName(projectName))
	utils.CheckErr(app.ConfigRemoveProject(project.Name, true))

	printSuccess := func() { fmt.Println(ui.GetTheme().Success("Project " + projectName + " successfully removed")) }

	utils.CheckErr(ProjectRemoveFromGitignore(project, false))

	if !rmdir {
		printSuccess()
		fmt.Println("You should now delete the project directory.")
	} else {
		if !withConfirm || ui.ConfirmInline("Do you want to delete "+project.Name, false) {
			utils.CheckErr(utils.RmDir(project.Path()))
		}
		printSuccess()
	}
}
