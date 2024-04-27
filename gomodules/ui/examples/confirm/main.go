package main

import (
	"fmt"

	"github.com/software-t-rex/monospace/gomodules/ui"
)

var themeChanged = false

func main() {
	confirm := ui.NewConfirm("This is an inline prompt without help do you want to continue ?", true).
		Inline().
		WithoutHelp().
		Run()
	if confirm {
		fmt.Println("Ok Here it is:")
		confirmAgain := ui.NewConfirm("This is a confirm with help can you confirm ?", false).
			WithCleanup(true). // ignored in fallback mode
			Run()
		if confirmAgain {
			fmt.Println("That was expected.")
		} else {
			fmt.Println("WTF it really should have been, you should report a bug !")
		}
	} else {
		fmt.Println("Bye !")
	}

	if ui.EnhancedEnabled() {
		if !themeChanged && ui.ConfirmInline("Do you want to retry with another theme ?", true) {
			ui.SetTheme(ui.ThemeMonoSpace)
			themeChanged = true
			main()
			return
		}
		fmt.Println("--- Re run in fallback mode ---")
		ui.ToggleEnhanced(false)
		main()
	}

}
