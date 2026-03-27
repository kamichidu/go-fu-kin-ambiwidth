package faw

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/kamichidu/go-fu-kin-ambiwidth/internal"
	"golang.org/x/text/width"
)

type Tracker struct {
	base Registry

	mu *sync.Mutex

	m map[rune]struct{}

	mapFileName string
}

func NewTracker(base Registry, mapFileName string) *Tracker {
	return &Tracker{
		base:        base,
		mu:          &sync.Mutex{},
		m:           make(map[rune]struct{}, 1024),
		mapFileName: mapFileName,
	}
}

func (t *Tracker) Lookup(ch rune) ([]rune, bool) {
	t.track(ch)
	alt, ok := t.base.Lookup(ch)
	return alt, ok
}

func (t *Tracker) track(ch rune) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.m[ch]; ok {
		return
	}
	if width.LookupRune(ch).Kind() == width.EastAsianAmbiguous {
		t.m[ch] = struct{}{}
	}

	if len(t.m) > 1000 {
		func() {
			t.mu.Unlock()
			defer t.mu.Lock()
			if err := t.Flush(); err != nil {
				// TODO: error handling
				panic(err)
			}
		}()
	}
}

func (t *Tracker) Flush() error {
	var tracked map[rune]struct{}
	func() {
		t.mu.Lock()
		defer t.mu.Unlock()
		tracked = make(map[rune]struct{}, len(t.m))
		for k, v := range t.m {
			tracked[k] = v
		}
		clear(t.m)
	}()
	if len(tracked) == 0 {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(t.mapFileName), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(t.mapFileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if stat, err := file.Stat(); err != nil {
		return err
	} else if stat.Size() == 0 {
		// write header lines on created
		for _, line := range internal.MapFileHeaderLines {
			fmt.Fprintf(file, "%s\n", line)
		}
	}

	m, err := MapFromReader(file)
	if err != nil {
		return err
	}
	l := make([]rune, 0, len(tracked))
	for ch := range tracked {
		if _, ok := m.Lookup(ch); ok {
			continue
		}
		l = append(l, ch)
	}
	if len(l) == 0 {
		return nil
	}
	sort.SliceStable(l, func(i, j int) bool {
		return l[i] < l[j]
	})
	if stat, err := file.Stat(); err != nil {
		return err
	} else if stat.Size() > 0 {
		var lastByte [1]byte
		n, err := file.ReadAt(lastByte[:], stat.Size()-1)
		if err != nil && err != io.EOF {
			return err
		}
		if n > 0 && lastByte[0] != '\n' {
			fmt.Fprint(file, "\n")
		}
	}
	for _, ch := range l {
		fmt.Fprintf(file, "%U  %c\n", ch, ch)
	}
	return nil
}
