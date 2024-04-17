/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/
package ui

import "github.com/charmbracelet/lipgloss"

var (
	accentColor  = lipgloss.AdaptiveColor{Light: "#f2804a", Dark: "#f2804a"}
	accentText   = lipgloss.AdaptiveColor{Light: "0", Dark: "0"}
	errorColor   = lipgloss.AdaptiveColor{Light: "9", Dark: "9"}
	errorText    = lipgloss.AdaptiveColor{Light: "15", Dark: "15"}
	successColor = lipgloss.AdaptiveColor{Light: "2", Dark: "10"}
	successText  = lipgloss.AdaptiveColor{Light: "233", Dark: "233"}
	warningColor = lipgloss.AdaptiveColor{Light: "11", Dark: "11"}
	warningText  = lipgloss.AdaptiveColor{Light: "233", Dark: "233"}
	infoColor    = lipgloss.AdaptiveColor{Light: "12", Dark: "12"}
	infoText     = lipgloss.AdaptiveColor{Light: "233", Dark: "233"}
	moreColor    = lipgloss.AdaptiveColor{Light: "245", Dark: "245"}

	buttonStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "251", Dark: "236"}).
			Foreground(lipgloss.AdaptiveColor{Light: "233", Dark: "254"})

	buttonAccentuatedStyle = buttonStyle.Copy().Background(accentColor).Foreground(accentText)
)

var ThemeDefaultConfig = ThemeConfig{
	AccentColor:  accentColor,
	AccentText:   accentText,
	ErrorColor:   errorColor,
	ErrorText:    errorText,
	SuccessColor: successColor,
	SuccessText:  successText,
	WarningColor: warningColor,
	WarningText:  warningText,
	InfoColor:    infoColor,
	InfoText:     infoText,

	SelectedIndicatorString:    lipgloss.NewStyle().Foreground(successColor).Bold(true).Render("✔"),
	UnSelectedIndicatorString:  "∙",
	FocusItemIndicatorString:   lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(">"),
	UnFocusItemIndicatorString: " ",
	MoreUpIndicatorString:      lipgloss.NewStyle().Foreground(moreColor).Render("▴"),
	MoreDownIndicatorString:    lipgloss.NewStyle().Foreground(moreColor).Render("▾"),
	KeySeparator:               "/",
	KeyBindingSeparator:        "•",

	ButtonStyle:            buttonStyle,
	ButtonAccentuatedStyle: buttonAccentuatedStyle,
}

var ThemeDefault = NewTheme(ThemeDefaultConfig)
