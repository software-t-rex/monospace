/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

// SequenceKeysMap is a map of sequence keys to their corresponding key names.
// This map is used to convert escape sequences to key names.
//
// Here are some volountary ignored keys and the reason why:
// - ctrl+h: is detected as backspace
// - ctrl+i: is detected as tab
// - ctrl+j: is detected as enter
// - ctrl+m: is detected as enter
// - ctrl+v: is not detected and is an os paste on windows
// - ctrl+z: is not detected and trigger EOF on windows and may suspend on *nix
// - ctrl+h: is detected as backspace
// - f10 is unreachable on my computer and not tested
// - f11 is full screen on some terminals
// - find, select keys which are not present on my keyboard and not tested
//
// For other keys to be added you can propose a PR with the key.
// there's some tools to help you to get the key sequence in the tools directory
//
// some reference:
//
// - https://en.wikipedia.org/wiki/ANSI_escape_code
// - https://invisible-island.net/xterm/ctlseqs/ctlseqs.html
// - https://learn.microsoft.com/fr-fr/windows/console/console-virtual-terminal-sequences
var SequenceKeysMap = map[string]string{
	"\x1b": "esc",
	"\n":   "enter",
	"\r":   "enter",
	"\t":   "tab",
	"\x7f": "backspace",
	"\b":   "backspace",

	"\x1b[A": "up",    // normal mode
	"\x1b[B": "down",  // normal mode
	"\x1b[C": "right", // normal mode
	"\x1b[D": "left",  // normal mode
	"\x1b[H": "home",  // normal mode
	"\x1b[F": "end",   // normal mode

	"\x1bOA": "up",    // application mode
	"\x1bOB": "down",  // application mode
	"\x1bOC": "right", // application mode
	"\x1bOD": "left",  // application mode
	"\x1bOH": "home",  // application mode
	"\x1bOF": "end",   // application mode

	"\x1b[2~": "insert",
	"\x1b[3~": "delete",
	"\x1b[5~": "pageup",
	"\x1b[6~": "pagedown",
	"\x1b[7~": "home",
	"\x1b[8~": "end",

	"\x1b[3;5~": "ctrl+delete",
	"\x1b[3;3~": "alt+delete",

	"\x01": "ctrl+a",
	"\x02": "ctrl+b",
	"\x03": "ctrl+c",
	"\x04": "ctrl+d",
	"\x05": "ctrl+e",
	"\x06": "ctrl+f",
	"\x07": "ctrl+g", // terminal bell
	"\x0b": "ctrl+k", // vertical tab = \v
	"\x0c": "ctrl+l", // clear screen or form feed = \f
	"\x0e": "ctrl+n",
	"\x0f": "ctrl+o",
	"\x10": "ctrl+p",
	"\x11": "ctrl+q",
	"\x12": "ctrl+r",
	"\x13": "ctrl+s",
	"\x14": "ctrl+t",
	"\x15": "ctrl+u", // kill line
	"\x17": "ctrl+w", // delete previous word
	"\x18": "ctrl+x",
	"\x19": "ctrl+y",

	"\x1b[11~": "F1", //(wanted "\x1b[11~", got "\x1bOP") (windows wanted "\x1b[11~", got "\x1bOP")
	"\x1bOP":   "F1",
	"\x1b[12~": "F2", //(wanted "\x1b[12~", got "\x1bOQ")
	"\x1bOQ":   "F2", // windows
	"\x1b[13~": "F3", // (wanted "\x1b[13~", got "\x1bOR")
	"\x1bOR":   "F3",
	"\x1b[14~": "F4",
	"\x1bOS":   "F4", //windows,  (linux wanted "\x1b[14~", got "\x1bOS")
	"\x1b[15~": "F5",
	"\x1b[17~": "F6",
	"\x1b[18~": "F7",
	"\x1b[19~": "F8",
	"\x1b[20~": "F9",
	"\x1b[24~": "F12",

	"\x1b[1;2A": "shift+up",
	"\x1b[1;2B": "shift+down",
	"\x1b[1;2C": "shift+right",
	"\x1b[1;2D": "shift+left",
	"\x1b[Z":    "shift+tab",

	"\x1b[1;3A": "alt+up",
	"\x1b[1;3B": "alt+down",
	"\x1b[1;3C": "alt+right",
	"\x1b[1;3D": "alt+left",

	"\x1b[1;5A": "ctrl+up",
	"\x1b[1;5B": "ctrl+down",
	"\x1b[1;5C": "ctrl+right",
	"\x1b[1;5D": "ctrl+left",

	// following keys are not standard and may be captured by terminal emulator
	"\x1b\x7f": "alt+backspace", // delete previous word
	"\x1b\x62": "alt+b",         // move to previous word
	"\x1b\x66": "alt+f",         // move to next word
	"\x1b\x64": "alt+d",         // delete word

	// Add more compound sequences as needed
}
