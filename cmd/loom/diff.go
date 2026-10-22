// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"ubunatic.com/loom"
	"ubunatic.com/loom/measure"
)

type diffExitError struct {
	code int
	err  error
}

func (e diffExitError) Error() string { return e.err.Error() }
func (e diffExitError) Unwrap() error { return e.err }

func errorExitCode(err error) int {
	var exit diffExitError
	if errors.As(err, &exit) {
		return exit.code
	}
	var sExit sessionExitError
	if errors.As(err, &sExit) {
		return sExit.code
	}
	return 1
}

type diffRegion struct{ row, col, width, height int }

func diffCommand() *cobra.Command {
	var regionText, gotRegionText, findText string
	var ignoreGlyph, ignoreColour bool
	cmd := &cobra.Command{
		Use: "diff <want.ansi> <got.ansi>", Short: "Compare ANSI screens cell by cell",
		Long: "Compare glyphs and foreground/background colours. Rows and columns use 1-based terminal cells (columns by display width). --region is row,col,width,height, also 1-based.",
		Args: cobra.ExactArgs(2), SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if gotRegionText != "" && findText != "" {
				return diffExitError{2, fmt.Errorf("--got-region and --find cannot be used together")}
			}
			var region *diffRegion
			if regionText != "" {
				parsed, err := parseDiffRegion(regionText)
				if err != nil {
					return diffExitError{2, err}
				}
				region = &parsed
			}
			if gotRegionText != "" || findText != "" {
				gotData, err := readAsset(args[1])
				if err != nil {
					return diffExitError{2, err}
				}
				got, err := parseDiffScreen(gotData)
				if err != nil {
					return diffExitError{2, err}
				}
				wantData, err := readAsset(args[0])
				if err != nil {
					return diffExitError{2, err}
				}
				want, err := parseDiffScreen(wantData)
				if err != nil {
					return diffExitError{2, err}
				}
				var crop diffRegion
				if gotRegionText != "" {
					crop, err = parseDiffRegion(gotRegionText)
				} else {
					crop, err = findDiffText(got, findText, want.Cols(), want.Rows())
				}
				if err != nil {
					return diffExitError{2, err}
				}
				if crop.row+crop.height > got.Rows() || crop.col+crop.width > got.Cols() {
					return diffExitError{2, fmt.Errorf("got region exceeds capture bounds")}
				}
				cropped := loom.NewAnsiBuffer(crop.width, crop.height)
				for y := 0; y < crop.height; y++ {
					for x := 0; x < crop.width; x++ {
						cropped.Set(x, y, got.Get(crop.col+x, crop.row+y))
					}
				}
				args[1] = ""
				gotData = []byte(cropped.Serialize())
				tmp, err := os.CreateTemp("", "loom-diff-crop-*.ansi")
				if err != nil {
					return diffExitError{2, err}
				}
				defer os.Remove(tmp.Name())
				if _, err = tmp.Write(gotData); err != nil {
					tmp.Close()
					return diffExitError{2, err}
				}
				if err = tmp.Close(); err != nil {
					return diffExitError{2, err}
				}
				args[1] = tmp.Name()
			}
			count, err := compareAnsiFiles(cmd.OutOrStdout(), args[0], args[1], region, ignoreGlyph, ignoreColour)
			if err != nil {
				return diffExitError{2, err}
			}
			if count > 0 {
				return diffExitError{1, fmt.Errorf("screens differ")}
			}
			fmt.Fprintln(cmd.OutOrStdout(), "screens match")
			return nil
		},
	}
	cmd.Flags().StringVar(&regionText, "region", "", "compare row,col,width,height (1-based terminal cells)")
	cmd.Flags().StringVar(&gotRegionText, "got-region", "", "crop got screen at row,col,width,height (1-based cells)")
	cmd.Flags().StringVar(&findText, "find", "", "find text in got screen and compare the reference-sized region there")
	cmd.Flags().BoolVar(&ignoreGlyph, "ignore-glyph", false, "ignore glyph differences")
	cmd.Flags().BoolVar(&ignoreColour, "ignore-colour", false, "ignore foreground and background colour differences")
	return cmd
}

