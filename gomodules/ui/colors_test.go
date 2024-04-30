package ui

import (
	"testing"
)

func TestAdaptiveColor(t *testing.T) {
	canEnhance = true
	enabledEnhanced = true
	mockTerm := &MockTerm{isTerm: false, isDarkBg: false}
	SetTerminal(mockTerm)
	tests := []struct {
		color         AdaptiveColor
		expectedDark  SGRParam
		expectedLight SGRParam
	}{
		{AdaptiveColor{Dark: Color("#eee"), Light: Color("#111")}, "238;238;238", "17;17;17"},
		{AdaptiveColor{Dark: Color("#222"), Light: Color("#aaa")}, "34;34;34", "170;170;170"},
	}

	for _, test := range tests {
		mockTerm.setIsDarkBg(false)
		resultFg := test.color.Foreground()
		resultBg := test.color.Background()
		if resultFg != "38;2;"+test.expectedLight {
			t.Errorf("Expected %v, got %v", test.expectedLight, resultFg)
		}
		if resultBg != "48;2;"+test.expectedLight {
			t.Errorf("Expected %v, got %v", test.expectedLight, resultBg)
		}
		mockTerm.setIsDarkBg(true)
		resultFg = test.color.Foreground()
		resultBg = test.color.Background()
		if resultFg != "38;2;"+test.expectedDark {
			t.Errorf("Expected %v, got %v", test.expectedDark, resultFg)
		}
		if resultBg != "48;2;"+test.expectedDark {
			t.Errorf("Expected %v, got %v", test.expectedDark, resultBg)
		}
	}
}

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		hex           string
		expected      RGB
		expectedError bool
	}{
		{"#ffffff", RGB{255, 255, 255}, false},
		{"#123456", RGB{18, 52, 86}, false},
		{"#8FA", RGB{136, 255, 170}, false},
		{"invalid", RGB{0, 0, 0}, true},
		{"#BADRGB", RGB{0, 0, 0}, true},
	}

	for _, test := range tests {
		result, err := hexToRGB(test.hex)
		if result != test.expected {
			t.Errorf("Expected %v, got %v from %s", test.expected, result, test.hex)
		}
		if (err != nil) != test.expectedError {
			t.Errorf("Expected error %v, got %v", test.expectedError, err)
		}
	}
}

func TestColorForeground(t *testing.T) {
	tests := []struct {
		color    Color
		expected SGRParam
	}{
		{"#ffffff", "38;2;255;255;255"},
		{"1", "31"},
		{"9", "38;5;9"},
		{"invalid", ""},
		{"555", ""}, // invalid ANSI256 color
	}

	for _, test := range tests {
		result := test.color.Foreground()
		if result != test.expected {
			t.Errorf("Expected '%v', got %v", test.expected, result)
		}
	}
}

func TestColorBackground(t *testing.T) {
	tests := []struct {
		color    Color
		expected SGRParam
	}{
		{"#123456", "48;2;18;52;86"},
		{"#fff", "48;2;255;255;255"},
		{"7", "47"},
		{"133", "48;5;133"},
		{"invalid", ""},
		{"555", ""}, // invalid ANSI256 color
		{"-1", ""},  // invalid ANSI256 color
	}

	for _, test := range tests {
		result := test.color.Background()
		if result != test.expected {
			t.Errorf("Expected '%v', got %v", test.expected, result)
		}
	}
}
