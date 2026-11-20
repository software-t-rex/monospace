package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestConfigRead_NullYAML_NoPanic checks that a YAML file containing "null"
// does not cause a nil pointer dereference (config remains nil after Unmarshal).
func TestConfigRead_NullYAML_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ConfigRead panicked on null YAML (nil pointer dereference): %v", r)
		}
	}()
	configPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(configPath, []byte("null"), 0640); err != nil {
		t.Fatal(err)
	}
	_, _ = ConfigRead(configPath)
}

// TestConfigAddOrUpdateProject_NilProjectsMap checks that calling
// ConfigAddOrUpdateProject when config.Projects is nil does not panic.
func TestConfigAddOrUpdateProject_NilProjectsMap(t *testing.T) {
	savedConfig := appConfig
	defer func() { appConfig = savedConfig }()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ConfigAddOrUpdateProject panicked with nil Projects: %v", r)
		}
	}()
	appConfig = &MonospaceConfig{GoModPrefix: "test.com"} // Projects is nil
	if err := ConfigAddOrUpdateProject("test/project", "local", false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if appConfig.Projects["test/project"] != "local" {
		t.Errorf("project not added, got: %v", appConfig.Projects)
	}
}

var sampleConfig = &MonospaceConfig{
	GoModPrefix:         "test.com",
	JSPM:                "yarn@xxx",
	PreferredOutputMode: "grouped",
	Projects:            map[string]string{"packages/test": "internal"},
	Aliases:             map[string]string{"test": "packages/test"},
}

func TestConfig(t *testing.T) {
	// config should not be loaded at first
	if ConfigIsLoaded() {
		t.Errorf("ConfigIsLoaded(): config should not be loaded before call to ConfigSet")
	}

	got, err := ConfigGet()
	if err == nil {
		t.Errorf("ConfigGet(): config get should nor return a config before call to ConfigSet, got %v", got)
	}

	config := &MonospaceConfig{}
	configSet(config)
	if !(config.JSPM == DfltJSPM && config.GoModPrefix == DfltGoModPrfx) {
		t.Errorf("ConfigSet(): doesn't set default values correctly")
	}

	if !ConfigIsLoaded() {
		t.Errorf("ConfigIsLoaded(): config should be loaded after call to ConfigSet")
	}

	got, err = ConfigGet()
	if err != nil {
		t.Errorf("ConfigGet(): should return loaded config")
	}

	if !reflect.DeepEqual(got, config) {
		t.Errorf("ConfigGet(): bad returned value want: %v, got: %v", config, got)
	}
}

func TestConfigInitAndSave(t *testing.T) {
	//reset config
	appConfig = nil
	var testConfig = *sampleConfig
	configPath := filepath.Join(t.TempDir(), "config.yml")
	testConfig.configPath = configPath
	testConfig.root = filepath.Dir(filepath.Dir(configPath))

	err := ConfigLoad(configPath)
	if err == nil {
		t.Errorf("ConfigInit(): should error on unexisting config file")
	}

	err = ConfigSave()
	if err == nil {
		t.Errorf("ConfigSave(): should error on unloaded config")
	}

	appConfig = &testConfig

	err = ConfigSave()
	if err != nil {
		t.Errorf("ConfigSave(): should save the config without error, %v", err)
	}
	res, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		expected := `# yaml-language-server: $schema=https://raw.githubusercontent.com/software-t-rex/monospace/main/apps/monospace/schemas/monospace.schema.json
go_mod_prefix: test.com
js_package_manager: yarn@xxx
preferred_output_mode: grouped
projects:
    packages/test: internal
projects_aliases:
    test: packages/test
`
		if expected != string(res) {
			t.Fatalf("Saved Config mismatch got '%s', expected '%s'", string(res), expected)
		}
	}

	//reset config
	appConfig = nil
	var got *MonospaceConfig
	got, err = ConfigGet()
	if err == nil || got != nil {
		t.Fatalf("ConfigGet(): should not return a config at this point")
	}
	err = ConfigLoad(configPath)
	if err != nil {
		t.Fatalf("ConfigInit(): failed to load config: %v", err)
	}
	got, err = ConfigGet()
	if err != nil || got == nil {
		t.Fatalf("ConfigGet(): should return an initialized config, got error: %v", err)
	}

	if !reflect.DeepEqual(got, &testConfig) {
		t.Fatalf("ConfigGet(): return different config than expected: %+v, want: %+v, err: %v", got, &testConfig, err)
	}
}

