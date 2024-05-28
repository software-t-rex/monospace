/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"fmt"
	"strings"
)

const (
	csiStart string = "\033["
	sgrEnd   string = "m"
	sgrReset string = "\033[0m"
)

type SGRParam string
type Styler = func(s ...string) string

const (
	Bold      SGRParam = "1"
	Faint     SGRParam = "2"
	Italic    SGRParam = "3"
	Underline SGRParam = "4"
	Blink     SGRParam = "5"
	Reversed  SGRParam = "7"
	Strike    SGRParam = "9"

	ResetBold       SGRParam = "22" // 21m is not specified this is not a mistake
	ResetFaint      SGRParam = "22" // so 22 reset both Bold and Faint
	ResetItalic     SGRParam = "23"
	ResetUnderline  SGRParam = "24"
	ResetBlink      SGRParam = "25"
	ResetReversed   SGRParam = "27"
	ResetStrike     SGRParam = "29"
	ResetForeground SGRParam = "39"
	ResetBackground SGRParam = "49"
	ResetAll        SGRParam = "0"
)

var resetParams = map[SGRParam]SGRParam{
	Bold:      ResetBold,
	Faint:     ResetFaint,
	Italic:    ResetItalic,
	Underline: ResetUnderline,
	Blink:     ResetBlink,
	Reversed:  ResetReversed,
	Strike:    ResetStrike,
	"30":      ResetForeground,
	"31":      ResetForeground,
	"32":      ResetForeground,
	"33":      ResetForeground,
	"34":      ResetForeground,
	"35":      ResetForeground,
	"36":      ResetForeground,
	"37":      ResetForeground,
	"90":      ResetForeground,
	"91":      ResetForeground,
	"92":      ResetForeground,
	"93":      ResetForeground,
	"94":      ResetForeground,
	"95":      ResetForeground,
	"96":      ResetForeground,
	"97":      ResetForeground,
	"40":      ResetBackground,
	"41":      ResetBackground,
	"42":      ResetBackground,
	"43":      ResetBackground,
	"44":      ResetBackground,
	"45":      ResetBackground,
	"46":      ResetBackground,
	"47":      ResetBackground,
	"100":     ResetBackground,
	"101":     ResetBackground,
	"102":     ResetBackground,
	"103":     ResetBackground,
	"104":     ResetBackground,
	"105":     ResetBackground,
	"106":     ResetBackground,
	"107":     ResetBackground,
}

func isResetParam(p SGRParam) bool {
	switch p {
	// resetFaint == resetBold
	case ResetBold, ResetItalic, ResetUnderline, ResetBlink, ResetReversed, ResetStrike, ResetForeground, ResetBackground:
		return true
	}
	return false
}

// Combine multiple ANSI SGR params.
func SGRCombine(styles ...SGRParam) SGRParam {
	var sb strings.Builder
	for i, style := range styles {
		if i > 0 {
			sb.WriteString(";")
		}
		sb.WriteString(string(style))
	}
	return SGRParam(sb.String())
}

// Return an ANSI SGR escape sequence from given SGR parameters.
// Sample usage: ui.SGREscapeSequence(colors.Red, colors.Bold)
func SGREscapeSequence(styles ...SGRParam) string {
	if len(styles) == 0 {
		return ""
	}
	return csiStart + string(SGRCombine(styles...)) + sgrEnd
}

// Return the reset ANSI SGR escape sequence corresponding for given SGR parameters.
// this is more computationally expensive than simply use a reset all sequence \033[0m
// but it's more accurate as it will only reset the styles that were set (waring Faint/Bold use the same reset).
// thus allowing to embed styles in styles without having to worry about resetting them.
//
// If no styles are given it will return a reset all sequence.
// If it can't determine the reset style for one of its arguments, it will return a reset all sequence.
func SGRResetSequence(styles ...SGRParam) string {
	if len(styles) == 0 {
		return "\033[0m"
	}
	addedSeq := make(map[SGRParam]struct{})
	resetSeq := []SGRParam{}
	for _, style := range styles {
		reset, ok := resetParams[style]
		if !ok {
			if isResetParam(style) {
				continue
			} else if strings.HasPrefix(string(style), "38;") { // this is a foreground color
				reset = ResetForeground
			} else if strings.HasPrefix(string(style), "48;") { // this is a background color
				reset = ResetBackground
			} else { // we can't determine the reset style so we return a reset all
				return sgrReset
			}
		}
		if _, visited := addedSeq[reset]; !visited {
			addedSeq[reset] = struct{}{}
			resetSeq = append(resetSeq, reset)
		}
	}

	return csiStart + string(SGRCombine(resetSeq...)) + sgrEnd
}

// Directly apply styles to a string.
// If Enhanced is disabled, the string is returned as is.
// Sample usage: ui.ApplyStyle("This will be red and bold.", colors.Red, colors.Bold)
func ApplyStyle(s string, styles ...SGRParam) string {
	if !EnhancedEnabled() {
		return s
	}
	sgrSeq := SGREscapeSequence(styles...)
	sgrResetSeq := SGRResetSequence(styles...)
	return sgrSeq + s + sgrResetSeq
}

// Returns a function that will apply given styles to the received String
// The returned function will return an unstyled string if ui.ToggleEnhanced(false)
// has been called (or set to false on init by the env var NO_COLOR, ACCESSIBLE or TERM=dumb)
// Sample usage: ui.NewStyler(colors.Red, colors.Bold)("This will be red and bold.")
func NewStyler(styles ...SGRParam) Styler {
	// cache the style string
	sgrSeq := SGREscapeSequence(styles...)
	sgrResetSeq := SGRResetSequence(styles...)
	return func(s ...string) string {
		str := strings.Join(s, " ")
		if !EnhancedEnabled() {
			return str
		}
		return sgrSeq + str + sgrResetSeq
	}
}

// utility function to print a string with given styles
func Println(s string, styles ...SGRParam) {
	fmt.Println(ApplyStyle(s, styles...))
}
