package utils

import (
	"fmt"
	"strings"

	"github.com/software-t-rex/monospace/colors"
)

var ErrorStyle = colors.Error
var EmphasedError = colors.EmphasedError
var Success = colors.Success
var EmphasedSuccess = colors.EmphasedSuccess
var Info = colors.Info
var EmphasedInfo = colors.EmphasedInfo
var Warning = colors.Warning
var EmphasedWarning = colors.EmphasedWarning

var Underline = colors.Style(colors.Underline)
var Bold = colors.Style(colors.Bold)
var Italic = colors.Style(colors.Italic)

var Red = colors.Style(colors.Red)
var Green = colors.Style(colors.Green)
var Blue = colors.Style(colors.Blue)
var Yellow = colors.Style(colors.Yellow)
var BrightBlue = colors.Style(colors.BrightBlue)

func Indent(s string, indentation string) string {
	return fmt.Sprintf("%s%s\n", indentation, strings.Replace(s, "\n", "\n"+indentation, -1))
}
