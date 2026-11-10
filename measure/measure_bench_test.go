// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package measure_test

import (
	"testing"
	"ubunatic.com/loom/measure"
)

var asciiStr = "Hello World! This is a simple ASCII string for testing."
var cjkStr = "你好，世界！這是一個用於測試的簡單 ASCII 字符串。"
var emojiStr = "👋🌍👨‍👩‍👧‍👦 This is a test 🚀🔥💯"

func BenchmarkStringWidth_ASCII(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = measure.StringWidth(asciiStr)
	}
}

func BenchmarkStringWidth_CJK(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = measure.StringWidth(cjkStr)
	}
}

func BenchmarkStringWidth_Emoji(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = measure.StringWidth(emojiStr)
	}
}

func BenchmarkClusters_ASCII(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = measure.Clusters(asciiStr)
	}
}

func BenchmarkClusters_CJK(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = measure.Clusters(cjkStr)
	}
}

func BenchmarkClusters_Emoji(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = measure.Clusters(emojiStr)
	}
}