func Dump(o ...any) {
	for _, val := range o {
		fmt.Printf("%+v\n", val)
		fmt.Printf("%#v\n", val)
		out, _ := json.MarshalIndent(val, "", "  ")
		fmt.Print(string(out) + "\n")
	}
}
func TestConfigAddProject(t *testing.T) {
	appConfig = sampleConfig

	err := ConfigAddProject("packages/test", "whatever", false)
	if err == nil {
		t.Errorf("ConfigAddProject(): should report error on adding existing project, err: %v", err)
	}

	err = ConfigAddProject("test/toto", "local", false)
	if appConfig.Projects["test/toto"] != "local" {
		t.Errorf("ConfigAddProject(): should add project to config want: %v, got: %v, err: %v", "local", appConfig.Projects["test/toto"], err)
	}

}
func TestConfigAddProjectAlias(t *testing.T) {
	appConfig = sampleConfig

	err := ConfigAddProjectAlias("packages/test", "test", false)
	if err == nil {
		t.Errorf("ConfigAddProjectAlias(): should report error on existing project alias, err: %v", err)
	}

	err = ConfigAddProjectAlias("packages/test", "aliasname", false)
	if appConfig.Aliases["aliasname"] != "packages/test" {
		t.Errorf("ConfigAddProjectAlias(): should add project alias to config want: %v, got: %v, err: %v", "local", appConfig.Projects["test/toto"], err)
	}
}

func TestConfigRemoveProject(t *testing.T) {
	appConfig = sampleConfig
	appConfig.Projects["test/removableProject"] = "internal"
	appConfig.Aliases["alias1"] = "test/removableProject"
	appConfig.Aliases["alias2"] = "test/removableProject"
	err := ConfigRemoveProject("test/removableProject", false)
	if err != nil {
		t.Errorf("ConfigRemoveProject(): Should not error on removing existing project, err: %v", err)
	}
	p, ok := appConfig.Projects["test/removableProject"]
	if ok || p != "" {
		t.Errorf("ConfigRemoveProject(): should remove project from config")
	}

	err = ConfigRemoveProject("test/unknownProject", false)
	if err != nil {
		t.Errorf("ConfigRemoveProject(): Should not error on removing unknown project, err: %v", err)
	}

	if appConfig.Aliases["alias1"] != "" || appConfig.Aliases["alias2"] != "" {
		t.Errorf("ConfigRemoveProject(): correctly remove all aliases for projects")
	}

}

func TestConfigRemoveProjectAlias(t *testing.T) {
	appConfig = sampleConfig

	appConfig.Aliases["myalias"] = "packages/test"
	err := ConfigRemoveProjectAlias("myalias", false)
	if err != nil {
		t.Errorf("Should not error on removing existing project alias, err: %v", err)
	}
	p, ok := appConfig.Aliases["myalias"]
	if ok || p != "" {
		t.Errorf("ConfigRemoveProjectAlias(): should remove project alias from config")
	}

	err = ConfigRemoveProjectAlias("unknownAlias", false)
	if err != nil {
		t.Errorf("Should not error on removing unknown project alias, err: %v", err)
	}
}

