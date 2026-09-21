package graph

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTreemapLayoutProperties081(t *testing.T) {
	rng := rand.New(rand.NewSource(81))
	for layout := TreemapLayoutSliceDice; layout <= TreemapLayoutSquarified; layout++ {
		for n := 0; n < 220; n++ {
			w, h := 1+rng.Intn(80), 1+rng.Intn(40)
			segments := make([]TreemapSegment, 1+rng.Intn(12))
			for i := range segments {
				segments[i] = TreemapSegment{Name: fmt.Sprintf("n%d", i), Value: float64(rng.Intn(20) - 4)}
			}
			if n%11 == 0 {
				segments[rng.Intn(len(segments))].Value = 0
			}
			cells := LayoutTreemap(segments, w, h, layout)
			checkCells081(t, cells, segments, w, h)
			again := LayoutTreemap(segments, w, h, layout)
			if fmt.Sprint(cells) != fmt.Sprint(again) {
				t.Fatalf("layout %d is not deterministic", layout)
			}
		}
		for _, dimensions := range [][2]int{{0, 4}, {4, 0}, {-1, 2}, {2, -1}} {
			if got := LayoutTreemap(nil, dimensions[0], dimensions[1], layout); got != nil {
				t.Fatalf("invalid dimensions returned cells: %+v", got)
			}
		}
	}
}

func checkCells081(t *testing.T, cells []TreemapCell, segments []TreemapSegment, w, h int) {
	t.Helper()
	positive := false
	for _, segment := range segments {
		positive = positive || segment.Value > 0 && !math.IsNaN(segment.Value) && !math.IsInf(segment.Value, 0)
	}
	grid := make([]int, w*h)
	for _, cell := range cells {
		if cell.Value <= 0 {
			t.Fatalf("non-positive cell: %+v", cell)
		}
		r := cell.Rect
		if r.X < 0 || r.Y < 0 || r.W <= 0 || r.H <= 0 || r.X+r.W > w || r.Y+r.H > h {
			t.Fatalf("out-of-bounds cell %+v in %dx%d", r, w, h)
		}
		if cell.Index < 0 || cell.Index >= len(segments) || segments[cell.Index] != cell.Segment {
			t.Fatalf("bad metadata: %+v", cell)
		}
		for y := r.Y; y < r.Y+r.H; y++ {
			for x := r.X; x < r.X+r.W; x++ {
				grid[y*w+x]++
			}
		}
	}
	for i, count := range grid {
		if !positive {
			if count != 0 {
				t.Fatalf("empty input covered grid index %d", i)
			}
			continue
		}
		if count != 1 {
			t.Fatalf("grid index %d covered %d times for %dx%d", i, count, w, h)
		}
	}
}

func TestTreemapAspectRatio081(t *testing.T) {
	cases := [][]float64{{1, 2, 3, 5, 8}, {1, 1, 1, 1, 1, 1}, {2, 20, 3, 10, 7}, {1, 4, 16, 2, 8, 32}, {9, 8, 7, 6, 5, 4, 3}}
	var sliceMean, squareMean float64
	for _, values := range cases {
		segments := make([]TreemapSegment, len(values))
		for i, value := range values {
			segments[i] = TreemapSegment{Value: value}
		}
		s := worstAspect081(LayoutTreemap(segments, 80, 30, TreemapLayoutSliceDice))
		q := worstAspect081(LayoutTreemap(segments, 80, 30, TreemapLayoutSquarified))
		sliceMean += s
		squareMean += q
		t.Logf("distribution=%v slice-dice=%.3f squarified=%.3f", values, s, q)
	}
	sliceMean /= float64(len(cases))
	squareMean /= float64(len(cases))
	t.Logf("mean worst aspect: slice-dice=%.3f squarified=%.3f", sliceMean, squareMean)
	if squareMean > sliceMean+1e-9 {
		t.Fatalf("squarified aspect %.3f worse than slice-dice %.3f", squareMean, sliceMean)
	}
}

