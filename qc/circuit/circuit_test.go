package circuit_test

import (
	"sort"
	"strconv"
	"testing"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/kegliz/qplay/qc/gate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromBuilderRoundtrip(t *testing.T) {
	require := require.New(t)

	b := builder.New(builder.Q(2))
	b.H(0).CNOT(0, 1)

	dr, err := b.BuildDAG()
	require.NoError(err)

	c := circuit.FromDAG(dr)

	require.Equal(2, c.Qubits())
	require.Equal(0, c.Clbits())
	require.Equal(1, c.MaxStep())
	require.Equal(2, c.Depth())
}

func TestCircuit_Properties(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	b := builder.New(builder.Q(3), builder.C(1))
	b.H(0)
	b.CNOT(0, 1)
	b.Toffoli(0, 1, 2)
	b.Measure(2, 0)

	// Build the DAG first
	dr, err := b.BuildDAG() // Use BuildDAG interface
	require.NoError(err, "building DAG failed")
	require.NotNil(dr, "built DAG should not be nil")

	// Create the Circuit from the DAG
	c := circuit.FromDAG(dr) // Use FromDAG with DAGReader
	require.NotNil(c, "Circuit should not be nil")

	assert.Equal(3, c.Qubits(), "Qubit count mismatch")
	assert.Equal(1, c.Clbits(), "Classical bit count mismatch")

	// Depth calculation depends on the longest path in the DAG
	// H(0) -> CNOT(0,1) -> Toffoli(0,1,2) -> Measure(2,0)
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
	b := builder.New(builder.Q(3))
	b.H(0)       // Step 0, Line 0
	b.H(1)       // Step 0, Line 1
	b.CNOT(0, 2) // Step 1, Line 0 (depends on H(0))
	b.X(1)       // Step 1, Line 1 (depends on H(1))
	b.CZ(0, 1)   // Step 2, Line 0 (depends on both paths)

	// Build the DAG first
	dr, err := b.BuildDAG() // Use BuildDAG interface
	require.NoError(err, "building DAG failed")
	require.NotNil(dr, "built DAG should not be nil")

	// Create the Circuit from the DAG
	c := circuit.FromDAG(dr) // Use FromDAG with DAGReader
	require.NotNil(c)

	ops := c.Operations()
	require.Len(ops, 5)

	// Expected layout:
	// Step 0: H(0) [line 0], H(1) [line 1]
	// Step 1: CNOT(0, 2) [line 0],  X(1) [line 1]
	// Step 2: CZ(0, 1) [line 0]

	assert.Equal(2, c.MaxStep(), "MaxStep should be 2")
	assert.Equal(3, c.Depth(), "Depth should be 3")

	// Verify timestep and line for each operation
	opMap := make(map[string]circuit.Operation)
	for _, op := range ops {
		key := op.G.Name()
		if len(op.Qubits) > 0 { // Add qubit info for uniqueness
			qubitsCopy := append([]int(nil), op.Qubits...)
			sort.Ints(qubitsCopy)
			qubitStr := ""
			for i, q := range qubitsCopy {
				if i > 0 {
					qubitStr += ","
				}
				qubitStr += strconv.Itoa(q)
			}
			key += "_" + qubitStr
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
	cnot02, ok := opMap["CNOT_0,2"]
	require.True(ok, "CNOT(0, 2) not found")
	assert.Equal(1, cnot02.TimeStep, "CNOT(0, 2) timestep")
	assert.Equal(0, cnot02.Line, "CNOT(0, 2) line") // Line is min qubit index

	// Check X(1)
	x1, ok := opMap["X_1"]
	require.True(ok, "X(1) not found")
	assert.Equal(1, x1.TimeStep, "X(1) timestep")
	assert.Equal(1, x1.Line, "X(1) line")

	// Check CZ(0, 1)
	cz01, ok := opMap["CZ_0,1"]
	require.True(ok, "CZ(0, 1) not found")
	assert.Equal(2, cz01.TimeStep, "CZ(0, 1) timestep") // Depends on previous steps
	assert.Equal(0, cz01.Line, "CZ(0, 1) line")
}

func TestCircuit_Empty(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	b := builder.New(builder.Q(2), builder.C(1))

	// Build the DAG first
	dr, err := b.BuildDAG() // Use BuildDAG interface
	require.NoError(err, "building empty DAG failed")
	require.NotNil(dr, "built empty DAG should not be nil")

	// Create the Circuit from the DAG
	c := circuit.FromDAG(dr) // Use FromDAG with DAGReader
	require.NotNil(c)

	assert.Equal(2, c.Qubits())
	assert.Equal(1, c.Clbits())
	assert.Equal(-1, c.MaxStep()) // MaxStep is -1 for empty circuit
	assert.Equal(0, c.Depth())    // Depth is 0 for empty circuit
	assert.Empty(c.Operations())
}

func TestCircuit_FromBuildCircuit(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	b := builder.New(builder.Q(3), builder.C(1))
	b.H(0)
	b.CNOT(0, 1)
	b.Toffoli(0, 1, 2)
	b.Measure(2, 0)

	// Build the Circuit directly
	c, err := b.BuildCircuit()
	require.NoError(err, "building circuit failed")
	require.NotNil(c, "built circuit should not be nil")

	assert.Equal(3, c.Qubits(), "Qubit count mismatch")
	assert.Equal(1, c.Clbits(), "Classical bit count mismatch")
}
