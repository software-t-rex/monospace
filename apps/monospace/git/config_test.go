package git

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestConfig(t *testing.T) {
	// setup
	tmpdir := t.TempDir()
	assert.NilError(t, Init(tmpdir, false), "Failed to init git repo")

	runStep := func(t *testing.T, name string, fn func(t *testing.T)) {
		if !t.Run(name, fn) {
			t.FailNow()
		}
	}

	runStep(t, "Should be able to get and set a config setting", func(t *testing.T) {
		// get
		got, err := ConfigGet(tmpdir, "core.bare")
		assert.NilError(t, err, "Failed to get config")
		assert.Equal(t, got, "false", "Config value is not the expected one")
		// set
		err = ConfigSet(tmpdir, "core.pager", "more")
		assert.NilError(t, err, "Failed to set pager config")
		// get
		got, err = ConfigGet(tmpdir, "core.pager")
		assert.NilError(t, err, "Failed to get pager config")
		assert.Equal(t, got, "more", "Config pager value is not the expected one")
	})

	runStep(t, "Should be able to get and set hooks dir", func(t *testing.T) {
		// set
		err := HooksPathSet(tmpdir, "test/hooks")
		assert.NilError(t, err, "Failed to set hooks dir")
		// get
		got, err := HooksPathGet(tmpdir)
		assert.NilError(t, err, "Failed to get hooks dir")
		assert.Equal(t, got, "test/hooks", "Hooks dir is not the expected one")
	})

}
