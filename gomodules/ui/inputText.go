/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"fmt"
	"strings"
)

type InputTextModel struct {
	msg       string
	value     string
	validator func(string) error
	errorMsg  string
	inline    bool
	prompt    string
	uiApi     *ComponentApi
}

func (m *InputTextModel) GetComponentApi() *ComponentApi {
	return m.uiApi
}

func (m *InputTextModel) Inline() *InputTextModel {
	m.inline = true
	return m
}

// Allow to set a function to validate the input
// if the function return an error the input will be rejected and the error will be displayed
// the user will then be asked to enter a new value
func (m *InputTextModel) WithValidator(validator func(string) error) *InputTextModel {
	m.validator = validator
	return m
}

// Remove the input from the output when done
func (m *InputTextModel) WithCleanup() *InputTextModel {
	m.uiApi.Cleanup = true
	return m
}

// Set the prompt to display before the input (ignored in inline mode)
func (m *InputTextModel) SetPrompt(prompt string) *InputTextModel {
	m.prompt = prompt
	return m
}

func (m *InputTextModel) GetReadlineValue() string {
	return m.value
}
func (m *InputTextModel) Init() Cmd {
	return nil
}

func (m *InputTextModel) Update(msg Msg) Cmd {
	switch msg := msg.(type) {
	case MsgLine:
		m.value = msg.Value
		if m.validator != nil {
			if err := m.validator(msg.Value); err != nil {
				m.errorMsg = err.Error()
				return nil
			}
		}
		m.errorMsg = ""
		m.uiApi.Done = true
		return nil
	}
	return nil
}

func (m *InputTextModel) Render() string {
	theme := GetTheme()
	sb := strings.Builder{}
	if m.errorMsg != "" {
		sb.WriteString(theme.Error(m.errorMsg))
		sb.WriteRune('\n')
	}
	sb.WriteString(theme.Bold(theme.Accentuated(m.msg)))
	if !m.inline {
		sb.WriteString("\n")
	} else {
		sb.WriteString(" ")
	}
	sb.WriteString(theme.FocusItemIndicator())
	sb.WriteString(" ")
	// don't display back the value if it's a password
	if m.uiApi.InputReader != PasswordReader {
		sb.WriteString(m.value)
	}
	return sb.String()
}

func (m *InputTextModel) Fallback() Model {
	if m.errorMsg != "" {
		fmt.Printf("%s\n", m.errorMsg)
	}
	prompt := m.msg
	if m.inline {
		prompt += " "
	} else {
		prompt += "\n"
	}
	if m.prompt != "" {
		prompt += m.prompt
	} else if !m.inline {
		prompt += m.prompt
	}
	line, err := Readline(prompt)
	if err != nil {
		m.errorMsg = err.Error()
		return m.Fallback()
	}
	if m.validator != nil {
		if err := m.validator(line.Value); err != nil {
			m.errorMsg = err.Error()
			return m.Fallback()
		}
	}
	m.value = line.Value
	return m
}

func (m *InputTextModel) Run() string {
	return RunComponent(m).value
}

func NewInputText(msg string) *InputTextModel {
	theme := GetTheme()
	return &InputTextModel{
		msg:    msg,
		prompt: theme.FocusItemIndicator(),
		uiApi: &ComponentApi{
			InputReader: LineReader,
		},
	}
}

// Beware that it will work as an inputText in fallback mode
func NewInputPassword(msg string) *InputTextModel {
	theme := GetTheme()
	return &InputTextModel{
		msg:    msg,
		prompt: theme.FocusItemIndicator(),
		uiApi: &ComponentApi{
			InputReader: PasswordReader,
		},
	}
}
