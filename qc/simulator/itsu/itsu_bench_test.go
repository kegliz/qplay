package itsu

import (
	"testing"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/renderer"
	"github.com/kegliz/qplay/qc/simulator"
	"github.com/kegliz/qplay/qc/testutil"
)

// getBenchmarkConfig returns appropriate configuration based on test flags
func getBenchmarkConfig(b *testing.B) testutil.TestConfig {
	b.Helper()

	if testing.Short() {
		return testutil.QuickTestConfig
	}
	return testutil.BenchmarkTestConfig
}

// complexCircuit creates a moderately complex circuit for benchmarking.
// It applies H to all qubits, then Y/Z gates, then a chain of CNOTs, then measures all.
func complexCircuit(numQubits int) builder.Builder {
	b := builder.New(builder.Q(numQubits), builder.C(numQubits))
	// Apply H to all qubits
	for i := range numQubits {
		b.H(i)
	}
	// Apply Y gates to odd qubits and Z gates to even qubits for variety
	for i := range numQubits {
		if i%2 == 0 {
			b.Z(i)
		} else {
			b.Y(i)
		}
	}
	// Apply a chain of CNOTs
	for i := range numQubits - 1 {
		b.CNOT(i, i+1)
	}
	// Measure all qubits
	for i := range numQubits {
		b.Measure(i, i)
	}
	return b
}

func BenchmarkSerial(b *testing.B) {
	config := getBenchmarkConfig(b)

	build := complexCircuit(config.Qubits)
	circ, err := build.BuildCircuit()
	if err != nil {
		b.Fatalf("build error: %v", err)
	}

	// Test file creation and cleanup using testutil
	renderer := renderer.NewRenderer(80)
	filePath, cleanup := testutil.TempFileB(b, testutil.PNGTestSuffix)
	defer cleanup()

	err = renderer.Save(filePath, circ)
	if err != nil {
		b.Fatalf("image save error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer() // Reset timer after setup

	for i := 0; i < b.N; i++ {
		sim := simulator.NewSimulator(simulator.SimulatorOptions{
			Shots:  config.Shots,
			Runner: NewItsuOneShotRunner(),
		})
		sim.SetVerbose(false) // Disable verbose for benchmarks

		if _, err := sim.RunSerial(circ); err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}

func BenchmarkParallel(b *testing.B) {
	config := getBenchmarkConfig(b)

	build := complexCircuit(config.Qubits)
	circ, err := build.BuildCircuit()
	if err != nil {
		b.Fatalf("build error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer() // Reset timer after setup

	for i := 0; i < b.N; i++ {
		sim := simulator.NewSimulator(simulator.SimulatorOptions{
			Shots:  config.Shots,
			Runner: NewItsuOneShotRunner(),
		})
		sim.SetVerbose(false) // Disable verbose for benchmarks

		if _, err := sim.RunParallelChan(circ); err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}

// BenchmarkParallelStatic is a benchmark for the static partitioning of the parallel run.
func BenchmarkParallelStatic(b *testing.B) {
	config := getBenchmarkConfig(b)

	build := complexCircuit(config.Qubits)
	circ, err := build.BuildCircuit()
	if err != nil {
		b.Fatalf("build error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer() // Reset timer after setup

	for i := 0; i < b.N; i++ {
		sim := simulator.NewSimulator(simulator.SimulatorOptions{
			Shots:  config.Shots,
			Runner: NewItsuOneShotRunner(),
		})
		sim.SetVerbose(false) // Disable verbose for benchmarks

		if _, err := sim.RunParallelStatic(circ); err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}
