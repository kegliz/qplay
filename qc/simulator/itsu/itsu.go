package itsu

import (
	// "github.com/itsubaki/q"
	"github.com/kegliz/qplay/qc/circuit"
)

// Simulator implements a simple quantum simulator using the q package.
// It runs a circuit for a specified number of shots and returns the histogram of results.
// It uses a topological sort to determine the order of operations.
// It supports a limited set of gates: H, X, S, CNOT, SWAP, Toffoli, Fredkin, and MEASURE.
// The simulator uses a classical bitstring to record measurement results.
// The simulator is not optimized for performance and is intended for educational purposes only.
type Simulator struct{ Shots int }

func New(shots int) *Simulator { return &Simulator{Shots: shots} }

func (s *Simulator) Run(c circuit.Circuit) (hist map[string]int, err error) {
	hist = make(map[string]int)
	// TODO

	return hist, nil
}
