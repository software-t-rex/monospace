/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/software-t-rex/monospace/gomodules/ui/pkg/lineEditor"
	"github.com/software-t-rex/monospace/gomodules/ui/pkg/sequencesKeys"
)

type inputReader int

const (
	// this is the default reader it will send MsgKey
	KeyReader inputReader = iota
	// LineReader will send MsgLine (same as LineReaderFramed if terminal width is available else default to LineReaderUnbounded)
	LineReader
	// LineReaderUnbounded will send MsgLine (doesn't manage terminal width)
	// this is more a fallback mode when the terminal width is not available
	LineReaderUnbounded
	// LineReaderWrapped will send MsgLine
	// it will wrap the text if it's too long for the terminal width
	// will default to LineReaderUnbounded if the terminal width is not available
	//
	// suitable to edit long text at the end of display (doesn't support new lines or carriage returns)
	//
	// WARNING: this mode is still highly experimental and may not work as expected
	//          any help to improve it is welcome
	LineReaderWrapped
	// this will send MsgLine, text will be displayed in a windowed mode
	// this means that the text will be displayed on a single line and
	// scroll horizontally if it's too long for the terminal width
	// will default to LineReader if the terminal width is not available
	//
	// suitable to edit long text with other content displayed after
	lineReaderFramed
	// this will send MsgLine and will not manage any display of the value
	// suitable for password input
	PasswordReader
	// this will send MsgInt
	IntReader
	// this will send MsgInts
	IntsReader
)

var lineReaderVisualMode = map[inputReader]lineEditor.VisualEditMode{
	LineReader:          lineEditor.VisualEditMode(-1), // auto
	LineReaderUnbounded: lineEditor.VModeUnbounded,
	LineReaderWrapped:   lineEditor.VModeWrappedLine,
	lineReaderFramed:    lineEditor.VModeFramedLine,
	PasswordReader:      lineEditor.VModeMaskedLine,
}

type (
	// MsgKey is a message sent by the KeyReader
	// it can be a known key press event or an unknown sequence
	// if it's an unknown sequence the Value will be the raw sequence
	// and the Unknown field will be set to true
	// if it's a known sequence the Value will be the name of the key
	// and the IsSeq field will be set to true
	// if it's a paste event the IsSeq will be set to true and the Value will be "paste"
	// and the pasted value will be returned by the IsPasteEvent method
	MsgKey struct {
		Value   string
		Unknown bool
		IsSeq   bool
		ByteSeq []byte
	}
	MsgLine interface {
		// Value returns the value of the line
		Value() string
		// return a string representation of the line as displayed while editing
		Sprint() (string, error)
	}
	MsgLineEnhanced struct {
		line *lineEditor.LineEditor
	}
	MsgLineSimple struct {
		value string
	}
	MsgInt struct {
		Value int
	}
	MsgInts struct {
		Value []int
	}
	// re-export the lineEditor.LineEditorOptions for convenience
	LineEditorOptions      lineEditor.LineEditorOptions
	readlineConfigProvider interface {
		ReadlineConfig() LineEditorOptions
	}
	readlinekeyHandlerProvider interface {
		// the keyHandler is called when a key is pressed in enhanced mode
		// it should return a message to update the model or an error
		// if nil,nil is returned then the normal input handling will be done
		// if a non nil msg or a non nil error is returned then the input will be considered handled
		// and enhanced readline will simply ignore the key
		// (only used in enhanced mode)
		ReadlineKeyHandler(string) (Msg, error)
	}
	readlineCompletionProvider interface {
		// ReadlineCompletion is called by the ReadlineEnhanced function to get completion suggestions
		//   - wordStart is the start of the word to complete (start of the word to cursor position)
		//   - word is the whole word under cursor (start to end of the word)
		ReadlineCompletion(string, string) ([]string, error)
	}
	// readLineHistoryProvider interface {
	// 	ReadLineHistory(n int) ([]string, error)
	// }
)

func (msg MsgKey) String() string { return msg.Value }

