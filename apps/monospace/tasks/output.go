/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package tasks

import "io"

// taskOutputWriter writes task output prefixed by the task display name.
//
// It tracks line boundaries instead of prefixing every Write call: command
// output is received in arbitrary chunks, so a single line can be split across
// multiple writes. Only line starts are prefixed, and partial lines are written
// as they arrive.
//
// A writer instance belongs to a single job and is not safe for concurrent use.
type taskOutputWriter struct {
	out         io.Writer
	prefix      []byte
	atLineStart bool
}

func newTaskOutputWriter(out io.Writer, taskName string) *taskOutputWriter {
	return &taskOutputWriter{
		out:         out,
		prefix:      []byte(taskName + ": "),
		atLineStart: true,
	}
}

func (w *taskOutputWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	out := make([]byte, 0, len(p)+len(w.prefix))
	for _, b := range p {
		if w.atLineStart {
			out = append(out, w.prefix...)
			w.atLineStart = false
		}
		out = append(out, b)
		if b == '\n' {
			w.atLineStart = true
		}
	}
	// Output errors are deliberately ignored: a closed stdout (for example
	// `monospace run | head`) must not fail the task nor prevent the cache fill.
	_, _ = w.out.Write(out)
	return len(p), nil
}

// Flush terminates the current line if the task output did not end with a
// newline, so subsequent task outputs stay on their own line.
func (w *taskOutputWriter) Flush() {
	if w.atLineStart {
		return
	}
	_, _ = w.out.Write([]byte{'\n'})
	w.atLineStart = true
}
