package qprog

import (
	"fmt"
	"testing"

	"github.com/itsubaki/q"
	"github.com/stretchr/testify/assert"
)

// TestBela is a low level q test for teleportation
func TestTeleportation(t *testing.T) {
	qsim := q.New()

	// generate qubits of |phi>|0>|0>
	phi := qsim.New(1+2i, 3+4i)
	// phi := qsim.One()
	// phi := qsim.Zero()
	fmt.Println("phi")
	for _, s := range qsim.State(phi) {
		fmt.Println(s)
	}
	// qx := qsim.New(11+22i, 33+44i)
	// fmt.Println("qx")

	// for _, s := range qsim.State(qx) {
	// 	fmt.Println(s)
	// }
	// fmt.Println()

	q0 := qsim.Zero()
	//q0 := qsim.New(11+22i, 33+44i)
	q1 := qsim.Zero()
	fmt.Println("phi")

	for _, s := range qsim.State(phi) {
		fmt.Println(s)
	}
	fmt.Println()
	// for _, s := range qsim.State(qx) {
	// 	fmt.Println(s)
	// }
	// fmt.Println()

	for _, s := range qsim.State(q0) {
		fmt.Println(s)
	}
	fmt.Println(qsim)
	for _, s := range qsim.State() {
		fmt.Println(s)
	}

	qsim.H(q0)

	fmt.Println(qsim)
	for _, s := range qsim.State() {
		fmt.Println(s)
	}

	qsim.CNOT(q0, q1)
	fmt.Println(qsim)
	for _, s := range qsim.State() {
		fmt.Println(s)
	}

	qsim.CNOT(phi, q0)
	qsim.H(phi)
	fmt.Println(qsim)
	for _, s := range qsim.State() {
		fmt.Println(s)
	}

	mphi := qsim.Measure(phi)
	mq0 := qsim.Measure(q0)

	fmt.Println(qsim)
	for _, s := range qsim.State() {
		fmt.Println(s)
	}

	qsim.CondX(mq0.IsOne(), q1)
	qsim.CondZ(mphi.IsOne(), q1)
	fmt.Println(qsim)
	for _, s := range qsim.State() {
		fmt.Println(s)
	}

	for _, s := range qsim.State(q1) {
		fmt.Println(s)
	}

}

// test AddStep error when qubit is out of range
func TestAddStepQubitOutOfRangeError(t *testing.T) {
	assert := assert.New(t)

	p := NewProgram(1)
	s := NewStep()
	err := s.AddGate(NewXGate(2))
	assert.NoError(err)

	err = p.AddStep(s)
	assert.Error(err, "qubit is out of range while adding step")
}

// test AddStep error when step is empty
func TestAddStepEmptyError(t *testing.T) {
	assert := assert.New(t)

	p := NewProgram(1)
	s := NewStep()

	err := p.AddStep(s)
	assert.Error(err, "step is empty while adding step")
}

// test AddGate error when target or controll is duplicated
func TestAddGateQubitDuplicatedError(t *testing.T) {
	assert := assert.New(t)
	s := NewStep()

	err := s.AddGate(NewXGate(1))
	assert.NoError(err)
	err = s.AddGate(NewXGate(1))
	assert.Error(err, "target is duplicated while adding gate")
}

func TestCheck(t *testing.T) {
	assert := assert.New(t)

	p := NewProgram(1)
	s := NewStep()

	err := s.AddGate(NewXGate(0))
	assert.NoError(err)
	err = p.AddStep(s)
	assert.NoError(err)

	err = p.Check()
	assert.NoError(err)
}

// test Check error when target is out of range
func TestCheckTargetOutOfRangeError(t *testing.T) {
	assert := assert.New(t)

	p := &Program{
		NumOfQubits: 1,
		Steps: []Step{
			{
				Gates: []Gate{
					{Type: HGate, Targets: []int{1}},
				},
			},
		},
	}

	err := p.Check()
	assert.Error(err, "target is out of range while checking program")
}

// test Check error when target is duplication
func TestCheckTargetDuplicationError(t *testing.T) {
	assert := assert.New(t)

	p := &Program{
		NumOfQubits: 1,
		Steps: []Step{
			{
				Gates: []Gate{
					{Type: HGate, Targets: []int{0}},
					{Type: HGate, Targets: []int{0}},
				},
			},
		},
	}

	err := p.Check()
	assert.Error(err, "target is duplicated while checking program")
}

// test Check error when target is duplication
func TestCheckTargetDuplicationWithDifferentSteps(t *testing.T) {
	assert := assert.New(t)

	p := &Program{
		NumOfQubits: 1,
		Steps: []Step{
			{
				Gates: []Gate{
					{Type: HGate, Targets: []int{0}},
				},
			},
			{
				Gates: []Gate{
					{Type: HGate, Targets: []int{0}},
				},
			},
		},
	}

	err := p.Check()
	assert.NoError(err)
}

// TODO: check cnot, tofoli gates with programs: teleportation and quantum arithmetics
