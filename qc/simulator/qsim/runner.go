// Package qsim - Main runner implementation
package qsim

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kegliz/qplay/qc/circuit"
	"github.com/kegliz/qplay/qc/simulator"
)

// Supported gates for the QSim backend
var supportedGates = []string{
	"H", "X", "Y", "Z", "S", "CNOT", "CZ", "SWAP", "TOFFOLI", "FREDKIN", "MEASURE",
}

// OneShotRunner implementation
func (r *QSimRunner) RunOnce(c circuit.Circuit) (string, error) {
	return r.RunOnceWithContext(context.Background(), c)
}

// ContextualRunner implementation
func (r *QSimRunner) RunOnceWithContext(ctx context.Context, c circuit.Circuit) (string, error) {
	start := time.Now()
	r.metrics.totalExecutions.Add(1)
	r.metrics.lastRunTime.Store(start)

	defer func() {
		duration := time.Since(start)
		r.metrics.totalTime.Add(duration.Nanoseconds())
	}()

	// Check context cancellation
	select {
	case <-ctx.Done():
		r.metrics.failedRuns.Add(1)
		r.metrics.lastError.Store("context cancelled")
		return "", ctx.Err()
	default:
	}

	// Initialize quantum state
	state := NewQuantumState(c.Qubits(), c.Clbits())

	// Execute circuit operations
	for _, op := range c.Operations() {
		// Check context cancellation during execution
		select {
		case <-ctx.Done():
			r.metrics.failedRuns.Add(1)
			r.metrics.lastError.Store("context cancelled during execution")
			return "", ctx.Err()
		default:
		}

		if op.G.Name() == "MEASURE" {
			// Perform measurement
			if len(op.Qubits) != 1 {
				err := fmt.Errorf("measurement requires exactly one qubit, got %d", len(op.Qubits))
				r.metrics.failedRuns.Add(1)
				r.metrics.lastError.Store(err.Error())
				return "", err
			}

			qubit := op.Qubits[0]
			result := state.Measure(qubit)

			// Store classical bit if specified
			if op.Cbit >= 0 && op.Cbit < len(state.classicalBits) {
				state.classicalBits[op.Cbit] = result
			}
		} else {
			// Apply quantum gate
			if err := state.ApplyGate(op.G, op.Qubits); err != nil {
				r.metrics.failedRuns.Add(1)
				r.metrics.lastError.Store(err.Error())
				return "", fmt.Errorf("failed to apply gate %s: %w", op.G.Name(), err)
			}
		}
	}

	// Convert classical bits to result string
	result := r.formatResult(state.classicalBits)

	r.metrics.successfulRuns.Add(1)
	r.metrics.lastError.Store("")

	if r.verbose {
		fmt.Printf("QSim: Circuit executed successfully, result: %s\n", result)
	}

	return result, nil
}

// formatResult converts classical bits to string representation
func (r *QSimRunner) formatResult(bits []bool) string {
	if len(bits) == 0 {
		return "0" // Default result for circuits without measurements
	}

	var result strings.Builder
	for i := len(bits) - 1; i >= 0; i-- { // MSB first
		if bits[i] {
			result.WriteByte('1')
		} else {
			result.WriteByte('0')
		}
	}
	return result.String()
}

// BackendProvider implementation
func (r *QSimRunner) GetBackendInfo() simulator.BackendInfo {
	return simulator.BackendInfo{
		Name:        "QSim Quantum Simulator",
		Version:     "v1.0.0",
		Description: "Custom statevector-based quantum circuit simulator built from scratch",
		Vendor:      "qplay",
		Capabilities: map[string]bool{
			"context_support":    true,
			"batch_execution":    true,
			"circuit_validation": true,
			"metrics_collection": true,
			"configuration":      true,
			"reset":              true,
		},
		Metadata: map[string]string{
			"backend_type":   "statevector_simulator",
			"language":       "go",
			"license":        "MIT",
			"implementation": "from_scratch",
		},
	}
}

// ConfigurableRunner implementation
func (r *QSimRunner) SetVerbose(verbose bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.verbose = verbose
}

func (r *QSimRunner) Configure(options map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for key, value := range options {
		switch key {
		case "verbose":
			if verbose, ok := value.(bool); ok {
				r.verbose = verbose
				r.config[key] = value
			} else {
				return fmt.Errorf("invalid type for 'verbose' option: expected bool, got %T", value)
			}
		case "log_level":
			if _, ok := value.(string); ok {
				r.config[key] = value
			} else {
				return fmt.Errorf("invalid type for 'log_level' option: expected string, got %T", value)
			}
		case "seed":
			if _, ok := value.(int64); ok {
				r.config[key] = value
				// TODO: Set random seed for reproducible results
			} else {
				return fmt.Errorf("invalid type for 'seed' option: expected int64, got %T", value)
			}
		default:
			r.config[key] = value
		}
	}
	return nil
}

