package observability

import (
	"testing"
	"time"

	"github.com/codcod/repos/internal/core"
)

func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()
	if collector == nil {
		t.Fatal("NewMetricsCollector should not return nil")
	}

	if collector.counters == nil {
		t.Error("Counters map should be initialized")
	}

	if collector.gauges == nil {
		t.Error("Gauges map should be initialized")
	}

	if collector.histograms == nil {
		t.Error("Histograms map should be initialized")
	}

	if collector.timers == nil {
		t.Error("Timers map should be initialized")
	}

	if collector.startTime.IsZero() {
		t.Error("Start time should be set")
	}
}

func TestCounterOperations(t *testing.T) {
	collector := NewMetricsCollector()

	// Test initial value
	if collector.GetCounter("test") != 0 {
		t.Error("Initial counter value should be 0")
	}

	// Test increment
	collector.IncrementCounter("test")
	if collector.GetCounter("test") != 1 {
		t.Errorf("Expected counter value 1, got %d", collector.GetCounter("test"))
	}

	// Test add
	collector.AddToCounter("test", 5)
	if collector.GetCounter("test") != 6 {
		t.Errorf("Expected counter value 6, got %d", collector.GetCounter("test"))
	}
}

func TestGaugeOperations(t *testing.T) {
	collector := NewMetricsCollector()

	// Test initial value
	if collector.GetGauge("test") != 0 {
		t.Error("Initial gauge value should be 0")
	}

	// Test set
	collector.SetGauge("test", 42.5)
	if collector.GetGauge("test") != 42.5 {
		t.Errorf("Expected gauge value 42.5, got %f", collector.GetGauge("test"))
	}

	// Test overwrite
	collector.SetGauge("test", 100.0)
	if collector.GetGauge("test") != 100.0 {
		t.Errorf("Expected gauge value 100.0, got %f", collector.GetGauge("test"))
	}
}

func TestHistogramOperations(t *testing.T) {
	collector := NewMetricsCollector()

	// Test initial value
	hist := collector.GetHistogram("test")
	if hist != nil {
		t.Error("Initial histogram should be nil")
	}

	// Test record
	collector.RecordHistogram("test", 10.0)
	collector.RecordHistogram("test", 20.0)
	collector.RecordHistogram("test", 30.0)

	hist = collector.GetHistogram("test")
	if hist == nil {
		t.Fatal("Histogram should exist after recording values")
	}

	summary := hist.Summary()
	if summary.Count != 3 {
		t.Errorf("Expected count 3, got %d", summary.Count)
	}

	if summary.Mean != 20.0 {
		t.Errorf("Expected mean 20.0, got %f", summary.Mean)
	}

	if summary.Min != 10.0 {
		t.Errorf("Expected min 10.0, got %f", summary.Min)
	}

	if summary.Max != 30.0 {
		t.Errorf("Expected max 30.0, got %f", summary.Max)
	}
}

func TestTimerOperations(t *testing.T) {
	collector := NewMetricsCollector()

	timer := collector.StartTimer("test")
	if timer == nil {
		t.Fatal("StartTimer should return a timer")
	}

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	duration := timer.Stop()
	if duration < 10*time.Millisecond {
		t.Errorf("Duration should be at least 10ms, got %v", duration)
	}

	retrievedTimer := collector.GetTimer("test")
	if retrievedTimer != timer {
		t.Error("Retrieved timer should be the same as the started timer")
	}
}

func TestRecordRepositoryProcessed(t *testing.T) {
	collector := NewMetricsCollector()

	initial := collector.GetCounter("repositories_processed")
	collector.RecordRepositoryProcessed()

	if collector.GetCounter("repositories_processed") != initial+1 {
		t.Error("Repository processed counter should increment")
	}

	if collector.repositoryCount != 1 {
		t.Error("Repository count should be 1")
	}
}

func TestRecordCheckResult(t *testing.T) {
	collector := NewMetricsCollector()

	result := core.CheckResult{
		Name:     "test-check",
		Status:   core.StatusHealthy,
		Score:    85,
		Duration: 100 * time.Millisecond,
	}

	collector.RecordCheckResult(result)

	if collector.GetCounter("checks_executed") != 1 {
		t.Error("Checks executed counter should be 1")
	}

	if collector.GetCounter("checks_healthy") != 1 {
		t.Error("Checks healthy counter should be 1")
	}

	scoreHist := collector.GetHistogram("check_scores")
	if scoreHist == nil {
		t.Fatal("Check scores histogram should exist")
	}

	summary := scoreHist.Summary()
	if summary.Count != 1 {
		t.Error("Check scores histogram should have 1 entry")
	}

	if summary.Mean != 85.0 {
		t.Errorf("Expected mean score 85.0, got %f", summary.Mean)
	}
}

func TestGetSummary(t *testing.T) {
	collector := NewMetricsCollector()

	// Add some data
	collector.IncrementCounter("test_counter")
	collector.SetGauge("test_gauge", 42.0)
	collector.RecordHistogram("test_hist", 10.0)
	collector.RecordRepositoryProcessed()

	// Wait a bit to ensure duration > 0
	time.Sleep(1 * time.Millisecond)

	summary := collector.GetSummary()

	if summary.StartTime.IsZero() {
		t.Error("Summary should have start time")
	}

	if summary.TotalDuration <= 0 {
		t.Error("Summary should have positive total duration")
	}

	if summary.RepositoriesCount != 1 {
		t.Errorf("Expected repositories count 1, got %d", summary.RepositoriesCount)
	}

	if summary.Counters["test_counter"] != 1 {
		t.Error("Summary should include test counter")
	}

	if summary.Gauges["test_gauge"] != 42.0 {
		t.Error("Summary should include test gauge")
	}

	if _, exists := summary.Histograms["test_hist"]; !exists {
		t.Error("Summary should include test histogram")
	}

	if summary.RepositoriesPerSecond <= 0 {
		t.Error("Should calculate repositories per second")
	}
}

