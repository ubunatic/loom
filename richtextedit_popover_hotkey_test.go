package loom

import (
	"strings"
	"testing"

	"ubunatic.com/loom/measure"
)

func TestRichTextEditPopoverLabelsAndHotkeys(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "selected"}}}}})
	edit.SetSelection(RichPosition{}, RichPosition{Offset: 8})
	edit.ConsumeKey(KeyEvent{Key: "ctrl-space"})
	canvas := NewCanvas(70, 5)
	edit.Draw(canvas, canvas.Bounds())

	var got strings.Builder
	bar := edit.popoverButtons[0].rect
	for x := bar.X; x < edit.popoverButtons[len(edit.popoverButtons)-1].rect.X+edit.popoverButtons[len(edit.popoverButtons)-1].rect.W; x++ {
		got.WriteString(canvas.Get(x, bar.Y).Text)
	}
	want := " B ▏I ▏U ▏S ▏🎨 ▏Frame ▏Link ▏≡ "
	if got.String() != want {
		t.Fatalf("toolbar = %q, want %q", got.String(), want)
	}
	for _, tc := range []struct {
		label  string
		button int
		column int
	}{{"Frame", 5, 1}, {"Link", 6, 1}} {
		cell := canvas.Get(edit.popoverButtons[tc.button].rect.X+tc.column, bar.Y)
		if !cell.Style.Bold {
			t.Errorf("hotkey rune for %s is not bold: %+v", tc.label, cell.Style)
		}
		if canvas.Get(edit.popoverButtons[tc.button].rect.X+tc.column+1, bar.Y).Style.Bold {
			t.Errorf("rune after %s hotkey is bold", tc.label)
		}
	}
	if measure.StringWidth(SpeccedDefaults.RichTextEdit.SeparatorGlyph) != 1 {
		t.Fatalf("separator width = %d, want one cell", measure.StringWidth(SpeccedDefaults.RichTextEdit.SeparatorGlyph))
	}
}

func TestRichTextEditFocusedPopoverHighlightsWholeItemByDisplayColumn(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "selected"}}}}})
	edit.SetSelection(RichPosition{}, RichPosition{Offset: 8})
	edit.ConsumeKey(KeyEvent{Key: "ctrl-space"})
	canvas := NewCanvas(100, 5)
	defs := SpeccedDefaults.RichTextEdit
	focusBG := ColorIndex(uint8(defs.PopoverFocusBG))
	focusFG := ColorIndex(uint8(defs.PopoverFocusFG))
	separatorFG := ColorIndex(uint8(defs.SeparatorFG))
	separatorBG := ColorIndex(uint8(defs.SeparatorBG))

	for _, action := range []string{"B", "S", "ColorPicker", "Hamburger"} {
		focused := popoverActionIndex(action)
		edit.popoverFocus, edit.popoverFocusSet = focused, true
		edit.Draw(canvas, canvas.Bounds())
		button := edit.popoverButtons[focused]
		for column := 0; column < button.rect.W; column++ {
			cell := canvas.Get(button.rect.X+column, button.rect.Y)
			wantFG := focusFG
			if focused > 0 && column == 0 {
				wantFG = separatorFG
			}
			if cell.Style.BG != focusBG || cell.Style.FG != wantFG {
				t.Errorf("focus %q column %d cell %+v, want FG=%v BG=%v", action, column, cell, wantFG, focusBG)
			}
		}
		if focused == 0 {
			if got := canvas.Get(button.rect.X, button.rect.Y).Text; got != " " {
				t.Errorf("B leading edge pad = %q, want space", got)
			}
		} else if got := canvas.Get(button.rect.X, button.rect.Y).Text; got != defs.SeparatorGlyph {
			t.Errorf("%s preceding separator = %q, want %q", action, got, defs.SeparatorGlyph)
		}
		if got := canvas.Get(button.rect.X+1, button.rect.Y).Text; got != button.label {
			t.Errorf("%s label cell = %q, want %q", action, got, button.label)
		}
		if got := canvas.Get(button.rect.X+button.rect.W-1, button.rect.Y).Text; got != " " {
			t.Errorf("%s trailing pad = %q, want space", action, got)
		}
		if action == "ColorPicker" && button.rect.W != measure.StringWidth("🎨")+2 {
			t.Errorf("color item width = %d, want display width %d plus separator and pad", button.rect.W, measure.StringWidth("🎨")+2)
		}
		if focused < len(edit.popoverButtons)-1 {
			separator := canvas.Get(button.rect.X+button.rect.W, button.rect.Y)
			if separator.Text != defs.SeparatorGlyph || separator.Style.BG != separatorBG {
				t.Errorf("separator after %s = %+v, want unhighlighted separator", action, separator)
			}
		}
	}
}

