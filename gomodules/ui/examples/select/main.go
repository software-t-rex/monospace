package main

import (
	"fmt"
	"os"

	"github.com/software-t-rex/monospace/gomodules/ui"
)

func printError(err error) {
	fmt.Fprintln(os.Stderr, ui.GetTheme().Error("Error: "+err.Error()))
}

func main() {
	options := []ui.SelectOption[int]{
		{Value: 1, Label: "int 1"},
		{Value: 2, Label: "int 2"},
		{Value: 3, Label: "int 3"},
		{Value: 4, Label: "int 4"},
		{Value: 5, Label: "int 5"},
		{Value: 6, Label: "int 6"},
		{Value: 7, Label: "int 7"},
		{Value: 8, Label: "int 8"},
		{Value: 9, Label: "int 9"},
		{Value: 10, Label: "int 10"},
	}
	selection, err := ui.NewSelect("Please select an option", options).
		MaxVisibleOptions(4).
		SelectedIndex(6).
		Escapable(true).
		Run()
	if err != nil {
		printError(err)
		return
	}
	fmt.Printf("You choose: %v\n", selection)
	options2 := []string{"option 1", "option 2", "option 3"}
	selection2, err2 := ui.NewSelectStrings("Please select an option from a list of strings", options2).
		WithCleanup(true).
		Run()
	if err2 != nil {
		printError(err2)
		return
	}
	fmt.Printf("You choose: %v\n", selection2)

	if ui.EnhancedEnabled() {
		fmt.Println("--- Re run in fallback mode ---")
		ui.ToggleEnhanced(false)
		main()
	}
}
