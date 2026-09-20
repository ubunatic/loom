// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"sort"
	"sync"
	"time"
)

// MetricSample is one observation in a named metric series.
//
// Samples are kept in publish order. At is the timestamp supplied to Publish;
// it is not used to reorder observations.
type MetricSample struct {
	Value float64
	At    time.Time
}

// MetricPoint is an alias for MetricSample, useful when referring to a series
// as a sequence of points.
type MetricPoint = MetricSample

// MetricSnapshot is a detached view of all metric series at one instant.
// The map and all sample slices are copied from the store. Mutating them does
// not change the store or any other snapshot.
type MetricSnapshot struct {
	Series map[string][]MetricSample
}

// Get returns a copy of the named series. It returns nil when the series is
// not present.
func (s MetricSnapshot) Get(name string) []MetricSample {
	return cloneMetricSamples(s.Series[name])
}

// Values returns the values in the named series, oldest first.
func (s MetricSnapshot) Values(name string) []float64 {
	samples := s.Series[name]
	values := make([]float64, len(samples))
	for i, sample := range samples {
		values[i] = sample.Value
	}
	return values
}

// Latest returns the newest published sample for name.
func (s MetricSnapshot) Latest(name string) (MetricSample, bool) {
	samples := s.Series[name]
	if len(samples) == 0 {
		return MetricSample{}, false
	}
	return samples[len(samples)-1], true
}

// Names returns all series names in stable lexical order.
func (s MetricSnapshot) Names() []string {
	names := make([]string, 0, len(s.Series))
	for name := range s.Series {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Clone returns an independent copy of the snapshot.
func (s MetricSnapshot) Clone() MetricSnapshot {
	return cloneMetricSnapshot(s)
}

type metricStoreConfig struct {
	capacity int
}

// MetricStoreOption configures a MetricStore.
type MetricStoreOption func(*metricStoreConfig)

// Retention limits each named series to the newest capacity samples. A
// non-positive capacity leaves series unbounded.
func Retention(capacity int) MetricStoreOption {
	return func(config *metricStoreConfig) {
		config.capacity = capacity
	}
}

// Capacity is an alias for Retention for callers that describe the option in
// terms of the backing ring capacity.
func Capacity(capacity int) MetricStoreOption { return Retention(capacity) }

// WithRetention is an explicit alias for Retention.
func WithRetention(capacity int) MetricStoreOption { return Retention(capacity) }

// MetricStore stores named rolling metric series. Publishing and snapshotting
// may safely happen concurrently from different goroutines.
type MetricStore struct {
	mu       sync.RWMutex
	capacity int
	series   map[string][]MetricSample
}

// NewMetricStore creates an empty, concurrency-safe metric store.
func NewMetricStore(opts ...MetricStoreOption) *MetricStore {
	config := metricStoreConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&config)
		}
	}
	return &MetricStore{capacity: config.capacity, series: make(map[string][]MetricSample)}
}

// Publish appends a sample to the named series. Empty names are ignored. When
// retention is configured, older samples are discarded independently for each
// series.
func (s *MetricStore) Publish(name string, value float64, at time.Time) {
	if s == nil || name == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.series == nil {
		s.series = make(map[string][]MetricSample)
	}
	values := append(s.series[name], MetricSample{Value: value, At: at})
	if s.capacity > 0 && len(values) > s.capacity {
		values = values[len(values)-s.capacity:]
	}
	s.series[name] = values
}

// PublishNow appends a sample using time.Now as its timestamp.
func (s *MetricStore) PublishNow(name string, value float64) {
	s.Publish(name, value, time.Now())
}

// Snapshot returns a deeply detached view of every series. Later publishes do
// not alter the returned map or slices.
func (s *MetricStore) Snapshot() MetricSnapshot {
	if s == nil {
		return MetricSnapshot{Series: make(map[string][]MetricSample)}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneMetricSnapshot(MetricSnapshot{Series: s.series})
}

// Get returns a detached copy of a named series from the current store state.
func (s *MetricStore) Get(name string) []MetricSample {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneMetricSamples(s.series[name])
}

// Values returns the values in a named series from the current store state.
func (s *MetricStore) Values(name string) []float64 {
	samples := s.Get(name)
	values := make([]float64, len(samples))
	for i, sample := range samples {
		values[i] = sample.Value
	}
	return values
}

// Latest returns the newest published sample for name.
func (s *MetricStore) Latest(name string) (MetricSample, bool) {
	if s == nil {
		return MetricSample{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	samples := s.series[name]
	if len(samples) == 0 {
		return MetricSample{}, false
	}
	return samples[len(samples)-1], true
}

// Names returns all currently known series names in stable lexical order.
func (s *MetricStore) Names() []string { return s.Snapshot().Names() }

func cloneMetricSamples(samples []MetricSample) []MetricSample {
	if samples == nil {
		return nil
	}
	return append([]MetricSample(nil), samples...)
}

func cloneMetricSnapshot(snapshot MetricSnapshot) MetricSnapshot {
	series := make(map[string][]MetricSample, len(snapshot.Series))
	for name, samples := range snapshot.Series {
		series[name] = cloneMetricSamples(samples)
	}
	return MetricSnapshot{Series: series}
}
