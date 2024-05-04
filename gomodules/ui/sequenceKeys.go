/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

// SequenceKeysMap is a map of sequence keys to their corresponding key names.
// This map is used to convert escape sequences to key names.
// here are some volountary ignored keys and the reason why:
// - ctrl+h: is detected as backspace
// - ctrl+i: is detected as tab
// - ctrl+j: is detected as enter
// - ctrl+m: is detected as enter
// - ctrl+v: is not detected and is an os paste on windows
// - ctrl+z: is not detected and trigger EOF on windows and may suspend on *nix
// - ctrl+h: is detected as backspace
// - f10 is unreachable on my computer and not tested
// - f11 is full screen on some terminals
// For other keys to be added you can propose a PR with the key.
// there's some tools to help you to get the key sequence in the tools directory

var KeyNames = map[string]string{
	"\x1b": "esc",
	"\n":   "enter", // (wanted "\n", got "\r") (windows wanted "\r", got "\n")
	"\r":   "enter",
	"\t":   "tab",
	"\x7f": "backspace",
	"\b":   "backspace", // (wanted "\b", got "\x7f")

	"\x1b[A":  "up",
	"\x1bOA":  "up", // (windows wanted "\x1bOA", got "\x1b[A")
	"\x1b[B":  "down",
	"\x1bOB":  "down", // (windows wanted "\x1bOB", got "\x1b[B")
	"\x1b[C":  "right",
	"\x1bOC":  "right",    // (wanted "\x1bOC", got "\x1b[C")
	"\x1b[D":  "left",     // (windows wanted "\x1bOD", got "\x1b[D")
	"\x1bOD":  "left",     // (wanted "\x1bOD", got "\x1b[D")
	"\x1b[1~": "home",     // Some terminals like xterm (wanted "\x1b[1~", got "\x1b[H") ( windows wanted "\x1b[1~", got "\x1b[H")
	"\x1b[H":  "home",     // Some terminals like iTerm2, Linux console (mycomputer even windows terminal)
	"\x1bOH":  "home",     // Some terminals in application keypad mode(wanted "\x1bOH", got "\x1b[H")
	"\x1b[7~": "home",     // rxvt (wanted "\x1b[7~", got "\x1b[H") (windows wanted "\x1b[7~", got "\x1b[H")
	"\x1b[4~": "end",      // Some terminals like xterm (wanted "\x1b[4~", got "\x1b[F") (windows wanted "\x1b[4~", got "\x1b[F")
	"\x1b[F":  "end",      // Some terminals like iTerm2, Linux console
	"\x1bOF":  "end",      // Some terminals in application keypad mode (wanted "\x1bOF", got "\x1b[F")  (windows wanted "\x1bOF", got "\x1b[F")
	"\x1b[8~": "end",      // rxvt (wanted "\x1b[8~", got "\x1b[F") (windows wanted "\x1b[8~", got "\x1b[F")
	"\x1b[5~": "pageup",   // Linux console
	"\x1b[6~": "pagedown", // Linux console

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
	"\x18": "ctrl+x",
	"\x17": "ctrl+w", // delete previous word
	"\x19": "ctrl+y",

	"\x1b[2~":  "insert",
	"\x1b[3~":  "delete",
	"\x1b[11~": "f1", //(wanted "\x1b[11~", got "\x1bOP") (windows wanted "\x1b[11~", got "\x1bOP")
	"\x1bOP":   "f1",
	"\x1b[12~": "f2", //(wanted "\x1b[12~", got "\x1bOQ")
	"\x1bOQ":   "f2", // windows
	"\x1b[13~": "f3", // (wanted "\x1b[13~", got "\x1bOR")
	"\x1bOR":   "f3",
	"\x1b[14~": "f4",
	"\x1bOS":   "f4", //windows,  (linux wanted "\x1b[14~", got "\x1bOS")
	"\x1b[15~": "f5",
	"\x1b[17~": "f6",
	"\x1b[18~": "f7",
	"\x1b[19~": "f8",
	"\x1b[20~": "f9",
	"\x1b[24~": "f12",

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
