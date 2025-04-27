package gate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinGates(t *testing.T) {
	tests := []struct {
		name       string
		gate       Gate
		wantName   string
		wantSpan   int
		wantSymbol string
		wantTgts   []int
		wantCtrls  []int
	}{
		{"Hadamard", H(), "H", 1, "H", []int{0}, []int{}},
		{"PauliX", X(), "X", 1, "X", []int{0}, []int{}},
		{"PhaseS", S(), "S", 1, "S", []int{0}, []int{}},
		{"Measure", Measure(), "MEASURE", 1, "M", []int{0}, []int{}},
		{"SWAP", Swap(), "SWAP", 2, "×", []int{0, 1}, []int{}},
		{"CNOT", CNOT(), "CNOT", 2, "⊕", []int{1}, []int{0}},             // Target=1, Control=0
		{"CZ", CZ(), "CZ", 2, "●", []int{1}, []int{0}},                   // Added CZ test case
		{"Toffoli", Toffoli(), "TOFFOLI", 3, "T", []int{2}, []int{0, 1}}, // Target=2, Controls=0,1
		{"Fredkin", Fredkin(), "FREDKIN", 3, "F", []int{1, 2}, []int{0}}, // Targets=1,2, Control=0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tt.wantName, tt.gate.Name(), "Name mismatch")
			assert.Equal(tt.wantSpan, tt.gate.QubitSpan(), "QubitSpan mismatch")
			assert.Equal(tt.wantSymbol, tt.gate.DrawSymbol(), "DrawSymbol mismatch")
			assert.Equal(tt.wantTgts, tt.gate.Targets(), "Targets mismatch")
			assert.Equal(tt.wantCtrls, tt.gate.Controls(), "Controls mismatch")
		})
	}
}

func TestFactory(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	testCases := []struct {
		alias    string
		expected Gate
	}{
		{"h", H()},
		{" H ", H()}, // Test trimming/normalization
		{"x", X()},
		{"s", S()},
		{"swap", Swap()},
		{"SWAP", Swap()},
		{"cx", CNOT()},
		{"cnot", CNOT()},
		{"CNOT", CNOT()},
		{"cz", CZ()}, // Added CZ alias test
		{"CZ", CZ()}, // Added CZ alias test (uppercase)
		{"t", Toffoli()},
		{"toffoli", Toffoli()},
		{"ccx", Toffoli()},
		{"fredkin", Fredkin()},
		{"cswap", Fredkin()},
		{"m", Measure()},
		{"measure", Measure()},
		{"meas", Measure()},
	}

	for _, tc := range testCases {
		t.Run("Alias_"+tc.alias, func(t *testing.T) {
			g, err := Factory(tc.alias)
			require.NoError(err, "Factory failed for alias: %s", tc.alias)
			// Check for tc.expected is the same singleton as g
			assert.Same(tc.expected, g, "Factory should return singleton instance for alias: %s", tc.alias)
		})
	}

	// Test unknown gate
	unknownName := "unknown_gate"
	g, err := Factory(unknownName)
	assert.Nil(g, "Factory should return nil for unknown gate")
	require.Error(err, "Factory should return error for unknown gate")
	assert.ErrorIs(err, ErrUnknownGate{unknownName}, "Error type should be ErrUnknownGate")
	assert.Contains(err.Error(), unknownName, "Error message should contain the unknown name")
}

// Test Factory with a non-existent gate
func TestFactory_NonExistentGate(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Assuming Z gate doesn't exist yet
	nonExistentGate := "nonExistent_gate"
	g, err := Factory(nonExistentGate)
	assert.Nil(g, "Factory should return nil for non-existent gate")
	require.Error(err, "Factory should return error for non-existent gate")
	assert.ErrorIs(err, ErrUnknownGate{nonExistentGate}, "Error type should be ErrUnknownGate")
	assert.Contains(err.Error(), nonExistentGate, "Error message should contain the non-existent gate name")
}
