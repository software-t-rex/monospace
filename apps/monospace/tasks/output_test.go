/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package tasks

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"

	jobExecutor "github.com/software-t-rex/go-jobExecutor/v2"
)

func TestTaskOutputWriter_PartialWrites(t *testing.T) {
	var buf bytes.Buffer
	w := newTaskOutputWriter(&buf, "task")

	if _, err := w.Write([]byte("hel")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := w.Write([]byte("lo\nnext ")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := w.Write([]byte("line")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	w.Flush()

	if want := "task: hello\ntask: next line\n"; buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

func TestTaskOutputWriter_IgnoresOutputErrors(t *testing.T) {
	w := newTaskOutputWriter(failingWriter{}, "task")
	n, err := w.Write([]byte("hello\n"))
	if err != nil {
		t.Errorf("Write: unexpected error: %v", err)
	}
	if n != len("hello\n") {
		t.Errorf("Write returned %d, want %d", n, len("hello\n"))
	}
	w.Flush() // must not panic nor propagate the failing writer error
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("broken output")
}

// TestInterleaved_DirectCmd_PartialWrites checks the real executor wiring:
// output split across several writes must be prefixed once, and a missing
// trailing newline must be flushed when the job completes.
func TestInterleaved_DirectCmd_PartialWrites(t *testing.T) {
	e := NewExecutor("interleaved")
	e.AddJob(jobExecutor.NamedJob{
		Name: "myproject#partial",
		Job:  exec.Command("sh", "-c", "printf 'par'; sleep 0.05; printf 'tial'"),
	})

	captured := captureStdout(t, func() {
		if errs := e.DagExecute(); errs.Len() > 0 {
			t.Errorf("DagExecute: unexpected errors: %v", errs)
		}
	})
	clean := stripAnsi(captured)

	if n := strings.Count(clean, "myproject#partial: partial"); n != 1 {
		t.Errorf("streamed output should be prefixed exactly once, got %d occurrences:\n%s", n, clean)
	}
	if !strings.Contains(clean, "myproject#partial: partial\n") {
		t.Errorf("streamed output should end with a flushed newline:\n%q", clean)
	}
}

// TestInterleaved_DirectCmd_ReportsStartError checks that an error that
// produced no output (here: command not found) is still displayed.
func TestInterleaved_DirectCmd_ReportsStartError(t *testing.T) {
	for _, mode := range []string{"interleaved", "errors-only"} {
		t.Run(mode, func(t *testing.T) {
			e := NewExecutor(mode)
			e.AddJob(jobExecutor.NamedJob{
				Name: "myproject#missing",
				Job:  exec.Command("monospace-test-no-such-binary"),
			})
			captured := captureStdout(t, func() {
				e.DagExecute()
			})
			clean := stripAnsi(captured)
			if !strings.Contains(clean, "myproject#missing: ") || !strings.Contains(clean, "monospace-test-no-such-binary") {
				t.Errorf("start error should be displayed with the task prefix:\n%q", clean)
			}
		})
	}
}
