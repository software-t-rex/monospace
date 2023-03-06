package utils

import (
	"errors"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type MonospaceConfigPipeline struct {
	DependsOn  []string `mapstructure:"name,omitempty"`
	Env        []string `mapstructure:"env,omitempty"`
	Outputs    []string `mapstructure:"outputs,omitempty"`
	Inputs     []string `mapstructure:"inputs,omitempty"`
	Cache      bool     `mapstructure:"cache,omitempty"`
	OutputMode string   `mapstructure:",omitempty"`
	Persistent bool     `mapstructure:"peristent,omitempty"`
}
type MonospaceConfig struct {
	GoModPrefix string                             `mapstructure:"go_mod_prefix,omitempty"`
	JSPM        string                             `mapstructure:"js_package_manager,omitempty"`
	Projects    map[string]string                  `mapstructure:"projects, omitempty"`
	Aliases     map[string]string                  `mapstructure:"project_aliases,omitempty"`
	Pipeline    map[string]MonospaceConfigPipeline `mapstructure:"pipeline, omitempty"`
}

var dfltJSPM string = "^pnpm@7.27.0"
var dfltGoModPrfx string = "example.com"

var appConfig *MonospaceConfig

func AppConfigSet(config *MonospaceConfig) {
	c := config
	if c.GoModPrefix == "" {
		c.GoModPrefix = dfltGoModPrfx
	}
	if c.JSPM == "" {
		c.JSPM = dfltJSPM
	}
	appConfig = config
}

func AppConfigIsLoaded() bool {
	return appConfig != nil
}

func AppConfigGet() (*MonospaceConfig, error) {
	if !AppConfigIsLoaded() {
		return nil, errors.New("app config not loaded")
	}
	return appConfig, nil
}

func AppConfigInit(configPath string) (*MonospaceConfig, error) {
	if AppConfigIsLoaded() {
		return nil, errors.New("config already loaded")
	}
	_, err := FileExists(configPath)
	if err != nil {
		return nil, err
	}
	viper.SetConfigFile(configPath)
	viper.SetDefault("js_package_manager", dfltJSPM)
	viper.SetDefault("go_mod_prefix", dfltGoModPrfx)
	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	var config *MonospaceConfig
	err = viper.Unmarshal(&config)
	if err == nil {
		AppConfigSet(config)
	}
	return config, err
}

func AppConfigSave() error {
	config, err := AppConfigGet()
	if err != nil {
		return err
	}
	r := reflect.ValueOf(config).Elem()
	rType := r.Type()
	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)
		fName := strings.Split(rType.Field(i).Tag.Get("mapstructure"), ",")[0]
		viper.Set(fName, f.Interface())
	}
	return viper.WriteConfig()
}

// ///////////////////////////////
func AppConfigAddProjectAlias(projectName string, alias string, save bool) error {
	config, err := AppConfigGet()
	if err != nil {
		return err
	}
	if config.Aliases[alias] != "" {
		return errors.New("alias " + alias + " already exists")
	}
	config.Aliases[alias] = projectName
	if save {
		return AppConfigSave()
	}
	return err
}

func AppConfigRemoveProjectAlias(alias string, save bool) error {
	config, err := AppConfigGet()
	if err != nil {
		return err
	}
	if config.Aliases[alias] != "" {
		delete(config.Aliases, alias)
	}
	if save {
		return AppConfigSave()
	}
	return nil
}

func AppConfigAddProject(projectName string, repoUrl string, save bool) error {
	config, err := AppConfigGet()
	if err != nil {
		return err
	}
	_, ok := config.Projects[projectName]
	if ok {
		return errors.New("project " + projectName + " already exists")
	}
	config.Projects[projectName] = repoUrl
	if save {
		return AppConfigSave()
	}
	return err
}

func AppConfigRemoveProject(projectName string, save bool) error {
	config, err := AppConfigGet()
	if err != nil {
		return err
	}
	delete(config.Projects, projectName)
	// lookup for aliases to remove
	for k, v := range config.Aliases {
		if v == projectName {
			delete(config.Aliases, k)
			continue
		}
	}
	if save {
		return AppConfigSave()
	}
	return err
}
