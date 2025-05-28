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
	for _, amp := range qs.amplitudes {
		norm += real(amp * cmplx.Conj(amp))
	}
	norm = math.Sqrt(norm)

	if norm > 1e-10 { // Avoid division by zero
		for i := range qs.amplitudes {
			qs.amplitudes[i] /= complex(norm, 0)
		}
	}
}

// GetProbabilities returns measurement probabilities for each computational basis state
func (qs *QuantumState) GetProbabilities() []float64 {
	probs := make([]float64, len(qs.amplitudes))
	for i, amp := range qs.amplitudes {
		probs[i] = real(amp * cmplx.Conj(amp))
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

	for i, amp := range qs.amplitudes {
		if (i & mask) != 0 { // Bit is set (|1⟩)
			probOne += real(amp * cmplx.Conj(amp))
		}
	}

	// Perform measurement
	result := rand.Float64() < probOne

	// Collapse the state
	var norm float64
	for i := range qs.amplitudes {
		bitSet := (i & mask) != 0
		if bitSet != result {
			qs.amplitudes[i] = 0 // Zero out incompatible amplitudes
		} else {
			norm += real(qs.amplitudes[i] * cmplx.Conj(qs.amplitudes[i]))
		}
	}

	// Renormalize
	if norm > 1e-10 {
		norm = math.Sqrt(norm)
		for i := range qs.amplitudes {
			if (i&mask != 0) == result {
				qs.amplitudes[i] /= complex(norm, 0)
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

	newAmplitudes := make([]complex128, len(qs.amplitudes))

	for i := range qs.amplitudes {
		if (i & mask) == 0 { // |0⟩ component
			j := i | mask // Corresponding |1⟩ state
			newAmplitudes[i] = invSqrt2 * (qs.amplitudes[i] + qs.amplitudes[j])
			newAmplitudes[j] = invSqrt2 * (qs.amplitudes[i] - qs.amplitudes[j])
		}
	}

	qs.amplitudes = newAmplitudes
	return nil
}

func (qs *QuantumState) applyPauliX(qubit int) error {
	if qubit >= qs.numQubits {
		return fmt.Errorf("invalid qubit %d for %d-qubit system", qubit, qs.numQubits)
	}

	mask := 1 << qubit

	for i := range qs.amplitudes {
		if (i & mask) == 0 { // |0⟩ component
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

	for idx := range qs.amplitudes {
		if (idx & mask) == 0 { // |0⟩ component
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

	for idx := 0; idx < len(qs.amplitudes); idx++ {
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

	for i := 0; i < len(qs.amplitudes); i++ {
		if (i & controlMask) != 0 { // Control is |1⟩
			if (i & targetMask) == 0 { // Target is |0⟩
				j := i | targetMask // Flip target to |1⟩
				qs.amplitudes[i], qs.amplitudes[j] = qs.amplitudes[j], qs.amplitudes[i]
			}
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

	for i := 0; i < len(qs.amplitudes); i++ {
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

	for i := 0; i < len(qs.amplitudes); i++ {
		bit1 := (i & mask1) != 0
		bit2 := (i & mask2) != 0

		if bit1 != bit2 { // Only swap if bits are different
			j := i
			if bit1 { // qubit1 is 1, qubit2 is 0
				j = (i &^ mask1) | mask2 // Set qubit1 to 0, qubit2 to 1
			} else { // qubit1 is 0, qubit2 is 1
				j = (i &^ mask2) | mask1 // Set qubit1 to 1, qubit2 to 0
			}

			if i < j { // Avoid double swapping
				qs.amplitudes[i], qs.amplitudes[j] = qs.amplitudes[j], qs.amplitudes[i]
			}
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

	for i := 0; i < len(qs.amplitudes); i++ {
		if (i&mask1) != 0 && (i&mask2) != 0 { // Both controls are |1⟩
			if (i & targetMask) == 0 { // Target is |0⟩
				j := i | targetMask // Flip target to |1⟩
				qs.amplitudes[i], qs.amplitudes[j] = qs.amplitudes[j], qs.amplitudes[i]
			}
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

	for i := 0; i < len(qs.amplitudes); i++ {
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
