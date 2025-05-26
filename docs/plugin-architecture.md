# Quantum Backend Plugin Architecture

The QC simulator package now supports a flexible plugin architecture that allows different quantum backends to be registered and used interchangeably. This design enables easy extension with new quantum simulation engines while maintaining backward compatibility.

## Overview

The plugin system consists of several key components:

1. **Runner Registry**: Manages registration and creation of quantum backend runners
2. **Enhanced Interfaces**: Optional capabilities that runners can implement
3. **Plugin Registration**: Automatic registration system using Go's `init()` functions
4. **Backward Compatibility**: Existing code continues to work without changes

## Core Interfaces

### OneShotRunner (Base Interface)
```go
type OneShotRunner interface {
    RunOnce(circuit.Circuit) (string, error)
}
```

### Enhanced Interfaces

**BackendProvider**: Provides metadata about the backend
```go
type BackendProvider interface {
    GetBackendInfo() BackendInfo
}
```

**ContextualRunner**: Supports context-based execution with cancellation
```go
type ContextualRunner interface {
    RunOnceWithContext(ctx context.Context, c circuit.Circuit) (string, error)
}
```

**ConfigurableRunner**: Allows runtime configuration
```go
type ConfigurableRunner interface {
    SetVerbose(verbose bool)
    Configure(options map[string]interface{}) error
    GetConfiguration() map[string]interface{}
}
```

**MetricsCollector**: Provides execution statistics
```go
type MetricsCollector interface {
    GetMetrics() ExecutionMetrics
    ResetMetrics()
}
```

**ValidatingRunner**: Can validate circuits before execution
```go
type ValidatingRunner interface {
    ValidateCircuit(c circuit.Circuit) error
    GetSupportedGates() []string
}
```

**BatchRunner**: Supports efficient batch execution
```go
type BatchRunner interface {
    RunBatch(c circuit.Circuit, shots int) ([]string, error)
}
```

## Using the Plugin System

### Basic Usage

```go
// Create a simulator using a named runner
sim, err := simulator.NewSimulatorWithDefaults("itsu")
if err != nil {
    log.Fatal(err)
}

// Or with custom options
sim, err := simulator.NewSimulatorWithRunner("itsu", simulator.SimulatorOptions{
    Shots:   2048,
    Workers: 4,
})
```

### Discovering Available Runners

```go
// List all registered runners
runners := simulator.ListRunners()
fmt.Println("Available runners:", runners)

// Get information about a specific runner
runner, err := simulator.CreateRunner("itsu")
if err != nil {
    log.Fatal(err)
}

if info := simulator.GetBackendInfo(runner); info != nil {
    fmt.Printf("Backend: %s v%s\n", info.Name, info.Version)
    fmt.Printf("Description: %s\n", info.Description)
}
```

### Checking Capabilities

```go
runner, _ := simulator.CreateRunner("itsu")

// Check what interfaces the runner supports
if simulator.SupportsContext(runner) {
    fmt.Println("Runner supports context-based execution")
}

if simulator.SupportsMetrics(runner) {
    fmt.Println("Runner provides execution metrics")
}

if simulator.SupportsBatch(runner) {
    fmt.Println("Runner supports batch execution")
}
```

### Using Enhanced Features

```go
runner, _ := simulator.CreateRunner("itsu")

// Configuration
if configurable, ok := runner.(simulator.ConfigurableRunner); ok {
    err := configurable.Configure(map[string]interface{}{
        "verbose": true,
        "timeout": 30,
    })
}

// Context-based execution
if contextual, ok := runner.(simulator.ContextualRunner); ok {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    result, err := contextual.RunOnceWithContext(ctx, circuit)
}

// Batch execution
if batch, ok := runner.(simulator.BatchRunner); ok {
    results, err := batch.RunBatch(circuit, 1000)
}

// Metrics
if metrics, ok := runner.(simulator.MetricsCollector); ok {
    stats := metrics.GetMetrics()
    fmt.Printf("Total executions: %d\n", stats.TotalExecutions)
    fmt.Printf("Average time: %v\n", stats.AverageTime)
}
```

## Creating a Custom Backend

To create a custom quantum backend, implement the `OneShotRunner` interface and optionally the enhanced interfaces:

```go
package mybackend

import (
    "github.com/kegliz/qplay/qc/simulator"
    "github.com/kegliz/qplay/qc/circuit"
)

type MyRunner struct {
    // Your backend implementation
}

func NewMyRunner() *MyRunner {
    return &MyRunner{}
}

func (r *MyRunner) RunOnce(c circuit.Circuit) (string, error) {
    // Implement your quantum simulation logic here
    return "00101", nil
}

// Optionally implement enhanced interfaces
func (r *MyRunner) GetBackendInfo() simulator.BackendInfo {
    return simulator.BackendInfo{
        Name:        "My Custom Backend",
        Version:     "v1.0.0",
        Description: "Custom quantum simulator",
        Vendor:      "my-company",
        Capabilities: map[string]bool{
            "context_support": true,
            "batch_execution": false,
        },
    }
}

// Register your backend
func init() {
    simulator.MustRegisterRunner("my-backend", func() simulator.OneShotRunner {
        return NewMyRunner()
    })
}
```

## Built-in Backends

### Itsu Backend
The default backend based on `github.com/itsubaki/q`:

- **Name**: "itsu", "itsubaki", "default"
- **Features**: Full feature support including context, batch, metrics, validation
- **Supported Gates**: H, X, S, CNOT, CZ, SWAP, TOFFOLI, FREDKIN, MEASURE

## Command Line Demo

A demonstration utility is available at `cmd/plugin-demo/`:

```bash
# List all available runners
go run cmd/plugin-demo/main.go list

# Show detailed information about a runner
go run cmd/plugin-demo/main.go info itsu

# Run an example circuit
go run cmd/plugin-demo/main.go run itsu

# Benchmark performance
go run cmd/plugin-demo/main.go benchmark itsu
```

## Migration Guide

### Existing Code
No changes needed! Existing code continues to work:

```go
// This still works exactly as before
sim := simulator.NewSimulator(simulator.SimulatorOptions{
    Shots:  1024,
    Runner: itsu.NewItsuOneShotRunner(),
})
```

### New Plugin-Based Code
Take advantage of the new plugin system:

```go
// New plugin-based approach
sim, err := simulator.NewSimulatorWithDefaults("itsu")
if err != nil {
    log.Fatal(err)
}
```

## Testing

The plugin system includes comprehensive tests covering:

- Runner registration and discovery
- Enhanced interface functionality
- Capability checking
- Error handling
- Thread safety

Run tests with:
```bash
go test ./qc/simulator/...
```

## Thread Safety

The plugin registry is thread-safe and can be safely used from multiple goroutines. Runner factories are called on-demand and should return new instances for each call.

## Best Practices

1. **Register in init()**: Use `init()` functions to register runners automatically
2. **Check Capabilities**: Always check if a runner supports optional interfaces before using them
3. **Handle Errors**: Plugin creation can fail, always check errors
4. **Use Contexts**: For long-running operations, use context-based execution when available
5. **Collect Metrics**: Use metrics to monitor performance and debug issues
