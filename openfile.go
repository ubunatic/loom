package loom

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
)

// CommandStarter starts a platform command and returns once it has launched.
type CommandStarter func(name string, args ...string) error

// FileOpener launches files using the desktop opener for its platform.
// Start and GOOS may be supplied by callers to make launching deterministic in tests.
type FileOpener struct {
	GOOS  string
	Start CommandStarter
}

// OpenFile launches path with the current platform's desktop file opener.
func OpenFile(path string) error {
	return (FileOpener{}).Open(path)
}

// Open launches path, falling back to another platform opener only when a
// candidate executable is unavailable.
func (o FileOpener) Open(path string) error {
	goos := o.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	start := o.Start
	if start == nil {
		start = startDetached
	}
	candidates, err := openerCommands(goos, path)
	if err != nil {
		return err
	}
	for i, command := range candidates {
		err = start(command[0], command[1:]...)
		if err == nil {
			return nil
		}
		if !errors.Is(err, exec.ErrNotFound) || i == len(candidates)-1 {
			return fmt.Errorf("open file %q: %w", path, err)
		}
	}
	return nil
}

func openerCommands(goos, path string) ([][]string, error) {
	switch goos {
	case "darwin":
		return [][]string{{"open", path}}, nil
	case "linux", "freebsd", "openbsd", "netbsd", "dragonfly":
		return [][]string{{"xdg-open", path}, {"gio", "open", path}}, nil
	case "windows":
		return [][]string{{"rundll32", "url.dll,FileProtocolHandler", path}}, nil
	default:
		return nil, fmt.Errorf("open file %q: unsupported platform %q", path, goos)
	}
}

func startDetached(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = nil
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
