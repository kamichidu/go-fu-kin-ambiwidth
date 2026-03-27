package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	anyascii "github.com/anyascii/go"
	"github.com/kamichidu/go-fu-kin-ambiwidth/internal"
	"golang.org/x/text/width"
)

// AmbiwidthCharacters holds all Unicode ambiguous width characters.
// The format is a slice of structs, each containing the rune and its string representation.
var AmbiwidthCharacters []struct {
	R rune
	S string
}

func init() {
	for i := rune(0); i <= 0x10ffff; i++ {
		switch {
		case 0x0080 <= i && i <= 0x00ff: // Latin-1 Supplement
		case 0x2000 <= i && i <= 0x206f: // General Punctuation
		case 0x2100 <= i && i <= 0x214f: // Letterlike Symbols
		case 0x2190 <= i && i <= 0x21ff: // Arrows
		case 0x2200 <= i && i <= 0x22FF: // Mathematical Operators
		case 0x2500 <= i && i <= 0x257f: // Box Drawing
		case 0x2580 <= i && i <= 0x259f: // Block Elements
		case 0x25a0 <= i && i <= 0x25ff: // Geometric Shapes
		case 0x2700 <= i && i <= 0x27bf: // Dingbats
		default:
			continue
		}
		if width.LookupRune(i).Kind() == width.EastAsianAmbiguous {
			AmbiwidthCharacters = append(AmbiwidthCharacters, struct {
				R rune
				S string
			}{i, string(i)})
		}
	}
}

func main() {
	for _, line := range internal.MapFileHeaderLines {
		fmt.Printf("%s\n", line)
	}
	n := 0
	for _, c := range AmbiwidthCharacters {
		n = max(n, utf8.RuneCountInString(fmt.Sprintf("%U", c.R)))
	}
	for _, c := range AmbiwidthCharacters {
		a := anyascii.Transliterate(c.S)
		if a == "" {
			a = "."
		}

		// Try to find a better single character than just the first one if it's a parenthesis
		var r rune
		if len(a) > 1 && a[0] == '(' && a[len(a)-1] == ')' {
			// e.g., (R) -> R
			r, _ = utf8.DecodeRuneInString(a[1:])
		} else {
			r, _ = utf8.DecodeRuneInString(a)
		}

		m := utf8.RuneCountInString(fmt.Sprintf("%U", c.R))
		fmt.Printf("%U%s %c\n", c.R, strings.Repeat(" ", n-m), r)
	}
}
