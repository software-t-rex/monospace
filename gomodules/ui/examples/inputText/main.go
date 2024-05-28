package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/software-t-rex/monospace/gomodules/ui"
)

func main() {
	theme := ui.SetTheme(ui.ThemeMonoSpace)
	res, errInput := ui.NewInputText("Enter your name (you can try to autocomplete)").
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
		WithCompletion(func(wordStart, fullWord string) ([]string, error) {
			knownNames := []string{
				"Adrien", "Adele", "Alice",
				"Bob", "Brenda",
				"Carlos", "Carol", "Cecil", "Charlie",
				"David", "Deborah", "Derek", "Diane", "Don",
				"Edward", "Eric", "Eve", "Eva", "Evans", "Ezra",
				"Fiona", "Frank", "Fred", "Frederic",
				"Grace", "George", "Gina",
				"Heidi", "Howard",
				"Ivan",
				"Jack", "James", "Jane", "Janet", "Jason", "John", "Judy",
				"Kevin", "Karen", "Kirk",
				"Larry", "Linda", "Lionel", "Liz", "Lloyd", "Lucy", "Luke", "Lynn",
				"Mallory", "Megan",
				"Nancy", "Nathan", "Nina", "Nuno",
				"Olivia", "Oliver", "Oscar",
				"Pam", "Paul", "Peter", "Philip", "Peggy",
				"Quentin", "Quill",
				"Richard", "Romeo",
				"Sally", "Steve", "Sybil",
				"Trent", "Tom", "Thomas",
				"Ursula", "Victor", "Vince", "Vincent", "Walter", "Xavier", "Yvonne", "Zach", "Zelda",
			}

			res := []string{}
			for _, name := range knownNames {
				if wordStart == "" || len(wordStart) >= 1 && strings.HasPrefix(strings.ToLower(name), strings.ToLower(wordStart)) {
					res = append(res, name)
				}
			}
			// if any completion add the provided word to the list of possible completion
			if len(res) > 0 && fullWord != wordStart {
				res = append(res, fullWord)
			}
			return res, nil
		}).
		WithCleanup().
		WithKeyHandler(func(key string) (ui.Msg, error) {
			if key == "esc" {
				return ui.MsgQuit{}, nil
			}
			return nil, nil // returning nil,nil will let the input handle the key
		}).
		// Inline().
		Run()
	if errInput != nil {
		ui.Println(fmt.Sprintf("Error: %s", errInput), ui.BrightRed.Foreground())
		return
	}
	ui.Println(fmt.Sprintf("Hello "+theme.Bold("%#v!"), res), ui.BrightBlue.Foreground())

	pass, ErrPass := ui.NewInputPassword("Enter your password").
		Inline().
		WithValidator(func(str string) error {
			if str == "" {
				return fmt.Errorf("password can't be empty")
			}
			if len(str) < 8 {
				return fmt.Errorf("password must be at least 8 characters long")
			}
			tests := []string{".{7,}", "[a-zA-Z]", "[0-9]", "[^\\d\\w]"}
			for _, test := range tests {
				t, _ := regexp.MatchString(test, str)
				if !t {
					return fmt.Errorf("password must contains letters, digits and special characters")
				}
			}
			return nil
		}).
		Run()
	if ErrPass != nil {
		ui.Println(fmt.Sprintf("Error: %s", ErrPass), ui.BrightRed.Foreground())
		return
	}
	ui.Println(fmt.Sprintf("Your password is '%s'", theme.Bold(pass)), ui.BrightBlue.Foreground())

}