func TestConfigRemoveProjectAliasUpdatesPipeline(t *testing.T) {
	cfg := &MonospaceConfig{
		GoModPrefix: "test.com",
		Projects:    map[string]string{"packages/test": "internal", "packages/other": "internal"},
		Aliases:     map[string]string{"myalias": "packages/test"},
		Pipeline: map[string]MonospaceConfigTask{
			"myalias#build":       {Cmd: []string{"make build"}},
			"myalias#test":        {Cmd: []string{"make test"}, DependsOn: []string{"myalias#build"}},
			"packages/other#ci":   {Cmd: []string{"ci"}, DependsOn: []string{"myalias#test", "packages/other#lint"}},
			"packages/other#lint": {Cmd: []string{"lint"}},
		},
	}
	appConfig = cfg

	err := ConfigRemoveProjectAlias("myalias", false)
	if err != nil {
		t.Fatalf("ConfigRemoveProjectAlias(): unexpected error: %v", err)
	}

	// alias removed
	if _, ok := cfg.Aliases["myalias"]; ok {
		t.Errorf("ConfigRemoveProjectAlias(): alias should be removed")
	}

	// task keys renamed
	if _, ok := cfg.Pipeline["myalias#build"]; ok {
		t.Errorf("ConfigRemoveProjectAlias(): old task key myalias#build should be renamed")
	}
	if _, ok := cfg.Pipeline["packages/test#build"]; !ok {
		t.Errorf("ConfigRemoveProjectAlias(): task key should be renamed to packages/test#build")
	}
	if _, ok := cfg.Pipeline["packages/test#test"]; !ok {
		t.Errorf("ConfigRemoveProjectAlias(): task key should be renamed to packages/test#test")
	}

	// dependsOn updated in renamed task
	testTask := cfg.Pipeline["packages/test#test"]
	if len(testTask.DependsOn) != 1 || testTask.DependsOn[0] != "packages/test#build" {
		t.Errorf("ConfigRemoveProjectAlias(): dependsOn in renamed task should be updated, got: %v", testTask.DependsOn)
	}

	// dependsOn updated in unrelated task
	ciTask := cfg.Pipeline["packages/other#ci"]
	if len(ciTask.DependsOn) != 2 || ciTask.DependsOn[0] != "packages/test#test" {
		t.Errorf("ConfigRemoveProjectAlias(): dependsOn in other tasks should be updated, got: %v", ciTask.DependsOn)
	}
}

func TestConfigRemoveProjectAliasCollision(t *testing.T) {
	cfg := &MonospaceConfig{
		GoModPrefix: "test.com",
		Projects:    map[string]string{"packages/test": "internal", "packages/other": "internal"},
		Aliases:     map[string]string{"myalias": "packages/test"},
		Pipeline: map[string]MonospaceConfigTask{
			"myalias#build":       {Cmd: []string{"make build"}},
			"packages/test#build": {Cmd: []string{"make build2"}}, // collision target
		},
	}
	appConfig = cfg

	err := ConfigRemoveProjectAlias("myalias", false)
	if err == nil {
		t.Fatalf("ConfigRemoveProjectAlias(): expected error for collision, got nil")
	}
	if !strings.Contains(err.Error(), "would overwrite") {
		t.Errorf("ConfigRemoveProjectAlias(): expected 'would overwrite' error, got: %v", err)
	}
}

func TestConfigRemoveProjectCleansPipeline(t *testing.T) {
	cfg := &MonospaceConfig{
		GoModPrefix: "test.com",
		Projects:    map[string]string{"packages/removable": "internal", "packages/other": "internal"},
		Aliases:     map[string]string{"alias1": "packages/removable"},
		Pipeline: map[string]MonospaceConfigTask{
			"packages/removable#build": {Cmd: []string{"build"}},
			"alias1#test":              {Cmd: []string{"test"}, DependsOn: []string{"alias1#build"}},
			"*#lint":                   {Cmd: []string{"lint"}, DependsOn: []string{"build"}},
			"root#ci":                  {Cmd: []string{"ci"}},
			"packages/other#ci": {Cmd: []string{"ci"}, DependsOn: []string{
				"alias1#test", "packages/other#lint", "*#lint", "build",
			}},
			"packages/other#lint": {Cmd: []string{"lint"}},
		},
	}
	appConfig = cfg

	err := ConfigRemoveProject("packages/removable", false)
	if err != nil {
		t.Fatalf("ConfigRemoveProject(): unexpected error: %v", err)
	}

	for _, key := range []string{"packages/removable#build", "alias1#test"} {
		if _, ok := cfg.Pipeline[key]; ok {
			t.Errorf("ConfigRemoveProject(): pipeline task %q should be removed", key)
		}
	}
	// alias of the removed project should be gone
	if _, ok := cfg.Aliases["alias1"]; ok {
		t.Errorf("ConfigRemoveProject(): alias of the removed project should be removed")
	}
	// dependencies on the removed project tasks should be gone, others should be kept
	ci := cfg.Pipeline["packages/other#ci"]
	expectedDeps := []string{"packages/other#lint", "*#lint", "build"}
	if !reflect.DeepEqual(ci.DependsOn, expectedDeps) {
		t.Errorf("ConfigRemoveProject(): unexpected dependsOn, want %v, got %v", expectedDeps, ci.DependsOn)
	}
	// global and root tasks as well as bare dependencies must be kept
	if _, ok := cfg.Pipeline["*#lint"]; !ok {
		t.Errorf("ConfigRemoveProject(): global task *#lint should be kept")
	}
	if _, ok := cfg.Pipeline["root#ci"]; !ok {
		t.Errorf("ConfigRemoveProject(): root task root#ci should be kept")
	}
	// no orphan must remain
	if report := cfg.PipelineFindOrphans(); !report.IsEmpty() {
		t.Errorf("ConfigRemoveProject(): should not leave orphans, got %+v", report)
	}
}

