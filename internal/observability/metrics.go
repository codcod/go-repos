// Package observability provides metrics collection and monitoring capabilities
package observability

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/codcod/repos/internal/core"
)

// MetricsCollector collects and manages metrics for the repos tool
type MetricsCollector struct {
	mu              sync.RWMutex
	counters        map[string]int64
	gauges          map[string]float64
	histograms      map[string]*Histogram
	timers          map[string]*Timer
	startTime       time.Time
	checkResults    []core.CheckResult
	repositoryCount int
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		counters:   make(map[string]int64),
		gauges:     make(map[string]float64),
		histograms: make(map[string]*Histogram),
		timers:     make(map[string]*Timer),
		startTime:  time.Now(),
	}
}

// Counter operations
func (mc *MetricsCollector) IncrementCounter(name string) {
	mc.AddToCounter(name, 1)
}

func (mc *MetricsCollector) AddToCounter(name string, value int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.counters[name] += value
}

func (mc *MetricsCollector) GetCounter(name string) int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.counters[name]
}

// Gauge operations
func (mc *MetricsCollector) SetGauge(name string, value float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.gauges[name] = value
}

func (mc *MetricsCollector) GetGauge(name string) float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.gauges[name]
}

// Histogram operations
func (mc *MetricsCollector) RecordHistogram(name string, value float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.histograms[name] == nil {
		mc.histograms[name] = NewHistogram()
	}
	mc.histograms[name].Record(value)
}

func (mc *MetricsCollector) GetHistogram(name string) *Histogram {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.histograms[name]
}

// Timer operations
func (mc *MetricsCollector) StartTimer(name string) *Timer {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	timer := NewTimer()
	mc.timers[name] = timer
	return timer
}

func (mc *MetricsCollector) GetTimer(name string) *Timer {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.timers[name]
}

// Repository and check tracking
func (mc *MetricsCollector) RecordRepositoryProcessed() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.counters["repositories_processed"]++
	mc.repositoryCount++
}

func (mc *MetricsCollector) RecordCheckResult(result core.CheckResult) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.checkResults = append(mc.checkResults, result)

	// Use internal methods to avoid mutex deadlock
	mc.counters["checks_executed"]++
	mc.counters[fmt.Sprintf("checks_%s", result.Status)]++

	// Record histogram data directly
	scoreName := "check_scores"
	if mc.histograms[scoreName] == nil {
		mc.histograms[scoreName] = NewHistogram()
	}
	mc.histograms[scoreName].Record(float64(result.Score))

	durationName := "check_duration_ms"
	if mc.histograms[durationName] == nil {
		mc.histograms[durationName] = NewHistogram()
	}
	mc.histograms[durationName].Record(float64(result.Duration.Nanoseconds()) / 1e6)
}

// Summary operations
func (mc *MetricsCollector) GetSummary() MetricsSummary {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	totalDuration := time.Since(mc.startTime)

	summary := MetricsSummary{
		StartTime:         mc.startTime,
		TotalDuration:     totalDuration,
		RepositoriesCount: mc.repositoryCount,
		ChecksExecuted:    mc.counters["checks_executed"],
		Counters:          make(map[string]int64),
		Gauges:            make(map[string]float64),
		Histograms:        make(map[string]HistogramSummary),
	}

	// Copy counters
	for k, v := range mc.counters {
		summary.Counters[k] = v
	}

	// Copy gauges
	for k, v := range mc.gauges {
		summary.Gauges[k] = v
	}

	// Copy histogram summaries
	for k, v := range mc.histograms {
		if v != nil {
			summary.Histograms[k] = v.Summary()
		}
	}

	// Calculate rates
	if totalDuration.Seconds() > 0 {
		summary.RepositoriesPerSecond = float64(mc.repositoryCount) / totalDuration.Seconds()
		summary.ChecksPerSecond = float64(mc.counters["checks_executed"]) / totalDuration.Seconds()
	}

	return summary
}

