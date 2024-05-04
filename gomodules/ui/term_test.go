package ui

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"golang.org/x/term"
)

// only contains mock for termInterface used in term.go
type MockTerm struct {
	isTerm   bool
	isDarkBg bool
	reader   *bufio.Reader
	writer   *bufio.Writer
}

func (t *MockTerm) MakeRaw() (*term.State, error) {
	return nil, nil
}
func (t *MockTerm) Restore(*term.State) error { return nil }
func (t *MockTerm) HandleState(bool) (func(), error) {
	return nil, nil
}
func (t *MockTerm) IsTerminal() bool {
	return t.isTerm
}
func (t *MockTerm) HasDarkBackground() (bool, error) { return t.isDarkBg, nil }

// convenience methods for tests to set mock term is a term, it will rerun detectCapability
func (t *MockTerm) setIsTerm(isTerm bool) *MockTerm {
	t.isTerm = isTerm
	detectCapability(t)
	return t
}
func (t *MockTerm) setIsDarkBg(isDarkBg bool) *MockTerm {
	t.isDarkBg = isDarkBg
	return t
}
func (t *MockTerm) NewReader() *bufio.Reader {
	return t.reader
}
func (t *MockTerm) NewScanner() *bufio.Scanner {
	return bufio.NewScanner(t.reader)
}
func (t *MockTerm) Write(a ...any) (int, error) {
	return fmt.Fprint(t.writer, a...)
}

func (t *MockTerm) Tty() *os.File {
	return nil
}

func TestGetSetUsedTerm(t *testing.T) {
	// reset usedTerm
	usedTerm = nil
	// test set usedTerm
	mockTerm := &MockTerm{}
	SetTerminal(mockTerm)
	if GetTerminal() != mockTerm {
		t.Error("getUsedTerm() should return mockTerm")
	}
}
