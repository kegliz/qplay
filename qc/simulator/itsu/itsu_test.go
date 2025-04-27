package itsu

import (
	"sort"
	"testing"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pretty prints the histogram in a deterministic, sorted order
func pretty(t *testing.T, hist map[string]int, shots int) {
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
func TestBellState(t *testing.T) {
	shots := 1024
	b := builder.New(builder.Q(2), builder.C(2))
	b.H(0).CNOT(0, 1).Measure(0, 0).Measure(1, 1)

	c, err := b.BuildCircuit()
	require.NoError(t, err)

	sim := New(shots)
	hist, err := sim.Run(c)
	require.NoError(t, err)

	pretty(t, hist, shots)

	assert.InDelta(t, 0.5, float64(hist["00"])/float64(shots), 0.1)
	assert.InDelta(t, 0.5, float64(hist["11"])/float64(shots), 0.1)
	assert.Equal(t, 0, hist["01"], "unexpected outcome 01")
	assert.Equal(t, 0, hist["10"], "unexpected outcome 10")
}

// TestGrover2Qubit demonstrates one Grover iteration on 2‑qubit search space
// amplifying the |11⟩ state.
func TestGrover2Qubit(t *testing.T) {
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

	sim := New(shots)
	hist, err := sim.Run(c)
	require.NoError(t, err)

	pretty(t, hist, shots)

	assert.Greater(t, hist["11"], int(0.75*float64(shots)), "Grover did not amplify |11⟩ sufficiently")
}
