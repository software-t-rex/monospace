package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/software-t-rex/monospace/app"
	"gopkg.in/yaml.v3"
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

var unSkipedSteps []string = []string{
	"create",
	"aliases",     // require create
	"import",      // require create
	"clone",       // require import, create
	"rename",      // require import, create
	"remove",      // require create
	"ls",          // require all previous steps
	"run",         // require clone, and steps above
	"exec",        // require run and all steps required by run
	"check",       // require create, import and rename
	"externalize", // no deps
	"status",
}

func skipOrContinue(t *testing.T, stepName string) {
	for _, val := range unSkipedSteps {
		if val == stepName {
			return
		}
	}
	t.Skip()
}

func TestCmd_Suite(t *testing.T) {

	// SETUP TEST SUITE
	type testCase struct {
		name    string        // test name
		args    []string      // monospace args
		want    icmd.Expected // test command output
		pathOps fs.PathOp     // check directory structure agains pathOps
		match   string        // regexp to run against combined output empty string is ignored
	}

	// prepare binary to test
	monospaceDir := filepath.Join(dirname(), "../")
	tmpDir := fs.NewDir(t, "mstest-build")
	monospaceBin := filepath.Join(tmpDir.Path(), "monospace-test")
	if runtime.GOOS == "windows" {
		monospaceBin += ".exe"
	}
	icmd.RunCmd(icmd.Command(
		"go", "build", "-cover", "-o", monospaceBin, "main.go",
	), icmd.Dir(monospaceDir)).
		Assert(t, icmd.Success)

		// prepare some dirs to work with
	initDir := fs.NewDir(t, "mstest-init")
	initDirOp := icmd.Dir(initDir.Path())
	cloneDir := fs.NewDir(t, "mstest-clone")

	// test suite helpers methods
	runMonospace := func(args []string, ops ...icmd.CmdOp) *icmd.Result {
		cmd := icmd.Command(monospaceBin, args...)
		return icmd.RunCmd(cmd, append(ops,
			icmd.WithEnv(append(os.Environ(), "GOCOVERDIR="+tmpDir.Path(), "NO_COLOR=1")...),
			icmd.WithTimeout(time.Second*30),
		)...)
	}
	hasFile := func(name string) fs.PathOp { return fs.WithFile(name, "", fs.MatchAnyFileContent, fs.MatchAnyFileMode) }
	hasDir := func(name string, extraFiles bool, ops ...fs.PathOp) fs.PathOp {
		ops = append(ops, fs.MatchAnyFileMode)
		if extraFiles {
			ops = append(ops, fs.MatchExtraFiles)
		}
		return fs.WithDir(name, ops...)
	}
	runTestCases := func(path string) func(*testing.T, []testCase) {
		dirOp := icmd.Dir(path)
		return func(t *testing.T, tests []testCase) {
			t.Helper()
			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					t.Helper()
					result := runMonospace(tc.args, dirOp)
					result.Assert(t, tc.want)
					if tc.pathOps != nil {
						assert.Assert(t, fs.Equal(path, fs.Expected(t, tc.pathOps, fs.MatchExtraFiles, fs.MatchAnyFileMode)))
					}
					if tc.match != "" {
						exp := regexp.MustCompile(tc.match)
						assert.Assert(t, exp.MatchString(result.Combined()), "expected to match against %s, got: %s", tc.match, result.Combined())
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
		result.Assert(t, icmd.Success)
		expected := fs.Expected(t,
			hasFile("package.json"),
			hasFile(".npmrc"),
			hasFile(".gitignore"),
			hasFile("go.work"),
			hasFile("pnpm-workspace.yaml"),
			hasDir(".git", true),
			hasDir(".monospace", false,
				hasFile("monospace.yml"),
				hasDir("bin", false),
				hasDir("githooks", false,
					hasFile("post-merge"),
					hasFile("post-checkout"),
				),
			),
		)
		assert.Assert(t, fs.Equal(initDir.Path(), expected))
		runMonospace([]string{"ls"}, icmd.Dir(initDir.Path())).Assert(t, icmd.Expected{
			Out: "No projects found start by adding one to your monospace.",
		})
	})

	runStep(t, "create", func(t *testing.T) {
		skipOrContinue(t, "create")
		tests := []testCase{
			{"ls without project should display a message", []string{"ls"}, icmd.Expected{ExitCode: 0, Out: "No projects found start by adding one to your monospace."}, nil, ""},
			{"with no args", []string{"create"}, icmd.Expected{ExitCode: 1}, nil, ""},
			{"with 1 args", []string{"create", "local"}, icmd.Expected{ExitCode: 1}, nil, ""},
			{"with 2 args but invalid kind arg", []string{"create", "invalid", "apps/myapp"}, icmd.Expected{ExitCode: 1}, nil, ""},
			{"with invalid name", []string{"create", "--no-interactive", "internal", "apps!#/1myapp"}, icmd.Expected{ExitCode: 1}, nil, ""},
			{"local", []string{"create", "--no-interactive", "local", "apps/myapp"}, icmd.Success,
				hasDir("apps", false,
					hasDir("myapp", false,
						hasDir(".git", true),
						hasFile(".gitignore"),
					),
				),
				"",
			},
			{"internal", []string{"create", "--no-interactive", "internal", "packages/mylib"}, icmd.Success,
				hasDir("packages", false, hasDir("mylib", false, hasFile(".gitignore"))), "",
			},
			{"internal go type", []string{"create", "--no-interactive", "internal", "packages/golib", "-t", "go"}, icmd.Success,
				hasDir("packages", true, hasDir("golib", false,
					hasFile(".gitignore"),
					hasFile("go.mod"),
					hasFile("main.go"),
				)),
				"",
			},
			{"internal js type", []string{"create", "--no-interactive", "internal", "packages/jslib", "-t", "js"}, icmd.Success,
				hasDir("packages", true, hasDir("jslib", false,
					hasFile(".gitignore"),
					hasFile("package.json"),
					hasFile("index.js"),
				)),
				"",
			},
			{"ls should list added projects sorted", []string{"ls", "-C"}, icmd.Expected{ExitCode: 0, Out: "apps/myapp\npackages/golib\npackages/jslib\npackages/mylib"}, nil, ""},
		}

		t.Log("create")
		runTestCases(initDir.Path())(t, tests)
	})

	runStep(t, "aliases", func(t *testing.T) {
		skipOrContinue(t, "aliases")
		tests := []testCase{
			{"with no args and no aliases return message no aliases", []string{"aliases"}, icmd.Expected{ExitCode: 0, Out: "No aliases defined\n"}, nil, ""},
			{"create with invalid project should error", []string{"aliases", "add", "invalidPath", "invalidAlias", "-C"}, icmd.Expected{ExitCode: 1, Err: "Error: unknown project invalidPath"}, nil, ""},
			{"create with invalid alias should error", []string{"aliases", "add", "packages/golib", "1golib", "-C"}, icmd.Expected{ExitCode: 1, Err: "Error: invalid alias name 1golib"}, nil, ""},
			{"create with valid args should work", []string{"aliases", "add", "packages/golib", "golib"}, icmd.Success, nil, ""},
			{"create with valid args should work", []string{"aliases", "add", "packages/jslib", "jslib"}, icmd.Success, nil, ""},
			{"list should display existing aliases", []string{"aliases", "list"}, icmd.Success, nil, "^(golib: packages/golib\n|jslib: packages/jslib\n){2}$"},
			{"remove should fail silently on invalid alias", []string{"aliases", "remove", "test"}, icmd.Success, nil, ""},
			{"remove should remove with valid alias", []string{"aliases", "remove", "golib"}, icmd.Success, nil, ""},
			{"list should not display removed aliases", []string{"aliases", "list"}, icmd.Success, nil, "^jslib: packages/jslib\n$"},
		}
		t.Log("aliases")
		runTestCases(initDir.Path())(t, tests)
	})

	runStep(t, "import", func(t *testing.T) {
		skipOrContinue(t, "import")
		tests := []testCase{
			{"from other repo should work", []string{"import", "modules/external", "file://" + filepath.Join(monospaceDir, "../../gomodules/js-packagemanager")}, icmd.Success,
				hasDir("modules", false, hasDir("external", true,
					hasFile("go.mod"),
				)), "",
			},
			{"repo should appear when ls", []string{"ls", "-C"}, icmd.Expected{ExitCode: 0, Out: "apps/myapp\nmodules/external\npackages/golib\npackages/jslib\npackages/mylib"}, nil, ""},
		}
		t.Log("import")
		runTestCases(initDir.Path())(t, tests)
	})

	runStep(t, "clone", func(t *testing.T) {
		skipOrContinue(t, "clone")
		// we need to comit changes to the initialized repo if we want something to clone
		icmd.RunCmd(icmd.Command("git", "add", "."), initDirOp).Assert(t, icmd.Success)
		icmd.RunCmd(icmd.Command("git", "commit", "-m", "commitChanges"), initDirOp).Assert(t, icmd.Expected{ExitCode: 0, Out: ""})

		tests := []testCase{
			{"should clone external projects too", []string{"clone", initDir.Path(), "clonedRepo"}, icmd.Success,
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
				), "",
			},
			{"should error on non monospace repo", []string{"clone", "git@github.com:malko/rocketchat-jira-hook.git"}, icmd.Expected{ExitCode: 1},
				hasDir("rocketchat-jira-hook", true, hasDir(".git", true)), "&& monospace init",
			},
		}
		t.Log("clone")
		runTestCases(cloneDir.Path())(t, tests)
	})

	runStep(t, "rename", func(t *testing.T) {
		skipOrContinue(t, "rename")
		tests := []testCase{
			{"should error on invalid project", []string{"rename", "inexsiting", "renamed"}, icmd.Expected{ExitCode: 1}, nil, "Unkwown project"},
			{"should error with invalid project name", []string{"rename", "modules/external", "modules/&renamed"}, icmd.Expected{ExitCode: 1}, nil, "not a valid project name"},
			{"should error if newname point to existing project", []string{"rename", "modules/external", "packages/golib"}, icmd.Expected{ExitCode: 1}, nil, "already exists"},
			{"should error when renaming to the same name", []string{"rename", "modules/external", "modules/external"}, icmd.Expected{ExitCode: 1}, nil, "already exists"},
			{"should work with valid args", []string{"rename", "modules/external", "modules/renamed"}, icmd.Success,
				hasDir("modules", false, hasDir("renamed", true)), "",
			},
		}
		t.Log("rename")
		runTestCases(initDir.Path())(t, tests)
	})

	runStep(t, "remove", func(t *testing.T) {
		skipOrContinue(t, "remove")
		tests := []testCase{
			{"should error on unknown project", []string{"remove", "unknown"}, icmd.Expected{ExitCode: 1}, nil, "Unknown project"},
			{"should keep the directory if -y", []string{"remove", "packages/mylib", "-y"}, icmd.Success,
				hasDir("packages", false,
					hasDir("golib", true),
					hasDir("jslib", true),
					hasDir("mylib", true),
				), "",
			},
			{"should delete the directory if -rmdir", []string{"remove", "packages/golib", "--rmdir"}, icmd.Success,
				hasDir("packages", false,
					hasDir("jslib", true),
					hasDir("mylib", true),
				), "",
			},
		}
		t.Log("rename")
		runTestCases(initDir.Path())(t, tests)
	})

	runStep(t, "ls", func(t *testing.T) {
		skipOrContinue(t, "ls")
		longOutput := "apps/myapp (local)\nmodules/renamed (file:///home/malko/git/T-REx/monospace/gomodules/js-packagemanager)\npackages/jslib (internal)"
		cloneOutput := "apps/myapp\nmodules/external\npackages/golib\npackages/jslib\npackages/mylib"
		tests := []testCase{
			{"should return the list of projects", []string{"ls", "-C"}, icmd.Expected{ExitCode: 0, Out: "apps/myapp\nmodules/renamed\npackages/jslib"}, nil, ""},
			{"should add detail with -long flag", []string{"ls", "-C", "--long"}, icmd.Expected{ExitCode: 0, Out: longOutput}, nil, ""},
			{"should support list alias", []string{"list", "-C", "-l"}, icmd.Expected{ExitCode: 0, Out: longOutput}, nil, ""},
			{"should support monospaceRoot as argument", []string{"ls", filepath.Join(cloneDir.Path(), "clonedRepo")}, icmd.Expected{ExitCode: 0, Out: cloneOutput}, nil, ""},
			{"should support a monospace subdir path as argument", []string{"ls", filepath.Join(cloneDir.Path(), "clonedRepo/modules/")}, icmd.Expected{ExitCode: 0, Out: cloneOutput}, nil, ""},
			{"should print an error with a dir which is not a monospace as argument", []string{"ls", filepath.Join(cloneDir.Path(), "rocketchat-jira-hook")}, icmd.Expected{ExitCode: 1}, nil, "not part of a monospace"},
			{"should print an error with a dir which doesn't exists as argument", []string{"ls", filepath.Join(cloneDir.Path(), "notexists")}, icmd.Expected{ExitCode: 1}, nil, "no such file or directory"},
		}
		t.Log("ls")
		runTestCases(initDir.Path())(t, tests)
	})

	runStep(t, "run", func(t *testing.T) {
		skipOrContinue(t, "run")
		// add some tasks to initial repo
		confPath := cloneDir.Join("clonedRepo/.monospace/monospace.yml")
		config, _ := app.ConfigRead(confPath)
		saveConfig := func() {
			rawConfig, _ := yaml.Marshal(config)
			os.WriteFile(confPath, rawConfig, 0640)
		}
		sayOk := []string{"echo", "ok"}
		sayHello := []string{"echo", "hello"}
		config.Pipeline = make(map[string]app.MonospaceConfigTask)
		delete(config.Projects, "apps/myapp") // avoid error on local project
		config.Aliases["golib"] = "packages/golib"
		config.Aliases["mylib"] = "packages/mylib"
		config.Pipeline["list"] = app.MonospaceConfigTask{Cmd: []string{"ls"}}
		config.Pipeline["jslib#sayok"] = app.MonospaceConfigTask{Cmd: sayOk}
		config.Pipeline["golib#sayok"] = app.MonospaceConfigTask{Cmd: sayOk}
		config.Pipeline["mylib#sayhello"] = app.MonospaceConfigTask{Cmd: sayHello, DependsOn: []string{"golib#sayok"}}
		config.Pipeline["unknown#sayok"] = app.MonospaceConfigTask{Cmd: sayOk}
		config.Pipeline["sayhellofromroot"] = app.MonospaceConfigTask{Cmd: []string{"echo", "hello from root"}}
		// os.Remove(initDir.Join(".monospace/monospace.yml"))
		saveConfig()

		runTCs := runTestCases(cloneDir.Join("clonedRepo"))
		runTC := func(tc testCase) {
			t.Helper()
			runTCs(t, []testCase{tc})
		}

		runTC(testCase{
			"should error if pipeline define a task for unknown project",
			[]string{"run", "-C", "sayok"},
			icmd.Expected{ExitCode: 1},
			nil,
			"unknown is neither a project name or an alias",
		})
		// remove erroneous task
		delete(config.Pipeline, "unknown#sayok")
		saveConfig()

		runTC(testCase{
			"shoud error on unknown task",
			[]string{"run", "-C", "unknown"},
			icmd.Expected{ExitCode: 1},
			nil, "no tasks found",
		})
		runTC(testCase{
			"when unfiltered shoud run top level task on all project and root",
			[]string{"run", "-C", "sayhellofromroot"},
			icmd.Success,
			nil,
			`(?:(?:packages\/(?:mylib|golib|jslib)|modules\/external|root)#sayhellofromroot\s+succeed.*\s*hello from root[\s\S]*?){5}.*Tasks: ✔ 5 succeed / 5 total`,
		})
		runTC(testCase{
			"when filtered shoud run top level task on selected project only",
			[]string{"run", "-C", "sayhellofromroot", "-p", "jslib", "-p", "packages/mylib"},
			icmd.Success,
			nil,
			`(?:(?:packages\/(?:mylib|jslib))#sayhellofromroot\s+succeed.*\s*hello from root[\s\S]*?){2}.*Tasks: ✔ 2 succeed / 2 total`,
		})
		runTC(testCase{
			"when unfiltered shoud run not top level tasks on all project that supports it",
			[]string{"run", "-C", "sayok"},
			icmd.Success,
			nil,
			`(?:(?:packages\/(?:golib|jslib))#sayok\s+succeed.*\s*ok[\s\S]*?){2}.*Tasks: ✔ 2 succeed / 2 total`,
		})
		runTC(testCase{
			"when filtered shoud run not top level tasks on filtered project that supports it only",
			[]string{"run", "-C", "sayok", "-p", "jslib", "mylib"},
			icmd.Success,
			nil,
			`Tasks: ✔ 1 succeed / 1 total`,
		})
		runTC(testCase{
			"shoud run dependency task first",
			[]string{"run", "-C", "sayhello", "-p", "mylib"},
			icmd.Success,
			nil,
			`golib#sayok\s+succeed[\s\S]+mylib#sayhello\s+succeed`,
		})
	})

	runStep(t, "run on root project", func(t *testing.T) {
		skipOrContinue(t, "run")
		runTCs := runTestCases(cloneDir.Join("clonedRepo"))
		runTC := func(tc testCase) {
			t.Helper()
			runTCs(t, []testCase{tc})
		}
		// testing the interaction with root package
		runTC(testCase{
			"when filter root only shoud run top level task on root only",
			[]string{"run", "-C", "list", "-p", "root"},
			icmd.Success,
			nil,
			`root#list succeed.*\n\s+go.work\n\s+modules\n\s+package.json\n\s+packages\n\s+pnpm-workspace.yaml\s+Tasks: ✔ 1 succeed / 1 total`,
		})
		runTC(testCase{
			"when not only filter root shoud run top level task on root and others fitlered projects",
			[]string{"run", "-C", "sayhellofromroot", "-p", "root,mylib"},
			icmd.Success,
			nil,
			`(?:(?:packages/mylib|root)#sayhellofromroot\s+succeed.*\s*hello from root[\s\S]*?){2}.*Tasks: ✔ 2 succeed / 2 total`,
		})
	})

	runStep(t, "run with additional args", func(t *testing.T) {
		skipOrContinue(t, "run")
		runTCs := runTestCases(cloneDir.Join("clonedRepo"))
		runTC := func(tc testCase) {
			t.Helper()
			runTCs(t, []testCase{tc})
		}
		// testing the interaction with root package
		runTC(testCase{
			"shoud pass additional args to task",
			[]string{"run", "-C", "sayhellofromroot", "-p", "root", "--", "with additional arg"},
			icmd.Success,
			nil,
			`root#sayhellofromroot succeed.*\s*hello from root with additional arg`,
		})
	})

	runStep(t, "exec", func(t *testing.T) {
		skipOrContinue(t, "exec")
		runTCs := runTestCases(cloneDir.Join("clonedRepo"))
		runTC := func(tc testCase) {
			t.Helper()
			runTCs(t, []testCase{tc})
		}
		runTC(testCase{
			"shoud call command on all known projects but not on root",
			[]string{"exec", "-C", "echo", "ok"},
			icmd.Success,
			nil,
			`(?:(?:packages\/(?:mylib|golib|jslib)|modules\/external): echo ok\s+succeed.*\s*ok[\s\S]*?){4}.*Tasks: ✔ 4 succeed / 4 total`,
		})
		runTC(testCase{
			"with --include-root shoud call command on root too",
			[]string{"exec", "-C", "--include-root", "echo", "ok"},
			icmd.Success,
			nil,
			`(?:(?:packages\/(?:mylib|golib|jslib)|modules\/external|root): echo ok\s+succeed.*\s*ok[\s\S]*?){5}.*Tasks: ✔ 5 succeed / 5 total`,
		})
		runTC(testCase{
			"with root filter only shoud call command on root only",
			[]string{"exec", "-C", "ls", "-p", "root"},
			icmd.Success,
			nil,
			`root: ls succeed .*\s+go.work\s+modules\s+package.json\s+packages\s+pnpm-workspace.yaml\s+Tasks: ✔ 1 succeed / 1 total`,
		})
		runTC(testCase{
			"shoud pass additional args to underlying command",
			[]string{"exec", "-C", "echo", "ok", "-p", "root,mylib", "--", "with args"},
			icmd.Success,
			nil,
			`(?:(?:packages/mylib|root): echo ok with args succeed.*\s+ok with args[\s\S]+){2}Tasks: ✔ 2 succeed / 2 total`,
		})
		runTC(testCase{
			"with --git shoud call command on git projects only (no root)",
			[]string{"exec", "-C", "--git", "echo", "ok"},
			icmd.Success,
			nil,
			`(?:modules\/external: echo ok\s+succeed.*\s*ok[\s\S]*?){1}.*Tasks: ✔ 1 succeed / 1 total`,
		})
		runTC(testCase{
			"with --git --include-root shoud call command on git projects only (with root)",
			[]string{"exec", "-C", "--git", "-r", "echo", "ok"},
			icmd.Success,
			nil,
			`(?:(?:root|modules\/external): echo ok\s+succeed.*\s*ok[\s\S]*?){2}.*Tasks: ✔ 2 succeed / 2 total`,
		})
		runTC(testCase{
			"with --external shoud call command on external projects only",
			[]string{"exec", "-C", "--external", "echo", "ok"},
			icmd.Success,
			nil,
			`(?:modules\/external: echo ok\s+succeed.*\s*ok[\s\S]*?){1}.*Tasks: ✔ 1 succeed / 1 total`,
		})
		runTC(testCase{
			"with --internal shoud call command on internal projects only",
			[]string{"exec", "-C", "--internal", "echo", "ok"},
			icmd.Success,
			nil,
			`(?:(?:packages\/(?:mylib|golib|jslib)): echo ok\s+succeed.*\s*ok[\s\S]*?){3}.*Tasks: ✔ 3 succeed / 3 total`,
		})
		// Add a local project (setup for next test case)
		runTC(testCase{
			"should be able to add local project",
			[]string{"create", "local", "modules/local"},
			icmd.Success,
			hasDir("modules", true, hasDir("local", true)),
			"",
		})
		runTC(testCase{
			"with --local shoud call command on local projects only",
			[]string{"exec", "-C", "--local", "echo", "ok"},
			icmd.Success,
			nil,
			`(?:modules\/local: echo ok\s+succeed.*\s*ok[\s\S]*?){1}.*Tasks: ✔ 1 succeed / 1 total`,
		})
		runTC(testCase{
			"with --local --external shoud call command on local and external projects only",
			[]string{"exec", "-C", "--local", "--external", "echo", "ok"},
			icmd.Success,
			nil,
			`(?:(?:modules\/local|modules\/external): echo ok\s+succeed.*\s*ok[\s\S]*?){2}.*Tasks: ✔ 2 succeed / 2 total`,
		})
		runTC(testCase{
			"with --local --internal shoud call command on local and internal projects only",
			[]string{"exec", "-C", "--local", "--internal", "echo", "ok"},
			icmd.Success,
			nil,
			`(?:(?:modules\/local|packages\/(?:mylib|golib|jslib)): echo ok\s+succeed.*\s*ok[\s\S]*?){4}.*Tasks: ✔ 4 succeed / 4 total`,
		})
		runTC(testCase{
			"with --external --internal shoud call command on external and internal projects only",
			[]string{"exec", "-C", "--external", "--internal", "echo", "ok"},
			icmd.Success,
			nil,
			`(?:(?:modules\/external|packages\/(?:mylib|golib|jslib)): echo ok\s+succeed.*\s*ok[\s\S]*?){4}.*Tasks: ✔ 4 succeed / 4 total`,
		})
		runTC(testCase{
			"with --include-root --local --external --internal --git shoud call command only on git projects (with root)",
			[]string{"exec", "-C", "--include-root", "--local", "--external", "--internal", "--git", "echo", "ok"},
			icmd.Success,
			nil,
			`(?:(?:root|modules\/(?:loc|extern)al): echo ok\s+succeed.*\s*ok[\s\S]*?){3}.*Tasks: ✔ 3 succeed / 3 total`,
		})

	})

	runStep(t, "check local projects", func(t *testing.T) {
		skipOrContinue(t, "check")
		runTCs := runTestCases(initDir.Path())
		runTC := func(tc testCase) {
			t.Helper()
			runTCs(t, []testCase{tc})
		}
		checkIsClean := func(t *testing.T) {
			t.Helper()
			assert.Assert(t, runMonospace([]string{"check"}, initDirOp).Compare(icmd.Success), "check should not return errors")
		}

		runTC(testCase{
			"should not return errors with no anomalies", []string{"check"},
			icmd.Success, nil, "",
		})

		t.Log("local projects with origin")
		t.Run("Should be able to add origin to local project", func(t *testing.T) {
			icmd.RunCmd(icmd.Command(
				"git", "-C", initDir.Join("apps/myapp"), "remote", "add", "origin", "file:///home/malko/git/T-REx/monospace/gomodules/js-packagemanager",
			)).Assert(t, icmd.Success)
		})
		runTC(testCase{
			"should return error when local project has origin", []string{"check"},
			icmd.Expected{ExitCode: 1}, nil, "origin is set.*?for local project apps/myapp",
		})
		runTC(testCase{
			"with --fix should set local project to external if has origin", []string{"check", "--fix"},
			icmd.Success, nil, "setting project.+?as external",
		})
		t.Run("local project should be fixed to external", func(t *testing.T) {
			icmd.RunCmd(icmd.Command(
				"git", "-C", initDir.Join("apps/myapp"), "remote", "show",
			)).Assert(t, icmd.Expected{ExitCode: 0, Out: "origin"})
			checkIsClean(t)
		})

		t.Log("local projects without git repo")
		t.Run("Add another local project", func(t *testing.T) { // this is more setup than a test
			runMonospace([]string{"create", "local", "modules/local"}, initDirOp).Assert(t, icmd.Success)
			t.Log("Remove git directory from newly created project")
			fs.DirFromPath(t, initDir.Join("modules/local/.git")).Remove()
		})
		runTC(testCase{"should return errors when local is not a repo", []string{"check"},
			icmd.Expected{ExitCode: 1}, nil, "is not a git repository",
		})
		runTC(testCase{"with --fix should set to internal when local is not a repo", []string{"check", "--fix"},
			icmd.Success, nil, "",
		})
		t.Run("local project should be fixed to internal", func(t *testing.T) {
			cmdRes := runMonospace([]string{"ls", "-l"}, initDirOp)
			cmdRes.Assert(t, icmd.Success)
			assert.Assert(t, strings.Contains(cmdRes.Stdout(), "modules/local (internal)"), "modules/local should be internal")
		})
		t.Run("local project should be fixed to internal", func(t *testing.T) {
			cmdRes := runMonospace([]string{"ls", "-l"}, initDirOp)
			cmdRes.Assert(t, icmd.Success)
			assert.Assert(t, strings.Contains(cmdRes.Stdout(), "modules/local (internal)"), "modules/local should be internal")
		})

		t.Log("local projects without directory should not error")
		t.Run("Add another local project", func(t *testing.T) {
			cmdRes := runMonospace([]string{"create", "local", "modules/local2"}, initDirOp)
			cmdRes.Assert(t, icmd.Success)
			fs.DirFromPath(t, initDir.Join("modules/local2")).Remove()
		})
		runTC(testCase{"should not return errors when local project has no directory", []string{"check"},
			icmd.Success, nil, "",
		})
	})

	runStep(t, "check external projects", func(t *testing.T) {
		skipOrContinue(t, "check")
		runTCs := runTestCases(initDir.Path())
		runTC := func(tc testCase) {
			t.Helper()
			runTCs(t, []testCase{tc})
		}
		checkIsClean := func(t *testing.T) {
			t.Helper()
			assert.Assert(t, runMonospace([]string{"check"}, initDirOp).Compare(icmd.Success), "check should not return errors")
		}

		checkIsClean(t)
		t.Log("external projects without matching directory")
		t.Run("Remove project directory", func(t *testing.T) {
			fs.DirFromPath(t, initDir.Join("modules/renamed")).Remove()
		})
		runTC(testCase{
			"should errors if project directory doesn't exists", []string{"check"},
			icmd.Expected{ExitCode: 1}, nil, "",
		})
		runTC(testCase{
			"with --fix should clone remote project in dest dir", []string{"check", "--fix"},
			icmd.Success, nil, `modules/renamed\s+cloning...`,
		})
		t.Run("project should be cloned", func(t *testing.T) {
			assert.Assert(t, fs.Equal(initDir.Path(), fs.Expected(t,
				hasDir("modules", true,
					hasDir("renamed", true,
						hasDir(".git", true),
					),
				), fs.MatchExtraFiles,
			)), "project should be cloned")
			checkIsClean(t)
		})

		t.Log("external projects with mismatch origin")
		t.Run("change origin url", func(t *testing.T) {
			icmd.RunCmd(icmd.Command(
				"git", "-C", initDir.Join("modules/renamed"), "remote", "set-url", "origin", "git@github.com/whatever",
			)).Assert(t, icmd.Success)
		})
		runTC(testCase{
			"should errors if project origin doesn't match", []string{"check"},
			icmd.Expected{ExitCode: 1}, nil, "",
		})
		runTC(testCase{
			"with --fix should set update config to match origin", []string{"check", "--fix"},
			icmd.Success, nil, "",
		})
		t.Run("project external config should match project origin", func(t *testing.T) {
			cmdRes := runMonospace([]string{"ls", "-l"}, initDirOp)
			cmdRes.Assert(t, icmd.Success)
			assert.Assert(t, strings.Contains(cmdRes.Stdout(), "modules/renamed (git@github.com/whatever)"), "modules/renamed should be external")
			checkIsClean(t)
		})

		t.Log("external projects wich is not a git repo")
		t.Run("Remove git directory from project", func(t *testing.T) {
			fs.DirFromPath(t, initDir.Join("modules/renamed/.git")).Remove()
		})
		runTC(testCase{
			"should errors if project is not a git repo", []string{"check"},
			icmd.Expected{ExitCode: 1}, nil, "",
		})
		runTC(testCase{
			"with --fix should set project to internal", []string{"check", "--fix"},
			icmd.Success, nil, "",
		})
		t.Run("project should be set to internal", func(t *testing.T) {
			cmdRes := runMonospace([]string{"ls", "-l"}, initDirOp)
			cmdRes.Assert(t, icmd.Success)
			assert.Assert(t, strings.Contains(cmdRes.Stdout(), "modules/renamed (internal)"), "modules/renamed should be internal")
			checkIsClean(t)
		})
	})

	runStep(t, "check internal projects", func(t *testing.T) {
		skipOrContinue(t, "check")
		runTCs := runTestCases(initDir.Path())
		runTC := func(tc testCase) {
			t.Helper()
			runTCs(t, []testCase{tc})
		}
		checkIsClean := func(t *testing.T) {
			t.Helper()
			assert.Assert(t, runMonospace([]string{"check"}, initDirOp).Compare(icmd.Success), "check should not return errors")
		}

		checkIsClean(t)
		t.Log("internal projects without matching directory")
		t.Run("Remove project directory", func(t *testing.T) {
			fs.DirFromPath(t, initDir.Join("packages/jslib")).Remove()
		})
		runTC(testCase{
			"should errors if project directory doesn't exists", []string{"check"},
			icmd.Expected{ExitCode: 1}, nil, "",
		})
		runTC(testCase{
			"with --fix should remove project from config", []string{"check", "--fix"},
			icmd.Success, nil, "",
		})
		t.Run("project should be removed from config", func(t *testing.T) {
			cmdRes := runMonospace([]string{"ls", "-l"}, initDirOp)
			cmdRes.Assert(t, icmd.Success)
			assert.Assert(t, !strings.Contains(cmdRes.Stdout(), "packages/jslib"), "packages/jslib should be removed")
			checkIsClean(t)
		})

		t.Log("internal projects which is a git repo")
		t.Run("Add internal project with a repo", func(t *testing.T) {
			runMonospace([]string{"create", "internal", "packages/jslib"}, initDirOp).Assert(t, icmd.Success)
			icmd.RunCmd(icmd.Command("git", "-C", initDir.Join("packages/jslib"), "init")).Assert(t, icmd.Success)
		})
		runTC(testCase{
			"should errors if project is a git repo", []string{"check"},
			icmd.Expected{ExitCode: 1}, nil, "",
		})
		runTC(testCase{
			"with --fix should error if origin is not set", []string{"check", "--fix"},
			icmd.Expected{ExitCode: 1}, nil, "",
		})
		t.Run("set origin", func(t *testing.T) {
			icmd.RunCmd(icmd.Command(
				"git", "-C", initDir.Join("packages/jslib"), "remote", "add", "origin", "git@github.com/whatever",
			)).Assert(t, icmd.Success)
		})
		runTC(testCase{
			"with --fix should set project to external", []string{"check", "--fix"},
			icmd.Success, nil, "",
		})
		t.Run("project should be set to external", func(t *testing.T) {
			cmdRes := runMonospace([]string{"ls", "-l"}, initDirOp)
			cmdRes.Assert(t, icmd.Success)
			assert.Assert(t, strings.Contains(cmdRes.Stdout(), "packages/jslib (git@github.com/whatever)"), "packages/jslib should be external")
		})
	})

	runStep(t, "externalize", func(t *testing.T) {
		skipOrContinue(t, "externalize")
		// Setup test
		subExternalOp := hasDir("subExternal", true, fs.WithFile("sub-readme.md", "readme"))
		willExternalOp := hasDir("modules", true,
			hasDir("willExternalize", true,
				fs.WithFile("readme.md", "readme"),
				subExternalOp,
			),
		)
		runMonospace([]string{"create", "internal", "modules/willExternalize"}, initDirOp).Assert(t, icmd.Success)
		fs.Apply(t, initDir, hasDir("modules", true, willExternalOp))
		runMonospace([]string{"exec", "git", "-p", "root", "--", "-C", initDir.Path(), "add", "."}, initDirOp).Assert(t, icmd.Success)
		runMonospace([]string{"exec", "git", "-p", "root", "--", "-C", initDir.Path(), "commit", "-m", "add willExternalize"}, initDirOp).Assert(t, icmd.Success)
		runTCs := runTestCases(initDir.Path())

		// run tests
		tests := []testCase{
			{"should warn on --push without repoUrl", []string{"externalize", "modules/willExternalize", "--push"}, icmd.Expected{ExitCode: 1}, nil, "you must provide a repo url when using --push"},
			{"should correctly externalize an internal project", []string{"externalize", "modules/willExternalize", "-y", "-b", "dfltbranchname"}, icmd.Success, nil, "Externalization done"},
			{"should correctly set initial branch", []string{"exec", "git", "-p", "modules/willExternalize", "--", "branch"}, icmd.Success, nil, `\* dfltbranchname`},
		}
		runTCs(t, tests)

	})

	runStep(t, "status", func(t *testing.T) {
		skipOrContinue(t, "status")
	})
}