func TestConfigRemoveProjectRemovesAliasedTasks(t *testing.T) {
	cfg := &MonospaceConfig{
		GoModPrefix: "test.com",
		Projects:    map[string]string{"packages/test": "internal", "packages/other": "local"},
		Aliases:     map[string]string{"mylib": "packages/test"},
		Pipeline: map[string]MonospaceConfigTask{
			"mylib#build": {Cmd: []string{"build"}},
			"packages/other#ci": {Cmd: []string{"ci"}, DependsOn: []string{
				"mylib#build", "packages/other#lint",
			}},
			"packages/other#lint": {Cmd: []string{"lint"}},
		},
	}
	appConfig = cfg

	// collect keys before removal, like the remove command does to report them
	if tasks := cfg.ProjectPipelineTasks("packages/test"); !reflect.DeepEqual(tasks, []string{"mylib#build"}) {
		t.Errorf("ProjectPipelineTasks(): want [mylib#build], got %v", tasks)
	}

	if err := ConfigRemoveProject("packages/test", false); err != nil {
		t.Fatalf("ConfigRemoveProject(): unexpected error: %v", err)
	}
	if _, ok := cfg.Pipeline["mylib#build"]; ok {
		t.Errorf("ConfigRemoveProject(): aliased task should be removed")
	}
	ci := cfg.Pipeline["packages/other#ci"]
	if !reflect.DeepEqual(ci.DependsOn, []string{"packages/other#lint"}) {
		t.Errorf("ConfigRemoveProject(): unexpected dependsOn, got %v", ci.DependsOn)
	}
}

func TestConfigRemoveProjectSavesEmptyPipeline(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	cfg := &MonospaceConfig{
		GoModPrefix: "test.com",
		Projects:    map[string]string{"packages/test": "internal"},
		Pipeline:    map[string]MonospaceConfigTask{"packages/test#build": {Cmd: []string{"build"}}},
		configPath:  configPath,
	}
	appConfig = cfg

	if err := ConfigRemoveProject("packages/test", true); err != nil {
		t.Fatalf("ConfigRemoveProject(): unexpected error: %v", err)
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "pipeline:") {
		t.Errorf("ConfigRemoveProject(): empty pipeline should be omitted from config, got:\n%s", string(raw))
	}
}

func TestConfigPipelineFindOrphans(t *testing.T) {
	cfg := &MonospaceConfig{
		GoModPrefix: "test.com",
		Projects:    map[string]string{"packages/test": "internal"},
		Aliases:     map[string]string{"broken": "packages/unknown", "test": "packages/test"},
		Pipeline: map[string]MonospaceConfigTask{
			"packages/ghost#build": {Cmd: []string{"build"}},  // unknown project
			"broken#build":         {Cmd: []string{"build"}},  // alias to unknown project
			"packages/test#lint":   {Cmd: []string{"lint"}},   // valid
			"*#global":             {Cmd: []string{"global"}}, // valid
			"root#rootTask":        {Cmd: []string{"root"}},   // valid
			"test#build":           {Cmd: []string{"build"}},  // valid alias
			"packages/test#ci": {Cmd: []string{"ci"}, DependsOn: []string{ // orphan deps
				"packages/ghost#build", "packages/test#lint", "lint", "packages/ghost#build",
			}},
		},
	}
	appConfig = cfg

	report := cfg.PipelineFindOrphans()
	if report.IsEmpty() {
		t.Fatalf("PipelineFindOrphans(): should report orphans")
	}
	gotTasks := make([]string, 0, len(report.Tasks))
	for _, orphan := range report.Tasks {
		gotTasks = append(gotTasks, orphan.Task)
	}
	expectedTasks := []string{"broken#build", "packages/ghost#build"}
	if !reflect.DeepEqual(gotTasks, expectedTasks) {
		t.Errorf("PipelineFindOrphans(): want tasks %v, got %v", expectedTasks, gotTasks)
	}
	if len(report.Deps) != 1 {
		t.Fatalf("PipelineFindOrphans(): want 1 orphan dep, got %v", report.Deps)
	}
	if report.Deps[0].Task != "packages/test#ci" || report.Deps[0].Dep != "packages/ghost#build" {
		t.Errorf("PipelineFindOrphans(): unexpected orphan dep %+v", report.Deps[0])
	}
	// reading must not modify the config
	if _, ok := cfg.Pipeline["packages/ghost#build"]; !ok {
		t.Errorf("PipelineFindOrphans(): should not modify the pipeline")
	}
}

