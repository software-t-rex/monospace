//go:build !windows
// +build !windows

/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

func openTTY() (*os.File, error) {
	return os.OpenFile("/dev/tty", os.O_RDWR, 0)
}

func makeRaw(fd int) (*term.State, error) {
	return term.MakeRaw(fd)
}

func terminalHasDarkBackground(terminal interface {
	TermWithReader
	TermWithRawMode
}) bool {
	// at this point we must have checked that the file is a terminal
	state, _ := terminal.MakeRaw()
	defer terminal.Restore(state)
	fmt.Print("\033]11;?\a")
	reader := terminal.NewReader()
	responseChan := make(chan string, 1)
	go func() {
		response, err := reader.ReadString('\a')
		if err == nil {
			responseChan <- response
		}
	}()

	select {
	case response := <-responseChan:
		var r, g, b int
		n, err := fmt.Sscanf(response, "\033]11;rgb:%04x/%04x/%04x\a", &r, &g, &b)
		if err != nil || n != 3 {
			return true
		}
		// Calculate luminance.
		luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535
		return luminance < 0.5
	case <-time.After(time.Millisecond * 20): // 2 seconds timeout
		// If OSC query didn't work, check the COLORFGBG environment variable.
		colorFgBg := os.Getenv("COLORFGBG")
		parts := strings.Split(colorFgBg, ";")
		if len(parts) >= 2 {
			if parts[len(parts)-1] == "0" {
				return true
			} else if parts[len(parts)-1] == "15" {
				return false
			}
		}
	}

	// If neither method worked, default to dark.
	return true
}
