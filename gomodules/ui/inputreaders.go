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
	"time"
)

var KeyNames = map[string]string{
	"\x1b": "esc",
	"\n":   "enter",
	"\r":   "enter",
	"\t":   "tab",
	"\x7f": "backspace",

	"\x03": "ctrl+c",
	"\x04": "ctrl+d",
	"\x18": "ctrl+x",
	"\x01": "ctrl+a",
	"\x16": "ctrl+v",
	"\x1a": "ctrl+z", // may be pause on some terminals
	"\x19": "ctrl+y",

	"\x1b[A":  "up",
	"\x1bOA":  "up",
	"\x1b[B":  "down",
	"\x1bOB":  "down",
	"\x1b[C":  "right",
	"\x1bOC":  "right",
	"\x1b[D":  "left",
	"\x1bOD":  "left",
	"\x1b[1~": "home",     // Some terminals like xterm
	"\x1b[H":  "home",     // Some terminals like iTerm2, Linux console
	"\x1bOH":  "home",     // Some terminals in application keypad mode
	"\x1b[7~": "home",     // rxvt
	"\x1b[4~": "end",      // Some terminals like xterm
	"\x1b[F":  "end",      // Some terminals like iTerm2, Linux console
	"\x1bOF":  "end",      // Some terminals in application keypad mode
	"\x1b[8~": "end",      // rxvt
	"\x1b[5~": "pageup",   // Linux console
	"\x1b[6~": "pagedown", // Linux console

	"\x1b[2~":  "insert", // Linux console
	"\x1b[3~":  "delete", // Linux console
	"\x1b[11~": "f1",     // Linux console
	"\x1b[12~": "f2",     // Linux console
	"\x1b[13~": "f3",     // Linux console
	"\x1b[14~": "f4",     // Linux console
	"\x1b[15~": "f5",     // Linux console
	"\x1b[17~": "f6",     // Linux console
	"\x1b[18~": "f7",     // Linux console
	"\x1b[19~": "f8",     // Linux console
	"\x1b[20~": "f9",     // Linux console
	"\x1b[21~": "f10",    // Linux console
	"\x1b[23~": "f11",    // Linux console
	"\x1b[24~": "f12",    // Linux console

	// Add more compound sequences as needed
}

type KeyMsg struct {
	Value string
}

func (k KeyMsg) String() string {
	return k.Value
}

func ReadKeyPressEvent(terminal interface {
	TermIsTerminal
	TermWithRawMode
	TermWithReader
}) (Msg, error) {
	if !terminal.IsTerminal() {
		return nil, nil
	}
	state, err := terminal.MakeRaw()
	fmt.Print("\033[?25l")
	defer func() {
		fmt.Print("\033[?25h")
		terminal.Restore(state)
	}()
	if err != nil {
		return nil, err
	}

	reader := terminal.NewReader()
	key, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if key == '\x1b' { // start of an escape sequence
		seq := []byte{key}
		for {
			if reader.Buffered() == 0 {
				time.Sleep(50 * time.Millisecond) // add a delay
				if reader.Buffered() == 0 {       // check again after the delay
					if key == '\x1b' {
						return KeyMsg{Value: "esc"}, nil // return "esc" if the escape key was pressed alone
					} else {
						return KeyMsg{Value: string(seq)}, nil // return "fn" if the function key was pressed alone
					}
				}
			}
			key, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}
			seq = append(seq, key)
			if name, exists := KeyNames[string(seq)]; exists {
				return KeyMsg{Value: name}, nil
			}
		}
	} else {
		if name, exists := KeyNames[string(key)]; exists {
			return KeyMsg{Value: name}, nil
		}
	}
	return KeyMsg{Value: string(key)}, nil
}

/** read a line from stdin */
type LineMsg struct {
	Value string
}

func (k LineMsg) String() string {
	return k.Value
}

func Readline(prompt string) (LineMsg, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	scanner.Scan()
	return LineMsg{Value: scanner.Text()}, scanner.Err()
}

type IntMsg struct {
	Value int
}

func ReadInt(prompt string) (IntMsg, error) {
	input, err := Readline(prompt)
	if err != nil {
		return IntMsg{}, err
	}
	v, err := strconv.Atoi(input.Value)
	if err != nil {
		return IntMsg{}, err
	}
	return IntMsg{Value: v}, nil
}

type IntsMsg struct {
	Value []int
}

/** Parse a string of space separated integers into a list of integers */
func ParseInts(str string) (IntsMsg, error) {
	inputs := strings.Fields(str)
	ints := make([]int, len(inputs))
	for k, input := range inputs {
		i, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			return IntsMsg{}, err
		}
		ints[k] = i
	}
	return IntsMsg{Value: ints}, nil
}

func ReadInts(prompt string) (IntsMsg, error) {
	line, err := Readline(prompt)
	if err != nil {
		return IntsMsg{}, err
	}
	return ParseInts(line.Value)
}
