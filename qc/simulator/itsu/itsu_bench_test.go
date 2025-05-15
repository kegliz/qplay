package itsu

import (
	"runtime" // Added runtime import
	"testing"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/renderer"
	"github.com/kegliz/qplay/qc/simulator"
	//"github.com/kegliz/qplay/qc/simulator/itsu"
)

// complexCircuit creates a moderately complex circuit for benchmarking.
// It applies H to all qubits, then a chain of CNOTs, then measures all.
func complexCircuit(numQubits int) builder.Builder {
	b := builder.New(builder.Q(numQubits), builder.C(numQubits))
	// Apply H to all qubits
	for i := range numQubits {
		b.H(i)
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

const shots = 1024 * 8 // Number of shots for the benchmark
const numBenchmarkQubits = 7

func BenchmarkSerial(b *testing.B) {
	build := complexCircuit(numBenchmarkQubits)
	circ, err := build.BuildCircuit()
	if err != nil {
		b.Fatalf("build error: %v", err)
	}

	renderer := renderer.NewRenderer(80)
	filePath1 := "benchmark.png"
	//defer os.Remove(filePath1) // Clean up

	err = renderer.Save(filePath1, circ) // Save first circuit
	if err != nil {
		b.Fatalf("image save error: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer() // Reset timer after setup
	for i := 0; i < b.N; i++ {
		sim := simulator.NewSimulator(simulator.SimulatorOptions{Shots: shots, Workers: 0, Runner: NewItsuOneShotRunner()}) // Added Workers
		sim.SetVerbose(true)
		if _, err := sim.RunSerial(circ); err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}

func BenchmarkParallel(b *testing.B) {
	build := complexCircuit(numBenchmarkQubits)
	circ, err := build.BuildCircuit()
	if err != nil {
		b.Fatalf("build error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer() // Reset timer after setup
	for i := 0; i < b.N; i++ {
		sim := simulator.NewSimulator(simulator.SimulatorOptions{Shots: shots, Workers: runtime.NumCPU(), Runner: NewItsuOneShotRunner()}) // Added Workers
		sim.SetVerbose(true)
		// s.Workers is set by New, no need to set it again here
		if _, err := sim.RunParallelChan(circ); err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}

// BenchmarkParallelStatic is a benchmark for the static partitioning of the parallel run.
func BenchmarkParallelStatic(b *testing.B) {
	build := complexCircuit(numBenchmarkQubits)
	circ, err := build.BuildCircuit()
	if err != nil {
		b.Fatalf("build error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer() // Reset timer after setup
	for i := 0; i < b.N; i++ {
		sim := simulator.NewSimulator(simulator.SimulatorOptions{Shots: shots, Workers: runtime.NumCPU(), Runner: NewItsuOneShotRunner()}) // Added Workers
		sim.SetVerbose(true)
		// s.Workers is set by New, no need to set it again here
		if _, err := sim.RunParallelStatic(circ); err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}

// BenchmarkParallelStatic is a benchmark for the static partitioning of the parallel run.
func BenchmarkPooledParallelStatic(b *testing.B) {
	build := complexCircuit(numBenchmarkQubits)
	circ, err := build.BuildCircuit()
	if err != nil {
		b.Fatalf("build error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer() // Reset timer after setup
	for i := 0; i < b.N; i++ {
		sim := simulator.NewSimulator(simulator.SimulatorOptions{Shots: shots, Workers: runtime.NumCPU(), Runner: NewPooledItsuOneShotRunner()}) // Added Workers
		sim.SetVerbose(true)
		// s.Workers is set by New, no need to set it again here
		if _, err := sim.RunParallelStatic(circ); err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}
