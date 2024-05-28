//go:build windows
// +build windows

/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"os"

	"golang.org/x/sys/windows"
)

type state struct {
	mode uint32
}

func openTTY() (*os.File, error) {
	return os.OpenFile("CONIN$", os.O_RDWR, 0644)
}

func terminalHasDarkBackground(_ interface{}) bool {
	// consider all windows terminal as having dark background
	return true
}

func makeRaw(fd int) (*TermState, error) {
	var st uint32
	handle := windows.Handle(fd)
	if err := windows.GetConsoleMode(handle, &st); err != nil {
		return nil, err
	}
	raw := st &^ (windows.ENABLE_ECHO_INPUT | windows.ENABLE_PROCESSED_INPUT | windows.ENABLE_LINE_INPUT | windows.ENABLE_PROCESSED_OUTPUT)
	raw = raw | windows.ENABLE_VIRTUAL_TERMINAL_INPUT
	if err := windows.SetConsoleMode(handle, raw); err != nil {
		return nil, err
	}

	return &TermState{mode: st}, nil
}

func restore(fd int, state *TermState) error {
	return windows.SetConsoleMode(windows.Handle(fd), state.mode)
}
