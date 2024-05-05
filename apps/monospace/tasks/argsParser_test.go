/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package tasks

import (
	"reflect"
	"testing"
)

type argsParserTest struct {
	name             string
	line             string
	want             []string
	strictShouldFail error
}

func testsCases() []argsParserTest {
	return []argsParserTest{
		{
			name: "with quoted spaces",
			line: `ls "/tmp/dir/with space" '/tmp/other dir/with space'`,
			want: []string{"ls", "/tmp/dir/with space", "/tmp/other dir/with space"},
		},
		{
			name: "with inverted quote",
			line: `ls "/tmp/dir/'withQuote'" '/tmp/dir/"withDoubleQuote"'`,
			want: []string{"ls", `/tmp/dir/'withQuote'`, `/tmp/dir/"withDoubleQuote"`},
		},
		{
			name: "with embedded quote",
			line: `ls '/tmp/dir/\'withQuote\'' "/tmp/dir/\"withDoubleQuote\""`,
			want: []string{"ls", `/tmp/dir/'withQuote'`, `/tmp/dir/"withDoubleQuote"`},
		},
		{
			name: "No quotes",
			line: "test.sh -f arg1 -key=value",
			want: []string{"test.sh", "-f", "arg1", "-key=value"},
		},
		{
			name: "With quotes",
			line: `testWithQuote.sh -f arg1 -key=value -f="space string"`,
			want: []string{"testWithQuote.sh", "-f", "arg1", "-key=value", `-f="space string"`},
		},
		{
			name: "Single quotes",
			line: `testSingle.sh -f 'arg 1' -key='value' -f='tata arg1'`,
			want: []string{"testSingle.sh", "-f", "arg 1", "-key='value'", "-f='tata arg1'"},
		},
		{
			name: "Mixed quotes",
			line: `testMixed.sh -f 'arg 1' -key="value" -f="tata arg1"`,
			want: []string{"testMixed.sh", "-f", "arg 1", `-key="value"`, `-f="tata arg1"`},
		},
		{
			name: "Embedded quotes",
			line: `testEmbed.sh -f="tata 'arg1'" x="a 'b '"`,
			want: []string{"testEmbed.sh", `-f="tata 'arg1'"`, `x="a 'b '"`},
		},
		{
			name: "Embedded quotes of same type",
			line: `testEmbed2.sh 'single \'embedded\' quote' "double \"embedded\" quote" -f="tata \"arg1\"" x='a \'b \''`,
			want: []string{"testEmbed2.sh", "single 'embedded' quote", "double \"embedded\" quote", `-f="tata \"arg1\""`, `x='a \'b \''`},
		},
		{
			name:             "Unterminated quote",
			line:             `unterminated.sh args "unterminated`,
			want:             []string{"unterminated.sh", "args", "unterminated"},
			strictShouldFail: ErrParseQuoteError,
		},
		{
			name:             "Unterminated quote in arg",
			line:             `unterminated.sh args"unterminated`,
			want:             []string{"unterminated.sh", "args\"unterminated"},
			strictShouldFail: ErrParseQuoteError,
		},
		{
			name:             "Unterminated quote at end",
			line:             `unterminated.sh args"`,
			want:             []string{"unterminated.sh", "args\""},
			strictShouldFail: ErrParseQuoteError,
		},
		{
			name: "Ingnore leading and trailing spaces",
			line: "\t unterminated.sh args -x \t\n ",
			want: []string{"unterminated.sh", "args", "-x"},
		},
	}
}

func TestParseArgs(t *testing.T) {
	for _, tt := range testsCases() {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseArgs(tt.line)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseArgs(%#v) = %#v, want %#v", tt.line, got, tt.want)
			}
		})
	}
}

func TestParseArgsStrict(t *testing.T) {
	for _, tt := range testsCases() {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArgsStrict(tt.line)
			if err != tt.strictShouldFail {
				if err == nil {
					t.Errorf("ParseArgs(%#v) should have failed but didn't", tt.line)
				} else if tt.strictShouldFail == nil {
					t.Errorf("ParseArgs(%#v) failed with error: %s", tt.line, err)
				} else {
					t.Errorf("ParseArgs(%#v) failed with error: %s, expected error: %s", tt.line, err, tt.strictShouldFail)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseArgs(%#v) = %#v, want %#v", tt.line, got, tt.want)
			}
		})
	}
}
