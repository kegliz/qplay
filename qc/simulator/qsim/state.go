// Package qsim implements a quantum circuit simulator from scratch
// This package provides a statevector-based quantum simulator that implements
// the OneShotRunner interface and enhanced capabilities for benchmarking and validation.
package qsim

import (
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kegliz/qplay/qc/gate"
)

// QSimRunner is a quantum circuit simulator built from scratch
type QSimRunner struct {
	config  map[string]interface{}
	mu      sync.RWMutex
	metrics QSimMetrics
	verbose bool
}

// QSimMetrics tracks execution statistics
type QSimMetrics struct {
	totalExecutions atomic.Int64
	successfulRuns  atomic.Int64
	failedRuns      atomic.Int64
	totalTime       atomic.Int64 // nanoseconds
	lastError       atomic.Value // string
	lastRunTime     atomic.Value // time.Time
}

// QuantumState represents the statevector of a quantum system
type QuantumState struct {
	numQubits     int
	amplitudes    []complex128 // State vector amplitudes
	numClassical  int          // Number of classical bits
	classicalBits []bool       // Classical bit values
}

// NewQSimRunner creates a new quantum simulator instance
func NewQSimRunner() *QSimRunner {
	runner := &QSimRunner{
		config:  make(map[string]interface{}),
		verbose: false,
	}

	// Initialize metrics
	runner.metrics.lastRunTime.Store(time.Time{})
	runner.metrics.lastError.Store("")

	return runner
}

// NewQuantumState creates a new quantum state with n qubits in |0...0⟩ state
func NewQuantumState(numQubits, numClassical int) *QuantumState {
	numStates := 1 << numQubits // 2^numQubits
	amplitudes := make([]complex128, numStates)
	amplitudes[0] = 1.0 // |0...0⟩ state has amplitude 1

	return &QuantumState{
		numQubits:     numQubits,
		amplitudes:    amplitudes,
		numClassical:  numClassical,
		classicalBits: make([]bool, numClassical),
	}
}

// Clone creates a deep copy of the quantum state
func (qs *QuantumState) Clone() *QuantumState {
	newState := &QuantumState{
		numQubits:     qs.numQubits,
		amplitudes:    make([]complex128, len(qs.amplitudes)),
		numClassical:  qs.numClassical,
		classicalBits: make([]bool, len(qs.classicalBits)),
	}

	copy(newState.amplitudes, qs.amplitudes)
	copy(newState.classicalBits, qs.classicalBits)

	return newState
}

// Normalize ensures the state vector has unit magnitude
func (qs *QuantumState) Normalize() {
	var norm float64
	// Optimized norm calculation
	for i := 0; i < len(qs.amplitudes); i++ {
		amp := qs.amplitudes[i]
		norm += real(amp)*real(amp) + imag(amp)*imag(amp)
	}

	if norm > 1e-10 { // Avoid division by zero
		norm = math.Sqrt(norm)
		invNorm := complex(1.0/norm, 0)
		for i := 0; i < len(qs.amplitudes); i++ {
			qs.amplitudes[i] *= invNorm
		}
	}
}

// GetProbabilities returns measurement probabilities for each computational basis state
func (qs *QuantumState) GetProbabilities() []float64 {
	probs := make([]float64, len(qs.amplitudes))
	// Optimized probability calculation using manual loop unrolling
	for i := range qs.amplitudes {
		amp := qs.amplitudes[i]
		probs[i] = real(amp)*real(amp) + imag(amp)*imag(amp)
	}
	return probs
}

// Measure performs a measurement of specified qubit and collapses the state
func (qs *QuantumState) Measure(qubit int) bool {
	if qubit >= qs.numQubits {
		return false // Invalid qubit
	}

	// Calculate probability of measuring |1⟩
	var probOne float64
	mask := 1 << qubit

	// Optimized probability calculation
	for i := mask; i < len(qs.amplitudes); i += 2 << qubit {
		end := min(i+(1<<qubit), len(qs.amplitudes))
		for j := i; j < end; j++ {
			amp := qs.amplitudes[j]
			probOne += real(amp * cmplx.Conj(amp))
		}
	}

	// Perform measurement
	result := rand.Float64() < probOne

	// Collapse the state - optimized normalization
	var norm float64
	if result {
		// Keep |1⟩ states, zero |0⟩ states
		for i := range qs.amplitudes {
			if (i & mask) != 0 {
				amp := qs.amplitudes[i]
				norm += real(amp * cmplx.Conj(amp))
			} else {
				qs.amplitudes[i] = 0
			}
		}
	} else {
		// Keep |0⟩ states, zero |1⟩ states
		for i := range qs.amplitudes {
			if (i & mask) == 0 {
				amp := qs.amplitudes[i]
				norm += real(amp * cmplx.Conj(amp))
			} else {
				qs.amplitudes[i] = 0
			}
		}
	}

	// Renormalize
	if norm > 1e-10 {
		norm = math.Sqrt(norm)
		invNorm := complex(1.0/norm, 0)
		for i := range qs.amplitudes {
			if (i&mask != 0) == result {
				qs.amplitudes[i] *= invNorm
			}
		}
	}

	return result
}

