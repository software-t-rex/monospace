/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var defaultMaxVisibleOptions = 5

// set a number either 0 or between 3 and 15
func SetDefaultMaxVisibleOptions(max int) {
	if max > 0 && max < 3 {
		max = 3
	} else if max < 0 || max > 15 {
		max = 0
	}
	defaultMaxVisibleOptions = max
}

type SelectOption[T comparable] struct {
	Value T
	Label string
}

type multiSelectModel[T comparable] struct {
	title             string
	options           []SelectOption[T]
	selected          map[int]bool
	selectionMaxLen   int
	selectionMinLen   int // will default to 1
	maxVisibleOptions int
	startVisibleIndex int
	focusedIndex      int
	errorMsg          string
	done              bool
	bindings          *KeyBindings[*multiSelectModel[T]]
	singleSelect      bool
	cleanup           bool
}

//#region - handle model settings

func (m *multiSelectModel[T]) clampLen(l int) int {
	if l < 0 {
		l = 0
	} else if l > len(m.options) {
		l = len(m.options)
	}
	return l
}

// Set the maximum number of options that can be selected
// Set to 0 for no limit (the default)
func (m *multiSelectModel[T]) SelectionMaxLen(maxLen int) *multiSelectModel[T] {
	m.selectionMaxLen = m.clampLen(maxLen)
	if m.selectionMaxLen > 0 && (m.selectionMaxLen < m.selectionMinLen) {
		m.selectionMinLen = m.selectionMaxLen
	}
	return m
}

// Set the maximum number of options that can be selected
// Set to 0 for no limit (default to 1)
func (m *multiSelectModel[T]) SelectionMinLen(minLen int) *multiSelectModel[T] {
	m.selectionMinLen = m.clampLen(minLen)
	if m.selectionMinLen > m.selectionMaxLen {
		m.selectionMaxLen = m.selectionMinLen
	}
	return m
}

// this is an alias for SelectionMinLen(0)
func (m *multiSelectModel[T]) AllowEmptySelection() *multiSelectModel[T] {
	m.selectionMinLen = 0
	return m
}

// this is an alias for calling SelectionMaxLen(1) and SelectionMinLen(1) with the same value
func (m *multiSelectModel[T]) SelectionExactLen(l int) *multiSelectModel[T] {
	l = m.clampLen(l)
	m.selectionMaxLen = l
	m.selectionMinLen = l
	return m
}

/** Beware that this feature is not supported in fallback mode */
func (m *multiSelectModel[T]) SelectedIndexes(indexes ...int) *multiSelectModel[T] {
	m.selected = make(map[int]bool, len(m.options))
	if m.selectionMaxLen > 0 && len(indexes) > m.selectionMaxLen {
		indexes = indexes[:m.selectionMaxLen]
	}
	for _, index := range indexes {
		if index < 0 || index >= len(m.options) {
			continue
		}
		m.selected[index] = true
	}
	// place focus on the first selected item
	if len(indexes) > 0 {
		m.focusedIndex = indexes[0]
	}
	return m
}

// Set the max number of options to display at once (between 3 to 15)
// Set to 0 for no limit default to defaultMaxVisibleOptions
// You can call SetDefaultMaxVisibleOptions(max int) to set a default value
func (m *multiSelectModel[T]) MaxVisibleOptions(nb int) *multiSelectModel[T] {
	if nb > 0 && nb < 3 {
		nb = 3
	} else if nb < 0 || nb > 15 {
		nb = 0
	}
	m.maxVisibleOptions = nb
	return m
}

// Remove the menu from the output when done
// default to false (menu will remain visible)
// Ignored in fallback mode.
func (m *multiSelectModel[T]) WithCleanup(clear bool) *multiSelectModel[T] {
	m.cleanup = clear
	return m
}

