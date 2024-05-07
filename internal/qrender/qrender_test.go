package qrender

import (
	"testing"

	"github.com/kegliz/qplay/internal/qprog"
	"github.com/stretchr/testify/assert"
)

// TestRenderCircuit is a test for RenderCircuit.
func TestRenderCircuit(t *testing.T) {
	assert := assert.New(t)
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
	// program with 2 qubit - 1 step with 2 gates
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
	// Render the circuit
	qr := NewDefaultQRenderer()
	img := qr.RenderCircuit(p1)
	err := SaveImage(img, "circuit.png")
	assert.NoError(err, "saving image failed")

	img = qr.RenderCircuit(p2)
	err = SaveImage(img, "circuit2.png")
	assert.NoError(err, "saving image failed")

	// Check that the image has the expected dimensions
	assert.Equal(qr.imageWidth, img.Bounds().Dx(), "Rendered image has unexpected width")
	//	assert.Equal(qr.imageHeight, img.Bounds().Dy(), "Rendered image has unexpected height")

	img = qr.RenderCircuit(p3)
	err = SaveImage(img, "circuit3.png")
	assert.NoError(err, "saving image failed")

	img = qr.RenderCircuit(p4)
	err = SaveImage(img, "circuit4.png")
	assert.NoError(err, "saving image failed")

	img = qr.RenderCircuit(p5)
	err = SaveImage(img, "circuit5.png")
	assert.NoError(err, "saving image failed")

	// TODO: Add more tests to check that the image has the expected content
}

func TestRenderTeleportationCircuit(t *testing.T) {
	assert := assert.New(t)

	// 0: qubit to teleport
	// 1-2: entangled pair
	p := qprog.NewProgram(3)
	// 1. create entangled pair
	step := qprog.NewStep()
	step.AddGate(qprog.NewHGate(1))
	p.AddStep(step)
	step = qprog.NewStep()
	step.AddGate(qprog.NewCNotGate(1, 2))
	p.AddStep(step)
	// 2. prepare qubit to teleport
	step = qprog.NewStep()
	step.AddGate(qprog.NewCNotGate(0, 1))
	p.AddStep(step)
	step = qprog.NewStep()
	step.AddGate(qprog.NewHGate(0))
	p.AddStep(step)
	// 3. measure qubit to teleport
	step = qprog.NewStep()
	step.AddGate(qprog.NewMeasurement(0))
	step.AddGate(qprog.NewMeasurement(1))
	p.AddStep(step)
	// 4. apply gates based on measurement results
	step = qprog.NewStep()
	step.AddGate(qprog.NewCNotGate(1, 2))
	p.AddStep(step)
	step = qprog.NewStep()
	step.AddGate(qprog.NewCZGate(0, 2))
	p.AddStep(step)

	// Render the circuit
	qr := NewDefaultQRenderer()
	img := qr.RenderCircuit(p)
	err := SaveImage(img, "teleportation.png")
	assert.NoError(err, "saving image failed")

}
