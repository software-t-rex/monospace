/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/
package ui

// do all initialisation here
func init() {
	terminal, _ := NewTerminal(nil)
	SetTerminal(terminal)
	detectCapability(usedTerm)
	SetTheme(ThemeDefault)
}