func findDiffText(buf *loom.AnsiBuffer, text string, width, height int) (diffRegion, error) {
	needle := []rune(text)
	if len(needle) == 0 {
		return diffRegion{}, fmt.Errorf("--find text must not be empty")
	}
	found := diffRegion{width: width, height: height}
	matches := 0
	for y := 0; y < buf.Rows(); y++ {
		for x := 0; x < buf.Cols(); x++ {
			if x+measure.StringWidth(text) > buf.Cols() {
				continue
			}
			ok := true
			cellX := x
			for _, r := range needle {
				if buf.Get(cellX, y).Rune != r {
					ok = false
					break
				}
				cellX += max(1, measure.RuneWidth(r))
			}
			if ok {
				found.row, found.col = y, x
				matches++
			}
		}
	}
	if matches != 1 {
		return diffRegion{}, fmt.Errorf("--find %q matched %d locations", text, matches)
	}
	return found, nil
}

func parseDiffRegion(text string) (diffRegion, error) {
	parts := strings.Split(text, ",")
	if len(parts) != 4 {
		return diffRegion{}, fmt.Errorf("region must be row,col,width,height: %q", text)
	}
	values := make([]int, 4)
	for i, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n < 1 {
			return diffRegion{}, fmt.Errorf("region values must be positive integers: %q", text)
		}
		values[i] = n
	}
	return diffRegion{values[0] - 1, values[1] - 1, values[2], values[3]}, nil
}

func compareAnsiFiles(out io.Writer, wantPath, gotPath string, region *diffRegion, ignoreGlyph, ignoreColour bool) (int, error) {
	wantData, err := readAsset(wantPath)
	if err != nil {
		return 0, err
	}
	gotData, err := readAsset(gotPath)
	if err != nil {
		return 0, err
	}
	want, err := parseDiffScreen(wantData)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", wantPath, err)
	}
	got, err := parseDiffScreen(gotData)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", gotPath, err)
	}
	x0, y0, width, height := 0, 0, max(want.Cols(), got.Cols()), max(want.Rows(), got.Rows())
	if region != nil {
		x0, y0, width, height = region.col, region.row, region.width, region.height
	}
	differences := 0
	for y := y0; y < y0+height; y++ {
		for x := x0; x < x0+width; x++ {
			w, g := diffCell(want, x, y), diffCell(got, x, y)
			glyphDiff := !ignoreGlyph && w.Rune != g.Rune
			fgDiff, bgDiff := !ignoreColour && w.FG != g.FG, !ignoreColour && w.BG != g.BG
			if !glyphDiff && !fgDiff && !bgDiff {
				continue
			}
			differences++
			fmt.Fprintf(out, "row=%d col=%d want={glyph=%q fg=%s bg=%s} got={glyph=%q fg=%s bg=%s}\n", y+1, x+1, string(w.Rune), formatDiffColor(w.FG), formatDiffColor(w.BG), string(g.Rune), formatDiffColor(g.FG), formatDiffColor(g.BG))
		}
	}
	if differences > 0 {
		fmt.Fprintf(out, "differences: %d\n", differences)
	}
	return differences, nil
}

func parseDiffScreen(data []byte) (*loom.AnsiBuffer, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	return loom.ParseAnsiBuffer(text, 1, 1)
}

func diffCell(buf *loom.AnsiBuffer, x, y int) loom.AnsiCell {
	cell := buf.Get(x, y)
	if x > 0 {
		previous := buf.Get(x-1, y)
		if previous.Rune != ' ' && measure.RuneWidth(previous.Rune) > 1 && cell.Rune == ' ' {
			cell.FG, cell.BG = previous.FG, previous.BG
		}
	}
	return cell
}

func formatDiffColor(c loom.Color) string {
	if c == loom.ColorReset() {
		return "default"
	}
	for index := 0; index < 256; index++ {
		if c == loom.ColorIndex(uint8(index)) {
			return fmt.Sprintf("index:%d", index)
		}
	}
	r, g, b, ok := c.RGB()
	if ok {
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	return "default"
}
