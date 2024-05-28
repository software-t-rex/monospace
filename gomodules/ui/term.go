/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/software-t-rex/monospace/gomodules/ui/pkg/ansi"
	"golang.org/x/term"
)

var (
	ErrNOTERM    = fmt.Errorf("not a terminal")
	ErrTimeout   = fmt.Errorf("timeout")
	ErrBadFormat = fmt.Errorf("bad response format")
)

type (
	TermIsTerminal interface {
		IsTerminal() bool
	}
	TermWithRawMode interface {
		MakeRaw() (*TermState, error)
		Restore(*TermState) error

		// put terminal in raw mode and return a restore function and an error if any
		//
		// Usage:
		//
		//	restore, err := handleTerminalState(terminal)
		//	defer restore()
		HandleState(hideCursor bool) (func(), error)
	}
	TermWithBackground interface {
		// WARNING: you should have checked that the terminal is a terminal before calling this function.
		// you can do it like this term.IsTerminal(int(os.stdin.Fd()))
		// Try to use OSC query to get the background color.
		// it is the most reliable method to get the background color when available,
		// but it is not supported in all terminals.
		// If it fails, it will default to the COLORFGBG environment variable.
		// but it is not reliable for example Konsole keep reporting the wrong
		// background color after a change.
		HasDarkBackground() (bool, error)
	}
	TermWithExclusiveReader interface {
		// Returns a mutex locked reader on the terminal input and a function to unlock it when done
		// this is useful to avoid concurrent access to the reader.
		// You should have checked that the terminal is a terminal before calling this function.
		ExclusiveReader() (*bufio.Reader, func())
	}
	TermWithSize interface {
		// GetSize returns the number of columns and rows in the terminal.
		GetSize() (int, int, error)
		// DeviceStatusReport returns the current cursor position in the terminal.
		DeviceStatusReport() (row, col int, err error)
	}
	TTYFileDescriptor interface {
		Tty() *os.File
	}
	TermInterface interface {
		TermIsTerminal
		TermWithRawMode
		TermWithBackground
		TermWithExclusiveReader
		TermWithSize
		TTYFileDescriptor
	}
	TermState state
)

type Terminal struct {
	tty       *os.File
	fd        int
	isTerm    bool
	isRaw     bool
	hasDarkBg bool
	bgChecked bool
	reader    *bufio.Reader
	readMutex sync.Mutex
}

func NewTerminal(tty *os.File) (*Terminal, error) {
	terminal := &Terminal{
		hasDarkBg: true, // default to most common case
		fd:        -1,
	}
	if tty != nil {
		terminal.tty = tty
	} else {
		file, err := openTTY()
		if err != nil {
			terminal.bgChecked = true
			return terminal, err
		}
		terminal.tty = file
	}
	terminal.fd = int(terminal.tty.Fd())
	terminal.isTerm = term.IsTerminal(terminal.fd)
	if !terminal.isTerm {
		return terminal, fmt.Errorf("file descriptor %d is %w", terminal.fd, ErrNOTERM)
	}
	terminal.reader = bufio.NewReader(terminal.tty)
	return terminal, nil
}

func (t *Terminal) IsTerminal() bool {
	return t.isTerm
}
func (t *Terminal) MakeRaw() (*TermState, error) {
	if t.isRaw {
		return nil, nil
	}
	if !t.isTerm {
		return nil, fmt.Errorf("MakeRaw: %w", ErrNOTERM)
	}
	state, err := makeRaw(t.fd)
	if err == nil {
		t.isRaw = true
	}
	return state, err
}
func (t *Terminal) Restore(state *TermState) error {
	if state == nil {
		return nil
	}
	if !t.isTerm {
		return fmt.Errorf("Restore: %w", ErrNOTERM)
	}
	t.isRaw = false
	return restore(t.fd, state)
}

func (t *Terminal) HandleState(hideCursor bool) (func(), error) {
	restore := func() {}
	state, err := t.MakeRaw()
	if errors.Is(err, ErrNOTERM) {
		return restore, err
	}
	if hideCursor && err == nil {
		fmt.Fprint(os.Stdout, ansi.CtrlHideCursor)
	}
	restore = func() {
		t.Restore(state) // error will panic
		if hideCursor && err == nil {
			fmt.Fprint(os.Stdout, ansi.CtrlShowCursor)
		}
	}
	return restore, err
}

func (t *Terminal) GetSize() (int, int, error) {
	if !t.isTerm {
		return 0, 0, fmt.Errorf("GetSize: %w", ErrNOTERM)
	}
	return term.GetSize(int(os.Stdout.Fd()))
}

func (t *Terminal) Tty() *os.File {
	return t.tty
}

func (t *Terminal) ExclusiveReader() (*bufio.Reader, func()) {
	if !t.isTerm {
		panic(fmt.Errorf("ExclusiveReader: %w", ErrNOTERM))
	}
	t.readMutex.Lock()
	return t.reader, func() { t.readMutex.Unlock() }
}

func (t *Terminal) HasDarkBackground() (bool, error) {
	if !t.isTerm {
		return true, fmt.Errorf("HasDarkBackground: %w", ErrNOTERM)
	}
	if !t.bgChecked {
		t.hasDarkBg = terminalHasDarkBackground(t)
		t.bgChecked = true
	}
	return t.hasDarkBg, nil
}

func (t *Terminal) SetBracketedPasteMode(on bool) error {
	if !t.isTerm {
		return fmt.Errorf("SetBracketedPasteMode: %w", ErrNOTERM)
	}
	err := error(nil)
	if on {
		_, err = ansi.CtrlBracketedPasteModeOn.Print()
	} else {
		_, err = ansi.CtrlBracketedPasteModeOff.Print()
	}
	return fmt.Errorf("SetBracketedPasteMode: %w", err)
}