// Print metrics to output
func (mc *MetricsCollector) PrintSummary() {
	summary := mc.GetSummary()

	fmt.Printf("\n=== Metrics Summary ===\n")
	fmt.Printf("Total Duration: %v\n", summary.TotalDuration)
	fmt.Printf("Repositories Processed: %d (%.2f/sec)\n", summary.RepositoriesCount, summary.RepositoriesPerSecond)
	fmt.Printf("Checks Executed: %d (%.2f/sec)\n", summary.ChecksExecuted, summary.ChecksPerSecond)

	fmt.Printf("\nCounters:\n")
	for name, value := range summary.Counters {
		fmt.Printf("  %s: %d\n", name, value)
	}

	if len(summary.Gauges) > 0 {
		fmt.Printf("\nGauges:\n")
		for name, value := range summary.Gauges {
			fmt.Printf("  %s: %.2f\n", name, value)
		}
	}

	if len(summary.Histograms) > 0 {
		fmt.Printf("\nHistograms:\n")
		for name, hist := range summary.Histograms {
			fmt.Printf("  %s: count=%d, min=%.2f, max=%.2f, mean=%.2f, p95=%.2f\n",
				name, hist.Count, hist.Min, hist.Max, hist.Mean, hist.P95)
		}
	}
}

// MetricsSummary contains a snapshot of all metrics
type MetricsSummary struct {
	StartTime             time.Time                   `json:"start_time"`
	TotalDuration         time.Duration               `json:"total_duration"`
	RepositoriesCount     int                         `json:"repositories_count"`
	RepositoriesPerSecond float64                     `json:"repositories_per_second"`
	ChecksExecuted        int64                       `json:"checks_executed"`
	ChecksPerSecond       float64                     `json:"checks_per_second"`
	Counters              map[string]int64            `json:"counters"`
	Gauges                map[string]float64          `json:"gauges"`
	Histograms            map[string]HistogramSummary `json:"histograms"`
}

// Histogram tracks distribution of values
type Histogram struct {
	values []float64
	count  int64
	sum    float64
}

// NewHistogram creates a new histogram
func NewHistogram() *Histogram {
	return &Histogram{
		values: make([]float64, 0),
	}
}

// Record adds a value to the histogram
func (h *Histogram) Record(value float64) {
	h.values = append(h.values, value)
	h.count++
	h.sum += value
}

// Summary returns statistical summary of the histogram
func (h *Histogram) Summary() HistogramSummary {
	if h.count == 0 {
		return HistogramSummary{}
	}

	// Sort values for percentile calculation
	sorted := make([]float64, len(h.values))
	copy(sorted, h.values)
	sort.Float64s(sorted)

	summary := HistogramSummary{
		Count: h.count,
		Sum:   h.sum,
		Mean:  h.sum / float64(h.count),
		Min:   sorted[0],
		Max:   sorted[len(sorted)-1],
	}

	// Calculate percentiles
	if len(sorted) > 0 {
		summary.P50 = percentile(sorted, 0.5)
		summary.P95 = percentile(sorted, 0.95)
		summary.P99 = percentile(sorted, 0.99)
	}

	return summary
}

// HistogramSummary contains statistical information about a histogram
type HistogramSummary struct {
	Count int64   `json:"count"`
	Sum   float64 `json:"sum"`
	Mean  float64 `json:"mean"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	P50   float64 `json:"p50"`
	P95   float64 `json:"p95"`
	P99   float64 `json:"p99"`
}

// Timer tracks execution time
type Timer struct {
	startTime time.Time
	endTime   *time.Time
	duration  time.Duration
}

// NewTimer creates a new timer and starts it
func NewTimer() *Timer {
	return &Timer{
		startTime: time.Now(),
	}
}

// Stop stops the timer and records the duration
func (t *Timer) Stop() time.Duration {
	now := time.Now()
	t.endTime = &now
	t.duration = now.Sub(t.startTime)
	return t.duration
}

// Duration returns the elapsed time (stopping the timer if not already stopped)
func (t *Timer) Duration() time.Duration {
	if t.endTime == nil {
		return time.Since(t.startTime)
	}
	return t.duration
}

// percentile calculates the percentile value from sorted data
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}

	index := p * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// Helper functions for common metrics patterns

// MeasureOperation measures the duration of an operation
func (mc *MetricsCollector) MeasureOperation(name string, fn func() error) error {
	timer := mc.StartTimer(name)
	defer func() {
		duration := timer.Stop()
		mc.RecordHistogram(name+"_duration_ms", float64(duration.Nanoseconds())/1e6)
	}()

	err := fn()
	if err != nil {
		mc.IncrementCounter(name + "_errors")
	} else {
		mc.IncrementCounter(name + "_success")
	}

	return err
}
