/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package tasks

import (
	"fmt"
	"strings"
	"unicode"
)

var ErrParseError = fmt.Errorf("parse error")
var ErrParseQuoteError = fmt.Errorf("unmatched quote %w", ErrParseError)

type argsParser struct {
	Args       []string
	CmdStr     string
	len        int
	position   int
	strictMode bool
}

func NewArgsParser() *argsParser {
	return &argsParser{Args: []string{}}
}

func (p *argsParser) StrictMode(shouldFail bool) *argsParser {
	p.strictMode = shouldFail
	return p
}

func (p *argsParser) initParsing(CmdStr string) *argsParser {
	p.Args = []string{}
	p.CmdStr = strings.TrimSpace(CmdStr)
	p.len = len(p.CmdStr)
	p.position = 0
	return p
}

func (p *argsParser) parse() ([]string, error) {
	var err error
	for p.position < p.len {
		err = p.parseNext()
		if err != nil {
			return p.Args, err
		}
	}
	return p.Args, err
}

// Parse a string into arguments suitable for exec.Command arguments
func (p *argsParser) Parse(CmdStr string) ([]string, error) {
	args, err := p.StrictMode(false).initParsing(CmdStr).parse()
	if err != nil && err == ErrParseQuoteError {
		return args, nil
	}
	return args, err
}

func (p *argsParser) ParseStrict(CmdStr string) ([]string, error) {
	return p.StrictMode(true).initParsing(CmdStr).parse()
}

func (p *argsParser) parseNext() error {
	// skip spaces
	for p.position < p.len && unicode.IsSpace(rune(p.CmdStr[p.position])) {
		p.position++
	}
	// end of string
	if p.position >= p.len {
		return nil
	}
	// check for quoted argument
	if p.CmdStr[p.position] == '"' || p.CmdStr[p.position] == '\'' {
		arg, err := p.parseQuoted()
		p.Args = append(p.Args, arg)
		return err
	}
	// check for unquoted argument
	arg, err := p.parseUnquoted()
	p.Args = append(p.Args, arg)
	return err
}

func (p *argsParser) parseQuoted() (string, error) {
	quote := rune(p.CmdStr[p.position])
	p.position++
	start := p.position
	for p.position < p.len && (rune(p.CmdStr[p.position]) != quote || p.CmdStr[p.position-1] == '\\') {
		p.position++
	}
	arg := p.CmdStr[start:p.position]
	// unescape potential escaped quotes
	arg = strings.ReplaceAll(arg, fmt.Sprintf("\\%c", quote), string(quote))
	p.position++
	if p.position > p.len && p.strictMode {
		return arg, ErrParseQuoteError
	}
	return arg, nil
}
func (p *argsParser) parseUnquoted() (string, error) {
	start := p.position
	// parse until next space unless we found a quote
	for p.position < p.len {
		c := rune(p.CmdStr[p.position])
		if unicode.IsSpace(c) {
			break
		} else if c == '"' || c == '\'' {
			err := p.parseEmbeddedQuoted()
			if err != nil {
				return p.CmdStr[start:], err
			}
		} else {
			p.position++
		}
	}
	if p.position > p.len {
		if p.strictMode {
			return p.CmdStr[start:], ErrParseError
		} else {
			return p.CmdStr[start:], nil
		}
	}
	return p.CmdStr[start:p.position], nil
}

// parseEmbeddedQuoted parses a quoted string, handling escaped quotes.
func (p *argsParser) parseEmbeddedQuoted() error {
	quote := rune(p.CmdStr[p.position])
	p.position++
	for p.position < p.len && (rune(p.CmdStr[p.position]) != quote || p.CmdStr[p.position-1] == '\\') {
		p.position++
	}
	if p.position >= p.len {
		if p.strictMode {
			return ErrParseQuoteError
		}
		return nil
	}
	p.position++
	return nil
}

// ParseArgs parse a string into arguments suitable for exec.Command arguments
// If the command name is part of the string then the first argument will be the command name
// leading and trailing spaces are ignored.
// Errors during parsing are ignored.
func ParseArgs(str string) []string {
	args, _ := NewArgsParser().Parse(str)
	return args
}

// This is the same as ParseArgs but will return an error if an unmatched quote is found
func ParseArgsStrict(str string) ([]string, error) {
	return NewArgsParser().ParseStrict(str)
}
