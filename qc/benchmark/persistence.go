// Package benchmark provides result persistence and comparison capabilities
package benchmark

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// BenchmarkHistory stores historical benchmark results
type BenchmarkHistory struct {
	Results    []TimestampedResult `json:"results"`
	Metadata   HistoryMetadata     `json:"metadata"`
	LastUpdate time.Time           `json:"last_update"`
}

// TimestampedResult wraps a benchmark result with timestamp
type TimestampedResult struct {
	Timestamp time.Time       `json:"timestamp"`
	GitHash   string          `json:"git_hash,omitempty"`
	Version   string          `json:"version,omitempty"`
	Result    BenchmarkResult `json:"result"`
}

// HistoryMetadata contains metadata about benchmark history
type HistoryMetadata struct {
	RunnerName    string    `json:"runner_name"`
	CircuitType   string    `json:"circuit_type"`
	Scenario      string    `json:"scenario"`
	CreatedAt     time.Time `json:"created_at"`
	TotalRuns     int       `json:"total_runs"`
	RetentionDays int       `json:"retention_days"`
}

// ComparisonResult contains the result of comparing benchmark results
type ComparisonResult struct {
	Baseline     TimestampedResult `json:"baseline"`
	Current      TimestampedResult `json:"current"`
	Improvements []string          `json:"improvements"`
	Regressions  []string          `json:"regressions"`
	Summary      ComparisonSummary `json:"summary"`
}

// ComparisonSummary provides high-level comparison metrics
type ComparisonSummary struct {
	DurationChange   float64 `json:"duration_change_percent"`
	MemoryChange     float64 `json:"memory_change_percent"`
	ThroughputChange float64 `json:"throughput_change_percent"`
	OverallRating    string  `json:"overall_rating"` // "improved", "degraded", "neutral"
}

// BenchmarkPersistence manages benchmark result storage
type BenchmarkPersistence struct {
	StorageDir    string
	RetentionDays int
}

// NewBenchmarkPersistence creates a new persistence manager
func NewBenchmarkPersistence(storageDir string) *BenchmarkPersistence {
	return &BenchmarkPersistence{
		StorageDir:    storageDir,
		RetentionDays: 30, // Default retention period
	}
}

// SaveResult saves a benchmark result to persistent storage
func (bp *BenchmarkPersistence) SaveResult(result BenchmarkResult, gitHash, version string) error {
	// Ensure storage directory exists
	if err := os.MkdirAll(bp.StorageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %v", err)
	}

	// Generate filename based on benchmark configuration
	filename := bp.generateFilename(result)
	filepath := filepath.Join(bp.StorageDir, filename)

	// Load existing history or create new
	history, err := bp.loadHistory(filepath)
	if err != nil {
		// Create new history
		history = &BenchmarkHistory{
			Results: []TimestampedResult{},
			Metadata: HistoryMetadata{
				RunnerName:    result.RunnerName,
				CircuitType:   string(result.CircuitType),
				Scenario:      string(result.Scenario),
				CreatedAt:     time.Now(),
				RetentionDays: bp.RetentionDays,
			},
		}
	}

	// Add new result
	timestampedResult := TimestampedResult{
		Timestamp: time.Now(),
		GitHash:   gitHash,
		Version:   version,
		Result:    result,
	}

	history.Results = append(history.Results, timestampedResult)
	history.LastUpdate = time.Now()
	history.Metadata.TotalRuns++

	// Clean up old results based on retention policy
	history.Results = bp.cleanupOldResults(history.Results)

	// Save updated history
	return bp.saveHistory(history, filepath)
}

// LoadHistory loads benchmark history for a specific configuration
func (bp *BenchmarkPersistence) LoadHistory(runnerName string, circuitType CircuitType, scenario BenchmarkScenario) (*BenchmarkHistory, error) {
	result := BenchmarkResult{
		RunnerName:  runnerName,
		CircuitType: circuitType,
		Scenario:    scenario,
	}

	filename := bp.generateFilename(result)
	filepath := filepath.Join(bp.StorageDir, filename)

	return bp.loadHistory(filepath)
}

// CompareWithBaseline compares current result with baseline from history
func (bp *BenchmarkPersistence) CompareWithBaseline(current BenchmarkResult, gitHash, version string) (*ComparisonResult, error) {
	// Load history
	history, err := bp.LoadHistory(current.RunnerName, current.CircuitType, current.Scenario)
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %v", err)
	}

	if len(history.Results) == 0 {
		return nil, fmt.Errorf("no baseline results found")
	}

	// Find baseline (could be last stable version, or average of recent runs)
	baseline := bp.findBaseline(history.Results)

	// Create current result entry
	currentEntry := TimestampedResult{
		Timestamp: time.Now(),
		GitHash:   gitHash,
		Version:   version,
		Result:    current,
	}

	// Perform comparison
	return bp.compareResults(baseline, currentEntry), nil
}

