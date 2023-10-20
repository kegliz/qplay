package qservice

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"kegnet.dev/qplay/internal/qprog"
)

// test programStore SaveProgram and GetProgram
func TestProgramStore(t *testing.T) {
	assert := assert.New(t)

	ps := NewProgramStore()

	// program with 1 qubit - no steps
	p1 := &qprog.Program{
		NumOfQubits: 1,
		Steps:       []qprog.Step{},
	}

	// program with 1 qubit - 1 step with 1 gate
	p2 := &qprog.Program{
		NumOfQubits: 1,
		Steps: []qprog.Step{
			{
				Gates: []qprog.Gate{
					{Type: qprog.HGate, Targets: []int{0}},
				},
			},
		},
	}
	// program with 2 qubit - no steps
	p3 := &qprog.Program{
		NumOfQubits: 2,
		Steps:       []qprog.Step{},
	}
	// program with 2 qubit - 1 step with 1 gate
	p4 := &qprog.Program{
		NumOfQubits: 2,
		Steps: []qprog.Step{
			{
				Gates: []qprog.Gate{
					{Type: qprog.HGate, Targets: []int{0}},
				},
			},
		},
	}
	//	// program with 2 qubit - 1 step with 2 gates
	p5 := &qprog.Program{
		NumOfQubits: 2,
		Steps: []qprog.Step{
			{
				Gates: []qprog.Gate{
					{Type: qprog.HGate, Targets: []int{0}},
					{Type: qprog.XGate, Targets: []int{1}},
				},
			},
		},
	}

	// test SaveProgram
	id1, err := ps.SaveProgram(p1)
	assert.NoError(err, "saving program failed")
	id2, err := ps.SaveProgram(p2)
	assert.NoError(err, "saving program failed")
	id3, err := ps.SaveProgram(p3)
	assert.NoError(err, "saving program failed")
	id4, err := ps.SaveProgram(p4)
	assert.NoError(err, "saving program failed")
	id5, err := ps.SaveProgram(p5)
	assert.NoError(err, "saving program failed")

	// test GetProgram
	p, err := ps.GetProgram(id1)
	assert.NoError(err, "getting program failed")
	assert.Equal(p1, p, "program mismatch")
	p, err = ps.GetProgram(id2)
	assert.NoError(err, "getting program failed")
	assert.Equal(p2, p, "program mismatch")
	p, err = ps.GetProgram(id3)
	assert.NoError(err, "getting program failed")
	assert.Equal(p3, p, "program mismatch")
	p, err = ps.GetProgram(id4)
	assert.NoError(err, "getting program failed")
	assert.Equal(p4, p, "program mismatch")
	p, err = ps.GetProgram(id5)
	assert.NoError(err, "getting program failed")
	assert.Equal(p5, p, "program mismatch")

	// test GetProgram with invalid id
	p, err = ps.GetProgram("invalid")
	assert.Error(err, "getting program with invalid id should fail")
	assert.Nil(p, "program should be nil")
}
