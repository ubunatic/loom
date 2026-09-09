// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"context"
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

func TestCommand(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
		fail bool
	}{
		{"help", []string{"-h"}, "--watch", false},
		{"once", nil, "All Usage", false},
		{"unknown", []string{"--watc"}, "", true},
		{"argument", []string{"extra"}, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			err := execute(context.Background(), tc.args, &out)
			if (err != nil) != tc.fail {
				t.Fatalf("got %v", err)
			}
			if !strings.Contains(out.String(), tc.want) {
				t.Fatal(out.String())
			}
		})
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
