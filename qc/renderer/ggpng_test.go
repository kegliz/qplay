package renderer

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants for better maintainability
const (
	defaultTestTimeout = 10 * time.Second
	defaultCellSize    = 80
)

// tempTestFile creates a temporary test file and returns cleanup function
func tempTestFile(t *testing.T, filename string) (string, func()) {
	t.Helper()

	tempDir := t.TempDir() // Automatically cleaned up by Go
	fullPath := filepath.Join(tempDir, filename)

	cleanup := func() {
		if _, err := os.Stat(fullPath); err == nil {
			os.Remove(fullPath)
		}
	}

	return fullPath, cleanup
}

// withTimeout runs a function with timeout
func withTimeout(t *testing.T, timeout time.Duration, fn func() error) {
	t.Helper()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(timeout):
		t.Fatalf("operation timed out after %v", timeout)
	}
}

// TestInterfaces ensures the DAG type implements the interfaces
func TestInterfaces(t *testing.T) {
	// compile-time check
	var _ Renderer = (*GGPNG)(nil) // GGPNG implements Renderer
}

func TestGGPNG_Render(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Build circuit using Builder
	b := builder.New(builder.Q(3), builder.C(1))
	b.H(0)
	b.Toffoli(0, 1, 2)
	b.Measure(2, 0) // Measure q2 into cbit 0

	c, err := b.BuildCircuit() // Use BuildCircuit interface
	require.NoError(err, "building circuit failed")
	require.NotNil(c, "built circuit should not be nil")

	renderer := NewRenderer(80)
	img, err := renderer.Render(c)
	assert.NoError(err, "image rendered")
	require.NotNil(img, "image should not be nil")

	assert.Greater(img.Bounds().Dx(), 0, "image should not be empty")
	assert.Greater(img.Bounds().Dy(), 0, "image should not be empty")

	// Test rendering an empty circuit
	bEmpty := builder.New(builder.Q(1))
	drEmpty, err := bEmpty.BuildDAG()
	require.NoError(err, "building empty DAG failed")
	require.NotNil(drEmpty, "built empty DAG should not be nil")
	cEmpty := circuit.FromDAG(drEmpty)
	require.NotNil(cEmpty, "creating circuit from empty DAG failed")
	imgEmpty, err := renderer.Render(cEmpty)
	assert.NoError(err)
	require.NotNil(imgEmpty)
	assert.Greater(imgEmpty.Bounds().Dx(), 0) // Should still have width for wires
	assert.Greater(imgEmpty.Bounds().Dy(), 0) // Should still have height for wires
}

func TestGGPNG_Save(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Build circuit using Builder
	b := builder.New(builder.Q(3), builder.C(1))
	b.H(0)
	b.Toffoli(0, 1, 2)
	b.Measure(2, 0)

	// Build the circuit first
	c1, err := b.BuildCircuit() // Use BuildCircuit interface
	require.NoError(err, "building circuit 1 failed")
	require.NotNil(c1, "built circuit should not be nil")

	renderer := NewRenderer(80)
	filePath1, cleanup1 := tempTestFile(t, "ggpng_test1.png")
	defer cleanup1()

	err = renderer.Save(filePath1, c1) // Save first circuit
	assert.NoError(err, "image saved")

	// Check if the file exists and is valid PNG
	f1, err := os.Open(filePath1)
	require.NoError(err, "file %s should exist", filePath1)
	defer f1.Close()
	_, err = png.Decode(f1)
	assert.NoError(err, "file %s should be a valid PNG", filePath1)

	// Draw a more complex circuit
	b2 := builder.New(builder.Q(3))
	b2.H(0)
	b2.CNOT(0, 1)
	b2.CZ(1, 2) // Added CZ gate
	b2.SWAP(0, 2)
	b2.Fredkin(1, 0, 2) // Control q1, swap q0 and q2

	// Build the circuit first
	c2, err := b2.BuildCircuit() // Use BuildCircuit interface
	require.NoError(err, "building circuit 2 failed")
	require.NotNil(c2, "built circuit 2 should not be nil")

	filePath2, cleanup2 := tempTestFile(t, "ggpng_test2.png")
	defer cleanup2()

	err = renderer.Save(filePath2, c2) // Save second circuit
	assert.NoError(err, "image saved")

	// Check if the file exists and is valid PNG
	f2, err := os.Open(filePath2)
	require.NoError(err, "file %s should exist", filePath2)
	defer f2.Close()
	_, err = png.Decode(f2)
	assert.NoError(err, "file %s should be a valid PNG", filePath2)
}
