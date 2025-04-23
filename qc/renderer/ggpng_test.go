package renderer

import (
	"image/png"
	"os"
	"testing"

	"github.com/kegliz/qplay/qc/circuit"     // Import circuit
	"github.com/kegliz/qplay/qc/dag/builder" // Import builder

	// "github.com/kegliz/qplay/qc/dag" // Remove direct dag import if not needed elsewhere
	// "github.com/kegliz/qplay/qc/gate" // Keep if specific gates needed for builder
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGGPNG_Render(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Build circuit using Builder
	b := builder.New(builder.Q(3), builder.C(1)) // Use builder
	b.H(0)
	b.Toffoli(0, 1, 2)
	b.Measure(2, 0) // Measure q2 into cbit 0

	// Build the DAG first
	d, err := b.Build()
	require.NoError(err, "building DAG failed")
	require.NotNil(d, "built DAG should not be nil")

	// Create the Circuit from the DAG
	c := circuit.FromDAG(d)
	require.NotNil(c, "creating circuit from DAG failed")

	renderer := NewRenderer(80)
	img, err := renderer.Render(c)
	assert.NoError(err, "image rendered")
	require.NotNil(img, "image should not be nil")

	assert.Greater(img.Bounds().Dx(), 0, "image should not be empty")
	assert.Greater(img.Bounds().Dy(), 0, "image should not be empty")

	// Test rendering an empty circuit
	bEmpty := builder.New(builder.Q(1)) // Use builder
	dEmpty, err := bEmpty.Build()
	require.NoError(err, "building empty DAG failed")
	require.NotNil(dEmpty, "built empty DAG should not be nil")
	cEmpty := circuit.FromDAG(dEmpty) // Create circuit from empty DAG
	require.NotNil(cEmpty, "creating circuit from empty DAG failed")
	imgEmpty, err := renderer.Render(cEmpty)
	assert.NoError(err)
	require.NotNil(imgEmpty)
	assert.Greater(imgEmpty.Bounds().Dx(), 0) // Should still have width for wires
	assert.Greater(imgEmpty.Bounds().Dy(), 0) // Should still have height for wires

	// Test rendering circuit with unsupported gate (if any were defined)
	// For now, the default case handles unknown single-qubit gates
}
func TestGGPNG_Save(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Build circuit using Builder
	b := builder.New(builder.Q(3), builder.C(1)) // Use builder
	b.H(0)
	b.Toffoli(0, 1, 2)
	b.Measure(2, 0)

	// Build the DAG first
	d1, err := b.Build()
	require.NoError(err, "building DAG 1 failed")
	require.NotNil(d1, "built DAG 1 should not be nil")

	// Create the Circuit from the DAG
	c1 := circuit.FromDAG(d1)
	require.NotNil(c1, "creating circuit 1 from DAG failed")

	renderer := NewRenderer(80)
	filePath1 := "ggpng_test1.png"
	//defer os.Remove(filePath1) // Clean up

	err = renderer.Save(filePath1, c1) // Save first circuit
	assert.NoError(err, "image saved")

	// Check if the file exists and is valid PNG
	f1, err := os.Open(filePath1)
	require.NoError(err, "file %s should exist", filePath1)
	defer f1.Close()
	_, err = png.Decode(f1)
	assert.NoError(err, "file %s should be a valid PNG", filePath1)

	// Draw a more complex circuit
	b2 := builder.New(builder.Q(3)) // Use builder
	b2.H(0)
	b2.CNOT(0, 1)
	b2.CNOT(1, 2)
	b2.SWAP(0, 2)
	b2.Fredkin(1, 0, 2) // Control q1, swap q0 and q2

	// Build the DAG first
	d2, err := b2.Build()
	require.NoError(err, "building DAG 2 failed")
	require.NotNil(d2, "built DAG 2 should not be nil")

	// Create the Circuit from the DAG
	c2 := circuit.FromDAG(d2)
	require.NotNil(c2, "creating circuit 2 from DAG failed")

	filePath2 := "ggpng_test2.png"
	//defer os.Remove(filePath2) // Clean up

	err = renderer.Save(filePath2, c2) // Save second circuit
	assert.NoError(err, "image saved")

	// Check if the file exists and is valid PNG
	f2, err := os.Open(filePath2)
	require.NoError(err, "file %s should exist", filePath2)
	defer f2.Close()
	_, err = png.Decode(f2)
	assert.NoError(err, "file %s should be a valid PNG", filePath2)
}
