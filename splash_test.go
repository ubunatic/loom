// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"context"
	"testing"
	"time"
)

func TestSplashControllerProgression(t *testing.T) {
	cfg := SplashConfig{
		Tasks: []ProviderTask{
			{Name: "mic", Symbol: "●", Duration: 10 * time.Millisecond},
			{Name: "claude", Symbol: "✳", Duration: 10 * time.Millisecond},
		},
		TickInterval: 5 * time.Millisecond,
		HoldDuration: 5 * time.Millisecond,
	}

	sc := NewSplashController(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	sc.Start(ctx)

	select {
	case <-sc.Done():
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for splash controller completion")
	}

	snap := sc.Snapshot()
	if snap.State != SplashCompleted {
		t.Errorf("state = %v, want SplashCompleted", snap.State)
	}
	if snap.Progress != 100.0 {
		t.Errorf("progress = %v, want 100.0", snap.Progress)
	}
	for i, pill := range snap.Pills {
		if pill.State != ProviderDone {
			t.Errorf("pill %d (%s) state = %v, want ProviderDone", i, pill.Name, pill.State)
		}
	}
}

func TestSplashControllerDismissal(t *testing.T) {
	cfg := SplashConfig{
		Tasks: []ProviderTask{
			{Name: "mic", Symbol: "●", Duration: 500 * time.Millisecond},
			{Name: "claude", Symbol: "✳", Duration: 500 * time.Millisecond},
		},
		TickInterval: 10 * time.Millisecond,
	}

	sc := NewSplashController(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sc.Start(ctx)

	// Simulate Esc key after short delay
	time.Sleep(20 * time.Millisecond)
	handled := sc.HandleKey(KeyEvent{Key: "esc"})
	if !handled {
		t.Errorf("HandleKey(esc) = false, want true")
	}

	select {
	case <-sc.Done():
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for splash dismissal")
	}

	snap := sc.Snapshot()
	if snap.State != SplashDismissed {
		t.Errorf("state = %v, want SplashDismissed", snap.State)
	}
	if !snap.Dismissed {
		t.Errorf("snap.Dismissed = false, want true")
	}
}

func TestSplashControllerDismissalWithQKey(t *testing.T) {
	sc := NewSplashController(SplashConfig{})
	handled := sc.HandleKey(KeyEvent{Text: "q"})
	if !handled {
		t.Errorf("HandleKey(text 'q') = false, want true")
	}
	if !sc.Snapshot().Dismissed {
		t.Errorf("state not dismissed after 'q'")
	}
}

