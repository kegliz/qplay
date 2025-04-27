package dag

import (
	"testing"

	"github.com/kegliz/qplay/qc/gate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInterfaces ensures the DAG type implements the interfaces
func TestInterfaces(t *testing.T) {
	// Compile-time checks
	var _ DAGBuilder = (*DAG)(nil)
	var _ DAGReader = (*DAG)(nil)
}

func TestDAG_New(t *testing.T) {
	assert := assert.New(t)
	d := New(5, 2)
	assert.NotNil(d)
	assert.Equal(5, d.Qubits())
	assert.Equal(2, d.Clbits())
	assert.NotNil(d.nodes)
	assert.Len(d.nodes, 0) // Nodes map should be empty initially
	assert.Len(d.byQ, 5)
	assert.Len(d.last, 5)
	// Check initial state of byQ slices
	for i := 0; i < 5; i++ {
		assert.Len(d.byQ[i], 0)
	}
	// Check initial state of last slice (should be all zeros)
	for i := 0; i < 5; i++ {
		assert.Equal(NodeID(0), d.last[i]) // Initial value is zero NodeID
	}
	assert.False(d.valid)
}

func TestDAG_AddGate(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	d := New(3, 0) // Use direct integer arguments

	// Add H(0)
	err := d.AddGate(gate.H(), []int{0})
	require.NoError(err)
	assert.Len(d.nodes, 1)
	var h0Node *Node
	for _, n := range d.nodes { // Get the node (only one)
		h0Node = n
	}
	require.NotNil(h0Node)
	assert.Equal(gate.H(), h0Node.G)
	assert.Equal([]int{0}, h0Node.Qubits)
	assert.Equal(-1, h0Node.Cbit)
	assert.Empty(h0Node.parents)
	assert.Empty(h0Node.children)
	assert.Equal(h0Node.ID, d.last[0])
	assert.Equal([]NodeID{h0Node.ID}, d.byQ[0])

	// Add CNOT(0, 1)
	err = d.AddGate(gate.CNOT(), []int{0, 1})
	require.NoError(err)
	assert.Len(d.nodes, 2)
	var cnotNode *Node
	for id, n := range d.nodes {
		if id != h0Node.ID {
			cnotNode = n
			break
		}
	}
	require.NotNil(cnotNode)
	assert.Equal(gate.CNOT(), cnotNode.G)
	assert.Equal([]int{0, 1}, cnotNode.Qubits)
	// CNOT depends on the last op on qubit 0 (H(0)) and potentially qubit 1 (none initially)
	// Since last[1] was 0, only H(0) should be a parent.
	require.Len(cnotNode.parents, 1)
	assert.Contains(cnotNode.parents, h0Node.ID)
	assert.Empty(cnotNode.children)
	assert.Equal(cnotNode.ID, d.last[0]) // CNOT is now last on qubit 0
	assert.Equal(cnotNode.ID, d.last[1]) // CNOT is now last on qubit 1
	assert.Equal([]NodeID{h0Node.ID, cnotNode.ID}, d.byQ[0])
	assert.Equal([]NodeID{cnotNode.ID}, d.byQ[1]) // Only CNOT added to qubit 1

	// Check H(0) children updated
	assert.Equal([]NodeID{cnotNode.ID}, h0Node.children)

	// Test errors
	err = d.AddGate(gate.H(), []int{3}) // Qubit out of range
	assert.ErrorIs(err, ErrBadQubit)
	err = d.AddGate(gate.CNOT(), []int{0}) // Wrong span
	assert.ErrorIs(err, ErrSpan)

	// Validate and try adding again
	require.NoError(d.Validate())
	assert.True(d.valid)
	err = d.AddGate(gate.X(), []int{2}) // Add after validation
	assert.Error(err)
	assert.Contains(err.Error(), "already validated") // Check error message
}

func TestDAG_AddMeasure(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	d := New(2, 1)

	// Add H(0)
	err := d.AddGate(gate.H(), []int{0})
	require.NoError(err)
	h0Node := d.nodes[d.last[0]] // Get H(0) node

	// Add Measure(0, 0)
	err = d.AddMeasure(0, 0)
	require.NoError(err)
	assert.Len(d.nodes, 2)
	var mNode *Node
	for id, n := range d.nodes {
		if id != h0Node.ID {
			mNode = n
			break
		}
	}
	require.NotNil(mNode)
	assert.Equal(gate.Measure(), mNode.G)
	assert.Equal([]int{0}, mNode.Qubits)
	assert.Equal(0, mNode.Cbit)
	// Measure depends on the last op on qubit 0 (H(0))
	require.Len(mNode.parents, 1)
	assert.Contains(mNode.parents, h0Node.ID)
	assert.Empty(mNode.children)
	assert.Equal(mNode.ID, d.last[0]) // Measure is now last on qubit 0
	assert.Equal([]NodeID{h0Node.ID, mNode.ID}, d.byQ[0])

	// Check H(0) children updated
	assert.Equal([]NodeID{mNode.ID}, h0Node.children)

	// Test errors
	err = d.AddMeasure(2, 0) // Qubit out of range
	assert.ErrorIs(err, ErrBadQubit)
	err = d.AddMeasure(1, 1) // Clbit out of range
	assert.ErrorIs(err, ErrBadClbit)

	// Validate and try adding again
	require.NoError(d.Validate())
	assert.True(d.valid)
	err = d.AddMeasure(1, 0) // Add after validation
	assert.Error(err)
	assert.Contains(err.Error(), "already validated") // Check error message
}

func TestDAG_Validate_Success(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	d := New(2, 0)
	d.AddGate(gate.H(), []int{0})
	d.AddGate(gate.CNOT(), []int{0, 1})
	err := d.Validate()
	require.NoError(err)
	assert.True(d.valid)
	// Validate again should be no-op
	err = d.Validate()
	require.NoError(err)
	assert.True(d.valid)
}

func TestDAG_TopoSort_Depth_Operations(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	// H(0) --- CNOT(0,1) --- X(1)
	//          |
	// H(2) ----+  (CNOT depends on H(0) and H(2))
	d := New(3, 0)

	err := d.AddGate(gate.H(), []int{0}) // id 1 (assume) -> nodeA
	require.NoError(err)
	nodeA := d.nodes[d.last[0]]

	err = d.AddGate(gate.H(), []int{2}) // id 2 -> nodeB
	require.NoError(err)
	nodeB := d.nodes[d.last[2]]

	// CNOT(0, 1) depends on H(0) [nodeA] and last op on qubit 1 (none initially)
	// Let's add H(1) first to make dependencies clearer for CNOT(0,1)
	// H(0) --- CNOT(0,1) --- X(1)
	//          |
	// H(1) ----+
	// H(2) --- ? (Independent for now)
	// Let's redo the example slightly:
	// H(0) --- CNOT(0,1) --- X(1)
	// H(2) -----------------+
	// CNOT depends on H(0) and last op on qubit 1 (none)
	// X(1) depends on CNOT(0,1)
	// Let's stick to the original comment's intent: CNOT depends on H(0) and H(2)
	// This means CNOT must involve qubit 2, e.g., CNOT(0, 2) or CNOT(2, 1) etc.
	// Let's assume the comment meant:
	// H(0) --- CNOT(0,1) --- X(1)
	// H(2) -----------------+ (CNOT(0,1) depends on H(0) and H(2)) -> This is wrong, CNOT(0,1) only touches 0 and 1.
	// Let's assume the intended graph was:
	// H(0) --- CNOT(0,1) --- X(1)
	// H(2) ---+
	// CNOT(0,1) depends on H(0) (last on q0) and potentially last on q1 (none)
	// X(1) depends on CNOT(0,1) (last on q1)
	// H(2) is independent initially.

	// Resetting based on the code:
	// H(0) -> nodeA (last[0]=A)
	// H(2) -> nodeB (last[2]=B)
	// CNOT(0, 1) -> nodeC. Parents: last[0]=A, last[1]=0. So parent is A. (last[0]=C, last[1]=C)
	// X(1) -> nodeD. Parents: last[1]=C. So parent is C. (last[1]=D)

	err = d.AddGate(gate.CNOT(), []int{0, 1}) // id 3, parent: A -> nodeC
	require.NoError(err)
	nodeC := d.nodes[d.last[0]] // CNOT is last on 0 and 1
	require.Len(nodeC.parents, 1, "CNOT should have 1 parent (H(0))")
	assert.Contains(nodeC.parents, nodeA.ID)

	err = d.AddGate(gate.X(), []int{1}) // id 4, parent: C -> nodeD
	require.NoError(err)
	nodeD := d.nodes[d.last[1]] // X is last on 1
	require.Len(nodeD.parents, 1, "X should have 1 parent (CNOT)")
	assert.Contains(nodeD.parents, nodeC.ID)

	require.NoError(d.Validate())

	// Expected Topo Order: [A, B, C, D] or [B, A, C, D]
	// Expected Depth: 3 (layers: {A,B}, {C}, {D}) -> MaxStep = 2, Depth = 3
	// Expected Operations: Nodes in topological order

	order := d.calculateTopoSort()
	assert.Len(order, 4)
	// Check if A and B appear before C, and C before D
	posA, posB, posC, posD := -1, -1, -1, -1 // Corrected initialization
	for i, node := range order {
		switch node.ID {
		case nodeA.ID:
			posA = i
		case nodeB.ID:
			posB = i
		case nodeC.ID:
			posC = i
		case nodeD.ID:
			posD = i
		}
	}
	require.NotEqual(-1, posA, "Node A not found in order")
	require.NotEqual(-1, posB, "Node B not found in order")
	require.NotEqual(-1, posC, "Node C not found in order")
	require.NotEqual(-1, posD, "Node D not found in order")

	assert.True(posA < posC, "A should be before C")
	// B is independent of C in this corrected graph structure
	// assert.True(posB < posC, "B should be before C") // This assertion is not guaranteed
	assert.True(posC < posD, "C should be before D")

	depth := d.Depth()
	// Layers: {A, B}, {C}, {D} -> Indices 0, 1, 2 -> Max index is 2 -> Depth is 3
	assert.Equal(3, depth) // Layers 0, 1, 2 -> Depth 3

	ops := d.Operations()
	require.Len(ops, 4)
	assert.Equal(order[0].ID, ops[0].ID)
	assert.Equal(order[1].ID, ops[1].ID)
	assert.Equal(order[2].ID, ops[2].ID)
	assert.Equal(order[3].ID, ops[3].ID)
}

// TestCycleDetect uses the existing test logic but ensures it uses AddGate
func TestCycleDetect(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	d := New(1, 0) // Use direct integer arguments

	// Add two gates sequentially on the same qubit
	err := d.AddGate(gate.H(), []int{0}) // Node A
	require.NoError(err)
	nodeA := d.nodes[d.last[0]]

	err = d.AddGate(gate.X(), []int{0}) // Node B, parent: A
	require.NoError(err)
	nodeB := d.nodes[d.last[0]]

	// Manually create a cycle B -> A
	// This simulates an invalid state that Validate should catch.
	// Note: This bypasses the normal AddGate logic for testing Validate directly.
	nodeB.children = append(nodeB.children, nodeA.ID)
	nodeA.parents = append(nodeA.parents, nodeB.ID)

	// Reset valid flag to false before calling Validate, as AddGate doesn't set it
	d.valid = false
	err = d.Validate()
	assert.Error(err, "Validate should detect the cycle")
	assert.Contains(err.Error(), "cycle detected", "Error message should mention cycle")
	assert.False(d.valid, "DAG should remain invalid after cycle detection")
}
