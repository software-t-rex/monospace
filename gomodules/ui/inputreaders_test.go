/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/software-t-rex/monospace/gomodules/ui/pkg/sequencesKeys"
)

func TestReadKeyPressEvent(t *testing.T) {
	tests := []struct {
		input         string
		want          MsgKey
		expectedError error
	}{
		{
			input: "a",
			want:  MsgKey{Value: "a", Unknown: false, IsSeq: false, ByteSeq: []byte("a")},
		},
		{
			input: "\x1b[200~paste\x1b[201~",
			want:  MsgKey{Value: "paste", Unknown: false, IsSeq: true, ByteSeq: []byte("\x1b[200~paste\x1b[201~")},
		},
	}
	// check sequencesKeys are correctly read
	for seq, key := range sequencesKeys.Map {
		if seq == "\x1b[200~" {
			continue
		}
		// Skip keys that are not tested
		tests = append(tests, struct {
			input         string
			want          MsgKey
			expectedError error
		}{
			input:         seq,
			want:          MsgKey{Value: key, Unknown: false, IsSeq: true, ByteSeq: []byte(seq)},
			expectedError: nil,
		})

	}

	for _, test := range tests {
		reader := bufio.NewReader(strings.NewReader(test.input))
		msgKey, err := readKeyPressEvent(reader)
		if test.expectedError == nil {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		} else {
			if err == nil {
				t.Errorf("Expected error: %v", test.expectedError)
			} else if !errors.Is(err, test.expectedError) {
				t.Errorf("Expected error %v, but got %v", test.expectedError, err)
			}
		}
		if msgKey.Value != test.want.Value {
			t.Errorf("Expected value %q, but got %q", test.want.Value, msgKey.Value)
		}
		if msgKey.Unknown != test.want.Unknown {
			t.Errorf("Expected unknown %t, but got %t", test.want.Unknown, msgKey.Unknown)
		}
		if msgKey.IsSeq != test.want.IsSeq {
			t.Errorf("Expected isSeq %t, but got %t", test.want.IsSeq, msgKey.IsSeq)
		}
		if !bytes.Equal(msgKey.ByteSeq, test.want.ByteSeq) {
			t.Errorf("Expected byteSeq %#v, but got %#v", string(test.want.ByteSeq), string(msgKey.ByteSeq))
		}
	}
}
