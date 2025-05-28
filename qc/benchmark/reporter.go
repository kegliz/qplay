package benchmark

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/kegliz/qplay/qc/simulator"
)

// BenchmarkReport contains comprehensive benchmark results
type BenchmarkReport struct {
	Timestamp    time.Time         `json:"timestamp"`
	TotalRunners int               `json:"total_runners"`
	Results      []BenchmarkResult `json:"results"`
	Summary      BenchmarkSummary  `json:"summary"`
}

// BenchmarkSummary provides aggregated statistics
type BenchmarkSummary struct {
	TotalTests      int                        `json:"total_tests"`
	SuccessfulTests int                        `json:"successful_tests"`
	FailedTests     int                        `json:"failed_tests"`
	AverageDuration time.Duration              `json:"average_duration"`
	ByRunner        map[string]RunnerSummary   `json:"by_runner"`
	ByCircuit       map[string]CircuitSummary  `json:"by_circuit"`
	ByScenario      map[string]ScenarioSummary `json:"by_scenario"`
}

// RunnerSummary contains statistics for a specific runner
type RunnerSummary struct {
	Name            string                 `json:"name"`
	TotalTests      int                    `json:"total_tests"`
	SuccessfulTests int                    `json:"successful_tests"`
	AverageDuration time.Duration          `json:"average_duration"`
	BackendInfo     *simulator.BackendInfo `json:"backend_info,omitempty"`
}

// CircuitSummary contains statistics for a specific circuit type
type CircuitSummary struct {
	Type            CircuitType   `json:"type"`
	TotalTests      int           `json:"total_tests"`
	SuccessfulTests int           `json:"successful_tests"`
	AverageDuration time.Duration `json:"average_duration"`
}

// ScenarioSummary contains statistics for a specific scenario
type ScenarioSummary struct {
	Scenario        BenchmarkScenario `json:"scenario"`
	TotalTests      int               `json:"total_tests"`
	SuccessfulTests int               `json:"successful_tests"`
	AverageDuration time.Duration     `json:"average_duration"`
}

// BenchmarkReporter handles collection and reporting of benchmark results
type BenchmarkReporter struct {
	results []BenchmarkResult
}

// NewBenchmarkReporter creates a new benchmark reporter
func NewBenchmarkReporter() *BenchmarkReporter {
	return &BenchmarkReporter{
		results: make([]BenchmarkResult, 0),
	}
}

// AddResult adds a benchmark result to the reporter
func (r *BenchmarkReporter) AddResult(result BenchmarkResult) {
	r.results = append(r.results, result)
}

// GenerateReport creates a comprehensive benchmark report
func (r *BenchmarkReporter) GenerateReport() BenchmarkReport {
	report := BenchmarkReport{
		Timestamp: time.Now(),
		Results:   r.results,
		Summary:   r.generateSummary(),
	}

	// Count unique runners
	runners := make(map[string]bool)
	for _, result := range r.results {
		runners[result.RunnerName] = true
	}
	report.TotalRunners = len(runners)

	return report
}

// generateSummary creates aggregated statistics
func (r *BenchmarkReporter) generateSummary() BenchmarkSummary {
	summary := BenchmarkSummary{
		ByRunner:   make(map[string]RunnerSummary),
		ByCircuit:  make(map[string]CircuitSummary),
		ByScenario: make(map[string]ScenarioSummary),
	}

	var totalDuration time.Duration

	// Initialize counters
	runnerStats := make(map[string]*RunnerSummary)
	circuitStats := make(map[string]*CircuitSummary)
	scenarioStats := make(map[string]*ScenarioSummary)

	for _, result := range r.results {
		summary.TotalTests++
		totalDuration += result.Duration

		if result.Success {
			summary.SuccessfulTests++
		} else {
			summary.FailedTests++
		}

		// Runner statistics
		if _, exists := runnerStats[result.RunnerName]; !exists {
			runnerStats[result.RunnerName] = &RunnerSummary{
				Name:        result.RunnerName,
				BackendInfo: result.BackendInfo,
			}
		}
		runnerStat := runnerStats[result.RunnerName]
		runnerStat.TotalTests++
		if result.Success {
			runnerStat.SuccessfulTests++
		}

		// Circuit statistics
		circuitKey := string(result.CircuitType)
		if _, exists := circuitStats[circuitKey]; !exists {
			circuitStats[circuitKey] = &CircuitSummary{
				Type: result.CircuitType,
			}
		}
		circuitStat := circuitStats[circuitKey]
		circuitStat.TotalTests++
		if result.Success {
			circuitStat.SuccessfulTests++
		}

		// Scenario statistics
		scenarioKey := string(result.Scenario)
		if _, exists := scenarioStats[scenarioKey]; !exists {
			scenarioStats[scenarioKey] = &ScenarioSummary{
				Scenario: result.Scenario,
			}
		}
		scenarioStat := scenarioStats[scenarioKey]
		scenarioStat.TotalTests++
		if result.Success {
			scenarioStat.SuccessfulTests++
		}
	}

	// Calculate averages
	if summary.TotalTests > 0 {
		summary.AverageDuration = totalDuration / time.Duration(summary.TotalTests)
	}

	// Copy stats to summary
	for name, stat := range runnerStats {
		if stat.TotalTests > 0 {
			stat.AverageDuration = totalDuration / time.Duration(stat.TotalTests)
		}
		summary.ByRunner[name] = *stat
	}

	for name, stat := range circuitStats {
		if stat.TotalTests > 0 {
			stat.AverageDuration = totalDuration / time.Duration(stat.TotalTests)
		}
		summary.ByCircuit[name] = *stat
	}

	for name, stat := range scenarioStats {
		if stat.TotalTests > 0 {
			stat.AverageDuration = totalDuration / time.Duration(stat.TotalTests)
		}
		summary.ByScenario[name] = *stat
	}

	return summary
}

