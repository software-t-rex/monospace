package mono

import (
	"path/filepath"
	"testing"

	"github.com/software-t-rex/monospace/app"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestProjectGetPath(t *testing.T) {
	projectName := "apps/monospace"
	expectedPath := filepath.Join(SpaceGetRoot(), projectName)
	assert.Equal(t, ProjectGetPath(projectName), expectedPath, "ProjectGetPath did not return the expected path")
}

func TestProjectExists(t *testing.T) {
	// require config to be loaded
	if !app.ConfigIsLoaded() {
		app.ConfigLoad(SpaceGetConfigPath())
	}
	projectName := "apps/monospace"
	assert.Assert(t, ProjectExists(projectName), "ProjectExists should return true for an existing project")
	assert.Assert(t, !ProjectExists("nonexistent"), "ProjectExists should return false for a nonexistent project")
}

func TestProjectIsValidName(t *testing.T) {
	// require config to be loaded
	if !app.ConfigIsLoaded() {
		app.ConfigLoad(SpaceGetConfigPath())
	}
	testCases := []struct {
		projectName string
		want        bool
	}{
		{"apps/monospace", true},
		// tests for first character
		{"1project", false},
		{"/project", false},
		{"_project", true},
		{".project", false},
		{"$project", false},
		{"#project", false},
		// tests for first character after a slash
		{"apps/project", true},
		{"apps/_project", true},
		{"apps/1project", false},
		{"apps/.project", false},
		{"apps/@project", false},
		{"apps/#project", false},
		// should not have a final slash
		{"apps/monospace/", false},
		{"pro.ject", true},
	}
	for _, tc := range testCases {
		got := ProjectIsValidName(tc.projectName)
		if got != tc.want {
			t.Errorf("ProjectIsValidName(%q) = %v; want %v", tc.projectName, got, tc.want)
		}
	}
}

func TestProjectsGetAll(t *testing.T) {
	// require config to be loaded
	if !app.ConfigIsLoaded() {
		app.ConfigLoad(SpaceGetConfigPath())
	}
	attendedProjects := []string{
		"apps/monospace",
		"gomodules/go-jobexecutor",
		"gomodules/js-packagemanager",
		"gomodules/packageJson",
		"gomodules/scaffolders",
		"gomodules/ui",
		"gomodules/utils",
		"packages/monospace",
	}
	projects := ProjectsGetAll()
	foundProjects := make(map[string]bool, len(projects))
	for _, project := range projects {
		foundProjects[project.Name] = true
	}
	for _, project := range attendedProjects {
		if !foundProjects[project] {
			t.Errorf("ProjectsGetAll() did not return the project %q", project)
		}
	}
}

func TestProjectsGetAllNameOnly(t *testing.T) {
	// require config to be loaded
	if !app.ConfigIsLoaded() {
		app.ConfigLoad(SpaceGetConfigPath())
	}
	want := []string{
		"apps/monospace",
		"gomodules/go-jobexecutor",
		"gomodules/js-packagemanager",
		"gomodules/packageJson",
		"gomodules/scaffolders",
		"gomodules/ui",
		"gomodules/utils",
		"packages/monospace",
	}
	projects := ProjectsGetAllNameOnly()
	assert.Equal(t, len(projects), len(want), "ProjectsGetAllNameOnly() did not return the expected number of projects")
	assert.Assert(t, cmp.DeepEqual(projects, want), "ProjectsGetAllNameOnly() did not return the expected projects")
}

func TestProjectsGetAliasesNameOnly(t *testing.T) {
	// require config to be loaded
	if !app.ConfigIsLoaded() {
		app.ConfigLoad(SpaceGetConfigPath())
	}
	want := []string{
		"executor",
		"jspm",
		"monospace",
		"npmpkg",
		"packageJson",
		"ui",
		"utils",
	}
	aliases := ProjectsGetAliasesNameOnly()
	assert.Equal(t, len(aliases), len(want), "ProjectsGetAllNameOnly() did not return the expected number of projects")
	assert.Assert(t, cmp.DeepEqual(aliases, want), "ProjectsGetAllNameOnly() did not return the expected projects")
}

func TestProjectAsStruct(t *testing.T) {
	// require config to be loaded
	projectName := "apps/fakeProject"
	projectPath := filepath.Join(SpaceGetRoot(), projectName)
	project := ProjectAsStruct(projectName, "")
	assert.Equal(t, project.Name, projectName, "ProjectAsStruct did not return the expected project name")
	assert.Equal(t, projectPath, project.Path(), "ProjectAsStruct did not return the expected project path")
	assert.Equal(t, project.Kind, Internal, "ProjectAsStruct did not return the expected project type")
	assert.Assert(t, project.IsInternal(), "ProjectAsStruct did not return the expected project type")
}
