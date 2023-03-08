package app

import (
	"errors"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

type MonospaceConfigPipeline struct {
	DependsOn  []string `json:"name,omitempty"`
	Env        []string `json:"env,omitempty"`
	Outputs    []string `json:"outputs,omitempty"`
	Inputs     []string `json:"inputs,omitempty"`
	Cache      bool     `json:"cache,omitempty"`
	OutputMode string   `json:"outputMode,omitempty"`
	Persistent bool     `json:"peristent,omitempty"`
	Cmd        []string `json:"cmd,omitempty"`
}
type MonospaceConfig struct {
	GoModPrefix string                             `json:"go_mod_prefix,omitempty"`
	JSPM        string                             `json:"js_package_manager,omitempty"`
	Projects    map[string]string                  `json:"projects,omitempty"`
	Aliases     map[string]string                  `json:"projects_aliases,omitempty"`
	Pipeline    map[string]MonospaceConfigPipeline `json:"pipeline,omitempty"`
	configPath  string
	root        string
}

var dfltJSPM string = "^pnpm@7.27.0"
var dfltGoModPrfx string = "example.com"

var appConfig *MonospaceConfig

var ErrNotLoadedConfig = errors.New("config not loaded")

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, err
	} else if os.IsNotExist(err) { // not exists is not an error in this context
		return false, nil
	}
	return false, err
}
func writeFile(filePath string, body []byte) error {
	return os.WriteFile(filePath, body, 0640)
}

func configSet(config *MonospaceConfig) {
	c := config
	if c.GoModPrefix == "" {
		c.GoModPrefix = dfltGoModPrfx
	}
	if c.JSPM == "" {
		c.JSPM = dfltJSPM
	}
	appConfig = config
}

func ConfigIsLoaded() bool {
	return appConfig != nil
}

func ConfigGet() (*MonospaceConfig, error) {
	if !ConfigIsLoaded() {
		return nil, ErrNotLoadedConfig
	}
	return appConfig, nil
}

func ConfigRead(configPath string) (*MonospaceConfig, error) {
	_, err := fileExists(configPath)
	if err != nil {
		return nil, err
	}
	var raw []byte
	var config *MonospaceConfig
	raw, err = os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(raw, &config)
	config.configPath = configPath
	config.root = filepath.Dir(configPath)
	return config, err
}

func ConfigInit(configPath string) error {
	if ConfigIsLoaded() {
		return errors.New("config already loaded")
	}
	config, err := ConfigRead(configPath)
	if err == nil {
		configSet(config)
	}
	return err
}

func ConfigSave() error {
	config, err := ConfigGet()
	if err != nil {
		return err
	}
	if config.configPath == "" {
		return errors.New("missing a configPath to save to")
	}
	var raw []byte

	raw, err = yaml.Marshal(config)
	if err != nil {
		return err
	}
	raw = append([]byte("# yaml-language-server: $schema=./apps/monospace/schemas/monospace.schema.json\n"), raw...)

	return writeFile(config.configPath, raw)
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

func PopulateEnv(env map[string]string) error {
	if !ConfigIsLoaded() {
		return ErrNotLoadedConfig
	}
	env["ROOT"] = appConfig.root
	env["VERSION"] = Version
	env["JSPM"] = appConfig.JSPM
	env["GOPREFIX"] = appConfig.GoModPrefix

	for k, v := range env {
		if err := os.Setenv("MONOSPACE_"+k, v); err != nil {
			return err
		}
	}
	return nil
}