// test if the key press event is a paste event
// if it is a paste event it will return true and the pasted value as []rune
func (msg MsgKey) IsPasteEvent() (bool, []rune) {
	isPasteEvent := msg.IsSeq && string(msg.ByteSeq[:6]) == "\x1b[200~" && string(msg.ByteSeq[len(msg.ByteSeq)-6:]) == "\x1b[201~"
	var pastedValue []rune
	if isPasteEvent {
		pastedValue = lineEditor.Sanitize([]rune(string(msg.ByteSeq[6:len(msg.ByteSeq)-6])), true)
	}
	return isPasteEvent, pastedValue
}

func (msg MsgLineEnhanced) Value() string { return msg.line.GetStringValue() }
func (msg MsgLineEnhanced) Sprint() (string, error) {
	return msg.line.Sprint()
}
func (msg MsgLineSimple) Value() string { return msg.value }
func (msg MsgLineSimple) Sprint() (string, error) {
	return msg.value, nil
}

var (
	ErrSIGINT       = fmt.Errorf("SIGINT")
	ErrReadKey      = fmt.Errorf("read key error")
	ErrPasteTimeout = fmt.Errorf("%w: paste timeout", ErrReadKey)
	ErrUnknownKey   = fmt.Errorf("%w: unknown key", ErrReadKey)
	ErrComp         = fmt.Errorf("completion error")
	ErrNaN          = fmt.Errorf("not a number")
)

var pasteTimeoutDuration = 1 * time.Second

// you MUST already have a terminal with activated raw mode
// and bracketed paste mode should be activated to detect paste events
func readKeyPressEvent(reader *bufio.Reader) (MsgKey, error) {
	var buf bytes.Buffer
	var msgKey MsgKey
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return msgKey, fmt.Errorf("%w: %w", ErrReadKey, err)
		}
		buf.WriteByte(b)
		bufLen := buf.Len()
		if msgKey.Value == "paste" { // inside bracketed paste (buf already contains one more byte at this point)
			pasteTimer := time.NewTimer(pasteTimeoutDuration)
			// @implement go routines to read without blocking the main loop
			responseChan := make(chan []byte, 1)
			errChan := make(chan error, 1)
			go func() {
				pasteContent, err := reader.ReadBytes('~')
				if err != nil {
					errChan <- err
					return
				}
				responseChan <- pasteContent
			}()
			for {
				select {
				case <-pasteTimer.C: // paste event takes too long we will consider this as an unknown sequence
					msgKey.Value = "esc"
					msgKey.ByteSeq = buf.Bytes()
					msgKey.Unknown = true
					return msgKey, fmt.Errorf("%w: %w", ErrUnknownKey, ErrPasteTimeout)
				case err := <-errChan:
					return msgKey, fmt.Errorf("%w: %w", ErrReadKey, err)
				case pasteContent := <-responseChan:
					pasteTimer.Stop()
					msgKey.Value = "paste"
					msgKey.ByteSeq = append(buf.Bytes(), pasteContent...)
					return msgKey, nil
				}
			}
		}
		// is buf content considered a known escape sequence (such as tab, enter, etc...)
		if buf.Bytes()[0] != '\x1b' || bufLen > 1 {
			if name, exists := sequencesKeys.Map[buf.String()]; exists {
				msgKey.Value = name
				msgKey.IsSeq = true
				if name == "paste" { // should read until the end of the paste event
					continue
				}
				msgKey.ByteSeq = buf.Bytes()
				return msgKey, nil
			}
		}

		if bufLen == 1 && b == '\x1b' { // start of an escape sequence
			msgKey.IsSeq = true
			if reader.Buffered() == 0 { // check if we have more bytes to read
				msgKey.Value = "esc"
				msgKey.ByteSeq = buf.Bytes()
				return msgKey, nil
			}
			// more bytes to read we continue to read the sequence
			continue
		}
		if msgKey.IsSeq { // inside an escape sequence
			if b == '\x1b' { // we should not have any escape key inside the sequence
				msgKey.Value = "unknown"
				msgKey.ByteSeq = buf.Bytes()[:bufLen-1] // discard the last escape
				reader.UnreadByte()                     // restore the last escape for another read
				msgKey.Unknown = true
				return msgKey, fmt.Errorf("%w: %#v", ErrUnknownKey, string(msgKey.ByteSeq))
			}
			if _, exists := sequencesKeys.PartialLookupMap[string(buf.Bytes())]; !exists || bufLen > sequencesKeys.MaxLen {
				msgKey.Value = "unknown"
				msgKey.ByteSeq = buf.Bytes()
				msgKey.Unknown = true
				return msgKey, fmt.Errorf("%w: %#v", ErrUnknownKey, buf.String())
			}
			continue
		}
		if bufLen > 4 { // utf8 max length is 4 bytes
			msgKey.Value = buf.String()
			msgKey.ByteSeq = buf.Bytes()
			msgKey.Unknown = true
			return msgKey, fmt.Errorf("%w: %#v", ErrUnknownKey, buf.String())
		}
		if utf8.FullRune(buf.Bytes()) {
			msgKey.Value = buf.String()
			msgKey.ByteSeq = buf.Bytes()
			return msgKey, nil
		}
	}
}

