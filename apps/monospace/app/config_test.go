package app

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

var sampleConfig = &MonospaceConfig{
	GoModPrefix: "test.com",
	JSPM:        "yarn@xxx",
	Projects:    map[string]string{"packages/test": "internal"},
	Aliases:     map[string]string{"test": "packages/test"},
}

func TestAppConfig(t *testing.T) {
	// config should not be loaded at first
	if ConfigIsLoaded() {
		t.Errorf("AppConfigIsLoaded(): config should not be loaded before call to AppConfigSet")
	}

	got, err := ConfigGet()
	if err == nil {
		t.Errorf("AppConfigGet(): config get should nor return a config before call to AppConfigSet, got %v", got)
	}

	config := &MonospaceConfig{}
	ConfigSet(config)
	if !(config.JSPM == dfltJSPM && config.GoModPrefix == dfltGoModPrfx) {
		t.Errorf("AppConfigSet(): doesn't set default values correctly")
	}

	if !ConfigIsLoaded() {
		t.Errorf("AppConfigIsLoaded(): config should be loaded after call to AppConfigSet")
	}

	got, err = ConfigGet()
	if err != nil {
		t.Errorf("AppConfigGet(): should return loaded config")
	}

	if !reflect.DeepEqual(got, config) {
		t.Errorf("AppConfigGet(): bad returned value want: %v, got: %v", config, got)
	}
}

func TestAppConfigInitAndSave(t *testing.T) {
	//reset config
	appConfig = nil

	var configPath = filepath.Join(t.TempDir(), "config.yml")
	viper.SetConfigFile(configPath)

	got, err := ConfigInit(configPath)
	if err == nil {
		t.Errorf("AppConfigInit(): should error on unexisiting config file, got %v", got)
	}

	err = ConfigSave()
	if err == nil {
		t.Errorf("AppConfigSave(): should error on unloaded config")
	}

	appConfig = sampleConfig

	err = ConfigSave()
	if err != nil {
		t.Errorf("AppConfigSave(): should save the config without error, %v", err)
	}
	res, err := exec.Command("cat", configPath).CombinedOutput()
	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Println(string(res))
	//reset config
	appConfig = nil

	got, err = ConfigGet()
	if err == nil || got != nil {
		t.Errorf("AppConfigGet(): should not return a config at this point")
	}

	got, err = ConfigInit(configPath)
	if !reflect.DeepEqual(got, sampleConfig) {
		t.Errorf("AppConfigInit(): return different config than expected: %v, want: %v, err: %v", got, sampleConfig, err)
	}
}

func TestAppConfigAddProject(t *testing.T) {
	appConfig = sampleConfig

	err := ConfigAddProject("packages/test", "whathever", false)
	if err == nil {
		t.Errorf("AppConfigAddProject(): should report error on adding existing project, err: %v", err)
	}

	err = ConfigAddProject("test/toto", "local", false)
	if appConfig.Projects["test/toto"] != "local" {
		t.Errorf("AppConfigAddProject(): shoud add project to config want: %v, got: %v, err: %v", "local", appConfig.Projects["test/toto"], err)
	}

}
func TestAppConfigAddProjectAlias(t *testing.T) {
	appConfig = sampleConfig

	err := ConfigAddProjectAlias("packages/test", "test", false)
	if err == nil {
		t.Errorf("AppConfigAddProjectAlias(): should report error on existing project alias, err: %v", err)
	}

	err = ConfigAddProjectAlias("packages/test", "aliasname", false)
	if appConfig.Aliases["aliasname"] != "packages/test" {
		t.Errorf("AppConfigAddProjectAlias(): shoud add project alias to config want: %v, got: %v, err: %v", "local", appConfig.Projects["test/toto"], err)
	}
}

func TestAppConfigRemoveProject(t *testing.T) {
	appConfig = sampleConfig
	appConfig.Projects["test/removableProject"] = "internal"
	appConfig.Aliases["alias1"] = "test/removableProject"
	appConfig.Aliases["alias2"] = "test/removableProject"
	err := ConfigRemoveProject("test/removableProject", false)
	if err != nil {
		t.Errorf("AppConfigRemoveProject(): Should not error on removing existing project, err: %v", err)
	}
	p, ok := appConfig.Projects["test/removableProject"]
	if ok || p != "" {
		t.Errorf("AppConfigRemoveProject(): shoud remove project from config")
	}

	err = ConfigRemoveProject("test/unknownProject", false)
	if err != nil {
		t.Errorf("AppConfigRemoveProject(): Should not error on removing unknown project, err: %v", err)
	}

	if appConfig.Aliases["alias1"] != "" || appConfig.Aliases["alias2"] != "" {
		t.Errorf("AppConfigRemoveProject(): correctly remove all aliases for projects")
	}

}

func TestAppConfigRemoveProjectAlias(t *testing.T) {

	appConfig.Aliases["myalias"] = "packages/test"
	err := ConfigRemoveProjectAlias("myalias", false)
	if err != nil {
		t.Errorf("Should not error on removing existing project alias, err: %v", err)
	}
	p, ok := appConfig.Aliases["myalias"]
	if ok || p != "" {
		t.Errorf("AppConfigRemoveProjectAlias(): shoud remove project alias from config")
	}

	err = ConfigRemoveProjectAlias("unknownAlias", false)
	if err != nil {
		t.Errorf("Should not error on removing unknown project alias, err: %v", err)
	}

}
