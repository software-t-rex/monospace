//go:build windows

/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package utils

import (
	"os/exec"
)

func Open(file string) *exec.Cmd {
	return exec.Command("rundll32", "url.dll", "FileProtocolHandler", file).Start()
}
