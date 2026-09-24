// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// spec.go provides runtime configuration and override lookups for emoji and sequence widths.
package measure

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"gopkg.in/yaml.v3"
)

// Package measure defines and loads emoji/sequence specifications.
// EmojiSpec defines specced emoji and sequence width policy.
type EmojiSpec struct {
	VS16DefaultWidth  int             `yaml:"vs16_default_width"`
	VS16VTEMode       string          `yaml:"vs16_vte_mode"`
	VS16NonVTEMode    string          `yaml:"vs16_non_vte_mode,omitempty"`
	ZWJDefaultWidth   int             `yaml:"zwj_default_width"`
	FlagDefaultWidth  int             `yaml:"flag_default_width"`
	Overrides         []EmojiOverride `yaml:"overrides"`
}

// EmojiOverride defines an explicit glyph-to-width mapping.
type EmojiOverride struct {
	Glyph      string   `yaml:"glyph"`
	Codepoints []string `yaml:"codepoints"`
	Width      int      `yaml:"width"`
	VTEMode    string   `yaml:"vte_mode,omitempty"`
	NonVTEMode string   `yaml:"non_vte_mode,omitempty"`
	Note       string   `yaml:"note"`
}

// RenderPath identifies the terminal rendering strategy and width alignment profile.
type RenderPath string

const (
	// RenderPathVTE targets VTE-based terminals (e.g. GNOME Terminal, Tilix, Ptyxis)
	// which render VS16 emojis across 2 cells but only advance the cursor by 1 cell.
	RenderPathVTE RenderPath = "vte"

	// RenderPathStandard targets modern non-VTE / Wayland terminals (e.g. Foot, Alacritty, Kitty, WezTerm, Ghostty)
	// which advance cursor by 2 cells natively for wide emoji.
	RenderPathStandard RenderPath = "standard"

	// RenderPathNonVTE is an alias for RenderPathStandard.
	RenderPathNonVTE RenderPath = "non-vte"
)

type emojiRuntime struct {
	spec            EmojiSpec
	runeOverrides   map[rune]int
	glyphOverrides  map[string]int
	vteModes        map[string]string
	nonVTEModes     map[string]string
	vs16VTEMode     string
	vs16NonVTEMode  string
}

var currentEmojiRuntime atomic.Pointer[emojiRuntime]
var currentRenderPath atomic.Pointer[RenderPath]

func init() {
	def := defaultEmojiSpec()
	currentEmojiRuntime.Store(&emojiRuntime{
		spec:           def,
		runeOverrides:  buildRuneOverrides(def.Overrides),
		glyphOverrides: buildGlyphOverrides(def.Overrides),
		vteModes:       buildVTEModes(def.Overrides),
		nonVTEModes:    buildNonVTEModes(def.Overrides),
		vs16VTEMode:    def.VS16VTEMode,
		vs16NonVTEMode: def.VS16NonVTEMode,
	})
}