// WriteJSON writes the report as JSON to the provided writer
func (r *BenchmarkReporter) WriteJSON(w io.Writer) error {
	report := r.GenerateReport()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// PrintSummary prints a human-readable summary to the provided writer
func (r *BenchmarkReporter) PrintSummary(w io.Writer) {
	report := r.GenerateReport()

	fmt.Fprintf(w, "🚀 Quantum Backend Benchmark Report\n")
	fmt.Fprintf(w, "=====================================\n")
	fmt.Fprintf(w, "Generated: %s\n", report.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(w, "Total Runners: %d\n", report.TotalRunners)
	fmt.Fprintf(w, "Total Tests: %d\n", report.Summary.TotalTests)
	fmt.Fprintf(w, "Successful: %d\n", report.Summary.SuccessfulTests)
	fmt.Fprintf(w, "Failed: %d\n", report.Summary.FailedTests)
	fmt.Fprintf(w, "Average Duration: %v\n\n", report.Summary.AverageDuration)

	// Runner summary
	fmt.Fprintf(w, "📊 Results by Runner:\n")
	fmt.Fprintf(w, "--------------------\n")

	// Sort runners by name for consistent output
	var runnerNames []string
	for name := range report.Summary.ByRunner {
		runnerNames = append(runnerNames, name)
	}
	sort.Strings(runnerNames)

	for _, name := range runnerNames {
		stat := report.Summary.ByRunner[name]
		fmt.Fprintf(w, "• %s: %d/%d tests passed (%.1f%%), avg: %v\n",
			stat.Name,
			stat.SuccessfulTests,
			stat.TotalTests,
			float64(stat.SuccessfulTests)/float64(stat.TotalTests)*100,
			stat.AverageDuration)

		if stat.BackendInfo != nil {
			fmt.Fprintf(w, "  └─ %s v%s\n", stat.BackendInfo.Name, stat.BackendInfo.Version)
		}
	}

	// Circuit summary
	fmt.Fprintf(w, "\n🔄 Results by Circuit Type:\n")
	fmt.Fprintf(w, "----------------------------\n")
	for circuitType, stat := range report.Summary.ByCircuit {
		fmt.Fprintf(w, "• %s: %d/%d tests passed (%.1f%%), avg: %v\n",
			circuitType,
			stat.SuccessfulTests,
			stat.TotalTests,
			float64(stat.SuccessfulTests)/float64(stat.TotalTests)*100,
			stat.AverageDuration)
	}

	// Scenario summary
	fmt.Fprintf(w, "\n⚡ Results by Scenario:\n")
	fmt.Fprintf(w, "----------------------\n")
	for scenario, stat := range report.Summary.ByScenario {
		fmt.Fprintf(w, "• %s: %d/%d tests passed (%.1f%%), avg: %v\n",
			scenario,
			stat.SuccessfulTests,
			stat.TotalTests,
			float64(stat.SuccessfulTests)/float64(stat.TotalTests)*100,
			stat.AverageDuration)
	}

	// Failed tests details
	if report.Summary.FailedTests > 0 {
		fmt.Fprintf(w, "\n❌ Failed Tests:\n")
		fmt.Fprintf(w, "----------------\n")
		for _, result := range report.Results {
			if !result.Success {
				fmt.Fprintf(w, "• %s/%s/%s: %s\n",
					result.RunnerName,
					result.CircuitType,
					result.Scenario,
					result.Error)

				// Show resource limit violations if any
				if len(result.LimitsExceeded) > 0 {
					fmt.Fprintf(w, "  └─ Limits exceeded: %v\n", result.LimitsExceeded)
				}
			}
		}
	}

	// Resource usage summary
	fmt.Fprintf(w, "\n💾 Resource Usage Summary:\n")
	fmt.Fprintf(w, "--------------------------\n")

	var totalMemoryDelta int64
	var maxMemoryUsage uint64
	var averageCircuitDepth float64
	var totalTests int

	for _, result := range report.Results {
		if result.Success {
			totalMemoryDelta += result.ResourceUsage.MemoryDelta
			if result.ResourceUsage.PeakMemory > maxMemoryUsage {
				maxMemoryUsage = result.ResourceUsage.PeakMemory
			}
			averageCircuitDepth += float64(result.ResourceUsage.CircuitDepth)
			totalTests++
		}
	}

	if totalTests > 0 {
		avgMemoryDelta := totalMemoryDelta / int64(totalTests)
		averageCircuitDepth /= float64(totalTests)

		fmt.Fprintf(w, "• Average memory delta: %d bytes\n", avgMemoryDelta)
		fmt.Fprintf(w, "• Peak memory usage: %d MB\n", maxMemoryUsage/(1024*1024))
		fmt.Fprintf(w, "• Average circuit depth: %.1f\n", averageCircuitDepth)
	}
}