// ApplyGate applies a quantum gate to the state
func (qs *QuantumState) ApplyGate(g gate.Gate, qubits []int) error {
	switch g.Name() {
	case "H":
		return qs.applyHadamard(qubits[0])
	case "X":
		return qs.applyPauliX(qubits[0])
	case "Y":
		return qs.applyPauliY(qubits[0])
	case "Z":
		return qs.applyPauliZ(qubits[0])
	case "S":
		return qs.applyS(qubits[0])
	case "CNOT":
		return qs.applyCNOT(qubits[0], qubits[1])
	case "CZ":
		return qs.applyCZ(qubits[0], qubits[1])
	case "SWAP":
		return qs.applySwap(qubits[0], qubits[1])
	case "TOFFOLI":
		return qs.applyToffoli(qubits[0], qubits[1], qubits[2])
	case "FREDKIN":
		return qs.applyFredkin(qubits[0], qubits[1], qubits[2])
	default:
		return fmt.Errorf("unsupported gate: %s", g.Name())
	}
}

// Single-qubit gate implementations

func (qs *QuantumState) applyHadamard(qubit int) error {
	if qubit >= qs.numQubits {
		return fmt.Errorf("invalid qubit %d for %d-qubit system", qubit, qs.numQubits)
	}

	mask := 1 << qubit
	invSqrt2 := complex(1.0/math.Sqrt(2), 0)

	// Process only half the states (avoid double processing)
	// Work in-place to avoid memory allocation
	for i := 0; i < len(qs.amplitudes); i++ {
		if (i & mask) == 0 { // |0⟩ component
			j := i | mask // Corresponding |1⟩ state
			a0, a1 := qs.amplitudes[i], qs.amplitudes[j]
			qs.amplitudes[i] = invSqrt2 * (a0 + a1)
			qs.amplitudes[j] = invSqrt2 * (a0 - a1)
		}
	}

	return nil
}

func (qs *QuantumState) applyPauliX(qubit int) error {
	if qubit >= qs.numQubits {
		return fmt.Errorf("invalid qubit %d for %d-qubit system", qubit, qs.numQubits)
	}

	mask := 1 << qubit

	// Optimized X gate: only process pairs once
	for i := range qs.amplitudes {
		if (i & mask) == 0 { // Only process |0⟩ states
			j := i | mask // Corresponding |1⟩ state
			qs.amplitudes[i], qs.amplitudes[j] = qs.amplitudes[j], qs.amplitudes[i]
		}
	}

	return nil
}

func (qs *QuantumState) applyPauliY(qubit int) error {
	if qubit >= qs.numQubits {
		return fmt.Errorf("invalid qubit %d for %d-qubit system", qubit, qs.numQubits)
	}

	mask := 1 << qubit
	i := complex(0, 1) // Imaginary unit

	// Optimized Y gate: only process pairs once
	for idx := range qs.amplitudes {
		if (idx & mask) == 0 { // Only process |0⟩ states
			j := idx | mask // Corresponding |1⟩ state
			temp := qs.amplitudes[idx]
			qs.amplitudes[idx] = -i * qs.amplitudes[j]
			qs.amplitudes[j] = i * temp
		}
	}

	return nil
}

func (qs *QuantumState) applyPauliZ(qubit int) error {
	if qubit >= qs.numQubits {
		return fmt.Errorf("invalid qubit %d for %d-qubit system", qubit, qs.numQubits)
	}

	mask := 1 << qubit

	for i := range qs.amplitudes {
		if (i & mask) != 0 { // |1⟩ component gets phase flip
			qs.amplitudes[i] = -qs.amplitudes[i]
		}
	}

	return nil
}