func defaultEmojiSpec() EmojiSpec {
	return EmojiSpec{
		VS16DefaultWidth:  2,
		VS16VTEMode:       "pad-1",
		VS16NonVTEMode:    "default",
		ZWJDefaultWidth:   2,
		FlagDefaultWidth:  2,
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

func buildVTEModes(overrides []EmojiOverride) map[string]string {
	m := make(map[string]string, len(overrides))
	for _, o := range overrides {
		if o.Glyph != "" && o.VTEMode != "" {
			m[o.Glyph] = o.VTEMode
		}
	}
	return m
}

func buildNonVTEModes(overrides []EmojiOverride) map[string]string {
	m := make(map[string]string, len(overrides))
	for _, o := range overrides {
		if o.Glyph != "" && o.NonVTEMode != "" {
			m[o.Glyph] = o.NonVTEMode
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
	if spec.VS16VTEMode == "" {
		spec.VS16VTEMode = "pad-1"
	}
	if spec.VS16NonVTEMode == "" {
		spec.VS16NonVTEMode = "default"
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
		vteModes:       buildVTEModes(spec.Overrides),
		nonVTEModes:    buildNonVTEModes(spec.Overrides),
		vs16VTEMode:    spec.VS16VTEMode,
		vs16NonVTEMode: spec.VS16NonVTEMode,
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

// DetectRenderPath inspects the process environment and terminal variables to determine the active RenderPath.
func DetectRenderPath() RenderPath {
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("LOOM_RENDER_PATH"))); v != "" {
		switch v {
		case "vte":
			return RenderPathVTE
		case "standard", "non-vte", "foot", "modern":
			return RenderPathStandard
		}
	}
	// Detect foot terminal via dedicated environment variables
	if os.Getenv("FOOT_TERMINAL_PID") != "" || os.Getenv("FOOT_APP_ID") != "" {
		return RenderPathStandard
	}
	term := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	if term == "foot" || term == "foot-extra" || term == "alacritty" || term == "xterm-kitty" || term == "ghostty" {
		return RenderPathStandard
	}
	prog := strings.ToLower(strings.TrimSpace(os.Getenv("TERM_PROGRAM")))
	if prog == "foot" || prog == "ghostty" || prog == "wezterm" || prog == "alacritty" || prog == "kitty" {
		return RenderPathStandard
	}
	if os.Getenv("VTE_VERSION") != "" {
		return RenderPathVTE
	}
	// Default to VTE path for current VTE-focused workloads
	return RenderPathVTE
}

// ActiveRenderPath returns the currently active RenderPath (either manually overridden or autodetected).
func ActiveRenderPath() RenderPath {
	if p := currentRenderPath.Load(); p != nil {
		return *p
	}
	return DetectRenderPath()
}

// SetRenderPath sets an explicit override for the active RenderPath.
func SetRenderPath(p RenderPath) {
	currentRenderPath.Store(&p)
}

// ResetRenderPath clears any explicit RenderPath override, reverting to autodetection.
func ResetRenderPath() {
	currentRenderPath.Store(nil)
}

// VTEMode returns the authoritative VTE rendering mode for a glyph from the spec.
func VTEMode(glyph string) string {
	if glyph == "" {
		return "default"
	}
	rt := currentEmojiRuntime.Load()
	if rt == nil {
		return "default"
	}
	if mode, ok := rt.vteModes[glyph]; ok {
		return mode
	}
	if strings.ContainsRune(glyph, '\uFE0F') && !strings.ContainsRune(glyph, '\u200D') {
		if rt.vs16VTEMode != "" {
			return rt.vs16VTEMode
		}
		return "pad-1"
	}
	return "default"
}

// NonVTEMode returns the authoritative modern Non-VTE / standard rendering mode for a glyph from the spec.
func NonVTEMode(glyph string) string {
	if glyph == "" {
		return "default"
	}
	rt := currentEmojiRuntime.Load()
	if rt == nil {
		return "default"
	}
	if mode, ok := rt.nonVTEModes[glyph]; ok {
		return mode
	}
	if strings.ContainsRune(glyph, '\uFE0F') && !strings.ContainsRune(glyph, '\u200D') {
		if rt.vs16NonVTEMode != "" {
			return rt.vs16NonVTEMode
		}
		return "default"
	}
	return "default"
}

// RenderMode returns the authoritative rendering mode for a glyph for the given RenderPath.
func RenderMode(glyph string, path RenderPath) string {
	switch path {
	case RenderPathStandard, RenderPathNonVTE:
		return NonVTEMode(glyph)
	case RenderPathVTE:
		fallthrough
	default:
		return VTEMode(glyph)
	}
}

// ApplyRenderMode transforms a glyph string according to the requested mode.
func ApplyRenderMode(glyph, mode string) string {
	switch mode {
	case "pad-1":
		return glyph + " "
	case "no-vs16":
		return strings.ReplaceAll(glyph, "\uFE0F", "")
	case "no-vs16-pad":
		return strings.ReplaceAll(glyph, "\uFE0F", "") + " "
	case "force-vs16":
		if strings.ContainsRune(glyph, '\uFE0F') {
			return glyph
		}
		return glyph + "\uFE0F"
	case "split-zwj":
		return strings.ReplaceAll(glyph, "\u200D", " ")
	case "base-only":
		runes := []rune(glyph)
		if len(runes) > 0 {
			return string(runes[0])
		}
		return glyph
	default:
		return glyph
	}
}

// ApplyVTEMode transforms a glyph string using its authoritative VTE mode from the spec.
func ApplyVTEMode(glyph string) string {
	return ApplyRenderMode(glyph, VTEMode(glyph))
}

// ApplyNonVTEMode transforms a glyph string using its authoritative Non-VTE mode from the spec.
func ApplyNonVTEMode(glyph string) string {
	return ApplyRenderMode(glyph, NonVTEMode(glyph))
}

// ApplyRenderPath transforms a glyph string according to the requested RenderPath.
func ApplyRenderPath(glyph string, path RenderPath) string {
	return ApplyRenderMode(glyph, RenderMode(glyph, path))
}

// ApplyAutoRenderMode transforms a glyph string according to the active RenderPath.
func ApplyAutoRenderMode(glyph string) string {
	return ApplyRenderPath(glyph, ActiveRenderPath())
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

