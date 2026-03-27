package faw

import (
	"bufio"
	"io"
	"unicode/utf8"
)

type Registry interface {
	Lookup(rune) ([]rune, bool)
}

type filter struct {
	r *bufio.Reader

	idx int

	buffer []rune

	backBuffer []rune

	lastErr error

	reg Registry

	state stateFn
}

func Wrap(src io.Reader, reg Registry) io.Reader {
	f := &filter{
		reg: reg,
	}
	if r, ok := src.(*bufio.Reader); ok {
		f.r = r
	} else {
		f.r = bufio.NewReader(src)
	}
	f.buffer = make([]rune, 0, 1024)
	f.backBuffer = make([]rune, 0, 64)
	f.state = stateDefault
	return f
}

func (f *filter) Read(p []byte) (int, error) {
	if len(p) < utf8.UTFMax {
		return 0, io.ErrShortBuffer
	}
	var n int
	for {
		if len(p)-n < utf8.UTFMax {
			return n, nil
		}
		if f.idx < len(f.buffer) {
			c := f.buffer[f.idx]
			f.idx++
			n += utf8.EncodeRune(p[n:], c)
			if f.idx == len(f.buffer) {
				f.buffer = f.buffer[:0]
				f.idx = 0
			}
			continue
		}
		if n > 0 && f.r.Buffered() == 0 {
			return n, nil
		}
		if f.lastErr != nil {
			return n, f.lastErr
		}
		c, _, err := f.r.ReadRune()
		if err != nil {
			if len(f.backBuffer) > 0 {
				f.buffer = append(f.buffer, f.backBuffer...)
				f.backBuffer = f.backBuffer[:0]
			}
			if len(f.buffer) == 0 {
				return n, err
			} else {
				f.lastErr = err
				continue
			}
		}
		f.state = f.state(f, c)
	}
}

type stateFn func(*filter, rune) stateFn

func stateDefault(f *filter, ch rune) stateFn {
	switch ch {
	case '\x1b':
		f.backBuffer = f.backBuffer[:0]
		f.backBuffer = append(f.backBuffer, ch)
		return stateEsc
	}
	if alt, ok := f.reg.Lookup(ch); ok {
		f.buffer = append(f.buffer, alt...)
		return stateDefault
	}
	f.buffer = append(f.buffer, ch)
	return stateDefault
}

func stateEsc(f *filter, ch rune) stateFn {
	switch ch {
	case '[': // csi
		f.backBuffer = append(f.backBuffer, ch)
		return stateCSI
	case ']': // osc
		f.backBuffer = append(f.backBuffer, ch)
		return stateOSC
	default: // single esc
		f.backBuffer = append(f.backBuffer, ch)
		f.buffer = append(f.buffer, f.backBuffer...)
		f.backBuffer = f.backBuffer[:0]
		return stateDefault
	}
}

func stateCSI(f *filter, ch rune) stateFn {
	if ch >= 0x40 && ch <= 0x7e {
		f.backBuffer = append(f.backBuffer, ch)
		f.buffer = append(f.buffer, f.backBuffer...)
		f.backBuffer = f.backBuffer[:0]
		return stateDefault
	}
	f.backBuffer = append(f.backBuffer, ch)
	return stateCSI
}

func stateOSC(f *filter, ch rune) stateFn {
	// bell
	if ch == 0x07 {
		f.backBuffer = append(f.backBuffer, ch)
		f.buffer = append(f.buffer, f.backBuffer...)
		f.backBuffer = f.backBuffer[:0]
		return stateDefault
	}
	// esc backslash
	if ch == '\\' && len(f.backBuffer) > 0 && f.backBuffer[len(f.backBuffer)-1] == '\x1b' {
		f.backBuffer = append(f.backBuffer, ch)
		f.buffer = append(f.buffer, f.backBuffer...)
		f.backBuffer = f.backBuffer[:0]
		return stateDefault
	}
	f.backBuffer = append(f.backBuffer, ch)
	return stateOSC
}
