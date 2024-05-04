/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

var ErrNOTERM = fmt.Errorf("not a terminal")

type (
	TermIsTerminal interface {
		IsTerminal() bool
	}
	TermWithRawMode interface {
		MakeRaw() (*term.State, error)
		Restore(*term.State) error
		HandleState(bool) (func(), error)
	}
	TermWithBackground interface {
		HasDarkBackground() (bool, error)
	}
	TermWithReader interface {
		NewReader() *bufio.Reader
		NewScanner() *bufio.Scanner
	}
	TTYFileDescriptor interface {
		Tty() *os.File
	}
	TermInterface interface {
		TermIsTerminal
		TermWithRawMode
		TermWithBackground
		TermWithReader
		TTYFileDescriptor
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
			terminal.bgChecked = true
			return terminal, err
		}
		terminal.tty = file
	}
	terminal.fd = int(terminal.tty.Fd())
	terminal.isTerm = term.IsTerminal(terminal.fd)
	if !terminal.isTerm {
		return terminal, fmt.Errorf("file descriptor %d is %w", terminal.fd, ErrNOTERM)
	}
	return terminal, nil
}

func (t *Terminal) IsTerminal() bool {
	return t.isTerm
}
func (t *Terminal) MakeRaw() (*term.State, error) {
	if !t.isTerm {
		return nil, fmt.Errorf("MakeRaw: %w", ErrNOTERM)
	}
	return makeRaw(t.fd)
}
func (t *Terminal) Restore(state *term.State) error {
	if !t.isTerm {
		return fmt.Errorf("Restore: %w", ErrNOTERM)
	}
	return term.Restore(t.fd, state)
}

// put terminal in raw mode and return a restore function and an error if any
//
// Usage:
//
//	restore, err := handleTerminalState(terminal)
//	defer restore()
func (t *Terminal) HandleState(hideCursor bool) (func(), error) {
	restore := func() {}
	state, err := t.MakeRaw()
	if errors.Is(err, ErrNOTERM) {
		return restore, err
	}
	tty := t.Tty()
	if hideCursor && err == nil {
		fmt.Fprint(tty, "\033[?25l")
	}
	restore = func() {
		t.Restore(state)
		if hideCursor && err == nil {
			fmt.Fprint(tty, "\033[?25h")
		}
	}
	return restore, err
}

func (t *Terminal) Tty() *os.File {
	return t.tty
}

// you should have checked that the terminal is a terminal before calling this function.
func (t *Terminal) NewReader() *bufio.Reader {
	return bufio.NewReader(t.tty)
}
func (t *Terminal) NewScanner() *bufio.Scanner {
	return bufio.NewScanner(t.tty)
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
		return true, fmt.Errorf("HasDarkBackground: %w", ErrNOTERM)
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
	if usedTerm == nil {
		usedTerm, _ = NewTerminal(nil)
	}
	return usedTerm
}
