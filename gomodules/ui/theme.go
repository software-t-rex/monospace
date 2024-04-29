/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"strings"
)

type (
	ThemeConfig struct {
		AccentColor ColorInterface
		/** color to use for text on top of accent color */
		AccentText   ColorInterface
		ErrorColor   ColorInterface
		ErrorText    ColorInterface
		SuccessColor ColorInterface
		SuccessText  ColorInterface
		WarningColor ColorInterface
		WarningText  ColorInterface
		InfoColor    ColorInterface
		InfoText     ColorInterface

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

		ButtonStyler            func(...string) string
		ButtonAccentuatedStyler func(...string) string
	}

	themePrebuiltRenderers struct {
		accentuated        func(...string) string
		reverseAccentuated func(...string) string
		error              func(...string) string
		emphasedError      func(...string) string
		success            func(...string) string
		emphasedSuccess    func(...string) string
		warning            func(...string) string
		emphasedWarning    func(...string) string
		info               func(...string) string
		emphasedInfo       func(...string) string
		title              func(...string) string
	}

	Theme struct {
		isDefined bool
		Config    ThemeConfig
		renderers themePrebuiltRenderers
	}

	ThemeInitializer func() ThemeConfig
)

// keep a global theme to be able to use it in the whole application
var usedTheme *Theme

// Set the global theme to use in the whole application
// if theme is nil it will use the default theme
func SetTheme(theme ThemeInitializer) *Theme {
	usedTheme = NewTheme(theme)
	return usedTheme
}

// Get the global theme to use in the whole application
// if theme is not set it will set theme to the default theme
func GetTheme() *Theme {
	if !usedTheme.isDefined {
		SetTheme(nil)
	}
	return usedTheme
}

func (t *ThemeConfig) Copy() ThemeConfig {
	return *t
}

var Error = NewStyler(Red.Foreground())
var EmphasedError = NewStyler(Red.Background(), White.Foreground(), Bold)

var Success = NewStyler(Green.Foreground())
var EmphasedSuccess = NewStyler(Green.Background(), Bold)

var Warning = NewStyler(Yellow.Foreground())
var EmphasedWarning = NewStyler(Yellow.Background(), Black.Foreground(), Bold)

var Info = NewStyler(Blue.Foreground())
var EmphasedInfo = NewStyler(Blue.Background(), White.Foreground(), Bold)

func NewTheme(themeInitializer ThemeInitializer) *Theme {
	var config ThemeConfig
	if themeInitializer == nil {
		config = ThemeDefault()
	} else {
		config = themeInitializer()
	}
	return &Theme{
		isDefined: true,
		Config:    config,
		renderers: themePrebuiltRenderers{
			accentuated:        NewStyler(config.AccentColor.Foreground()),
			reverseAccentuated: NewStyler(config.AccentColor.Background(), config.AccentText.Foreground()),
			error:              NewStyler(config.ErrorColor.Foreground()),
			emphasedError:      NewStyler(config.ErrorColor.Background(), config.ErrorText.Foreground(), Bold),
			success:            NewStyler(config.SuccessColor.Foreground()),
			emphasedSuccess:    NewStyler(config.SuccessColor.Background(), config.SuccessText.Foreground(), Bold),
			warning:            NewStyler(config.WarningColor.Foreground()),
			emphasedWarning:    NewStyler(config.WarningColor.Background(), config.WarningText.Foreground(), Bold),
			info:               NewStyler(config.InfoColor.Foreground()),
			emphasedInfo:       NewStyler(config.InfoColor.Background(), config.InfoText.Foreground(), Bold),
			title:              NewStyler(config.AccentColor.Foreground(), Bold),
		},
	}
	return t
}

// @todo Check copy really copy for renderers
func (t *Theme) Copy() Theme { return *t }

// beware that Bold also reset the Faint style
func (t *Theme) Bold(s ...string) string {
	return ApplyStyle(strings.Join(s, " "), Bold)
}

// beware that Faint also reset the Bold style
func (t *Theme) Faint(s ...string) string {
	return ApplyStyle(strings.Join(s, " "), Faint)
}

func (t *Theme) Italic(s ...string) string {
	return ApplyStyle(strings.Join(s, " "), Italic)
}

func (t *Theme) Underline(s ...string) string {
	return ApplyStyle(strings.Join(s, " "), Underline)
}

func (t *Theme) Blink(s ...string) string {
	return ApplyStyle(strings.Join(s, " "), Blink)
}

func (t *Theme) Reversed(s ...string) string {
	return ApplyStyle(strings.Join(s, " "), Reversed)
}

func (t *Theme) Strike(s ...string) string {
	return ApplyStyle(strings.Join(s, " "), Strike)
}

func (t *Theme) Accentuated(s ...string) string        { return t.renderers.accentuated(s...) }
func (t *Theme) ReverseAccentuated(s ...string) string { return t.renderers.reverseAccentuated(s...) }
func (t *Theme) Error(s ...string) string              { return t.renderers.error(s...) }
func (t *Theme) EmphasedError(s ...string) string      { return t.renderers.emphasedError(s...) }
func (t *Theme) Success(s ...string) string            { return t.renderers.success(s...) }
func (t *Theme) EmphasedSuccess(s ...string) string    { return t.renderers.emphasedSuccess(s...) }
func (t *Theme) Warning(s ...string) string            { return t.renderers.warning(s...) }
func (t *Theme) EmphasedWarning(s ...string) string    { return t.renderers.emphasedWarning(s...) }
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
	return t.Config.ButtonStyler(str)
}
func (t *Theme) ButtonAccentuated(s ...string) string {
	str := " " + strings.Join(s, " ") + " "
	return t.Config.ButtonAccentuatedStyler(str)
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
