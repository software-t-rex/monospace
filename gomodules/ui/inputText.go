/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var sgrExp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

type InputTextModel struct {
	msg             string
	value           string
	validator       func(string) error
	errorMsg        string
	inline          bool
	prompt          string
	compSuggester   func(string, string) ([]string, error)
	keyHandler      func(string) (Msg, error)
	uiApi           *ComponentApi
	readlineOptions LineEditorOptions
	visualPrint     string
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

// Allow to set a function to suggest completion for the input
// (only used in enhanced mode)
func (m *InputTextModel) WithCompletion(compSuggester func(string, string) ([]string, error)) *InputTextModel {
	m.compSuggester = compSuggester
	return m
}

// the keyHandler is called when a key is pressed in enhanced mode
// it should return a message to update the model or an error
// if nil,nil is returned then the normal input handling will be done
// if a non nil msg or a non nil error is returned then the input will be considered handled
// and enhanced readline will simply ignore the key
// (only used in enhanced mode)
func (m *InputTextModel) WithKeyHandler(keyHandler func(string) (Msg, error)) *InputTextModel {
	m.keyHandler = keyHandler
	return m
}

// Remove the input from the output when done
func (m *InputTextModel) WithCleanup() *InputTextModel {
	m.uiApi.Cleanup = true
	return m
}

func (m *InputTextModel) SetMaxLen(n int) *InputTextModel {
	m.readlineOptions.MaxLen = n
	return m
}

func (m *InputTextModel) SetMaxWidth(n int) *InputTextModel {
	m.readlineOptions.MaxWidth = n
	return m
}

func (m *InputTextModel) SetInitValue(value string) *InputTextModel {
	m.value = value
	return m
}

// Set the prompt to display before the input (ignored in inline mode)
func (m *InputTextModel) SetPrompt(prompt string) *InputTextModel {
	m.prompt = prompt
	return m
}

func (m *InputTextModel) Init() Cmd {
	m.readlineOptions.Value = m.value
	if !m.inline {
		m.readlineOptions.VisualPos = [2]int{0, len(sgrExp.ReplaceAllString(m.prompt, ""))}
	}
	return nil
}

func (m *InputTextModel) Update(msg Msg) Cmd {
	switch msg := msg.(type) {
	case MsgLine:
		m.value = msg.Value()
		if m.validator != nil {
			if err := m.validator(msg.Value()); err != nil {
				m.errorMsg = err.Error()
				return nil
			}
		}
		m.errorMsg = ""
		m.uiApi.Done = true
		m.visualPrint, _ = msg.Sprint() // ignore errors
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
	sb.WriteString(m.prompt)

	// displaying the value is handled by the input reader unless we are done
	if m.uiApi.Done {
		sb.WriteString(m.visualPrint)
	}
	return sb.String()
}

func (m *InputTextModel) Fallback() (Model, error) {
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
		if errors.Is(err, ErrSIGINT) || errors.Is(err, io.EOF) {
			return m, err
		}
		m.errorMsg = err.Error()
		return m.Fallback()
	}
	if m.validator != nil {
		if err := m.validator(line.Value()); err != nil {
			m.errorMsg = err.Error()
			return m.Fallback()
		}
	}
	m.value = line.Value()
	return m, nil
}

//#region ReadlineProviders interface implementation

func (m *InputTextModel) ReadlineConfig() LineEditorOptions {
	if m.uiApi.InputReader != PasswordReader {
		// don't provide current value to password input
		m.readlineOptions.Value = m.value
	}
	return m.readlineOptions
}

// ReadlineCompletion is called by the ReadlineEnhanced function to get completion suggestions
//   - wordStart is the start of the word to complete (start of the word to cursor position)
//   - word is the whole word under cursor (start to end of the word)
func (m *InputTextModel) ReadlineCompletion(wordStart string, word string) ([]string, error) {
	if m.compSuggester != nil {
		return m.compSuggester(wordStart, word)
	}
	return nil, nil
}

func (m *InputTextModel) ReadlineKeyHandler(key string) (Msg, error) {
	if m.keyHandler != nil {
		return m.keyHandler(key)
	}
	return nil, nil
}

//#endregion ReadlineProviders interface implementation

func (m *InputTextModel) Run() (string, error) {
	model, err := RunComponent(m)
	return model.value, err
}

func newInputText(msg string, reader inputReader) *InputTextModel {
	theme := GetTheme()
	return &InputTextModel{
		msg:             msg,
		prompt:          theme.FocusItemIndicator() + " ",
		readlineOptions: LineEditorOptions{},
		uiApi: &ComponentApi{
			InputReader: reader,
		},
	}
}

func NewInputText(msg string) *InputTextModel {
	return newInputText(msg, LineReader)
}

// Beware that it will work as an inputText in fallback mode
func NewInputPassword(msg string) *InputTextModel {
	return newInputText(msg, PasswordReader)
}
