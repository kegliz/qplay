package qsim

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/kegliz/qplay/qc/simulator"
	_ "github.com/kegliz/qplay/qc/simulator/itsu" // Import reference implementation
)

// Helper function to create a simple H-gate circuit
func createHadamardCircuit() circuit.Circuit {
	b := builder.New(builder.Q(1), builder.C(1))
	b.H(0)
	b.Measure(0, 0)
	c, _ := b.BuildCircuit()
	return c
}

// Helper function to create Bell state circuit
func createBellStateCircuit() circuit.Circuit {
	b := builder.New(builder.Q(2), builder.C(2))
	b.H(0)
	b.CNOT(0, 1)
	b.Measure(0, 0)
	b.Measure(1, 1)
	c, _ := b.BuildCircuit()
	return c
}

// Helper function to create superposition circuit
func createSuperpositionCircuit(qubits int) circuit.Circuit {
	b := builder.New(builder.Q(qubits), builder.C(qubits))

	for i := range qubits {
		b.H(i)
	}

	for i := 0; i < qubits; i++ {
		b.Measure(i, i)
	}

	c, _ := b.BuildCircuit()
	return c
}

func TestQSimRunner_BasicFunctionality(t *testing.T) {
	runner := NewQSimRunner()

	// Test simple circuit
	circ := createHadamardCircuit()
	result, err := runner.RunOnce(circ)
	if err != nil {
		t.Fatalf("Failed to run simple circuit: %v", err)
	}

	if result != "0" && result != "1" {
		t.Errorf("Expected result '0' or '1', got '%s'", result)
	}

	t.Logf("Hadamard circuit result: %s", result)
}

func TestQSimRunner_BellState(t *testing.T) {
	runner := NewQSimRunner()

	circ := createBellStateCircuit()

	// Run multiple times to check correlation
	results := make(map[string]int)
	runs := 1000

	for range runs {
		result, err := runner.RunOnce(circ)
		if err != nil {
			t.Fatalf("Failed to run Bell state circuit: %v", err)
		}
		results[result]++
	}

	// Check that we get mostly 00 and 11 (Bell state correlation)
	correlated := results["00"] + results["11"]

	correlationRatio := float64(correlated) / float64(runs)

	t.Logf("Bell state results: 00=%d, 01=%d, 10=%d, 11=%d",
		results["00"], results["01"], results["10"], results["11"])
	t.Logf("Correlation ratio: %.3f", correlationRatio)

	// Should have high correlation (>80%) due to Bell state
	if correlationRatio < 0.8 {
		t.Errorf("Expected high correlation (>0.8), got %.3f", correlationRatio)
	}
}

func TestQSimRunner_CompareWithItsubaki(t *testing.T) {
	qsimRunner := NewQSimRunner()
	itsubakiRunner, err := simulator.CreateRunner("itsu")
	if err != nil {
		t.Skipf("Itsubaki runner not available: %v", err)
	}

	testCases := []struct {
		name string
		circ circuit.Circuit
	}{
		{"Hadamard", createHadamardCircuit()},
		{"Bell State", createBellStateCircuit()},
		{"2-Qubit Superposition", createSuperpositionCircuit(2)},
		{"3-Qubit Superposition", createSuperpositionCircuit(3)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runs := 1000
			qsimResults := make(map[string]int)
			itsubakiResults := make(map[string]int)

			// Run with QSim
			for range runs {
				result, err := qsimRunner.RunOnce(tc.circ)
				if err != nil {
					t.Fatalf("QSim failed: %v", err)
				}
				qsimResults[result]++
			}

			// Run with Itsubaki
			for range runs {
				result, err := itsubakiRunner.RunOnce(tc.circ)
				if err != nil {
					t.Fatalf("Itsubaki failed: %v", err)
				}
				itsubakiResults[result]++
			}

			// Compare distributions
			t.Logf("QSim results: %v", qsimResults)
			t.Logf("Itsubaki results: %v", itsubakiResults)

			// Check that both simulators produce similar distributions
			for result, qsimCount := range qsimResults {
				qsimProb := float64(qsimCount) / float64(runs)
				itsubakiCount := itsubakiResults[result]
				itsubakiProb := float64(itsubakiCount) / float64(runs)

				diff := math.Abs(qsimProb - itsubakiProb)
				if diff > 0.1 { // Allow 10% difference due to randomness
					t.Errorf("Large difference for result %s: QSim=%.3f, Itsubaki=%.3f, diff=%.3f",
						result, qsimProb, itsubakiProb, diff)
				}
			}
		})
	}
}

