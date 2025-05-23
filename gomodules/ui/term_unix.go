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
	"time"

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

// -1 means unknown, 0 means false, 1 means true
var hasDarkBackgroundCache int = -1
var hasDarkBackgroundLastCheck time.Time = time.Now()

func terminalHasDarkBackground(terminal interface {
	TermWithExclusiveReader
	TermWithRawMode
}) bool {
	if hasDarkBackgroundCache != -1 && time.Since(hasDarkBackgroundLastCheck) < 5*time.Second {
		return hasDarkBackgroundCache == 1
	}
	// term := os.Getenv("TERM")
	// if !canEnhance || regexp.MustCompile(`^(screen|tmux|dumb|rxvt)`).MatchString(term) {
	// 	return true
	// }
	res, err := terminalQuery(terminal, terminalQueryOpts{
		querySequence:  "\x1b]11;?\a\x1b\\",
		expectedPrefix: "\x1b]11;rgb:",
		maxLen:         25,
		endVerifier: func(b byte, res []byte) (bool, error) {
			hasDarkBackgroundLastCheck = time.Now()
			if len(res) > 1 && b == '\a' || b == '\x1b' || b == '\\' {
				hasDarkBackgroundCache = 1
				return true, nil
			} else if b == '\n' || b == '\r' {
				hasDarkBackgroundCache = 1
				return true, fmt.Errorf("bad response")
			}
			hasDarkBackgroundCache = 0
			return false, nil
		},
	})

	if err != nil {
		if errors.Is(err, ErrTimeout) { // fallback to COLORFGBG
			if darkBackgroundFromEnv() {
				hasDarkBackgroundCache = 1
			} else {
				hasDarkBackgroundCache = 0
			}
			return hasDarkBackgroundCache == 1
		}
		// fallback to dark
		hasDarkBackgroundCache = 1
		return true
	}
	var r, g, b int
	n, err := fmt.Sscanf(strings.TrimRight(res, "\a\x1b\\"), "%04x/%04x/%04x", &r, &g, &b)
	if err != nil || n != 3 {
		if darkBackgroundFromEnv() {
			hasDarkBackgroundCache = 1
		} else {
			hasDarkBackgroundCache = 0
		}
		return hasDarkBackgroundCache == 1
	}
	// Calculate luminance.
	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535
	if luminance < 0.5 {
		hasDarkBackgroundCache = 1
	} else {
		hasDarkBackgroundCache = 0
	}
	return hasDarkBackgroundCache == 1
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