func (qs *QuantumState) applyS(qubit int) error {
	if qubit >= qs.numQubits {
		return fmt.Errorf("invalid qubit %d for %d-qubit system", qubit, qs.numQubits)
	}

	mask := 1 << qubit
	i := complex(0, 1) // Imaginary unit

	for idx := range qs.amplitudes {
		if (idx & mask) != 0 { // |1⟩ component gets i phase
			qs.amplitudes[idx] = i * qs.amplitudes[idx]
		}
	}

	return nil
}

// Two-qubit gate implementations

func (qs *QuantumState) applyCNOT(control, target int) error {
	if control >= qs.numQubits || target >= qs.numQubits {
		return fmt.Errorf("invalid qubits %d,%d for %d-qubit system", control, target, qs.numQubits)
	}

	controlMask := 1 << control
	targetMask := 1 << target

	// Only process states where control is |1⟩ and target is |0⟩
	for i := 0; i < len(qs.amplitudes); i++ {
		if (i&controlMask) != 0 && (i&targetMask) == 0 {
			j := i | targetMask
			qs.amplitudes[i], qs.amplitudes[j] = qs.amplitudes[j], qs.amplitudes[i]
		}
	}

	return nil
}

func (qs *QuantumState) applyCZ(control, target int) error {
	if control >= qs.numQubits || target >= qs.numQubits {
		return fmt.Errorf("invalid qubits %d,%d for %d-qubit system", control, target, qs.numQubits)
	}

	controlMask := 1 << control
	targetMask := 1 << target

	for i := range qs.amplitudes {
		if (i&controlMask) != 0 && (i&targetMask) != 0 { // Both |1⟩
			qs.amplitudes[i] = -qs.amplitudes[i]
		}
	}

	return nil
}

func (qs *QuantumState) applySwap(qubit1, qubit2 int) error {
	if qubit1 >= qs.numQubits || qubit2 >= qs.numQubits {
		return fmt.Errorf("invalid qubits %d,%d for %d-qubit system", qubit1, qubit2, qs.numQubits)
	}

	mask1 := 1 << qubit1
	mask2 := 1 << qubit2

	// Optimized SWAP: only process states where qubits have different values
	for i := range qs.amplitudes {
		if (i&mask1) != 0 && (i&mask2) == 0 { // qubit1=1, qubit2=0
			j := (i &^ mask1) | mask2 // qubit1=0, qubit2=1
			qs.amplitudes[i], qs.amplitudes[j] = qs.amplitudes[j], qs.amplitudes[i]
		}
	}

	return nil
}

// Three-qubit gate implementations

func (qs *QuantumState) applyToffoli(control1, control2, target int) error {
	if control1 >= qs.numQubits || control2 >= qs.numQubits || target >= qs.numQubits {
		return fmt.Errorf("invalid qubits %d,%d,%d for %d-qubit system", control1, control2, target, qs.numQubits)
	}

	mask1 := 1 << control1
	mask2 := 1 << control2
	targetMask := 1 << target
	controlMask := mask1 | mask2

	// Only process states where both controls are |1⟩ and target is |0⟩
	for i := range qs.amplitudes {
		if (i&controlMask) == controlMask && (i&targetMask) == 0 {
			j := i | targetMask
			qs.amplitudes[i], qs.amplitudes[j] = qs.amplitudes[j], qs.amplitudes[i]
		}
	}

	return nil
}

func (qs *QuantumState) applyFredkin(control, target1, target2 int) error {
	if control >= qs.numQubits || target1 >= qs.numQubits || target2 >= qs.numQubits {
		return fmt.Errorf("invalid qubits %d,%d,%d for %d-qubit system", control, target1, target2, qs.numQubits)
	}

	controlMask := 1 << control
	mask1 := 1 << target1
	mask2 := 1 << target2

	for i := range qs.amplitudes {
		if (i & controlMask) != 0 { // Control is |1⟩
			bit1 := (i & mask1) != 0
			bit2 := (i & mask2) != 0

			if bit1 != bit2 { // Only swap if bits are different
				j := i
				if bit1 { // target1 is 1, target2 is 0
					j = (i &^ mask1) | mask2 // Set target1 to 0, target2 to 1
				} else { // target1 is 0, target2 is 1
					j = (i &^ mask2) | mask1 // Set target1 to 1, target2 to 0
				}

				qs.amplitudes[i], qs.amplitudes[j] = qs.amplitudes[j], qs.amplitudes[i]
			}
		}
	}

	return nil
}