func worstAspect081(cells []TreemapCell) float64 {
	worst := 0.0
	for _, cell := range cells {
		r := cell.Rect
		a, b := float64(r.W), float64(r.H*2)
		if a < b {
			a, b = b, a
		}
		worst = math.Max(worst, a/b)
	}
	return worst
}

func TestColorScaleEdges081(t *testing.T) {
	scale := HeatColorScale()
	checks := []struct{ value, min, max float64 }{{-1, 10, 0}, {5, 5, 5}, {math.Inf(1), 0, 1}, {math.Inf(-1), 0, 1}, {math.NaN(), 0, 1}, {-5, -10, 0}, {-5, -10, 0}}
	for _, check := range checks {
		_ = scale(check.value, check.min, check.max)
	}
	if got := scale(5, 0, 10).ForegroundANSI; got != "38;2;128;0;128" {
		t.Fatalf("midpoint = %q", got)
	}
	if scale(0, 0, 10).ForegroundANSI >= scale(10, 0, 10).ForegroundANSI {
		t.Fatal("heat scale is not monotonic by red endpoint")
	}
	_ = scale(-5, 10, 0)
}

func TestRenderTreemapGolden081(t *testing.T) {
	got := strings.Join(RenderTreemap([]TreemapSegment{{Name: "alpha", Value: 5}, {Name: "beta", Value: 3}}, TreemapOptions{Width: 18, Height: 5}), "\n")
	want := "┌─────────┐┌─────┐\n│█████████││▓▓▓▓▓│\n│alpha████││beta▓│\n│█████████││▓▓▓▓▓│\n└─────────┘└─────┘"
	if got != want {
		t.Fatalf("renderer golden changed:\n%q", got)
	}
}

func TestTreemapEvidence081(t *testing.T) {
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("LOOM_EVIDENCE is not enabled")
	}
	root := filepath.Join("..", "docs", "progress", "081")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	data := []TreemapSegment{{Name: "alpha", Value: 10}, {Name: "beta", Value: 7}, {Name: "gamma", Value: 4}, {Name: "delta", Value: 2}}
	for _, layout := range []struct {
		name  string
		value TreemapLayout
	}{{"slicedice", TreemapLayoutSliceDice}, {"squarified", TreemapLayoutSquarified}} {
		for i, size := range [][2]int{{32, 10}, {48, 14}, {64, 18}} {
			prefix := "M1-"
			if layout.value == TreemapLayoutSquarified {
				prefix = "M2-"
			}
			writeEvidence081(t, filepath.Join(root, fmt.Sprintf("%s%s-s%d.ansi", prefix, layout.name, i+1)), data, size[0], size[1], layout.value, nil)
		}
	}
	writeEvidence081(t, filepath.Join(root, "M3-colorscale.ansi"), data, 48, 14, TreemapLayoutSquarified, HeatColorScale())
}

func writeEvidence081(t *testing.T, path string, segments []TreemapSegment, w, h int, layout TreemapLayout, scale ColorScale) {
	cells := LayoutTreemap(segments, w, h, layout)
	grid := make([][]rune, h)
	for y := range grid {
		grid[y] = []rune(strings.Repeat(" ", w))
	}
	for _, cell := range cells {
		r := cell.Rect
		for x := r.X; x < r.X+r.W; x++ {
			grid[r.Y][x], grid[r.Y+r.H-1][x] = '─', '─'
		}
		for y := r.Y; y < r.Y+r.H; y++ {
			grid[y][r.X], grid[y][r.X+r.W-1] = '│', '│'
		}
		grid[r.Y][r.X], grid[r.Y][r.X+r.W-1], grid[r.Y+r.H-1][r.X], grid[r.Y+r.H-1][r.X+r.W-1] = '┌', '┐', '└', '┘'
		if r.W > len(cell.Label)+2 && r.H >= 3 {
			copy(grid[r.Y+r.H/2][r.X+1:], []rune(cell.Label))
		}
	}
	var out strings.Builder
	for _, row := range grid {
		out.WriteString(string(row))
		out.WriteByte('\n')
	}
	if scale != nil {
		out.WriteString("heat: blue low -> red high\n")
	}
	if err := os.WriteFile(path, []byte(out.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Log(out.String())
}
