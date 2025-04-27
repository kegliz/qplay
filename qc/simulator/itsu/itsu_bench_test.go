package itsu_test

import (
	"runtime"
	"testing"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/simulator/itsu"
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

const shots = 4096
const numBenchmarkQubits = 10 // Use 10 qubits for benchmarks

func BenchmarkSerial(b *testing.B) {
	build := complexCircuit(numBenchmarkQubits) // Use complex circuit
	circ, err := build.BuildCircuit()
	if err != nil {
		b.Fatalf("build error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer() // Reset timer after setup
	for i := 0; i < b.N; i++ {
		sim := itsu.New(shots)
		sim.Workers = 1 // force single‑thread
		if _, err := sim.Run(circ); err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}

func BenchmarkParallel(b *testing.B) {
	build := complexCircuit(numBenchmarkQubits) // Use complex circuit
	circ, err := build.BuildCircuit()
	if err != nil {
		b.Fatalf("build error: %v", err)
	}

	workers := runtime.NumCPU()

	b.ReportAllocs()
	b.ResetTimer() // Reset timer after setup
	for i := 0; i < b.N; i++ {
		sim := itsu.New(shots)
		sim.Workers = workers
		if _, err := sim.Run(circ); err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}