// Set options from a slice of strings
// This is a shortcut for when you only need to select between strings
func (m *multiSelectModel[T]) SetStringsOptions(options []string) *multiSelectModel[T] {
	selectOptions := make([]SelectOption[T], len(options))
	for i, option := range options {
		selectOptions[i] = SelectOption[T]{Label: option, Value: any(option).(T)}
	}
	return m.SetOptions(selectOptions)
}

// Set options from a slice of SelectOption
func (m *multiSelectModel[T]) SetOptions(options []SelectOption[T]) *multiSelectModel[T] {
	m.options = options
	m.selected = make(map[int]bool, len(options))
	if m.singleSelect {
		m.selected[0] = true
		m.focusedIndex = 0
	}
	return m
}

//#endregion - handle model settings

/** return a slice of selected items */
func (m *multiSelectModel[T]) getSelected() []T {
	res := []T{}
	for i, option := range m.options {
		if m.selected[i] {
			res = append(res, any(option.Value).(T))
		}
	}
	return res
}

func (m *multiSelectModel[T]) Init() tea.Cmd {
	// if single selection mode, focus on the first selected item
	if m.focusedIndex >= m.maxVisibleOptions {
		// place selected option in the middle of the visible options
		m.startVisibleIndex = m.focusedIndex - m.maxVisibleOptions/2 + 1
		if m.startVisibleIndex > len(m.options)-m.maxVisibleOptions {
			m.startVisibleIndex = len(m.options) - m.maxVisibleOptions
		}
	}
	// if single selection mode, focus on the first selected item
	// always return nil for convenience
	ifSingleSelectFocusIndex := func() tea.Cmd {
		if m.singleSelect {
			m.selected = make(map[int]bool, len(m.options))
			m.selected[m.focusedIndex] = true
		}
		return nil
	}
	m.bindings = NewKeyBindings[*multiSelectModel[T]]().
		AddBinding("down,j", Msgs["down"], func(m *multiSelectModel[T]) tea.Cmd {
			if m.focusedIndex < len(m.options)-1 {
				m.focusedIndex++
			}
			if (m.focusedIndex - m.startVisibleIndex) >= m.maxVisibleOptions {
				m.startVisibleIndex++
			}
			return ifSingleSelectFocusIndex()
		}).
		AddBinding("up,k", Msgs["up"], func(m *multiSelectModel[T]) tea.Cmd {
			if m.focusedIndex > 0 {
				m.focusedIndex--
			}
			if m.focusedIndex >= 0 && m.focusedIndex < m.startVisibleIndex {
				m.startVisibleIndex = m.focusedIndex
			}
			return ifSingleSelectFocusIndex()
		})
	if !m.singleSelect {
		m.bindings.AddBinding(" ,x", Msgs["select"], func(m *multiSelectModel[T]) tea.Cmd {
			// if index is selected unselect it
			if m.selected[m.focusedIndex] {
				m.selected[m.focusedIndex] = false
				return nil
			}
			// single selection mode
			if m.selectionMaxLen == 1 {
				m.selected = make(map[int]bool, len(m.options))
				m.selected[m.focusedIndex] = true
			}
			// multi selection mode: check selection has not reached the max length
			if m.selectionMaxLen > 0 {
				selectedCount := 0
				for _, isSelected := range m.selected {
					if isSelected {
						selectedCount++
					}
				}
				if selectedCount >= m.selectionMaxLen && !m.selected[m.focusedIndex] {
					m.errorMsg = Msgs["limitReached"]
					return nil
				}
			}
			// mark focus item as selected
			m.selected[m.focusedIndex] = true
			return nil
		})
	}
	m.bindings.AddBinding("enter", Msgs["confirm"], func(m *multiSelectModel[T]) tea.Cmd {
		if m.selectionMinLen > 0 {
			selectedCount := 0
			for _, isSelected := range m.selected {
				if isSelected {
					selectedCount++
				}
			}
			if selectedCount < m.selectionMinLen {
				m.errorMsg = fmt.Sprintf(Msgs["outBoundMin"], m.selectionMinLen)
				return nil
			}
		}
		m.done = true
		return tea.ClearScreen
	}).
		// following bindings are not displayed in the help message
		AddBinding("ctrl+c", "", func(m *multiSelectModel[T]) tea.Cmd {
			return AbortTeaProgram // this will exit the program
		}).
		AddBinding("home", "", func(m *multiSelectModel[T]) tea.Cmd {
			m.focusedIndex = 0
			m.startVisibleIndex = 0
			return ifSingleSelectFocusIndex()
		}).
		AddBinding("end", "", func(m *multiSelectModel[T]) tea.Cmd {
			m.focusedIndex = len(m.options) - 1
			m.startVisibleIndex = len(m.options) - m.maxVisibleOptions
			if m.startVisibleIndex < 0 {
				m.startVisibleIndex = 0
			}
			return ifSingleSelectFocusIndex()
		}).
		AddBinding("left,h", "", func(m *multiSelectModel[T]) tea.Cmd {
			m.focusedIndex -= m.maxVisibleOptions
			if m.focusedIndex < 0 {
				m.focusedIndex = 0
			}
			m.startVisibleIndex -= m.maxVisibleOptions
			if m.startVisibleIndex < 0 {
				m.startVisibleIndex = 0
			}
			return ifSingleSelectFocusIndex()
		}).
		AddBinding("right,l", "", func(m *multiSelectModel[T]) tea.Cmd {
			m.focusedIndex += m.maxVisibleOptions
			if m.focusedIndex >= len(m.options) {
				m.focusedIndex = len(m.options) - 1
			}
			m.startVisibleIndex += m.maxVisibleOptions
			if m.startVisibleIndex >= len(m.options)-m.maxVisibleOptions {
				m.startVisibleIndex = len(m.options) - m.maxVisibleOptions
			}
			return ifSingleSelectFocusIndex()
		})
	return nil
}

