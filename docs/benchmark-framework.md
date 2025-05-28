# Quantum Benchmarking Framework

A comprehensive, production-ready benchmarking framework for quantum computing backends with resource management, stress testing, and CI/CD integration.

## Features

### 🚀 Core Capabilities
- **Plugin-Level Architecture**: Seamlessly integrates with the quantum backend plugin system
- **Resource Management**: Memory limits, timeouts, and circuit complexity validation
- **Multiple Test Scenarios**: Serial, parallel, batch, context, and metrics collection
- **Standard Circuit Library**: Simple, entanglement, superposition, and mixed gate circuits
- **Comprehensive Reporting**: JSON and console output with detailed metrics

### 🛡️ Safety & Resource Management
- **Memory Monitoring**: Real-time tracking with configurable limits (default: 500MB)
- **Timeout Protection**: Prevents runaway benchmarks (default: 30 seconds)
- **Circuit Validation**: Complexity limits for qubits and circuit depth
- **Graceful Degradation**: Safe fallbacks when limits are exceeded

### 🔬 Advanced Testing
- **Stress Testing**: Concurrent operations with memory pressure simulation
- **Performance Analysis**: Throughput, latency, and volatility metrics
- **Regression Detection**: Automated comparison with historical baselines
- **Memory Leak Detection**: Identifies potential memory issues

### 🏗️ CI/CD Integration
- **Auto-Detection**: Supports GitHub Actions, GitLab CI, Jenkins, Azure DevOps
- **Automated Reporting**: JSON reports, summaries, and badge generation
- **Regression Analysis**: Automated performance regression detection
- **Artifact Generation**: Ready-to-use CI artifacts and reports

## Quick Start

### Basic Usage

```go
package main

import (
    "github.com/kegliz/qplay/qc/benchmark"
    "github.com/kegliz/qplay/qc/testutil"
)

func main() {
    // Create a benchmark suite
    suite := benchmark.NewPluginBenchmarkSuite().
        WithRunners("itsu").
        WithCircuits(benchmark.SimpleCircuit, benchmark.EntanglementCircuit).
        WithScenarios(benchmark.SerialExecution, benchmark.ParallelExecution).
        WithConfig(testutil.QuickTestConfig)

    // Run benchmarks (implementation would use testing.B)
    // results := suite.RunAll()
}
```

### CLI Demo

```bash
# List available quantum backends
go run cmd/benchmark-demo/main.go -cmd=list

# Run a specific benchmark
go run cmd/benchmark-demo/main.go -cmd=benchmark -runner=itsu -circuit=simple

# Run all benchmarks with JSON output
go run cmd/benchmark-demo/main.go -cmd=benchmark-all -output=json
```

## Configuration

### Resource Limits

```go
limits := benchmark.ResourceLimits{
    MaxMemoryMB:     500,               // Memory limit in MB
    MaxDuration:     30 * time.Second,  // Per-benchmark timeout
    MaxCircuitDepth: 20,                // Maximum circuit depth
    MaxQubits:       5,                 // Maximum number of qubits
}

config := benchmark.BenchmarkConfig{
    CircuitType: benchmark.SimpleCircuit,
    Scenario:    benchmark.SerialExecution,
    Config:      testutil.QuickTestConfig,
    RunnerName:  "itsu",
    Limits:      limits,
}
```

### Test Configurations

The framework provides several pre-configured test scenarios:

- `testutil.QuickTestConfig`: Fast tests with minimal resources
- `testutil.ConservativeTestConfig`: Resource-constrained environments
- `testutil.StandardTestConfig`: Balanced performance and accuracy
- `testutil.LargeTestConfig`: Comprehensive testing with more resources

## Benchmark Scenarios

### 1. Serial Execution
Tests basic sequential circuit execution:
```go
scenario := benchmark.SerialExecution
```

### 2. Parallel Execution
Tests concurrent circuit execution:
```go
scenario := benchmark.ParallelExecution
```

### 3. Batch Execution
Tests efficient batch processing (if supported by backend):
```go
scenario := benchmark.BatchExecution
```

### 4. Context Execution
Tests context-based execution with cancellation:
```go
scenario := benchmark.ContextExecution
```

### 5. Metrics Collection
Tests performance monitoring capabilities:
```go
scenario := benchmark.MetricsCollection
```

## Circuit Types

### Simple Circuit
Basic single-qubit operations:
```go
circuitType := benchmark.SimpleCircuit
```

### Entanglement Circuit
Multi-qubit entangling operations:
```go
circuitType := benchmark.EntanglementCircuit
```

### Superposition Circuit
Hadamard-based superposition states:
```go
circuitType := benchmark.SuperpositionCircuit
```

### Mixed Gates Circuit
Combination of various quantum gates:
```go
circuitType := benchmark.MixedGatesCircuit
```

## Advanced Features

### Stress Testing

```go
import "github.com/kegliz/qplay/qc/benchmark"

// Configure stress test
config := benchmark.StressTestConfig{
    Duration:        30 * time.Second,
    ConcurrentOps:   10,
    MemoryPressure:  true,
    CircuitSizes:    []int{2, 3, 4, 5},
    MaxMemoryMB:     1024,
    RecoveryEnabled: true,
}

// Run stress test
result := benchmark.RunStressTest("itsu", config)

if !result.Success {
    fmt.Printf("Stress test failed: %s\n", result.Error)
} else {
    fmt.Printf("Stress test passed: %d/%d operations successful\n", 
        result.SuccessfulOps, result.TotalOperations)
}
```

### Benchmark Persistence

