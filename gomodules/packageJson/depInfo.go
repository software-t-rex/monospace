/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package packageJson

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
)

type PackageJsonDepInfo struct {
	Name string
	// can be a version range or tarball, or git url as defined in package.json
	VersionRange string
	// one of dependecies, devDependecies, peerDependecies, optionalDependecies
	Key string
	// contains protocol if defined
	Protocol string
	// keep info on the package that require this
	FromName    string
	FromVersion string
	FromFile    string
}

func newDependencyInfo(p *PackageJSON, moduleName string) (i PackageJsonDepInfo, ok bool) {
	var depKey, depVal string
	if depVal, ok = p.Dependencies[moduleName]; ok {
		depKey = "dependencies"
	} else if depVal, ok = p.DevDependencies[moduleName]; ok {
		depKey = "devDependencies"
	} else if depVal, ok = p.OptionalDependencies[moduleName]; ok {
		depKey = "optionalDependencies"
	} else if depVal, ok = p.PeerDependencies[moduleName]; ok {
		depKey = "peerDependencies"
	}

	if !ok {
		return
	}
	i.Name = moduleName
	i.Key = depKey
	i.FromName = p.Name
	i.FromVersion = p.Version
	i.FromFile = p.file

	parts := strings.Split(depVal, ":")
	if len(parts) == 1 {
		i.VersionRange = parts[0]
	} else {
		i.Protocol = parts[0]
		i.VersionRange = strings.Join(parts[1:], ":")
	}
	return
}

type WorkSpacePackage struct {
	Name    string
	Version string
	Dir     string
}
type WorkSpacePackageList map[string]WorkSpacePackage

// workspacePkgs is a map of package name as keys and package version as values
func (i PackageJsonDepInfo) IsWorkspaceDep(workspaceRootDir string, workspacePkgs WorkSpacePackageList) (bool, error) {
	depPackage, ok := workspacePkgs[i.Name]
	if !ok {
		return false, nil
	}
	// workspace protocol => considered ok no more check
	if i.Protocol == "workspace" {
		return true, nil
	}
	// local protocols we should look at the resolved path
	if i.Protocol == "file" || i.Protocol == "link" || i.Protocol == "portal" {
		depPath := filepath.Join(i.FromFile, i.VersionRange)
		if depPath != depPackage.Dir {
			return false, errors.New("Package " + i.Name + " path does not match workspace package " + depPackage.Name)
		}
		return true, nil
	}
	// any other protocols than defaults should be considered remote and so not ok
	if i.Protocol != "" && i.Protocol != "npm" {
		return false, nil
	}
	// at this point We need to check the version range
	if i.VersionRange == "*" || i.VersionRange == "^" || i.VersionRange == "~" {
		return true, nil
	}
	constraint, errc := semver.NewConstraint(i.VersionRange)
	if errc != nil {
		// if we can't parse the version, we can't check the range and will assume it is ok but will return an error
		return true, errors.New("can't parse version range: " + i.VersionRange + " for package " + i.Name)
	}
	version, errv := semver.NewVersion(depPackage.Version)
	if errv != nil {
		// if we can't parse the version, we can't check the range and will assume it is ok but will return an error
		return true, errors.New("can't parse version: " + depPackage.Version + " for package " + depPackage.Name)
	}
	return constraint.Check(version), nil
}
