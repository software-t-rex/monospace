/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

type MonospaceConfigPipeline struct {
	DependsOn  []string `yaml:"dependsOn,omitempty,flow"`
	Env        []string `yaml:"env,omitempty,flow"`
	Outputs    []string `yaml:"outputs,omitempty"`
	Inputs     []string `yaml:"inputs,omitempty"`
	Cache      bool     `yaml:"cache,omitempty"`
	OutputMode string   `yaml:"output_mode,omitempty"`
	Persistent bool     `yaml:"persistent,omitempty"`
	Cmd        []string `yaml:"cmd,omitempty,flow"`
}
type MonospaceConfig struct {
	GoModPrefix         string                             `yaml:"go_mod_prefix,omitempty"`
	JSPM                string                             `yaml:"js_package_manager,omitempty"`
	PreferredOutputMode string                             `yaml:"preferred_output_mode,omitempty"`
	Projects            map[string]string                  `yaml:"projects,omitempty"`
	Aliases             map[string]string                  `yaml:"projects_aliases,omitempty"`
	Pipeline            map[string]MonospaceConfigPipeline `yaml:"pipeline,omitempty"`
	configPath          string
	root                string
}

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

func (c *MonospaceConfig) GetPath() string {
	return c.configPath
}
func (c *MonospaceConfig) GetRoot() string {
	return c.root
}
func (c *MonospaceConfig) Save() error {
	return ConfigSave()
}

func configSet(config *MonospaceConfig) {
	if config == nil {
		panic("configSet called with nil config")
	}
	c := config
	if c.GoModPrefix == "" {
		c.GoModPrefix = dfltGoModPrfx
	}
	if c.JSPM == "" {
		c.JSPM = dfltJSPM
	}
	if c.PreferredOutputMode == "" {
		c.PreferredOutputMode = dfltPreferredOutputMode
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
	raw = append([]byte("# yaml-language-server: $schema=https://raw.githubusercontent.com/software-t-rex/monospace/main/apps/monospace/schemas/monospace.schema.json\n"), raw...)

	return writeFile(config.configPath, raw)
}

func ConfigAddProjectAlias(projectName string, alias string, save bool) error {
	config, err := ConfigGet()
	if err != nil {
		return err
	}
	aliasOk, err := regexp.MatchString("^[a-zA-Z_][a-zA-Z0-9_-]*(\\/[a-zA-Z_][a-zA-Z0-9_-]*)*$", alias)
	if !aliasOk {
		return fmt.Errorf("invalid alias name %s", alias)
	} else if config.Projects == nil || config.Projects[projectName] == "" {
		return fmt.Errorf("unknown project %s", projectName)
	}
	if config.Aliases == nil {
		config.Aliases = make(map[string]string)
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

func ConfigAddOrUpdateProject(projectName string, repoUrl string, save bool) error {
	config, err := ConfigGet()
	if err != nil {
		return err
	}
	config.Projects[projectName] = repoUrl
	if save {
		return ConfigSave()
	}
	return err
}
func ConfigAddProject(projectName string, repoUrl string, save bool) error {
	config, err := ConfigGet()
	if err != nil {
		return err
	}
	if config.Projects == nil {
		config.Projects = make(map[string]string)
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
	if env == nil {
		env = make(map[string]string)
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
