package itsu

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"maps"
	"slices"

	"github.com/itsubaki/q"
	"github.com/kegliz/qplay/internal/logger"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/kegliz/qplay/qc/simulator"
	"github.com/rs/zerolog"
)

type ItsuOneShotRunner struct {
	log     logger.Logger
	config  map[string]interface{}
	mu      sync.RWMutex
	metrics ItsuMetrics
}

type ItsuMetrics struct {
	totalExecutions atomic.Int64
	successfulRuns  atomic.Int64
	failedRuns      atomic.Int64
	totalTime       atomic.Int64 // nanoseconds
	lastError       atomic.Value // string
	lastRunTime     atomic.Value // time.Time
}

// Supported gates for the Itsu backend
var supportedGates = []string{
	"H", "X", "Y", "S", "Z", "CNOT", "CZ", "SWAP", "TOFFOLI", "FREDKIN", "MEASURE",
}

func NewItsuOneShotRunner() *ItsuOneShotRunner {
	return &ItsuOneShotRunner{
		log: *logger.NewLogger(logger.LoggerOptions{
			Debug: false,
		}),
		config: make(map[string]any),
	}
}

// BackendProvider implementation
func (s *ItsuOneShotRunner) GetBackendInfo() simulator.BackendInfo {
	return simulator.BackendInfo{
		Name:        "Itsu Quantum Simulator",
		Version:     "v0.0.5",
		Description: "Go-based quantum circuit simulator using github.com/itsubaki/q",
		Vendor:      "itsubaki",
		Capabilities: map[string]bool{
			"context_support":    true,
			"batch_execution":    true,
			"circuit_validation": true,
			"metrics_collection": true,
			"configuration":      true,
			"reset":              true,
		},
		Metadata: map[string]string{
			"backend_type": "statevector_simulator",
			"language":     "go",
			"license":      "MIT",
		},
	}
}

// ConfigurableRunner implementation
func (s *ItsuOneShotRunner) Configure(options map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, value := range options {
		switch key {
		case "verbose":
			if verbose, ok := value.(bool); ok {
				s.SetVerbose(verbose)
				s.config[key] = value
			} else {
				return fmt.Errorf("invalid type for 'verbose' option: expected bool, got %T", value)
			}
		case "log_level":
			if level, ok := value.(string); ok {
				s.config[key] = value
				// TODO: Apply log level configuration based on level value
				_ = level // Acknowledge we're not using it yet
			} else {
				return fmt.Errorf("invalid type for 'log_level' option: expected string, got %T", value)
			}
		default:
			s.config[key] = value
		}
	}
	return nil
}

func (s *ItsuOneShotRunner) GetConfiguration() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config := make(map[string]any)
	maps.Copy(config, s.config)
	return config
}
func (s *ItsuOneShotRunner) SetVerbose(verbose bool) {
	if verbose {
		s.log.Logger = s.log.Logger.Level(zerolog.DebugLevel) // Log all messages if verbose
	} else {
		s.log.Logger = s.log.Logger.Level(zerolog.InfoLevel)
	}
}

func (s *ItsuOneShotRunner) RunOnce(c circuit.Circuit) (string, error) {
	start := time.Now()
	defer func() {
		s.metrics.totalExecutions.Add(1)
		s.metrics.totalTime.Add(int64(time.Since(start)))
		s.metrics.lastRunTime.Store(start)
	}()

	sim := q.New()
	result, err := runOnce(sim, c)

	if err != nil {
		s.metrics.failedRuns.Add(1)
		s.metrics.lastError.Store(err.Error())
	} else {
		s.metrics.successfulRuns.Add(1)
	}

	return result, err
}

// runOnce plays the circuit exactly one time on the provided simulator,
// returning the measured classical bit‑string.
func runOnce(sim *q.Q, c circuit.Circuit) (string, error) {
	qs := sim.ZeroWith(c.Qubits())
	//cbits := bytes.Repeat([]byte{'0'}, c.Clbits())
	cbits := make([]byte, c.Clbits())
	for i := range cbits {
		cbits[i] = '0' // Explicitly initialize to '0'
	}

	for i, op := range c.Operations() {
		// Check qubit indices are valid for the gate's operation before applying
		// (This is defensive programming; circuit/DAG validation should catch this)
		for _, qIndex := range op.Qubits {
			if qIndex < 0 || qIndex >= len(qs) {
				// Add operation index to error message
				return "", fmt.Errorf("itsu: invalid qubit index %d for gate %s (op %d) in runOnce", qIndex, op.G.Name(), i)
			}
		}
		if op.G.Name() == "MEASURE" && (op.Cbit < 0 || op.Cbit >= len(cbits)) {
			// Add operation index to error message
			return "", fmt.Errorf("itsu: invalid classical bit index %d for MEASURE (op %d) in runOnce", op.Cbit, i)
		}

		switch op.G.Name() {
		case "H":
			sim.H(qs[op.Qubits[0]])
		case "X":
			sim.X(qs[op.Qubits[0]])
		case "Y":
			sim.Y(qs[op.Qubits[0]])
		case "S":
			sim.S(qs[op.Qubits[0]])
		case "Z":
			sim.Z(qs[op.Qubits[0]])
		case "CNOT":
			sim.CNOT(qs[op.Qubits[0]], qs[op.Qubits[1]])
		case "CZ":
			sim.CZ(qs[op.Qubits[0]], qs[op.Qubits[1]])
		case "SWAP":
			sim.Swap(qs[op.Qubits[0]], qs[op.Qubits[1]])
		case "TOFFOLI":
			sim.Toffoli(qs[op.Qubits[0]], qs[op.Qubits[1]], qs[op.Qubits[2]])
		case "FREDKIN":
			ctrl, a, b := qs[op.Qubits[0]], qs[op.Qubits[1]], qs[op.Qubits[2]]
			// Standard decomposition: CNOT(b,a) Toffoli(ctrl,a,b) CNOT(b,a)
			sim.CNOT(b, a)
			sim.Toffoli(ctrl, a, b)
			sim.CNOT(b, a)
		case "MEASURE":
			m := sim.Measure(qs[op.Qubits[0]]) // collapses state & returns result
			if m.IsOne() {
				cbits[op.Cbit] = '1'
			} else {
				cbits[op.Cbit] = '0'
			}
		default:
			// Add operation index to error message
			return "", fmt.Errorf("itsu: unsupported gate %s (op %d) encountered in runOnce", op.G.Name(), i)
		}
	}
	// Return the final classical bit string (little-endian)
	return string(cbits), nil
}

