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
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// isCacheEnabled returns true if the cache mode is "skip" or "restore".
// Empty or any other value is considered "disabled".
func isCacheEnabled(mode string) bool {
	return mode == "skip" || mode == "restore"
}

type MonospaceConfigTask struct {
	Description     string            `yaml:"description,omitempty"`
	Cmd             []string          `yaml:"cmd,omitempty,flow"`
	DependsOn       []string          `yaml:"dependsOn,omitempty,flow"`
	Env             map[string]string `yaml:"env,omitempty,flow"`
	Persistent      bool              `yaml:"persistent,omitempty"`
	OutputMode      string            `yaml:"output_mode,omitempty"`
	Cache           string            `yaml:"cache,omitempty"`             // "skip" | "restore" | "" (disabled)
	CacheStrategy   string            `yaml:"cache_strategy,omitempty"`    // "content" | "mtime" | ""
	CacheMaxEntries int               `yaml:"cache_max_entries,omitempty"` // 0 = use global default
	Inputs          []string          `yaml:"inputs,omitempty"`
	Outputs         []string          `yaml:"outputs,omitempty"`
}
type MonospaceConfig struct {
	GoModPrefix         string                         `yaml:"go_mod_prefix,omitempty"`
	JSPM                string                         `yaml:"js_package_manager,omitempty"`
	PreferredOutputMode string                         `yaml:"preferred_output_mode,omitempty"`
	CacheMaxEntries     int                            `yaml:"cache_max_entries,omitempty"` // global default, 0 = use DefaultCacheMaxEntries
	Projects            map[string]string              `yaml:"projects,omitempty"`
	Aliases             map[string]string              `yaml:"projects_aliases,omitempty"`
	Pipeline            map[string]MonospaceConfigTask `yaml:"pipeline,omitempty"`
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
func (c *MonospaceConfig) GetDir() string {
	return filepath.Dir(c.configPath)
}
func (c *MonospaceConfig) GetRoot() string {
	return c.root
}

// returns a map indexed by project name with the alias as value
func (c *MonospaceConfig) GetProjectsAliases() map[string]string {
	projectAliases := make(map[string]string)
	if c.Aliases != nil {
		for alias, projectName := range c.Aliases {
			projectAliases[projectName] = alias
		}
	}
	return projectAliases
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
		c.GoModPrefix = DfltGoModPrfx
	}
	if c.JSPM == "" {
		c.JSPM = DfltJSPM
	}
	if c.PreferredOutputMode == "" {
		c.PreferredOutputMode = DfltPreferredOutputMode
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
	if config == nil {
		config = &MonospaceConfig{}
	}
	config.configPath = configPath
	config.root = filepath.Dir(filepath.Dir(configPath))
	return config, err
}

func ConfigLoadNoCheck(configPath string) error {
	config, err := ConfigRead(configPath)
	if err == nil {
		configSet(config)
	}
	return err
}

func ConfigLoad(configPath string) error {
	if ConfigIsLoaded() {
		return errors.New("config already loaded")
	}
	return ConfigLoadNoCheck(configPath)
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
	if alias == "root" {
		return fmt.Errorf("alias %s is reserved", alias)
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
	projectName := config.Aliases[alias]
	if projectName == "" {
		if save {
			return ConfigSave()
		}
		return nil
	}
	// update pipeline: rename task keys and dependsOn entries that use the alias
	// (errors are reported before any mutation, the alias is kept on collision)
	if _, _, errRemap := config.remapPipelinePrefixes(map[string]string{alias: projectName}); errRemap != nil {
		return fmt.Errorf("cannot remove alias %q: %w", alias, errRemap)
	}
	delete(config.Aliases, alias)
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
	if config.Projects == nil {
		config.Projects = make(map[string]string)
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
	// remove pipeline tasks belonging to the project and the references to them.
	// prefixes have to be collected before removing the project aliases.
	config.removePipelineRefs(config.projectPipelinePrefixes(projectName))
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
	return nil
}

// ConfigRenameProject renames a project, its pipeline tasks and every reference to them,
// so no pipeline task is lost. Aliases pointing to the project are updated too.
func ConfigRenameProject(oldName, newName string, save bool) error {
	config, err := ConfigGet()
	if err != nil {
		return err
	}
	repoUrl, exists := config.Projects[oldName]
	if !exists {
		return fmt.Errorf("unknown project %s", oldName)
	}
	if _, exists := config.Projects[newName]; exists {
		return fmt.Errorf("project %s already exists", newName)
	}
	// an alias with the target name would be shadowed by the renamed project
	if _, exists := config.Aliases[newName]; exists {
		return fmt.Errorf("cannot rename project %s to %s: %s is already used as an alias", oldName, newName, newName)
	}
	// rename pipeline task keys and their dependsOn entries,
	// errors are reported before any mutation to avoid losing tasks
	if _, _, errRemap := config.remapPipelinePrefixes(map[string]string{oldName: newName}); errRemap != nil {
		return fmt.Errorf("cannot rename project %q: %w", oldName, errRemap)
	}
	config.Projects[newName] = repoUrl
	delete(config.Projects, oldName)
	// aliases targeting the project have to target the new project name,
	// pipeline tasks using an alias prefix are left untouched since they still resolve to the project
	for alias, target := range config.Aliases {
		if target == oldName {
			config.Aliases[alias] = newName
		}
	}
	if save {
		return ConfigSave()
	}
	return nil
}

// this regexp mirrors the task name syntax accepted by tasks.ParseTaskName
var pipelineTaskNameRegexp = regexp.MustCompile("^(?:([^#]+)#)?([^#]+)$")

// splitPipelineTaskName splits a pipeline task key into its project prefix and task name.
// A missing project prefix is reported as an empty string meaning "*" (ie: "build", "#build").
// It mirrors tasks.ParseTaskName exactly, including its handling of whitespace.
func splitPipelineTaskName(key string) (project string, task string, ok bool) {
	match := pipelineTaskNameRegexp.FindStringSubmatch(strings.TrimPrefix(key, "#"))
	if match == nil {
		return "", "", false
	}
	return match[1], match[2], true
}

// pipelineProjectPrefix normalizes a project prefix, an empty prefix meaning "*"
func pipelineProjectPrefix(prefix string) string {
	if prefix == "" {
		return "*"
	}
	return prefix
}

// pipelineProjectIsValid returns true if the prefix refers to a known project or a valid alias,
// or is one of the reserved "*" and "root" prefixes.
// It mirrors tasks.getStandardProjectName() validity rules.
func (c *MonospaceConfig) pipelineProjectIsValid(prefix string) bool {
	prefix = pipelineProjectPrefix(prefix)
	if prefix == "*" || prefix == "root" {
		return true
	}
	if _, ok := c.Projects[prefix]; ok {
		return true
	}
	if aliased, ok := c.Aliases[prefix]; ok {
		_, ok = c.Projects[aliased]
		return ok
	}
	return false
}

// projectPipelinePrefixes returns the project name plus every alias referring to it
func (c *MonospaceConfig) projectPipelinePrefixes(projectName string) map[string]bool {
	prefixes := map[string]bool{projectName: true}
	for alias, target := range c.Aliases {
		if target == projectName {
			prefixes[alias] = true
		}
	}
	return prefixes
}

// ProjectPipelineTasks returns the sorted list of pipeline task keys belonging to the given project
// (either using its name or one of its aliases as prefix)
func (c *MonospaceConfig) ProjectPipelineTasks(projectName string) []string {
	prefixes := c.projectPipelinePrefixes(projectName)
	pipelineTasks := make([]string, 0)
	for key := range c.Pipeline {
		project, _, ok := splitPipelineTaskName(key)
		if ok && prefixes[pipelineProjectPrefix(project)] {
			pipelineTasks = append(pipelineTasks, key)
		}
	}
	sort.Strings(pipelineTasks)
	return pipelineTasks
}

// ProjectPipelineTasks returns the sorted list of pipeline task keys belonging to the given project
// from the loaded config
func ProjectPipelineTasks(projectName string) ([]string, error) {
	config, err := ConfigGet()
	if err != nil {
		return nil, err
	}
	return config.ProjectPipelineTasks(projectName), nil
}

// remapPipelinePrefixes renames pipeline task keys and dependsOn entries using one of the given
// prefixes. Errors are reported before any mutation to avoid losing tasks on collision.
// Dependencies without '#' are relative to their own project, they are left untouched.
func (c *MonospaceConfig) remapPipelinePrefixes(prefixMap map[string]string) (changedKeys, changedDeps []string, err error) {
	renamedKeys := map[string]string{}
	for key := range c.Pipeline {
		project, task, ok := splitPipelineTaskName(key)
		if !ok {
			continue
		}
		newPrefix, found := prefixMap[pipelineProjectPrefix(project)]
		if !found {
			continue
		}
		if newKey := newPrefix + "#" + task; newKey != key {
			renamedKeys[key] = newKey
		}
	}
	// check for collisions before mutating anything
	for oldKey, newKey := range renamedKeys {
		if _, exists := c.Pipeline[newKey]; exists {
			return nil, nil, fmt.Errorf("renaming pipeline task %q to %q would overwrite an existing task", oldKey, newKey)
		}
	}
	for oldKey, newKey := range renamedKeys {
		c.Pipeline[newKey] = c.Pipeline[oldKey]
		delete(c.Pipeline, oldKey)
		changedKeys = append(changedKeys, oldKey)
	}
	for key, task := range c.Pipeline {
		changed := false
		for i, dep := range task.DependsOn {
			if !strings.Contains(dep, "#") {
				continue
			}
			project, depTask, ok := splitPipelineTaskName(dep)
			if !ok {
				continue
			}
			newPrefix, found := prefixMap[pipelineProjectPrefix(project)]
			if !found {
				continue
			}
			task.DependsOn[i] = newPrefix + "#" + depTask
			changedDeps = append(changedDeps, key+" -> "+dep)
			changed = true
		}
		if changed {
			c.Pipeline[key] = task
		}
	}
	sort.Strings(changedKeys)
	sort.Strings(changedDeps)
	return
}

// removePipelineRefs removes pipeline tasks belonging to one of the given prefixes and filters out
// the dependsOn entries referring to them. Dependencies without '#' are relative to their own
// project, they are left untouched.
func (c *MonospaceConfig) removePipelineRefs(prefixes map[string]bool) (removedTasks, cleanedDeps []string) {
	if len(prefixes) == 0 {
		return
	}
	for key := range c.Pipeline {
		project, _, ok := splitPipelineTaskName(key)
		if ok && prefixes[pipelineProjectPrefix(project)] {
			delete(c.Pipeline, key)
			removedTasks = append(removedTasks, key)
		}
	}
	for key, task := range c.Pipeline {
		if len(task.DependsOn) == 0 {
			continue
		}
		kept := make([]string, 0, len(task.DependsOn))
		changed := false
		for _, dep := range task.DependsOn {
			if strings.Contains(dep, "#") {
				project, _, ok := splitPipelineTaskName(dep)
				if ok && prefixes[pipelineProjectPrefix(project)] {
					changed = true
					cleanedDeps = append(cleanedDeps, key+" -> "+dep)
					continue
				}
			}
			kept = append(kept, dep)
		}
		if changed {
			task.DependsOn = kept
			c.Pipeline[key] = task
		}
	}
	sort.Strings(removedTasks)
	sort.Strings(cleanedDeps)
	return
}

// PipelineOrphanRef describes either a pipeline task belonging to an unknown project or a
// dependency referring to such a task.
type PipelineOrphanRef struct {
	Task    string // the pipeline task holding the reference
	Project string // the unknown project prefix
	Dep     string // the offending dependency (empty for orphan tasks)
	IsDep   bool
}

// PipelineOrphanReport lists orphan pipeline tasks and dependencies referring to them
type PipelineOrphanReport struct {
	Tasks []PipelineOrphanRef
	Deps  []PipelineOrphanRef
}

// IsEmpty returns true when no orphan was found
func (r PipelineOrphanReport) IsEmpty() bool {
	return len(r.Tasks) == 0 && len(r.Deps) == 0
}

// PipelineFindOrphans reports pipeline tasks belonging to an unknown project or alias and the
// dependencies referring to them. The config is left untouched.
func (c *MonospaceConfig) PipelineFindOrphans() PipelineOrphanReport {
	report := PipelineOrphanReport{}
	orphanTasks := map[string]bool{}
	for key := range c.Pipeline {
		project, _, ok := splitPipelineTaskName(key)
		if !ok {
			report.Tasks = append(report.Tasks, PipelineOrphanRef{Task: key})
			orphanTasks[key] = true
			continue
		}
		prefix := pipelineProjectPrefix(project)
		if !c.pipelineProjectIsValid(prefix) {
			report.Tasks = append(report.Tasks, PipelineOrphanRef{Task: key, Project: prefix})
			orphanTasks[key] = true
		}
	}
	seenDeps := map[string]bool{}
	for key, task := range c.Pipeline {
		if orphanTasks[key] {
			continue
		}
		for _, dep := range task.DependsOn {
			if !strings.Contains(dep, "#") {
				continue
			}
			// a dependency may be declared more than once, report it only once
			depKey := key + "\x00" + dep
			if seenDeps[depKey] {
				continue
			}
			seenDeps[depKey] = true
			project, _, ok := splitPipelineTaskName(dep)
			if !ok {
				report.Deps = append(report.Deps, PipelineOrphanRef{Task: key, Dep: dep, IsDep: true})
				continue
			}
			prefix := pipelineProjectPrefix(project)
			if !c.pipelineProjectIsValid(prefix) {
				report.Deps = append(report.Deps, PipelineOrphanRef{Task: key, Project: prefix, Dep: dep, IsDep: true})
			}
		}
	}
	sort.Slice(report.Tasks, func(i, j int) bool { return report.Tasks[i].Task < report.Tasks[j].Task })
	sort.Slice(report.Deps, func(i, j int) bool {
		if report.Deps[i].Task == report.Deps[j].Task {
			return report.Deps[i].Dep < report.Deps[j].Dep
		}
		return report.Deps[i].Task < report.Deps[j].Task
	})
	return report
}

// PipelineFixOrphans removes orphan pipeline tasks and the dependencies referring to them,
// it returns the report of what has been removed.
func (c *MonospaceConfig) PipelineFixOrphans() PipelineOrphanReport {
	report := c.PipelineFindOrphans()
	for _, orphan := range report.Tasks {
		delete(c.Pipeline, orphan.Task)
	}
	for _, orphan := range report.Deps {
		c.PipelineRemoveDep(orphan.Task, orphan.Dep)
	}
	return report
}

// PipelineConfigWithoutOrphans returns a copy of the config where orphan pipeline tasks and
// dependencies have been removed. It is used to run the remaining pipeline checks on a config
// which can be standardized, without mutating the original config nor exiting.
func (c *MonospaceConfig) PipelineConfigWithoutOrphans() *MonospaceConfig {
	clone := *c
	clone.Pipeline = make(map[string]MonospaceConfigTask, len(c.Pipeline))
	for key, task := range c.Pipeline {
		clone.Pipeline[key] = task
	}
	clone.PipelineFixOrphans()
	return &clone
}

// PipelineRemoveDep removes the given dependency from a pipeline task
func (c *MonospaceConfig) PipelineRemoveDep(taskKey, dep string) bool {
	task, ok := c.Pipeline[taskKey]
	if !ok {
		return false
	}
	kept := make([]string, 0, len(task.DependsOn))
	removed := false
	for _, taskDep := range task.DependsOn {
		if taskDep == dep {
			removed = true
			continue
		}
		kept = append(kept, taskDep)
	}
	if !removed {
		return false
	}
	task.DependsOn = kept
	c.Pipeline[taskKey] = task
	return true
}

// Add given env vars to the current env prefixing them with MONOSPACE_
// It also add MONOSPACE_ROOT, MONOSPACE_VERSION, MONOSPACE_JSPM, MONOSPACE_GOPREFIX to the env
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
