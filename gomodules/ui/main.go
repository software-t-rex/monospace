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
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// this reflects user preferences and default to true
var enabledEnhanced = true

// this reflects the ability to enhance the rendering
// try to detect if the terminal is not able to render colors or if env var NO_COLOR|ACCESSIBLE are set
var canEnhance = true

// keep trace of the running program to be able to kill it
var killTeaProgram func()

type TeaModelWithFallback interface {
	tea.Model
	Fallback() TeaModelWithFallback
}

//#region utility functions

func CheckErr(err error) {
	if err != nil {
		if killTeaProgram != nil {
			killTeaProgram()
		}
		os.Exit(1)
	}
}

/** Parse a string of space separated integers into a list of integers */
func ParseInts(str string) ([]int, error) {
	inputs := strings.Fields(str)
	ints := make([]int, len(inputs))
	for k, input := range inputs {
		i, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			return nil, err
		}
		ints[k] = i
	}
	return ints, nil
}

//#enregion utility functions

// Allow to turn off colouring for all Style methods
// Be careful: if you do string(colors.Red) + "a red string" + string(Reset)
// it will still be rendered in red as you use colors codes directly.
func ToggleEnhenced(enable bool) {
	enabledEnhanced = enable
}

// Returns whether enhanced rendering is enabled
func EnhencedEnabled() bool {
	return enabledEnhanced && canEnhance
}

func init() {
	nocolor := os.Getenv("NO_COLOR")
	accessible := os.Getenv("ACCESSIBLE")
	if (nocolor != "" && nocolor != "0" && nocolor != "false") || (accessible != "" && accessible != "0" && accessible != "false") {
		canEnhance = false
	}
}

/** read a line from stdin */
func Readline(prompt string) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}
func ReadInt(prompt string) (int, error) {
	input, err := Readline(prompt)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(input)
}
func ReadInts(prompt string) ([]int, error) {
	inputs, err := Readline(prompt)
	if err != nil {
		return nil, err
	}
	return ParseInts(inputs)
}

/** helper function to run a tea program and return the result model */
func runTeaProgram[T TeaModelWithFallback](initModel T) T {
	if killTeaProgram != nil {
		panic("a program is already running")
	}
	if !EnhencedEnabled() { // use fallback mode
		initModel.Init()
		return initModel.Fallback().(T)
	}
	p := tea.NewProgram(initModel)
	killTeaProgram = p.Kill
	res, err := p.Run()
	killTeaProgram = nil
	if err != nil {
		CheckErr(err)
	}
	return res.(T)
}

/** helper function to call when user abort the program like with ctrl+c */
func AbortTeaProgram() tea.Msg {
	fmt.Printf("\n" + usedTheme.Error("User Aborted") + "\n")
	if killTeaProgram != nil {
		killTeaProgram()
	}
	return tea.Quit
}
