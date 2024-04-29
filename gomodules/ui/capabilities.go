/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"os"
	"strings"
)

// this reflects user preferences and default to true
var enabledEnhanced = true

// this reflects the ability to enhance the rendering
// try to detect if the terminal is not able to render colors or if env var NO_COLOR|ACCESSIBLE are set
var canEnhance = true

// env vars that can be used to disable color rendering
var env_nocolor = os.Getenv("NO_COLOR")
var env_accessible = os.Getenv("ACCESSIBLE")
var env_term = os.Getenv("TERM")
var env_CI = os.Getenv("CI")

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
	canEnhance = true
	if (env_nocolor != "" && env_nocolor != "0" && env_nocolor != "false") ||
		(env_CI != "" && env_CI != "0" && env_CI != "false") ||
		(env_accessible != "" && env_accessible != "0" && env_accessible != "false") ||
		strings.HasPrefix(env_term, "dumb") || !t.IsTerminal() {
		canEnhance = false
	}
}