// GetTrends analyzes trends in benchmark performance over time
func (bp *BenchmarkPersistence) GetTrends(runnerName string, circuitType CircuitType, scenario BenchmarkScenario, days int) (*TrendAnalysis, error) {
	history, err := bp.LoadHistory(runnerName, circuitType, scenario)
	if err != nil {
		return nil, err
	}

	// Filter results by time range
	cutoff := time.Now().AddDate(0, 0, -days)
	var filteredResults []TimestampedResult

	for _, result := range history.Results {
		if result.Timestamp.After(cutoff) {
			filteredResults = append(filteredResults, result)
		}
	}

	return bp.analyzeTrends(filteredResults), nil
}

// TrendAnalysis contains trend analysis results
type TrendAnalysis struct {
	Period          string            `json:"period"`
	SampleCount     int               `json:"sample_count"`
	DurationTrend   TrendMetric       `json:"duration_trend"`
	MemoryTrend     TrendMetric       `json:"memory_trend"`
	ThroughputTrend TrendMetric       `json:"throughput_trend"`
	Volatility      VolatilityMetrics `json:"volatility"`
}

// TrendMetric represents a performance trend
type TrendMetric struct {
	Direction     string  `json:"direction"`  // "improving", "degrading", "stable"
	Slope         float64 `json:"slope"`      // Rate of change
	Confidence    float64 `json:"confidence"` // 0-1, how confident we are in the trend
	StartValue    float64 `json:"start_value"`
	EndValue      float64 `json:"end_value"`
	ChangePercent float64 `json:"change_percent"`
}

// VolatilityMetrics measures consistency of performance
type VolatilityMetrics struct {
	DurationStdDev   time.Duration `json:"duration_std_dev"`
	MemoryStdDev     float64       `json:"memory_std_dev"`
	ConsistencyScore float64       `json:"consistency_score"` // 0-1, higher is more consistent
}

// generateFilename creates a consistent filename for benchmark results
func (bp *BenchmarkPersistence) generateFilename(result BenchmarkResult) string {
	return fmt.Sprintf("bench_%s_%s_%s.json",
		result.RunnerName,
		result.CircuitType,
		result.Scenario)
}

// loadHistory loads benchmark history from file
func (bp *BenchmarkPersistence) loadHistory(filepath string) (*BenchmarkHistory, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var history BenchmarkHistory
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&history); err != nil {
		return nil, err
	}

	return &history, nil
}

// saveHistory saves benchmark history to file
func (bp *BenchmarkPersistence) saveHistory(history *BenchmarkHistory, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(history)
}

