package faw

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "simple ascii",
			input:   "hello world",
			want:    "hello world",
			wantErr: io.EOF,
		},
		{
			name:    "utf-8",
			input:   "こんにちは世界",
			want:    "こんにちは世界",
			wantErr: io.EOF,
		},
		{
			name:    "with newlines",
			input:   "line1\nline2\r\nline3",
			want:    "line1\nline2\r\nline3",
			wantErr: io.EOF,
		},
		{
			name:    "empty input",
			input:   "",
			want:    "",
			wantErr: io.EOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dst bytes.Buffer
			src := strings.NewReader(tt.input)
			err := Copy(&dst, src)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Copy() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := dst.String(); got != tt.want {
				t.Errorf("Copy() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopy_WriteError(t *testing.T) {
	writer := &errorWriter{err: errors.New("write error")}
	src := strings.NewReader("hello")
	err := Copy(writer, src)
	if err == nil || !strings.Contains(err.Error(), "write error") {
		t.Errorf("Copy() expected write error, got %v", err)
	}
}

type errorWriter struct {
	err error
}

func (w *errorWriter) Write(p []byte) (int, error) {
	return 0, w.err
}

func TestCopy_Flush(t *testing.T) {
	// Copy flushes on \n, \r, or when r.Buffered() == 0.
	// We can use a custom writer that record each Write call.
	var writeCalls []string
	writer := &recordWriter{
		onWrite: func(p []byte) {
			writeCalls = append(writeCalls, string(p))
		},
	}

	input := "abc\ndef"
	src := strings.NewReader(input)
	// Copy uses bufio.NewWriter(dst), which has a default buffer size (usually 4096).
	// It flushes on '\n'.
	err := Copy(writer, src)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Copy() error = %v, want %v", err, io.EOF)
	}

	// We expect at least one write call containing "abc\n" because it flushes on '\n'.
	foundNewlineFlush := false
	for _, call := range writeCalls {
		if strings.Contains(call, "abc\n") {
			foundNewlineFlush = true
			break
		}
	}
	if !foundNewlineFlush {
		t.Errorf("Expected flush on newline, but write calls were: %v", writeCalls)
	}
}

type recordWriter struct {
	onWrite func([]byte)
}

func (w *recordWriter) Write(p []byte) (int, error) {
	w.onWrite(p)
	return len(p), nil
}