// h,j,k,l = left, down, up, right
func (m *multiSelectModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.done = false
	m.errorMsg = ""
	cmd := m.bindings.Handle(m, msg)
	if m.done {
		return m, tea.Quit
	}
	return m, cmd
}

func (m *multiSelectModel[T]) View() string {
	if m.done && m.cleanup {
		return ""
	}
	theme := GetTheme()
	needMoreIndicator := false
	if m.maxVisibleOptions < len(m.options) {
		needMoreIndicator = true
	}
	var sb strings.Builder
	sb.WriteString(theme.Title(m.title))
	sb.WriteString("\n")
	for i, option := range m.options {
		if i < m.startVisibleIndex {
			continue
		} else if i >= m.startVisibleIndex+m.maxVisibleOptions {
			break
		}
		label := option.Label
		if m.selected[i] {
			label = theme.Bold(label)
		}
		if i == m.focusedIndex {
			label = theme.Success(label)
		}
		if needMoreIndicator {
			if i == m.startVisibleIndex && i > 0 {
				sb.WriteString(theme.MoreUpIndicator())
			} else if i == m.startVisibleIndex+m.maxVisibleOptions-1 && i < len(m.options)-1 {
				sb.WriteString(theme.MoreDownIndicator())
			} else {
				sb.WriteString(" ")
			}
			sb.WriteString(" ")
		}
		sb.WriteString(theme.ConditionalFocusIndicator(m.focusedIndex == i))
		sb.WriteString(" ")
		if !m.singleSelect {
			sb.WriteString(theme.ConditionalSelectedIndicator(m.selected[i]))
			sb.WriteString(" ")
		}
		sb.WriteString(label)
		sb.WriteString("\n")
	}
	if !m.done {
		// sb.WriteString("\n")
		if m.errorMsg != "" {
			sb.WriteString(theme.Error(m.errorMsg))
		} else {
			sb.WriteString(m.bindings.GetDescription())
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (m *multiSelectModel[T]) Fallback() TeaModelWithFallback {
	var sb strings.Builder
	// reset selected (we don't want to keep the previous selection while in fallback mode)
	m.selected = make(map[int]bool, len(m.options))
	sb.WriteString(m.title)
	sb.WriteString("\n")
	for i, option := range m.options {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, option.Label))
	}
	if m.singleSelect {
		sb.WriteString(Msgs["fallbackSelectPrompt"])
	} else {
		sb.WriteString(Msgs["fallbackMultiSelectPrompt"])
	}
	sb.WriteString("\n")
	if m.errorMsg != "" {
		sb.WriteString(Msgs["errorPrefix"])
		sb.WriteString(m.errorMsg)
		sb.WriteString("\n")
		m.errorMsg = ""
	}
	// prompt user for input
	ints, err := ReadInts(sb.String())
	// update the model
	if err != nil {
		m.errorMsg = Msgs["notANumber"]
		return m.Fallback()
	} else if m.selectionMinLen > 0 && len(ints) < m.selectionMinLen {
		m.errorMsg = fmt.Sprintf(Msgs["outBoundMin"], m.selectionMinLen)
		return m.Fallback()
	} else if m.selectionMaxLen > 0 && len(ints) > m.selectionMaxLen {
		m.errorMsg = fmt.Sprintf(Msgs["outBoundMax"], m.selectionMaxLen)
		return m.Fallback()
	}
	for _, i := range ints {
		if i < 1 || i > len(m.options) {
			m.errorMsg = fmt.Sprintf(Msgs["outOfRange"], len(m.options))
			return m.Fallback()
		}
		m.selected[i-1] = true
	}
	if m.errorMsg != "" {
		return m.Fallback()
	}
	return m
}

func newMultiSelect[T comparable](title string, singleSelect bool) *multiSelectModel[T] {
	m := &multiSelectModel[T]{title: title, singleSelect: singleSelect}
	m.maxVisibleOptions = defaultMaxVisibleOptions
	m.selectionMinLen = 1
	if singleSelect {
		m.selectionMaxLen = 1
	}
	return m
}

// NewMultiSelect creates a new multi-select model with the given title and options.
// The options should be a slice of MultiSelectOption, where each option is a value that implements the comparable interface.
//
// Usage:
//
//	type MyOption struct {
//	    Name string
//	    Value int // can be any comparable type
//	}
//
//	options := []MultiSelectOption[MyOption]{
//	    MyOption{Name: "Option 1", Value: 1},
//	    MyOption{Name: "Option 2", Value: 2},
//	    MyOption{Name: "Option 3", Value: 3},
//	}
//
// multiSelect := NewMultiSelect[MyOption]("Choose an option", options)
// selected := multiSelect.Run()
//
//	for _, option := range selected {
//	    fmt.Printf("Selected option: %s, value: %d\n", option.Name, option.Value)
//	}
func NewMultiSelect[T comparable](title string, options []SelectOption[T]) *multiSelectModel[T] {
	return newMultiSelect[T](title, false).SetOptions(options)
}

// NewMultiSelectStrings creates a new multi-select model with the given title
// and options. The options should be a slice of strings.
// This is a shortcut for when you only need to select between strings
// without having to manually prepare a slice of SelectOption.
//
// Usage:
//
// options := []string{"Option 1", "Option 2", "Option 3"}
//
// multiSelect := NewMultiSelectStrings("Choose an option", options)
// selected := multiSelect.Run()
//
//	for _, option := range selected {
//	    fmt.Printf("Selected option: %s\n", option)
//	}
func NewMultiSelectStrings(title string, options []string) *multiSelectModel[string] {
	return newMultiSelect[string](title, false).SetStringsOptions(options)
}

// Run the multi-select model and return a slice of selected items.
func (m *multiSelectModel[T]) Run() []T {
	return runTeaProgram(m).getSelected()
}