// cleanupOldResults removes results older than retention period
func (bp *BenchmarkPersistence) cleanupOldResults(results []TimestampedResult) []TimestampedResult {
	cutoff := time.Now().AddDate(0, 0, -bp.RetentionDays)

	var filtered []TimestampedResult
	for _, result := range results {
		if result.Timestamp.After(cutoff) {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// findBaseline selects an appropriate baseline for comparison
func (bp *BenchmarkPersistence) findBaseline(results []TimestampedResult) TimestampedResult {
	if len(results) == 0 {
		return TimestampedResult{}
	}

	// Sort by timestamp (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	// For now, use the most recent result as baseline
	// In the future, this could be more sophisticated (e.g., last stable release)
	if len(results) >= 2 {
		return results[1] // Second most recent
	}

	return results[0]
}

// compareResults performs detailed comparison between two results
func (bp *BenchmarkPersistence) compareResults(baseline, current TimestampedResult) *ComparisonResult {
	comparison := &ComparisonResult{
		Baseline: baseline,
		Current:  current,
	}

	// Calculate changes
	durationChange := bp.calculateChange(
		float64(baseline.Result.Duration),
		float64(current.Result.Duration))

	memoryChange := bp.calculateChange(
		float64(baseline.Result.ResourceUsage.MemoryDelta),
		float64(current.Result.ResourceUsage.MemoryDelta))

	// Analyze improvements and regressions
	if durationChange < -5.0 { // 5% improvement threshold
		comparison.Improvements = append(comparison.Improvements,
			fmt.Sprintf("Duration improved by %.2f%%", -durationChange))
	} else if durationChange > 5.0 { // 5% regression threshold
		comparison.Regressions = append(comparison.Regressions,
			fmt.Sprintf("Duration regressed by %.2f%%", durationChange))
	}

	if memoryChange < -10.0 { // 10% memory improvement threshold
		comparison.Improvements = append(comparison.Improvements,
			fmt.Sprintf("Memory usage improved by %.2f%%", -memoryChange))
	} else if memoryChange > 10.0 { // 10% memory regression threshold
		comparison.Regressions = append(comparison.Regressions,
			fmt.Sprintf("Memory usage regressed by %.2f%%", memoryChange))
	}

	// Calculate overall rating
	overallRating := "neutral"
	if len(comparison.Improvements) > len(comparison.Regressions) {
		overallRating = "improved"
	} else if len(comparison.Regressions) > len(comparison.Improvements) {
		overallRating = "degraded"
	}

	comparison.Summary = ComparisonSummary{
		DurationChange: durationChange,
		MemoryChange:   memoryChange,
		OverallRating:  overallRating,
	}

	return comparison
}

// calculateChange computes percentage change between two values
func (bp *BenchmarkPersistence) calculateChange(baseline, current float64) float64 {
	if baseline == 0 {
		return 0
	}
	return ((current - baseline) / baseline) * 100.0
}

// analyzeTrends performs trend analysis on historical data
func (bp *BenchmarkPersistence) analyzeTrends(results []TimestampedResult) *TrendAnalysis {
	if len(results) < 2 {
		return &TrendAnalysis{
			SampleCount: len(results),
		}
	}

	// Sort by timestamp
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.Before(results[j].Timestamp)
	})

	// Extract time series data
	durations := make([]float64, len(results))
	memories := make([]float64, len(results))

	for i, result := range results {
		durations[i] = float64(result.Result.Duration)
		memories[i] = float64(result.Result.ResourceUsage.MemoryDelta)
	}

	// Calculate simple linear trends
	durationTrend := bp.calculateLinearTrend(durations)
	memoryTrend := bp.calculateLinearTrend(memories)

	return &TrendAnalysis{
		Period:        fmt.Sprintf("%d days", len(results)),
		SampleCount:   len(results),
		DurationTrend: durationTrend,
		MemoryTrend:   memoryTrend,
		Volatility:    bp.calculateVolatility(durations, memories),
	}
}

// calculateLinearTrend computes simple linear trend metrics
func (bp *BenchmarkPersistence) calculateLinearTrend(values []float64) TrendMetric {
	if len(values) < 2 {
		return TrendMetric{}
	}

	start := values[0]
	end := values[len(values)-1]
	change := ((end - start) / start) * 100.0

	direction := "stable"
	if change > 5.0 {
		direction = "degrading"
	} else if change < -5.0 {
		direction = "improving"
	}

	return TrendMetric{
		Direction:     direction,
		StartValue:    start,
		EndValue:      end,
		ChangePercent: change,
		Confidence:    0.8, // Simple confidence measure
	}
}

// calculateVolatility measures consistency of performance
func (bp *BenchmarkPersistence) calculateVolatility(durations, memories []float64) VolatilityMetrics {
	if len(durations) == 0 {
		return VolatilityMetrics{}
	}

	// Calculate standard deviation for durations
	durationMean := bp.mean(durations)
	durationVariance := bp.variance(durations, durationMean)
	durationStdDev := time.Duration(bp.sqrt(durationVariance))

	// Calculate standard deviation for memory
	memoryMean := bp.mean(memories)
	memoryVariance := bp.variance(memories, memoryMean)
	memoryStdDev := bp.sqrt(memoryVariance)

	// Simple consistency score (inverse of coefficient of variation)
	consistencyScore := 1.0 / (1.0 + (bp.sqrt(durationVariance) / durationMean))

	return VolatilityMetrics{
		DurationStdDev:   durationStdDev,
		MemoryStdDev:     memoryStdDev,
		ConsistencyScore: consistencyScore,
	}
}

// Helper math functions
func (bp *BenchmarkPersistence) mean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (bp *BenchmarkPersistence) variance(values []float64, mean float64) float64 {
	sum := 0.0
	for _, v := range values {
		diff := v - mean
		sum += diff * diff
	}
	return sum / float64(len(values))
}

func (bp *BenchmarkPersistence) sqrt(value float64) float64 {
	// Simple approximation for square root
	if value == 0 {
		return 0
	}

	x := value
	for i := 0; i < 10; i++ { // Newton's method iterations
		x = (x + value/x) / 2
	}
	return x
}

// ExportResults exports benchmark results to various formats
func (bp *BenchmarkPersistence) ExportResults(history *BenchmarkHistory, format string, writer io.Writer) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(history)
	case "csv":
		return bp.exportCSV(history, writer)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportCSV exports results in CSV format
func (bp *BenchmarkPersistence) exportCSV(history *BenchmarkHistory, writer io.Writer) error {
	// Write CSV header
	_, err := writer.Write([]byte("timestamp,git_hash,version,success,duration_ns,memory_delta_bytes,allocs_per_op,bytes_per_op\n"))
	if err != nil {
		return err
	}

	// Write data rows
	for _, result := range history.Results {
		line := fmt.Sprintf("%s,%s,%s,%t,%d,%d,%d,%d\n",
			result.Timestamp.Format(time.RFC3339),
			result.GitHash,
			result.Version,
			result.Result.Success,
			result.Result.Duration.Nanoseconds(),
			result.Result.ResourceUsage.MemoryDelta,
			result.Result.AllocsPerOp,
			result.Result.BytesPerOp,
		)
		_, err := writer.Write([]byte(line))
		if err != nil {
			return err
		}
	}

	return nil
}
