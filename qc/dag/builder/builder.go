package builder

import (
	"github.com/kegliz/qplay/qc/dag"
	"github.com/kegliz/qplay/qc/gate"
)

// ---------------------------- public API -----------------------------

// Builder implements a *fluent* declarative DSL:
//
//	c, _ := builder.New(Q(3), C(2)).
//	    H(0).
//	    CNOT(0, 1).
//	    Toffoli(0, 1, 2).
//	    Measure(2, 0).
//	    Build()
type Builder interface {
	// Single-qubit gates
	H(q int) Builder
	X(q int) Builder
	S(q int) Builder

	// Multi-qubit gates
	CNOT(ctrl, tgt int) Builder
	SWAP(q1, q2 int) Builder
	Toffoli(c1, c2, tgt int) Builder
	Fredkin(ctrl, t1, t2 int) Builder

	// Measurement
	Measure(q, c int) Builder

	// Finalise
	Build() (*dag.DAG, error) // Changed return type
}

// New returns a fresh Builder with the requested qubits/classical bits.
func New(opts ...Option) Builder { return newBuilder(opts...) }

// ---------------------------- implementation -------------------------

type b struct {
	d   *dag.DAG // mutable during build
	err error
}

func newBuilder(opts ...Option) *b {
	cfg := config{qubits: 1}
	for _, o := range opts {
		o(&cfg)
	}
	return &b{d: dag.New(cfg.qubits, cfg.clbits)}
}

// helper: bail-out pattern
func (b *b) bail(err error) Builder { b.err = err; return b }

func (b *b) H(q int) Builder               { return b.add1(gate.H(), q) }
func (b *b) X(q int) Builder               { return b.add1(gate.X(), q) }
func (b *b) S(q int) Builder               { return b.add1(gate.S(), q) }
func (b *b) CNOT(c, t int) Builder         { return b.add2(gate.CNOT(), c, t) }
func (b *b) SWAP(q1, q2 int) Builder       { return b.add2(gate.Swap(), q1, q2) }
func (b *b) Toffoli(a, bq, t int) Builder  { return b.add3(gate.Toffoli(), a, bq, t) }
func (b *b) Fredkin(c, t1, t2 int) Builder { return b.add3(gate.Fredkin(), c, t1, t2) }
func (b *b) Measure(q, cbit int) Builder {
	if b.err != nil {
		return b
	}
	if err := b.d.AddMeasure(q, cbit); err != nil {
		return b.bail(err)
	}
	return b
}

func (b *b) Build() (*dag.DAG, error) { // Changed return type
	if b.err != nil {
		return nil, b.err
	}
	if err := b.d.Validate(); err != nil {
		return nil, err
	}
	return b.d, nil // Return the validated DAG
}

// ------------------------- private helpers ---------------------------

func (b *b) add1(g gate.Gate, q int) Builder {
	if b.err != nil {
		return b
	}
	if err := b.d.AddGate(g, []int{q}); err != nil {
		return b.bail(err)
	}
	return b
}
func (b *b) add2(g gate.Gate, q0, q1 int) Builder {
	if b.err != nil {
		return b
	}
	if err := b.d.AddGate(g, []int{q0, q1}); err != nil {
		return b.bail(err)
	}
	return b
}
func (b *b) add3(g gate.Gate, q0, q1, q2 int) Builder {
	if b.err != nil {
		return b
	}
	if err := b.d.AddGate(g, []int{q0, q1, q2}); err != nil {
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
