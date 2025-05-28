// Package benchmark provides stress testing capabilities for quantum backend performance
package benchmark

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/kegliz/qplay/qc/simulator"
)

// StressTestConfig defines parameters for stress testing
type StressTestConfig struct {
	Duration        time.Duration // How long to run the stress test
	ConcurrentOps   int           // Number of concurrent operations
	MemoryPressure  bool          // Whether to apply memory pressure
	CircuitSizes    []int         // Different circuit sizes to test
	MaxMemoryMB     int64         // Memory limit during stress test
	RecoveryEnabled bool          // Whether to enable panic recovery
}

// DefaultStressConfig provides reasonable defaults for stress testing
var DefaultStressConfig = StressTestConfig{
	Duration:        30 * time.Second,
	ConcurrentOps:   10,
	MemoryPressure:  true,
	CircuitSizes:    []int{2, 3, 4, 5},
	MaxMemoryMB:     1024, // 1GB for stress testing
	RecoveryEnabled: true,
}

// StressTestResult contains results from stress testing
type StressTestResult struct {
	Config           StressTestConfig `json:"config"`
	RunnerName       string           `json:"runner_name"`
	Success          bool             `json:"success"`
	Error            string           `json:"error,omitempty"`
	Duration         time.Duration    `json:"duration"`
	TotalOperations  int              `json:"total_operations"`
	SuccessfulOps    int              `json:"successful_ops"`
	FailedOps        int              `json:"failed_ops"`
	PanicRecoveries  int              `json:"panic_recoveries"`
	PeakMemoryMB     int64            `json:"peak_memory_mb"`
	MemoryLeaks      []MemoryLeak     `json:"memory_leaks,omitempty"`
	PerformanceStats PerformanceStats `json:"performance_stats"`
}

// MemoryLeak represents a detected memory leak
type MemoryLeak struct {
	Timestamp      time.Time `json:"timestamp"`
	MemoryMB       int64     `json:"memory_mb"`
	ExpectedMB     int64     `json:"expected_mb"`
	LeakSizeMB     int64     `json:"leak_size_mb"`
	GoroutineCount int       `json:"goroutine_count"`
}

// PerformanceStats tracks performance during stress testing
type PerformanceStats struct {
	AvgDuration         time.Duration `json:"avg_duration"`
	MinDuration         time.Duration `json:"min_duration"`
	MaxDuration         time.Duration `json:"max_duration"`
	Percentile95        time.Duration `json:"p95_duration"`
	Percentile99        time.Duration `json:"p99_duration"`
	ThroughputOpsPerSec float64       `json:"throughput_ops_per_sec"`
}

// RunStressTest executes a comprehensive stress test on a quantum backend
func RunStressTest(runnerName string, config StressTestConfig) StressTestResult {
	result := StressTestResult{
		Config:     config,
		RunnerName: runnerName,
	}

	// Create runner
	runner, err := simulator.CreateRunner(runnerName)
	if err != nil {
		result.Error = fmt.Sprintf("failed to create runner: %v", err)
		return result
	}

	// Track memory baseline
	var baselineMemory int64
	if config.MemoryPressure {
		runtime.GC()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		baselineMemory = int64(m.Alloc) / (1024 * 1024)
	}

	// Setup performance tracking
	durations := make([]time.Duration, 0, 1000)
	var mutex sync.Mutex

	// Create cancellation context
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()

	// Track operations
	var totalOps, successOps, failedOps, panicRecoveries int32
	var wg sync.WaitGroup

	// Start concurrent workers
	for i := range config.ConcurrentOps {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Execute stress operation with recovery
					if config.RecoveryEnabled {
						func() {
							defer func() {
								if r := recover(); r != nil {
									panicRecoveries++
									failedOps++
								}
							}()

							if executeStressOperation(runner, config, &durations, &mutex) {
								successOps++
							} else {
								failedOps++
							}
							totalOps++
						}()
					} else {
						if executeStressOperation(runner, config, &durations, &mutex) {
							successOps++
						} else {
							failedOps++
						}
						totalOps++
					}

					// Small delay to prevent excessive CPU usage
					time.Sleep(time.Microsecond)
				}
			}
		}(i)
	}

	// Monitor memory usage during test
	var peakMemoryMB int64
	var memoryLeaks []MemoryLeak

	if config.MemoryPressure {
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					var m runtime.MemStats
					runtime.ReadMemStats(&m)
					currentMemoryMB := int64(m.Alloc) / (1024 * 1024)

					if currentMemoryMB > peakMemoryMB {
						peakMemoryMB = currentMemoryMB
					}

					// Check for memory leaks (significant increase from baseline)
					if currentMemoryMB > baselineMemory*2 && len(memoryLeaks) < 10 {
						leak := MemoryLeak{
							Timestamp:      time.Now(),
							MemoryMB:       currentMemoryMB,
							ExpectedMB:     baselineMemory,
							LeakSizeMB:     currentMemoryMB - baselineMemory,
							GoroutineCount: runtime.NumGoroutine(),
						}
						memoryLeaks = append(memoryLeaks, leak)
					}

					// Emergency brake if memory usage is too high
					if currentMemoryMB > config.MaxMemoryMB {
						cancel()
						result.Error = fmt.Sprintf("emergency stop: memory usage %dMB exceeded limit %dMB",
							currentMemoryMB, config.MaxMemoryMB)
						return
					}
				}
			}
		}()
	}

	// Wait for test completion
	start := time.Now()
	wg.Wait()
	result.Duration = time.Since(start)

	// Calculate results
	result.TotalOperations = int(totalOps)
	result.SuccessfulOps = int(successOps)
	result.FailedOps = int(failedOps)
	result.PanicRecoveries = int(panicRecoveries)
	result.PeakMemoryMB = peakMemoryMB
	result.MemoryLeaks = memoryLeaks
	result.Success = failedOps == 0 && len(memoryLeaks) == 0

	// Calculate performance statistics
	mutex.Lock()
	result.PerformanceStats = calculatePerformanceStats(durations, result.Duration)
	mutex.Unlock()

	return result
}

