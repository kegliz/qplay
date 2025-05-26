package main

import (
	"fmt"
	"log"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/kegliz/qplay/qc/renderer"
	"github.com/kegliz/qplay/qc/simulator"

	// Import the itsu package to register the plugin
	_ "github.com/kegliz/qplay/qc/simulator/itsu"
)

func main() {
	fmt.Println("🔥 Z Gate (Pauli-Z) Demonstration")
	fmt.Println("=================================")

	// Simple Z gate demonstration
	fmt.Println("\n1. Basic Z Gate Circuit:")
	runBasicZGateDemo()

	fmt.Println("\n2. Z Gate with Hadamard:")
	runHadamardZDemo()

	fmt.Println("\n🎉 Z Gate demonstration completed!")
}

func runBasicZGateDemo() {
	// Create circuit with Z gate
	b := builder.New(builder.Q(1), builder.C(1))
	b.Z(0).Measure(0, 0)

	// Build and run circuit
	circ, err := b.BuildCircuit()
	if err != nil {
		log.Fatalf("Failed to build circuit: %v", err)
	}

	runAndDisplay("Basic Z Gate", circ, 1024)
}

func runHadamardZDemo() {
	// Create circuit with H-Z-H sequence
	b := builder.New(builder.Q(1), builder.C(1))
	b.H(0).Z(0).H(0).Measure(0, 0)

	// Build and run circuit
	circ, err := b.BuildCircuit()
	if err != nil {
		log.Fatalf("Failed to build circuit: %v", err)
	}

	runAndDisplay("Hadamard + Z + Hadamard", circ, 1024)
}

func runAndDisplay(name string, circ circuit.Circuit, shots int) {
	// Create runner
	runner, err := simulator.CreateRunner("itsu")
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	// Create and run simulator
	sim := simulator.NewSimulator(simulator.SimulatorOptions{
		Shots:  shots,
		Runner: runner,
	})

	results, err := sim.RunSerial(circ)
	if err != nil {
		log.Fatalf("Failed to run circuit '%s': %v", name, err)
	}

	// Display results
	fmt.Printf("   Circuit: %s\n", name)
	fmt.Printf("   Shots: %d\n", shots)
	fmt.Printf("   Results:\n")

	for state, count := range results {
		probability := float64(count) / float64(shots)
		fmt.Printf("     |%s⟩: %d counts (%.1f%%)\n", state, count, probability*100)
	}

	// Save circuit diagram
	filename := fmt.Sprintf("z-gate-%s.png", sanitizeName(name))
	r := renderer.NewRenderer(80)
	if err := r.Save(filename, circ); err != nil {
		log.Printf("Warning: Failed to save diagram '%s': %v", filename, err)
	} else {
		fmt.Printf("   📊 Circuit diagram saved as: %s\n", filename)
	}
}

func sanitizeName(name string) string {
	result := ""
	for _, char := range name {
		if char == ' ' {
			result += "-"
		} else if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' {
			result += string(char)
		}
	}
	return result
}
