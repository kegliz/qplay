package circuit

import (
	"testing"

	"github.com/kegliz/qplay/qc/dag/builder" // Import builder
	// "github.com/kegliz/qplay/qc/dag" // Remove direct dag import if not needed
	"github.com/kegliz/qplay/qc/gate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuit_Properties(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	b := builder.New(builder.Q(3), builder.C(1)) // Use builder
	b.H(0)
	b.CNOT(0, 1)
	b.Toffoli(0, 1, 2)
	b.Measure(2, 0)

	// Build the DAG first
	d, err := b.Build()
	require.NoError(err, "building DAG failed")
	require.NotNil(d, "built DAG should not be nil")

	// Create the Circuit from the DAG
	c := FromDAG(d) // Use FromDAG directly as we are in the circuit package
	require.NotNil(c, "Circuit should not be nil")

	assert.Equal(3, c.Qubits(), "Qubit count mismatch")
	assert.Equal(1, c.Clbits(), "Classical bit count mismatch")

	// Depth calculation depends on the longest path in the DAG
	// H(0) -> CNOT(0,1) -> Toffoli(0,1,2) -> Measure(2,0)
	// Path 0: H(0) -> CNOT(0,1) -> Toffoli(0,1,2) (length 3 nodes, depth 3 layers)
	// Path 1: CNOT(0,1) -> Toffoli(0,1,2)
	// Path 2: Toffoli(0,1,2) -> Measure(2,0)
	// Longest path involves 4 operations, so 4 layers/timesteps (0, 1, 2, 3)
	// Depth = MaxStep + 1
	assert.Equal(3, c.MaxStep(), "MaxStep mismatch")
	assert.Equal(4, c.Depth(), "Depth mismatch")

	ops := c.Operations()
	assert.Len(ops, 4, "Operation count mismatch")

	// Check properties of the first operation (H(0))
	assert.Equal(gate.H(), ops[0].G, "First gate mismatch")
	assert.Equal([]int{0}, ops[0].Qubits, "First gate qubits mismatch")
	assert.Equal(-1, ops[0].Cbit, "First gate cbit mismatch")
	assert.Equal(0, ops[0].TimeStep, "First gate timestep mismatch")
	assert.Equal(0, ops[0].Line, "First gate line mismatch")

	// Check properties of the last operation (Measure(2,0))
	assert.Equal(gate.Measure(), ops[3].G, "Last gate mismatch")
	assert.Equal([]int{2}, ops[3].Qubits, "Last gate qubits mismatch")
	assert.Equal(0, ops[3].Cbit, "Last gate cbit mismatch")
	assert.Equal(3, ops[3].TimeStep, "Last gate timestep mismatch")
	assert.Equal(2, ops[3].Line, "Last gate line mismatch")

	// Check timestep ordering
	for i := 0; i < len(ops)-1; i++ {
		assert.LessOrEqual(ops[i].TimeStep, ops[i+1].TimeStep, "Operations should be sorted by timestep")
		if ops[i].TimeStep == ops[i+1].TimeStep {
			assert.LessOrEqual(ops[i].Line, ops[i+1].Line, "Operations at same timestep should be sorted by line")
		}
	}
}

func TestCircuit_Layout(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Circuit where gates can run in parallel
	// H(0) | H(1)
	// CNOT(0, 2) | X(1)
	b := builder.New(builder.Q(3)) // Use builder
	b.H(0)
	b.H(1)       // Should be at timestep 0, line 1
	b.CNOT(0, 2) // Depends on H(0), should be at timestep 1, line 0
	b.X(1)       // Depends on H(1), should be at timestep 1, line 1

	// Build the DAG first
	d, err := b.Build()
	require.NoError(err, "building DAG failed")
	require.NotNil(d, "built DAG should not be nil")

	// Create the Circuit from the DAG
	c := FromDAG(d)
	require.NotNil(c)

	ops := c.Operations()
	require.Len(ops, 4)

	// Expected layout:
	// Step 0: H(0) [line 0], H(1) [line 1]
	// Step 1: CNOT(0, 2) [line 0], X(1) [line 1]

	assert.Equal(1, c.MaxStep(), "MaxStep should be 1")
	assert.Equal(2, c.Depth(), "Depth should be 2")

	// Verify timestep and line for each operation
	opMap := make(map[string]Operation)
	for _, op := range ops {
		key := op.G.Name()
		if len(op.Qubits) > 0 { // Add qubit info for uniqueness if needed
			key += "_"
			// Simple string conversion for keys, ensure uniqueness for test
			qubitStr := ""
			for i, q := range op.Qubits {
				if i > 0 {
					qubitStr += ","
				}
				qubitStr += string(rune(q + '0'))
			}
			key += qubitStr
		}
		opMap[key] = op
	}

	// Check H(0)
	h0, ok := opMap["H_0"]
	require.True(ok, "H(0) not found")
	assert.Equal(0, h0.TimeStep, "H(0) timestep")
	assert.Equal(0, h0.Line, "H(0) line")

	// Check H(1)
	h1, ok := opMap["H_1"]
	require.True(ok, "H(1) not found")
	assert.Equal(0, h1.TimeStep, "H(1) timestep")
	assert.Equal(1, h1.Line, "H(1) line")

	// Check CNOT(0, 2)
	cnot02, ok := opMap["CNOT_0,2"] // Adjusted key based on logic above
	require.True(ok, "CNOT(0, 2) not found")
	assert.Equal(1, cnot02.TimeStep, "CNOT(0, 2) timestep")
	assert.Equal(0, cnot02.Line, "CNOT(0, 2) line") // Line is min qubit index

	// Check X(1)
	x1, ok := opMap["X_1"]
	require.True(ok, "X(1) not found")
	assert.Equal(1, x1.TimeStep, "X(1) timestep")
	assert.Equal(1, x1.Line, "X(1) line")
}

func TestCircuit_Empty(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	b := builder.New(builder.Q(2), builder.C(1)) // Use builder
	// Build the DAG first
	d, err := b.Build()
	require.NoError(err, "building empty DAG failed")
	require.NotNil(d, "built empty DAG should not be nil")

	// Create the Circuit from the DAG
	c := FromDAG(d)
	require.NotNil(c)

	assert.Equal(2, c.Qubits())
	assert.Equal(1, c.Clbits())
	assert.Equal(-1, c.MaxStep()) // MaxStep is -1 for empty circuit (no operations)
	assert.Equal(0, c.Depth())    // Depth is 0 (MaxStep + 1)
	assert.Empty(c.Operations())
}