func TestQSimRunner_ProbabilityValidation(t *testing.T) {
	runner := NewQSimRunner()

	// Test known quantum states with exact probabilities
	testCases := []struct {
		name     string
		builder  func() circuit.Circuit
		expected map[string]float64
	}{
		{
			name: "Single H gate",
			builder: func() circuit.Circuit {
				b := builder.New(builder.Q(1), builder.C(1))
				b.H(0)
				c, _ := b.BuildCircuit()
				return c
			},
			expected: map[string]float64{
				"0": 0.5,
				"1": 0.5,
			},
		},
		{
			name: "Two H gates",
			builder: func() circuit.Circuit {
				b := builder.New(builder.Q(2), builder.C(2))
				b.H(0)
				b.H(1)
				c, _ := b.BuildCircuit()
				return c
			},
			expected: map[string]float64{
				"00": 0.25,
				"01": 0.25,
				"10": 0.25,
				"11": 0.25,
			},
		},
		{
			name: "Bell state",
			builder: func() circuit.Circuit {
				b := builder.New(builder.Q(2), builder.C(2))
				b.H(0)
				b.CNOT(0, 1)
				c, _ := b.BuildCircuit()
				return c
			},
			expected: map[string]float64{
				"00": 0.5,
				"11": 0.5,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			circ := tc.builder()
			probs, err := runner.GetResultProbabilities(circ)
			if err != nil {
				t.Fatalf("Failed to get probabilities: %v", err)
			}

			t.Logf("Calculated probabilities: %v", probs)
			t.Logf("Expected probabilities: %v", tc.expected)

			// Check that probabilities match expectations
			for state, expectedProb := range tc.expected {
				actualProb, exists := probs[state]
				if !exists {
					t.Errorf("Expected state %s not found in results", state)
					continue
				}

				diff := math.Abs(actualProb - expectedProb)
				if diff > 1e-10 {
					t.Errorf("Probability mismatch for state %s: expected %.6f, got %.6f, diff=%.2e",
						state, expectedProb, actualProb, diff)
				}
			}

			// Check for unexpected states
			for state, prob := range probs {
				if _, expected := tc.expected[state]; !expected && prob > 1e-10 {
					t.Errorf("Unexpected state %s with probability %.6f", state, prob)
				}
			}
		})
	}
}

func TestQSimRunner_GateImplementations(t *testing.T) {
	runner := NewQSimRunner()

	testCases := []struct {
		name     string
		builder  func() circuit.Circuit
		expected map[string]float64
	}{
		{
			name: "X gate",
			builder: func() circuit.Circuit {
				b := builder.New(builder.Q(1), builder.C(1))
				b.X(0)
				c, _ := b.BuildCircuit()
				return c
			},
			expected: map[string]float64{"1": 1.0},
		},
		{
			name: "Y gate",
			builder: func() circuit.Circuit {
				b := builder.New(builder.Q(1), builder.C(1))
				b.Y(0)
				c, _ := b.BuildCircuit()
				return c
			},
			expected: map[string]float64{"1": 1.0},
		},
		{
			name: "Z gate (no effect on |0⟩)",
			builder: func() circuit.Circuit {
				b := builder.New(builder.Q(1), builder.C(1))
				b.Z(0)
				c, _ := b.BuildCircuit()
				return c
			},
			expected: map[string]float64{"0": 1.0},
		},
		{
			name: "S gate (no effect on |0⟩)",
			builder: func() circuit.Circuit {
				b := builder.New(builder.Q(1), builder.C(1))
				b.S(0)
				c, _ := b.BuildCircuit()
				return c
			},
			expected: map[string]float64{"0": 1.0},
		},
		{
			name: "SWAP gate",
			builder: func() circuit.Circuit {
				b := builder.New(builder.Q(2), builder.C(2))
				b.X(0)       // Set first qubit to |1⟩
				b.SWAP(0, 1) // Swap qubits
				c, _ := b.BuildCircuit()
				return c
			},
			expected: map[string]float64{"10": 1.0}, // |1⟩|0⟩ becomes |0⟩|1⟩
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			circ := tc.builder()
			probs, err := runner.GetResultProbabilities(circ)
			if err != nil {
				t.Fatalf("Failed to get probabilities: %v", err)
			}

			t.Logf("Gate %s probabilities: %v", tc.name, probs)

			for state, expectedProb := range tc.expected {
				actualProb, exists := probs[state]
				if !exists {
					t.Errorf("Expected state %s not found", state)
					continue
				}

				diff := math.Abs(actualProb - expectedProb)
				if diff > 1e-10 {
					t.Errorf("Probability mismatch for %s state %s: expected %.6f, got %.6f",
						tc.name, state, expectedProb, actualProb)
				}
			}
		})
	}
}

