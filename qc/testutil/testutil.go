// Package testutil provides testing utilities and constants for the qc package tests.
// This improves maintainability by centralizing test configuration and common patterns.
package testutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/stretchr/testify/require"
)

// Test constants for consistent configuration across tests
const (
	// Test timeouts
	DefaultTestTimeout = 10 * time.Second
	LongTestTimeout    = 30 * time.Second
	BenchmarkTimeout   = 60 * time.Second

	// Simulation parameters
	DefaultShots   = 1024
	SmallShots     = 100
	LargeShots     = 2048
	BenchmarkShots = 8192
	DefaultWorkers = 8

	// Circuit parameters
	DefaultQubits = 3
	SmallQubits   = 2
	LargeQubits   = 7

	// Statistical tolerances
	DefaultTolerance = 0.1  // 10% tolerance for statistical tests
	StrictTolerance  = 0.05 // 5% tolerance for precise tests

	// File testing
	TestFilePrefix = "qc_test_"
	PNGTestSuffix  = ".png"
)

// TestConfig holds configuration for test scenarios
type TestConfig struct {
	Shots     int
	Qubits    int
	Workers   int
	Timeout   time.Duration
	Tolerance float64
}

// Predefined test configurations
var (
	QuickTestConfig = TestConfig{
		Shots:     SmallShots,
		Qubits:    SmallQubits,
		Workers:   4,
		Timeout:   DefaultTestTimeout,
		Tolerance: DefaultTolerance,
	}

	StandardTestConfig = TestConfig{
		Shots:     DefaultShots,
		Qubits:    DefaultQubits,
		Workers:   DefaultWorkers,
		Timeout:   DefaultTestTimeout,
		Tolerance: DefaultTolerance,
	}

	BenchmarkTestConfig = TestConfig{
		Shots:     BenchmarkShots,
		Qubits:    LargeQubits,
		Workers:   DefaultWorkers,
		Timeout:   BenchmarkTimeout,
		Tolerance: StrictTolerance,
	}

	// ConservativeTestConfig provides very conservative settings for resource-constrained environments
	ConservativeTestConfig = TestConfig{
		Shots:     50,              // Very small shot count
		Qubits:    2,               // Minimal qubits
		Workers:   2,               // Few workers
		Timeout:   5 * time.Second, // Short timeout
		Tolerance: DefaultTolerance,
	}
)

// WithTimeout creates a context with timeout for test operations
func WithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// TempFile creates a temporary test file and returns cleanup function
func TempFile(t *testing.T, suffix string) (string, func()) {
	t.Helper()

	tempDir := t.TempDir() // Go 1.15+ automatically cleans this up
	filename := TestFilePrefix + t.Name() + suffix
	filepath := filepath.Join(tempDir, filename)

	cleanup := func() {
		if _, err := os.Stat(filepath); err == nil {
			os.Remove(filepath)
		}
	}

	return filepath, cleanup
}

// TempFileB creates a temporary test file for benchmarks and returns cleanup function
func TempFileB(b *testing.B, suffix string) (string, func()) {
	b.Helper()

	// Create temp directory manually for benchmarks since b.TempDir() doesn't exist
	tempDir := os.TempDir()
	filename := TestFilePrefix + b.Name() + suffix
	filepath := filepath.Join(tempDir, filename)

	cleanup := func() {
		if _, err := os.Stat(filepath); err == nil {
			os.Remove(filepath)
		}
	}

	return filepath, cleanup
}

// NewBellStateCircuit creates a standard Bell state circuit for testing
func NewBellStateCircuit(t *testing.T) circuit.Circuit {
	t.Helper()

	b := builder.New(builder.Q(2), builder.C(2))
	b.H(0).CNOT(0, 1).Measure(0, 0).Measure(1, 1)

	c, err := b.BuildCircuit()
	require.NoError(t, err, "failed to build Bell state circuit")
	return c
}

// NewGroverCircuit creates a standard 2-qubit Grover circuit for testing
func NewGroverCircuit(t *testing.T) circuit.Circuit {
	t.Helper()

	b := builder.New(builder.Q(2), builder.C(2))

	// Initial superposition
	b.H(0).H(1)

	// Oracle marks |11⟩ by phase flip
	b.CZ(0, 1)

	// Diffusion operator
	b.H(0).H(1)
	b.X(0).X(1)
	b.CZ(0, 1)
	b.X(0).X(1)
	b.H(0).H(1)

	// Measurement
	b.Measure(0, 0).Measure(1, 1)

	c, err := b.BuildCircuit()
	require.NoError(t, err, "failed to build Grover circuit")
	return c
}

// AssertHistogramDistribution validates histogram results within tolerance
func AssertHistogramDistribution(t *testing.T, hist map[string]int, expected map[string]float64, totalShots int, tolerance float64) {
	t.Helper()

	for state, expectedProb := range expected {
		actualCount := hist[state]
		actualProb := float64(actualCount) / float64(totalShots)

		if expectedProb == 0 {
			require.Equal(t, 0, actualCount, "state %s should have 0 count", state)
		} else {
			require.InDelta(t, expectedProb, actualProb, tolerance,
				"state %s probability mismatch: expected %.3f, got %.3f",
				state, expectedProb, actualProb)
		}
	}
}

// RequireWithinTimeout runs a function with timeout and fails the test if it times out
func RequireWithinTimeout(t *testing.T, timeout time.Duration, fn func() error, msgAndArgs ...interface{}) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		require.NoError(t, err, msgAndArgs...)
	case <-ctx.Done():
		t.Fatalf("operation timed out after %v: %v", timeout, msgAndArgs)
	}
}

// SkipIfShort skips the test if running with -short flag
func SkipIfShort(t *testing.T, reason string) {
	t.Helper()
	if testing.Short() {
		t.Skipf("skipping test in short mode: %s", reason)
	}
}

// SkipIfCI skips the test if running in CI environment
func SkipIfCI(t *testing.T, reason string) {
	t.Helper()
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("skipping test in CI: %s", reason)
	}
}

// Parallel marks the test as safe to run in parallel
func Parallel(t *testing.T) {
	t.Helper()
	t.Parallel()
}
