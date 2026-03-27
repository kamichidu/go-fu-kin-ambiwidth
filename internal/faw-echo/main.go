package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/width"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("〓〓〓 FAW-Echo Interactive 〓〓〓")
	fmt.Println("※ This tool uses many East Asian Ambiguous width characters.")
	fmt.Println("They may appear broken if your terminal and environment disagree on their width.")
	fmt.Println("Enter text to decorate (empty line to quit):")

	for {
		fmt.Print("\n≫ ")
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()
		if text == "" {
			break
		}

		// calculateCellWidth computes the visual width in cells.
		// We'll treat ambiguous characters as 1-cell here because if they are 
		// treated as 2-cells in the terminal, the border characters (also 
		// ambiguous) will also be 2-cells, and the alignment will be preserved 
		// as long as they are consistent.
		w := calculateCellWidth(text)

		// Border 1: Heavy Box Drawings (Ambiguous)
		// ┏ (U+250F), ━ (U+2501), ┓ (U+2513), ┃ (U+2503), ┛ (U+251B), ┗ (U+2517)
		h1 := strings.Repeat("━", w+2)
		fmt.Printf("┏%s┓\n", h1)
		fmt.Printf("┃ %s ┃\n", text)
		fmt.Printf("┗%s┛\n", h1)

		// Border 2: Double Box Drawings (Ambiguous)
		// ╔ (U+2554), ═ (U+2550), ╗ (U+2557), ║ (U+2551), ╝ (U+255D), ╚ (U+255A)
		h2 := strings.Repeat("═", w+2)
		fmt.Printf("╔%s╗\n", h2)
		fmt.Printf("║ %s ║\n", text)
		fmt.Printf("╚%s╝\n", h2)

		// Type 3: Symbol Decoration (Ambiguous)
		// ★ (U+2605), ◆ (U+25C6), ■ (U+25A0)
		fmt.Printf("★ %s ★\n", text)
		fmt.Printf("◆ %s ◆\n", text)
		fmt.Printf("■ %s ■\n", text)
	}
	fmt.Println("\nBye! (※)")
}

func calculateCellWidth(s string) int {
	w := 0
	for _, r := range s {
		k := width.LookupRune(r).Kind()
		switch k {
		case width.EastAsianWide, width.EastAsianFullwidth:
			w += 2
		case width.EastAsianAmbiguous:
			// Treat as 1-cell here. If the terminal treats it as 2-cells, 
			// it will also treat our border characters as 2-cells.
			w += 1
		default:
			w += 1
		}
	}
	return w
}
