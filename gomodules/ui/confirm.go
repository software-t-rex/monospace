/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"fmt"
	"os"
	"strings"
)

type confirmModel struct {
	msg       string
	bindings  *KeyBindings[*confirmModel]
	inline    bool
	confirmed bool
	help      bool
	yesLabel  string
	noLabel   string
	errorMsg  string // used only in fallback mode
	uiApi     *ComponentApi
}

func (m *confirmModel) GetComponentApi() *ComponentApi {
	return m.uiApi
}

func (m *confirmModel) Inline() *confirmModel {
	m.inline = true
	return m
}

func (m *confirmModel) WithoutHelp() *confirmModel {
	m.help = false
	return m
}

// Remove the menu from the output when done
// default to false (menu will remain visible)
// Ignored in fallback mode.
func (m *confirmModel) WithCleanup(clear bool) *confirmModel {
	m.uiApi.Cleanup = clear
	return m
}

// this is ignored in fallback mode
func (m *confirmModel) SetYesLabel(label string) *confirmModel {
	m.yesLabel = label
	return m
}

// this is ignored in fallback mode
func (m *confirmModel) SetNoLabel(label string) *confirmModel {
	m.noLabel = label
	return m
}

func (m *confirmModel) Init() Cmd {
	m.bindings = NewKeyBindings[*confirmModel]().
		AddBinding("y", "Confirm", func(m *confirmModel) Cmd {
			m.confirmed = true
			m.uiApi.Done = true
			return CmdQuit
		}).
		AddBinding("n", "Cancel", func(m *confirmModel) Cmd {
			m.confirmed = false
			m.uiApi.Done = true
			return CmdQuit
		}).
		AddBinding("left,right,up,down,tab,h,j,k,l", "", func(m *confirmModel) Cmd {
			m.confirmed = !m.confirmed
			return nil
		}).
		AddToDescription("Arrow/tab to switch").
		AddBinding("enter", "Validate answer", func(m *confirmModel) Cmd {
			m.uiApi.Done = true
			return CmdQuit
		}).
		AddBinding("ctrl+c", "", func(m *confirmModel) Cmd {
			return CmdKill // this will exit the program
		})
	return nil
}

func (m *confirmModel) Update(msg Msg) Cmd {
	return m.bindings.Handle(m, msg)
}

func (m *confirmModel) Render() string {
	theme := GetTheme()
	var sb strings.Builder
	if m.uiApi.Done {
		if m.uiApi.Cleanup {
			return "" // we don't want to display anything if we are in clearScreen mode
		}
		sb.WriteString(theme.Title(m.msg))
		sb.WriteString(" ")
		if m.confirmed {
			sb.WriteString(theme.Success("Yes"))
		} else {
			sb.WriteString(theme.Error("No"))
		}
		return sb.String()
	}
	sb.WriteString(theme.Title(m.msg))
	if !m.inline {
		sb.WriteString("\n")
	}

	sb.WriteString(" ")
	if m.confirmed {
		sb.WriteString(theme.Button(m.noLabel) + " " + theme.ButtonSuccess(m.yesLabel))
	} else {
		sb.WriteString(theme.ButtonError(m.noLabel) + " " + theme.Button(m.yesLabel))
	}
	if !m.uiApi.Done && m.help {
		sb.WriteString("\n")
		sb.WriteString(m.bindings.GetDescription())
	}
	return sb.String()
}

func (m *confirmModel) Fallback() Model {
	sb := strings.Builder{}
	if m.errorMsg != "" {
		sb.WriteString(Msgs["errorPrefix"])
		sb.WriteString(m.errorMsg)
		sb.WriteString("\n")
	}
	dfltString := Msgs["fallbackConfirmPromptTrue"]
	if !m.help && !m.confirmed {
		dfltString = Msgs["fallbackConfirmPromptFalse"]
	} else if m.help && m.confirmed {
		dfltString = Msgs["fallbackConfirmHelpPromptTrue"]
	} else if m.help && !m.confirmed {
		dfltString = Msgs["fallbackConfirmHelpPromptFalse"]
	}
	sb.WriteString(m.msg)
	if m.inline {
		sb.WriteString(" ")
	} else {
		sb.WriteString("\n")
	}
	sb.WriteString(dfltString)

	input, err := Readline(sb.String())
	if err != nil {
		os.Exit(1)
	}
	switch strings.ToLower(input.Value) {
	case "y", "yes":
		m.confirmed = true
		m.uiApi.Done = true
	case "n", "no":
		m.confirmed = false
		m.uiApi.Done = true
	case "": // do nothing
		m.uiApi.Done = true
	default:
		m.errorMsg = Msgs["fallbackConfirmError"]
	}
	if !m.uiApi.Done {
		return m.Fallback()
	}
	if m.uiApi.Cleanup {
		fmt.Print("\033[2K\033[1A\033[2K")
	}
	return m
}

func NewConfirm(msg string, dflt bool) *confirmModel {
	theme := GetTheme()
	return &confirmModel{
		msg:       msg,
		confirmed: dflt,
		yesLabel:  theme.Underline("Y") + "es",
		noLabel:   theme.Underline("N") + "o",
		help:      true,
		uiApi:     &ComponentApi{},
	}
}

func (m *confirmModel) Run() bool {
	return RunComponent(m).confirmed
}

// Shorthand for NewConfirm(msg, dflt).Run()
func Confirm(msg string, dflt bool) bool {
	return NewConfirm(msg, dflt).Run()
}

func ConfirmInline(msg string, dflt bool) bool {
	return NewConfirm(msg, dflt).Inline().Run()
}
func ConfirmInlineNoHelp(msg string, dflt bool) bool {
	return NewConfirm(msg, dflt).Inline().WithoutHelp().Run()
}
