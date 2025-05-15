package itsu

import (
	"sort"
	"testing"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/simulator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pretty prints the histogram in a deterministic, sorted order
func prettyPS(t *testing.T, hist map[string]int, shots int) {
	keys := make([]string, 0, len(hist))
	for k := range hist {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	t.Log("Histogram (key : count / %):")
	for _, k := range keys {
		c := hist[k]
		pct := 100 * float64(c) / float64(shots)
		t.Logf("  %s : %4d (%.1f%%)", k, c, pct)
	}
}

// TestBellState prepares the |Φ⁺⟩ Bell state and checks ~50/50 statistics.
func TestBellStatePS(t *testing.T) {
	shots := 2048
	b := builder.New(builder.Q(2), builder.C(2))
	b.H(0).CNOT(0, 1).Measure(0, 0).Measure(1, 1)

	c, err := b.BuildCircuit()
	require.NoError(t, err)

	sim := simulator.NewSimulator(simulator.SimulatorOptions{Shots: shots, Runner: NewItsuOneShotRunner()})
	hist, err := sim.RunParallelStatic(c)
	require.NoError(t, err)

	prettyPS(t, hist, shots)

	assert.InDelta(t, 0.5, float64(hist["00"])/float64(shots), 0.1)
	assert.InDelta(t, 0.5, float64(hist["11"])/float64(shots), 0.1)
	assert.Equal(t, 0, hist["01"], "unexpected outcome 01")
	assert.Equal(t, 0, hist["10"], "unexpected outcome 10")
}

// TestGrover2Qubit demonstrates one Grover iteration on 2‑qubit search space
// amplifying the |11⟩ state.
func TestGrover2QubitPS(t *testing.T) {
	shots := 1024
	b := builder.New(builder.Q(2), builder.C(2))

	// — initial superposition —
	b.H(0).H(1)

	// — oracle marks |11⟩ by phase flip (controlled‑Z) —
	b.CZ(0, 1)

	// — diffusion operator —
	b.H(0).H(1)
	b.X(0).X(1)
	b.CZ(0, 1)
	b.X(0).X(1)
	b.H(0).H(1)

	// — measurement —
	b.Measure(0, 0).Measure(1, 1)

	c, err := b.BuildCircuit()
	require.NoError(t, err)

	sim := simulator.NewSimulator(simulator.SimulatorOptions{Shots: shots, Runner: NewItsuOneShotRunner()})
	hist, err := sim.RunParallelStatic(c)
	require.NoError(t, err)

	prettyPS(t, hist, shots)

	assert.Greater(t, hist["11"], int(0.75*float64(shots)), "Grover did not amplify |11⟩ sufficiently")
}

func TestGrover3QubitPS(t *testing.T) {
	shots := 1024
	b := builder.New(builder.Q(3), builder.C(3))

	// — initial superposition —
	b.H(0).H(1).H(2)

	// — oracle marks |111⟩ by phase flip (CCZ) —
	// Implement CCZ using H and Toffoli: H(target) Toffoli(c1, c2, target) H(target)
	b.H(2).Toffoli(0, 1, 2).H(2)

	// — diffusion operator (3 qubits) —
	// HHH - XXX - CCZ - XXX - HHH
	b.H(0).H(1).H(2)
	b.X(0).X(1).X(2)
	// CCZ
	b.H(2).Toffoli(0, 1, 2).H(2)
	b.X(0).X(1).X(2)
	b.H(0).H(1).H(2)

	// — measurement —
	b.Measure(0, 0).Measure(1, 1).Measure(2, 2)

	c, err := b.BuildCircuit()
	require.NoError(t, err)

	sim := simulator.NewSimulator(simulator.SimulatorOptions{Shots: shots, Runner: NewItsuOneShotRunner()})
	hist, err := sim.RunParallelStatic(c)
	require.NoError(t, err)

	prettyPS(t, hist, shots)

	assert.Greater(t, hist["111"], int(0.75*float64(shots)), "Grover did not amplify |111⟩ sufficiently")
}
