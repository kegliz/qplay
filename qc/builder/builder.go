package builder

import (
	"fmt"

	"github.com/kegliz/qplay/qc/circuit"
	"github.com/kegliz/qplay/qc/dag"
	"github.com/kegliz/qplay/qc/gate"
)

// Builder implements a *fluent* declarative DSL for building quantum circuits.
type Builder interface {
	// Single-qubit gates
	H(q int) Builder
	X(q int) Builder
	S(q int) Builder

	// Multi-qubit gates
	CNOT(ctrl, tgt int) Builder
	CZ(ctrl, tgt int) Builder
	SWAP(q1, q2 int) Builder
	Toffoli(c1, c2, tgt int) Builder
	Fredkin(ctrl, t1, t2 int) Builder

	// Measurement
	Measure(q, cbit int) Builder

	// Finalise
	// BuildDAG returns a validated DAGReader interface.
	// It returns an error if the DAG is invalid.
	BuildDAG() (dag.DAGReader, error)
	BuildCircuit() (circuit.Circuit, error) // convenience façade
}

// New returns a fresh Builder with the requested qubits/classical bits.
func New(opts ...Option) Builder { return newBuilder(opts...) }

// ---------------------------- implementation -------------------------

type b struct {
	dagBuilder dag.DAGBuilder
	err        error
	built      bool
}

func newBuilder(opts ...Option) *b {
	cfg := config{qubits: 1}
	for _, o := range opts {
		o(&cfg)
	}
	return &b{dagBuilder: dag.New(cfg.qubits, cfg.clbits)}
}

// helper: bail-out pattern
func (b *b) bail(err error) Builder {
	if b.err == nil {
		b.err = err
	}
	return b
}

// Check if already built or if an error occurred
func (b *b) checkState() bool {
	return b.built || b.err != nil
}

func (b *b) H(q int) Builder               { return b.add1(gate.H(), q) }
func (b *b) X(q int) Builder               { return b.add1(gate.X(), q) }
func (b *b) S(q int) Builder               { return b.add1(gate.S(), q) }
func (b *b) CNOT(c, t int) Builder         { return b.add2(gate.CNOT(), c, t) }
func (b *b) CZ(c, t int) Builder           { return b.add2(gate.CZ(), c, t) }
func (b *b) SWAP(q1, q2 int) Builder       { return b.add2(gate.Swap(), q1, q2) }
func (b *b) Toffoli(a, bq, t int) Builder  { return b.add3(gate.Toffoli(), a, bq, t) }
func (b *b) Fredkin(c, t1, t2 int) Builder { return b.add3(gate.Fredkin(), c, t1, t2) }

func (b *b) Measure(q, cbit int) Builder {
	if b.checkState() {
		return b
	}
	if err := b.dagBuilder.AddMeasure(q, cbit); err != nil {
		return b.bail(err)
	}
	return b
}

// BuildDAG validates the internal DAG and returns it as a DAGReader.
// The builder becomes invalid after this call.
func (b *b) BuildDAG() (dag.DAGReader, error) {
	if b.built {
		return nil, fmt.Errorf("builder: BuildDAG or BuildCircuit already called: %w", dag.ErrBuild)
	}
	if b.err != nil {
		return nil, b.err
	}

	// Validate the DAG
	if err := b.dagBuilder.Validate(); err != nil {
		return nil, err
	}

	b.built = true // Mark as built

	// The concrete type (*dag.DAG) should implement DAGReader
	reader, ok := b.dagBuilder.(dag.DAGReader)
	if !ok {
		return nil, fmt.Errorf("builder: internal error - DAG does not implement DAGReader")
	}

	return reader, nil
}

// BuildCircuit is syntactic sugar for the common case where the caller
// immediately converts the DAG into the immutable, renderer‑friendly
// Circuit façade.
func (b *b) BuildCircuit() (circuit.Circuit, error) {
	dagReader, err := b.BuildDAG() // reuse existing validation logic
	if err != nil {
		return nil, err
	}
	return circuit.FromDAG(dagReader), nil
}

// ------------------------- private helpers ---------------------------

func (b *b) add1(g gate.Gate, q int) Builder {
	if b.checkState() {
		return b
	}
	if err := b.dagBuilder.AddGate(g, []int{q}); err != nil {
		return b.bail(err)
	}
	return b
}

func (b *b) add2(g gate.Gate, q0, q1 int) Builder {
	if b.checkState() {
		return b
	}
	if err := b.dagBuilder.AddGate(g, []int{q0, q1}); err != nil {
		return b.bail(err)
	}
	return b
}

func (b *b) add3(g gate.Gate, q0, q1, q2 int) Builder {
	if b.checkState() {
		return b
	}
	if err := b.dagBuilder.AddGate(g, []int{q0, q1, q2}); err != nil {
		return b.bail(err)
	}
	return b
}

// ------------------------- options -----------------------------------

type config struct {
	qubits int
	clbits int
}
type Option func(*config)

func Q(n int) Option { return func(c *config) { c.qubits = n } }
func C(n int) Option { return func(c *config) { c.clbits = n } }
