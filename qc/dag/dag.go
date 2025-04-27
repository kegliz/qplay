package dag

import (
	"fmt"
	"sync/atomic"

	"github.com/kegliz/qplay/qc/gate"
)

// NodeID is stable across passes/serialisation.
type NodeID uint64

var idCtr uint64 // atomic counter for NodeIDs

// Node holds one DAG vertex = Gate or Measure op.
// It contains the gate, its qubit targets, and its classical target.
type Node struct {
	ID     NodeID
	G      gate.Gate
	Qubits []int // logical qubit indices       (len = G.QubitSpan())
	Cbit   int   // classical target; -1 if none
	// Fast adjacency
	parents  []NodeID
	children []NodeID
}

// Parents returns a copy of the parent node IDs.
func (n *Node) Parents() []NodeID {
	result := make([]NodeID, len(n.parents))
	copy(result, n.parents)
	return result
}

// DAGBuilder defines the interface for constructing a DAG.
type DAGBuilder interface {
	AddGate(g gate.Gate, qs []int) error
	AddMeasure(q, c int) error
	Validate() error
	Qubits() int
	Clbits() int
}

// DAGReader defines the interface for reading a validated DAG.
type DAGReader interface {
	Operations() []*Node // Returns nodes in topological order
	Depth() int          // Returns the circuit depth
	Qubits() int
	Clbits() int
}

// DAG is *mutable* until Validate() is called; then considered frozen.
// It implements both DAGBuilder and DAGReader interfaces.
type DAG struct {
	qubits int
	clbits int

	nodes map[NodeID]*Node // all vertices
	byQ   [][]NodeID       // per-qubit chronological list
	last  []NodeID         // last op on each qubit (for hazards)

	valid bool // set by Validate()

	// Cached results after validation
	topoOrder []*Node
	depth     int
}

// New creates a new DAG with the specified number of qubits and classical bits.
func New(qb, cb int) *DAG {
	return &DAG{
		qubits: qb,
		clbits: cb,
		nodes:  make(map[NodeID]*Node),
		byQ:    make([][]NodeID, qb),
		last:   make([]NodeID, qb),
		depth:  -1, // Initialize depth as uncalculated
	}
}

// nextID generates a new unique NodeID.
func nextID() NodeID { return NodeID(atomic.AddUint64(&idCtr, 1)) }

// Qubits returns the number of qubits.
func (d *DAG) Qubits() int { return d.qubits }

// Clbits returns the number of classical bits.
func (d *DAG) Clbits() int { return d.clbits }

// AddGate adds a gate operation to the DAG.
func (d *DAG) AddGate(g gate.Gate, qs []int) error {
	if d.valid {
		return ErrValidated
	}
	if err := d.checkGate(g, qs); err != nil {
		return err
	}
	n := &Node{
		ID:     nextID(),
		G:      g,
		Qubits: append([]int(nil), qs...),
		Cbit:   -1,
	}
	d.nodes[n.ID] = n

	// Build edges: parent = last op on each incident qubit.
	// Use a set to prevent duplicate parents if a gate touches the same qubit twice
	parentSet := make(map[NodeID]struct{})
	for _, q := range qs {
		if prev := d.last[q]; prev != 0 {
			if _, exists := parentSet[prev]; !exists {
				parentSet[prev] = struct{}{}
				n.parents = append(n.parents, prev)
				d.nodes[prev].children = append(d.nodes[prev].children, n.ID)
			}
		}
		d.last[q] = n.ID
		d.byQ[q] = append(d.byQ[q], n.ID)
	}
	return nil
}

// AddMeasure adds a measurement operation to the DAG.
func (d *DAG) AddMeasure(q, c int) error {
	if d.valid {
		return ErrValidated
	}
	if q < 0 || q >= d.qubits {
		return ErrBadQubit
	}
	if c < 0 || c >= d.clbits {
		return ErrBadClbit
	}
	n := &Node{
		ID:     nextID(),
		G:      gate.Measure(),
		Qubits: []int{q},
		Cbit:   c,
	}
	d.nodes[n.ID] = n
	if prev := d.last[q]; prev != 0 {
		n.parents = []NodeID{prev}
		d.nodes[prev].children = append(d.nodes[prev].children, n.ID)
	}
	d.last[q] = n.ID
	d.byQ[q] = append(d.byQ[q], n.ID)
	return nil
}

