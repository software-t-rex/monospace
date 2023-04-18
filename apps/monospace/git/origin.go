package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/software-t-rex/monospace/utils"
)

func GitGetOrigin(directory string) (string, error) {
	cmd := exec.Command("git", "-C", directory, "remote", "get-url", "origin")
	var errMsg bytes.Buffer
	cmd.Stderr = &errMsg
	origin, err := cmd.Output()
	if err == nil && string(origin) != "" {
		return strings.TrimSpace(string(origin)), nil
	}
	if errMsg.Len() > 0 {
		return "", fmt.Errorf(strings.TrimSpace(errMsg.String()))
	}
	return "", err
}
func GitHasOrigin(directory string) (bool, error) {
	cmd := exec.Command("git", "-C", directory, "remote", "show")
	remotes, err := cmd.Output()
	if err != nil {
		return false, err
	}
	if utils.SliceContains(strings.Split(string(remotes), "\n"), "origin") {
		return true, nil
	}
	return false, nil
}
func GitSetOrigin(directory string, origin string) error {
	has, err := GitHasOrigin(directory)
	if err != nil {
		return err
	}
	if has {
		return gitExec("-C", directory, "remote", "set-url", "origin", origin)
	}
	return gitExec("-C", directory, "remote", "add", "origin", origin)
}

func GitRemoveOrigin(directory string) error {
	return gitExec("-C", directory, "remote", "remove", "origin")
}
