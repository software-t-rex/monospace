package ui

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSGRCombine(t *testing.T) {
	result := SGRCombine(Bold, Underline)
	expected := "1;4"
	if string(result) != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
func TestSGREscapeSequence(t *testing.T) {
	result := SGREscapeSequence(Bold, Underline)
	expected := "\033[1;4m"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
	// test for empty styles
	result = SGREscapeSequence([]SGRParam{}...)
	expected = ""
	if result != expected {
		t.Errorf("Expected '%#v', got '%#v'", expected, result)
	}
}
func TestSGRResetSequence(t *testing.T) {
	tests := []struct {
		name     string
		styles   []SGRParam
		expected string
	}{
		{
			name:     "Test with foreground color",
			styles:   []SGRParam{Red.Foreground(), Bold, Underline},
			expected: csiStart + "39;22;24" + csiEnd,
		},
		{
			name:     "Test with Reverse Black background that blink",
			styles:   []SGRParam{Black.Background(), Reversed, Blink},
			expected: csiStart + "49;27;25" + csiEnd,
		},
		// Add more test cases as needed
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SGRResetSequence(test.styles...)
			if !strings.Contains(result, test.expected) {
				t.Errorf("Expected %q to contain %q", result, test.expected)
			}
		})
	}
}
func TestApplyStyle(t *testing.T) {
	canEnhance = true
	enabledEnhanced = true
	result := ApplyStyle("test", Bold, Underline)
	expected := "\033[1;4mtest\033[22;24m"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	enabledEnhanced = false
	result = ApplyStyle("test", Bold, Underline)
	expected = "test"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestNewStyler(t *testing.T) {
	canEnhance = true
	enabledEnhanced = true
	styler := NewStyler(Bold, Underline)
	result := styler("test")
	expected := "\033[1;4mtest\033[22;24m"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	enabledEnhanced = false
	result = styler("test")
	expected = "test"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestPrintln(t *testing.T) {
	canEnhance = true
	enabledEnhanced = true
	// Redirect standard output to a buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var buf bytes.Buffer
	var mu sync.Mutex
	scanner := bufio.NewScanner(r)
	go func() {
		for scanner.Scan() {
			mu.Lock()
			buf.WriteString(scanner.Text() + "\n")
			mu.Unlock()
		}
	}()

	defer func() {
		os.Stdout = old
		w.Close()
	}()
	redBold := []SGRParam{Red.Foreground(), Bold}
	tests := []struct {
		style           []SGRParam
		canEnhance      bool
		enhancedEnabled bool
		expected        string
	}{
		{redBold, true, true, "\033[31;1mtest\033[39;22m\n"},
		{redBold, true, false, "test\n"},
		{redBold, false, true, "test\n"},
		{redBold, false, false, "test\n"},
	}

	for _, test := range tests {
		canEnhance = test.canEnhance
		enabledEnhanced = test.enhancedEnabled
		Println("test", test.style...)
		time.Sleep(10 * time.Millisecond) // give some time for the goroutine to read from the pipe
		mu.Lock()
		output := buf.String()
		mu.Unlock()
		if output != test.expected {
			t.Errorf("Expected %q, got %q", test.expected, output)
		}
		mu.Lock()
		buf.Reset()
		mu.Unlock()
	}
}
