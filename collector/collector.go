// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package collector defines the small boundary between changing data and
// rendering. Collectors return records; callers decide how records become
// snapshots and widgets.
package collector

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// Type identifies a configured collector implementation.
type Type string

const (
	// TypeFile reads a bounded file at the caller's cadence.
	TypeFile Type = "file"
	// DefaultRetention is the in-memory history window when a spec omits one.
	DefaultRetention = 15 * time.Minute
)

// Record is one immutable collection result. Data is owned by the record and
// must be copied before a caller retains or transforms it asynchronously.
type Record struct {
	At   time.Time
	Data []byte
}

// Collector reads one source without knowing anything about widgets or
// rendering.
type Collector interface {
	Type() Type
	Collect(context.Context, time.Time) (Record, error)
}

// Spec is the minimal declaration consumed by a file collector prototype.
// Rate is a duration string such as "1s"; source-specific parsing remains
// outside this package.
type Spec struct {
	ID        string `yaml:"id"`
	Type      Type   `yaml:"type"`
	Path      string `yaml:"path"`
	Rate      string `yaml:"rate"`
	Retention string `yaml:"retention"`
	MaxBytes  int64  `yaml:"max_bytes"`
}

// Validate checks the bounded prototype declaration.
func (s Spec) Validate() error {
	if s.ID == "" || s.Type == "" || s.Path == "" || s.Rate == "" {
		return fmt.Errorf("collector spec: id, type, path and rate are required")
	}
	if s.Type != TypeFile {
		return fmt.Errorf("collector spec: unsupported type %q", s.Type)
	}
	interval, err := time.ParseDuration(s.Rate)
	if err != nil || interval <= 0 {
		return fmt.Errorf("collector spec: rate must be a positive duration")
	}
	if _, err := s.RetentionDuration(); err != nil {
		return err
	}
	if s.MaxBytes < 0 {
		return fmt.Errorf("collector spec: max_bytes cannot be negative")
	}
	return nil
}

// Build creates the typed collector described by the declaration.
func (s Spec) Build() (Collector, time.Duration, time.Duration, error) {
	if err := s.Validate(); err != nil {
		return nil, 0, 0, err
	}
	interval, _ := time.ParseDuration(s.Rate)
	retention, _ := s.RetentionDuration()
	return FileCollector{Path: s.Path, MaxBytes: s.MaxBytes}, interval, retention, nil
}

// RetentionDuration resolves the in-memory history window. An omitted value
// uses DefaultRetention; no collector state is persisted to disk.
func (s Spec) RetentionDuration() (time.Duration, error) {
	if s.Retention == "" {
		return DefaultRetention, nil
	}
	retention, err := time.ParseDuration(s.Retention)
	if err != nil || retention <= 0 {
		return 0, fmt.Errorf("collector spec: retention must be a positive duration")
	}
	return retention, nil
}

// History retains timestamped records only within a live in-memory window.
type History struct {
	retention time.Duration
	records   []Record
}

// NewHistory creates a live history with the requested retention window.
func NewHistory(retention time.Duration) (*History, error) {
	if retention <= 0 {
		return nil, fmt.Errorf("collector history: retention must be positive")
	}
	return &History{retention: retention}, nil
}

// Append publishes a record and removes samples older than the retention
// window relative to the appended record timestamp.
func (h *History) Append(record Record) {
	if h == nil {
		return
	}
	record.Data = append([]byte(nil), record.Data...)
	h.records = append(h.records, record)
	cutoff := record.At.Add(-h.retention)
	first := 0
	for first < len(h.records) && h.records[first].At.Before(cutoff) {
		first++
	}
	if first > 0 {
		h.records = h.records[first:]
	}
}

// Snapshot returns a copy of the live records in chronological order.
func (h *History) Snapshot() []Record {
	if h == nil {
		return nil
	}
	result := make([]Record, len(h.records))
	for i, record := range h.records {
		result[i] = Record{At: record.At, Data: append([]byte(nil), record.Data...)}
	}
	return result
}

// Run collects immediately and then at interval until cancellation. The sink
// receives records independently of rendering and owns snapshot publication.
func Run(ctx context.Context, c Collector, interval time.Duration, sink func(Record) error) error {
	if c == nil || sink == nil {
		return fmt.Errorf("collector: collector and sink are required")
	}
	if interval <= 0 {
		return fmt.Errorf("collector: interval must be positive")
	}
	collect := func(at time.Time) error {
		record, err := c.Collect(ctx, at)
		if err != nil {
			return err
		}
		return sink(record)
	}
	if err := collect(time.Now()); err != nil {
		return err
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case at := <-ticker.C:
			if err := collect(at); err != nil {
				return err
			}
		}
	}
}

// FileCollector reads a file, rejecting records larger than MaxBytes.
type FileCollector struct {
	Path     string
	MaxBytes int64
}

// Type identifies this collector as a file source.
func (FileCollector) Type() Type { return TypeFile }

// Collect reads one bounded file record. The timestamp is supplied by the
// scheduler so collection timing remains testable and independent of Draw.
func (c FileCollector) Collect(ctx context.Context, at time.Time) (Record, error) {
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	if c.Path == "" {
		return Record{}, fmt.Errorf("file collector: path is required")
	}
	limit := c.MaxBytes
	if limit <= 0 {
		limit = 1 << 20
	}
	f, err := os.Open(c.Path)
	if err != nil {
		return Record{}, fmt.Errorf("file collector: open %q: %w", c.Path, err)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return Record{}, fmt.Errorf("file collector: read %q: %w", c.Path, err)
	}
	if int64(len(data)) > limit {
		return Record{}, fmt.Errorf("file collector: %q exceeds %d bytes", c.Path, limit)
	}
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	return Record{At: at, Data: data}, nil
}