func TestMeasureOperation(t *testing.T) {
	collector := NewMetricsCollector()

	// Test successful operation
	err := collector.MeasureOperation("test_op", func() error {
		time.Sleep(1 * time.Millisecond)
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if collector.GetCounter("test_op_success") != 1 {
		t.Error("Success counter should be incremented")
	}

	if collector.GetCounter("test_op_errors") != 0 {
		t.Error("Error counter should be 0")
	}

	durationHist := collector.GetHistogram("test_op_duration_ms")
	if durationHist == nil {
		t.Fatal("Duration histogram should exist")
	}

	summary := durationHist.Summary()
	if summary.Count != 1 {
		t.Error("Duration histogram should have 1 entry")
	}

	if summary.Mean <= 0 {
		t.Error("Duration should be positive")
	}
}

func TestNewHistogram(t *testing.T) {
	hist := NewHistogram()
	if hist == nil {
		t.Fatal("NewHistogram should not return nil")
	}

	if hist.count != 0 {
		t.Error("Initial count should be 0")
	}

	if hist.sum != 0 {
		t.Error("Initial sum should be 0")
	}

	if len(hist.values) != 0 {
		t.Error("Initial values slice should be empty")
	}
}

func TestHistogramRecord(t *testing.T) {
	hist := NewHistogram()

	hist.Record(10.0)
	hist.Record(20.0)
	hist.Record(30.0)

	if hist.count != 3 {
		t.Errorf("Expected count 3, got %d", hist.count)
	}

	if hist.sum != 60.0 {
		t.Errorf("Expected sum 60.0, got %f", hist.sum)
	}

	if len(hist.values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(hist.values))
	}
}

func TestHistogramSummary(t *testing.T) {
	hist := NewHistogram()

	// Test empty histogram
	summary := hist.Summary()
	if summary.Count != 0 {
		t.Error("Empty histogram should have count 0")
	}

	// Add values
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, v := range values {
		hist.Record(v)
	}

	summary = hist.Summary()
	if summary.Count != 10 {
		t.Errorf("Expected count 10, got %d", summary.Count)
	}

	if summary.Mean != 5.5 {
		t.Errorf("Expected mean 5.5, got %f", summary.Mean)
	}

	if summary.Min != 1.0 {
		t.Errorf("Expected min 1.0, got %f", summary.Min)
	}

	if summary.Max != 10.0 {
		t.Errorf("Expected max 10.0, got %f", summary.Max)
	}

	if summary.P50 != 5.5 {
		t.Errorf("Expected P50 5.5, got %f", summary.P50)
	}
}

func TestNewTimer(t *testing.T) {
	timer := NewTimer()
	if timer == nil {
		t.Fatal("NewTimer should not return nil")
	}

	if timer.startTime.IsZero() {
		t.Error("Timer should have start time set")
	}

	if timer.endTime != nil {
		t.Error("Timer should not have end time set initially")
	}
}

func TestTimerStop(t *testing.T) {
	timer := NewTimer()

	time.Sleep(1 * time.Millisecond)

	duration := timer.Stop()
	if duration <= 0 {
		t.Error("Duration should be positive")
	}

	if timer.endTime == nil {
		t.Error("End time should be set after stop")
	}

	if timer.duration != duration {
		t.Error("Stored duration should match returned duration")
	}
}

func TestTimerDuration(t *testing.T) {
	timer := NewTimer()

	time.Sleep(1 * time.Millisecond)

	// Test duration before stop
	duration1 := timer.Duration()
	if duration1 <= 0 {
		t.Error("Duration should be positive before stop")
	}

	// Stop timer
	timer.Stop()

	// Test duration after stop
	duration2 := timer.Duration()
	if duration2 != timer.duration {
		t.Error("Duration should return stored duration after stop")
	}
}

func TestPercentile(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		p        float64
		expected float64
	}{
		{0.0, 1.0},
		{0.5, 5.5},
		{1.0, 10.0},
		{0.25, 3.25},
		{0.75, 7.75},
	}

	for _, test := range tests {
		result := percentile(values, test.p)
		if result != test.expected {
			t.Errorf("percentile(%f) = %f, expected %f", test.p, result, test.expected)
		}
	}

	// Test empty slice
	empty := []float64{}
	if percentile(empty, 0.5) != 0 {
		t.Error("Percentile of empty slice should be 0")
	}
}

func TestPrintSummary(t *testing.T) {
	collector := NewMetricsCollector()

	// Add some data
	collector.IncrementCounter("test_counter")
	collector.SetGauge("test_gauge", 42.0)
	collector.RecordHistogram("test_hist", 10.0)
	collector.RecordRepositoryProcessed()

	// This is mainly a smoke test - we can't easily capture the output
	// In a real implementation, you might want to make the output configurable
	collector.PrintSummary()
}

func BenchmarkCounterOperations(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.IncrementCounter("benchmark")
	}
}

func BenchmarkHistogramRecord(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordHistogram("benchmark", float64(i))
	}
}

func BenchmarkTimerOperations(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		timer := collector.StartTimer("benchmark")
		timer.Stop()
	}
}
