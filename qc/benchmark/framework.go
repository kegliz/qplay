// Package benchmark provides a standardized benchmarking framework for quantum backend plugins
package benchmark

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/kegliz/qplay/qc/circuit"
	"github.com/kegliz/qplay/qc/simulator"
	"github.com/kegliz/qplay/qc/testutil"
)

// ResourceLimits defines limits for benchmark execution
type ResourceLimits struct {
	MaxMemoryMB     int64         // Maximum memory usage in MB
	MaxDuration     time.Duration // Maximum duration per benchmark
	MaxCircuitDepth int           // Maximum circuit depth
	MaxQubits       int           // Maximum number of qubits
}

// DefaultResourceLimits provides safe defaults for benchmark execution
var DefaultResourceLimits = ResourceLimits{
	MaxMemoryMB:     500, // 500MB memory limit
	MaxDuration:     30 * time.Second,
	MaxCircuitDepth: 20, // Reasonable circuit depth
	MaxQubits:       5,  // Conservative qubit limit
}

// BenchmarkScenario represents different types of benchmark tests
type BenchmarkScenario string

const (
	SerialExecution   BenchmarkScenario = "serial"
	ParallelExecution BenchmarkScenario = "parallel"
	BatchExecution    BenchmarkScenario = "batch"
	ContextExecution  BenchmarkScenario = "context"
	MetricsCollection BenchmarkScenario = "metrics"
)

// BenchmarkConfig holds configuration for benchmark execution
type BenchmarkConfig struct {
	CircuitType CircuitType
	Scenario    BenchmarkScenario
	Config      testutil.TestConfig
	RunnerName  string
	Limits      ResourceLimits // Resource limits for safe execution
}

// ResourceUsage tracks resource consumption during benchmarks
type ResourceUsage struct {
	StartMemory   uint64        `json:"start_memory"`
	PeakMemory    uint64        `json:"peak_memory"`
	EndMemory     uint64        `json:"end_memory"`
	MemoryDelta   int64         `json:"memory_delta"`
	GCCount       uint32        `json:"gc_count"`
	Duration      time.Duration `json:"duration"`
	CircuitDepth  int           `json:"circuit_depth"`
	CircuitQubits int           `json:"circuit_qubits"`
}

// BenchmarkResult contains the results and metadata from a benchmark run
type BenchmarkResult struct {
	RunnerName     string                      `json:"runner_name"`
	CircuitType    CircuitType                 `json:"circuit_type"`
	Scenario       BenchmarkScenario           `json:"scenario"`
	Success        bool                        `json:"success"`
	Error          string                      `json:"error,omitempty"`
	Duration       time.Duration               `json:"duration"`
	AllocsPerOp    int64                       `json:"allocs_per_op"`
	BytesPerOp     int64                       `json:"bytes_per_op"`
	BackendInfo    *simulator.BackendInfo      `json:"backend_info,omitempty"`
	Metrics        *simulator.ExecutionMetrics `json:"metrics,omitempty"`
	ResourceUsage  ResourceUsage               `json:"resource_usage"`            // NEW: Resource tracking
	LimitsExceeded []string                    `json:"limits_exceeded,omitempty"` // NEW: Limit violations
}

// PluginBenchmarkSuite provides comprehensive benchmarking for all registered quantum backends
type PluginBenchmarkSuite struct {
	runners   []string
	circuits  []CircuitType
	scenarios []BenchmarkScenario
	config    testutil.TestConfig
	limits    ResourceLimits // NEW: Resource limits
}

// NewPluginBenchmarkSuite creates a new benchmark suite with default configuration
func NewPluginBenchmarkSuite() *PluginBenchmarkSuite {
	return &PluginBenchmarkSuite{
		runners:   simulator.ListRunners(),
		circuits:  []CircuitType{SimpleCircuit, EntanglementCircuit, SuperpositionCircuit, MixedGatesCircuit},
		scenarios: []BenchmarkScenario{SerialExecution, ParallelExecution, BatchExecution, ContextExecution},
		config:    testutil.QuickTestConfig, // Use quick config for benchmarking
		limits:    DefaultResourceLimits,    // NEW: Default resource limits
	}
}

// WithRunners configures which runners to benchmark
func (s *PluginBenchmarkSuite) WithRunners(runners ...string) *PluginBenchmarkSuite {
	s.runners = runners
	return s
}

// WithCircuits configures which circuit types to test
func (s *PluginBenchmarkSuite) WithCircuits(circuits ...CircuitType) *PluginBenchmarkSuite {
	s.circuits = circuits
	return s
}

// WithScenarios configures which scenarios to test
func (s *PluginBenchmarkSuite) WithScenarios(scenarios ...BenchmarkScenario) *PluginBenchmarkSuite {
	s.scenarios = scenarios
	return s
}

// WithConfig sets the test configuration
func (s *PluginBenchmarkSuite) WithConfig(config testutil.TestConfig) *PluginBenchmarkSuite {
	s.config = config
	return s
}

