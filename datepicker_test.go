// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
	"time"
)

func TestDatePickerNavigationBoundsAndClock(t *testing.T) {
	selected := time.Date(2024, time.January, 31, 0, 0, 0, 0, time.UTC)
	picker := NewDatePicker(&selected)
	picker.Now = func() time.Time { return time.Date(2024, time.January, 20, 0, 0, 0, 0, time.UTC) }
	picker.Min = time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC)
	picker.Max = time.Date(2024, time.March, 2, 0, 0, 0, 0, time.UTC)
	rows := strings.Join(Render(picker, 64, 11), "\n")
	if !strings.Contains(rows, "January 2024") || !strings.Contains(rows, "*20") {
		t.Fatalf("missing injected-clock month or today marker: %s", rows)
	}
	picker.HandleKey(KeyEvent{Key: "right"})
	if got := picker.Cursor().Format("2006-01-02"); got != "2024-02-01" {
		t.Fatalf("right at month edge = %s", got)
	}
	picker.HandleKey(KeyEvent{Key: "pgdown"})
	if picker.DisplayMonth() != time.March {
		t.Fatalf("PgDn month = %v", picker.DisplayMonth())
	}
	picker.HandleKey(KeyEvent{Key: "ctrl-right"})
	if got := picker.Cursor().Format("2006-01-02"); got != "2024-03-02" {
		t.Fatalf("year navigation escaped maximum: %s", got)
	}
	picker.setCursor(time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC))
	picker.HandleKey(KeyEvent{Key: "left"})
	if got := picker.Cursor().Format("2006-01-02"); got != "2024-01-30" {
		t.Fatalf("left navigation escaped minimum: %s", got)
	}
}

func TestDatePickerPageKeyAliases(t *testing.T) {
	selected := time.Date(2024, time.June, 15, 0, 0, 0, 0, time.UTC)
	picker := NewDatePicker(&selected)
	for _, key := range []string{"pgdn", "pagedown"} {
		picker.ConsumeKey(KeyEvent{Key: key})
	}
	if got := picker.Cursor().Format("2006-01-02"); got != "2024-08-15" {
		t.Fatalf("page aliases cursor = %s", got)
	}
	picker.ConsumeKey(KeyEvent{Key: "pageup"})
	if got := picker.Cursor().Format("2006-01-02"); got != "2024-07-15" {
		t.Fatalf("pageup alias cursor = %s", got)
	}
}

func TestDatePickerTypedISOCommitAndMouse(t *testing.T) {
	selected := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	picker := NewDatePicker(&selected)
	for _, r := range "2024-02-29" {
		picker.HandleKey(KeyEvent{Text: string(r)})
	}
	picker.HandleKey(KeyEvent{Key: "enter"})
	if got := selected.Format("2006-01-02"); got != "2024-02-29" {
		t.Fatalf("typed commit = %s", got)
	}
	picker.Draw(NewCanvas(23, 9), Rect{W: 23, H: 9})
	if !picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 3, Y: 4}).Consumed {
		t.Fatal("date click not consumed")
	}
	if got := selected.Format("2006-01-02"); got == "2024-02-29" {
		t.Fatal("mouse click did not select calendar cell")
	}
}

func TestDatePickerWeekStartAndInvalidTypedDate(t *testing.T) {
	selected := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	picker := NewDatePicker(&selected)
	picker.WeekStart = time.Sunday
	for _, r := range "2023-02-29" {
		picker.HandleKey(KeyEvent{Text: string(r)})
	}
	picker.HandleKey(KeyEvent{Key: "enter"})
	if selected.Format("2006-01-02") != "2024-01-01" {
		t.Fatal("invalid ISO date changed selection")
	}
	rows := strings.Join(Render(picker, 23, 9), "\n")
	if !strings.Contains(rows, "Su Mo Tu We Th Fr Sa") {
		t.Fatalf("week start not applied: %s", rows)
	}
}

func TestDatePickerInitializesEmptyValueWithInjectedClock(t *testing.T) {
	var selected time.Time
	picker := NewDatePickerWithClock(&selected, func() time.Time {
		return time.Date(2024, time.April, 9, 23, 45, 0, 0, time.FixedZone("test", 3600))
	})
	if got := selected.Format("2006-01-02 15:04"); got != "2024-04-09 00:00" {
		t.Fatalf("empty value initialization = %s", got)
	}
	if picker.Cursor().Day() != 9 {
		t.Fatalf("initial cursor = %v", picker.Cursor())
	}
}
