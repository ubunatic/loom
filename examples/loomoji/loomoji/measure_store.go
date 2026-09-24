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

	_ "codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/measure"
)

// TerminalProfile identifies the terminal whose rendered widths were recorded.
type TerminalProfile struct {
	Term        string `json:"term"`
	TermProgram string `json:"term_program,omitempty"`
	VTEVersion  string `json:"vte_version,omitempty"`
}

// MeasureFilter selects one category of recorded measurements for review.
type MeasureFilter string

const (
	MeasureFilterAll             MeasureFilter = "all"
	MeasureFilterUnassessed      MeasureFilter = "unassessed"
	MeasureFilterWidth1          MeasureFilter = "width-1"
	MeasureFilterWidth2          MeasureFilter = "width-2"
	MeasureFilterWidth3          MeasureFilter = "width-3"
	MeasureFilterWidth4          MeasureFilter = "width-4"
	MeasureFilterUnsure          MeasureFilter = "unsure"
	MeasureFilterDiffersFromLoom MeasureFilter = "differs-from-loom"
	MeasureFilterHasComment      MeasureFilter = "has-comment"
)

// Measurement contains the observed and computed width of one glyph.
// MeasuredWidth is 1 through 4 for a confirmed result and 0 for unsure/other.
type Measurement struct {
	Glyph         string   `json:"glyph"`
	Codepoints    []string `json:"codepoints"`
	ComputedWidth int      `json:"computed_width"`
	VTEWidth      int      `json:"vte_width,omitempty"`
	MeasuredWidth int      `json:"measured_width"`
	RenderMode    string   `json:"render_mode,omitempty"`
	Answered      bool     `json:"answered,omitempty"`
	Comment       string   `json:"comment,omitempty"`
}

// Render mode identifiers
const (
	RenderModeDefault   = "default"
	RenderModePad1      = "pad-1"
	RenderModeNoVS16    = "no-vs16"
	RenderModeNoVS16Pad = "no-vs16-pad"
	RenderModeForceVS16 = "force-vs16"
	RenderModeSplitZWJ  = "split-zwj"
	RenderModeBaseOnly  = "base-only"
)

// AvailableRenderModes returns the rendering modes applicable to a glyph.
func AvailableRenderModes(glyph string) []string {
	modes := []string{RenderModeDefault, RenderModePad1}
	if strings.ContainsRune(glyph, '\uFE0F') {
		modes = append(modes, RenderModeNoVS16, RenderModeNoVS16Pad)
	}
	if !strings.ContainsRune(glyph, '\uFE0F') && !strings.ContainsRune(glyph, '\u200D') && len([]rune(glyph)) == 1 {
		modes = append(modes, RenderModeForceVS16)
	}
	if strings.ContainsRune(glyph, '\u200D') {
		modes = append(modes, RenderModeSplitZWJ)
	}
	runes := []rune(glyph)
	if len(runes) > 1 && (strings.ContainsRune(glyph, '\u200D') || strings.ContainsRune(glyph, '\uFE0F') || len(runes) > 2) {
		modes = append(modes, RenderModeBaseOnly)
	}
	return modes
}

// ApplyRenderMode transforms a glyph string according to the requested mode.
func ApplyRenderMode(glyph, mode string) string {
	switch mode {
	case RenderModePad1:
		return glyph + " "
	case RenderModeNoVS16:
		return strings.ReplaceAll(glyph, "\uFE0F", "")
	case RenderModeNoVS16Pad:
		return strings.ReplaceAll(glyph, "\uFE0F", "") + " "
	case RenderModeForceVS16:
		if strings.ContainsRune(glyph, '\uFE0F') {
			return glyph
		}
		return glyph + "\uFE0F"
	case RenderModeSplitZWJ:
		return strings.ReplaceAll(glyph, "\u200D", " ")
	case RenderModeBaseOnly:
		runes := []rune(glyph)
		if len(runes) > 0 {
			return string(runes[0])
		}
		return glyph
	default:
		return glyph
	}
}

// EvaluateVTEMode returns the recommended rendering mode for a glyph in VTE terminals.
func EvaluateVTEMode(glyph string) string {
	if glyph == "" {
		return RenderModeDefault
	}
	if glyph == "⟵" || glyph == "⟶" || glyph == "⟷" || glyph == "⟹" || glyph == "⟺" || glyph == "🐻‍❄️" || glyph == "👁️‍🗨️" {
		return RenderModePad1
	}
	if strings.ContainsRune(glyph, '\uFE0F') && !strings.ContainsRune(glyph, '\u200D') && glyph != "⭐️" {
		return RenderModePad1
	}
	return RenderModeDefault
}

// EvaluateVTEWidth predicts/evaluates the cell width expected for a glyph in VTE-based terminals.
func EvaluateVTEWidth(glyph string) int {
	return measure.StringWidth(glyph)
}

// NewMeasurement creates an unanswered record using Loom's current width policy and VTE evaluation.
func NewMeasurement(glyph string) Measurement {
	codepoints := make([]string, 0, len([]rune(glyph)))
	for _, r := range glyph {
		codepoints = append(codepoints, fmt.Sprintf("U+%04X", r))
	}
	mode := EvaluateVTEMode(glyph)
	return Measurement{
		Glyph:         glyph,
		Codepoints:    codepoints,
		ComputedWidth: measure.StringWidth(glyph),
		RenderMode:    mode,
		MeasuredWidth: EvaluateVTEWidth(glyph),
	}
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
		if m.ComputedWidth == 0 {
			m.ComputedWidth = measure.StringWidth(m.Glyph)
		}
		if m.RenderMode == "" {
			m.RenderMode = RenderModeDefault
		}
		if _, exists := s.Entries[m.Glyph]; !exists {
			s.Entries[m.Glyph] = m
		}
	}
}

