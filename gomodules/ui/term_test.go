package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

// only contains mock for termInterface used in term.go
type MockTerm struct {
	isTerm    bool
	isDarkBg  bool
	reader    *bufio.Reader
	writer    *bufio.Writer
	readMutex sync.Mutex
}

func (t *MockTerm) MakeRaw() (*TermState, error) {
	return nil, nil
}
func (t *MockTerm) Restore(*TermState) error { return nil }
func (t *MockTerm) HandleState(bool) (func(), error) {
	return func() {}, nil
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
func (t *MockTerm) ExclusiveReader() (*bufio.Reader, func()) {
	t.readMutex.Lock()
	return t.reader, func() { t.readMutex.Unlock() }
}
func (t *MockTerm) NewScanner() *bufio.Scanner {
	return bufio.NewScanner(t.reader)
}
func (t *MockTerm) Write(a ...any) (int, error) {
	return fmt.Fprint(t.writer, a...)
}
func (t *MockTerm) GetSize() (int, int, error) {
	return 0, 0, nil
}
func (t *MockTerm) Tty() *os.File {
	return nil
}
func (t *MockTerm) DeviceStatusReport() (row, col int, err error) {
	return 0, 0, nil
}

// start capturing stdout
// returns a restore function which will restore stdout to its original state and returns the captured output
func startCaptureStdout() func() string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	return func() string {
		w.Close()
		os.Stdout = old
		out, _ := io.ReadAll(r)
		return string(out)
	}
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

func TestTerminalQuery(t *testing.T) {
	testCases := []struct {
		name           string
		querySequence  string
		expectedPrefix string
		endVerifier    func(b byte, res []byte) (bool, error)
		fakeReply      string
		want           string
		expectedErr    error
	}{
		{
			name:           "working case",
			querySequence:  "\x1b[6n",
			expectedPrefix: "\x1b[",
			endVerifier: func(b byte, _ []byte) (bool, error) {
				return b == 'R', nil
			},
			fakeReply: "\x1b[2;5R",
			want:      "2;5R",
		},
		{
			name:           "timeout",
			querySequence:  "\x1b[6n",
			expectedPrefix: "\x1b[",
			endVerifier: func(b byte, _ []byte) (bool, error) {
				return b == 'R', nil
			},
			want:        "",
			expectedErr: ErrTimeout,
		},
		{
			name:           "neverEnding",
			querySequence:  "\x1b[6n",
			expectedPrefix: "\x1b[",
			endVerifier: func(b byte, _ []byte) (bool, error) {
				return b == 'R', nil
			},
			fakeReply:   "\x1b[2;5",
			want:        "",
			expectedErr: ErrTimeout,
		},
		{
			name:           "badFormat",
			querySequence:  "\x1b[6n",
			expectedPrefix: "\x1b[",
			endVerifier: func(b byte, _ []byte) (bool, error) {
				return b == 'R', nil
			},
			fakeReply:   "not a valid reply",
			want:        "",
			expectedErr: ErrBadFormat,
		},
	}

	rIn, wIn, _ := os.Pipe()
	mockTerminal := &MockTerm{
		reader: bufio.NewReader(rIn),
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// capture output
			restoreOutput := startCaptureStdout()
			opts := terminalQueryOpts{
				querySequence:  tc.querySequence,
				expectedPrefix: tc.expectedPrefix,
				endVerifier:    tc.endVerifier,
				timeout:        time.Millisecond * 500,
			}
			// send fakeReply in advance
			if tc.fakeReply != "" {
				wIn.WriteString(tc.fakeReply)
			}
			res, err := terminalQuery(mockTerminal, opts)
			// restore stdout
			received := restoreOutput()
			assert.Equal(t, string(received), tc.querySequence)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NilError(t, err)
			}
			assert.Equal(t, tc.want, res)
		})
	}

}
