package utils

import (
	"path/filepath"
)

func WriteTemplateGitinore(path string) error {
	return WriteFile(filepath.Join(path, "/", ".gitignore"), "node_modules\n.vscode\ndist\ncoverage\n")
}
