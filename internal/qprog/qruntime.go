package qprog

import (
	"fmt"

	"github.com/itsubaki/q"
)

type (

	// RuntimeFactory is a factory of quantum computer runtime.
	RuntimeFactory interface {
		// NewRuntime creates a new quantum computer runtime.
		NewRuntime() Runtime
	}

	// Result is a result of quantum circuit.
	Result struct {
		Q  *q.Q
		QB QuantumBoard
	}

	// Runtime is a quantum computer runtime (simulator or other)
	Runtime interface {
		// Run runs the program.
		Run(p *Program) (*Result, error)
		// Result returns the result of the program.
		// Result() *Result
	}

	// Register is a quantum register.
	Register interface {
		IsMeasured() bool
		Id() int
		Name() string
		IsOne() bool
		IsZero() bool
		Q() q.Qubit
	}

	QuantumBoard interface {
		R(int) Register
		RbyName(string) Register
	}
)

type (
	// qRuntimeFactory is a factory of quantum computer simulation runtime.
	qRuntimeFactory struct{}

	// qRuntime is a quantum computer simulation runtime.
	qRuntime struct{}

	// qRegister is a quantum register.
	qRegister struct {
		name       string
		q          q.Qubit
		isMeasured bool
		classical  bool
	}
	qBoard struct {
		regs []*qRegister
	}
)

var _ RuntimeFactory = (*qRuntimeFactory)(nil)
var _ Runtime = (*qRuntime)(nil)
var _ Register = (*qRegister)(nil)
var _ QuantumBoard = (*qBoard)(nil)

func (qr *qBoard) R(i int) Register {
	if len(qr.regs) <= i {
		return nil
	}
	return qr.regs[i]
}

func (qr *qBoard) RbyName(name string) Register {
	for _, r := range qr.regs {
		if r.Name() == name {
			return r
		}
	}
	return nil
}

func (qr *qRegister) IsMeasured() bool {
	return qr.isMeasured
}

func (qr *qRegister) Id() int {
	return qr.q.Index()
}

func (qr *qRegister) Name() string {
	if qr.name != "" {
		return qr.name
	}
	return fmt.Sprintf("q%d", qr.q.Index())
}

func (qr *qRegister) IsOne() bool {
	return qr.classical
}

func (qr *qRegister) IsZero() bool {
	return !qr.classical
}

func (qr *qRegister) Q() q.Qubit {
	return qr.q
}

// NewRuntimeFactory creates a new quantum computer runtime factory.
func NewRuntimeFactory() RuntimeFactory {
	return &qRuntimeFactory{}
}

// NewRuntime creates a new quantum computer runtime.
func (f *qRuntimeFactory) NewRuntime() Runtime {
	return &qRuntime{}
}

// Run executes quantum circuit.
// It returns the result of quantum circuit.
func (qrun *qRuntime) Run(p *Program) (*Result, error) {

	regs := make([]*qRegister, p.NumOfQubits)

	qsim := q.New()
	var found bool
	for i := range regs {
		found = false // if i in p.initialAlfaBeta initiiate with alfa beta
		for _, init := range p.initializations {
			if init.index == i {
				regs[i] = &qRegister{
					name: init.name,
					q:    qsim.New(init.alfa, init.beta),
				}
				found = true
				break
			}
		}
		// otherwise start with |0> state
		if !found {
			regs[i] = &qRegister{
				q: qsim.Zero(),
			}
		}
	}
	// apply quantum circuit
	for i, step := range p.Steps {
		fmt.Printf("step %d input", i)
		fmt.Println()
		fmt.Println(qsim)
		for _, s := range qsim.State() {
			fmt.Println(s)
		}

		for _, gate := range step.Gates {
			switch gate.Type {
			case HGate:
				qsim.H(regs[gate.Targets[0]].q)
			case XGate:
				qsim.X(regs[gate.Targets[0]].q)
			case ZGate:
				qsim.Z(regs[gate.Targets[0]].q)
			case CNotGate:
				if regs[gate.Controls[0]].isMeasured {
					qsim.CondX(regs[gate.Controls[0]].IsOne(), regs[gate.Targets[0]].q)
				} else {
					qsim.CNOT(regs[gate.Controls[0]].q, regs[gate.Targets[0]].q)
				}
			case ToffoliGate:
				qsim.Toffoli(regs[gate.Controls[0]].q, regs[gate.Controls[1]].q, regs[gate.Targets[0]].q)
			case CZGate:
				if regs[gate.Controls[0]].isMeasured {
					qsim.CondZ(regs[gate.Controls[0]].IsOne(), regs[gate.Targets[0]].q)
				} else {
					qsim.CZ(regs[gate.Controls[0]].q, regs[gate.Targets[0]].q)
				}
			case Measurement:
				m := qsim.Measure(regs[gate.Targets[0]].q)
				regs[gate.Targets[0]].classical = m.IsOne()
				regs[gate.Targets[0]].isMeasured = true
			}
		}
	}

	return &Result{
		Q: qsim,
		QB: &qBoard{
			regs: regs,
		},
	}, nil
}