func ReadKeyPressEvent(terminal interface {
	TermIsTerminal
	TermWithRawMode
	TermWithExclusiveReader
	TTYFileDescriptor
}) (Msg, error) {
	restore, err := terminal.HandleState(true)
	defer restore()
	if err != nil {
		return nil, err
	}
	exclusiveReader, releaseReader := terminal.ExclusiveReader()
	msg, err := readKeyPressEvent(exclusiveReader)
	releaseReader()
	return msg, err
}

func readlineGetVisualMode(visualMode lineEditor.VisualEditMode, termWidth int) lineEditor.VisualEditMode {
	if visualMode == -1 { // auto mode
		if termWidth < 1 {
			visualMode = lineEditor.VModeUnbounded
		} else {
			visualMode = lineEditor.VModeFramedLine
		}
	} else if termWidth < 1 && visualMode != lineEditor.VModeHidden {
		// if we don't have the terminal width we default to unbounded line
		visualMode = lineEditor.VModeUnbounded
	}
	return visualMode
}

// return the line typed by the user or an error if any
// common terminal shortcuts are supported:
// - ctrl+c to cancel return empty MsgKill and error "SIGINT"
// - ctrl+d delete or if empty return empty Msg and error "EOF"
// - ctrl+l to clear line value
// - ctrl+u to delete from cursor to start of line
// - ctrl+k to delete from cursor to end of line
// - ctrl+w/alt+backspace to delete previous word
// - ctrl+delete/alt+d/alt+delete to delete next word
// - ctrl+left/alt+left/alt+b to move to previous word
// - ctrl+right/alt+right/alt+f to move to next word
// - ctrl+a/home to move to start of line
// - ctrl+e/end to move to end of line
// - left to move left
// - right to move right
// - backspace to delete previous character
// - delete to delete
// - enter to return the line
// - paste operation are supported only if terminal is set to bracketed paste mode
func readLineEnhanced(terminal interface {
	TermIsTerminal
	TermWithRawMode
	TermWithExclusiveReader
	TermWithSize
}, visualMode lineEditor.VisualEditMode, provider any) (Msg, error) {
	restoreState, errState := terminal.HandleState(false)
	defer restoreState()
	if errState != nil {
		return nil, fmt.Errorf("readlineEnhanced: %w", errState)
	}
	terminalWithBracketedPasteMode, hasBracketedPasteMode := terminal.(interface{ SetBracketedPasteMode(bool) error })
	if hasBracketedPasteMode {
		errBracketedPasteMode := terminalWithBracketedPasteMode.SetBracketedPasteMode(true)
		if errBracketedPasteMode != nil {
			defer terminalWithBracketedPasteMode.SetBracketedPasteMode(false)
		}
	}

	// check for termWidth and set visualMode accordingly (if not forced)
	termWidth, _, getSizeErr := terminal.GetSize()
	if getSizeErr != nil {
		// default to 80 columns if we can't get the terminal width
		termWidth = 80
	}
	visualMode = readlineGetVisualMode(visualMode, termWidth)
	// set visual offset
	line := lineEditor.NewLineEditor(os.Stdout, visualMode).SetTermWidth(termWidth)

	_, col, errReport := terminal.DeviceStatusReport()
	if errReport == nil {
		line.SetVisualStartOffset(col - 1)
	}

	var keyHandlerProvider readlinekeyHandlerProvider
	var haskeyHandlerProvider bool
	var hasCompletionProvider bool
	var completionProvider readlineCompletionProvider
	if provider != nil {
		configProvider, hasConfigProvider := any(provider).(readlineConfigProvider)
		keyHandlerProvider, haskeyHandlerProvider = any(provider).(readlinekeyHandlerProvider)
		completionProvider, hasCompletionProvider = any(provider).(readlineCompletionProvider)
		if hasConfigProvider {
			config := configProvider.ReadlineConfig()
			line.SetConfig(lineEditor.LineEditorOptions(config))
		}
	}
	exclusiveReader, releaseReader := terminal.ExclusiveReader()
	for {
		keyPressEvt, errEvt := readKeyPressEvent(exclusiveReader)
		if errEvt != nil && !errors.Is(errEvt, ErrUnknownKey) { // ignore ErrReadKey
			releaseReader()
			return MsgLineEnhanced{line: line}, errEvt
		}
		if haskeyHandlerProvider {
			msg, errHandler := keyHandlerProvider.ReadlineKeyHandler(keyPressEvt.Value)
			if msg != nil || errHandler != nil {
				releaseReader()
				return msg, errHandler
			}
		}
		if line.IsCompleting() && keyPressEvt.Value != "tab" {
			line.CompletionEnd()
		}
		// if evt is not a sequence we insert it in the line
		if !keyPressEvt.IsSeq {
			if !keyPressEvt.Unknown { // ignore evt marked as unkown
				line.Insert([]rune(keyPressEvt.Value))
			}
		} else {
			errCmd := error(nil)
			switch keyPressEvt.Value {
			case "ctrl+c": // cancel
				releaseReader()
				return MsgKill{}, ErrSIGINT
			case "ctrl+d": // delete / or cancel if empty
				if line.Len() < 1 {
					releaseReader()
					return MsgLineEnhanced{line: line}, io.EOF
				}
				errCmd = line.Delete()
			case "enter": // return the line
				releaseReader()
				return MsgLineEnhanced{line: line}, nil
			case "up":
				if visualMode == lineEditor.VModeHidden {
					line.RingBell()
				} else if visualMode == lineEditor.VModeWrappedLine {
					errCmd = line.MoveUp()
				}
				// @fixme implement history
			case "down":
				if visualMode == lineEditor.VModeHidden {
					line.RingBell()
				} else if visualMode == lineEditor.VModeWrappedLine {
					errCmd = line.MoveDown()
				}
				// @fixme implement history
			case "left": // move left
				errCmd = line.MoveLeft()
			case "right": // move right
				errCmd = line.MoveRight()
			case "backspace": // delete previous character
				errCmd = line.DeleteBackward()
			case "delete": // delete
				errCmd = line.Delete()
			case "home", "ctrl+a": // move to start of line
				errCmd = line.MoveStartOfLine()
			case "end", "ctrl+e": // move to end of line
				errCmd = line.MoveEndOfLine()
			case "ctrl+home": // move to start of content
				errCmd = line.MoveStart()
			case "ctrl+end": // move to end of content
				errCmd = line.MoveEnd()
			case "ctrl+l": // clear line
				errCmd = line.Clear()
			case "ctrl+u": // delete from cursor to start of line
				errCmd = line.DeleteToStart()
			case "ctrl+k": // delete from cursor to end of line
				errCmd = line.DeleteToEnd()
			case "ctrl+w", "alt+backspace": // delete previous word
				errCmd = line.DeleteWordBackward()
			case "alt+d", "ctrl+delete", "alt+delete": // delete next word
				errCmd = line.DeleteWord()
			case "alt+b", "ctrl+left", "alt+left": // move to previous word
				errCmd = line.MoveWordLeft()
			case "alt+f", "ctrl+right", "alt+right": // move to next word
				errCmd = line.MoveWordRight()
			case "paste":
				_, pastedValue := keyPressEvt.IsPasteEvent()
				// we don't support new lines or tab in the content of a line so we replace them with single space
				for k, v := range pastedValue {
					if unicode.IsSpace(v) {
						pastedValue[k] = ' '
					}
				}
				errCmd = line.Insert(pastedValue)
				if errors.Is(errCmd, lineEditor.ErrMaxLen) {
					errCmd = nil
				}
			case "tab":
				if !hasCompletionProvider {
					line.RingBell()
				} else {
					if !line.IsCompleting() { // start completing
						startOfWordToComp, wordToComp := line.CompletionStart()
						if startOfWordToComp != "" {
							suggestions, err := completionProvider.ReadlineCompletion(startOfWordToComp, wordToComp)
							if err != nil {
								line.CompletionEnd()
								releaseReader()
								return MsgLineEnhanced{line: line}, fmt.Errorf("%w: %w", ErrComp, err)
							}
							if len(suggestions) > 0 {
								line.CompletionSuggests(append(suggestions, wordToComp))
							} else {
								line.CompletionEnd()
								line.RingBell()
							}
						}
					} else {
						errCmd = line.CompletionNext()
						if errors.Is(errCmd, lineEditor.ErrMaxLen) { // ignore max len error on completion
							errCmd = nil
						}
					}
				}
			}
			if errCmd != nil {
				releaseReader()
				return MsgLineEnhanced{line: line}, errCmd
			}
		}
	}
}

