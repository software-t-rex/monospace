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
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

type state term.State

func openTTY() (*os.File, error) {
	return os.OpenFile("/dev/tty", os.O_RDWR, 0)
}

func makeRaw(fd int) (*TermState, error) {
	st, err := term.MakeRaw(fd)
	return (*TermState)(st), err
}
func restore(fd int, state *TermState) error {
	return term.Restore(fd, (*term.State)(state))
}

func terminalHasDarkBackground(terminal interface {
	TermWithExclusiveReader
	TermWithRawMode
}) bool {
	// term := os.Getenv("TERM")
	// if !canEnhance || regexp.MustCompile(`^(screen|tmux|dumb|rxvt)`).MatchString(term) {
	// 	return true
	// }
	res, err := terminalQuery(terminal, terminalQueryOpts{
		querySequence:  "\x1b]11;?\a\x1b\\",
		expectedPrefix: "\x1b]11;rgb:",
		maxLen:         25,
		endVerifier: func(b byte, res []byte) (bool, error) {
			if len(res) > 1 && b == '\a' || b == '\x1b' || b == '\\' {
				return true, nil
			} else if b == '\n' || b == '\r' {
				return true, fmt.Errorf("bad response")
			}
			return false, nil
		},
	})

	if err != nil {
		if errors.Is(err, ErrTimeout) { // fallback to COLORFGBG
			return darkBackgroundFromEnv()
		}
		return true // fallback to dark
	}
	var r, g, b int
	n, err := fmt.Sscanf(strings.TrimRight(res, "\a\x1b\\"), "%04x/%04x/%04x", &r, &g, &b)
	if err != nil || n != 3 {
		return darkBackgroundFromEnv()
	}
	// Calculate luminance.
	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535
	return luminance < 0.5
}

func darkBackgroundFromEnv() bool {
	colorFgBg := os.Getenv("COLORFGBG")
	parts := strings.Split(colorFgBg, ";")
	if len(parts) >= 2 {
		if parts[len(parts)-1] == "0" {
			return true
		} else if parts[len(parts)-1] == "15" {
			return false
		}
	}
	return true
}
