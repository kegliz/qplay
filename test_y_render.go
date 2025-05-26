package main

import (
	"fmt"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/renderer"
)

func main() {
	// Create a simple circuit with Y gate
	b := builder.New(builder.Q(3))
	b.H(0).Y(1).X(2).CNOT(0, 1).Y(2)

	// Build the circuit
	c, err := b.BuildCircuit()
	if err != nil {
		fmt.Printf("Error building circuit: %v\n", err)
		return
	}

	// Create renderer
	r := renderer.NewRenderer(80)

	// Render and save the circuit
	fmt.Println("Rendering circuit with Y gates...")
	err = r.Save("y-gate-demo.png", c)
	if err != nil {
		fmt.Printf("Error saving image: %v\n", err)
		return
	}

	fmt.Println("Circuit with Y gates rendered successfully to y-gate-demo.png")
}
