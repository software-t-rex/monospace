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
			"myalias#build": {Cmd: []string{"make build"}},
			"myalias#test":  {Cmd: []string{"make test"}, DependsOn: []string{"myalias#build"}},
			"packages/other#ci": {Cmd: []string{"ci"}, DependsOn: []string{"myalias#test", "packages/other#lint"}},
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
			"myalias#build":     {Cmd: []string{"make build"}},
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
