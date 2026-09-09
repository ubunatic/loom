// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestShowOnce(t *testing.T) {
	var out bytes.Buffer
	if err := run(&out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "\x1b") {
		t.Fatal("plain output contains terminal controls")
	}
	rows := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(rows) != 9 {
		t.Fatalf("got %d rows", len(rows))
	}
	for i, row := range rows {
		if len([]rune(row)) != 64 {
			t.Errorf("row %d is not 64 cells: %q", i, row)
		}
	}
	if !strings.Contains(rows[1], "All Usage") || !strings.Contains(rows[1], "Load") {
		t.Fatal("embedded declaration not rendered")
	}
}

type brokenWriter struct{ err error }

func (w brokenWriter) Write([]byte) (int, error) { return 0, w.err }

func TestOutputFailure(t *testing.T) {
	want := errors.New("closed output")
	if err := run(brokenWriter{want}); !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}
