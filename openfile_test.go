package loom

import (
	"errors"
	"os/exec"
	"reflect"
	"testing"
)

func TestFileOpenerInjectionAndFallback(t *testing.T) {
	var calls [][]string
	opener := FileOpener{
		GOOS: "linux",
		Start: func(name string, args ...string) error {
			call := append([]string{name}, args...)
			calls = append(calls, call)
			if name == "xdg-open" {
				return exec.ErrNotFound
			}
			return nil
		},
	}
	if err := opener.Open("a file.txt"); err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"xdg-open", "a file.txt"}, {"gio", "open", "a file.txt"}}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func TestFileOpenerDoesNotMaskLaunchError(t *testing.T) {
	wantErr := errors.New("launch failed")
	calls := 0
	opener := FileOpener{GOOS: "linux", Start: func(string, ...string) error {
		calls++
		return wantErr
	}}
	if err := opener.Open("file"); !errors.Is(err, wantErr) {
		t.Fatalf("Open() error = %v, want launch failure", err)
	}
	if calls != 1 {
		t.Fatalf("starter called %d times, want 1", calls)
	}
}

func TestFileOpenerRejectsUnsupportedPlatform(t *testing.T) {
	called := false
	opener := FileOpener{GOOS: "plan9", Start: func(string, ...string) error {
		called = true
		return nil
	}}
	if err := opener.Open("file"); err == nil {
		t.Fatal("Open() succeeded on an unsupported platform")
	}
	if called {
		t.Fatal("starter called on an unsupported platform")
	}
}
