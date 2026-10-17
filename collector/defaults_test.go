package collector

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
	"ubunatic.com/loom/spec"
)

func TestDefaultRetentionFromSpec(t *testing.T) {
	data, err := spec.DefaultsYAML()
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		Collector map[string]string `yaml:"collector"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	want, err := time.ParseDuration(raw.Collector["default_retention"])
	if err != nil || want <= 0 {
		t.Fatalf("invalid spec retention: %v, %v", want, err)
	}
	got, err := (Spec{}).RetentionDuration()
	if err != nil || got != want || DefaultRetention != want {
		t.Fatalf("default retention = %v, exported = %v, want %v: %v", got, DefaultRetention, want, err)
	}
	// The default controls actual eviction at its boundary.
	history, err := NewHistory(got)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(1000, 0)
	history.Append(Record{At: start, Data: []byte("old")})
	history.Append(Record{At: start.Add(want), Data: []byte("boundary")})
	if len(history.Snapshot()) != 2 {
		t.Fatal("history evicted a record at the retention boundary")
	}
	history.Append(Record{At: start.Add(want + time.Nanosecond), Data: []byte("new")})
	if records := history.Snapshot(); len(records) != 2 || string(records[0].Data) != "boundary" {
		t.Fatalf("history did not apply specced retention: %+v", records)
	}
}