// WithLimits sets the resource limits
func (s *PluginBenchmarkSuite) WithLimits(limits ResourceLimits) *PluginBenchmarkSuite {
	s.limits = limits
	return s
}

// validateCircuitComplexity checks if a circuit exceeds complexity limits
func validateCircuitComplexity(circ circuit.Circuit, limits ResourceLimits) []string {
	var violations []string

	if circ.Qubits() > limits.MaxQubits {
		violations = append(violations, fmt.Sprintf("circuit has %d qubits, limit is %d", circ.Qubits(), limits.MaxQubits))
	}

	// Estimate circuit depth by counting gate layers
	depth := estimateCircuitDepth(circ)
	if depth > limits.MaxCircuitDepth {
		violations = append(violations, fmt.Sprintf("circuit depth %d exceeds limit %d", depth, limits.MaxCircuitDepth))
	}

	return violations
}

// estimateCircuitDepth provides a rough estimate of circuit depth
func estimateCircuitDepth(circ circuit.Circuit) int {
	// Use the built-in depth calculation if available
	if depth := circ.Depth(); depth > 0 {
		return depth
	}

	// Fallback: simple heuristic based on operations
	ops := circ.Operations()
	return len(ops) / max(1, circ.Qubits()) // Rough estimate
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getMemoryUsage returns current memory statistics
func getMemoryUsage() (uint64, uint32) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc, m.NumGC
}

// checkMemoryLimit verifies that current memory usage is within limits
func checkMemoryLimit(maxMemoryMB int64) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	currentMemoryMB := int64(m.Alloc) / (1024 * 1024)
	if currentMemoryMB > maxMemoryMB {
		return fmt.Errorf("current memory usage %dMB exceeds limit %dMB", currentMemoryMB, maxMemoryMB)
	}
	return nil
}

// RunSingleBenchmark executes a single benchmark configuration with resource monitoring
func RunSingleBenchmark(b *testing.B, config BenchmarkConfig) BenchmarkResult {
	result := BenchmarkResult{
		RunnerName:  config.RunnerName,
		CircuitType: config.CircuitType,
		Scenario:    config.Scenario,
	}

	// Initialize resource tracking
	startMem, startGC := getMemoryUsage()
	result.ResourceUsage.StartMemory = startMem

	// Force initial GC to get clean baseline
	runtime.GC()
	debug.FreeOSMemory()

	// Create the runner
	runner, err := simulator.CreateRunner(config.RunnerName)
	if err != nil {
		result.Error = fmt.Sprintf("failed to create runner: %v", err)
		return result
	}

	// Get backend info if available
	if info := simulator.GetBackendInfo(runner); info != nil {
		result.BackendInfo = info
	}

	// Configure the runner if it supports configuration
	if configurable, ok := runner.(simulator.ConfigurableRunner); ok {
		configurable.SetVerbose(false) // Disable verbose for benchmarking
	}

	// Build the circuit
	circuitBuilder := StandardCircuits[config.CircuitType]

	// Apply resource limits to circuit building
	qubits := min(config.Config.Qubits, config.Limits.MaxQubits)
	build := circuitBuilder(qubits)
	circ, err := build.BuildCircuit()
	if err != nil {
		result.Error = fmt.Sprintf("failed to build circuit: %v", err)
		return result
	}

	// Validate circuit complexity
	if violations := validateCircuitComplexity(circ, config.Limits); len(violations) > 0 {
		result.LimitsExceeded = violations
		result.Error = fmt.Sprintf("circuit exceeds resource limits: %v", violations)
		return result
	}

	// Record circuit characteristics
	result.ResourceUsage.CircuitQubits = circ.Qubits()
	result.ResourceUsage.CircuitDepth = estimateCircuitDepth(circ)

	b.ReportAllocs()
	b.ResetTimer()

	// Execute the benchmark based on scenario
	start := time.Now()
	err = runBenchmarkScenario(b, runner, circ, config)
	result.Duration = time.Since(start)

	// Record memory usage after execution
	endMem, endGC := getMemoryUsage()
	result.ResourceUsage.EndMemory = endMem
	result.ResourceUsage.GCCount = endGC - startGC
	result.ResourceUsage.MemoryDelta = int64(endMem - startMem)

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Success = true
	}

	// Collect metrics if supported
	if metrics, ok := runner.(simulator.MetricsCollector); ok {
		execMetrics := metrics.GetMetrics()
		result.Metrics = &execMetrics
	}

	return result
}

// runBenchmarkScenario executes the appropriate benchmark based on the scenario
func runBenchmarkScenario(b *testing.B, runner simulator.OneShotRunner, circ circuit.Circuit, config BenchmarkConfig) error {
	switch config.Scenario {
	case SerialExecution:
		return runSerialBenchmark(b, runner, circ, config)
	case ParallelExecution:
		return runParallelBenchmark(b, runner, circ, config)
	case BatchExecution:
		return runBatchBenchmark(b, runner, circ, config)
	case ContextExecution:
		return runContextBenchmark(b, runner, circ, config)
	case MetricsCollection:
		return runMetricsBenchmark(b, runner, circ, config)
	default:
		return fmt.Errorf("unknown scenario: %s", config.Scenario)
	}
}

