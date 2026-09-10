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
	ID       string `yaml:"id"`
	Type     Type   `yaml:"type"`
	Path     string `yaml:"path"`
	Rate     string `yaml:"rate"`
	MaxBytes int64  `yaml:"max_bytes"`
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
	if s.MaxBytes < 0 {
		return fmt.Errorf("collector spec: max_bytes cannot be negative")
	}
	return nil
}

// Build creates the typed collector described by the declaration.
func (s Spec) Build() (Collector, time.Duration, error) {
	if err := s.Validate(); err != nil {
		return nil, 0, err
	}
	interval, _ := time.ParseDuration(s.Rate)
	return FileCollector{Path: s.Path, MaxBytes: s.MaxBytes}, interval, nil
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
