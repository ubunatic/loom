// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"strings"
	"testing"

	"ubunatic.com/loom/loomtest"
)

func TestFmtIsIdempotentAndKeepsShownScreen(t *testing.T) {
	original := "\x1b[0m\x1b[38;5;1;3;1mred   \x1b[0m\x1b[0m\n\x1b[38;5;196mX\x1b[0m plain   \x1b[0m\n"
	path := writeFixture(t, original)
	var first bytes.Buffer
	if err := execute([]string{"fmt", path}, &first); err != nil {
		t.Fatalf("fmt: %v", err)
	}
	if first.String() == original {
		t.Fatal("fmt did not canonicalize the ANSI encoding")
	}
	if strings.Contains(first.String(), "plain   ") {
		t.Fatalf("fmt retained unstyled trailing spaces: %q", first.String())
	}
	if out, err := runDiff(t, original, first.String()); err != nil || out != "screens match\n" {
		t.Fatalf("shown screen changed: output=%q err=%v", out, err)
	}
	secondPath := writeFixture(t, first.String())
	var second bytes.Buffer
	if err := execute([]string{"fmt", secondPath}, &second); err != nil {
		t.Fatalf("second fmt: %v", err)
	}
	if second.String() != first.String() {
		t.Fatalf("fmt is not idempotent:\nfirst=%q\nsecond=%q", first.String(), second.String())
	}
}

func TestFmtPreservesAllSGRAttributes(t *testing.T) {
	original := "\x1b[3mA   \x1b[0m"
	path := writeFixture(t, original)
	var out bytes.Buffer
	if err := execute([]string{"fmt", path}, &out); err != nil {
		t.Fatalf("fmt: %v", err)
	}
	before, after := loomtest.NewVT(4, 1), loomtest.NewVT(4, 1)
	before.Write([]byte(original)) //nolint:errcheck
	after.Write(out.Bytes())       //nolint:errcheck
	for x := 0; x < 4; x++ {
		if before.Cell(x, 0) != after.Cell(x, 0) {
			t.Fatalf("shown cell changed at %d: before=%+v after=%+v", x, before.Cell(x, 0), after.Cell(x, 0))
		}
	}
}