// runSerialBenchmark tests basic serial execution with timeout protection
func runSerialBenchmark(b *testing.B, runner simulator.OneShotRunner, circ circuit.Circuit, config BenchmarkConfig) error {
	for i := 0; i < b.N; i++ {
		// Check memory usage before each iteration
		if err := checkMemoryLimit(config.Limits.MaxMemoryMB); err != nil {
			return fmt.Errorf("memory limit exceeded: %v", err)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), config.Limits.MaxDuration)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			sim := simulator.NewSimulator(simulator.SimulatorOptions{
				Shots:  config.Config.Shots,
				Runner: runner,
			})

			_, err := sim.RunSerial(circ)
			done <- err
		}()

		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("serial run failed: %v", err)
			}
		case <-ctx.Done():
			return fmt.Errorf("serial run timed out after %v", config.Limits.MaxDuration)
		}
	}
	return nil
}

// runParallelBenchmark tests parallel execution if supported with resource monitoring
func runParallelBenchmark(b *testing.B, runner simulator.OneShotRunner, circ circuit.Circuit, config BenchmarkConfig) error {
	for i := 0; i < b.N; i++ {
		// Check memory usage before each iteration
		if err := checkMemoryLimit(config.Limits.MaxMemoryMB); err != nil {
			return fmt.Errorf("memory limit exceeded: %v", err)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), config.Limits.MaxDuration)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			sim := simulator.NewSimulator(simulator.SimulatorOptions{
				Shots:   config.Config.Shots,
				Workers: config.Config.Workers,
				Runner:  runner,
			})

			_, err := sim.RunParallelChan(circ)
			done <- err
		}()

		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("parallel run failed: %v", err)
			}
		case <-ctx.Done():
			return fmt.Errorf("parallel run timed out after %v", config.Limits.MaxDuration)
		}
	}
	return nil
}

// runBatchBenchmark tests batch execution if supported with resource limits
func runBatchBenchmark(b *testing.B, runner simulator.OneShotRunner, circ circuit.Circuit, config BenchmarkConfig) error {
	if batchRunner, ok := runner.(simulator.BatchRunner); ok {
		// Limit batch size to prevent memory issues
		maxBatchSize := min(config.Config.Shots, 1000) // Cap at 1000 shots

		for i := 0; i < b.N; i++ {
			// Check memory usage before each iteration
			if err := checkMemoryLimit(config.Limits.MaxMemoryMB); err != nil {
				return fmt.Errorf("memory limit exceeded: %v", err)
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), config.Limits.MaxDuration)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				_, err := batchRunner.RunBatch(circ, maxBatchSize)
				done <- err
			}()

			select {
			case err := <-done:
				if err != nil {
					return fmt.Errorf("batch run failed: %v", err)
				}
			case <-ctx.Done():
				return fmt.Errorf("batch run timed out after %v", config.Limits.MaxDuration)
			}
		}
	} else {
		b.Skip("Runner does not support batch execution")
	}
	return nil
}

// runContextBenchmark tests context-based execution if supported
func runContextBenchmark(b *testing.B, runner simulator.OneShotRunner, circ circuit.Circuit, config BenchmarkConfig) error {
	if contextRunner, ok := runner.(simulator.ContextualRunner); ok {
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), config.Config.Timeout)
			_, err := contextRunner.RunOnceWithContext(ctx, circ)
			cancel()

			if err != nil {
				return fmt.Errorf("context run failed: %v", err)
			}
		}
	} else {
		b.Skip("Runner does not support context execution")
	}
	return nil
}

// runMetricsBenchmark tests with metrics collection enabled and resource monitoring
func runMetricsBenchmark(b *testing.B, runner simulator.OneShotRunner, circ circuit.Circuit, config BenchmarkConfig) error {
	if metricsRunner, ok := runner.(simulator.MetricsCollector); ok {
		metricsRunner.ResetMetrics()

		for i := 0; i < b.N; i++ {
			// Check memory usage before each iteration
			if err := checkMemoryLimit(config.Limits.MaxMemoryMB); err != nil {
				return fmt.Errorf("memory limit exceeded: %v", err)
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), config.Limits.MaxDuration)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				_, err := runner.RunOnce(circ)
				done <- err
			}()

			select {
			case err := <-done:
				if err != nil {
					return fmt.Errorf("metrics run failed: %v", err)
				}
			case <-ctx.Done():
				return fmt.Errorf("metrics run timed out after %v", config.Limits.MaxDuration)
			}
		}
	} else {
		b.Skip("Runner does not support metrics collection")
	}
	return nil
}

func GetBenchmarkName(runnerName string, circuitType CircuitType, scenario BenchmarkScenario) string {
	return fmt.Sprintf("%s_%s_%s", runnerName, circuitType, scenario)
}
