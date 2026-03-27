package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/creack/pty"
	faw "github.com/kamichidu/go-fu-kin-ambiwidth"
	"github.com/kamichidu/go-fu-kin-ambiwidth/assets"
	"github.com/kamichidu/go-fu-kin-ambiwidth/internal"
	"golang.org/x/term"
)

func run(stdin *os.File, stdout, stderr io.Writer, args []string) int {
	var (
		mapFile      = filepath.Join(internal.UserConfigDir(), "faw/faw.map")
		trackEnabled = false
	)
	flgs := flag.NewFlagSet("faw", flag.ContinueOnError)
	flgs.StringVar(&mapFile, "map-file", mapFile, "")
	flgs.BoolVar(&trackEnabled, "track", trackEnabled, "")
	flgs.SetOutput(io.Discard)
	if err := flgs.Parse(args[1:]); errors.Is(err, flag.ErrHelp) {
		flgs.SetOutput(stdout)
		flgs.Usage()
		return 0
	} else if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if flgs.NArg() == 0 {
		flgs.SetOutput(stderr)
		flgs.Usage()
		return 1
	}

	m, err := loadMap(mapFile)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if trackEnabled {
		m = faw.NewTracker(m, mapFile)
		defer func() {
			tracker := m.(*faw.Tracker)
			if err := tracker.Flush(); err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
			}
		}()
	}

	cmd := exec.Command(flgs.Arg(0), flgs.Args()[1:]...)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	defer ptmx.Close()

	oldState, err := term.MakeRaw(int(stdin.Fd()))
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	defer func() {
		if err := term.Restore(int(stdin.Fd()), oldState); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
		}
	}()

	// Handle pty size.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	defer signal.Stop(sigCh)
	go func() {
		for range sigCh {
			if err := pty.InheritSize(stdin, ptmx); err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
			}
		}
	}()
	// Initial resize.
	sigCh <- syscall.SIGWINCH

	// stdin
	go func() {
		if _, err := io.Copy(ptmx, stdin); err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Fprintf(stderr, "error: stdin io error: %v\n", err)
			}
		}
	}()
	// stdout
	if err := faw.Copy(stdout, faw.Wrap(ptmx, m)); err != nil {
		if !errors.Is(err, io.EOF) {
			fmt.Fprintf(stderr, "error: stdout io error: %v\n", err)
		}
	}
	err = cmd.Wait()
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	fmt.Fprintf(stderr, "error: command error: %v\n", err)
	return 1
}

func loadMap(name string) (faw.Registry, error) {
	m, err := faw.MapFromReader(bytes.NewReader(assets.DefaultMap))
	if err != nil {
		panic(err)
	}

	m2, err := faw.MapFromFile(name)
	if errors.Is(err, os.ErrNotExist) {
		return m, nil
	} else if err != nil {
		return nil, err
	}
	m.Merge(m2)
	return m, nil
}

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, os.Args))
}
