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

var KeyDescriptors = map[string]string{
	"left":      "←",
	"right":     "→",
	"up":        "↑",
	"down":      "↓",
	"enter":     "↵",
	" ":         "space",
	"esc":       "escape",
	"tab":       "⇥",
	"shift+tab": "⇤",
	"backspace": "⌫",
	"delete":    "⌦",
	"home":      "⇱",
	"end":       "⇲",
	"pgup":      "⇞",
	"pgdown":    "⇟",
}

type KeyBindings[T tea.Model] struct {
	Handlers    map[string]func(T) tea.Cmd
	Description []string
}

func NewKeyBindings[T tea.Model]() *KeyBindings[T] {
	return &KeyBindings[T]{Handlers: make(map[string]func(T) tea.Cmd)}
}

// Add a key binding to the key bindings
// @param keysToBind multiple keys can be added separated by a comma
func (k *KeyBindings[T]) AddBinding(keysToBind string, desc string, handler func(T) tea.Cmd) *KeyBindings[T] {
	theme := GetTheme()
	keys := strings.Split(keysToBind, ",")
	cleanKeys := []string{}
	for _, key := range keys {
		if desc, ok := KeyDescriptors[key]; ok {
			cleanKeys = append(cleanKeys, desc)
		} else {
			cleanKeys = append(cleanKeys, key)
		}
	}
	// reset faint and bold with 22 and reenable them with 1;2
	keySeparator := "\033[22m\033[2m" + theme.KeySeparator() + "\033[1m"
	if desc != "" {
		// we assume that escapes code are safe to use in description as it won't be used for fallback mode
		k.Description = append(k.Description, fmt.Sprintf("\033[37;2;1m%s\033[0m \033[37;2m%s\033[0m", strings.Join(cleanKeys, keySeparator), desc))
	}
	for _, key := range keys {
		k.Handlers[key] = handler
	}
	return k
}

// this method should be called from the Update method of the model to handle key bindings
func (k *KeyBindings[T]) Handle(m T, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if handler, ok := k.Handlers[msg.String()]; ok {
			return handler(m)
		}
	}
	return nil
}

func (k *KeyBindings[T]) AddToDescription(desc ...string) *KeyBindings[T] {
	for _, d := range desc {
		k.Description = append(k.Description, "\033[37;2m"+d+"\033[0m")
	}
	return k
}

// Should be called from the View method of the model to display the key bindings
func (k *KeyBindings[T]) GetDescription() string {
	theme := GetTheme()
	bindingSeparator := " \033[37;2m" + theme.KeyBindingSeparator() + "\033[0m "
	return strings.Join(k.Description, bindingSeparator)
}
