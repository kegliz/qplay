// Command benchmark-demo demonstrates the plugin-level benchmark framework
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kegliz/qplay/qc/benchmark"
	"github.com/kegliz/qplay/qc/simulator"
	_ "github.com/kegliz/qplay/qc/simulator/itsu" // Import to register the runner
	"github.com/kegliz/qplay/qc/testutil"
)

func main() {
	var (
		command    = flag.String("cmd", "run", "Command to execute: list, info, run, benchmark, benchmark-all, stress, ci")
		runner     = flag.String("runner", "itsu", "Runner name to use")
		circuit    = flag.String("circuit", "simple", "Circuit type: simple, entanglement, superposition, mixed")
		scenario   = flag.String("scenario", "serial", "Scenario: serial, parallel, batch, context, metrics")
		output     = flag.String("output", "console", "Output format: console, json")
		shots      = flag.Int("shots", 100, "Number of shots for benchmark")
		qubits     = flag.Int("qubits", 2, "Number of qubits")
		workers    = flag.Int("workers", 4, "Number of worker threads")
		duration   = flag.Duration("duration", 10*time.Second, "Duration for stress test")
		concurrent = flag.Int("concurrent", 5, "Number of concurrent operations for stress test")
		outputDir  = flag.String("output-dir", "./benchmark-results", "Output directory for reports")
	)
	flag.Parse()

	switch *command {
	case "list":
		listRunners()
	case "info":
		showRunnerInfo(*runner)
	case "run":
		runExample(*runner, *circuit)
	case "benchmark":
		runBenchmark(*runner, *circuit, *scenario, *output, *shots, *qubits, *workers)
	case "benchmark-all":
		runAllBenchmarks(*output)
	case "stress":
		runStressTest(*runner, *duration, *concurrent, *output)
	case "ci":
		runCIBenchmarks(*outputDir, *output)
	default:
		fmt.Printf("Unknown command: %s\n", *command)
		flag.Usage()
		os.Exit(1)
	}
}

func listRunners() {
	fmt.Println("🔌 Available Quantum Backend Runners:")
	fmt.Println("====================================")

	runners := simulator.ListRunners()
	if len(runners) == 0 {
		fmt.Println("No runners registered")
		return
	}

	for i, name := range runners {
		fmt.Printf("%d. %s\n", i+1, name)

		// Get additional info if available
		if runner, err := simulator.CreateRunner(name); err == nil {
			if info := simulator.GetBackendInfo(runner); info != nil {
				fmt.Printf("   └─ %s v%s\n", info.Name, info.Version)
				fmt.Printf("   └─ %s\n", info.Description)
			}
		}
	}
}

