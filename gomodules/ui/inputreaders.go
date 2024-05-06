/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type (
	MsgKey struct {
		Value   string
		Unknown bool
		IsSeq   bool
		ByteSeq []byte
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
	readlineValueProvider interface {
		ReadlineValue() string
	}
	readlinekeyHandlerProvider interface {
		ReadlineKeyHandler(string) (Msg, error)
	}
	readlineCompletionProvider interface {
		ReadlineCompletion(string) ([]string, error)
	}
)

func (k MsgKey) String() string  { return k.Value }
func (k MsgLine) String() string { return k.Value }

var ErrSIGINT = fmt.Errorf("SIGINT")
var ErrEOF = fmt.Errorf("EOF")
var ErrComp = fmt.Errorf("completion error")

// you MUST already have a terminal with activated raw mode
func readKeyPressEvent(terminal interface {
	TermWithReader
}) (MsgKey, error) {
	reader := terminal.NewReader()
	key, err := reader.ReadByte()
	if err != nil {
		return MsgKey{}, err
	}
	if key == '\x1b' { // start of an escape sequence
		seq := []byte{key}
		for {
			if reader.Buffered() == 0 { // check if we have more bytes to read
				time.Sleep(30 * time.Millisecond) // add a delay just in case we have a sequence that is not yet complete
				if reader.Buffered() == 0 {       // check again after the delay
					if len(seq) == 1 { // if the escape key was pressed alone
						return MsgKey{Value: "esc", IsSeq: true, ByteSeq: []byte{key}}, nil
					} else { // return unknown sequence
						seq = append(seq, key)
						return MsgKey{Value: string(seq), IsSeq: true, Unknown: true, ByteSeq: seq}, nil
					}
				}
			}
			key, err := reader.ReadByte()
			if err != nil {
				return MsgKey{}, err
			}
			// we should not have any escape key inside the sequence
			// if so we consider it as an unknown sequence but return Value esc
			if key == '\x1b' {
				return MsgKey{Value: "esc", IsSeq: true, Unknown: true, ByteSeq: seq}, nil
			}
			seq = append(seq, key)
			if name, exists := SequenceKeysMap[string(seq)]; exists {
				return MsgKey{Value: name, IsSeq: true, ByteSeq: seq}, nil
			}
			// if len(seq) > 25 { // limit the size of the sequence to 25 characters
			// 	return MsgKey{Value: string(seq), IsKnownSeq: false, ByteSeq: seq}, nil
			// }
		}
	} else {
		if name, exists := SequenceKeysMap[string(key)]; exists {
			return MsgKey{Value: name, IsSeq: true, ByteSeq: []byte{key}}, nil
		}
	}
	// consider any other key as a single key and as a known sequence
	return MsgKey{Value: string(key), ByteSeq: []byte{key}}, nil
}

func ReadKeyPressEvent(terminal interface {
	TermIsTerminal
	TermWithRawMode
	TermWithReader
	TTYFileDescriptor
}) (Msg, error) {
	restore, err := terminal.HandleState(true)
	defer restore()
	if err != nil {
		return nil, err
	}
	return readKeyPressEvent(terminal)
}

// return the line typed by the user or an error if any
// common terminal shortcuts are supported:
// - ctrl+c to cancel return empty MsgKill and error "SIGINT"
// - ctrl+d delete or if empty return empty Msg and error "EOF"
// - ctrl+l to clear line
// - ctrl+u to delete from cursor to start of line
// - ctrl+k to delete from cursor to end of line
// - ctrl+w to delete previous word
// - alt+d to delete next word
// - alt+b to move to previous word
// - alt+f to move to next word
// - ctrl+a/home to move to start of line
// - ctrl+e/end to move to end of line
// - left to move left
// - right to move right
// - backspace to delete previous character
// - delete to delete
// - enter to return the line
func readLineEnhanced(terminal interface {
	TermIsTerminal
	TermWithRawMode
	TermWithReader
	TTYFileDescriptor
}, provider any, hiddenInput bool) (Msg, error) {

	restore, err := terminal.HandleState(hiddenInput)
	defer restore()
	if err != nil {
		return nil, err
	}
	tty := terminal.Tty()
	line := NewLineEditor(tty, !hiddenInput)
	var keyHandlerProvider readlinekeyHandlerProvider
	var haskeyHandlerProvider bool
	var hasCompletionProvider bool
	var completionProvider readlineCompletionProvider
	if provider != nil {
		valProvider, hasValProvider := any(provider).(readlineValueProvider)
		keyHandlerProvider, haskeyHandlerProvider = any(provider).(readlinekeyHandlerProvider)
		completionProvider, hasCompletionProvider = any(provider).(readlineCompletionProvider)
		if hasValProvider {
			line.value = valProvider.ReadlineValue()
			line.cursorPos = line.len()
		}
	}
	for {
		key, err := readKeyPressEvent(terminal)
		if err != nil {
			return MsgLine{Value: line.value}, err
		}
		if haskeyHandlerProvider {
			msg, err := keyHandlerProvider.ReadlineKeyHandler(key.Value)
			if msg != nil || err != nil {
				return msg, err
			}
		}
		if line.completing && key.Value != "tab" {
			line.completionEnd()
		}
		// unknown sequence are just appended to the line
		if !key.IsSeq { // @TODO should we ignore unknown sequences?
			line.insert(key.Value)
		} else {
			switch key.Value {
			case "ctrl+c": // cancel
				return MsgKill{}, ErrSIGINT
			case "ctrl+d": // delete / or cancel if empty
				if line.value == "" {
					return MsgLine{}, ErrEOF
				}
				line.delete()
			case "enter": // return the line
				return MsgLine{Value: line.value}, nil
			case "left": // move left
				line.moveLeft()
			case "right": // move right
				line.moveRight()
			case "backspace": // delete previous character
				line.deleteBackward()
			case "delete": // delete
				line.delete()
			case "home", "ctrl+a": // move to start of line
				line.moveStart()
			case "end", "ctrl+e": // move to end of line
				line.moveEnd()
			case "ctrl+l": // clear line
				line.clear()
			case "ctrl+u": // delete from cursor to start of line
				line.deleteToStart()
			case "ctrl+k": // delete from cursor to end of line
				line.deleteToEnd()
			case "ctrl+w", "alt+backspace": // delete previous word
				line.deleteWordBackward()
			case "alt+d", "ctrl+delete", "alt+delete": // delete next word
				line.deleteWord()
			case "alt+b", "ctrl+left", "alt+left": // move to previous word
				line.moveWordLeft()
			case "alt+f", "ctrl+right", "alt+right": // move to next word
				line.moveWordRight()
			case "tab":
				if !hasCompletionProvider {
					fmt.Fprint(tty, "\a") // terminal bell
				} else {
					if !line.completing { // start completing
						suggestions, err := completionProvider.ReadlineCompletion(line.completionStart())
						if err != nil {
							line.completionEnd()
							return MsgLine{Value: line.value}, fmt.Errorf("%w: %w", ErrComp, err)
						}
						if len(suggestions) > 0 {
							line.completionSuggests(suggestions)
						} else {
							line.completionEnd()
							fmt.Fprint(tty, "\a") // terminal bell
						}
					} else {
						line.completionNext()
					}
				}
			}

		}

	}
}

// This is an advanced version of ReadLine that allow to provide a value to edit.
// It requires a terminal which supports raw mode and is a tty to work.
// Most common keyboard shortcuts should be supported for navigation and editing.
//
// The provider can be nil or an object that implements one or more of the
// following interfaces:
//   - readlineValueProvider to provide a value to the line editor
//   - readlineKeyHandlerProvider to provide custom key bindings
//
// Providing a value to edit is possible by providing an object that implements
// interface{ReadlineValue() string}. If the given object does not implement
// this interface the value will be ignored.
//
// The provider can also implements interface{ReadlineKeyHandler() ui.Msg}
// to provide custom key bindings. it will be called for each key press with the
// key as argument.
// The method should return a ui.Msg (can be any value) if the key was handled
// and nil if the default behavior should be executed.
// If ReadlineKeyHandler returns a ui.Msg it will be returned by ReadLineEnhanced.
// If ReadlineKeyHandler returns nil the default behavior will be executed and the
// editor will continue to wait for input.
//
// Note: this function is made public for advanced usage only.
// You shouldn't need to use it directly in most cases but simply define a model
// implementing the ComponentApi with InputReader: ui.LineReader.
func ReadLineEnhanced(terminal interface {
	TermIsTerminal
	TermWithRawMode
	TermWithReader
	TTYFileDescriptor
}, provider any) (Msg, error) {
	return readLineEnhanced(terminal, provider, false)
}

// This is an advanced version of ReadLine which don't print to output.
// It requires a terminal which supports raw mode and is a tty to work.
// Most common keyboard shortcuts should be supported for navigation and editing.
//
// Note: this function is made public for advanced usage only.
// You shouldn't need to use it directly in most cases but simply define a model
// implementing the ComponentApi with InputReader: ui.PasswordReader.
func ReadPassword(terminal interface {
	TermIsTerminal
	TermWithRawMode
	TermWithReader
	TTYFileDescriptor
}) (Msg, error) {
	return readLineEnhanced(terminal, nil, true)
}

/** read a line from stdin */
func Readline(prompt string) (MsgLine, error) {
	term := GetTerminal()
	if !term.IsTerminal() {
		return MsgLine{}, ErrNOTERM
	}
	scanner := term.NewScanner()
	fmt.Print(prompt)
	scanner.Scan()
	return MsgLine{Value: scanner.Text()}, scanner.Err()
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
