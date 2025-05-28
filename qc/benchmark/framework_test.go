package benchmark

import (
	"testing"

	"github.com/kegliz/qplay/qc/simulator"
	_ "github.com/kegliz/qplay/qc/simulator/itsu" // Import to register the runner
	"github.com/kegliz/qplay/qc/testutil"
)

// TestFrameworkBasics tests the basic functionality of the benchmark framework
func TestFrameworkBasics(t *testing.T) {
	// Test circuit creation
	t.Run("CircuitCreation", func(t *testing.T) {
		for circuitType, builder := range StandardCircuits {
			t.Run(string(circuitType), func(t *testing.T) {
				build := builder(2)
				_, err := build.BuildCircuit()
				if err != nil {
					t.Errorf("Failed to build %s circuit: %v", circuitType, err)
				}
			})
		}
	})

	// Test runner availability
	t.Run("RunnersAvailable", func(t *testing.T) {
		runners := simulator.ListRunners()
		if len(runners) == 0 {
			t.Skip("No runners registered")
		}

		t.Logf("Available runners: %v", runners)

		// Test creating each runner
		for _, runnerName := range runners {
			t.Run(runnerName, func(t *testing.T) {
				runner, err := simulator.CreateRunner(runnerName)
				if err != nil {
					t.Errorf("Failed to create runner %s: %v", runnerName, err)
					return
				}

				// Test basic execution
				build := buildSimpleCircuit(1)
				circ, err := build.BuildCircuit()
				if err != nil {
					t.Errorf("Failed to build circuit: %v", err)
					return
				}

				result, err := runner.RunOnce(circ)
				if err != nil {
					t.Errorf("Failed to run circuit: %v", err)
					return
				}

				t.Logf("Runner %s result: %s", runnerName, result)
			})
		}
	})

	// Test benchmark suite creation
	t.Run("SuiteCreation", func(t *testing.T) {
		suite := NewPluginBenchmarkSuite()
		if suite == nil {
			t.Error("Failed to create benchmark suite")
		}

		if len(suite.runners) == 0 {
			t.Skip("No runners available for testing")
		}

		t.Logf("Suite has %d runners, %d circuits, %d scenarios",
			len(suite.runners), len(suite.circuits), len(suite.scenarios))
	})

	// Test single benchmark execution
	t.Run("SingleBenchmark", func(t *testing.T) {
		runners := simulator.ListRunners()
		if len(runners) == 0 {
			t.Skip("No runners available")
		}

		config := BenchmarkConfig{
			CircuitType: SimpleCircuit,
			Scenario:    SerialExecution,
			Config:      testutil.QuickTestConfig,
			RunnerName:  runners[0],
			Limits:      DefaultResourceLimits, // Add resource limits
		}

		// Create a fake benchmark for testing
		b := &testing.B{}
		result := RunSingleBenchmark(b, config)

		if !result.Success {
			t.Errorf("Benchmark failed: %s", result.Error)
		} else {
			t.Logf("Benchmark succeeded in %v", result.Duration)
		}
	})
}