func TestQSimRunner_EnhancedInterfaces(t *testing.T) {
	runner := NewQSimRunner()

	// Test BackendProvider
	info := runner.GetBackendInfo()
	if info.Name != "QSim Quantum Simulator" {
		t.Errorf("Expected name 'QSim Quantum Simulator', got '%s'", info.Name)
	}

	// Test ConfigurableRunner
	err := runner.Configure(map[string]interface{}{
		"verbose": true,
		"seed":    int64(12345),
	})
	if err != nil {
		t.Errorf("Failed to configure runner: %v", err)
	}

	config := runner.GetConfiguration()
	if config["verbose"] != true {
		t.Errorf("Expected verbose=true, got %v", config["verbose"])
	}

	// Test MetricsCollector
	circ := createHadamardCircuit()
	_, err = runner.RunOnce(circ)
	if err != nil {
		t.Fatalf("Failed to run circuit: %v", err)
	}

	metrics := runner.GetMetrics()
	if metrics.TotalExecutions != 1 {
		t.Errorf("Expected 1 execution, got %d", metrics.TotalExecutions)
	}

	// Test ValidatingRunner
	err = runner.ValidateCircuit(circ)
	if err != nil {
		t.Errorf("Failed to validate valid circuit: %v", err)
	}

	gates := runner.GetSupportedGates()
	if len(gates) == 0 {
		t.Error("Expected non-empty supported gates list")
	}

	// Test BatchRunner
	results, err := runner.RunBatch(circ, 10)
	if err != nil {
		t.Errorf("Failed to run batch: %v", err)
	}
	if len(results) != 10 {
		t.Errorf("Expected 10 results, got %d", len(results))
	}

	// Test ContextualRunner
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = runner.RunOnceWithContext(ctx, circ)
	if err != nil {
		t.Errorf("Failed to run with context: %v", err)
	}
}

func TestQSimRunner_ErrorHandling(t *testing.T) {
	runner := NewQSimRunner()

	// Test invalid circuit with too many qubits
	b := builder.New(builder.Q(25), builder.C(25)) // Exceeds limit of 20
	invalidCirc, _ := b.BuildCircuit()

	err := runner.ValidateCircuit(invalidCirc)
	if err == nil {
		t.Error("Expected validation error for circuit with too many qubits")
	}

	// Test cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	circ := createHadamardCircuit()
	_, err = runner.RunOnceWithContext(ctx, circ)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}

	// Test invalid configuration
	err = runner.Configure(map[string]interface{}{
		"verbose": "not a boolean", // Invalid type
	})
	if err == nil {
		t.Error("Expected configuration error for invalid type")
	}
}

func BenchmarkQSimRunner_vs_Itsubaki(b *testing.B) {
	qsimRunner := NewQSimRunner()
	itsubakiRunner, err := simulator.CreateRunner("itsu")
	if err != nil {
		b.Skipf("Itsubaki runner not available: %v", err)
	}

	circ := createBellStateCircuit()

	b.Run("QSim", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := qsimRunner.RunOnce(circ)
			if err != nil {
				b.Fatalf("QSim failed: %v", err)
			}
		}
	})

	b.Run("Itsubaki", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := itsubakiRunner.RunOnce(circ)
			if err != nil {
				b.Fatalf("Itsubaki failed: %v", err)
			}
		}
	})
}