func showRunnerInfo(runnerName string) {
	fmt.Printf("🔍 Runner Information: %s\n", runnerName)
	fmt.Println("========================")

	runner, err := simulator.CreateRunner(runnerName)
	if err != nil {
		fmt.Printf("❌ Failed to create runner: %v\n", err)
		return
	}

	// Basic info
	info := simulator.GetBackendInfo(runner)
	if info != nil {
		fmt.Printf("Name: %s\n", info.Name)
		fmt.Printf("Version: %s\n", info.Version)
		fmt.Printf("Description: %s\n", info.Description)
		fmt.Printf("Vendor: %s\n", info.Vendor)

		fmt.Println("\nCapabilities:")
		for capability, supported := range info.Capabilities {
			status := "❌"
			if supported {
				status = "✅"
			}
			fmt.Printf("  %s %s\n", status, capability)
		}

		if len(info.Metadata) > 0 {
			fmt.Println("\nMetadata:")
			for key, value := range info.Metadata {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	// Check enhanced interface support
	fmt.Println("\nEnhanced Interface Support:")
	fmt.Printf("  ✅ Basic Execution\n")
	fmt.Printf("  %s Context Support\n", checkmark(simulator.SupportsContext(runner)))
	fmt.Printf("  %s Configuration\n", checkmark(simulator.SupportsConfiguration(runner)))
	fmt.Printf("  %s Metrics Collection\n", checkmark(simulator.SupportsMetrics(runner)))
	fmt.Printf("  %s Batch Execution\n", checkmark(simulator.SupportsBatch(runner)))
	fmt.Printf("  %s Circuit Validation\n", checkmark(simulator.SupportsValidation(runner)))
	fmt.Printf("  %s Backend Info\n", checkmark(simulator.SupportsBackendInfo(runner)))
}

func checkmark(supported bool) string {
	if supported {
		return "✅"
	}
	return "❌"
}

func runExample(runnerName, circuitType string) {
	fmt.Printf("🚀 Running Example: %s with %s circuit\n", runnerName, circuitType)
	fmt.Println("=============================================")

	// Parse circuit type
	ct := parseCircuitType(circuitType)
	if ct == "" {
		fmt.Printf("❌ Unknown circuit type: %s\n", circuitType)
		return
	}

	// Create runner
	runner, err := simulator.CreateRunner(runnerName)
	if err != nil {
		fmt.Printf("❌ Failed to create runner: %v\n", err)
		return
	}

	// Build circuit
	circuitBuilder := benchmark.StandardCircuits[ct]
	build := circuitBuilder(2) // Use 2 qubits for demo
	circ, err := build.BuildCircuit()
	if err != nil {
		fmt.Printf("❌ Failed to build circuit: %v\n", err)
		return
	}

	fmt.Printf("📋 Circuit: %s\n", benchmark.GetCircuitDescription(ct))

	// Run the circuit
	start := time.Now()
	result, err := runner.RunOnce(circ)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Execution failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Result: %s\n", result)
	fmt.Printf("⏱️  Duration: %v\n", duration)

	// Show metrics if available
	if metrics, ok := runner.(simulator.MetricsCollector); ok {
		execMetrics := metrics.GetMetrics()
		fmt.Printf("📊 Metrics: %d executions, %v avg time\n",
			execMetrics.TotalExecutions,
			execMetrics.AverageTime)
	}
}

func runBenchmark(runnerName, circuitType, scenario, output string, shots, qubits, workers int) {
	fmt.Printf("🏁 Running Benchmark: %s\n", runnerName)
	fmt.Println("======================")

	// Parse parameters
	ct := parseCircuitType(circuitType)
	if ct == "" {
		fmt.Printf("❌ Unknown circuit type: %s\n", circuitType)
		return
	}

	sc := parseScenario(scenario)
	if sc == "" {
		fmt.Printf("❌ Unknown scenario: %s\n", scenario)
		return
	}

	// Create configuration
	config := benchmark.BenchmarkConfig{
		CircuitType: ct,
		Scenario:    sc,
		RunnerName:  runnerName,
		Config: testutil.TestConfig{
			Shots:     shots,
			Qubits:    qubits,
			Workers:   workers,
			Timeout:   testutil.DefaultTestTimeout,
			Tolerance: testutil.DefaultTolerance,
		},
		Limits: benchmark.ResourceLimits{
			MaxMemoryMB:     300, // 300MB limit for CLI demo
			MaxDuration:     20 * time.Second,
			MaxCircuitDepth: 15,
			MaxQubits:       min(qubits, 4), // Cap at 4 qubits for demo
		},
	}

	// Create a dummy benchmark
	b := &testing.B{}
	result := benchmark.RunSingleBenchmark(b, config)

	// Output results
	if output == "json" {
		reporter := benchmark.NewBenchmarkReporter()
		reporter.AddResult(result)
		reporter.WriteJSON(os.Stdout)
	} else {
		fmt.Printf("📋 Circuit: %s\n", benchmark.GetCircuitDescription(ct))
		fmt.Printf("🎯 Scenario: %s\n", scenario)
		fmt.Printf("⚙️  Config: %d shots, %d qubits, %d workers\n", shots, qubits, workers)
		fmt.Println()

		if result.Success {
			fmt.Printf("✅ Status: Success\n")
			fmt.Printf("⏱️  Duration: %v\n", result.Duration)
			if result.AllocsPerOp > 0 {
				fmt.Printf("🧠 Memory: %d allocs/op, %d bytes/op\n", result.AllocsPerOp, result.BytesPerOp)
			}
			if result.Metrics != nil {
				fmt.Printf("📊 Metrics: %d executions, %d successful\n",
					result.Metrics.TotalExecutions,
					result.Metrics.SuccessfulRuns)
			}
		} else {
			fmt.Printf("❌ Status: Failed\n")
			fmt.Printf("💥 Error: %s\n", result.Error)
		}
	}
}

func runAllBenchmarks(output string) {
	fmt.Println("🚀 Running Comprehensive Benchmark Suite")
	fmt.Println("========================================")

	reporter := benchmark.NewBenchmarkReporter()

	// Get available runners
	runners := simulator.ListRunners()
	if len(runners) == 0 {
		fmt.Println("❌ No runners available")
		return
	}

	// Run benchmarks for all combinations
	for _, runnerName := range runners {
		for _, circuitType := range []benchmark.CircuitType{benchmark.SimpleCircuit, benchmark.EntanglementCircuit} {
			for _, scenario := range []benchmark.BenchmarkScenario{benchmark.SerialExecution} {
				config := benchmark.BenchmarkConfig{
					CircuitType: circuitType,
					Scenario:    scenario,
					RunnerName:  runnerName,
					Config:      testutil.QuickTestConfig,
					Limits: benchmark.ResourceLimits{
						MaxMemoryMB:     200, // Conservative limits for all benchmarks
						MaxDuration:     15 * time.Second,
						MaxCircuitDepth: 10,
						MaxQubits:       3,
					},
				}

				fmt.Printf("Running %s/%s/%s...\n", runnerName, circuitType, scenario)

				b := &testing.B{}
				result := benchmark.RunSingleBenchmark(b, config)
				reporter.AddResult(result)
			}
		}
	}

	// Output results
	if output == "json" {
		reporter.WriteJSON(os.Stdout)
	} else {
		reporter.PrintSummary(os.Stdout)
	}
}

func runStressTest(runnerName string, duration time.Duration, concurrent int, output string) {
	fmt.Printf("🔥 Running Stress Test: %s\n", runnerName)
	fmt.Printf("Duration: %v, Concurrent Ops: %d\n", duration, concurrent)
	fmt.Println("==========================================")

	config := benchmark.StressTestConfig{
		Duration:        duration,
		ConcurrentOps:   concurrent,
		MemoryPressure:  true,
		CircuitSizes:    []int{2, 3, 4},
		MaxMemoryMB:     1024,
		RecoveryEnabled: true,
	}

	result := benchmark.RunStressTest(runnerName, config)

	if output == "json" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		encoder.Encode(result)
		return
	}

	// Console output
	fmt.Printf("\n📊 Stress Test Results:\n")
	fmt.Printf("Status: %s\n", getStatusEmoji(result.Success))
	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
	}

	fmt.Printf("Actual Duration: %v\n", result.Duration)
	fmt.Printf("Total Operations: %d\n", result.TotalOperations)
	fmt.Printf("Successful: %d (%.1f%%)\n", result.SuccessfulOps,
		float64(result.SuccessfulOps)/float64(result.TotalOperations)*100)
	fmt.Printf("Failed: %d\n", result.FailedOps)
	fmt.Printf("Panic Recoveries: %d\n", result.PanicRecoveries)
	fmt.Printf("Peak Memory: %d MB\n", result.PeakMemoryMB)

	if len(result.MemoryLeaks) > 0 {
		fmt.Printf("⚠️  Memory Leaks Detected: %d\n", len(result.MemoryLeaks))
		for i, leak := range result.MemoryLeaks {
			fmt.Printf("  %d. %v: %d MB leaked\n", i+1, leak.Timestamp.Format("15:04:05"), leak.LeakSizeMB)
		}
	}

	// Performance stats
	stats := result.PerformanceStats
	fmt.Printf("\n📈 Performance Stats:\n")
	fmt.Printf("Throughput: %.2f ops/sec\n", stats.ThroughputOpsPerSec)
	fmt.Printf("Average Duration: %v\n", stats.AvgDuration)
	fmt.Printf("P95 Duration: %v\n", stats.Percentile95)
	fmt.Printf("P99 Duration: %v\n", stats.Percentile99)
}

func runCIBenchmarks(outputDir, output string) {
	fmt.Println("🏗️  Running CI/CD Benchmark Suite")
	fmt.Println("================================")

	// Create CI runner
	ciRunner := benchmark.NewCIBenchmarkRunner(outputDir)

	// Display detected environment
	fmt.Printf("Environment: %s\n", ciRunner.Config.Environment)
	fmt.Printf("Branch: %s\n", ciRunner.Config.Branch)
	fmt.Printf("Commit: %s\n", ciRunner.Config.CommitHash)
	fmt.Printf("Build: %s\n", ciRunner.Config.BuildNumber)

	// Run the benchmark suite
	start := time.Now()
	report, err := ciRunner.RunBenchmarkSuite()
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ CI benchmarks failed: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		encoder.Encode(report)
		return
	}

	// Console output
	fmt.Printf("\n📊 CI Benchmark Results (completed in %v):\n", elapsed)

	// Calculate summary
	var totalTests, passedTests, failedTests int
	var totalDuration time.Duration

	for _, result := range report.Results {
		totalTests++
		totalDuration += result.BenchmarkResult.Duration
		if result.BenchmarkResult.Success {
			passedTests++
		} else {
			failedTests++
		}
	}

	fmt.Printf("Total Tests: %d\n", totalTests)
	fmt.Printf("Passed: %d (%.1f%%)\n", passedTests, float64(passedTests)/float64(totalTests)*100)
	fmt.Printf("Failed: %d\n", failedTests)
	fmt.Printf("Total Duration: %v\n", totalDuration)

	if totalTests > 0 {
		fmt.Printf("Average Duration: %v\n", totalDuration/time.Duration(totalTests))
	}

	// Regression analysis
	analysis := report.RegressionAnalysis
	fmt.Printf("\n🔍 Regression Analysis:\n")
	fmt.Printf("Status: %s %s\n", getRegressionStatusEmoji(analysis.OverallStatus), analysis.OverallStatus)
	fmt.Printf("Comparisons: %d\n", analysis.TotalComparisons)

	if len(analysis.Regressions) > 0 {
		fmt.Printf("⚠️  Regressions: %d\n", len(analysis.Regressions))
		for _, reg := range analysis.Regressions {
			fmt.Printf("  - %s: %s %.1f%% (%s)\n",
				reg.TestName, reg.ChangeType, reg.ChangePercent, reg.Significance)
		}
	}

	if len(analysis.Improvements) > 0 {
		fmt.Printf("✨ Improvements: %d\n", len(analysis.Improvements))
		for _, imp := range analysis.Improvements {
			fmt.Printf("  - %s: %s %.1f%%\n",
				imp.TestName, imp.ChangeType, imp.ChangePercent)
		}
	}

	fmt.Printf("\n📁 Artifacts generated in: %s\n", outputDir)
	fmt.Printf("  - benchmark-report.json\n")
	fmt.Printf("  - benchmark-summary.txt\n")
	fmt.Printf("  - benchmark-badge.json\n")

	// Exit with appropriate code for CI
	if analysis.OverallStatus == "fail" {
		fmt.Printf("\n❌ Benchmark suite failed due to regressions\n")
		os.Exit(1)
	} else if analysis.OverallStatus == "warning" {
		fmt.Printf("\n⚠️  Benchmark suite completed with warnings\n")
		os.Exit(0) // Don't fail CI for warnings, but log them
	} else {
		fmt.Printf("\n✅ Benchmark suite passed\n")
	}
}

// Helper functions for pretty output
func getStatusEmoji(success bool) string {
	if success {
		return "✅ PASSED"
	}
	return "❌ FAILED"
}

func getRegressionStatusEmoji(status string) string {
	switch status {
	case "pass":
		return "✅"
	case "warning":
		return "⚠️"
	case "fail":
		return "❌"
	default:
		return "❓"
	}
}

func parseCircuitType(circuitType string) benchmark.CircuitType {
	switch strings.ToLower(circuitType) {
	case "simple":
		return benchmark.SimpleCircuit
	case "entanglement":
		return benchmark.EntanglementCircuit
	case "superposition":
		return benchmark.SuperpositionCircuit
	case "mixed":
		return benchmark.MixedGatesCircuit
	default:
		return ""
	}
}

func parseScenario(scenario string) benchmark.BenchmarkScenario {
	switch strings.ToLower(scenario) {
	case "serial":
		return benchmark.SerialExecution
	case "parallel":
		return benchmark.ParallelExecution
	case "batch":
		return benchmark.BatchExecution
	case "context":
		return benchmark.ContextExecution
	case "metrics":
		return benchmark.MetricsCollection
	default:
		return ""
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