// executeStressOperation performs a single stress test operation
func executeStressOperation(runner simulator.OneShotRunner, config StressTestConfig,
	durations *[]time.Duration, mutex *sync.Mutex) bool {

	// Select random circuit size (fix the modulo operation)
	circuitSize := config.CircuitSizes[len(*durations)%len(config.CircuitSizes)]

	// Build a simple circuit for stress testing
	build := StandardCircuits[SimpleCircuit](circuitSize)
	circ, err := build.BuildCircuit()
	if err != nil {
		return false
	}

	// Execute with timing
	start := time.Now()
	_, err = runner.RunOnce(circ)
	duration := time.Since(start)

	// Record timing
	mutex.Lock()
	*durations = append(*durations, duration)
	mutex.Unlock()

	return err == nil
}

// calculatePerformanceStats computes performance statistics from duration data
func calculatePerformanceStats(durations []time.Duration, totalTime time.Duration) PerformanceStats {
	if len(durations) == 0 {
		return PerformanceStats{}
	}

	// Sort durations for percentile calculation
	sortedDurations := make([]time.Duration, len(durations))
	copy(sortedDurations, durations)

	// Simple bubble sort for percentiles (good enough for stress test data)
	for i := 0; i < len(sortedDurations); i++ {
		for j := i + 1; j < len(sortedDurations); j++ {
			if sortedDurations[i] > sortedDurations[j] {
				sortedDurations[i], sortedDurations[j] = sortedDurations[j], sortedDurations[i]
			}
		}
	}

	// Calculate statistics
	var total time.Duration
	min := sortedDurations[0]
	max := sortedDurations[len(sortedDurations)-1]

	for _, d := range durations {
		total += d
	}

	avg := total / time.Duration(len(durations))
	p95Index := int(float64(len(sortedDurations)) * 0.95)
	p99Index := int(float64(len(sortedDurations)) * 0.99)

	throughput := float64(len(durations)) / totalTime.Seconds()

	return PerformanceStats{
		AvgDuration:         avg,
		MinDuration:         min,
		MaxDuration:         max,
		Percentile95:        sortedDurations[p95Index],
		Percentile99:        sortedDurations[p99Index],
		ThroughputOpsPerSec: throughput,
	}
}

// StressBenchmark provides a benchmark-compatible stress test
func StressBenchmark(b *testing.B, runnerName string, config StressTestConfig) {
	b.Helper()

	// Adjust config for benchmark duration
	config.Duration = time.Duration(b.N) * time.Millisecond

	result := RunStressTest(runnerName, config)

	if !result.Success {
		b.Errorf("Stress test failed: %s", result.Error)
	}

	// Report metrics
	b.ReportMetric(float64(result.SuccessfulOps), "ops")
	b.ReportMetric(result.PerformanceStats.ThroughputOpsPerSec, "ops/sec")
	b.ReportMetric(float64(result.PeakMemoryMB), "MB")

	if len(result.MemoryLeaks) > 0 {
		b.Errorf("Memory leaks detected: %d", len(result.MemoryLeaks))
	}
}
