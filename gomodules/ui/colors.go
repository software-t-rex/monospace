/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	Black        colorANSI8 = 0
	Red          colorANSI8 = 1
	Green        colorANSI8 = 2
	Yellow       colorANSI8 = 3
	Blue         colorANSI8 = 4
	Magenta      colorANSI8 = 5
	Cyan         colorANSI8 = 6
	White        colorANSI8 = 7
	ResetColor   colorANSI8 = 9
	DefaultColor colorANSI8 = ResetColor
)

type RGB [3]int
type ColorInterface interface {
	Foreground() SGRParam
	Background() SGRParam
}

type Color string

func (c Color) Foreground() SGRParam {
	if strings.HasPrefix(string(c), "#") {
		return ForegroundHex(string(c))
	}
	i, err := strconv.Atoi(string(c))
	if err != nil || i < 0 {
		return ""
	}
	if i < 8 {
		return colorANSI8(i).Foreground()
	}
	if i < 256 {
		return colorANSI256(i).Foreground()
	}
	return ""
}
func (c Color) Background() SGRParam {
	if strings.HasPrefix(string(c), "#") {
		return BackgroundHex(string(c))
	}
	i, err := strconv.Atoi(string(c))
	if err != nil || i < 0 {
		return ""
	}
	if i < 8 {
		return colorANSI8(i).Background()
	}
	if i < 256 {
		return colorANSI256(i).Background()
	}
	return ""
}

// colorANSI8 represents an ANSI color from 0 to 7
type colorANSI8 int

func (c colorANSI8) Foreground() SGRParam {
	return SGRParam(fmt.Sprintf("3%d", c))
}
func (c colorANSI8) Background() SGRParam {
	return SGRParam(fmt.Sprintf("4%d", c))
}

type colorANSI256 int

func (c colorANSI256) Foreground() SGRParam {
	return ForegroundANSI256(int(c))
}
func (c colorANSI256) Background() SGRParam {
	return BackgroundANSI256(int(c))
}

type AdaptiveColor struct {
	Light Color
	Dark  Color
}

func (c AdaptiveColor) Color() Color {
	var terminal TermWithBackground = GetTerminal()
	if !EnhancedEnabled() {
		return c.Dark
	} else if isDark, _ := terminal.HasDarkBackground(); !isDark {
		return c.Light
	}
	return c.Dark
}
func (c AdaptiveColor) Foreground() SGRParam {
	return c.Color().Foreground()
}
func (c AdaptiveColor) Background() SGRParam {
	return c.Color().Background()
}

func ForegroundANSI256(c int) SGRParam {
	return SGRParam(fmt.Sprintf("38;5;%d", c))
}

func BackgroundANSI256(c int) SGRParam {
	return SGRParam(fmt.Sprintf("48;5;%d", c))
}

func ForegroundRGB(rgb RGB) SGRParam {
	return SGRParam(fmt.Sprintf("38;2;%d;%d;%d", rgb[0], rgb[1], rgb[2]))
}

func BackgroundRGB(rgb RGB) SGRParam {
	return SGRParam(fmt.Sprintf("48;2;%d;%d;%d", rgb[0], rgb[1], rgb[2]))
}

func ForegroundHex(hex string) SGRParam {
	rgb, err := hexToRGB(hex)
	if err != nil {
		return ""
	}
	return ForegroundRGB(rgb)
}

func BackgroundHex(hex string) SGRParam {
	rgb, err := hexToRGB(hex)
	if err != nil {
		return ""
	}
	return BackgroundRGB(rgb)
}

var validHexColor = regexp.MustCompile(`^#([0-9a-fA-F]{6}|[0-9a-fA-F]{3})$`)

// only read rgb in the form of "#RRGGBB" / "#RGB"
func hexToRGB(hex string) (rgb RGB, err error) {
	// check if the hex string is valid with a regexp
	if !validHexColor.MatchString(hex) {
		return rgb, fmt.Errorf("invalid hex color: %s", hex)
	}
	if len(hex) == 4 {
		hex = fmt.Sprintf("#%c%c%c%c%c%c", hex[1], hex[1], hex[2], hex[2], hex[3], hex[3])
	}
	_, err = fmt.Sscanf(hex, "#%02x%02x%02x", &rgb[0], &rgb[1], &rgb[2])
	if err != nil {
		return rgb, err // should never happen as we already validated the hex string
	}
	return rgb, nil
}
