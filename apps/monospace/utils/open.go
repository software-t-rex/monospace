//go:build !darwin && !windows

/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package utils

import "os/exec"

func Open(file string) error {
	return exec.Command("xdg-open", file).Run()
}