// Set stores a measurement, replacing any previous record for its glyph.
func (s *MeasurementStore) Set(measurement Measurement) {
	if measurement.Glyph == "" {
		return
	}
	if measurement.RenderMode == "" {
		measurement.RenderMode = RenderModeDefault
	}
	if s.Entries == nil {
		s.Entries = make(map[string]Measurement)
	}
	s.Entries[measurement.Glyph] = measurement
}

// Unmeasured returns glyphs without a saved result, preserving input order.
func (s *MeasurementStore) Unmeasured(glyphs []string) []string {
	result := make([]string, 0, len(glyphs))
	for _, glyph := range glyphs {
		measurement, exists := s.Entries[glyph]
		if !exists || (!measurement.Answered && measurement.MeasuredWidth == 0) {
			result = append(result, glyph)
		}
	}
	return result
}

// Filter returns glyphs matching filter, preserving input order.
func (s *MeasurementStore) Filter(glyphs []string, filter MeasureFilter) []string {
	if filter == "" || filter == MeasureFilterAll {
		return append([]string(nil), glyphs...)
	}
	result := make([]string, 0, len(glyphs))
	for _, glyph := range glyphs {
		measurement, exists := s.Entries[glyph]
		if filter == MeasureFilterUnassessed || filter == "unmeasured" {
			if !exists || (!measurement.Answered && measurement.MeasuredWidth == 0) {
				result = append(result, glyph)
			}
			continue
		}
		if !exists || (!measurement.Answered && measurement.MeasuredWidth == 0) {
			continue
		}
		if measurementMatchesFilter(measurement, filter) {
			result = append(result, glyph)
		}
	}
	return result
}

// Review returns recorded glyphs matching filter, preserving input order.
func (s *MeasurementStore) Review(glyphs []string, filter MeasureFilter) []string {
	return s.Filter(glyphs, filter)
}

func measurementMatchesFilter(measurement Measurement, filter MeasureFilter) bool {
	switch filter {
	case MeasureFilterAll:
		return true
	case MeasureFilterWidth1:
		return measurement.MeasuredWidth == 1
	case MeasureFilterWidth2:
		return measurement.MeasuredWidth == 2
	case MeasureFilterWidth3:
		return measurement.MeasuredWidth == 3
	case MeasureFilterWidth4:
		return measurement.MeasuredWidth == 4
	case MeasureFilterUnsure:
		return measurement.MeasuredWidth == 0
	case MeasureFilterDiffersFromLoom:
		return measurement.MeasuredWidth != 0 && measurement.MeasuredWidth != measurement.ComputedWidth
	case MeasureFilterHasComment:
		return measurement.Comment != ""
	default:
		return false
	}
}

// LoadMeasurementStore reads a profile store from JSON.
func LoadMeasurementStore(path string) (*MeasurementStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewMeasurementStore(TerminalProfile{}), nil
		}
		return nil, fmt.Errorf("loomoji: read measurement store: %w", err)
	}
	var store MeasurementStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("loomoji: decode measurement store: %w", err)
	}
	if store.Entries == nil {
		store.Entries = make(map[string]Measurement)
	}
	for k, m := range store.Entries {
		if m.Glyph != "" {
			m.ComputedWidth = measure.StringWidth(m.Glyph)
			if m.RenderMode == "" {
				m.RenderMode = RenderModeDefault
			}
			store.Entries[k] = m
		}
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
	fmt.Fprintf(&b, "Terminal width measurements\nTERM=%s\nTERM_PROGRAM=%s\nVTE_VERSION=%s\n\n", s.Profile.Term, s.Profile.TermProgram, s.Profile.VTEVersion)
	glyphs := make([]string, 0, len(s.Entries))
	for glyph := range s.Entries {
		glyphs = append(glyphs, glyph)
	}
	sort.Strings(glyphs)
	for _, glyph := range glyphs {
		m := s.Entries[glyph]
		measured := "unanswered"
		if m.Answered || m.MeasuredWidth != 0 {
			if m.MeasuredWidth == 0 {
				measured = "other/unsure"
			} else {
				measured = fmt.Sprint(m.MeasuredWidth)
			}
		}
		comment := strings.NewReplacer("\t", " ", "\r", " ", "\n", " ").Replace(m.Comment)
		mode := m.RenderMode
		if mode == "" {
			mode = RenderModeDefault
		}
		if mode != RenderModeDefault {
			fmt.Fprintf(&b, "%s\t%s\tcomputed=%d\tmeasured=%s\tmode=%s\tcomment=%s\n", glyph, strings.Join(m.Codepoints, " "), m.ComputedWidth, measured, mode, comment)
		} else {
			fmt.Fprintf(&b, "%s\t%s\tcomputed=%d\tmeasured=%s\tcomment=%s\n", glyph, strings.Join(m.Codepoints, " "), m.ComputedWidth, measured, comment)
		}
	}
	if err := writeAtomic(path, []byte(b.String())); err != nil {
		return fmt.Errorf("loomoji: save measurement report: %w", err)
	}
	return nil
}

// Paths returns the JSON and text report paths for a terminal profile.
func (p TerminalProfile) Paths(dir string) (jsonPath, textPath string) {
	if p.VTEVersion != "" {
		stem := "vte-" + url.PathEscape(p.VTEVersion)
		return filepath.Join(dir, stem+".json"), filepath.Join(dir, stem+".txt")
	}
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
