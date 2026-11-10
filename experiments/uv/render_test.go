package main

import (
	"bytes"
	"strings"
	"testing"
	"fmt"

	"ubunatic.com/loom"
	"ubunatic.com/loom/vt"
	uv "github.com/charmbracelet/ultraviolet"
)

func assertSame(t *testing.T, loomOut, uvOut []byte, cols, rows int) {
	vLoom := vt.NewVT(cols, rows)
	vLoom.Write(loomOut)

	vUv := vt.NewVT(cols, rows)
	vUv.Write(uvOut)

	sLoom := strings.Join(vLoom.Screen(), "\n")
	sUv := strings.Join(vUv.Screen(), "\n")

	if sLoom != sUv {
		t.Errorf("Outputs differ!")
	}
}

func TestEquality(t *testing.T) {
	c := loom.NewCanvas(200, 60)
	c.Write(0, 0, "Hello World", loom.Style{Bold: true})
	c.Write(10, 10, "Line 10", loom.Style{})
	c.Write(150, 50, "Line 50", loom.Style{Italic: true})

	var loomBuf bytes.Buffer
	c.Flush(&loomBuf, 1)

	var uvBuf bytes.Buffer
	scr := canvasToUV(c, &uvBuf)
	scr.Render()
	scr.Flush()

	assertSame(t, loomBuf.Bytes(), uvBuf.Bytes(), 200, 60)
}

func BenchmarkFlushScenarios(b *testing.B) {
	cols, rows := 200, 60

	c := loom.NewCanvas(cols, rows)
	for y := 0; y < rows; y++ {
		c.Write(0, y, strings.Repeat("A", cols), loom.Style{})
	}

	// First full paint
	b.Run("FirstPaint/Loom", func(b *testing.B) {
		b.ReportAllocs()
		var out bytes.Buffer
		for i := 0; i < b.N; i++ {
			out.Reset()
			c.Flush(&out, 1)
			b.ReportMetric(float64(out.Len()), "bytes/op")
		}
	})

	b.Run("FirstPaint/Ultraviolet", func(b *testing.B) {
		b.ReportAllocs()
		var out bytes.Buffer
		for i := 0; i < b.N; i++ {
			out.Reset()
			scr := uv.NewTerminalScreen(&out, nil)
			scr.Resize(cols, rows)
			for y := 0; y < rows; y++ {
				for x := 0; x < cols; x++ {
					scr.SetCell(x, y, uv.NewCell(scr.WidthMethod(), "A"))
				}
			}
			scr.Render()
			scr.Flush()
			b.ReportMetric(float64(out.Len()), "bytes/op")
		}
	})

	// Frame where one line changes
	b.Run("OneLineChanged/Loom", func(b *testing.B) {
		b.ReportAllocs()
		var out bytes.Buffer
		for i := 0; i < b.N; i++ {
			c.Write(0, 10, fmt.Sprintf("B%d%s", i%10, strings.Repeat("A", cols-5)), loom.Style{})
			out.Reset()
			c.Flush(&out, 1)
			b.ReportMetric(float64(out.Len()), "bytes/op")
		}
	})

	b.Run("OneLineChanged/Ultraviolet", func(b *testing.B) {
		b.ReportAllocs()
		var out bytes.Buffer
		scr := uv.NewTerminalScreen(&out, nil)
		scr.Resize(cols, rows)
		for y := 0; y < rows; y++ {
			for x := 0; x < cols; x++ {
				scr.SetCell(x, y, uv.NewCell(scr.WidthMethod(), "A"))
			}
		}
		scr.Render()
		scr.Flush() // First paint
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			out.Reset()
			line := fmt.Sprintf("B%d%s", i%10, strings.Repeat("A", cols-5))
			for x, r := range line {
				scr.SetCell(x, 10, uv.NewCell(scr.WidthMethod(), string(r)))
			}
			scr.Render()
			scr.Flush()
			b.ReportMetric(float64(out.Len()), "bytes/op")
		}
	})

	// Frame where selection moves (two lines change)
	b.Run("SelectionMoves/Loom", func(b *testing.B) {
		b.ReportAllocs()
		var out bytes.Buffer
		for i := 0; i < b.N; i++ {
			c.Write(0, i%rows, strings.Repeat("B", cols), loom.Style{Bold: true})
			c.Write(0, (i-1+rows)%rows, strings.Repeat("A", cols), loom.Style{})
			out.Reset()
			c.Flush(&out, 1)
			b.ReportMetric(float64(out.Len()), "bytes/op")
		}
	})

	b.Run("SelectionMoves/Ultraviolet", func(b *testing.B) {
		b.ReportAllocs()
		var out bytes.Buffer
		scr := uv.NewTerminalScreen(&out, nil)
		scr.Resize(cols, rows)
		for y := 0; y < rows; y++ {
			for x := 0; x < cols; x++ {
				scr.SetCell(x, y, uv.NewCell(scr.WidthMethod(), "A"))
			}
		}
		scr.Render()
		scr.Flush()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			out.Reset()
			for x := 0; x < cols; x++ {
				cB := uv.NewCell(scr.WidthMethod(), "B")
				cB.Style.Attrs |= uv.AttrBold
				scr.SetCell(x, i%rows, cB)

				cA := uv.NewCell(scr.WidthMethod(), "A")
				scr.SetCell(x, (i-1+rows)%rows, cA)
			}
			scr.Render()
			scr.Flush()
			b.ReportMetric(float64(out.Len()), "bytes/op")
		}
	})

	// Scroll by one line
	b.Run("Scroll/Loom", func(b *testing.B) {
		b.ReportAllocs()
		var out bytes.Buffer
		for i := 0; i < b.N; i++ {
			for y := 0; y < rows-1; y++ {
				c.Write(0, y, fmt.Sprintf("L%d%s", (y+i)%10, strings.Repeat("A", cols-5)), loom.Style{})
			}
			c.Write(0, rows-1, fmt.Sprintf("L%d%s", (rows-1+i)%10, strings.Repeat("A", cols-5)), loom.Style{})
			out.Reset()
			c.Flush(&out, 1)
			b.ReportMetric(float64(out.Len()), "bytes/op")
		}
	})

	b.Run("Scroll/Ultraviolet", func(b *testing.B) {
		b.ReportAllocs()
		var out bytes.Buffer
		scr := uv.NewTerminalScreen(&out, nil)
		scr.Resize(cols, rows)
		for y := 0; y < rows; y++ {
			for x := 0; x < cols; x++ {
				scr.SetCell(x, y, uv.NewCell(scr.WidthMethod(), "A"))
			}
		}
		scr.Render()
		scr.Flush()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			out.Reset()
			// This tests diffing scroll, not uv's scroll primitives specifically if it has them
			for y := 0; y < rows-1; y++ {
				line := fmt.Sprintf("L%d%s", (y+i)%10, strings.Repeat("A", cols-5))
				for x, r := range line {
					scr.SetCell(x, y, uv.NewCell(scr.WidthMethod(), string(r)))
				}
			}
			line := fmt.Sprintf("L%d%s", (rows-1+i)%10, strings.Repeat("A", cols-5))
			for x, r := range line {
				scr.SetCell(x, rows-1, uv.NewCell(scr.WidthMethod(), string(r)))
			}

			scr.Render()
			scr.Flush()
			b.ReportMetric(float64(out.Len()), "bytes/op")
		}
	})
}