func TestSplitPipelineTaskName(t *testing.T) {
	tests := []struct {
		key     string
		project string
		task    string
		ok      bool
	}{
		{"build", "", "build", true},
		{"#build", "", "build", true},
		{"*#build", "*", "build", true},
		{"root#build", "root", "build", true},
		{"packages/test#build", "packages/test", "build", true},
		// whitespace must not be trimmed, it makes the project name invalid
		// exactly like tasks.ParseTaskName does
		{" #build", " ", "build", true},
		{"packages/test #build", "packages/test ", "build", true},
		{"build ", "", "build ", true},
		{"a#b#c", "", "", false},
		{"#", "", "", false},
	}
	for _, tt := range tests {
		project, task, ok := splitPipelineTaskName(tt.key)
		if project != tt.project || task != tt.task || ok != tt.ok {
			t.Errorf("splitPipelineTaskName(%q) = (%q, %q, %v), want (%q, %q, %v)",
				tt.key, project, task, ok, tt.project, tt.task, tt.ok)
		}
	}
}

func TestConfigPipelineFindOrphansMalformedTask(t *testing.T) {
	cfg := &MonospaceConfig{
		GoModPrefix: "test.com",
		Projects:    map[string]string{"packages/test": "internal"},
		Pipeline: map[string]MonospaceConfigTask{
			" #build":           {Cmd: []string{"build"}}, // invalid project name " "
			"invalid#task#name": {Cmd: []string{"task"}},  // invalid task name
		},
	}
	appConfig = cfg

	report := cfg.PipelineFindOrphans()
	if len(report.Tasks) != 2 {
		t.Fatalf("PipelineFindOrphans(): want 2 orphan tasks, got %+v", report.Tasks)
	}
	if report.Tasks[0].Project != " " || report.Tasks[1].Project != "" {
		t.Errorf("PipelineFindOrphans(): unexpected projects, got %+v", report.Tasks)
	}
}

func TestConfigPipelineConfigWithoutOrphansReturnsCopy(t *testing.T) {
	cfg := &MonospaceConfig{
		GoModPrefix: "test.com",
		Projects:    map[string]string{"packages/test": "internal"},
		Pipeline:    map[string]MonospaceConfigTask{"packages/test#build": {Cmd: []string{"build"}}},
	}
	appConfig = cfg

	clone := cfg.PipelineConfigWithoutOrphans()
	if clone == cfg {
		t.Fatalf("PipelineConfigWithoutOrphans(): should always return a copy")
	}
	clone.Pipeline["packages/test#build"] = MonospaceConfigTask{Cmd: []string{"modified"}}
	if got := cfg.Pipeline["packages/test#build"].Cmd[0]; got != "build" {
		t.Errorf("PipelineConfigWithoutOrphans(): original config should not be modified, got %q", got)
	}
}

func TestProjectPipelineTasksWithoutConfig(t *testing.T) {
	previous := appConfig
	t.Cleanup(func() { appConfig = previous })
	appConfig = nil

	if _, err := ProjectPipelineTasks("packages/test"); err == nil {
		t.Errorf("ProjectPipelineTasks(): should return an error when config is not loaded")
	}
}

