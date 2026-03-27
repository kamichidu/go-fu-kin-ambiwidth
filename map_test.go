package faw

import (
	"reflect"
	"strings"
	"testing"
)

func TestMap_line(t *testing.T) {
	tests := []struct {
		line    string
		wantCh  rune
		wantAlt string
		wantOk  bool
	}{
		{"U+00A1          !", '\u00A1', "!", true},
		{"U+00A1 !!", '\u00A1', "!!", true},
		{"u+00a1          !", '\u00A1', "!", true},
		{"U+2026          .", '\u2026', ".", true},
		{"U+2500          -", '\u2500', "-", true},
		{"  U+00A1  !  ", '\u00A1', "!", true},
		{";; comment", 0, "", false},
		{"U+00A1 ! ;; comment", '\u00A1', "!", true},
		{"", 0, "", false},
		{"invalid", 0, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			var m Map
			gotCh, gotAlt, gotOk := m.line(tt.line)
			if gotCh != tt.wantCh {
				t.Errorf("line() gotCh = %v, want %v", gotCh, tt.wantCh)
			}
			if gotAlt != tt.wantAlt {
				t.Errorf("line() gotAlt = %v, want %v", gotAlt, tt.wantAlt)
			}
			if gotOk != tt.wantOk {
				t.Errorf("line() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestMap_load(t *testing.T) {
	input := `
;; comment
U+00A1 !
U+00A4 $
`
	m := &Map{}
	err := m.load(strings.NewReader(input))
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}
	want := mapOf(map[rune][]rune{
		'\u00A1': []rune("!"),
		'\u00A4': []rune("$"),
	})
	if !reflect.DeepEqual(m, want) {
		t.Errorf("load() got = %v, want %v", m.m, want)
	}
}

func TestMapFromFile(t *testing.T) {
	m, err := MapFromFile("assets/faw.map")
	if err != nil {
		t.Fatalf("MapFromFile() error = %v", err)
	}
	if len(m.m) == 0 {
		t.Error("MapFromFile() returned empty map")
	}

	// Verify some known values from assets/faw.map
	tests := []struct {
		ch   rune
		want string
	}{
		{'\u00A1', "!"},
		{'\u00A4', "$"},
		{'\u2026', "."},
		{'\u2500', "-"},
	}
	for _, tt := range tests {
		if got, ok := m.Lookup(tt.ch); !ok || string(got) != tt.want {
			t.Errorf("m[%U] = %s, want %s (ok=%v)", tt.ch, string(got), tt.want, ok)
		}
	}
}
