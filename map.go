package faw

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Map struct {
	m map[rune][]rune

	err error
}

func mapOf(m map[rune][]rune) *Map {
	return &Map{
		m: m,
	}
}

func MapFromFile(name string) (*Map, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return MapFromReader(bytes.NewReader(data))
}

func MapFromReader(r io.Reader) (*Map, error) {
	m := &Map{}
	if err := m.load(r); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Map) Lookup(ch rune) ([]rune, bool) {
	v, ok := m.m[ch]
	return v, ok
}

func (m *Map) Merge(other *Map) {
	for k, v := range other.m {
		m.m[k] = v
	}
}

func (m *Map) load(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		ch, alt, ok := m.line(scanner.Text())
		if !ok {
			continue
		}
		if m.m == nil {
			m.m = map[rune][]rune{}
		}
		m.m[ch] = []rune(alt)
	}
	m.err = errors.Join(m.err, scanner.Err())
	return m.err
}

func (m *Map) line(line string) (rune, string, bool) {
	// ignore comment
	if l := strings.SplitN(line, ";;", 2); len(l) > 1 {
		line = l[0]
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return 0, "", false
	}

	var (
		cpStr string
		alt   string
	)
	if l := strings.SplitN(line, " ", 2); len(l) == 2 {
		cpStr, alt = l[0], l[1]
	} else {
		m.err = errors.Join(m.err, fmt.Errorf("invalid line: %v", line))
		return 0, "", false
	}

	// U+0000
	cpStr = strings.TrimSpace(cpStr)
	cpStr = strings.ToLower(cpStr)
	cpStr = strings.TrimPrefix(cpStr, "u+")
	cp, err := strconv.ParseInt(cpStr, 16, 64)
	if err != nil {
		m.err = errors.Join(m.err, fmt.Errorf("invalid line: %v", line))
		return 0, "", false
	}

	alt = strings.TrimSpace(alt)

	return rune(cp), alt, true
}
