// Command treesitter measures the cgo-free Tree-sitter/wazero canary.
// Run with: go run ./internal/canary/treesitter
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	sitter "github.com/malivvan/tree-sitter"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	started := time.Now()
	ts, err := sitter.New(ctx)
	if err != nil {
		return err
	}
	loadTime := time.Since(started)
	runtime.ReadMemStats(&after)
	fmt.Printf("runtime+module load: %s, Go heap delta %d B, wasm image embedded in module\n", loadTime, after.TotalAlloc-before.TotalAlloc)

	lang, err := ts.LanguageC(ctx)
	if err != nil {
		return err
	}
	query, err := ts.NewQuery(ctx, "(function_definition) @function", lang)
	if err != nil {
		return err
	}
	_ = query
	fmt.Println("grammar: C (bundled Wasm export); query: (function_definition) @function")

	for _, lines := range []int{100, 1000, 10000} {
		source := cSource(lines)
		parser, err := ts.NewParser(ctx)
		if err != nil {
			return err
		}
		if err := parser.SetLanguage(ctx, lang); err != nil {
			return err
		}
		var parseBefore runtime.MemStats
		runtime.ReadMemStats(&parseBefore)
		started = time.Now()
		tree, err := parser.ParseString(ctx, source)
		if err != nil {
			return err
		}
		initial := time.Since(started)
		var parseMem runtime.MemStats
		runtime.ReadMemStats(&parseMem)
		root, err := tree.RootNode(ctx)
		if err != nil {
			return err
		}
		kind, err := root.Kind(ctx)
		if err != nil {
			return err
		}
		cursor, err := ts.NewQueryCursor(ctx)
		if err != nil {
			return err
		}
		if err := cursor.Exec(ctx, query, root); err != nil {
			return err
		}
		matches := 0
		for {
			_, ok, err := cursor.NextMatch(ctx)
			if err != nil {
				return err
			}
			if !ok {
				break
			}
			matches++
		}

		// The published wrapper has no TSInputEdit/ts_tree_edit binding, so
		// this measures a full reparse after a one-byte edit, not incremental reuse.
		edited := source + "\n"
		started = time.Now()
		_, err = parser.ParseString(ctx, edited)
		if err != nil {
			return err
		}
		reparse := time.Since(started)
		fmt.Printf("lines=%d bytes=%d root=%s initial=%s Go-alloc=%d B query-matches=%d full-reparse-after-1-byte-edit=%s\n", lines, len(source), kind, initial, parseMem.TotalAlloc-parseBefore.TotalAlloc, matches, reparse)
		if err := parser.Close(ctx); err != nil {
			return err
		}
	}
	return nil
}

func cSource(lines int) string {
	var b strings.Builder
	b.Grow(lines * 32)
	for i := 0; i < lines; i++ {
		fmt.Fprintf(&b, "int value%d(void) { return %d; }\n", i, i)
	}
	return b.String()
}
