// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"sync"
	"testing"
	"time"
)

func TestMetricStoreRetentionAndSnapshot(t *testing.T) {
	store := NewMetricStore(Retention(3))
	at := time.Unix(100, 0)
	for i := 0; i < 5; i++ {
		store.Publish("cpu", float64(i), at.Add(time.Duration(i)*time.Second))
	}
	store.Publish("ram", 42, at)

	snapshot := store.Snapshot()
	cpu := snapshot.Get("cpu")
	if len(cpu) != 3 {
		t.Fatalf("retained %d cpu samples, want 3", len(cpu))
	}
	for i, want := range []float64{2, 3, 4} {
		if cpu[i].Value != want {
			t.Errorf("cpu[%d] = %v, want %v", i, cpu[i].Value, want)
		}
	}
	if got := snapshot.Values("ram"); len(got) != 1 || got[0] != 42 {
		t.Errorf("ram values = %v, want [42]", got)
	}

	// A snapshot and each returned series are detached from the store.
	cpu[0].Value = 999
	snapshot.Series["cpu"][1].Value = 888
	store.Publish("cpu", 5, at.Add(5*time.Second))
	if got := store.Values("cpu"); got[0] != 3 || got[1] != 4 || got[2] != 5 {
		t.Errorf("store changed through snapshot mutation: %v", got)
	}
	if got := snapshot.Values("cpu"); got[0] != 2 || got[1] != 888 || got[2] != 4 {
		t.Errorf("snapshot did not remain internally consistent after direct map mutation: %v", got)
	}
}

func TestMetricStoreConcurrentPublishAndSnapshot(t *testing.T) {
	store := NewMetricStore(Retention(64))
	const writers = 8
	const samples = 500
	var wg sync.WaitGroup
	for writer := 0; writer < writers; writer++ {
		writer := writer
		wg.Add(1)
		go func() {
			defer wg.Done()
			for sample := 0; sample < samples; sample++ {
				store.Publish("shared", float64(writer*samples+sample), time.Unix(int64(sample), 0))
				store.Publish("series", float64(sample), time.Unix(int64(sample), 0))
			}
		}()
	}
	for i := 0; i < 200; i++ {
		snapshot := store.Snapshot()
		if len(snapshot.Get("shared")) > 64 || len(snapshot.Get("series")) > 64 {
			t.Fatalf("snapshot exceeded retention: shared=%d series=%d", len(snapshot.Get("shared")), len(snapshot.Get("series")))
		}
	}
	wg.Wait()
	if got := len(store.Get("shared")); got != 64 {
		t.Errorf("shared retention = %d, want 64", got)
	}
	if got := len(store.Get("series")); got != 64 {
		t.Errorf("series retention = %d, want 64", got)
	}
}

func TestMetricStoreLatestAndNames(t *testing.T) {
	store := NewMetricStore()
	if _, ok := store.Latest("missing"); ok {
		t.Fatal("missing series reported a latest sample")
	}
	store.Publish("zeta", 1, time.Time{})
	store.Publish("alpha", 2, time.Time{})
	sample, ok := store.Latest("zeta")
	if !ok || sample.Value != 1 {
		t.Fatalf("Latest(zeta) = %+v, %v", sample, ok)
	}
	if got, want := store.Names(), []string{"alpha", "zeta"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("Names() = %v, want %v", got, want)
	}
}
