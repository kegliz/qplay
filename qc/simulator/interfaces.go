package simulator

import (
	"context"
	"time"

	"github.com/kegliz/qplay/qc/circuit"
)

// BackendInfo provides metadata about a quantum backend runner.
type BackendInfo struct {
	Name         string            `json:"name"`         // Human-readable name
	Version      string            `json:"version"`      // Backend version
	Description  string            `json:"description"`  // Description of the backend
	Vendor       string            `json:"vendor"`       // Vendor/author
	Capabilities map[string]bool   `json:"capabilities"` // Supported features
	Metadata     map[string]string `json:"metadata"`     // Additional metadata
}

// ExecutionMetrics contains performance and execution statistics.
type ExecutionMetrics struct {
	TotalExecutions int64         `json:"total_executions"`
	SuccessfulRuns  int64         `json:"successful_runs"`
	FailedRuns      int64         `json:"failed_runs"`
	AverageTime     time.Duration `json:"average_time"`
	TotalTime       time.Duration `json:"total_time"`
	LastError       string        `json:"last_error,omitempty"`
	LastRunTime     time.Time     `json:"last_run_time"`
}

// Enhanced interfaces for plugin capabilities

// BackendProvider provides information about the quantum backend.
type BackendProvider interface {
	// GetBackendInfo returns metadata about this backend implementation.
	GetBackendInfo() BackendInfo
}

// ContextualRunner supports context-based execution with cancellation and timeouts.
type ContextualRunner interface {
	// RunOnceWithContext executes a circuit with context support.
	RunOnceWithContext(ctx context.Context, c circuit.Circuit) (string, error)
}

// ConfigurableRunner allows runtime configuration of the runner.
type ConfigurableRunner interface {
	// SetVerbose enables or disables verbose logging.
	SetVerbose(verbose bool)

	// Configure applies configuration options to the runner.
	Configure(options map[string]interface{}) error

	// GetConfiguration returns current configuration.
	GetConfiguration() map[string]interface{}
}

// ResettableRunner allows resetting internal state.
type ResettableRunner interface {
	// Reset clears any internal state and counters.
	Reset()
}

// MetricsCollector provides execution metrics and statistics.
type MetricsCollector interface {
	// GetMetrics returns current execution metrics.
	GetMetrics() ExecutionMetrics

	// ResetMetrics clears all collected metrics.
	ResetMetrics()
}

// ValidatingRunner can validate circuits before execution.
type ValidatingRunner interface {
	// ValidateCircuit checks if the circuit can be executed by this runner.
	ValidateCircuit(c circuit.Circuit) error

	// GetSupportedGates returns the list of supported gate names.
	GetSupportedGates() []string
}

// BatchRunner supports batch execution for better performance.
type BatchRunner interface {
	// RunBatch executes multiple shots efficiently.
	RunBatch(c circuit.Circuit, shots int) ([]string, error)
}

// Enhanced OneShotRunner interface with optional capabilities
// The base OneShotRunner interface remains unchanged for backward compatibility.

// FullFeaturedRunner combines all optional interfaces.
// Implementations can choose which interfaces to implement based on their capabilities.
type FullFeaturedRunner interface {
	OneShotRunner
	BackendProvider
	ContextualRunner
	ConfigurableRunner
	ResettableRunner
	MetricsCollector
	ValidatingRunner
	BatchRunner
}

// Helper functions to check runner capabilities

// SupportsContext checks if a runner supports context-based execution.
func SupportsContext(runner OneShotRunner) bool {
	_, ok := runner.(ContextualRunner)
	return ok
}

// SupportsConfiguration checks if a runner supports runtime configuration.
func SupportsConfiguration(runner OneShotRunner) bool {
	_, ok := runner.(ConfigurableRunner)
	return ok
}

// SupportsMetrics checks if a runner provides execution metrics.
func SupportsMetrics(runner OneShotRunner) bool {
	_, ok := runner.(MetricsCollector)
	return ok
}

// SupportsValidation checks if a runner can validate circuits.
func SupportsValidation(runner OneShotRunner) bool {
	_, ok := runner.(ValidatingRunner)
	return ok
}

// SupportsBatch checks if a runner supports batch execution.
func SupportsBatch(runner OneShotRunner) bool {
	_, ok := runner.(BatchRunner)
	return ok
}

// SupportsBackendInfo checks if a runner provides backend information.
func SupportsBackendInfo(runner OneShotRunner) bool {
	_, ok := runner.(BackendProvider)
	return ok
}

// GetBackendInfo safely gets backend information if available.
func GetBackendInfo(runner OneShotRunner) *BackendInfo {
	if provider, ok := runner.(BackendProvider); ok {
		info := provider.GetBackendInfo()
		return &info
	}
	return nil
}
