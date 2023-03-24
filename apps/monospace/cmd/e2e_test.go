package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

func dirname() string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		fmt.Fprintln(os.Stderr, "unable to get the current filename and to run tests")
		os.Exit(1)
	}
	return filepath.Dir(filename)
}

func runStep(t *testing.T, name string, fn func(t *testing.T)) {
	if !t.Run(name, fn) {
		t.FailNow()
	}
}

func TestCmd_Suite(t *testing.T) {
	// SETUP TEST SUITE
	type testPathCase struct {
		name    string
		args    []string
		want    icmd.Expected
		pathOps fs.PathOp
	}
	type testMatchCase struct {
		name  string
		args  []string
		want  icmd.Expected
		match string
	}

	// prepare binary to test
	monospaceDir := filepath.Join(dirname(), "../")
	tmpDir := fs.NewDir(t, "mstest-build")
	monospaceBin := filepath.Join(tmpDir.Path(), "monospace-test")
	icmd.RunCmd(icmd.Command(
		"go", "build", "-cover", "-o", monospaceBin, "main.go",
	), icmd.Dir(monospaceDir)).
		Assert(t, icmd.Expected{ExitCode: 0})

		// prepare some dirs to work with
	initDir := fs.NewDir(t, "mstest-init")
	initDirOp := icmd.Dir(initDir.Path())
	cloneDir := fs.NewDir(t, "mstest-clone")

	// test suite helpers methods
	runMonospace := func(args []string, ops ...icmd.CmdOp) *icmd.Result {
		cmd := icmd.Command(monospaceBin, args...)
		return icmd.RunCmd(cmd, append(ops, icmd.WithEnv(append(os.Environ(), "GOCOVERDIR="+tmpDir.Path())...))...)
	}
	hasFile := func(name string) fs.PathOp { return fs.WithFile(name, "", fs.MatchAnyFileContent, fs.MatchAnyFileMode) }
	hasDir := func(name string, extraFiles bool, ops ...fs.PathOp) fs.PathOp {
		ops = append(ops, fs.MatchAnyFileMode)
		if extraFiles {
			ops = append(ops, fs.MatchExtraFiles)
		}
		return fs.WithDir(name, ops...)
	}
	runPathCase := func(path string) func(*testing.T, []testPathCase) {
		dirOp := icmd.Dir(path)
		return func(t *testing.T, tests []testPathCase) {
			t.Helper()
			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					result := runMonospace(tc.args, dirOp)
					result.Assert(t, tc.want)
					if tc.pathOps != nil {
						assert.Assert(t, fs.Equal(path, fs.Expected(t, tc.pathOps, fs.MatchExtraFiles, fs.MatchAnyFileMode)))
					}
				})
			}
		}
	}

	// generate coverage reports
	t.Cleanup(func() {
		covfile := filepath.Join(monospaceDir, "coverage/binary/coverage.out")
		covdir := filepath.Dir(covfile)
		icmd.RunCmd(icmd.Command("go", "tool", "covdata", "textfmt", "-i="+tmpDir.Path(), "-o", covfile))
		icmd.RunCmd(icmd.Command("go", "tool", "cover", "-html="+covfile, "-o", filepath.Join(covdir, "coverage.html")), icmd.Dir(covdir))
		icmd.RunCmd(icmd.Command("go", "tool", "cover", "-func="+covfile, "-o", filepath.Join(covdir, "coverage.txt")), icmd.Dir(covdir))
	})

	runStep(t, "init -y", func(t *testing.T) {
		// cmd := icmd.Command(monospaceBin, "init", "--no-interactive")
		result := runMonospace([]string{"init", "--no-interactive"}, icmd.Dir(initDir.Path()))
		// result := icmd.RunCmd(cmd, icmd.Dir(initDir.Path()))
		result.Assert(t, icmd.Expected{ExitCode: 0})
		expected := fs.Expected(t,
			hasFile("package.json"),
			hasFile(".npmrc"),
			hasFile(".gitignore"),
			hasFile("go.work"),
			hasFile("pnpm-workspace.yaml"),
			hasDir(".git", true),
			hasDir(".monospace", false, hasFile("monospace.yml")),
		)
		assert.Assert(t, fs.Equal(initDir.Path(), expected))
		runMonospace([]string{"ls"}, icmd.Dir(initDir.Path())).Assert(t, icmd.Expected{
			Out: "No projects found start by adding one to your monospace.",
		})
	})

	runStep(t, "create", func(t *testing.T) {
		tests := []testPathCase{
			{"ls without project should display a message", []string{"ls"}, icmd.Expected{ExitCode: 0, Out: "No projects found start by adding one to your monospace."}, nil},
			{"with no args", []string{"create"}, icmd.Expected{ExitCode: 1}, nil},
			{"with 1 args", []string{"create", "local"}, icmd.Expected{ExitCode: 1}, nil},
			{"with 2 args but invalid kind arg", []string{"create", "invalid", "apps/myapp"}, icmd.Expected{ExitCode: 1}, nil},
			{"with invalid name", []string{"create", "internal", "apps!#/1myapp"}, icmd.Expected{ExitCode: 1}, nil},
			{"local", []string{"create", "local", "apps/myapp"}, icmd.Expected{ExitCode: 0},
				hasDir("apps", false,
					hasDir("myapp", false,
						hasDir(".git", true),
						hasFile(".gitignore"),
					),
				),
			},
			{"internal", []string{"create", "internal", "packages/mylib"}, icmd.Expected{ExitCode: 0},
				hasDir("packages", false, hasDir("mylib", false, hasFile(".gitignore"))),
			},
			{"internal go type", []string{"create", "internal", "packages/golib", "-t", "go"}, icmd.Expected{ExitCode: 0},
				hasDir("packages", true, hasDir("golib", false,
					hasFile(".gitignore"),
					hasFile("go.mod"),
					hasFile("main.go"),
				)),
			},
			{"internal js type", []string{"create", "internal", "packages/jslib", "-t", "js"}, icmd.Expected{ExitCode: 0},
				hasDir("packages", true, hasDir("jslib", false,
					hasFile(".gitignore"),
					hasFile("package.json"),
					hasFile("index.js"),
				)),
			},
			{"ls should list added projects sorted", []string{"ls", "-C"}, icmd.Expected{ExitCode: 0, Out: "apps/myapp\npackages/golib\npackages/jslib\npackages/mylib"}, nil},
		}

		t.Log("create")
		runPathCase(initDir.Path())(t, tests)
	})

	runStep(t, "aliases", func(t *testing.T) {
		tests := []testMatchCase{
			{"with no args and no aliases return message no aliases", []string{"aliases"}, icmd.Expected{ExitCode: 0, Out: "No aliases defined\n"}, ""},
			{"create with invalid project should error", []string{"aliases", "add", "invalidPath", "invalidAlias", "-C"}, icmd.Expected{ExitCode: 1, Err: "Error: unknown project invalidPath"}, ""},
			{"create with invalid alias should error", []string{"aliases", "add", "packages/golib", "1golib", "-C"}, icmd.Expected{ExitCode: 1, Err: "Error: invalid alias name 1golib"}, ""},
			{"create with valid args should work", []string{"aliases", "add", "packages/golib", "golib"}, icmd.Expected{ExitCode: 0}, ""},
			{"create with valid args should work", []string{"aliases", "add", "packages/jslib", "jslib"}, icmd.Expected{ExitCode: 0}, ""},
			{"list should display existing aliases", []string{"aliases", "list"}, icmd.Expected{ExitCode: 0}, "^(golib: packages/golib\n|jslib: packages/jslib\n){2}$"},
			{"remove should fail silently on invalid alias", []string{"aliases", "remove", "test"}, icmd.Expected{ExitCode: 0}, ""},
			{"remove should remove with valid alias", []string{"aliases", "remove", "golib"}, icmd.Expected{ExitCode: 0}, ""},
			{"list should not display removed aliases", []string{"aliases", "list"}, icmd.Expected{ExitCode: 0}, "^jslib: packages/jslib\n$"},
		}
		t.Log("aliases")
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Log("case:", tc.name)
				result := runMonospace(tc.args, initDirOp)
				result.Assert(t, tc.want)
				if tc.match != "" {
					exp := regexp.MustCompile(tc.match)
					assert.Assert(t, exp.MatchString(result.Combined()))
				}
			})
		}
	})

	runStep(t, "import", func(t *testing.T) {
		tests := []testPathCase{
			{"from other repo should work", []string{"import", "modules/external", "file://" + filepath.Join(monospaceDir, "../../gomodules/js-packagemanager")}, icmd.Expected{ExitCode: 0},
				hasDir("modules", false, hasDir("external", true,
					hasFile("go.mod"),
				)),
			},
			{"repo should appear when ls", []string{"ls", "-C"}, icmd.Expected{ExitCode: 0, Out: "apps/myapp\nmodules/external\npackages/golib\npackages/jslib\npackages/mylib"}, nil},
		}
		t.Log("import")
		runPathCase(initDir.Path())(t, tests)
	})

	runStep(t, "clone", func(t *testing.T) {
		// we need to comit changes to the initialized repo if we want something to clone
		icmd.RunCmd(icmd.Command("git", "add", "."), initDirOp).Assert(t, icmd.Expected{ExitCode: 0})
		icmd.RunCmd(icmd.Command("git", "commit", "-m", "commitChanges"), initDirOp).Assert(t, icmd.Expected{ExitCode: 0, Out: ""})

		tests := []testPathCase{
			{"should clone external projects too", []string{"clone", initDir.Path(), "clonedRepo"}, icmd.Expected{ExitCode: 0},
				hasDir("clonedRepo", true,
					hasDir(".monospace", true),
					hasDir(".git", true),
					hasDir("packages", false,
						hasDir("mylib", true),
						hasDir("golib", true, hasFile("go.mod")),
						hasDir("jslib", true, hasFile("package.json")),
					),
					hasDir("modules", false,
						hasDir("external", true,
							hasDir(".git", true),
							hasFile("packagemanager.go"),
						),
					),
				),
			},
			{"should error on non monospace repo", []string{"clone", "git@github.com:malko/rocketchat-jira-hook.git"}, icmd.Expected{ExitCode: 1},
				hasDir("rocketchat-jira-hook", true, hasDir(".git", true)),
			},
		}
		t.Log("clone")
		t.Run(tests[0].name, func(t *testing.T) {
			test := tests[0]
			result := runMonospace(test.args, icmd.Dir(cloneDir.Path()))
			result.Assert(t, test.want)
			assert.Assert(t, fs.Equal(cloneDir.Path(), fs.Expected(t, test.pathOps, fs.MatchExtraFiles, fs.MatchAnyFileMode)))
		})
		t.Run(tests[1].name, func(t *testing.T) {
			test := tests[1]
			result := runMonospace(test.args, icmd.Dir(cloneDir.Path()))
			result.Assert(t, test.want)
			assert.Assert(t, fs.Equal(cloneDir.Path(), fs.Expected(t, test.pathOps, fs.MatchExtraFiles, fs.MatchAnyFileMode)))
			assert.Assert(t, strings.Contains(result.Combined(), "&& monospace init"))
		})
	})
}
