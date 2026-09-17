// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package measure

// Generated from Unicode emoji-data.txt (Emoji_Presentation), Unicode 18.0.
// https://www.unicode.org/Public/UCD/latest/ucd/emoji/emoji-data.txt
// Bits cover U+1F000 through U+1FAFF.
var emojiPresentationBits = [...]uint64{
	0x0000000000000010,
	0x0000000000000000,
	0x0000000000000000,
	0x0000000000008000,
	0x0000000000000000,
	0x0000000000000000,
	0x0000000007fe4000,
	0xffffffc000000000,
	0x077c800004000002,
	0x0000000000030000,
	0x0000000000000000,
	0x0000000000000000,
	0xffbfe001ffffffff,
	0xdfffffffffffffff,
	0xffffffff000fffff,
	0xff11ffff000f87ff,
	0x7fffffffffffffff,
	0xfffffffffffffffd,
	0xffffffffffffffff,
	0x9fffffffffffffff,
	0x3fffffffffffffff,
	0x040000ffffff7800,
	0x0000001000600000,
	0xf800000000000000,
	0xffffffffffffffff,
	0x000000000000ffff,
	0xffffffffffffffff,
	0x1ff01800f3e7103f,
	0x0000000000000000,
	0x0000000000000000,
	0x0000000000000000,
	0x00010fff00000000,
	0x0000000000000000,
	0x0000000000000000,
	0x0000000000000000,
	0x0000000000000000,
	0xf7fffffffffff000,
	0xffffffffffffffbf,
	0xffffffffffffffff,
	0xffffffffffffffff,
	0x0000000000000000,
	0x1fff000000000000,
	0xffffffffffffffff,
	0x07ff8fffbffff17f,
}

func emojiPresentationWidth(r rune) bool {
	offset := uint(r - 0x1f000)
	return emojiPresentationBits[offset/64]&(uint64(1)<<(offset%64)) != 0
}
