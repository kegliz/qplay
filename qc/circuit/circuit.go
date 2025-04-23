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
	TimeStep int   // Calculated layout column
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
	d   *dag.DAG
	ops []Operation // Cached operations with layout info
}

// ---------------- exported constructor -----------------
func FromDAG(d *dag.DAG) Circuit {
	nodes := d.Operations() // Nodes in topological order
	ops := make([]Operation, len(nodes))
	depth := make(map[dag.NodeID]int) // Store depth (timestep) for each node

	maxStep := 0
	for i, n := range nodes {
		// Calculate TimeStep (depth)
		nodeDepth := 0
		for _, pID := range n.Parents() { // Assuming Parents() method exists or accessing parents field
			if pDepth, ok := depth[pID]; ok {
				if pDepth+1 > nodeDepth {
					nodeDepth = pDepth + 1
				}
			}
		}
		depth[n.ID] = nodeDepth
		if nodeDepth > maxStep {
			maxStep = nodeDepth
		}

		// Calculate Line (minimum qubit index)
		minQubit := -1
		if len(n.Qubits) > 0 {
			minQubit = n.Qubits[0] // Assume sorted or find min
			// Ensure minQubit is actually the minimum
			for _, q := range n.Qubits {
				if q < minQubit {
					minQubit = q
				}
			}
		}

		ops[i] = Operation{
			G:        n.G,
			Qubits:   append([]int(nil), n.Qubits...), // Copy slice
			Cbit:     n.Cbit,
			TimeStep: nodeDepth,
			Line:     minQubit,
		}
	}

	// Sort operations primarily by TimeStep, secondarily by Line for consistent rendering
	sort.SliceStable(ops, func(i, j int) bool {
		if ops[i].TimeStep != ops[j].TimeStep {
			return ops[i].TimeStep < ops[j].TimeStep
		}
		return ops[i].Line < ops[j].Line
	})

	return &circuit{d: d, ops: ops}
}

// ---------------- interface methods --------------------
func (c *circuit) Qubits() int { return c.d.Qubits() }
func (c *circuit) Clbits() int { return c.d.Clbits() }

// Depth returns the number of layers/timesteps in the circuit.
func (c *circuit) Depth() int {
	return c.MaxStep() + 1
}

// MaxStep returns the maximum timestep index used in the circuit layout.
func (c *circuit) MaxStep() int {
	max := 0
	for _, o := range c.ops {
		if o.TimeStep > max {
			max = o.TimeStep
		}
	}
	return max
}

func (c *circuit) Operations() []Operation {
	// Return the cached & sorted operations
	return c.ops
}

// Note: The Parents() method is expected to be defined on dag.Node within the 'dag' package.
// The FromDAG function already relies on its existence.
