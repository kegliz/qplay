package benchmark

import (
	"fmt"
	"testing"

	"github.com/kegliz/qplay/qc/simulator"
	_ "github.com/kegliz/qplay/qc/simulator/itsu" // Import to register the runner
	"github.com/kegliz/qplay/qc/testutil"
)

// BenchmarkAllPlugins runs benchmarks for all registered quantum backend plugins
func BenchmarkAllPlugins(b *testing.B) {
	suite := NewPluginBenchmarkSuite()

	// For each registered runner
	for _, runnerName := range suite.runners {
		runnerName := runnerName // capture loop variable

		b.Run(runnerName, func(b *testing.B) {
			benchmarkRunner(b, runnerName, suite)
		})
	}
}

// benchmarkRunner runs all benchmark scenarios for a specific runner
func benchmarkRunner(b *testing.B, runnerName string, suite *PluginBenchmarkSuite) {
	// Test runner creation first
	runner, err := simulator.CreateRunner(runnerName)
	if err != nil {
		b.Fatalf("Failed to create runner %s: %v", runnerName, err)
	}

	// Display backend info if available
	if info := simulator.GetBackendInfo(runner); info != nil {
		b.Logf("Testing %s v%s: %s", info.Name, info.Version, info.Description)
	}

	// Run benchmarks for each circuit type
	for _, circuitType := range suite.circuits {
		circuitType := circuitType // capture loop variable

		b.Run(string(circuitType), func(b *testing.B) {
			benchmarkCircuitType(b, runnerName, circuitType, suite)
		})
	}
}

// benchmarkCircuitType runs all scenarios for a specific circuit type
func benchmarkCircuitType(b *testing.B, runnerName string, circuitType CircuitType, suite *PluginBenchmarkSuite) {
	for _, scenario := range suite.scenarios {
		scenario := scenario // capture loop variable

		b.Run(string(scenario), func(b *testing.B) {
			config := BenchmarkConfig{
				CircuitType: circuitType,
				Scenario:    scenario,
				Config:      suite.config,
				RunnerName:  runnerName,
			}

			result := RunSingleBenchmark(b, config)

			if !result.Success {
				if result.Error != "" {
					b.Logf("Benchmark failed: %s", result.Error)
				}
				// Don't fail the benchmark, just skip unsupported features
				return
			}

			// Log performance information
			if result.BackendInfo != nil {
				b.Logf("Backend: %s", result.BackendInfo.Name)
			}
			if result.Metrics != nil {
				b.Logf("Executions: %d, Success: %d, Avg Time: %v",
					result.Metrics.TotalExecutions,
					result.Metrics.SuccessfulRuns,
					result.Metrics.AverageTime)
			}
		})
	}
}

// BenchmarkSimpleCircuits focuses on simple circuits across all plugins
func BenchmarkSimpleCircuits(b *testing.B) {
	suite := NewPluginBenchmarkSuite().
		WithCircuits(SimpleCircuit).
		WithScenarios(SerialExecution)

	for _, runnerName := range suite.runners {
		runnerName := runnerName

		b.Run(runnerName, func(b *testing.B) {
			config := BenchmarkConfig{
				CircuitType: SimpleCircuit,
				Scenario:    SerialExecution,
				Config:      suite.config,
				RunnerName:  runnerName,
			}

			RunSingleBenchmark(b, config)
		})
	}
}

// BenchmarkEntanglementCircuits focuses on entanglement circuits across all plugins
func BenchmarkEntanglementCircuits(b *testing.B) {
	suite := NewPluginBenchmarkSuite().
		WithCircuits(EntanglementCircuit).
		WithScenarios(SerialExecution)

	for _, runnerName := range suite.runners {
		runnerName := runnerName

		b.Run(runnerName, func(b *testing.B) {
			config := BenchmarkConfig{
				CircuitType: EntanglementCircuit,
				Scenario:    SerialExecution,
				Config:      suite.config,
				RunnerName:  runnerName,
			}

			RunSingleBenchmark(b, config)
		})
	}
}

