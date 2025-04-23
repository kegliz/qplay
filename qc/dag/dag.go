package dag

import (
	"sync/atomic"

	"github.com/kegliz/qplay/qc/gate"
)

// NodeID is stable across passes/serialisation.
type NodeID uint64

var idCtr uint64 // atomic counter for NodeIDs

// Node holds one DAG vertex = Gate or Measure op.
type Node struct {
	ID     NodeID
	G      gate.Gate
	Qubits []int // logical qubit indices       (len = G.QubitSpan())
	Cbit   int   // classical target; -1 if none
	// Fast adjacency
	parents  []NodeID // Made accessible for FromDAG calculation
	children []NodeID
}

// Parents returns the parent node IDs.
func (n *Node) Parents() []NodeID {
	return n.parents
}

// DAG is *mutable* until Validate() is called; then considered frozen.
type DAG struct {
	qubits int
	clbits int

	nodes map[NodeID]*Node // all vertices
	byQ   [][]NodeID       // per-qubit chronological list
	last  []NodeID         // last op on each qubit (for hazards)

	valid bool // set by Validate()
}

func New(qb, cb int) *DAG {
	return &DAG{
		qubits: qb,
		clbits: cb,
		nodes:  make(map[NodeID]*Node),
		byQ:    make([][]NodeID, qb),
		last:   make([]NodeID, qb),
	}
}

func nextID() NodeID { return NodeID(atomic.AddUint64(&idCtr, 1)) }

// Add methods to access internal fields if they are private
func (d *DAG) Qubits() int             { return d.qubits }
func (d *DAG) Clbits() int             { return d.clbits }
func (d *DAG) Nodes() map[NodeID]*Node { return d.nodes } // Added for potential external use
