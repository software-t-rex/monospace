package main

import (
	"fmt"
	"os"

	"github.com/software-t-rex/monospace/gomodules/ui"
)

var choices = []string{"Banana", "Orange", "Strawberry", "Cherry", "Coconut", "Pineapple", "Mango", "Kiwi", "Papaya", "Pomegranate", "Grape", "Apple"}

func printError(err error) {
	fmt.Fprintln(os.Stderr, ui.GetTheme().Error("Error: "+err.Error()))
}

func main() {
	fruitSelector := ui.NewMultiSelectStrings("Select 2 to 3 fruits", choices).
		Escapable(true).
		SelectionMaxLen(3).
		SelectionMinLen(2).
		SelectedIndexes(0, 1, 2)
	selectedFruits, errFruits := fruitSelector.Run()
	if errFruits != nil {
		printError(errFruits)
		return
	}

	fmt.Println("Selected fruits:", selectedFruits)

	vegetables := []ui.SelectOption[int]{
		{Value: 1, Label: "Carrot (value int 1)"},
		{Value: 2, Label: "Cucumber (value int 2)"},
		{Value: 3, Label: "Tomato (value int 3)"},
		{Value: 4, Label: "Potato (value int 4)"},
		{Value: 5, Label: "Garlic (value int 5)"},
		{Value: 6, Label: "Spinach (value int 6)"},
		{Value: 7, Label: "Broccoli (value int 7)"},
		{Value: 8, Label: "Radish (value int 8)"},
	}
	selectedVegetables, errVegetables := ui.NewMultiSelect("Select some vegetables", vegetables).
		SelectedIndexes(4).
		Run()
	if errVegetables != nil {
		printError(errVegetables)
		return
	}

	fmt.Println("Selected vegetables: ", selectedVegetables)

	if ui.EnhancedEnabled() {
		fmt.Println("--- Re run in fallback mode ---")
		ui.ToggleEnhanced(false)
		main()
	}
}