func (t *Terminal) AltScreenBuffer() error {
	if !t.isTerm {
		return fmt.Errorf("AltScreenBuffer: %w", ErrNOTERM)
	}
	_, err := ansi.CtrlAltScreenBuffer.Print()
	return fmt.Errorf("AltScreenBuffer: %w", err)
}
func (t *Terminal) MainScreenBuffer() error {
	if !t.isTerm {
		return fmt.Errorf("MainScreenBuffer: %w", ErrNOTERM)
	}
	_, err := ansi.CtrlMainScreenBuffer.Print()
	return fmt.Errorf("MainScreenBuffer: %w", err)
}

func (t *Terminal) DeviceStatusReport() (row, col int, err error) {
	res, err := terminalQuery(t, terminalQueryOpts{
		querySequence:  ansi.CtrlDeviceStatusReport.String(),
		expectedPrefix: csiStart,
		endVerifier: func(b byte, _ []byte) (bool, error) {
			if b == 'R' {
				return true, nil
			} else if b == '\a' || b == '\x1b' || b == '\n' || b == '\r' || b == '\\' {
				return false, ErrBadFormat
			}
			return false, nil
		},
		maxLen: 10,
	})
	if err != nil {
		return 0, 0, err
	}
	rc := strings.Split(strings.TrimSuffix(res, "R"), ";")
	if len(rc) != 2 {
		return 0, 0, ErrBadFormat
	}
	if rc[1] != "" {
		col, err = strconv.Atoi(rc[1])
		if err != nil {
			return 0, 0, err
		}
	}
	if rc[0] != "" {
		row, err = strconv.Atoi(rc[0])
		if err != nil {
			return 0, col, err
		}
	}
	return row, col, nil
}

type terminalQueryOpts struct {
	// the sequence to send to the terminal
	querySequence string
	// the prefix of the response to expect this allow early exit if response is not what is expected
	expectedPrefix string
	// verifier to check if we should stop reading receives the last read byte and the response read so far.
	// It should return true and an error if any to stop reading and false, nil otherwhise.
	endVerifier func(byte, []byte) (bool, error)
	// timeout to wait for the response (default to 150ms)
	timeout time.Duration
	// res maxLen of the response to read (default to 25)
	maxLen int
}

// # this is a helper to query terminal for specific sequence
//
//   - querySequence is the sequence to send to the terminal
//   - expectedPrefix is the prefix of the response to expect this allow early exit if response is not what is expected
//   - endVerifier is a function to check if we should stop reading receives the response so far and should return true if we should stop reading
//     and false if we should continue reading the response. It also returns an error if any.
//
// # Notes:
//   - this function will put the terminal in raw mode and will restore it before returning.
//   - this function will read from the terminal in a goroutine and will return the response or an error if any.
//     the terminal implementation is responsible to protect the reader from concurrent access through the ExclusiveReader method.
//   - byte read after a timeout is restored to the reader so that it can be read again.
func terminalQuery(
	terminal interface {
		TermWithRawMode
		TermWithExclusiveReader
	},
	opts terminalQueryOpts,
) (string, error) {
	// setting default opts
	if opts.maxLen == 0 {
		opts.maxLen = 25
	}
	if opts.timeout <= 0 {
		opts.timeout = 150 * time.Millisecond
	}
	// ensure terminal is rawMode
	restore, err := terminal.HandleState(false)
	defer restore()
	if err != nil {
		return "", err
	}
	// print the query sequence
	fmt.Print(opts.querySequence)
	responseChan := make(chan string, 1)
	errChan := make(chan error, 1)
	canceled := new(atomic.Bool)
	go func() { // read the response in a goroutine
		reader, readReleaser := terminal.ExclusiveReader()
		response := make([]byte, 0, opts.maxLen) // should be enough for most responses
		for {
			if canceled.Load() {
				readReleaser()
				return
			}
			b, err := reader.ReadByte() // blocking read
			if err != nil {
				readReleaser()
				errChan <- err
				return
			}
			if canceled.Load() {
				// this is case where timeout happend during the blocking read
				reader.UnreadByte()
				readReleaser()
				return
			}
			response = append(response, b)
			// if start of the response is not the expected one then consider a bad response
			if len(response) < len(opts.expectedPrefix) {
				if b != opts.expectedPrefix[len(response)-1] {
					readReleaser()
					errChan <- fmt.Errorf("%w: %#v", ErrBadFormat, string(response)) // considered as bad response
					return
				}
				continue
			}
			if len(response) > opts.maxLen {
				readReleaser()
				errChan <- fmt.Errorf("%w: too long", ErrBadFormat) // consider as bad response
				return
			}

			stop, errVerifer := opts.endVerifier(b, response)
			if stop || errVerifer != nil {
				readReleaser()
				if errVerifer != nil {
					errChan <- fmt.Errorf("%w: %w", ErrBadFormat, errVerifer) // consider as bad response
					return
				}
				responseChan <- strings.TrimPrefix(string(response), opts.expectedPrefix) // consider as end of response
				return
			}
		}
	}()
	select {
	case response := <-responseChan:
		return response, nil
	case err := <-errChan:
		return "", err
	case <-time.After(opts.timeout):
		canceled.Store(true)
		return "", ErrTimeout
	}
}

var usedTerm TermInterface

// this should be mocked in tests
func SetTerminal(t TermInterface) {
	usedTerm = t
}
func GetTerminal() TermInterface {
	if usedTerm == nil {
		usedTerm, _ = NewTerminal(nil)
	}
	return usedTerm
}