// Validate checks if the DAG is acyclic, calculates topological order and depth,
// and marks it as valid (frozen).
// Once validated, no further operations can be added.
// This is a no-op if already validated.
func (d *DAG) Validate() error {
	if d.valid {
		return nil
	}

	// Check for cycles
	if err := d.acyclic(); err != nil {
		return err
	}

	// Calculate topological order and depth
	d.topoOrder = d.calculateTopoSort()
	d.depth = d.calculateDepth()

	d.valid = true
	return nil
}

// Operations returns nodes in topological order. Requires Validate() to be called first.
// It returns a copy of the slice to prevent external modification.
// If Validate() was not called, it returns nil.
func (d *DAG) Operations() []*Node {
	if !d.valid {
		return nil
	}
	// Return a copy to prevent external modification
	result := make([]*Node, len(d.topoOrder))
	copy(result, d.topoOrder)
	return result
}

// Depth returns the calculated depth. Requires Validate() to be called first.
func (d *DAG) Depth() int {
	return d.depth
}

// checkGate validates gate qubit span and indices.
func (d *DAG) checkGate(g gate.Gate, qs []int) error {
	if len(qs) != g.QubitSpan() {
		return ErrSpan
	}

	// Check for duplicate qubits within the same gate application
	seen := make(map[int]bool)
	for _, q := range qs {
		if q < 0 || q >= d.qubits {
			return ErrBadQubit
		}
		if seen[q] {
			return fmt.Errorf("dag: duplicate qubit %d specified for gate %s", q, g.Name())
		}
		seen[q] = true
	}
	return nil
}

// calculateTopoSort performs Kahn's algorithm for topological sorting.
func (d *DAG) calculateTopoSort() []*Node {
	inDeg := make(map[NodeID]int, len(d.nodes))
	for id, node := range d.nodes {
		inDeg[id] = len(node.parents)
	}

	// Initialize queue with nodes that have no dependencies
	queue := make([]NodeID, 0, len(d.nodes))
	for id, deg := range inDeg {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	order := make([]*Node, 0, len(d.nodes))
	for len(queue) > 0 {
		// Pop from queue
		id := queue[0]
		queue = queue[1:]

		// Add to result
		node := d.nodes[id]
		order = append(order, node)

		// Update dependencies
		for _, childID := range node.children {
			inDeg[childID]--
			if inDeg[childID] == 0 {
				queue = append(queue, childID)
			}
		}
	}

	// If we didn't visit all nodes, there's a cycle (should be caught by acyclic())
	if len(order) != len(d.nodes) {
		// This is a safety check - acyclic() should have caught any cycles
		panic("internal error: topological sort couldn't process all nodes; cycle not caught by acyclic()")
	}

	return order
}

// calculateDepth calculates the circuit depth (number of layers).
func (d *DAG) calculateDepth() int {
	if len(d.topoOrder) == 0 {
		return 0 // Empty DAG has depth 0
	}

	// Calculate node depths
	nodeDepth := make(map[NodeID]int)
	maxDepth := 0

	for _, node := range d.topoOrder {
		// Node's depth is 1 + max depth of its parents
		depth := 0
		for _, parentID := range node.parents {
			if parentDepth, ok := nodeDepth[parentID]; ok && parentDepth > depth {
				depth = parentDepth
			}
		}
		depth++ // Add 1 for this node's layer

		nodeDepth[node.ID] = depth
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

// acyclic performs DFS cycle-check.
func (d *DAG) acyclic() error {
	// 0: unvisited, 1: visiting (recursion stack), 2: visited
	state := make(map[NodeID]int)

	var dfs func(NodeID) error
	dfs = func(id NodeID) error {
		switch state[id] {
		case 1:
			return fmt.Errorf("dag: cycle detected involving node %d (%s)",
				id, d.nodes[id].G.Name())
		case 2:
			return nil // Already visited
		}

		// Mark as visiting
		state[id] = 1

		// Visit children
		for _, childID := range d.nodes[id].children {
			if err := dfs(childID); err != nil {
				return err
			}
		}

		// Mark as visited
		state[id] = 2
		return nil
	}

	// Try from each node (to handle disconnected subgraphs)
	for id := range d.nodes {
		if state[id] == 0 { // Not yet visited
			if err := dfs(id); err != nil {
				return err
			}
		}
	}

	return nil // No cycles found
}
