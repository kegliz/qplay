package circuit

import "github.com/kegliz/qplay/qc/gate"

// NOTE: This file defines CircuitStruct and associated methods.
// This appears to be an older or alternative representation compared to the
// DAG-based Circuit interface and its implementation in circuit.go.
// Consider removing this file and gate/gatestruct.go if the DAG-based
// approach is the primary method.

// ---------------------
// ---------------------
// ---------------------
// CircuitStruct is an immutable slice of Operations plus qubit count.
type CircuitStruct struct {
	qubits int
	Gates  []*gate.GateStruct
}

func NewCircuit(qubits int) *CircuitStruct {
	return &CircuitStruct{qubits: qubits}
}

// Qubits returns the number of qubits in the circuit
func (c *CircuitStruct) Qubits() int {
	return c.qubits
}

// Add adds a gate to the circuit
func (c *CircuitStruct) Add(gate *gate.GateStruct) *CircuitStruct {

	c.Gates = append(c.Gates, gate)
	return c
}

func (c *CircuitStruct) H(target int) *CircuitStruct {
	return c.Add(gate.NewHGate(target))
}
func (c *CircuitStruct) X(target int) *CircuitStruct {
	return c.Add(gate.NewXGate(target))
}
func (c *CircuitStruct) Z(target int) *CircuitStruct {
	return c.Add(gate.NewZGate(target))
}
func (c *CircuitStruct) CNot(control, target int) *CircuitStruct {
	return c.Add(gate.NewCNotGate(control, target))
}
func (c *CircuitStruct) Toffoli(control1, control2, target int) *CircuitStruct {
	return c.Add(gate.NewToffoliGate(control1, control2, target))
}
func (c *CircuitStruct) Swap(target1, target2 int) *CircuitStruct {
	return c.Add(gate.NewSwapGate(target1, target2))
}
func (c *CircuitStruct) Fredkin(control, target1, target2 int) *CircuitStruct {
	return c.Add(gate.NewFredkinGate(control, target1, target2))
}
func (c *CircuitStruct) Measure(target int) *CircuitStruct {
	return c.Add(gate.NewMeasurement(target))
}
