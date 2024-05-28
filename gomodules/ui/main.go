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
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/software-t-rex/monospace/gomodules/ui/pkg/ansi"
)

var ErrInputRead = fmt.Errorf("input read error")
var ErrRunComponent = fmt.Errorf("error while running component")

type (
	Msg          interface{}
	Cmd          func() Msg
	ComponentApi struct {
		Done        bool
		Cleanup     bool
		InputReader inputReader
	}
	Model interface {
		// set initial state of the component
		// can return a Cmd to run some initialisation commands
		Init() Cmd
		// return a rendered view of the model to display on the terminal
		Render() string
		// update the model with a message and return an optional Cmd to run
		// if this method returns a command it will then be executed again with the
		// msg returned by the command until it returns nil or the done flag is set to true
		// then the Render method will be called again to display the updated view
		// beware that the model should not be updated concurrently
		Update(Msg) Cmd
		// This method is called when we fallback to the non-enhanced mode
		// In such case there's no Render/Update loop but only a single call to Fallback
		// For examples of fallback usage see the confirm.go file
		Fallback() (Model, error)
		// return the embeded component api
		GetComponentApi() *ComponentApi
	}
)

func printError(err error) {
	fmt.Fprintln(os.Stderr, "\n"+GetTheme().Error(err.Error()))
}

func lineCounter(s string) int {
	return strings.Count(s, "\n")
}

func handleSpecialMsgs(m Model, msg Msg) bool {
	switch msg.(type) {
	case MsgQuit:
		m.GetComponentApi().Done = true
		return true
	case MsgKill:
		CmdKill()
		return true
	}
	return false
}
func handleMsgs(m Model, msg Msg) Cmd {
	if !handleSpecialMsgs(m, msg) {
		return m.Update(msg)
	}
	return nil
}
func executeCmds(m Model, cmd Cmd) {
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			return
		}
		cmd = handleMsgs(m, msg)
	}
}

func runComponent[M Model](m M) (M, error) {
	cmd := m.Init()
	executeCmds(m, cmd)
	linesPrinted := 0
	api := m.GetComponentApi()
	terminal := GetTerminal()
	for !api.Done {
		toPrint := lfToCrLf(m.Render())
		// Update the number of lines printed
		linesPrinted = lineCounter(toPrint)
		var msg Msg
		var err error
		switch api.InputReader {
		case LineReader, LineReaderWrapped, lineReaderFramed, PasswordReader:
			fmt.Print(toPrint)
			msg, err = ReadLineEnhanced(terminal, lineReaderVisualMode[api.InputReader], m)
			if err != nil {
				if errors.Is(err, ErrSIGINT) || errors.Is(err, io.EOF) {
					if api.Cleanup {
						EraseNLines(linesPrinted)
					}
					handleSpecialMsgs(m, msg)
					return m, err
				}
				return m, fmt.Errorf("runComponent %w: %w", ErrInputRead, err)
			}
		case KeyReader:
			fmt.Print(toPrint)
			msg, err = ReadKeyPressEvent(terminal)
		}
		if err != nil {
			return m, fmt.Errorf("runComponent %w: %w", ErrInputRead, err)
		}
		executeCmds(m, handleMsgs(m, msg))
		// Move the cursor up to the start of last render and erase from here
		EraseNLines(linesPrinted)
		// optional last render
		if api.Done && !api.Cleanup {
			// ensure we have a newline at the end of the last render
			view := m.Render()
			if !strings.HasSuffix(view, "\n") {
				view += "\r\n"
			}
			fmt.Print(view)
		}
	}
	return m, nil
}

func EraseNLines(n int) {
	if n > 0 {
		fmt.Printf(ansi.CtrlHorizAbs.Sprintf(0) + ansi.CtrlUp.Sprintf(n) + ansi.CtrlEraseD.Sprintf(0))
	} else {
		fmt.Printf(ansi.CtrlHorizAbs.Sprintf(0) + ansi.CtrlEraseD.Sprintf(0))
	}
}
func EraseText(text string) {
	EraseNLines(lineCounter(text))
}

/** helper function to run a component and return the result model */
func RunComponent[M Model](m M) (M, error) {
	if !EnhancedEnabled() { // use fallback mode
		cmd := m.Init()
		executeCmds(m, cmd)
		_m, err := m.Fallback()
		return _m.(M), err
	}
	res, err := runComponent(m)
	if err != nil {
		if errors.Is(err, ErrSIGINT) {
			CmdKill()
		} else if !errors.Is(err, io.EOF) { // EOF is not an error we want to handle
			_m, err := m.Fallback()
			return _m.(M), fmt.Errorf("%w: %w", ErrRunComponent, err)
		}
	}
	return res, err
}

type (
	// MsgQuit is a message to mark the component as done
	MsgQuit struct{}
	// MsgKill is a message to kill the program
	MsgKill struct{}
)

func CmdQuit() Msg { return MsgQuit{} }
func CmdKill() Msg {
	os.Exit(1)
	return nil
}
func CmdUserCancel() Msg {
	printError(fmt.Errorf("user cancelled"))
	return MsgQuit{}
}
func CmdUserAbort() Msg {
	printError(fmt.Errorf("user aborted"))
	os.Exit(1)
	return MsgKill{}
}

func lfToCrLf(prompt string) string {
	return regexp.MustCompile(`([^\r])\n`).ReplaceAllString(prompt, "$1\r\n")
}
