package circuit

import (
	"sort"

	"github.com/kegliz/qplay/qc/dag"
	"github.com/kegliz/qplay/qc/gate"
)

type Operation struct {
	G        gate.Gate
	Qubits   []int // Absolute qubit indices
	Cbit     int   // Absolute classical bit index (-1 if none)
	TimeStep int   // Calculated layout column (starting at 0)
	Line     int   // Calculated layout primary line (usually min qubit index)
}

type Circuit interface {
	Qubits() int
	Clbits() int
	Operations() []Operation // topological order with layout info
	Depth() int              // Max TimeStep + 1
	MaxStep() int            // Max TimeStep
}

type circuit struct {
	qubits  int
	clbits  int
	ops     []Operation // Cached operations with layout info
	depth   int         // Number of layers (MaxStep + 1)
	maxStep int         // Max timestep index
}

// FromDAG creates an immutable Circuit from a validated DAGReader.
// It calculates the layout (TimeStep, Line) for each operation.
func FromDAG(dr dag.DAGReader) Circuit {
	// Get topologically sorted nodes
	nodes := dr.Operations()
	if nodes == nil || len(nodes) == 0 {
		// Create an empty circuit
		return &circuit{
			qubits:  dr.Qubits(),
			clbits:  dr.Clbits(),
			ops:     []Operation{},
			depth:   0,
			maxStep: -1,
		}
	}

	ops := make([]Operation, len(nodes))
	// Store calculated timestep for each node ID
	nodeTimeStep := make(map[dag.NodeID]int)

	maxStep := -1
	for i, n := range nodes {
		// Calculate TimeStep based on parents' timesteps
		currentMaxParentStep := -1
		for _, pID := range n.Parents() {
			if pStep, ok := nodeTimeStep[pID]; ok && pStep > currentMaxParentStep {
				currentMaxParentStep = pStep
			}
		}

		// Node's timestep is 1 greater than its latest-finishing parent
		step := currentMaxParentStep + 1
		nodeTimeStep[n.ID] = step

		if step > maxStep {
			maxStep = step
		}

		// Calculate Line (minimum qubit index)
		minQubit := -1
		if len(n.Qubits) > 0 {
			minQubit = n.Qubits[0]
			for _, q := range n.Qubits[1:] {
				if q < minQubit {
					minQubit = q
				}
			}
		}

		ops[i] = Operation{
			G:        n.G,
			Qubits:   append([]int(nil), n.Qubits...), // Copy slice
			Cbit:     n.Cbit,
			TimeStep: step,
			Line:     minQubit,
		}
	}

	// Sort operations by TimeStep, then by Line for consistent rendering
	sort.SliceStable(ops, func(i, j int) bool {
		if ops[i].TimeStep != ops[j].TimeStep {
			return ops[i].TimeStep < ops[j].TimeStep
		}
		return ops[i].Line < ops[j].Line
	})

	return &circuit{
		qubits:  dr.Qubits(),
		clbits:  dr.Clbits(),
		ops:     ops,
		depth:   maxStep + 1,
		maxStep: maxStep,
	}
}

// ---------------- interface methods --------------------
func (c *circuit) Qubits() int { return c.qubits }
func (c *circuit) Clbits() int { return c.clbits }

func (c *circuit) Depth() int {
	return c.depth
}

func (c *circuit) MaxStep() int {
	return c.maxStep
}

func (c *circuit) Operations() []Operation {
	// Return a copy to prevent external modification
	result := make([]Operation, len(c.ops))
	copy(result, c.ops)
	return result
}