func (r *QSimRunner) GetConfiguration() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range r.config {
		result[k] = v
	}
	return result
}

// ResettableRunner implementation
func (r *QSimRunner) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Reset metrics
	r.metrics.totalExecutions.Store(0)
	r.metrics.successfulRuns.Store(0)
	r.metrics.failedRuns.Store(0)
	r.metrics.totalTime.Store(0)
	r.metrics.lastError.Store("")
	r.metrics.lastRunTime.Store(time.Time{})
}

// MetricsCollector implementation
func (r *QSimRunner) GetMetrics() simulator.ExecutionMetrics {
	totalExec := r.metrics.totalExecutions.Load()
	successRuns := r.metrics.successfulRuns.Load()
	failedRuns := r.metrics.failedRuns.Load()
	totalTimeNs := r.metrics.totalTime.Load()

	var avgTime time.Duration
	if totalExec > 0 {
		avgTime = time.Duration(totalTimeNs / totalExec)
	}

	lastError := ""
	if err := r.metrics.lastError.Load(); err != nil {
		lastError = err.(string)
	}

	lastRunTime := time.Time{}
	if t := r.metrics.lastRunTime.Load(); t != nil {
		lastRunTime = t.(time.Time)
	}

	return simulator.ExecutionMetrics{
		TotalExecutions: totalExec,
		SuccessfulRuns:  successRuns,
		FailedRuns:      failedRuns,
		AverageTime:     avgTime,
		TotalTime:       time.Duration(totalTimeNs),
		LastError:       lastError,
		LastRunTime:     lastRunTime,
	}
}

func (r *QSimRunner) ResetMetrics() {
	r.metrics.totalExecutions.Store(0)
	r.metrics.successfulRuns.Store(0)
	r.metrics.failedRuns.Store(0)
	r.metrics.totalTime.Store(0)
	r.metrics.lastError.Store("")
	r.metrics.lastRunTime.Store(time.Time{})
}

// ValidatingRunner implementation
func (r *QSimRunner) ValidateCircuit(c circuit.Circuit) error {
	if c.Qubits() > 20 { // Reasonable limit for demo
		return fmt.Errorf("circuit has too many qubits: %d (max 20)", c.Qubits())
	}

	if c.Depth() > 1000 { // Reasonable depth limit
		return fmt.Errorf("circuit is too deep: %d layers (max 1000)", c.Depth())
	}

	// Check all gates are supported
	for _, op := range c.Operations() {
		supported := false
		for _, supportedGate := range supportedGates {
			if op.G.Name() == supportedGate {
				supported = true
				break
			}
		}

		if !supported {
			return fmt.Errorf("unsupported gate: %s", op.G.Name())
		}

		// Validate qubit indices
		for _, qubit := range op.Qubits {
			if qubit < 0 || qubit >= c.Qubits() {
				return fmt.Errorf("invalid qubit index %d for %d-qubit circuit", qubit, c.Qubits())
			}
		}

		// Validate classical bit index
		if op.Cbit >= c.Clbits() {
			return fmt.Errorf("invalid classical bit index %d for %d-clbit circuit", op.Cbit, c.Clbits())
		}
	}

	return nil
}

func (r *QSimRunner) GetSupportedGates() []string {
	result := make([]string, len(supportedGates))
	copy(result, supportedGates)
	return result
}

// BatchRunner implementation
func (r *QSimRunner) RunBatch(c circuit.Circuit, shots int) ([]string, error) {
	if shots <= 0 {
		return nil, fmt.Errorf("shots must be positive, got %d", shots)
	}

	results := make([]string, shots)

	for i := 0; i < shots; i++ {
		result, err := r.RunOnce(c)
		if err != nil {
			return nil, fmt.Errorf("shot %d failed: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}

// GetResultProbabilities analyzes a circuit and returns theoretical probabilities
// This is useful for validation against known quantum states
func (r *QSimRunner) GetResultProbabilities(c circuit.Circuit) (map[string]float64, error) {
	// Create a copy of the state without measurements
	state := NewQuantumState(c.Qubits(), c.Clbits())

	// Apply all non-measurement operations
	for _, op := range c.Operations() {
		if op.G.Name() != "MEASURE" {
			if err := state.ApplyGate(op.G, op.Qubits); err != nil {
				return nil, fmt.Errorf("failed to apply gate %s: %w", op.G.Name(), err)
			}
		}
	}

	// Get probabilities for each computational basis state
	probs := state.GetProbabilities()
	result := make(map[string]float64)

	// Convert to string representation
	for i, prob := range probs {
		if prob > 1e-10 { // Only include non-zero probabilities
			bitString := fmt.Sprintf("%0*b", state.numQubits, i)
			result[bitString] = prob
		}
	}

	return result, nil
}

// Factory function for the plugin system
func init() {
	// Register the QSim runner with the plugin system
	simulator.MustRegisterRunner("qsim", func() simulator.OneShotRunner {
		return NewQSimRunner()
	})
}
