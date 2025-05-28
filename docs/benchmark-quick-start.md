# Quantum Benchmarking Framework - Quick Start Guide

## 🚀 Quick Start

### 1. Build the CLI
```bash
cd /Users/zoltankegli/work/qc/qplay
go build -o bin/benchmark-demo ./cmd/benchmark-demo
```

### 2. Run Basic Benchmark
```bash
# Single circuit test
./bin/benchmark-demo -cmd=run -circuit=simple

# Full benchmark
./bin/benchmark-demo -cmd=benchmark -circuit=entanglement -shots=100

# All circuits and runners
./bin/benchmark-demo -cmd=benchmark-all -output=json
```

### 3. Stress Testing
```bash
# Quick stress test (5 seconds, 3 workers)
./bin/benchmark-demo -cmd=stress -duration=5s -concurrent=3

# Production stress test (60 seconds, 10 workers)
./bin/benchmark-demo -cmd=stress -duration=60s -concurrent=10 -output=json
```

### 4. CI/CD Integration
```bash
# Run CI benchmark suite
./bin/benchmark-demo -cmd=ci -output=json > benchmark-results.json
```

## 📋 Available Commands

| Command | Description | Example |
|---------|-------------|---------|
| `run` | Execute single circuit | `-cmd=run -circuit=simple` |
| `benchmark` | Run benchmark with specified parameters | `-cmd=benchmark -shots=100 -circuit=entanglement` |
| `benchmark-all` | Run comprehensive benchmark suite | `-cmd=benchmark-all -output=json` |
| `stress` | Execute stress testing | `-cmd=stress -duration=30s -concurrent=5` |
| `ci` | Run CI/CD benchmark suite | `-cmd=ci -output=json` |
| `list` | List available runners | `-cmd=list` |
| `info` | Show runner information | `-cmd=info -runner=itsu` |

## 🔧 Configuration Parameters

| Parameter | Description | Default | Example |
|-----------|-------------|---------|---------|
| `-circuit` | Circuit type: simple, entanglement, superposition, mixed | `simple` | `-circuit=entanglement` |
| `-runner` | Quantum backend runner name | `itsu` | `-runner=itsubaki` |
| `-shots` | Number of shots for benchmark | `100` | `-shots=1000` |
| `-qubits` | Number of qubits | `2` | `-qubits=4` |
| `-scenario` | Execution scenario: serial, parallel, batch, context, metrics | `serial` | `-scenario=parallel` |
| `-output` | Output format: console, json | `console` | `-output=json` |
| `-output-dir` | Directory for output files | `./benchmark-results` | `-output-dir=/tmp/bench` |
| `-duration` | Duration for stress test | `10s` | `-duration=60s` |
| `-concurrent` | Concurrent workers for stress test | `5` | `-concurrent=20` |
| `-workers` | Number of worker threads | `4` | `-workers=8` |

## 🔬 Circuit Types

| Circuit | Gates | Qubits | Description |
|---------|-------|--------|-------------|
| `simple` | H, Measure | 2 | Basic superposition and measurement |
| `entanglement` | H, CNOT, Measure | 2 | Bell state creation |
| `superposition` | Multiple H | Variable | Multiple qubit superposition |
| `mixed` | X, Y, Z, H, CNOT | Variable | Comprehensive gate testing |

## 📊 Output Formats

### Console Output (Human Readable)
```
🏁 Running Benchmark: itsu
======================
✅ Circuit: Simple Circuit (H + Measure)
⏱️  Duration: 167µs
📊 Memory Usage: 208KB → 191KB (-17KB)
```

### JSON Output (Machine Readable)
```json
{
  "timestamp": "2025-05-28T13:34:32.442514+02:00",
  "total_runners": 1,
  "results": [{
    "runner_name": "itsu",
    "circuit_type": "simple",
    "success": true,
    "duration": 167,
    "resource_usage": {
      "start_memory": 208608,
      "end_memory": 191400,
      "memory_delta": -17208
    }
  }]
}
```

## 🔍 Programmatic Usage

### Basic Benchmark
```go
import "github.com/kegliz/qplay/qc/benchmark"

config := benchmark.BenchmarkConfig{
    CircuitType: benchmark.SimpleCircuit,
    Scenario:    benchmark.SerialExecution,
    RunnerName:  "itsu",
    Limits:      benchmark.DefaultResourceLimits,
}

result := benchmark.RunSingleBenchmark(b, config)
```