```go
// Create persistence manager
persistence := benchmark.NewBenchmarkPersistence("./benchmark-data")

// Save results
err := persistence.SaveResult(result, "abc123", "v1.0.0")

// Load history
history, err := persistence.LoadHistory("itsu", benchmark.SimpleCircuit, benchmark.SerialExecution)

// Compare with baseline
comparison, err := persistence.CompareWithBaseline(currentResult, "def456", "v1.0.1")
```

### CI/CD Integration

```go
// Create CI runner (auto-detects environment)
ciRunner := benchmark.NewCIBenchmarkRunner("./ci-artifacts")

// Run complete benchmark suite
report, err := ciRunner.RunBenchmarkSuite()

if err != nil {
    fmt.Printf("CI benchmarks failed: %v\n", err)
    os.Exit(1)
}

// Check for regressions
if report.RegressionAnalysis.OverallStatus == "fail" {
    fmt.Printf("Performance regressions detected!\n")
    os.Exit(1)
}
```

## CI/CD Environment Setup

### GitHub Actions

```yaml
name: Quantum Benchmarks
on: [push, pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Run Benchmarks
      run: |
        go run cmd/benchmark-demo/main.go -cmd=benchmark-all -output=json
    
    - name: Upload Artifacts
      uses: actions/upload-artifact@v3
      with:
        name: benchmark-results
        path: ci-artifacts/
```

### GitLab CI

```yaml
stages:
  - benchmark

quantum-benchmark:
  stage: benchmark
  script:
    - go run cmd/benchmark-demo/main.go -cmd=benchmark-all -output=json
  artifacts:
    reports:
      junit: ci-artifacts/benchmark-report.json
    paths:
      - ci-artifacts/
```

## Environment Variables

Configure behavior through environment variables:

- `BENCHMARK_VERBOSE=true`: Enable verbose output
- `BENCHMARK_MEMORY_LIMIT=1024`: Set memory limit in MB
- `BENCHMARK_TIMEOUT=60s`: Set timeout duration
- `BENCHMARK_OUTPUT_DIR=./results`: Set output directory

## Output Formats

### JSON Report
```json
{
  "runner_name": "itsu",
  "circuit_type": "simple",
  "scenario": "serial",
  "success": true,
  "duration": 1234567,
  "resource_usage": {
    "memory_delta": 1048576,
    "peak_memory": 2097152,
    "circuit_qubits": 2,
    "circuit_depth": 3
  },
  "backend_info": {
    "name": "Itsu Quantum Simulator",
    "version": "v1.0.0"
  }
}
```

### Console Output
```
🔌 Quantum Backend Plugin Benchmarks
===================================

Runner: itsu
Circuit: simple (2 qubits, depth 3)
Scenario: serial
Duration: 1.23ms
Memory: 1.0MB delta, 2.0MB peak
Status: ✅ PASSED
```

## Best Practices

### 1. Resource Management
- Always set appropriate resource limits for your environment
- Use conservative settings in CI/CD environments
- Monitor memory usage in long-running tests

### 2. Circuit Design
- Start with simple circuits for basic validation
- Use realistic circuit sizes that match your use case
- Consider quantum hardware limitations

### 3. Performance Testing
- Establish baselines before making changes
- Run benchmarks consistently in similar environments
- Track trends over time rather than single measurements

### 4. CI Integration
- Set appropriate timeouts for CI environments
- Use benchmark results to catch performance regressions
- Store historical data for trend analysis

## Troubleshooting

### Common Issues

#### Memory Limit Exceeded
```
Error: current memory usage 600MB exceeds limit 500MB
```
**Solution**: Increase memory limit or reduce circuit complexity:
```go
limits.MaxMemoryMB = 1024
```

#### Timeout Errors
```
Error: serial run timed out after 30s
```
**Solution**: Increase timeout or optimize backend:
```go
limits.MaxDuration = 60 * time.Second
```

#### Circuit Complexity Violations
```
Error: circuit has 6 qubits, limit is 5
```
**Solution**: Reduce circuit size or increase limits:
```go
limits.MaxQubits = 10
```

### Performance Issues

#### Slow Benchmarks
1. Check if backend supports batch execution
2. Reduce number of shots for quick tests
3. Use parallel execution when available

#### Memory Leaks
1. Enable stress testing to detect leaks
2. Monitor memory usage over time
3. Check backend implementation for resource cleanup

### CI/CD Issues

#### Environment Detection
If CI environment isn't detected automatically:
```bash
export CI_ENVIRONMENT=github-actions
```

#### Artifact Generation
Ensure output directory is writable:
```bash
mkdir -p ci-artifacts
chmod 755 ci-artifacts
```

## Contributing

### Adding New Circuit Types

1. Implement circuit builder function:
```go
func buildMyCircuit(qubits int) circuit.Build {
    // Implementation
}
```

2. Register in StandardCircuits:
```go
StandardCircuits[MyCircuit] = buildMyCircuit
```

### Adding New Scenarios

1. Implement scenario function:
```go
func runMyScenario(b *testing.B, runner simulator.OneShotRunner, 
    circ circuit.Circuit, config BenchmarkConfig) error {
    // Implementation
}
```

2. Add to scenario switch in `runBenchmarkScenario`

### Testing Changes

```bash
# Run unit tests
go test ./qc/benchmark/...

# Run integration tests
go test -tags=integration ./qc/benchmark/...

# Run benchmark tests
go test -bench=. ./qc/benchmark/...
```

## API Reference

See the [API documentation](./api.md) for detailed function signatures and usage examples.

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.