// This is an advanced version of ReadLine that allow to provide a value to edit.
// It requires a terminal which supports raw mode and is a tty to work.
// Most common keyboard shortcuts should be supported for navigation and editing.
//
// The visualMode determines how the text will be displayed to the user.
// see [VModeHidden][VModeUnbounded][VModeWrappedLine][VModeFramedLine]
//
// The provider can be nil or an object that implements one or more of the
// following interfaces:
//   - readlineConfigProvider to provide some config to the line editor
//   - readlineKeyHandlerProvider to provide custom key bindings
//   - readlineCompletionProvider to provide completion suggestions
//
// Providing a value to edit is possible by providing an object that implements
// interface{ReadlineConfig() LineEditorOptions}. If the given object does not implement
// this interface the value will be ignored. see [LineEditorOptions] for more details.
//
// The provider can also implements interface{ReadlineKeyHandler(key string) (ui.Msg, error)}
// to provide custom key bindings. it will be called for each key press with the
// key as argument.
// The method should return a ui.Msg (can be any value) to update model or an error
//   - if nil,nil is returned then the normal input handling will be done and
//     editor will continue to wait for input (unless the key is a completion key)
//   - if a non nil msg or a non nil error is returned then the input will be considered handled
//     and the editor will simply ignore the key and return the ui.Msg and error stopping the editor.
//
// It is also possible to provide auto-completion feature by implementing
// interface{ReadlineCompletion(wordStart, word string) ([]string, error)}.
// This method will be called when the user press the completion key (tab by default).
// It should return a list of suggestions for the word under the cursor.
// The wordStart is the start of the word to complete (start of the word to cursor position)
// and the word is the whole word under cursor (start to end of the word).
//
// Note: this function is made public for advanced usage only.
// You shouldn't need to use it directly in most cases but simply define a model
// implementing the ComponentApi with InputReader: ui.LineReader.
func ReadLineEnhanced(terminal interface {
	TermIsTerminal
	TermWithRawMode
	TermWithExclusiveReader
	TermWithSize
	TTYFileDescriptor
}, visualMode lineEditor.VisualEditMode, provider any) (Msg, error) {
	return readLineEnhanced(terminal, visualMode, provider)
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
	TermWithExclusiveReader
	TermWithSize
	TTYFileDescriptor
}) (Msg, error) {
	return ReadLineEnhanced(terminal, lineEditor.VModeMaskedLine, nil)
}

/** read a line from stdin */
func Readline(prompt string) (MsgLine, error) {
	term := GetTerminal()
	if !term.IsTerminal() {
		return MsgLineSimple{}, ErrNOTERM
	}
	reader, releaseReader := term.ExclusiveReader()
	defer releaseReader()
	fmt.Print(prompt)
	str, err := reader.ReadString('\n')
	return MsgLineSimple{value: strings.TrimRight(str, "\r\n")}, err
}

func ReadInt(prompt string) (MsgInt, error) {
	input, err := Readline(prompt)
	if err != nil {
		return MsgInt{}, err
	}
	v, err := strconv.Atoi(input.Value())
	if err != nil {
		return MsgInt{}, fmt.Errorf("%w:%w", ErrNaN, err)
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
			return MsgInts{}, fmt.Errorf("%w:%w", ErrNaN, err)
		}
		ints[k] = i
	}
	return MsgInts{Value: ints}, nil
}

func ReadInts(prompt string) (MsgInts, error) {
	input, err := Readline(prompt)
	if err != nil {
		return MsgInts{}, err
	}
	return ParseInts(input.Value())
}
