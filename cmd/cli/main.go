package main

import (
	"fmt"
	"sort" // Import the sort package

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/simulator"
	"github.com/kegliz/qplay/qc/simulator/itsu"
)

func main() {
	shots := 1024

	fmt.Println("--- Bell State Simulation ---")
	simulateBellState(shots)
	fmt.Println("\n--- 2-Qubit Grover Simulation (|11>) ---")
	simulateGrover2Qubit(shots)
	fmt.Println("\n--- 3-Qubit Grover Simulation (|111>) ---")
	simulateGrover3Qubit(shots)
}

// simulateBellState prepares the |Φ⁺⟩ Bell state and checks ~50/50 statistics.
func simulateBellState(shots int) {
	b := builder.New(builder.Q(2), builder.C(2))
	b.H(0).CNOT(0, 1).Measure(0, 0).Measure(1, 1)

	c, err := b.BuildCircuit()
	if err != nil {
		fmt.Printf("Error building Bell state circuit: %v\n", err)
		return
	}

	sim := simulator.NewSimulator(simulator.SimulatorOptions{Shots: shots, Runner: itsu.NewItsuOneShotRunner()})
	hist, err := sim.Run(c)
	if err != nil {
		fmt.Printf("Error running Bell state simulation: %v\n", err)
		return
	}

	pretty(hist, shots)
}

// simulateGrover2Qubit demonstrates one Grover iteration on 2‑qubit search space
// amplifying the |11⟩ state.
func simulateGrover2Qubit(shots int) {
	b := builder.New(builder.Q(2), builder.C(2))

	// — initial superposition —
	b.H(0).H(1)

	// — oracle marks |11⟩ by phase flip (controlled‑Z) —
	b.CZ(0, 1)

	// — diffusion operator —
	b.H(0).H(1)
	b.X(0).X(1)
	b.CZ(0, 1)
	b.X(0).X(1)
	b.H(0).H(1)

	// — measurement —
	b.Measure(0, 0).Measure(1, 1)

	c, err := b.BuildCircuit()
	if err != nil {
		fmt.Printf("Error building 2-qubit Grover circuit: %v\n", err)
		return
	}

	sim := simulator.NewSimulator(simulator.SimulatorOptions{Shots: shots, Runner: itsu.NewItsuOneShotRunner()})
	hist, err := sim.Run(c)
	if err != nil {
		fmt.Printf("Error running 2-qubit Grover simulation: %v\n", err)
		return
	}

	pretty(hist, shots)
}

// simulateGrover3Qubit demonstrates one Grover iteration on 3‑qubit search space
// amplifying the |111⟩ state.
func simulateGrover3Qubit(shots int) {
	b := builder.New(builder.Q(3), builder.C(3))

	// — initial superposition —
	b.H(0).H(1).H(2)

	// — oracle marks |111⟩ by phase flip (CCZ) —
	// Implement CCZ using H and Toffoli: H(target) Toffoli(c1, c2, target) H(target)
	b.H(2).Toffoli(0, 1, 2).H(2)

	// — diffusion operator (3 qubits) —
	// HHH - XXX - CCZ - XXX - HHH
	b.H(0).H(1).H(2)
	b.X(0).X(1).X(2)
	// CCZ
	b.H(2).Toffoli(0, 1, 2).H(2)
	b.X(0).X(1).X(2)
	b.H(0).H(1).H(2)

	// — measurement —
	b.Measure(0, 0).Measure(1, 1).Measure(2, 2)

	c, err := b.BuildCircuit()

	if err != nil {
		fmt.Printf("Error building 3-qubit Grover circuit: %v\n", err)
		return
	}

	sim := simulator.NewSimulator(simulator.SimulatorOptions{Shots: shots, Runner: itsu.NewItsuOneShotRunner()})
	hist, err := sim.Run(c)
	if err != nil {
		fmt.Printf("Error running 3-qubit Grover simulation: %v\n", err)
		return
	}

	pretty(hist, shots)
}

// pretty prints the histogram results in a readable, sorted format
func pretty(hist map[string]int, shots int) {
	// Extract keys for sorting
	keys := make([]string, 0, len(hist))
	for k := range hist {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Sort keys alphabetically

	// Print sorted results
	for _, state := range keys {
		count := hist[state]
		probability := float64(count) / float64(shots)
		fmt.Printf("State |%s>: %d counts (%.2f%%)\n", state, count, probability*100)
	}
}