### Stress Testing
```go
config := benchmark.StressTestConfig{
    Duration:      30 * time.Second,
    ConcurrentOps: 10,
    MemoryPressure: true,
    MaxMemoryMB:   512,
}

result := benchmark.RunStressTest("itsu", config)
```

## 🚨 Resource Management

### Memory Limits
- **Default**: 512MB per benchmark
- **Stress test**: 1GB limit
- **Emergency stop**: Automatic termination on limit exceeded
- **Leak detection**: Automatic memory leak monitoring

### Timeout Management
- **Default**: 5-minute timeout per benchmark
- **Configurable**: Set via `ResourceLimits.MaxDuration`
- **Context-based**: Graceful cancellation supported
- **Emergency stop**: Force termination on timeout

### Circuit Complexity
- **Max depth**: 100 gates (configurable)
- **Max qubits**: 20 qubits (configurable)
- **Validation**: Pre-execution complexity check
- **Scaling**: Automatic resource estimation

## 🏗️ CI/CD Integration

### GitHub Actions Example
```yaml
- name: Run Quantum Benchmarks
  run: |
    go build -o bin/benchmark-demo ./cmd/benchmark-demo
    ./bin/benchmark-demo -cmd=ci -output=json > benchmark-results.json
    
- name: Upload Benchmark Results
  uses: actions/upload-artifact@v3
  with:
    name: benchmark-results
    path: benchmark-results.json
```

### GitLab CI Example
```yaml
benchmark:
  script:
    - go build -o bin/benchmark-demo ./cmd/benchmark-demo
    - ./bin/benchmark-demo -cmd=ci -output=json
  artifacts:
    reports:
      performance: benchmark-results/benchmark-report.json
```

## 🔧 Troubleshooting

### Common Issues

#### Memory Leaks
```bash
# Monitor memory usage during stress test
./bin/benchmark-demo -cmd=stress -duration=60s -output=json | jq '.memory_leaks'
```

#### Timeout Issues
```bash
# Increase timeout for complex circuits
export BENCHMARK_TIMEOUT=600s
./bin/benchmark-demo -cmd=benchmark -circuit=mixed -qubits=10
```

#### Runner Not Found
```bash
# List available runners
./bin/benchmark-demo -cmd=list

# Check runner info
./bin/benchmark-demo -cmd=info -runner=itsu
```

### Performance Tuning

#### For High Throughput
```bash
# Use parallel execution with more workers
./bin/benchmark-demo -cmd=benchmark -scenario=parallel -workers=16
```

#### For Memory Efficiency
```bash
# Use batch execution
./bin/benchmark-demo -cmd=benchmark -scenario=batch -shots=1000
```

#### For Detailed Analysis
```bash
# Use metrics collection
./bin/benchmark-demo -cmd=benchmark -scenario=metrics -output=json
```

## 📈 Monitoring and Analysis

### Key Metrics to Watch
- **Execution time**: Latency per operation
- **Throughput**: Operations per second
- **Memory usage**: Peak and delta memory
- **Success rate**: Percentage of successful operations
- **Resource efficiency**: Memory/CPU usage trends

### Historical Analysis
```bash
# Check benchmark history
ls -la benchmark-results/benchmark-history/

# Compare results over time
cat benchmark-results/benchmark-history/bench_itsu_simple_serial.json | jq '.duration'
```

## 🎯 Best Practices

1. **Start small**: Begin with simple circuits and low shot counts
2. **Monitor resources**: Watch memory and CPU usage during benchmarks
3. **Use appropriate timeouts**: Set realistic limits based on circuit complexity
4. **Validate results**: Check success rates and error patterns
5. **Store history**: Keep benchmark results for trend analysis
6. **Test incrementally**: Gradually increase complexity and load
7. **Use CI integration**: Automate benchmarking in your development pipeline

## 📞 Support

- **Documentation**: `/docs/benchmark-framework.md`
- **Examples**: `/examples/` directory
- **Tests**: Run `go test ./qc/benchmark/...`
- **Issues**: Check error messages and logs for detailed information