// ResettableRunner implementation
func (s *ItsuOneShotRunner) Reset() {
	s.metrics.totalExecutions.Store(0)
	s.metrics.successfulRuns.Store(0)
	s.metrics.failedRuns.Store(0)
	s.metrics.totalTime.Store(0)
	s.metrics.lastError.Store("")
	s.metrics.lastRunTime.Store(time.Time{})
}

// MetricsCollector implementation
func (s *ItsuOneShotRunner) GetMetrics() simulator.ExecutionMetrics {
	totalExec := s.metrics.totalExecutions.Load()
	totalTimeNs := s.metrics.totalTime.Load()

	var avgTime time.Duration
	if totalExec > 0 {
		avgTime = time.Duration(totalTimeNs / totalExec)
	}

	lastErr, _ := s.metrics.lastError.Load().(string)
	lastRun, _ := s.metrics.lastRunTime.Load().(time.Time)

	return simulator.ExecutionMetrics{
		TotalExecutions: totalExec,
		SuccessfulRuns:  s.metrics.successfulRuns.Load(),
		FailedRuns:      s.metrics.failedRuns.Load(),
		AverageTime:     avgTime,
		TotalTime:       time.Duration(totalTimeNs),
		LastError:       lastErr,
		LastRunTime:     lastRun,
	}
}

func (s *ItsuOneShotRunner) ResetMetrics() {
	s.Reset()
}

// ValidatingRunner implementation
func (s *ItsuOneShotRunner) ValidateCircuit(c circuit.Circuit) error {
	for i, op := range c.Operations() {
		// Check if gate is supported
		supported := slices.Contains(supportedGates, op.G.Name())
		if !supported {
			return fmt.Errorf("itsu: unsupported gate %s at operation %d", op.G.Name(), i)
		}

		// Check qubit indices
		for _, qIndex := range op.Qubits {
			if qIndex < 0 || qIndex >= c.Qubits() {
				return fmt.Errorf("itsu: invalid qubit index %d for gate %s (op %d)", qIndex, op.G.Name(), i)
			}
		}

		// Check classical bit index for MEASURE
		if op.G.Name() == "MEASURE" && (op.Cbit < 0 || op.Cbit >= c.Clbits()) {
			return fmt.Errorf("itsu: invalid classical bit index %d for MEASURE (op %d)", op.Cbit, i)
		}
	}
	return nil
}

func (s *ItsuOneShotRunner) GetSupportedGates() []string {
	gates := make([]string, len(supportedGates))
	copy(gates, supportedGates)
	return gates
}

// ContextualRunner implementation
func (s *ItsuOneShotRunner) RunOnceWithContext(ctx context.Context, c circuit.Circuit) (string, error) {
	// Check for cancellation before starting
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	start := time.Now()
	defer func() {
		s.metrics.totalExecutions.Add(1)
		s.metrics.totalTime.Add(int64(time.Since(start)))
		s.metrics.lastRunTime.Store(start)
	}()

	// Create a channel to receive the result
	resultChan := make(chan struct {
		result string
		err    error
	}, 1)

	go func() {
		sim := q.New()
		result, err := runOnce(sim, c)
		resultChan <- struct {
			result string
			err    error
		}{result, err}
	}()

	select {
	case <-ctx.Done():
		s.metrics.failedRuns.Add(1)
		s.metrics.lastError.Store(ctx.Err().Error())
		return "", ctx.Err()
	case res := <-resultChan:
		if res.err != nil {
			s.metrics.failedRuns.Add(1)
			s.metrics.lastError.Store(res.err.Error())
		} else {
			s.metrics.successfulRuns.Add(1)
		}
		return res.result, res.err
	}
}

// BatchRunner implementation
func (s *ItsuOneShotRunner) RunBatch(c circuit.Circuit, shots int) ([]string, error) {
	if shots <= 0 {
		return nil, fmt.Errorf("shots must be positive, got %d", shots)
	}

	results := make([]string, shots)
	for i := range shots {
		result, err := s.RunOnce(c)
		if err != nil {
			return results[:i], fmt.Errorf("batch execution failed at shot %d: %w", i+1, err)
		}
		results[i] = result
	}
	return results, nil
}

// Register the Itsu runner with the plugin system
func init() {
	simulator.MustRegisterRunner("itsu", func() simulator.OneShotRunner {
		return NewItsuOneShotRunner()
	})

	// Also register with some aliases for convenience
	simulator.MustRegisterRunner("itsubaki", func() simulator.OneShotRunner {
		return NewItsuOneShotRunner()
	})

	simulator.MustRegisterRunner("default", func() simulator.OneShotRunner {
		return NewItsuOneShotRunner()
	})
}

// check that ItsuOneShotRunner implements the OneShotRunner interface
var _ simulator.OneShotRunner = (*ItsuOneShotRunner)(nil)
