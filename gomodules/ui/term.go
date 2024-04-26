/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/term"
)

type (
	TermIsTerminal interface {
		IsTerminal() bool
	}
	TermWithRawMode interface {
		MakeRaw() (*term.State, error)
		Restore(*term.State) error
	}
	TermWithBackground interface {
		HasDarkBackground() (bool, error)
	}
	TermWithReader interface {
		NewReader() *bufio.Reader
	}
	TermInterface interface {
		TermIsTerminal
		TermWithRawMode
		TermWithBackground
		TermWithReader
	}
)

type Terminal struct {
	tty       *os.File
	fd        int
	isTerm    bool
	hasDarkBg bool
	bgChecked bool
}

func NewTerminal(tty *os.File) (*Terminal, error) {
	terminal := &Terminal{
		hasDarkBg: true, // default to most common case
		fd:        -1,
	}
	if tty != nil {
		terminal.tty = tty
	} else {
		file, err := openTTY()
		if err != nil {
			return terminal, err
		}
		terminal.tty = file
	}
	terminal.fd = int(terminal.tty.Fd())
	terminal.isTerm = term.IsTerminal(terminal.fd)
	if !terminal.isTerm {
		return terminal, fmt.Errorf("file descriptor %d is not a terminal", terminal.fd)
	}
	return terminal, nil
}

func (t *Terminal) IsTerminal() bool {
	return t.isTerm
}
func (t *Terminal) MakeRaw() (*term.State, error) {
	if !t.isTerm {
		return nil, fmt.Errorf("trying to make raw a non terminal file descriptor %d", t.fd)
	}
	return makeRaw(t.fd)
}
func (t *Terminal) Restore(state *term.State) error {
	if !t.isTerm {
		return fmt.Errorf("trying to restore a non terminal file descriptor %d", t.fd)
	}
	return term.Restore(t.fd, state)
}

// you should have checked that the terminal is a terminal before calling this function.
func (t *Terminal) NewReader() *bufio.Reader {
	return bufio.NewReader(t.tty)
}

// WARNING: you should have checked that the terminal is a terminal before calling this function.
// you can do it like this term.IsTerminal(int(os.stdin.Fd()))
// Try to use OSC query to get the background color.
// it is the most reliable method to get the background color when available,
// but it is not supported in all terminals.
// If it fails, it will default to the COLORFGBG environment variable.
// but it is not reliable for example Konsole keep reporting the wrong
// background color after a change.
func (t *Terminal) HasDarkBackground() (bool, error) {
	if !t.isTerm {
		return true, fmt.Errorf("trying to get the background color of a non terminal file descriptor %d", t.fd)
	}
	if !t.bgChecked {
		t.hasDarkBg = terminalHasDarkBackground(t)
		t.bgChecked = true

	}
	return t.hasDarkBg, nil
}

var usedTerm TermInterface

// this should be mocked in tests
func SetTerminal(t TermInterface) {
	usedTerm = t
}
func GetTerminal() TermInterface {
	return usedTerm
}
