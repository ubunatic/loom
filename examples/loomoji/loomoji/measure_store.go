// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomoji

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"codeberg.org/ubunatic/loom/measure"
)

// TerminalProfile identifies the terminal whose rendered widths were recorded.
type TerminalProfile struct {
	Term        string `json:"term"`
	TermProgram string `json:"term_program,omitempty"`
}

// Measurement contains the observed and computed width of one glyph.
// MeasuredWidth is 1 or 2 for a confirmed result and 0 for unsure/other.
type Measurement struct {
	Glyph         string   `json:"glyph"`
	Codepoints    []string `json:"codepoints"`
	ComputedWidth int      `json:"computed_width"`
	MeasuredWidth int      `json:"measured_width"`
	Comment       string   `json:"comment,omitempty"`
}

// NewMeasurement creates an unanswered record using Loom's current width policy.
func NewMeasurement(glyph string) Measurement {
	codepoints := make([]string, 0, len([]rune(glyph)))
	for _, r := range glyph {
		codepoints = append(codepoints, fmt.Sprintf("U+%04X", r))
	}
	return Measurement{Glyph: glyph, Codepoints: codepoints, ComputedWidth: measure.StringWidth(glyph)}
}

// MeasurementStore holds observed widths for one terminal profile.
type MeasurementStore struct {
	Profile TerminalProfile        `json:"terminal"`
	Entries map[string]Measurement `json:"measurements"`
}

// NewMeasurementStore creates an empty store for a terminal profile.
func NewMeasurementStore(profile TerminalProfile) *MeasurementStore {
	return &MeasurementStore{Profile: profile, Entries: make(map[string]Measurement)}
}

// Merge adds measurements that are not already recorded, preserving prior answers.
func (s *MeasurementStore) Merge(measurements ...Measurement) {
	if s.Entries == nil {
		s.Entries = make(map[string]Measurement)
	}
	for _, m := range measurements {
		if m.Glyph == "" {
			continue
		}
		if _, exists := s.Entries[m.Glyph]; !exists {
			s.Entries[m.Glyph] = m
		}
	}
}

// Unmeasured returns glyphs without a saved result, preserving input order.
func (s *MeasurementStore) Unmeasured(glyphs []string) []string {
	result := make([]string, 0, len(glyphs))
	for _, glyph := range glyphs {
		if _, exists := s.Entries[glyph]; !exists {
			result = append(result, glyph)
		}
	}
	return result
}

// LoadMeasurementStore reads a profile store from JSON.
func LoadMeasurementStore(path string) (*MeasurementStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loomoji: read measurement store: %w", err)
	}
	var store MeasurementStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("loomoji: decode measurement store: %w", err)
	}
	if store.Entries == nil {
		store.Entries = make(map[string]Measurement)
	}
	return &store, nil
}

// Save writes the store as indented JSON, replacing the target atomically.
func (s *MeasurementStore) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("loomoji: encode measurement store: %w", err)
	}
	data = append(data, '\n')
	if err := writeAtomic(path, data); err != nil {
		return fmt.Errorf("loomoji: save measurement store: %w", err)
	}
	return nil
}

// SaveTextReport writes a stable, human-readable report for the store.
func (s *MeasurementStore) SaveTextReport(path string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "Terminal width measurements\nTERM=%s\nTERM_PROGRAM=%s\n\n", s.Profile.Term, s.Profile.TermProgram)
	glyphs := make([]string, 0, len(s.Entries))
	for glyph := range s.Entries {
		glyphs = append(glyphs, glyph)
	}
	sort.Strings(glyphs)
	for _, glyph := range glyphs {
		m := s.Entries[glyph]
		measured := fmt.Sprint(m.MeasuredWidth)
		if m.MeasuredWidth == 0 {
			measured = "other/unsure"
		}
		fmt.Fprintf(&b, "%s\t%s\tcomputed=%d\tmeasured=%s\tcomment=%s\n", glyph, strings.Join(m.Codepoints, " "), m.ComputedWidth, measured, m.Comment)
	}
	if err := writeAtomic(path, []byte(b.String())); err != nil {
		return fmt.Errorf("loomoji: save measurement report: %w", err)
	}
	return nil
}

// Paths returns the JSON and text report paths for a terminal profile.
func (p TerminalProfile) Paths(dir string) (jsonPath, textPath string) {
	term := p.Term
	if term == "" {
		term = "unknown"
	}
	program := p.TermProgram
	if program == "" {
		program = "unknown"
	}
	stem := url.PathEscape(term) + "--" + url.PathEscape(program)
	return filepath.Join(dir, stem+".json"), filepath.Join(dir, stem+".txt")
}

func writeAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".loomoji-widths-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
