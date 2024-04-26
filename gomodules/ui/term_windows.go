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
	"golang.org/x/term"
)

func openTTY() (*os.File, error) {
	// return os.Stdin, nil
	return os.OpenFile("CONIN$", os.O_RDWR, 0)
}

func terminalHasDarkBackground(_ interface{}) bool {
	// consider all windows terminal as having dark background
	return true
}

func makeRaw(fd int) (*term.State, error) {
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	handle := windows.Handle(fd)
	var mode uint32
	err = windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return nil, err
	}
	// this is needed to be able to read special keys like F1, F2, arrows keys, etc...
	err = windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_INPUT)
	if err != nil {
		return nil, err
	}
	return state, err
}
