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

type (
	MsgKey struct {
		Value      string
	}
	MsgLine struct {
		Value string
	}
	MsgInt struct {
		Value int
	}
	MsgInts struct {
		Value []int
	}
)

func (k MsgKey) String() string  { return k.Value }
func (k MsgLine) String() string { return k.Value }
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
					if len(seq) == 1 { // if the escape key was pressed alone
						return MsgKey{Value: "esc"}, nil
					} else { // return unknown sequence
						seq = append(seq, key)
						return MsgKey{Value: string(seq)}, nil
					}
				}
			}
			key, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}
			seq = append(seq, key)
			if name, exists := KeyNames[string(seq)]; exists {
				return MsgKey{Value: name}, nil
			}
		}
	} else {
			return MsgKey{Value: name}, nil
}

/** read a line from stdin */
func Readline(prompt string) (MsgLine, error) {
	fmt.Print(prompt)
	scanner.Scan()
	return MsgLine{Value: scanner.Text()}, scanner.Err()
}

func lfToCrLf(prompt string) string {
	return regexp.MustCompile(`([^\r])\n`).ReplaceAllString(prompt, "$1\r\n")
}

func ReadInt(prompt string) (MsgInt, error) {
	input, err := Readline(prompt)
	if err != nil {
		return MsgInt{}, err
	}
	v, err := strconv.Atoi(input.Value)
	if err != nil {
		return MsgInt{}, err
	}
	return MsgInt{Value: v}, nil
}

/** Parse a string of space separated integers into a list of integers */
func ParseInts(str string) (MsgInts, error) {
	inputs := strings.Fields(str)
	ints := make([]int, len(inputs))
	for k, input := range inputs {
		i, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			return MsgInts{}, err
		}
		ints[k] = i
	}
	return MsgInts{Value: ints}, nil
}

func ReadInts(prompt string) (MsgInts, error) {
	line, err := Readline(prompt)
	if err != nil {
		return MsgInts{}, err
	}
	return ParseInts(line.Value)
}
