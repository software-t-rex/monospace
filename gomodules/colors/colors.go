package colors

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

type style string

const (
	Reset style = "\033[0m"

	Black   style = "\033[30m"
	Red     style = "\033[31m"
	Green   style = "\033[32m"
	Yellow  style = "\033[33m"
	Blue    style = "\033[34m"
	Magenta style = "\033[35m"
	Cyan    style = "\033[36m"
	White   style = "\033[37m"
	Default style = "\033[39m"

	BrightBlack   style = "\033[30;1m"
	BrightRed     style = "\033[31;1m"
	BrightGreen   style = "\033[32;1m"
	BrightYellow  style = "\033[33;1m"
	BrightBlue    style = "\033[34;1m"
	BrightMagenta style = "\033[35;1m"
	BrightCyan    style = "\033[36;1m"
	BrightWhite   style = "\033[37;1m"

	Bold      style = "\033[1m"
	Italic    style = "\033[3m"
	Underline style = "\033[4m"
	Blink     style = "\033[5m"
	Reversed  style = "\033[7m"
	Strike    style = "\033[9m"

	ResetBold      style = "\033[22m" // 21m is not specified this is not a mistake
	ResetItalic    style = "\033[23m"
	ResetUnderline style = "\033[24m"
	ResetBlink     style = "\033[25m"
	ResetReversed  style = "\033[27m"
	ResetStrike    style = "\033[29m"

	BgBlack   style = "\033[40m"
	BgRed     style = "\033[41m"
	BgGreen   style = "\033[42m"
	BgYellow  style = "\033[43m"
	BgBlue    style = "\033[44m"
	BgMagenta style = "\033[45m"
	BgCyan    style = "\033[46m"
	BgWhite   style = "\033[47m"
	BgDefault style = "\033[49m"

	BrightBgBlack   style = "\033[40;1m"
	BrightBgRed     style = "\033[41;1m"
	BrightBgGreen   style = "\033[42;1m"
	BrightBgYellow  style = "\033[43;1m"
	BrightBgBlue    style = "\033[44;1m"
	BrightBgMagenta style = "\033[45;1m"
	BrightBgCyan    style = "\033[46;1m"
	BrightBgWhite   style = "\033[47;1m"
)

// var resets map[string]string = map[string]string{
// 	"\033[1m": "\033[22m",
// 	"\033[3m": "\033[23m",
// 	"\033[4m": "\033[24m",
// 	"\033[5m": "\033[25m",
// 	"\033[7m": "\033[27m",
// 	"\033[9m": "\033[29m",
// }

var enabled = true
var canColor = true

// Allow to turn off colouring for all Style methods
// Be careful: if you do string(colors.Red) + "a red string" + string(Reset)
// it will still be rendered in red as you use colors codes directly.
func Toggle(enable bool) {
	enabled = enable
}
func ColorEnabled() bool {
	return enabled && canColor
}

func init() {
	if runtime.GOOS == "windows" {
		canColor = false
	}
	nocolor := os.Getenv("NO_COLOR")
	if nocolor != "" && nocolor != "0" && nocolor != "false" {
		canColor = false
	}
}

// Returns a function that will apply given styles to the received String
// The returned function will return an unstyled string if colors.Toggle(false)
// have been called (or set to false on init by the env var NO_COLOR)
// Sample usage: colors.Style(colors.Red, colors.Bold)("This will be red and bold.")
func Style(styles ...style) func(s ...string) string {
	if !canColor {
		return func(s ...string) string { return strings.Join(s, " ") }
	}
	styleString := ""
	for _, style := range styles {
		styleString = styleString + string(style)
	}
	return func(s ...string) string {
		if enabled {
			return styleString + strings.Join(s, " ") + string(Reset)
		}
		return strings.Join(s, " ")
	}
}

var Error = Style(Red)
var EmphasedError = Style(BrightBgRed, White, Bold)

var Success = Style(Green)
var EmphasedSuccess = Style(BrightGreen, Bold)

var Warning = Style(Yellow)
var EmphasedWarning = Style(BrightBgYellow, Black, Bold)

var Info = Style(Blue)
var EmphasedInfo = Style(BrightBgBlue, White, Bold)

func Println(s string, styles ...style) {
	fmt.Println(Style(styles...)(s))
}