func TestConfigPipelineFixOrphans(t *testing.T) {
	cfg := &MonospaceConfig{
		GoModPrefix: "test.com",
		Projects:    map[string]string{"packages/test": "internal"},
		Pipeline: map[string]MonospaceConfigTask{
			"packages/ghost#build": {Cmd: []string{"build"}},
			"packages/test#ci":     {Cmd: []string{"ci"}, DependsOn: []string{"packages/ghost#build", "packages/test#lint"}},
			"packages/test#lint":   {Cmd: []string{"lint"}},
		},
	}
	appConfig = cfg

	removed := cfg.PipelineFixOrphans()
	if len(removed.Tasks) != 1 || len(removed.Deps) != 1 {
		t.Fatalf("PipelineFixOrphans(): unexpected report %+v", removed)
	}
	if _, ok := cfg.Pipeline["packages/ghost#build"]; ok {
		t.Errorf("PipelineFixOrphans(): orphan task should be removed")
	}
	ci := cfg.Pipeline["packages/test#ci"]
	if !reflect.DeepEqual(ci.DependsOn, []string{"packages/test#lint"}) {
		t.Errorf("PipelineFixOrphans(): orphan dependency should be removed, got %v", ci.DependsOn)
	}
	// idempotence
	if report := cfg.PipelineFindOrphans(); !report.IsEmpty() {
		t.Errorf("PipelineFixOrphans(): should be idempotent, still got %+v", report)
	}
}

func TestConfigRenameProject(t *testing.T) {
	cfg := &MonospaceConfig{
		GoModPrefix: "test.com",
		Projects:    map[string]string{"modules/old": "file://repo/old", "packages/other": "internal"},
		Aliases:     map[string]string{"old": "modules/old", "other": "packages/other"},
		Pipeline: map[string]MonospaceConfigTask{
			"modules/old#build":   {Cmd: []string{"build"}},
			"old#test":            {Cmd: []string{"test"}},
			"packages/other#ci":   {Cmd: []string{"ci"}, DependsOn: []string{"modules/old#build", "old#test", "packages/other#lint"}},
			"packages/other#lint": {Cmd: []string{"lint"}},
		},
	}
	appConfig = cfg

	if err := ConfigRenameProject("modules/old", "modules/new", false); err != nil {
		t.Fatalf("ConfigRenameProject(): unexpected error: %v", err)
	}
	// project renamed keeping its repo url
	if _, ok := cfg.Projects["modules/old"]; ok {
		t.Errorf("ConfigRenameProject(): old project should be removed from config")
	}
	if cfg.Projects["modules/new"] != "file://repo/old" {
		t.Errorf("ConfigRenameProject(): new project should keep the repo url, got %q", cfg.Projects["modules/new"])
	}
	// pipeline tasks renamed, none lost
	if _, ok := cfg.Pipeline["modules/old#build"]; ok {
		t.Errorf("ConfigRenameProject(): old task key should be renamed")
	}
	if _, ok := cfg.Pipeline["modules/new#build"]; !ok {
		t.Errorf("ConfigRenameProject(): task should be renamed to modules/new#build")
	}
	// aliased task keys are kept as is (the alias now targets the new name)
	if _, ok := cfg.Pipeline["old#test"]; !ok {
		t.Errorf("ConfigRenameProject(): aliased task key should be kept")
	}
	if cfg.Aliases["old"] != "modules/new" {
		t.Errorf("ConfigRenameProject(): alias should target the new project name, got %q", cfg.Aliases["old"])
	}
	// dependencies rewritten
	ci := cfg.Pipeline["packages/other#ci"]
	expectedDeps := []string{"modules/new#build", "old#test", "packages/other#lint"}
	if !reflect.DeepEqual(ci.DependsOn, expectedDeps) {
		t.Errorf("ConfigRenameProject(): unexpected dependsOn, want %v, got %v", expectedDeps, ci.DependsOn)
	}
	// no orphan introduced
	if report := cfg.PipelineFindOrphans(); !report.IsEmpty() {
		t.Errorf("ConfigRenameProject(): should not create orphans, got %+v", report)
	}
}