func TestRichTextEditPopoverSeparatorMouseHitActivatesFollowingItem(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "selected"}}}}}
	edit := NewRichTextEdit(doc)
	edit.SetSelection(RichPosition{}, RichPosition{Offset: 8})
	edit.ConsumeKey(KeyEvent{Key: "ctrl-space"})
	canvas := NewCanvas(100, 5)
	edit.Draw(canvas, Rect{W: 100, H: 5})
	button := edit.popoverButtons[popoverActionIndex("S")]
	previous := edit.popoverButtons[popoverActionIndex("U")].rect
	if button.rect.X != previous.X+previous.W || canvas.Get(button.rect.X, button.rect.Y).Text != SpeccedDefaults.RichTextEdit.SeparatorGlyph {
		t.Fatal("S hit area does not include its preceding separator")
	}
	if result := edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: button.rect.X, Y: button.rect.Y}); !result.Consumed {
		t.Fatal("separator click was not consumed")
	}
	if !doc.Lines[0].Spans[0].Style.Strike {
		t.Fatal("separator click did not activate the S item")
	}
}

func TestRichTextEditPopoverLetterHotkeys(t *testing.T) {
	for _, tc := range []struct {
		key    string
		action string
	}{{"b", "B"}, {"B", "B"}, {"i", "I"}, {"I", "I"}, {"u", "U"}, {"U", "U"}, {"s", "S"}, {"S", "S"}} {
		doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "selected"}}}}}
		edit := NewRichTextEdit(doc)
		edit.PopoverDefaults.AutoPopupOnSelection = false
		edit.SetSelection(RichPosition{}, RichPosition{Offset: 8})
		edit.ConsumeKey(KeyEvent{Key: "ctrl-space"})
		edit.Draw(NewCanvas(70, 5), Rect{W: 70, H: 5})
		if result := edit.ConsumeKey(KeyEvent{Key: tc.key, Text: tc.key}); !result.Consumed {
			t.Fatalf("%q was not consumed while toolbar focused", tc.key)
		}
		style := doc.Lines[0].Spans[0].Style
		if !styleForAction(style, tc.action) {
			t.Errorf("%q did not apply %s: %+v", tc.key, tc.action, style)
		}
	}
	for _, key := range []string{"f", "F", "l", "L"} {
		edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "selected"}}}}})
		edit.PopoverDefaults.AutoPopupOnSelection = false
		edit.SetSelection(RichPosition{}, RichPosition{Offset: 8})
		edit.ConsumeKey(KeyEvent{Key: "ctrl-space"})
		edit.Draw(NewCanvas(70, 5), Rect{W: 70, H: 5})
		if result := edit.ConsumeKey(KeyEvent{Key: key, Text: key}); !result.Consumed {
			t.Errorf("%q was not consumed while toolbar focused", key)
		}
		if key == "f" || key == "F" {
			if edit.popoverSubmenu != "Box" {
				t.Errorf("%q submenu = %q, want Box", key, edit.popoverSubmenu)
			}
		} else if edit.popoverSubmenu != "" {
			t.Errorf("disabled Link hotkey %q opened submenu %q", key, edit.popoverSubmenu)
		}
		if key == "l" || key == "L" {
			if got := edit.Document.ToPlainText(); got != "selected" {
				t.Errorf("disabled Link hotkey %q changed text to %q", key, got)
			}
		}
	}
}

func TestRichTextEditCtrlSpaceFocusesFirstItemForHotkeys(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello loom"}}}}}
	edit := NewRichTextEdit(doc)
	edit.ConsumeKey(KeyEvent{Key: "ctrl-space"})
	if !edit.popoverFocusSet || edit.popoverFocus != 0 || !edit.HasSelection {
		t.Fatalf("Ctrl+Space focus state: focused=%v index=%d selection=%v", edit.popoverFocusSet, edit.popoverFocus, edit.HasSelection)
	}
	edit.Draw(NewCanvas(70, 5), Rect{W: 70, H: 5})
	if result := edit.ConsumeKey(KeyEvent{Key: "b", Text: "b"}); !result.Consumed || !doc.Lines[0].Spans[0].Style.Bold {
		t.Fatalf("b after Ctrl+Space = %+v, style=%+v", result, doc.Lines[0].Spans[0].Style)
	}
	if got := doc.ToPlainText(); got != "hello loom" {
		t.Fatalf("b edited selected text to %q", got)
	}
}

