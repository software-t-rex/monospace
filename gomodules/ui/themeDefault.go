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

		successIndicatorString:     ApplyStyle("✔", Green.Foreground(), Bold),
		failureIndicatorString:     ApplyStyle("✘", Red.Foreground(), Bold),
		SelectedIndicatorString:    ApplyStyle("✔", Green.Foreground(), Bold),
		UnSelectedIndicatorString:  "∙",
		FocusItemIndicatorString:   ApplyStyle(">", Magenta.Foreground(), Bold),
		UnFocusItemIndicatorString: " ",
		MoreUpIndicatorString:      ApplyStyle("▴", Magenta.Foreground(), Faint),
		MoreDownIndicatorString:    ApplyStyle("▾", Magenta.Foreground(), Faint),
		KeySeparator:               "/",
		KeyBindingSeparator:        "•",

		ButtonStyler: NewStyler(
			AdaptiveColor{Light: colorANSI256(251), Dark: colorANSI256(236)}.Background(),
			AdaptiveColor{Light: colorANSI256(233), Dark: colorANSI256(254)}.Foreground(),
		),
		ButtonAccentuatedStyler: NewStyler(Magenta.Background(), Magenta.Foreground()),
	}
}

func ThemeMonoSpace() ThemeConfig {
	var (
		monoAccentColor  = Color("#f2804a")
		monoAccentText   = Black
		monoSuccessColor = AdaptiveColor{Light: Green, Dark: BrightGreen}
		monoErrorColor   = BrightRed
		monoMoreColor    = MidGrey
	)

	return ThemeConfig{
		AccentColor:  monoAccentColor,
		AccentText:   monoAccentText,
		ErrorColor:   monoErrorColor,
		ErrorText:    BrightWhite,
		SuccessColor: monoSuccessColor,
		SuccessText:  DarkGrey,
		WarningColor: BrightYellow,
		WarningText:  DarkGrey,
		InfoColor:    BrightBlue,
		InfoText:     DarkGrey,

		successIndicatorString:     ApplyStyle("✔", monoSuccessColor.Foreground(), Bold),
		failureIndicatorString:     ApplyStyle("✘", monoErrorColor.Foreground(), Bold),
		SelectedIndicatorString:    ApplyStyle("✔", monoSuccessColor.Foreground(), Bold),
		UnSelectedIndicatorString:  "∙",
		FocusItemIndicatorString:   ApplyStyle(">", monoAccentColor.Foreground(), Bold),
		UnFocusItemIndicatorString: " ",
		MoreUpIndicatorString:      ApplyStyle("▴", monoMoreColor.Foreground()),
		MoreDownIndicatorString:    ApplyStyle("▾", monoMoreColor.Foreground()),
		KeySeparator:               "/",
		KeyBindingSeparator:        "•",

		ButtonStyler: NewStyler(
			AdaptiveColor{Light: Gray20, Dark: Gray5}.Background(),
			AdaptiveColor{Light: DarkGrey, Dark: LightGrey}.Foreground(),
		),
		ButtonAccentuatedStyler: NewStyler(monoAccentColor.Background(), monoAccentText.Foreground()),
	}
}
