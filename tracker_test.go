package faw

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/text/width"
)

type mockRegistry struct {
	m map[rune][]rune
}

func (m *mockRegistry) Lookup(ch rune) ([]rune, bool) {
	alt, ok := m.m[ch]
	return alt, ok
}

func TestNewTracker(t *testing.T) {
	reg := &mockRegistry{m: make(map[rune][]rune)}
	mapFile := "test.map"
	tr := NewTracker(reg, mapFile)
	if tr == nil {
		t.Fatal("NewTracker returned nil")
	}
	if tr.base != reg {
		t.Errorf("NewTracker base mismatch")
	}
	if tr.m == nil {
		t.Error("NewTracker map not initialized")
	}
	if tr.mapFileName != mapFile {
		t.Errorf("NewTracker mapFileName mismatch")
	}
}

func TestTracker_Lookup(t *testing.T) {
	reg := &mockRegistry{
		m: map[rune][]rune{
			'\u00A1': []rune("!"), // Ambiguous
			'A':      []rune("A"), // Not Ambiguous
		},
	}
	tr := NewTracker(reg, "test.map")

	// Test Ambiguous character
	got, ok := tr.Lookup('\u00A1')
	if !ok || string(got) != "!" {
		t.Errorf("Lookup('\u00A1') = %s, %v; want \"!\", true", string(got), ok)
	}
	if _, tracked := tr.m['\u00A1']; !tracked {
		t.Errorf("'\u00A1' should be tracked")
	}

	// Test Non-Ambiguous character
	got, ok = tr.Lookup('A')
	if !ok || string(got) != "A" {
		t.Errorf("Lookup('A') = %s, %v; want \"A\", true", string(got), ok)
	}
	if _, tracked := tr.m['A']; tracked {
		t.Errorf("'A' should not be tracked")
	}

	// Test character not in registry
	got, ok = tr.Lookup('\u2026') // Ambiguous, but not in registry
	if ok || got != nil {
		t.Errorf("Lookup('\u2026') = %v, %v; want nil, false", got, ok)
	}
	if _, tracked := tr.m['\u2026']; !tracked {
		t.Errorf("'\u2026' should be tracked even if not in registry")
	}
}

func TestTracker_Flush_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	mapFile := filepath.Join(tmpDir, "new.map")

	reg := &mockRegistry{m: make(map[rune][]rune)}
	tr := NewTracker(reg, mapFile)

	// '\u00A1' is Ambiguous
	tr.Lookup('\u00A1')

	if err := tr.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	content, err := os.ReadFile(mapFile)
	if err != nil {
		t.Fatalf("Failed to read flushed file: %v", err)
	}

	sContent := string(content)
	if !strings.Contains(sContent, ";; faw.map") {
		t.Error("Flushed file missing header")
	}
	if !strings.Contains(sContent, "U+00A1") {
		t.Errorf("Flushed file missing tracked rune: %s", sContent)
	}
}

func TestTracker_Flush_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	mapFile := filepath.Join(tmpDir, "empty.map")

	reg := &mockRegistry{m: make(map[rune][]rune)}
	tr := NewTracker(reg, mapFile)

	if err := tr.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if _, err := os.Stat(mapFile); !os.IsNotExist(err) {
		t.Errorf("Empty flush should not create map file: %v", err)
	}
}

func TestTracker_Flush_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	mapFile := filepath.Join(tmpDir, "existing.map")

	initialContent := ";; header\nU+00A1  \u00A1\n"
	if err := os.WriteFile(mapFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}

	reg := &mockRegistry{m: make(map[rune][]rune)}
	tr := NewTracker(reg, mapFile)

	// '\u00A1' is already in file, '\u2026' is not
	tr.Lookup('\u00A1')
	tr.Lookup('\u2026')

	if err := tr.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	content, err := os.ReadFile(mapFile)
	if err != nil {
		t.Fatalf("Failed to read flushed file: %v", err)
	}

	sContent := string(content)
	if !strings.Contains(sContent, "U+2026") {
		t.Errorf("Flushed file missing new tracked rune: %s", sContent)
	}
}

func TestTracker_Flush_NoNewlineAtEnd(t *testing.T) {
	tmpDir := t.TempDir()
	mapFile := filepath.Join(tmpDir, "nonl.map")

	initialContent := ";; header\nU+00A1  \u00A1" // No newline at end
	if err := os.WriteFile(mapFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}

	reg := &mockRegistry{m: make(map[rune][]rune)}
	tr := NewTracker(reg, mapFile)

	tr.Lookup('\u2026')

	if err := tr.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	content, err := os.ReadFile(mapFile)
	if err != nil {
		t.Fatalf("Failed to read flushed file: %v", err)
	}

	sContent := string(content)
	if !strings.HasPrefix(sContent, initialContent) {
		t.Errorf("Initial content was modified: %s", sContent)
	}

	// Check if U+2026 is on a new line
	if !strings.Contains(sContent, "\nU+2026") {
		t.Errorf("New entry should be on a new line: %q", sContent)
	}
}

