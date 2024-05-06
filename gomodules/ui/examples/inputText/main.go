package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/software-t-rex/monospace/gomodules/ui"
)

func main() {
	theme := ui.SetTheme(ui.ThemeMonoSpace)
	res := ui.NewInputText("Enter your name (you can try to autocomplete)").
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
		WithCompletion(func(str string) ([]string, error) {
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
				if str == "" || len(str) >= 1 && strings.HasPrefix(strings.ToLower(name), strings.ToLower(str)) {
					res = append(res, name)
				}
			}
			return res, nil
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
