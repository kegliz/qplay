package qprog

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type QRuntimeTestSuite struct {
	suite.Suite
	R Runtime
}

func (s *QRuntimeTestSuite) SetupSuite() {
	f := NewRuntimeFactory()
	s.R = f.NewRuntime()
}

func (s *QRuntimeTestSuite) TestHadamard() {

	p := NewProgram(1)
	step := NewStep()
	err := step.AddGate(NewHGate(0))
	s.NoError(err)
	err = p.AddStep(step)
	s.NoError(err)
	result, err := s.R.Run(p)
	s.NoError(err)

	fmt.Println(result.Q.State())
	//result.Q.M(result.qc[0])
	fmt.Println(result.Q.State())
}

func (s *QRuntimeTestSuite) TestX() {

	p := NewProgram(1)
	step := NewStep()

	err := step.AddGate(NewXGate(0))
	s.NoError(err)
	err = p.AddStep(step)
	s.NoError(err)

	result, err := s.R.Run(p)
	s.NoError(err)

	fmt.Println("Result:", result.Q.State())
}

func (s *QRuntimeTestSuite) TestCNot() {

	p := NewProgram(2)
	step := NewStep()

	err := step.AddGate(NewCNotGate(0, 1))
	s.NoError(err)
	err = p.AddStep(step)
	s.NoError(err)

	result, err := s.R.Run(p)
	s.NoError(err)

	for _, r := range result.Q.State(0) {
		fmt.Println(r)
	}
	for _, r := range result.Q.State(1) {
		fmt.Println(r)
	}

}

func (s *QRuntimeTestSuite) TestRunTeleportation() {

	// 0: qubit to teleport
	// 1-2: entangled pair
	p := NewProgram(3)
	p.InitializeQubit(0, "Béla", 1+2i, 3+4i)
	// 1. create entangled pair
	step := NewStep()
	step.AddGate(NewHGate(1))
	p.AddStep(step)
	step = NewStep()
	step.AddGate(NewCNotGate(1, 2))
	p.AddStep(step)
	// 2. prepare qubit to teleport
	step = NewStep()
	step.AddGate(NewCNotGate(0, 1))
	p.AddStep(step)
	step = NewStep()
	step.AddGate(NewHGate(0))
	p.AddStep(step)
	// 3. measure qubit to teleport
	step = NewStep()
	step.AddGate(NewMeasurement(0))
	step.AddGate(NewMeasurement(1))
	p.AddStep(step)
	// 4. apply gates based on measurement results
	step = NewStep()

	/// ITT VAN BUG
	step.AddGate(NewCNotGate(1, 2))
	p.AddStep(step)
	step = NewStep()
	step.AddGate(NewCZGate(0, 2))
	p.AddStep(step)

	result, err := s.R.Run(p)
	s.NoError(err, "running program failed")

	fmt.Println()
	for _, r := range result.Q.State(result.QB.RbyName("Béla").Q()) {
		fmt.Println(r)
	}
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(QRuntimeTestSuite))
}
