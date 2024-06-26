package ui

import "testing"

func TestDetectCapability(t *testing.T) {
	mockTerm := &MockTerm{isTerm: true, isDarkBg: true}
	SetTerminal(mockTerm)
	type testCase struct {
		noColor    bool
		accessible bool
		term       string
		isTerm     bool
		want       bool
		msg        string
	}
	cases := []testCase{
		{false, false, "xterm", true, true, "Expected canEnhance=true, got false"},
		{false, false, "xterm", false, false, "Expected canEnhance=false when not in terminal, got true"},
		{true, false, "xterm", true, false, "Expected canEnhance=false when NO_COLOR is set, got true"},
		{false, true, "xterm", true, false, "Expected canEnhance=false when ACCESSIBLE is set, got true"},
		{false, false, "dumb", true, false, "Expected canEnhance=false when TERM=dumb, got true"},
	}

	for _, c := range cases {
		env_term = c.term
		if c.noColor {
			env_nocolor = "1"
		} else {
			env_nocolor = ""
		}
		if c.accessible {
			env_accessible = "1"
		} else {
			env_accessible = ""
		}
		mockTerm.setIsTerm(c.isTerm) // rerun detectCapability with new values
		if canEnhance != c.want {
			t.Errorf(c.msg)
		}
	}

}
func TestEnhancedEnabled(t *testing.T) {
	canEnhance = true
	ToggleEnhanced(true)
	if !EnhancedEnabled() {
		t.Errorf("Expected EnhancedEnabled() to return true, got false")
	}
	ToggleEnhanced(false)
	if EnhancedEnabled() {
		t.Errorf("Expected EnhancedEnabled() to return false, got true")
	}
	canEnhance = false
	if EnhancedEnabled() {
		t.Errorf("Expected EnhancedEnabled() to return false, got true")
	}
	ToggleEnhanced(true)
	if EnhancedEnabled() {
		t.Errorf("Expected EnhancedEnabled() to return false, got true")
	}
}
