// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// spec.go provides runtime configuration and override lookups for emoji and sequence widths.
package measure

import (
	"fmt"
	"sync/atomic"

	"gopkg.in/yaml.v3"
)

// Package measure defines and loads emoji/sequence specifications.
// EmojiSpec defines specced emoji and sequence width policy.
type EmojiSpec struct {
	VS16DefaultWidth int             `yaml:"vs16_default_width"`
	ZWJDefaultWidth  int             `yaml:"zwj_default_width"`
	FlagDefaultWidth int             `yaml:"flag_default_width"`
	Overrides        []EmojiOverride `yaml:"overrides"`
}

// EmojiOverride defines an explicit glyph-to-width mapping.
type EmojiOverride struct {
	Glyph      string   `yaml:"glyph"`
	Codepoints []string `yaml:"codepoints"`
	Width      int      `yaml:"width"`
	Note       string   `yaml:"note"`
}

type emojiRuntime struct {
	spec           EmojiSpec
	runeOverrides  map[rune]int
	glyphOverrides map[string]int
}

var currentEmojiRuntime atomic.Pointer[emojiRuntime]

func init() {
	def := defaultEmojiSpec()
	currentEmojiRuntime.Store(&emojiRuntime{
		spec:           def,
		runeOverrides:  buildRuneOverrides(def.Overrides),
		glyphOverrides: buildGlyphOverrides(def.Overrides),
	})
}

func defaultEmojiSpec() EmojiSpec {
	return EmojiSpec{
		VS16DefaultWidth: 2,
		ZWJDefaultWidth:  2,
		FlagDefaultWidth: 2,
	}
}

func buildRuneOverrides(overrides []EmojiOverride) map[rune]int {
	m := make(map[rune]int, len(overrides))
	for _, o := range overrides {
		rs := []rune(o.Glyph)
		if len(rs) == 1 {
			m[rs[0]] = o.Width
		}
	}
	return m
}

func buildGlyphOverrides(overrides []EmojiOverride) map[string]int {
	m := make(map[string]int, len(overrides))
	for _, o := range overrides {
		if o.Glyph != "" {
			m[o.Glyph] = o.Width
		}
	}
	return m
}

// LoadEmojiSpecYAML decodes and activates an emoji specification from YAML bytes.
func LoadEmojiSpecYAML(data []byte) error {
	var spec EmojiSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("measure: unmarshal emoji spec: %w", err)
	}
	LoadEmojiSpec(spec)
	return nil
}

// LoadEmojiSpec installs a validated EmojiSpec into the measurement runtime.
func LoadEmojiSpec(spec EmojiSpec) {
	if spec.VS16DefaultWidth <= 0 {
		spec.VS16DefaultWidth = 2
	}
	if spec.ZWJDefaultWidth <= 0 {
		spec.ZWJDefaultWidth = 2
	}
	if spec.FlagDefaultWidth <= 0 {
		spec.FlagDefaultWidth = 2
	}
	currentEmojiRuntime.Store(&emojiRuntime{
		spec:           spec,
		runeOverrides:  buildRuneOverrides(spec.Overrides),
		glyphOverrides: buildGlyphOverrides(spec.Overrides),
	})
}

// ActiveEmojiSpec returns a copy of the active emoji specification.
func ActiveEmojiSpec() EmojiSpec {
	rt := currentEmojiRuntime.Load()
	if rt == nil {
		return defaultEmojiSpec()
	}
	return rt.spec
}

func getRuneOverride(r rune) (int, bool) {
	rt := currentEmojiRuntime.Load()
	if rt == nil || rt.runeOverrides == nil {
		return 0, false
	}
	w, ok := rt.runeOverrides[r]
	return w, ok
}

func getGlyphOverride(g string) (int, bool) {
	rt := currentEmojiRuntime.Load()
	if rt == nil || rt.glyphOverrides == nil {
		return 0, false
	}
	w, ok := rt.glyphOverrides[g]
	return w, ok
}
