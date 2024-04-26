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
)

// this reflects user preferences and default to true
var enabledEnhanced = true

// this reflects the ability to enhance the rendering
// try to detect if the terminal is not able to render colors or if env var NO_COLOR|ACCESSIBLE are set
var canEnhance = true

type (
	Msg          interface{}
	Cmd          func() Msg
	ComponentApi struct {
		done    bool
		cleanup bool
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
		Fallback() Model
		// return the embeded component api
		GetComponentApi() *ComponentApi
	}
)

// Allow to turn off colouring for all Style methods
// Be careful: if you do string(colors.Red) + "a red string" + string(Reset)
// it will still be rendered in red as you use colors codes directly.
func ToggleEnhanced(enable bool) {
	enabledEnhanced = enable
}

// Returns whether enhanced rendering is enabled
func EnhancedEnabled() bool {
	return enabledEnhanced && canEnhance
}

// this mehod is used to detect if the terminal is able to render colors
// and set the canEnhance flag accordingly
func detectCapability(t TermIsTerminal) {
	nocolor := os.Getenv("NO_COLOR")
	accessible := os.Getenv("ACCESSIBLE")
	TERM := os.Getenv("TERM")
	isTerm := t.IsTerminal()
	canEnhance = true
	if (nocolor != "" && nocolor != "0" && nocolor != "false") ||
		(accessible != "" && accessible != "0" && accessible != "false") ||
		TERM == "dumb" || !isTerm {
		canEnhance = false
	}
}

func printError(err error) {
	fmt.Println("\n" + GetTheme().Error(err.Error()))
}

func lineCounter(s string) int {
	return strings.Count(s, "\n")
}

func executeCmds(m Model, cmd Cmd) {
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			return
		}
		switch msg := msg.(type) {
		case MsgQuit:
			m.GetComponentApi().done = true
		case MsgKill:
			CmdKill()
			return
		default:
			cmd = m.Update(msg)
		}
	}
}

func runComponent[M Model](m M) (M, error) {
	cmd := m.Init()
	executeCmds(m, cmd)
	linesPrinted := 0
	api := m.GetComponentApi()
	terminal := GetTerminal()
	for !api.done {
		toPrint := m.Render() + "\r\n"
		fmt.Print(toPrint)
		// Update the number of lines printed
		linesPrinted = lineCounter(toPrint)
		keyMsg, err := ReadKeyPressEvent(terminal)
		if err != nil {
			printError(fmt.Errorf("error reading keypress event: %v", err))
			return m, err
		}
		executeCmds(m, m.Update(keyMsg))
		// Move the cursor up to the start of the list and clear the list
		fmt.Printf("\033[0G\033[%dA\033[J", linesPrinted)
		// last render
		if api.done && !api.cleanup {
			fmt.Print(m.Render() + "\r\n")
		}
	}
	return m, nil
}

/** helper function to run a tea program and return the result model */
func RunComponent[T Model](initModel T) T {
	if !EnhancedEnabled() { // use fallback mode
		initModel.Init()
		return initModel.Fallback().(T)
	}
	res, err := runComponent(initModel)
	if err != nil {
		printError(fmt.Errorf("error while running component: %v", err))
		return initModel.Fallback().(T)
	}
	return res
}

type (
	MsgQuit struct{}
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
