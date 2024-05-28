/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ansi

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type ANSIControlSequence string

func (c ANSIControlSequence) String() string                  { return string(c) }
func (c ANSIControlSequence) Print() (int, error)             { return fmt.Print(c) }
func (c ANSIControlSequence) Fprint(w io.Writer) (int, error) { return fmt.Fprint(w, string(c)) }

type ANSIControlSequenceWithParams string

func (c ANSIControlSequenceWithParams) Sprintf(s ...any) string { return fmt.Sprintf(string(c), s...) }
func (c ANSIControlSequenceWithParams) Printf(s ...any) (int, error) {
	return fmt.Printf(string(c), s...)
}
func (c ANSIControlSequenceWithParams) Fprintf(w io.Writer, s ...any) (int, error) {
	return fmt.Fprintf(w, string(c), s...)
}

var (
	csiStart                      = "\x1b["
	CtrlUp                        = ANSIControlSequenceWithParams(csiStart + "%dA")    // cursor up (n lines)
	CtrlDown                      = ANSIControlSequenceWithParams(csiStart + "%dB")    // cursor down (n lines)
	CtrlForward                   = ANSIControlSequenceWithParams(csiStart + "%dC")    // cursor forward (n chars)
	CtrlBackward                  = ANSIControlSequenceWithParams(csiStart + "%dD")    // cursor backward (n chars)
	CtrlNextLine                  = ANSIControlSequenceWithParams(csiStart + "%dE")    // cursor next line (n lines)
	CtrlPrevLine                  = ANSIControlSequenceWithParams(csiStart + "%dF")    // cursor previous line
	CtrlHorizAbs                  = ANSIControlSequenceWithParams(csiStart + "%dG")    // cursor horizontal absolute (col)
	CtrlPos                       = ANSIControlSequenceWithParams(csiStart + "%d;%dH") // cursor position (row, col)
	CtrlEraseD                    = ANSIControlSequenceWithParams(csiStart + "%dJ")    // erase in display (0: from cursor to end, 1: from start to cursor, 2: all)
	CtrlEraseL                    = ANSIControlSequenceWithParams(csiStart + "%dK")    // erase in line (0: from cursor to end, 1: from start to cursor, 2: all)
	CtrlScrollUp                  = ANSIControlSequenceWithParams(csiStart + "%dS")    // scroll up (n lines)
	CtrlScrollDn                  = ANSIControlSequenceWithParams(csiStart + "%dT")    // scroll down (n lines)
	CtrlHVPos                     = ANSIControlSequenceWithParams(csiStart + "%d;%df") // horizontal and vertical position (row, col)
	CtrlDeviceStatusReport        = ANSIControlSequence(csiStart + "6n")               // device status report
	CtrlSaveCursorPosition_SCO    = ANSIControlSequence(csiStart + "s")                // save cursor position (SCO)
	CtrlRestoreCursorPosition_SCO = ANSIControlSequence(csiStart + "u")                // restore cursor position (SCO)
	CtrlSaveCursorPosition_DEC    = ANSIControlSequence(csiStart + "7")                // save cursor position (DEC)
	CtrlRestoreCursorPosition_DEC = ANSIControlSequence(csiStart + "8")                // restore cursor position (DEC)
	CtrlShowCursor                = ANSIControlSequence(csiStart + "?25h")             // show cursor
	CtrlHideCursor                = ANSIControlSequence(csiStart + "?25l")             // hide cursor
	CtrlAltScreenBuffer           = ANSIControlSequence(csiStart + "?1049h")           // use alternate screen buffer
	CtrlMainScreenBuffer          = ANSIControlSequence(csiStart + "?1049l")           // use main screen buffer
	CtrlBracketedPasteModeOn      = ANSIControlSequence(csiStart + "?2004h")           // bracketed paste mode on
	CtrlBracketedPasteModeOff     = ANSIControlSequence(csiStart + "?2004l")           // bracketed paste mode off
)

// send a CSI DSR query and returns the result of the query device status report
func QueryDeviceStatusReport() (row, col int, err error) {
	_, err = CtrlDeviceStatusReport.Print()
	if err != nil {
		return 0, 0, err
	}

	reader := bufio.NewReader(os.Stdin)
	responseChan := make(chan string, 1)
	go func() {
		response, err := reader.ReadString('R')
		if err == nil {
			responseChan <- response
		}
	}()

	select {
	case <-time.After(75 * time.Millisecond):
		// If the timeout has expired, return an error
		return 0, 0, fmt.Errorf("timeout while waiting for device status report")
	case response := <-responseChan:
		response = strings.TrimSuffix(strings.TrimPrefix(response, csiStart), "R")
		rc := strings.Split(response, ";")
		if len(rc) != 2 {
			return 0, 0, fmt.Errorf("bad response format")
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
}