func TestConfigRenameProjectErrors(t *testing.T) {
	newConfig := func() *MonospaceConfig {
		return &MonospaceConfig{
			GoModPrefix: "test.com",
			Projects:    map[string]string{"modules/old": "internal", "packages/other": "internal"},
			Pipeline: map[string]MonospaceConfigTask{
				"modules/old#build": {Cmd: []string{"build"}},
				"packages/other#ci": {Cmd: []string{"ci"}, DependsOn: []string{"modules/old#build"}},
			},
		}
	}

	// unknown project
	cfg := newConfig()
	appConfig = cfg
	if err := ConfigRenameProject("modules/unknown", "modules/new", false); err == nil {
		t.Errorf("ConfigRenameProject(): should error on unknown project")
	}

	// target already exists
	cfg = newConfig()
	appConfig = cfg
	if err := ConfigRenameProject("modules/old", "packages/other", false); err == nil {
		t.Errorf("ConfigRenameProject(): should error when new name already exists")
	}

	// target already exists as an alias, it would be shadowed by the renamed project
	cfg = newConfig()
	cfg.Aliases = map[string]string{"modules/new": "packages/other"}
	appConfig = cfg
	if err := ConfigRenameProject("modules/old", "modules/new", false); err == nil {
		t.Errorf("ConfigRenameProject(): should error when new name is already used as an alias")
	} else if !strings.Contains(err.Error(), "already used as an alias") {
		t.Errorf("ConfigRenameProject(): expected 'already used as an alias' error, got: %v", err)
	}
	if _, ok := cfg.Projects["modules/old"]; !ok {
		t.Errorf("ConfigRenameProject(): failed rename should not mutate the config")
	}

	// pipeline task collision must error before any mutation
	cfg = newConfig()
	cfg.Projects["modules/new"] = "internal"
	cfg.Pipeline["modules/new#build"] = MonospaceConfigTask{Cmd: []string{"build"}}
	appConfig = cfg
	delete(cfg.Projects, "modules/new") // only the pipeline task collides, not the project name
	if err := ConfigRenameProject("modules/old", "modules/new", false); err == nil {
		t.Errorf("ConfigRenameProject(): should error on pipeline task collision")
	} else if !strings.Contains(err.Error(), "would overwrite") {
		t.Errorf("ConfigRenameProject(): expected 'would overwrite' error, got: %v", err)
	}
	if _, ok := cfg.Projects["modules/old"]; !ok {
		t.Errorf("ConfigRenameProject(): failed rename should not mutate the config")
	}
	if _, ok := cfg.Pipeline["modules/old#build"]; !ok {
		t.Errorf("ConfigRenameProject(): failed rename should not mutate the pipeline")
	}
}

// TestCacheIsEnabled checks the isCacheEnabled helper.
func TestCacheIsEnabled(t *testing.T) {
	if isCacheEnabled("") {
		t.Errorf("isCacheEnabled(%q) should be false", "")
	}
	if isCacheEnabled("disabled") {
		t.Errorf("isCacheEnabled(%q) should be false", "disabled")
	}
	if isCacheEnabled("false") {
		t.Errorf("isCacheEnabled(%q) should be false", "false")
	}
	if !isCacheEnabled("skip") {
		t.Errorf("isCacheEnabled(%q) should be true", "skip")
	}
	if !isCacheEnabled("restore") {
		t.Errorf("isCacheEnabled(%q) should be true", "restore")
	}
	if isCacheEnabled("foo") {
		t.Errorf("isCacheEnabled(%q) should be false", "foo")
	}
}

// TestCacheStringValues checks that cache field works as plain string.
func TestCacheStringValues(t *testing.T) {
	type tmp struct {
		Cache string `yaml:"cache"`
	}
	tests := []struct {
		input    string
		expected string
	}{
		{"skip", "skip"},
		{"restore", "restore"},
		{"disabled", "disabled"},
		{"", ""},
	}
	for _, tt := range tests {
		var out tmp
		data := []byte("cache: " + tt.input)
		if err := yaml.Unmarshal(data, &out); err != nil {
			t.Errorf("unmarshaling %q: unexpected error: %v", tt.input, err)
		}
		if out.Cache != tt.expected {
			t.Errorf("unmarshaling %q: expected %q, got %q", tt.input, tt.expected, out.Cache)
		}
	}
}
