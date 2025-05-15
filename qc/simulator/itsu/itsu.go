package itsu

import (
	"fmt"

	"github.com/itsubaki/q"
	"github.com/kegliz/qplay/internal/logger"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/kegliz/qplay/qc/simulator"
	"github.com/rs/zerolog"
)

type ItsuOneShotRunner struct {
	log logger.Logger
}

func NewItsuOneShotRunner() *ItsuOneShotRunner {
	return &ItsuOneShotRunner{
		log: *logger.NewLogger(logger.LoggerOptions{
			Debug: false,
		}),
	}
}
func (s *ItsuOneShotRunner) SetVerbose(verbose bool) {
	if verbose {
		s.log.Logger = s.log.Logger.Level(zerolog.DebugLevel) // Log all messages if verbose
	} else {
		s.log.Logger = s.log.Logger.Level(zerolog.InfoLevel)
	}
}

func (s *ItsuOneShotRunner) RunOnce(c circuit.Circuit) (string, error) {
	sim := q.New()
	return runOnce(sim, c)
}

// runOnce plays the circuit exactly one time on the provided simulator,
// returning the measured classical bit‑string.
func runOnce(sim *q.Q, c circuit.Circuit) (string, error) {
	qs := sim.ZeroWith(c.Qubits())
	//cbits := bytes.Repeat([]byte{'0'}, c.Clbits())
	cbits := make([]byte, c.Clbits())
	for i := range cbits {
		cbits[i] = '0' // Explicitly initialize to '0'
	}

	for i, op := range c.Operations() {
		// Check qubit indices are valid for the gate's operation before applying
		// (This is defensive programming; circuit/DAG validation should catch this)
		for _, qIndex := range op.Qubits {
			if qIndex < 0 || qIndex >= len(qs) {
				// Add operation index to error message
				return "", fmt.Errorf("itsu: invalid qubit index %d for gate %s (op %d) in runOnce", qIndex, op.G.Name(), i)
			}
		}
		if op.G.Name() == "MEASURE" && (op.Cbit < 0 || op.Cbit >= len(cbits)) {
			// Add operation index to error message
			return "", fmt.Errorf("itsu: invalid classical bit index %d for MEASURE (op %d) in runOnce", op.Cbit, i)
		}

		switch op.G.Name() {
		case "H":
			sim.H(qs[op.Qubits[0]])
		case "X":
			sim.X(qs[op.Qubits[0]])
		case "S":
			sim.S(qs[op.Qubits[0]])
		case "CNOT":
			sim.CNOT(qs[op.Qubits[0]], qs[op.Qubits[1]])
		case "CZ":
			sim.CZ(qs[op.Qubits[0]], qs[op.Qubits[1]])
		case "SWAP":
			sim.Swap(qs[op.Qubits[0]], qs[op.Qubits[1]])
		case "TOFFOLI":
			sim.Toffoli(qs[op.Qubits[0]], qs[op.Qubits[1]], qs[op.Qubits[2]])
		case "FREDKIN":
			ctrl, a, b := qs[op.Qubits[0]], qs[op.Qubits[1]], qs[op.Qubits[2]]
			// Standard decomposition: CNOT(b,a) Toffoli(ctrl,a,b) CNOT(b,a)
			sim.CNOT(b, a)
			sim.Toffoli(ctrl, a, b)
			sim.CNOT(b, a)
		case "MEASURE":
			m := sim.Measure(qs[op.Qubits[0]]) // collapses state & returns result
			if m.IsOne() {
				cbits[op.Cbit] = '1'
			} else {
				cbits[op.Cbit] = '0'
			}
		default:
			// Add operation index to error message
			return "", fmt.Errorf("itsu: unsupported gate %s (op %d) encountered in runOnce", op.G.Name(), i)
		}
	}
	// Return the final classical bit string (little-endian)
	return string(cbits), nil
}

// check that ItsuOneShotRunner implements the OneShotRunner interface
var _ simulator.OneShotRunner = (*ItsuOneShotRunner)(nil)
