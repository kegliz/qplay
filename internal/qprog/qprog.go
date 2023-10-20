package qprog

import (
	"fmt"

	"github.com/itsubaki/q"
)

type (
	Program struct {
		ID          string `json:"id"`
		NumOfQubits int    `json:"numofqubits"`
		Steps       []Step `json:"steps"`
	}

	Step struct {
		Gates []Gate `json:"gates"`
	}

	// Gate is a quantum gate.
	// targets and controls are distinct qubit indices.
	Gate struct {
		Type     gateType `json:"name"`
		Targets  []int    `json:"targets"`
		Controls []int    `json:"controls"`
	}

	Result struct {
		q  *q.Q
		qc []q.Qubit
	}
)

func NewProgram(numOfQubits int) *Program {
	return &Program{
		NumOfQubits: numOfQubits,
		Steps:       []Step{},
	}
}

func NewProgramWithID(numOfQubits int, id string) *Program {
	return &Program{
		ID:          id,
		NumOfQubits: numOfQubits,
		Steps:       []Step{},
	}
}

func NewStep() *Step {
	return &Step{
		Gates: []Gate{},
	}
}

// AddStep adds a step to program.
func (p *Program) AddStep(step *Step) error {
	if len(step.Gates) == 0 {
		return fmt.Errorf("step is empty while adding step")
	}
	if step.maxIndex() >= p.NumOfQubits {
		return fmt.Errorf("qubit is out of range while adding step")
	}
	p.Steps = append(p.Steps, *step)
	return nil
}

// maxIndex returns the maximum index of target and control qubits.
func (s *Step) maxIndex() int {
	max := -1
	for _, gate := range s.Gates {
		for _, target := range gate.Targets {
			if target > max {
				max = target
			}
		}
		for _, control := range gate.Controls {
			if control > max {
				max = control
			}
		}
	}
	return max
}

// NewXGate returns a new XGate.
func NewXGate(target int) *Gate {
	return &Gate{
		Type:    XGate,
		Targets: []int{target},
	}
}

// NewHGate returns a new HGate.
func NewHGate(target int) *Gate {
	return &Gate{
		Type:    HGate,
		Targets: []int{target},
	}
}

// AddGate adds a gate to step.
func (step *Step) AddGate(gate *Gate) error {
	// iterate through step.gates and check that gate.targets and gate.controls are not duplicated with the current gate
	for _, g := range step.Gates {
		for _, t := range gate.Targets {
			for _, tt := range g.Targets {
				if t == tt {
					return fmt.Errorf("target qubit %d in gate is already used at step", t)
				}
			}
			for _, cc := range g.Controls {
				if t == cc {
					return fmt.Errorf("target qubit %d in gate is already used at step", t)
				}
			}
		}
		for _, c := range gate.Controls {
			for _, cc := range g.Controls {
				if c == cc {
					return fmt.Errorf("control qubit %d in gate is already used at step", c)
				}
			}
			for _, tt := range g.Targets {
				if c == tt {
					return fmt.Errorf("control qubit %d in gate is already used at step", c)
				}
			}
		}
	}
	step.Gates = append(step.Gates, *gate)
	return nil
}

func (p *Program) Check() error {
	for _, step := range p.Steps {
		err := step.Check(p.NumOfQubits)
		if err != nil {
			return err
		}
	}
	return nil
}

// Check checks if the step is valid.
func (s *Step) Check(maxQubit int) error {
	if len(s.Gates) == 0 {
		return fmt.Errorf("step has no gates")
	}
	// check if the target and control qubits are not out of range
	if max := s.maxIndex(); max >= maxQubit {
		return fmt.Errorf("qubit is out of range: %d", max)
	}
	// check if the union of all the target and control qubits of all the gates does not contain duplicates
	// make int slice for the union of all the target and control qubits
	qubits := make([]int, 0)
	for i, gate := range s.Gates {
		for _, target := range gate.Targets {
			// add the target qubit to the union if it is not in the union
			if !contains(qubits, target) {
				qubits = append(qubits, target)
			} else {
				return fmt.Errorf("target qubit %d in gate %d is duplicated", target, i)
			}
		}
		// add the control qubit to the union if it is not in the union
		for _, control := range gate.Controls {
			if !contains(qubits, control) {
				qubits = append(qubits, control)
			} else {
				return fmt.Errorf("control qubit %d in gate %d is duplicated", control, i)
			}
		}
	}
	return nil
}

// contains checks if a slice of integers contains a given integer.
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// Run executes quantum circuit.
// It returns the result of quantum circuit.
// TODO: separate the simulation from Program.
func (p *Program) Run() *Result {
	qsim := q.New()
	qc := make([]q.Qubit, p.NumOfQubits)
	for i := range qc {
		qc[i] = qsim.Zero()
	}
	// apply quantum circuit
	for _, step := range p.Steps {
		for _, gate := range step.Gates {
			switch gate.Type {
			case HGate:
				qsim.H(qc[gate.Targets[0]])
			case XGate:
				qsim.X(qc[gate.Targets[0]])
			}
		}
	}
	return &Result{
		q:  qsim,
		qc: qc,
	}
}
