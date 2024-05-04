package main

import (
	"fmt"
	"regexp"

	"github.com/software-t-rex/monospace/gomodules/ui"
)

func main() {
	theme := ui.SetTheme(ui.ThemeMonoSpace)
	res := ui.NewInputText("Enter your name").
		// WithCleanup().
		WithValidator(func(str string) error {
			if str == "" {
				return fmt.Errorf("name can't be empty")
			}
			if len(str) < 3 {
				return fmt.Errorf("name must be at least 3 characters long")
			}
			return nil
		}).
		Run()
	ui.Println(fmt.Sprintf("Hello %s!", theme.Bold(res)), ui.BrightBlue.Foreground())

	pass := ui.NewInputPassword("Enter your password").
		Inline().
		WithValidator(func(str string) error {
			if str == "" {
				return fmt.Errorf("password can't be empty")
			}
			if len(str) < 8 {
				return fmt.Errorf("password must be at least 8 characters long")
			}
			tests := []string{".{7,}", "[a-z]", "[A-Z]", "[0-9]", "[^\\d\\w]"}
			for _, test := range tests {
				t, _ := regexp.MatchString(test, str)
				if !t {
					return fmt.Errorf("password must contains uppercase and lowercase letters, digits and special characters")
				}
			}
			return nil
		}).
		Run()
	ui.Println(fmt.Sprintf("Your password is '%s'", theme.Bold(pass)), ui.BrightBlue.Foreground())

}
