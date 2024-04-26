/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

type selectModel[T comparable] struct {
	multiSelect *multiSelectModel[T]
}

func (s *selectModel[T]) SelectedIndex(index int) *selectModel[T] {
	s.multiSelect.SelectedIndexes(index)
	return s
}

func (s *selectModel[T]) MaxVisibleOptions(maxVisibleOptions int) *selectModel[T] {
	s.multiSelect.MaxVisibleOptions(maxVisibleOptions)
	return s
}

func (s *selectModel[T]) WithCleanup(withCleanup bool) *selectModel[T] {
	s.multiSelect.WithCleanup(withCleanup)
	return s
}

func (s *selectModel[T]) SetOptions(options []SelectOption[T]) *selectModel[T] {
	s.multiSelect.SetOptions(options)
	return s
}

func (s *selectModel[T]) SetStringsOptions(options []string) *selectModel[T] {
	s.multiSelect.SetStringsOptions(options)
	return s
}

func (s *selectModel[T]) Run() T {
	res := s.multiSelect.Run()
	return res[0]
}

func newSelect[T comparable](title string) *selectModel[T] {
	return &selectModel[T]{multiSelect: newMultiSelect[T](title, true)}
}

func NewSelect[T comparable](title string, options []SelectOption[T]) *selectModel[T] {
	return newSelect[T](title).SetOptions(options)
}

func NewSelectStrings(title string, options []string) *selectModel[string] {
	return newSelect[string](title).SetStringsOptions(options)
}
