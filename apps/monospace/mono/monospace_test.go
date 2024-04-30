package mono

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/software-t-rex/monospace/app"
	"gotest.tools/v3/assert"
)

func TestSpaceGetRoot(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NilError(t, err, "os.Getwd failed")
	assert.Equal(t, SpaceGetRoot(), strings.TrimSuffix(cwd, "/apps/monospace/mono"))
}

func TestSpaceGetRootForPath(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NilError(t, err, "os.Getwd failed")
	assert.Equal(t, SpaceGetRootForPath(cwd), strings.TrimSuffix(cwd, "/apps/monospace/mono"))
}

func TestSpaceGetConfigPath(t *testing.T) {
	expectedPath := filepath.Join(SpaceGetRoot(), "/.monospace/monospace.yml")
	assert.Equal(t, SpaceGetConfigPath(), expectedPath, "SpaceGetConfigPath did not return the expected path")
}

func TestSpaceHasProject(t *testing.T) {
	if !app.ConfigIsLoaded() {
		app.ConfigLoad(SpaceGetConfigPath())
	}
	assert.Assert(t, SpaceHasProject("apps/monospace"))
	assert.Assert(t, !SpaceHasProject("apps/monospace/docs"), "SpaceHasProject should not return true for a subdirectory of a project")
	assert.Assert(t, SpaceHasProject("gomodules/ui"))
	assert.Assert(t, !SpaceHasProject("ui"), "SpaceHasProject should not return true for project aliases")
}

func TestSpaceChdir(t *testing.T) {
	oldCwd, err := os.Getwd()
	assert.NilError(t, err, "os.Getwd failed")
	SpaceChdir()
	cwd, err := os.Getwd()
	assert.NilError(t, err, "os.Getwd failed")
	assert.Assert(t, SpaceGetRoot() == cwd, "SpaceChdir did not change the current working directory to the monospace root")
	os.Chdir(oldCwd)
}
