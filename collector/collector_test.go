// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package collector

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileCollectorReadsBoundedRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metrics")
	if err := os.WriteFile(path, []byte("cpu 42\n"), 0600); err != nil {
		t.Fatal(err)
	}
	at := time.Unix(123, 0)
	record, err := (FileCollector{Path: path, MaxBytes: 32}).Collect(context.Background(), at)
	if err != nil {
		t.Fatal(err)
	}
	if !record.At.Equal(at) || string(record.Data) != "cpu 42\n" {
		t.Fatalf("record = %+v", record)
	}
}

func TestFileCollectorRejectsOversizedAndCancelledReads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metrics")
	if err := os.WriteFile(path, []byte("0123456789"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := (FileCollector{Path: path, MaxBytes: 4}).Collect(context.Background(), time.Time{}); err == nil {
		t.Fatal("accepted oversized record")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (FileCollector{Path: path}).Collect(ctx, time.Time{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled read error = %v", err)
	}
}

func TestSpecBuildsTypedFileCollector(t *testing.T) {
	spec := Spec{ID: "cpu", Type: TypeFile, Path: "/proc/stat", Rate: "1s"}
	c, interval, retention, err := spec.Build()
	if err != nil {
		t.Fatal(err)
	}
	if c.Type() != TypeFile || interval != time.Second || retention != DefaultRetention {
		t.Fatalf("collector = %v, interval = %s, retention = %s", c.Type(), interval, retention)
	}
	for _, bad := range []Spec{
		{ID: "cpu", Type: "socket", Path: "/proc/stat", Rate: "1s"},
		{ID: "cpu", Type: TypeFile, Path: "/proc/stat", Rate: "0s"},
		{ID: "cpu", Type: TypeFile, Path: "/proc/stat", Rate: "1s", MaxBytes: -1},
	} {
		if _, _, _, err := bad.Build(); err == nil {
			t.Fatalf("accepted invalid spec %+v", bad)
		}
	}
}

func TestHistoryRetainsOnlyLiveWindow(t *testing.T) {
	history, err := NewHistory(15 * time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	base := time.Unix(1000, 0)
	history.Append(Record{At: base, Data: []byte("old")})
	history.Append(Record{At: base.Add(14 * time.Minute), Data: []byte("live")})
	history.Append(Record{At: base.Add(15*time.Minute + time.Nanosecond), Data: []byte("new")})
	records := history.Snapshot()
	if len(records) != 2 || string(records[0].Data) != "live" || string(records[1].Data) != "new" {
		t.Fatalf("retained records = %+v", records)
	}
	records[0].Data[0] = 'X'
	if string(history.Snapshot()[0].Data) != "live" {
		t.Fatal("history snapshot shares record data")
	}
}
