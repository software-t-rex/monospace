/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

// Package to read a package.json file into a struct
// with some utilities functions around package.json
package packageJson

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
)

type PackageJSONBug struct {
	Url   string `json:"url"`
	Email string `json:"email"`
}
type packageJSONPeerDependencyMeta struct {
	Optional bool `json:"optional"`
}

type PackageJSON struct {
	Name                 string                                   `json:"name,omitempty"`
	Version              string                                   `json:"version,omitempty"`
	Description          string                                   `json:"description,omitempty"`
	Keywords             []string                                 `json:"keywords,omitempty"`
	Homepage             string                                   `json:"homepage,omitempty"`
	Bugs                 PackageJSONBug                           `json:"bugs,omitempty"`
	License              string                                   `json:"license,omitempty"`
	Files                []string                                 `json:"files,omitempty"`
	Main                 string                                   `json:"main,omitempty"`
	Browser              string                                   `json:"browser,omitempty"`
	Directories          map[string]string                        `json:"directories,omitempty"`
	Scripts              map[string]string                        `json:"scripts,omitempty"`
	Dependencies         map[string]string                        `json:"dependencies,omitempty"`
	DevDependencies      map[string]string                        `json:"devDependencies,omitempty"`
	PeerDependencies     map[string]string                        `json:"peerDependencies,omitempty"`
	PeerDependenciesMeta map[string]packageJSONPeerDependencyMeta `json:"peerDependenciesMeta,omitempty"`
	OptionalDependencies map[string]string                        `json:"optionalDependencies,omitempty"`
	Engines              map[string]string                        `json:"engines,omitempty"`
	Os                   []string                                 `json:"os,omitempty"`
	Cpu                  []string                                 `json:"cpu,omitempty"`
	Private              bool                                     `json:"private,omitempty"`
	Workspaces           []string                                 `json:"workspaces,omitempty"`
	PackageManager       string                                   `json:"packageManager,omitempty"`

	// following properties can have multiple types or I was to lazy to dig in at the moment PR welcomes
	Author             any `json:"author,omitempty"`
	Contributors       any `json:"contributors,omitempty"`
	Funding            any `json:"funding,omitempty"`
	Bin                any `json:"bin,omitempty"`
	Man                any `json:"man,omitempty"`
	Repository         any `json:"repository,omitempty"`
	Config             any `json:"config,omitempty"`
	BundleDependencies any `json:"bundleDependencies,omitempty"`
	Overrides          any `json:"overrides,omitempty"`
	Resolutions        any `json:"resolutions,omitempty"` // resolutions is yarn overrides
	PublishConfig      any `json:"publishConfig,omitempty"`
	Pnpm               any `json:"pnpm,omitempty"` // specific to pnpm

	// following properties are not related to package.json but are used for some tooling
	file string     // this is the path of the package.json file used internally
	Mu   sync.Mutex `json:"-"`
}

func Read(path string) (p *PackageJSON, err error) {
	path, err = filepath.Abs(path)
	if err != nil {
		return
	}
	var raw []byte
	raw, err = os.ReadFile(path)
	if err != nil {
		return
	}
	err = json.Unmarshal(raw, &p)
	if err != nil {
		p.file = path
	}
	return p, err
}

// lookup for a given module name is the package dependencies
func (p *PackageJSON) GetDepencyInfoFor(moduleName string) (PackageJsonDepInfo, bool) {
	return newDependencyInfo(p, moduleName)
}

// return dependencies from dependencies, devDependencies, optionalDependecies
func (p *PackageJSON) GetMergedDependencies() map[string]string {
	res := make(map[string]string, len(p.Dependencies)+len(p.DevDependencies)+len(p.OptionalDependencies))
	for k, v := range p.OptionalDependencies {
		res[k] = v
	}
	for k, v := range p.DevDependencies {
		res[k] = v
	}
	for k, v := range p.Dependencies {
		res[k] = v
	}
	return res
}

// return tasks names in scipts
func (p *PackageJSON) GetAvailableTasks() []string {
	res := []string{}
	for k := range p.Scripts {
		res = append(res, k)
	}
	return res
}

// check a script task is set for that taskName
func (p *PackageJSON) HasTask(taskName string) bool {
	task, ok := p.Scripts[taskName]
	if ok && task != "" {
		return true
	}
	return false
}

// filter given dirList if they match a workspace defined in this package.json
// dirs in dirList should be either relative path to this package.json or absolute path
func (p *PackageJSON) FilterWorkspaceDirs(dirList []string, returnAbsPath bool) []string {
	var res []string
	var incPatterns []string
	var excPatterns []string
	rootDir := filepath.Dir(p.file)
	if len(p.Workspaces) == 0 || len(dirList) == 0 {
		return res
	}

	// separate includes form excludes
	for _, pattern := range p.Workspaces {
		if strings.HasPrefix(pattern, "!") {
			excPatterns = append(excPatterns, strings.TrimPrefix(pattern, "!"))
		} else {
			incPatterns = append(incPatterns, pattern)
		}
	}

	// filter included
	for _, dir := range dirList {
		dir = strings.TrimPrefix(dir, rootDir)
		for _, glob := range incPatterns {

			match, _ := doublestar.Match(glob, dir)
			if match {
				res = append(res, dir)
			}
		}
	}

	// remove excluded
	if len(res) > 0 && len(excPatterns) > 0 {
		for i, dir := range res {
			for _, glob := range excPatterns {
				match, _ := doublestar.Match(glob, dir)
				if match {
					res = append(res[i:], res[:i+1]...)
				}
			}
		}
	}

	if len(res) > 0 && returnAbsPath {
		for i, dir := range res {
			res[i] = filepath.Join(rootDir, dir)
		}
	}
	return res
}