// BenchmarkParallelExecution tests parallel execution capabilities
func BenchmarkParallelExecution(b *testing.B) {
	suite := NewPluginBenchmarkSuite().
		WithCircuits(SimpleCircuit, EntanglementCircuit).
		WithScenarios(ParallelExecution)

	for _, runnerName := range suite.runners {
		runnerName := runnerName

		b.Run(runnerName, func(b *testing.B) {
			for _, circuitType := range suite.circuits {
				circuitType := circuitType

				b.Run(string(circuitType), func(b *testing.B) {
					config := BenchmarkConfig{
						CircuitType: circuitType,
						Scenario:    ParallelExecution,
						Config:      suite.config,
						RunnerName:  runnerName,
					}

					RunSingleBenchmark(b, config)
				})
			}
		})
	}
}

// BenchmarkBatchExecution tests batch execution capabilities
func BenchmarkBatchExecution(b *testing.B) {
	suite := NewPluginBenchmarkSuite().
		WithCircuits(SimpleCircuit).
		WithScenarios(BatchExecution)

	for _, runnerName := range suite.runners {
		runnerName := runnerName

		// Check if runner supports batch execution
		runner, err := simulator.CreateRunner(runnerName)
		if err != nil {
			continue
		}

		if !simulator.SupportsBatch(runner) {
			b.Logf("Skipping %s: no batch support", runnerName)
			continue
		}

		b.Run(runnerName, func(b *testing.B) {
			config := BenchmarkConfig{
				CircuitType: SimpleCircuit,
				Scenario:    BatchExecution,
				Config:      suite.config,
				RunnerName:  runnerName,
			}

			RunSingleBenchmark(b, config)
		})
	}
}

// BenchmarkWithMetrics tests execution with metrics collection
func BenchmarkWithMetrics(b *testing.B) {
	suite := NewPluginBenchmarkSuite().
		WithCircuits(SimpleCircuit).
		WithScenarios(MetricsCollection)

	for _, runnerName := range suite.runners {
		runnerName := runnerName

		// Check if runner supports metrics
		runner, err := simulator.CreateRunner(runnerName)
		if err != nil {
			continue
		}

		if !simulator.SupportsMetrics(runner) {
			b.Logf("Skipping %s: no metrics support", runnerName)
			continue
		}

		b.Run(runnerName, func(b *testing.B) {
			config := BenchmarkConfig{
				CircuitType: SimpleCircuit,
				Scenario:    MetricsCollection,
				Config:      suite.config,
				RunnerName:  runnerName,
			}

			result := RunSingleBenchmark(b, config)

			if result.Success && result.Metrics != nil {
				b.Logf("Collected metrics: %d executions, %d successful",
					result.Metrics.TotalExecutions,
					result.Metrics.SuccessfulRuns)
			}
		})
	}
}

// Example of a custom benchmark using the framework
func BenchmarkCustomScenario(b *testing.B) {
	runners := simulator.ListRunners()
	if len(runners) == 0 {
		b.Skip("No runners registered")
	}

	// Create a custom configuration
	customConfig := testutil.TestConfig{
		Shots:     50, // Very quick for demo
		Qubits:    2,  // Small circuit
		Workers:   2,  // Minimal parallelism
		Timeout:   testutil.DefaultTestTimeout,
		Tolerance: testutil.DefaultTolerance,
	}

	// Test only the first available runner with custom config
	runnerName := runners[0]

	b.Run(fmt.Sprintf("Custom_%s", runnerName), func(b *testing.B) {
		config := BenchmarkConfig{
			CircuitType: SimpleCircuit,
			Scenario:    SerialExecution,
			Config:      customConfig,
			RunnerName:  runnerName,
		}

		result := RunSingleBenchmark(b, config)

		if result.Success {
			b.Logf("Custom benchmark completed for %s", runnerName)
		}
	})
}
