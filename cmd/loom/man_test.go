// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManPrintsGeneratedRootPage(t *testing.T) {
	var out bytes.Buffer
	if err := execute([]string{"man"}, &out); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`.TH "LOOM" "1"`, "Loom Manual", "loom-diff(1)", "loom-man(1)", "loom-markup(1)"} {
		if !strings.Contains(out.String(), expected) {
			t.Errorf("man output missing %q", expected)
		}
	}
}

func TestManInstallGeneratesCommandTree(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "man1")
	var out bytes.Buffer
	if err := execute([]string{"man", "--install", "--dir", dir}, &out); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"loom.1", "loom-diff.1", "loom-man.1", "loom-markup.1"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("reading generated %s: %v", name, err)
			continue
		}
		if !strings.Contains(string(data), `.TH "LOOM"`) {
			t.Errorf("%s lacks Loom header", name)
		}
	}
	if !strings.Contains(out.String(), dir) || !strings.Contains(out.String(), "man loom") {
		t.Errorf("install message = %q", out.String())
	}
}

func TestManDirImpliesInstall(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "custom")
	if err := execute([]string{"man", "--dir", dir}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "loom.1")); err != nil {
		t.Fatalf("--dir did not install pages: %v", err)
	}
}

func TestDefaultManDirResolution(t *testing.T) {
	tests := []struct {
		name, want string
		euid       int
		xdg, home  string
	}{
		{name: "root", euid: 0, xdg: "/xdg", home: "/home/user", want: "/usr/local/share/man/man1"},
		{name: "XDG", euid: 1000, xdg: "/xdg", home: "/home/user", want: "/xdg/man/man1"},
		{name: "home fallback", euid: 1000, home: "/home/user", want: "/home/user/.local/share/man/man1"},
		{name: "system fallback", euid: 1000, want: "/usr/local/share/man/man1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultManDirFor(tt.euid, tt.xdg, tt.home); got != tt.want {
				t.Errorf("defaultManDirFor() = %q, want %q", got, tt.want)
			}
		})
	}
}
