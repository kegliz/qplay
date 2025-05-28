// Package benchmark provides a standardized benchmarking framework for quantum backend plugins.
// It offers consistent benchmark circuits and scenarios that work across all registered backends.
package benchmark

import (
	"github.com/kegliz/qplay/qc/builder"
)

// CircuitType represents different categories of benchmark circuits
type CircuitType string

const (
	SimpleCircuit        CircuitType = "simple"        // Basic H + Measure
	EntanglementCircuit  CircuitType = "entanglement"  // H + CNOT + Measure
	SuperpositionCircuit CircuitType = "superposition" // Multiple H gates
	MixedGatesCircuit    CircuitType = "mixed"         // Variety of gates
)

// CircuitBuilder defines a function that creates a benchmark circuit
type CircuitBuilder func(qubits int) builder.Builder

// StandardCircuits contains predefined benchmark circuits for consistent testing
var StandardCircuits = map[CircuitType]CircuitBuilder{
	SimpleCircuit:        buildSimpleCircuit,
	EntanglementCircuit:  buildEntanglementCircuit,
	SuperpositionCircuit: buildSuperpositionCircuit,
	MixedGatesCircuit:    buildMixedGatesCircuit,
}

// buildSimpleCircuit creates a basic H + Measure circuit
// This tests fundamental gate application and measurement
func buildSimpleCircuit(qubits int) builder.Builder {
	if qubits < 1 {
		qubits = 1
	}

	b := builder.New(builder.Q(qubits), builder.C(qubits))

	// Apply Hadamard to first qubit only (simple)
	b.H(0)

	// Measure the first qubit
	b.Measure(0, 0)

	return b
}

// buildEntanglementCircuit creates an H + CNOT + Measure circuit
// This tests multi-qubit operations and entanglement
func buildEntanglementCircuit(qubits int) builder.Builder {
	if qubits < 2 {
		qubits = 2
	}

	b := builder.New(builder.Q(qubits), builder.C(qubits))

	// Create Bell state: H on first qubit, then CNOT
	b.H(0)
	b.CNOT(0, 1)

	// Measure both qubits
	b.Measure(0, 0)
	b.Measure(1, 1)

	return b
}

// buildSuperpositionCircuit creates multiple H gates + measurements
// This tests scaling with multiple superposition states
func buildSuperpositionCircuit(qubits int) builder.Builder {
	if qubits < 1 {
		qubits = 1
	}

	b := builder.New(builder.Q(qubits), builder.C(qubits))

	// Apply Hadamard to all qubits (up to a reasonable limit for benchmarking)
	maxQubits := min(qubits, 4) // Limit for benchmark performance
	for i := 0; i < maxQubits; i++ {
		b.H(i)
	}

	// Measure all used qubits
	for i := 0; i < maxQubits; i++ {
		b.Measure(i, i)
	}

	return b
}

// buildMixedGatesCircuit creates a circuit with variety of gates
// This tests backend support for different gate types
func buildMixedGatesCircuit(qubits int) builder.Builder {
	if qubits < 2 {
		qubits = 2
	}

	b := builder.New(builder.Q(qubits), builder.C(qubits))

	// Use at most 3 qubits for mixed circuit to keep it simple but meaningful
	maxQubits := min(qubits, 3)

	// Apply different single-qubit gates
	for i := 0; i < maxQubits; i++ {
		switch i % 4 {
		case 0:
			b.H(i) // Hadamard
		case 1:
			b.X(i) // Pauli-X
		case 2:
			b.Y(i) // Pauli-Y
		case 3:
			b.Z(i) // Pauli-Z
		}
	}

	// Add some two-qubit gates if we have enough qubits
	if maxQubits >= 2 {
		b.CNOT(0, 1)
	}
	if maxQubits >= 3 {
		b.CZ(1, 2)
	}

	// Measure all used qubits
	for i := 0; i < maxQubits; i++ {
		b.Measure(i, i)
	}

	return b
}

// GetCircuitDescription returns a human-readable description of the circuit type
func GetCircuitDescription(circuitType CircuitType) string {
	switch circuitType {
	case SimpleCircuit:
		return "Simple H + Measure (tests basic gates)"
	case EntanglementCircuit:
		return "H + CNOT + Measure (tests entanglement)"
	case SuperpositionCircuit:
		return "Multiple H + Measure (tests superposition scaling)"
	case MixedGatesCircuit:
		return "Mixed gates + CNOT + Measure (tests gate variety)"
	default:
		return "Unknown circuit type"
	}
}

// min returns the minimum of two integers (helper function)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
