package app

import (
	"errors"
	"os"
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
	Aliases     map[string]string                  `mapstructure:"projects_aliases,omitempty"`
	Pipeline    map[string]MonospaceConfigPipeline `mapstructure:"pipeline, omitempty"`
}

var dfltJSPM string = "^pnpm@7.27.0"
var dfltGoModPrfx string = "example.com"

var appConfig *MonospaceConfig

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, err
	} else if os.IsNotExist(err) { // not exists is not an error in this context
		return false, nil
	}
	return false, err
}

func ConfigSet(config *MonospaceConfig) {
	c := config
	if c.GoModPrefix == "" {
		c.GoModPrefix = dfltGoModPrfx
	}
	if c.JSPM == "" {
		c.JSPM = dfltJSPM
	}
	if c.Aliases == nil {
		c.Aliases = make(map[string]string)
	}
	if c.Projects == nil {
		c.Projects = make(map[string]string)
	}
	if c.Pipeline == nil {
		c.Pipeline = map[string]MonospaceConfigPipeline{}
	}
	appConfig = config
}

func ConfigIsLoaded() bool {
	return appConfig != nil
}

func ConfigGet() (*MonospaceConfig, error) {
	if !ConfigIsLoaded() {
		return nil, errors.New("app config not loaded")
	}
	return appConfig, nil
}

func ConfigInit(configPath string) (*MonospaceConfig, error) {
	if ConfigIsLoaded() {
		return nil, errors.New("config already loaded")
	}
	_, err := fileExists(configPath)
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
		ConfigSet(config)
	}
	return config, err
}

func ConfigSave() error {
	config, err := ConfigGet()
	if err != nil {
		return err
	}
	r := reflect.ValueOf(config).Elem()
	rType := r.Type()
	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)
		fName := strings.Split(rType.Field(i).Tag.Get("mapstructure"), ",")[0]
		val := f.Interface()
		viper.Set(fName, val)
	}
	return viper.WriteConfig()
}

func ConfigAddProjectAlias(projectName string, alias string, save bool) error {
	config, err := ConfigGet()
	if err != nil {
		return err
	}
	if config.Aliases[alias] != "" {
		return errors.New("alias " + alias + " already exists")
	}
	config.Aliases[alias] = projectName
	if save {
		return ConfigSave()
	}
	return err
}

func ConfigRemoveProjectAlias(alias string, save bool) error {
	config, err := ConfigGet()
	if err != nil {
		return err
	}
	if config.Aliases[alias] != "" {
		delete(config.Aliases, alias)
	}
	if save {
		return ConfigSave()
	}
	return nil
}

func ConfigAddProject(projectName string, repoUrl string, save bool) error {
	config, err := ConfigGet()
	if err != nil {
		return err
	}
	_, ok := config.Projects[projectName]
	if ok {
		return errors.New("project " + projectName + " already exists")
	}
	config.Projects[projectName] = repoUrl
	if save {
		return ConfigSave()
	}
	return err
}

func ConfigRemoveProject(projectName string, save bool) error {
	config, err := ConfigGet()
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
		return ConfigSave()
	}
	return err
}