func TestRichTextEditFocusedAutoBarAcceptsLetterHotkeys(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "selected"}}}}}
	edit := NewRichTextEdit(doc)
	edit.SetSelection(RichPosition{}, RichPosition{Offset: 8})
	edit.Draw(NewCanvas(70, 5), Rect{W: 70, H: 5})
	if result := edit.ConsumeKey(KeyEvent{Key: "down"}); !result.Consumed || !edit.popoverFocusSet {
		t.Fatalf("Down did not focus the auto-shown toolbar: %+v", result)
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "b", Text: "b"}); !result.Consumed || !doc.Lines[0].Spans[0].Style.Bold {
		t.Fatalf("focused toolbar hotkey result/style = %+v/%+v", result, doc.Lines[0].Spans[0].Style)
	}
}

func TestRichTextEditPopoverLetterTypesInUnfocusedAutoBar(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "word"}}}}}
	edit := NewRichTextEdit(doc)
	edit.SetSelection(RichPosition{}, RichPosition{Offset: 4})
	edit.Draw(NewCanvas(70, 5), Rect{W: 70, H: 5})
	if result := edit.ConsumeKey(KeyEvent{Key: "b", Text: "b"}); !result.Consumed {
		t.Fatal("editor did not consume ordinary typing")
	}
	if got := doc.ToPlainText(); got != "b" {
		t.Fatalf("typed text = %q, want replacement text", got)
	}
}

func TestRichTextEditFocusedBarReturnsTypingAfterDismissal(t *testing.T) {
	for _, dismissal := range []string{"esc"} {
		doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "word"}}}}}
		edit := NewRichTextEdit(doc)
		edit.SetSelection(RichPosition{}, RichPosition{Offset: 4})
		edit.Draw(NewCanvas(70, 5), Rect{W: 70, H: 5})
		if result := edit.ConsumeKey(KeyEvent{Key: "up"}); !result.Consumed || !edit.popoverFocusSet {
			t.Fatalf("Up did not focus auto-shown toolbar before %s: %+v", dismissal, result)
		}
		if result := edit.ConsumeKey(KeyEvent{Key: dismissal}); !result.Consumed || edit.popoverFocusSet {
			t.Fatalf("%s did not dismiss focused toolbar: %+v focus=%v", dismissal, result, edit.popoverFocusSet)
		}
		if result := edit.ConsumeKey(KeyEvent{Key: "b", Text: "b"}); !result.Consumed {
			t.Fatalf("b after %s was not typed", dismissal)
		}
		if got := doc.ToPlainText(); got != "bword" {
			t.Fatalf("text after %s and b = %q, want typing at the retained cursor", dismissal, got)
		}
	}
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "word"}}}}}
	edit := NewRichTextEdit(doc)
	edit.SetSelection(RichPosition{}, RichPosition{Offset: 4})
	edit.Draw(NewCanvas(70, 5), Rect{W: 70, H: 5})
	if result := edit.ConsumeKey(KeyEvent{Key: "down"}); !result.Consumed || !edit.popoverFocusSet {
		t.Fatalf("Down did not focus auto-shown toolbar: %+v", result)
	}
	_, _, _, pointer, _ := edit.popoverGeometry(edit.lastRect, edit.documentLines())
	away := "up"
	if pointer == SpeccedDefaults.RichTextEdit.PointerDownGlyph {
		away = "down"
	}
	if result := edit.ConsumeKey(KeyEvent{Key: away}); !result.Consumed || edit.popoverFocusSet {
		t.Fatalf("arrow away %s did not dismiss focused toolbar: %+v focus=%v pointer=%q", away, result, edit.popoverFocusSet, pointer)
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "b", Text: "b"}); !result.Consumed || doc.ToPlainText() != "bword" {
		t.Fatalf("b after arrow-away was not typed: %+v text=%q", result, doc.ToPlainText())
	}
}

func TestRichTextEditPopoverFrameMouseHitUsesNewGeometry(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "selected"}}}}})
	edit.PopoverDefaults.AutoPopupOnSelection = false
	edit.SetSelection(RichPosition{}, RichPosition{Offset: 8})
	edit.ConsumeKey(KeyEvent{Key: "ctrl-space"})
	edit.Draw(NewCanvas(70, 5), Rect{W: 70, H: 5})
	button := edit.popoverButtons[popoverActionIndex("Box")]
	if result := edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: button.rect.X + button.rect.W - 1, Y: button.rect.Y}); !result.Consumed {
		t.Fatal("Frame click was not consumed")
	}
	if edit.popoverSubmenu != "Box" {
		t.Fatalf("Frame click opened submenu %q, want Box", edit.popoverSubmenu)
	}
}

func styleForAction(style Style, action string) bool {
	switch action {
	case "B":
		return style.Bold
	case "I":
		return style.Italic
	case "U":
		return style.Underline
	case "S":
		return style.Strike
	default:
		return false
	}
}
