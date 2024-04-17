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

	"github.com/charmbracelet/lipgloss"
)

var usedTheme = ThemeDefault

func SetTheme(theme Theme) {
	usedTheme = theme
}

func GetTheme() Theme {
	return usedTheme
}

type ThemeConfig struct {
	AccentColor lipgloss.AdaptiveColor
	/** color to use for text on top of accent color */
	AccentText   lipgloss.AdaptiveColor
	ErrorColor   lipgloss.AdaptiveColor
	ErrorText    lipgloss.AdaptiveColor
	SuccessColor lipgloss.AdaptiveColor
	SuccessText  lipgloss.AdaptiveColor
	WarningColor lipgloss.AdaptiveColor
	WarningText  lipgloss.AdaptiveColor
	InfoColor    lipgloss.AdaptiveColor
	InfoText     lipgloss.AdaptiveColor

	SelectedIndicatorString    string
	UnSelectedIndicatorString  string
	FocusItemIndicatorString   string
	UnFocusItemIndicatorString string
	MoreUpIndicatorString      string
	MoreDownIndicatorString    string
	// key bindings separators should not be styled
	KeySeparator string
	// key bindings separators should not be styled
	KeyBindingSeparator string

	ButtonStyle            lipgloss.Style
	ButtonAccentuatedStyle lipgloss.Style
}

func (t *ThemeConfig) Copy() ThemeConfig {
	return *t
}

type themePrebuiltRenderers struct {
	accentuated        func(...string) string
	reverseAccentuated func(...string) string
	error              func(...string) string
	emphasedError      func(...string) string
	success            func(...string) string
	emphasedSuccess    func(...string) string
	warnings           func(...string) string
	emphasedWarnings   func(...string) string
	info               func(...string) string
	emphasedInfo       func(...string) string
	title              func(...string) string
}

type Theme struct {
	Config    ThemeConfig
	renderers themePrebuiltRenderers
}

func NewTheme(config ThemeConfig) Theme {
	success := lipgloss.NewStyle().Foreground(config.SuccessColor).Render
	t := Theme{
		Config: ThemeDefaultConfig,
		renderers: themePrebuiltRenderers{
			accentuated:        lipgloss.NewStyle().Foreground(config.AccentColor).Render,
			reverseAccentuated: lipgloss.NewStyle().Background(config.AccentColor).Foreground(config.AccentText).Render,
			error:              lipgloss.NewStyle().Foreground(config.ErrorColor).Render,
			emphasedError:      lipgloss.NewStyle().Background(config.ErrorColor).Foreground(config.ErrorText).Bold(true).Render,
			success:            success,
			emphasedSuccess:    lipgloss.NewStyle().Background(config.SuccessColor).Foreground(config.SuccessText).Bold(true).Render,
			warnings:           lipgloss.NewStyle().Foreground(config.WarningColor).Render,
			emphasedWarnings:   lipgloss.NewStyle().Background(config.WarningColor).Foreground(config.WarningText).Bold(true).Render,
			info:               lipgloss.NewStyle().Foreground(config.InfoColor).Render,
			emphasedInfo:       lipgloss.NewStyle().Background(config.InfoColor).Foreground(config.InfoText).Bold(true).Render,
			title:              lipgloss.NewStyle().Foreground(config.AccentColor).Bold(true).Render,
		},
	}
	return t
}

// @todo Check copy really copy for renderers
func (t *Theme) Copy() Theme { return *t }

// beware that Bold also reset the Faint style
func (t *Theme) Bold(s ...string) string {
	return fmt.Sprintf("\033[1m%s\033[22m", strings.Join(s, " "))
}

// beware that Faint also reset the Bold style
func (t *Theme) Faint(s ...string) string {
	return fmt.Sprintf("\033[2m%s\033[22m", strings.Join(s, " "))
}

func (t *Theme) Italic(s ...string) string {
	return fmt.Sprintf("\033[3m%s\033[23m", strings.Join(s, " "))
}

func (t *Theme) Underline(s ...string) string {
	return fmt.Sprintf("\033[4m%s\033[24m", strings.Join(s, " "))
}

func (t *Theme) Blink(s ...string) string {
	return fmt.Sprintf("\033[5m%s\033[25m", strings.Join(s, " "))
}

func (t *Theme) Reversed(s ...string) string {
	return fmt.Sprintf("\033[7m%s\033[27m", strings.Join(s, " "))
}

func (t *Theme) Strike(s ...string) string {
	return fmt.Sprintf("\033[9m%s\033[29m", strings.Join(s, " "))
}

func (t *Theme) Accentuated(s ...string) string        { return t.renderers.accentuated(s...) }
func (t *Theme) ReverseAccentuated(s ...string) string { return t.renderers.reverseAccentuated(s...) }
func (t *Theme) Error(s ...string) string              { return t.renderers.error(s...) }
func (t *Theme) EmphasedError(s ...string) string      { return t.renderers.emphasedError(s...) }
func (t *Theme) Success(s ...string) string            { return t.renderers.success(s...) }
func (t *Theme) EmphasedSuccess(s ...string) string    { return t.renderers.emphasedSuccess(s...) }
func (t *Theme) Warnings(s ...string) string           { return t.renderers.warnings(s...) }
func (t *Theme) EmphasedWarnings(s ...string) string   { return t.renderers.emphasedWarnings(s...) }
func (t *Theme) Info(s ...string) string               { return t.renderers.info(s...) }
func (t *Theme) EmphasedInfo(s ...string) string       { return t.renderers.emphasedInfo(s...) }

func (t *Theme) SelectedIndicator() string   { return t.Config.SelectedIndicatorString }
func (t *Theme) UnSelectedIndicator() string { return t.Config.UnSelectedIndicatorString }
func (t *Theme) ConditionalSelectedIndicator(selected bool) string {
	if selected {
		return t.SelectedIndicator()
	}
	return t.UnSelectedIndicator()
}
func (t *Theme) FocusItemIndicator() string   { return t.Config.FocusItemIndicatorString }
func (t *Theme) UnFocusItemIndicator() string { return t.Config.UnFocusItemIndicatorString }
func (t *Theme) ConditionalFocusIndicator(focus bool) string {
	if focus {
		return t.FocusItemIndicator()
	}
	return t.UnFocusItemIndicator()
}
func (t *Theme) MoreUpIndicator() string     { return t.Config.MoreUpIndicatorString }
func (t *Theme) MoreDownIndicator() string   { return t.Config.MoreDownIndicatorString }
func (t *Theme) KeySeparator() string        { return t.Config.KeySeparator }
func (t *Theme) KeyBindingSeparator() string { return t.Config.KeyBindingSeparator }

func (t *Theme) Button(s ...string) string {
	str := " " + strings.Join(s, " ") + " "
	return t.Config.ButtonStyle.Render(str)
}
func (t *Theme) ButtonAccentuated(s ...string) string {
	str := " " + strings.Join(s, " ") + " "
	return t.Config.ButtonAccentuatedStyle.Render(str)
}
func (t *Theme) ButtonSuccess(s ...string) string {
	str := " " + strings.Join(s, " ") + " "
	return t.renderers.emphasedSuccess(str)
}
func (t *Theme) ButtonError(s ...string) string {
	str := " " + strings.Join(s, " ") + " "
	return t.renderers.emphasedError(str)
}

func (t *Theme) Title(s ...string) string { return t.renderers.title(s...) }
