package colors

import (
	"fmt"
	"os"
	"runtime"
)

const (
	Reset = "\033[0m"

	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Purple  = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Default = "\033[39m"

	BrightBlack   = "\033[30;1m"
	BrightRed     = "\033[31;1m"
	BrightGreen   = "\033[32;1m"
	BrightYellow  = "\033[33;1m"
	BrightBlue    = "\033[34;1m"
	BrightMagenta = "\033[35;1m"
	BrightCyan    = "\033[36;1m"
	BrightWhite   = "\033[37;1m"

	Bold      = "\033[1m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	Reversed  = "\033[7m"
	Strike    = "\033[9m"

	ResetBold      = "\033[22m" // 21m is not specified
	ResetItalic    = "\033[23m"
	ResetUnderline = "\033[24m"
	ResetBlink     = "\033[25m"
	ResetReversed  = "\033[27m"
	ResetStrike    = "\033[29m"

	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgPurple  = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
	BgDefault = "\033[49m"

	BrightBgBlack  = "\033[40;1m"
	BrightBgRed    = "\033[41;1m"
	BrightBgGreen  = "\033[42;1m"
	BrightBgYellow = "\033[43;1m"
	BrightBgBlue   = "\033[44;1m"
	BrightBgPurple = "\033[45;1m"
	BrightBgCyan   = "\033[46;1m"
	BrightBgWhite  = "\033[47;1m"
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

func Toggle(enable bool) {
	enabled = enable
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

func Style(styles ...string) func(s string) string {
	if !canColor {
		return func(s string) string { return s }
	}
	styleString := ""
	for _, style := range styles {
		styleString = styleString + style
	}
	return func(s string) string {
		if enabled {
			return styleString + s + Reset
		}
		return s
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

func Println(s string, styles ...string) {
	fmt.Println(Style(styles...)(s))
}
