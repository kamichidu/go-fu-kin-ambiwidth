package faw

import (
	"bufio"
	"io"
)

func Copy(dst io.Writer, src io.Reader) error {
	var r *bufio.Reader
	if v, ok := src.(*bufio.Reader); ok {
		r = v
	} else {
		r = bufio.NewReader(src)
	}
	var w *bufio.Writer
	if v, ok := dst.(*bufio.Writer); ok {
		w = v
	} else {
		w = bufio.NewWriter(dst)
	}
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return err
		}
		if _, err := w.WriteRune(ch); err != nil {
			return err
		}
		// flush if needed
		switch {
		case r.Buffered() == 0:
			fallthrough
		case ch == '\r':
			fallthrough
		case ch == '\n':
			if err := w.Flush(); err != nil {
				return err
			}
		}
	}
}
