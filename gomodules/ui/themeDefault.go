/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

func ThemeDefault() ThemeConfig {
	return ThemeConfig{
		AccentColor:  Magenta,
		AccentText:   White,
		ErrorColor:   Red,
		ErrorText:    White,
		SuccessColor: Green,
		SuccessText:  Black,
		WarningColor: Yellow,
		WarningText:  Black,
		InfoColor:    Blue,
		InfoText:     White,

		SelectedIndicatorString:    ApplyStyle("✔", Green.Foreground(), Bold),
		UnSelectedIndicatorString:  "∙",
		FocusItemIndicatorString:   ApplyStyle(">", Magenta.Foreground(), Bold),
		UnFocusItemIndicatorString: " ",
		MoreUpIndicatorString:      ApplyStyle("▴", Magenta.Foreground(), Faint),
		MoreDownIndicatorString:    ApplyStyle("▾", Magenta.Foreground(), Faint),
		KeySeparator:               "/",
		KeyBindingSeparator:        "•",

		ButtonStyler: NewStyler(
			AdaptiveColor{Light: "251", Dark: "236"}.Background(),
			AdaptiveColor{Light: "233", Dark: "254"}.Foreground(),
		),
		ButtonAccentuatedStyler: NewStyler(Magenta.Background(), Magenta.Foreground()),
	}
}

func ThemeMonoSpace() ThemeConfig {
	var (
		monoAccentColor  = AdaptiveColor{Light: "#f2804a", Dark: "#f2804a"}
		monoAccentText   = AdaptiveColor{Light: "0", Dark: "0"}
		monoSuccessColor = AdaptiveColor{Light: "2", Dark: "10"}
		monoMoreColor    = AdaptiveColor{Light: "245", Dark: "245"}
	)

	return ThemeConfig{
		AccentColor:  monoAccentColor,
		AccentText:   monoAccentText,
		ErrorColor:   AdaptiveColor{Light: "9", Dark: "9"},
		ErrorText:    AdaptiveColor{Light: "15", Dark: "15"},
		SuccessColor: monoSuccessColor,
		SuccessText:  AdaptiveColor{Light: "233", Dark: "233"},
		WarningColor: AdaptiveColor{Light: "11", Dark: "11"},
		WarningText:  AdaptiveColor{Light: "233", Dark: "233"},
		InfoColor:    AdaptiveColor{Light: "12", Dark: "12"},
		InfoText:     AdaptiveColor{Light: "233", Dark: "233"},

		SelectedIndicatorString:    ApplyStyle("✔", monoSuccessColor.Foreground(), Bold),
		UnSelectedIndicatorString:  "∙",
		FocusItemIndicatorString:   ApplyStyle(">", monoAccentColor.Foreground(), Bold),
		UnFocusItemIndicatorString: " ",
		MoreUpIndicatorString:      ApplyStyle("▴", monoMoreColor.Foreground()),
		MoreDownIndicatorString:    ApplyStyle("▾", monoMoreColor.Foreground()),
		KeySeparator:               "/",
		KeyBindingSeparator:        "•",

		ButtonStyler: NewStyler(
			AdaptiveColor{Light: "251", Dark: "236"}.Background(),
			AdaptiveColor{Light: "233", Dark: "254"}.Foreground(),
		),
		ButtonAccentuatedStyler: NewStyler(monoAccentColor.Background(), monoAccentText.Foreground()),
	}
}
