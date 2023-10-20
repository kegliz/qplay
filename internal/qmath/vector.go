package qmath

import (
	"fmt"

	"github.com/itsubaki/q"
)

func ExampleNew() {
	qsim := q.New()

	// generate qubits of |0>|0>
	q0 := qsim.Zero()
	q1 := qsim.Zero()

	// apply quantum circuit
	qsim.H(q0).CNOT(q0, q1)

	for _, s := range qsim.State() {
		fmt.Println(s)
	}
	// [00][  0]( 0.7071 0.0000i): 0.5000
	// [11][  3]( 0.7071 0.0000i): 0.5000

	m0 := qsim.Measure(q0)
	m1 := qsim.Measure(q1)
	fmt.Println(m0.IsZero() == m1.IsZero()) // always true

	for _, s := range qsim.State() {
		fmt.Println(s)
	}
	// [00][  0]( 1.0000 0.0000i): 1.0000
	// or
	// [11][  3]( 1.0000 0.0000i): 1.0000
}
