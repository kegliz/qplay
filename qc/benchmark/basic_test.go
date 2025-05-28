package benchmark

import (
	"testing"

	"github.com/kegliz/qplay/qc/simulator"
	_ "github.com/kegliz/qplay/qc/simulator/itsu" // Import to register the runner
	"github.com/kegliz/qplay/qc/testutil"
)

// TestBasicFunctionality tests that the benchmark framework works correctly
func TestBasicFunctionality(t *testing.T) {
	// Test that runners are available
	runners := simulator.ListRunners()
	if len(runners) == 0 {
		t.Skip("No runners registered - this is expected if itsu package is not imported")
	}

	t.Logf("Available runners: %v", runners)

	// Test circuit creation
	for circuitType, builder := range StandardCircuits {
		t.Run(string(circuitType), func(t *testing.T) {
			build := builder(2)
			_, err := build.BuildCircuit()
			if err != nil {
				t.Errorf("Failed to build %s circuit: %v", circuitType, err)
			}
		})
	}

	// Test single benchmark execution if we have runners
	if len(runners) > 0 {
		config := BenchmarkConfig{
			CircuitType: SimpleCircuit,
			Scenario:    SerialExecution,
			Config:      testutil.QuickTestConfig,
			RunnerName:  runners[0],
			Limits:      DefaultResourceLimits, // Add resource limits
		}

		b := &testing.B{}
		result := RunSingleBenchmark(b, config)

		if !result.Success {
			t.Errorf("Benchmark failed: %s", result.Error)
		} else {
			t.Logf("Benchmark succeeded for %s in %v", runners[0], result.Duration)
		}
	}
}
