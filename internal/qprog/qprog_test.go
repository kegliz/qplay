package qprog

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHadamardGate(t *testing.T) {
	assert := assert.New(t)

	p := NewProgram(1)
	s := NewStep()
	err := s.AddGate(NewHGate(0))
	assert.NoError(err)
	err = p.AddStep(s)
	assert.NoError(err)
	result := p.Run()

	fmt.Println(result.q.State())
	result.q.M(result.qc[0])
	fmt.Println(result.q.State())
}

func TestX(t *testing.T) {
	assert := assert.New(t)

	p := NewProgram(1)
	s := NewStep()

	err := s.AddGate(NewXGate(0))
	assert.NoError(err)
	err = p.AddStep(s)
	assert.NoError(err)

	result := p.Run()

	fmt.Println(result.q.State())
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