func TestTracker_Flush_WithNewlineAtEnd(t *testing.T) {
	tmpDir := t.TempDir()
	mapFile := filepath.Join(tmpDir, "withnl.map")

	initialContent := ";; header\nU+00A1  \u00A1\n" // With newline at end
	if err := os.WriteFile(mapFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}

	reg := &mockRegistry{m: make(map[rune][]rune)}
	tr := NewTracker(reg, mapFile)

	tr.Lookup('\u2026')

	if err := tr.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	content, err := os.ReadFile(mapFile)
	if err != nil {
		t.Fatalf("Failed to read flushed file: %v", err)
	}

	sContent := string(content)
	// Should not contain double newline
	if strings.Contains(sContent, "\n\n") {
		t.Errorf("Should not contain empty lines: %q", sContent)
	}
}

func TestTracker_TrackAmbiguous(t *testing.T) {
	tr := NewTracker(nil, "test.map")

	// Check some runes
	tests := []struct {
		ch   rune
		want bool
	}{
		{'\u00A1', true},  // Inverted Exclamation Mark - Ambiguous
		{'A', false},      // Latin Capital Letter A - Narrow
		{'\u4E00', false}, // CJK Unified Ideograph - Wide
		{'\u2026', true},  // Horizontal Ellipsis - Ambiguous
	}

	for _, tt := range tests {
		tr.track(tt.ch)
		_, ok := tr.m[tt.ch]
		if ok != tt.want {
			t.Errorf("track(%U) tracked = %v, want %v (kind=%v)", tt.ch, ok, tt.want, width.LookupRune(tt.ch).Kind())
		}
	}
}

func BenchmarkTracker_Lookup_Narrow(b *testing.B) {
	reg := &mockRegistry{m: make(map[rune][]rune)}
	tr := NewTracker(reg, filepath.Join(b.TempDir(), "test.map"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Lookup('A')
	}
}

func BenchmarkTracker_Lookup_Ambiguous_AlreadyTracked(b *testing.B) {
	reg := &mockRegistry{m: make(map[rune][]rune)}
	tr := NewTracker(reg, filepath.Join(b.TempDir(), "test.map"))
	tr.Lookup('\u00A1')
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Lookup('\u00A1')
	}
}

func BenchmarkTracker_Lookup_Ambiguous_NewRune(b *testing.B) {
	reg := &mockRegistry{m: make(map[rune][]rune)}
	tmpDir := b.TempDir()
	mapFile := filepath.Join(tmpDir, "test.map")
	tr := NewTracker(reg, mapFile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Lookup(rune(0x00A1 + (i % 1000)))
	}
}

func BenchmarkTracker_Lookup_Wide(b *testing.B) {
	reg := &mockRegistry{m: make(map[rune][]rune)}
	tr := NewTracker(reg, filepath.Join(b.TempDir(), "test.map"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Lookup('\u4E00')
	}
}

func BenchmarkTracker_Flush_Small(b *testing.B) {
	tmpDir := b.TempDir()
	mapFile := filepath.Join(tmpDir, "test.map")
	tr := NewTracker(nil, mapFile)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for j := 0; j < 10; j++ {
			// Using Ambiguous characters
			tr.m[rune(0x00A1+j)] = struct{}{}
		}
		os.Remove(mapFile)
		b.StartTimer()

		if err := tr.Flush(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTracker_Flush_Large(b *testing.B) {
	tmpDir := b.TempDir()
	mapFile := filepath.Join(tmpDir, "test.map")
	tr := NewTracker(nil, mapFile)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for j := 0; j < 1000; j++ {
			// Using characters likely to be safe for Map.line
			tr.m[rune(0x4E00+j)] = struct{}{}
		}
		os.Remove(mapFile)
		b.StartTimer()

		if err := tr.Flush(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTracker_Flush_ExistingFile(b *testing.B) {
	tmpDir := b.TempDir()
	mapFile := filepath.Join(tmpDir, "test.map")
	tr := NewTracker(nil, mapFile)

	// Pre-fill file with 1000 entries (CJK range is safe)
	for j := 0; j < 1000; j++ {
		tr.m[rune(0x4E00+j)] = struct{}{}
	}
	tr.Flush()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for j := 0; j < 10; j++ {
			// Add some new ones
			tr.m[rune(0x5E00+j)] = struct{}{}
		}
		b.StartTimer()

		if err := tr.Flush(); err != nil {
			b.Fatal(err)
		}
	}
}
