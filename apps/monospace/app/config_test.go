package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

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

	err := ConfigInit(configPath)
	if err == nil {
		t.Errorf("ConfigInit(): should error on unexisiting config file")
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
		t.Fatalf(err.Error())
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
	err = ConfigInit(configPath)
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

	err := ConfigAddProject("packages/test", "whathever", false)
	if err == nil {
		t.Errorf("ConfigAddProject(): should report error on adding existing project, err: %v", err)
	}

	err = ConfigAddProject("test/toto", "local", false)
	if appConfig.Projects["test/toto"] != "local" {
		t.Errorf("ConfigAddProject(): shoud add project to config want: %v, got: %v, err: %v", "local", appConfig.Projects["test/toto"], err)
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
		t.Errorf("ConfigAddProjectAlias(): shoud add project alias to config want: %v, got: %v, err: %v", "local", appConfig.Projects["test/toto"], err)
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
		t.Errorf("ConfigRemoveProject(): shoud remove project from config")
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
		t.Errorf("ConfigRemoveProjectAlias(): shoud remove project alias from config")
	}

	err = ConfigRemoveProjectAlias("unknownAlias", false)
	if err != nil {
		t.Errorf("Should not error on removing unknown project alias, err: %v", err)
	}

}
