package utils

import (
	"os"
	"path/filepath"
)

func PackageGetPath(packageName string) string {
	return filepath.Join(MonospaceGetRoot(), "/packages/", filepath.Clean(packageName))
}

func PackageCreate(packageName string) error {
	// create a new directory in the packages folder
	return os.MkdirAll(PackageGetPath(packageName), 0750)
}

func PackageChdir(packageName string) error {
	return os.Chdir(PackageGetPath(packageName))
}
