package qprog

import (
	"fmt"
)

type (
	Program struct {
		ID              string `json:"id"`
		NumOfQubits     int    `json:"numofqubits"`
		Steps           []Step `json:"steps"`
		initializations []initQubit
	}

	initQubit struct {
		index int
		name  string
		alfa  complex128
		beta  complex128
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
)

func NewProgram(numOfQubits int) *Program {
	return &Program{
		NumOfQubits:     numOfQubits,
		Steps:           []Step{},
		initializations: []initQubit{},
	}
}

// InitializeQubitWithAlfa initializes a qubit with alfa.
func (p *Program) InitializeQubit(i int, name string, alfa complex128, beta complex128) error {
	if i >= p.NumOfQubits {
		return fmt.Errorf("qubit is out of range while initializing qubit")
	}
	p.initializations = append(p.initializations, initQubit{
		name:  name,
		index: i,
		alfa:  alfa,
		beta:  beta,
	})

	return nil
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
// TODO: check if we wrongly use a qubit after it is measured
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

// AddGate adds a gate to step.
func (step *Step) AddGate(gate *Gate) error {
	// iterate through step.gates and check that gate.targets and gate.controls are not duplicated with the current gates
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

// Check the validity of the program
// TODO: check the valid usage of measured qubits
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
