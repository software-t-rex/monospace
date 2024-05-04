package ui

import (
	"reflect"
	"testing"
)

type MockModel struct{ name string }

func TestAddBinding(t *testing.T) {
	SetTheme(ThemeDefault)
	kb := NewKeyBindings[MockModel]()
	nilHandler := func(m MockModel) Cmd { return nil }
	kb.AddBinding("a,b,c", "test", nilHandler)
	if len(kb.Handlers) != 3 {
		t.Errorf("AddBinding() failed, expected 3 handlers, got %v", len(kb.Handlers))
	}
	if _, ok := kb.Handlers["a"]; !ok {
		t.Errorf("AddBinding() failed, expected handler for key 'a'")
	}

	if _, ok := kb.Handlers["b"]; !ok {
		t.Errorf("AddBinding() failed, expected handler for key 'b'")
	}

	if _, ok := kb.Handlers["c"]; !ok {
		t.Errorf("AddBinding() failed, expected handler for key 'c'")
	}
	kb.AddBinding("ctrl+c", "exit", nilHandler)
	if len(kb.Handlers) != 4 {
		t.Errorf("AddBinding() failed, expected 4 handlers, got %v", len(kb.Handlers))
	}
	if _, ok := kb.Handlers["ctrl+c"]; !ok {
		t.Errorf("AddBinding() failed, expected handler for key 'ctrl+c'")
	}
}

func TestHandle(t *testing.T) {
	model := MockModel{"test"}
	nilCalls := 0
	quitCalls := 0
	nilHandler := func(m MockModel) Cmd {
		nilCalls += 1
		return nil
	}
	quitHandler := func(m MockModel) Cmd {
		quitCalls += 1
		return CmdQuit
	}
	kb := NewKeyBindings[MockModel]()
	kb.AddBinding("a,space", "test", nilHandler)
	kb.AddBinding("ctrl+c", "exit", quitHandler)
	kb.Handle(model, MsgKey{Value: "a"})
	if (nilCalls != 1) || (quitCalls != 0) {
		t.Errorf("Handle() failed, expected nilCalls=1, quitCalls=0, got nilCalls=%v, quitCalls=%v", nilCalls, quitCalls)
	}
	kb.Handle(model, MsgKey{Value: "space"})
	if (nilCalls != 2) || (quitCalls != 0) {
		t.Errorf("Handle() failed, expected nilCalls=2, quitCalls=0, got nilCalls=%v, quitCalls=%v", nilCalls, quitCalls)
	}
	cmd := kb.Handle(model, MsgKey{Value: "ctrl+c"})
	if quitCalls != 1 {
		t.Errorf("Handle() failed, expected quitHandler to have been called")
	}
	if reflect.ValueOf(cmd).Pointer() != reflect.ValueOf(CmdQuit).Pointer() {
		t.Errorf("Handle() failed, expected Quit")
	}
}

func TestGetDescription(t *testing.T) {
	theme := GetTheme()
	SetTheme(ThemeDefault)
	kb := NewKeyBindings[MockModel]()
	kb.AddBinding("a", "test", func(m MockModel) Cmd { return nil })
	kb.AddBinding("ctrl+c", "exit", func(m MockModel) Cmd { return CmdQuit })
	desc := kb.GetDescription()
	expected := "\n\x1b[39;2;1ma\x1b[22;39;2m test\x1b[0m\x1b[39;2m " + theme.KeyBindingSeparator() + " \x1b[22m\x1b[39;2;1mctrl+c\x1b[22;39;2m exit\x1b[0m"
	if desc != expected {
		t.Errorf("GetDescription() failed:\nexpected %#v\ngot      %#v", expected, desc)
	}
}
