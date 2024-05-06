/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"fmt"
	"io"
	"log"
	"os"
)

type lineEditor struct {
	value         string
	cursorPos     int
	visualEdition bool
	out           io.Writer
	debugLogger   *log.Logger
}

func NewLineEditor(out io.Writer, visualEdition bool) *lineEditor {
	logFile, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	debugLogger := log.New(logFile, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	return &lineEditor{out: out, visualEdition: visualEdition, debugLogger: debugLogger}
}

func (l *lineEditor) GetValue() string { return l.value }
func (l *lineEditor) len() int         { return len(l.value) }
func (l *lineEditor) insert(s string) {
	remain := s + l.value[l.cursorPos:]
	l.value = l.value[:l.cursorPos] + remain
	l.cursorPos++
	if l.visualEdition && remain != "" {
		fmt.Fprintf(l.out, "%s", remain)
		if len(l.value) > l.cursorPos {
			l.visualCursorMoveLeft(len(l.value) - l.cursorPos)
		}
	}

}
func (l *lineEditor) delete() {
	if l.cursorPos < len(l.value) {
		remain := l.value[l.cursorPos+1:]
		l.value = l.value[:l.cursorPos] + remain
		if l.visualEdition {
			l.visualClearToEnd()
			fmt.Fprintf(l.out, "%s", remain)
			l.visualCursorMoveLeft(len(remain))
		}
	}
}
func (l *lineEditor) deleteBackward() {
	if l.cursorPos > 0 {
		remain := l.value[l.cursorPos:]
		if len(remain) > 0 {
			l.value = l.value[:l.cursorPos-1] + remain
			if l.visualEdition {
				l.visualCursorMoveLeft(1)
			}
			l.cursorPos--
			if l.visualEdition {
				l.visualClearToEnd()
				fmt.Fprintf(l.out, "%s", remain)
				l.visualCursorMoveLeft(len(remain))
			}
		} else {
			l.cursorPos--
			l.value = l.value[:l.cursorPos]
			if l.visualEdition {
				l.visualCursorMoveLeft(1)
				l.visualClearToEnd()
			}
		}
	}
}
func (l *lineEditor) deleteToStart() {
	if l.cursorPos > 0 {
		l.value = l.value[l.cursorPos:]
		if l.visualEdition {
			l.visualCursorMoveLeft(l.cursorPos)
			l.visualClearToEnd()
			fmt.Fprintf(l.out, "%s", l.value)
			l.visualCursorMoveLeft(len(l.value))
		}
		l.cursorPos = 0
	}
}
func (l *lineEditor) deleteToEnd() {
	if l.cursorPos < len(l.value) {
		l.value = l.value[:l.cursorPos]
		if l.visualEdition {
			l.visualClearToEnd()
		}
	}
}
func (l *lineEditor) deleteWord() {
	if l.cursorPos < len(l.value) {
		// find the end of the word
		end := l.findWordBoundary("right")
		l.value = l.value[:l.cursorPos] + l.value[end:]
		if l.visualEdition {
			l.visualClearToEnd()
			fmt.Fprintf(l.out, "%s", l.value[l.cursorPos:])
			l.visualCursorMoveLeft(len(l.value) - l.cursorPos)
		}
	}
}
func (l *lineEditor) deleteWordBackward() {
	if l.cursorPos > 0 {
		// find the start of the word
		start := l.findWordBoundary("left")
		l.value = l.value[:start] + l.value[l.cursorPos:]
		if l.visualEdition {
			l.visualCursorMoveLeft(l.cursorPos - start)
			l.visualClearToEnd()
			fmt.Fprintf(l.out, "%s", l.value[start:])
			l.visualCursorMoveLeft(len(l.value) - start)
		}
		l.cursorPos = start
	}
}
func (l *lineEditor) moveStart() {
	if l.cursorPos > 0 {
		if l.visualEdition {
			l.visualCursorMoveLeft(l.cursorPos)
		}
		l.cursorPos = 0
	}
}
func (l *lineEditor) moveEnd() {
	length := l.len()
	if l.cursorPos < length {
		if l.visualEdition {
			l.visualCursorMoveRight(length - l.cursorPos)
		}
		l.cursorPos = length
	}
}
func (l *lineEditor) moveLeft() {
	if l.cursorPos > 0 {
		l.cursorPos--
		if l.visualEdition {
			l.visualCursorMoveLeft(1)
		}
	}
}
func (l *lineEditor) moveRight() {
	if l.cursorPos < len(l.value) {
		l.cursorPos++
		if l.visualEdition {
			l.visualCursorMoveRight(1)
		}
	}
}
func (l *lineEditor) moveWordLeft() {
	if l.cursorPos > 0 {
		// find the start of the word
		start := l.findWordBoundary("left")
		if l.visualEdition {
			l.visualCursorMoveLeft(l.cursorPos - start)
		}
		l.cursorPos = start
	}
}
func (l *lineEditor) moveWordRight() {
	if l.cursorPos < len(l.value) {
		// find the end of the word
		end := l.findWordBoundary("right")
		if l.visualEdition {
			l.visualCursorMoveRight(end - l.cursorPos)
		}
		l.cursorPos = end
	}
}
func (l *lineEditor) clear() {
	l.value = ""
	if l.visualEdition {
		l.visualCursorMoveLeft(l.cursorPos)
		l.visualClearToEnd()
	}
	l.cursorPos = 0
}
func (l *lineEditor) visualClearToEnd() {
	fmt.Fprintf(l.out, "\033[K")
}
func (l *lineEditor) visualCursorMoveLeft(n int) {
	if n > 0 {
		fmt.Fprintf(l.out, "\033[%dD", n)
	}
}
func (l *lineEditor) visualCursorMoveRight(n int) {
	if n > 0 {
		fmt.Fprintf(l.out, "\033[%dC", n)
	}
}

func (l *lineEditor) findWordBoundary(direction string) int {
	pos := l.cursorPos
	if direction == "left" {
		for pos > 0 && l.value[pos-1] == ' ' {
			pos--
		}
		for pos > 0 && l.value[pos-1] != ' ' {
			pos--
		}
	} else if direction == "right" {
		for pos < l.len() && l.value[pos] == ' ' {
			pos++
		}
		for pos < l.len() && l.value[pos] != ' ' {
			pos++
		}
	}
	return pos
}
