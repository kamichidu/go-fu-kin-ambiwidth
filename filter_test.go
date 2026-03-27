package faw

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestWrap(t *testing.T) {
	m := mapOf(map[rune][]rune{
		'\u00A1': []rune("!"),
		'\u2026': []rune("..."),
	})

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no mapping",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "with mapping",
			input: "h\u00A1llo\u2026",
			want:  "h!llo...",
		},
		{
			name:  "ansi escape sequence CSI",
			input: "\x1b[31mred\x1b[0m",
			want:  "\x1b[31mred\x1b[0m",
		},
		{
			name:  "ansi escape sequence OSC",
			input: "\x1b]0;title\x07",
			want:  "\x1b]0;title\x07",
		},
		{
			name:  "ansi escape sequence OSC with ST",
			input: "\x1b]0;title\x1b\\",
			want:  "\x1b]0;title\x1b\\",
		},
		{
			name:  "mixed mapping and escape sequences",
			input: "\x1b[31m\u00A1\x1b[0m",
			want:  "\x1b[31m!\x1b[0m",
		},
		{
			name:  "incomplete escape sequence at end",
			input: "\x1b[31",
			want:  "\x1b[31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Wrap(strings.NewReader(tt.input), m)
			var got bytes.Buffer
			_, err := io.Copy(&got, r)
			if err != nil && err != io.EOF {
				t.Fatalf("io.Copy error: %v", err)
			}
			if got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestFilter_ReadTooSmallBuffer(t *testing.T) {
	m := mapOf(nil)
	input := "abc"
	r := Wrap(strings.NewReader(input), m)

	p := make([]byte, utf8.UTFMax-1)
	_, err := r.Read(p)
	if err != io.ErrShortBuffer {
		t.Errorf("got error %v, want %v", err, io.ErrShortBuffer)
	}
}

func TestFilter_ReadSmallBuffer(t *testing.T) {
	m := mapOf(map[rune][]rune{
		'\u00A1': []rune("!!"),
	})
	input := "a\u00A1b"
	r := Wrap(strings.NewReader(input), m)

	// Read with utf8.UTFMax-byte buffer to force multiple reads and buffering
	p := make([]byte, utf8.UTFMax)
	var got bytes.Buffer
	for {
		n, err := r.Read(p)
		if n > 0 {
			got.Write(p[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("Read error: %v", err)
		}
	}

	want := "a!!b"
	if got.String() != want {
		t.Errorf("got %q, want %q", got.String(), want)
	}
}
