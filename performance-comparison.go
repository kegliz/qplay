package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/kegliz/qplay/qc/simulator"
	_ "github.com/kegliz/qplay/qc/simulator/itsu"
	_ "github.com/kegliz/qplay/qc/simulator/qsim"
)

type BenchmarkResult struct {
	Name      string
	QSimTime  time.Duration
	ItsuTime  time.Duration
	Speedup   float64
	Circuit   string
}

func createSimpleCircuit() circuit.Circuit {
	b := builder.New(builder.Q(1), builder.C(1))
	b.H(0) // Hadamard gate
	b.Measure(0, 0)
	circ, _ := b.BuildCircuit()
	return circ
}

func createBellState() circuit.Circuit {
	b := builder.New(builder.Q(2), builder.C(2))
	b.H(0) // Hadamard
	b.CNOT(0, 1) // CNOT to create entanglement
	b.Measure(0, 0)
	b.Measure(1, 1)
	circ, _ := b.BuildCircuit()
	return circ
}

func create3QubitSuperposition() circuit.Circuit {
	b := builder.New(builder.Q(3), builder.C(3))
	b.H(0)
	b.H(1)
	b.H(2)
	b.Measure(0, 0)
	b.Measure(1, 1)
	b.Measure(2, 2)
	circ, _ := b.BuildCircuit()
	return circ
}

func createComplexCircuit() circuit.Circuit {
	b := builder.New(builder.Q(3), builder.C(3))
	// Complex multi-qubit circuit
	b.H(0)
	b.H(1)
	b.CNOT(0, 1)
	b.X(2)
	b.Y(1)
	b.Z(0)
	b.CNOT(1, 2)
	b.CNOT(0, 2)
	b.H(2)
	for i := 0; i < 3; i++ {
		b.Measure(i, i)
	}
	circ, _ := b.BuildCircuit()
	return circ
}

func createDeepCircuit() circuit.Circuit {
	b := builder.New(builder.Q(3), builder.C(3))
	// Deep circuit with many layers
	for layer := 0; layer < 10; layer++ {
		b.H(0)
		b.X(1)
		b.Y(2)
		b.CNOT(0, 1)
		b.CNOT(1, 2)
	}
	for i := 0; i < 3; i++ {
		b.Measure(i, i)
	}
	circ, _ := b.BuildCircuit()
	return circ
}

func benchmarkRunner(runner simulator.OneShotRunner, circ circuit.Circuit, iterations int) time.Duration {
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := runner.RunOnce(circ)
		if err != nil {
			log.Printf("Error during benchmark: %v", err)
		}
	}
	return time.Since(start)
}

func main() {
	fmt.Println("🚀 QSim vs Itsubaki Performance Comparison")
	fmt.Println("===========================================")

	// Create runners
	qsimRunner, err := simulator.CreateRunner("qsim")
	if err != nil {
		log.Fatal("Failed to create QSim runner:", err)
	}

	itsuRunner, err := simulator.CreateRunner("itsu")
	if err != nil {
		log.Fatal("Failed to create Itsu runner:", err)
	}

	// Test circuits
	tests := []struct {
		name    string
		circuit circuit.Circuit
		iters   int
	}{
		{"Simple H+Measure", createSimpleCircuit(), 10000},
		{"Bell State", createBellState(), 10000},
		{"3-Qubit Superposition", create3QubitSuperposition(), 5000},
		{"Complex Multi-gate", createComplexCircuit(), 2000},
		{"Deep Circuit (10 layers)", createDeepCircuit(), 1000},
	}

	var results []BenchmarkResult

	fmt.Printf("%-25s %-12s %-12s %-10s %s\n", "Circuit", "QSim", "Itsubaki", "Speedup", "Description")
	fmt.Printf("%-25s %-12s %-12s %-10s %s\n", "=======", "====", "========", "=======", "===========")

	for _, test := range tests {
		fmt.Printf("Benchmarking %s (%d iterations)...\n", test.name, test.iters)

		// Benchmark QSim
		qsimTime := benchmarkRunner(qsimRunner, test.circuit, test.iters)

		// Benchmark Itsubaki
		itsuTime := benchmarkRunner(itsuRunner, test.circuit, test.iters)

		// Calculate speedup
		speedup := float64(itsuTime) / float64(qsimTime)

		result := BenchmarkResult{
			Name:     test.name,
			QSimTime: qsimTime,
			ItsuTime: itsuTime,
			Speedup:  speedup,
			Circuit:  fmt.Sprintf("%d iterations", test.iters),
		}
		results = append(results, result)

		// Format times
		qsimPerOp := qsimTime / time.Duration(test.iters)
		itsuPerOp := itsuTime / time.Duration(test.iters)

		fmt.Printf("%-25s %-12s %-12s %-10.2fx %s\n",
			test.name,
			qsimPerOp.String(),
			itsuPerOp.String(),
			speedup,
			test.circuit)
	}

	fmt.Println("\n📊 Summary:")
	fmt.Println("============")

	var totalSpeedup float64
	for _, result := range results {
		totalSpeedup += result.Speedup
	}
	avgSpeedup := totalSpeedup / float64(len(results))

	fmt.Printf("Average Speedup: %.2fx\n", avgSpeedup)
	
	// Find best and worst cases
	var bestSpeedup, worstSpeedup BenchmarkResult
	bestSpeedup.Speedup = 0
	worstSpeedup.Speedup = 999999

	for _, result := range results {
		if result.Speedup > bestSpeedup.Speedup {
			bestSpeedup = result
		}
		if result.Speedup < worstSpeedup.Speedup {
			worstSpeedup = result
		}
	}

	fmt.Printf("Best Performance: %s (%.2fx faster)\n", bestSpeedup.Name, bestSpeedup.Speedup)
	fmt.Printf("Worst Performance: %s (%.2fx faster)\n", worstSpeedup.Name, worstSpeedup.Speedup)

	if avgSpeedup > 1.0 {
		fmt.Printf("\n✅ QSim is %.2fx faster than Itsubaki on average!\n", avgSpeedup)
	} else {
		fmt.Printf("\n⚠️  QSim is %.2fx slower than Itsubaki on average.\n", 1.0/avgSpeedup)
	}
}
