package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/simulator"

	// Import the itsu package to register the plugin
	_ "github.com/kegliz/qplay/qc/simulator/itsu"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "list":
		listRunners()
	case "info":
		if len(os.Args) < 3 {
			fmt.Println("Usage: plugin-demo info <runner-name>")
			os.Exit(1)
		}
		showRunnerInfo(os.Args[2])
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: plugin-demo run <runner-name>")
			os.Exit(1)
		}
		runExample(os.Args[2])
	case "benchmark":
		if len(os.Args) < 3 {
			fmt.Println("Usage: plugin-demo benchmark <runner-name>")
			os.Exit(1)
		}
		benchmarkRunner(os.Args[2])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Plugin Architecture Demo")
	fmt.Println("Usage: plugin-demo <command> [args...]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list                    List all registered runners")
	fmt.Println("  info <runner-name>      Show detailed information about a runner")
	fmt.Println("  run <runner-name>       Run a simple circuit with the specified runner")
	fmt.Println("  benchmark <runner-name> Benchmark the runner performance")
}

func listRunners() {
	runners := simulator.ListRunners()
	fmt.Printf("Registered quantum backend runners (%d total):\n\n", len(runners))

	for _, name := range runners {
		runner, err := simulator.CreateRunner(name)
		if err != nil {
			fmt.Printf("  %-15s ERROR: %v\n", name, err)
			continue
		}

		if info := simulator.GetBackendInfo(runner); info != nil {
			fmt.Printf("  %-15s %s (%s)\n", name, info.Name, info.Version)
		} else {
			fmt.Printf("  %-15s (no backend info available)\n", name)
		}
	}
}

func showRunnerInfo(runnerName string) {
	runner, err := simulator.CreateRunner(runnerName)
	if err != nil {
		log.Fatalf("Failed to create runner %q: %v", runnerName, err)
	}

	fmt.Printf("Runner: %s\n", runnerName)
	fmt.Println("=" + fmt.Sprintf("%*s", len(runnerName)+7, ""))

	// Backend information
	if info := simulator.GetBackendInfo(runner); info != nil {
		fmt.Printf("\nBackend Information:\n")
		fmt.Printf("  Name:        %s\n", info.Name)
		fmt.Printf("  Version:     %s\n", info.Version)
		fmt.Printf("  Description: %s\n", info.Description)
		fmt.Printf("  Vendor:      %s\n", info.Vendor)

		fmt.Printf("\nCapabilities:\n")
		for capability, supported := range info.Capabilities {
			status := "❌"
			if supported {
				status = "✅"
			}
			fmt.Printf("  %s %s\n", status, capability)
		}

		if len(info.Metadata) > 0 {
			fmt.Printf("\nMetadata:\n")
			for key, value := range info.Metadata {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	// Supported gates
	if validator, ok := runner.(simulator.ValidatingRunner); ok {
		gates := validator.GetSupportedGates()
		fmt.Printf("\nSupported Gates (%d):\n", len(gates))
		for i, gate := range gates {
			if i > 0 && i%8 == 0 {
				fmt.Println()
			}
			fmt.Printf("  %-10s", gate)
		}
		fmt.Println()
	}

	// Interface support
	fmt.Printf("\nInterface Support:\n")
	fmt.Printf("  Context Support:    %s\n", boolToStatus(simulator.SupportsContext(runner)))
	fmt.Printf("  Configuration:      %s\n", boolToStatus(simulator.SupportsConfiguration(runner)))
	fmt.Printf("  Metrics Collection: %s\n", boolToStatus(simulator.SupportsMetrics(runner)))
	fmt.Printf("  Circuit Validation: %s\n", boolToStatus(simulator.SupportsValidation(runner)))
	fmt.Printf("  Batch Execution:    %s\n", boolToStatus(simulator.SupportsBatch(runner)))
}

func runExample(runnerName string) {
	fmt.Printf("Running Bell State example with %s runner...\n\n", runnerName)

	// Create a Bell state circuit
	b := builder.New(builder.Q(2), builder.C(2))
	b.H(0).CNOT(0, 1).Measure(0, 0).Measure(1, 1)

	circuit, err := b.BuildCircuit()
	if err != nil {
		log.Fatalf("Failed to build circuit: %v", err)
	}

	// Create simulator with the specified runner
	sim, err := simulator.NewSimulatorWithDefaults(runnerName)
	if err != nil {
		log.Fatalf("Failed to create simulator: %v", err)
	}

	// Run the simulation
	fmt.Printf("Running 1024 shots...\n")
	start := time.Now()
	results, err := sim.Run(circuit)
	duration := time.Since(start)

	if err != nil {
		log.Fatalf("Simulation failed: %v", err)
	}

	// Display results
	fmt.Printf("Results (completed in %v):\n", duration)
	for state, count := range results {
		probability := float64(count) / 1024.0
		fmt.Printf("  |%s⟩: %4d shots (%.1f%%)\n", state, count, probability*100)
	}

	// Show metrics if available
	if runner, err := simulator.CreateRunner(runnerName); err == nil {
		if collector, ok := runner.(simulator.MetricsCollector); ok {
			metrics := collector.GetMetrics()
			fmt.Printf("\nExecution Metrics:\n")
			fmt.Printf("  Total Executions: %d\n", metrics.TotalExecutions)
			fmt.Printf("  Successful Runs:  %d\n", metrics.SuccessfulRuns)
			fmt.Printf("  Failed Runs:      %d\n", metrics.FailedRuns)
			fmt.Printf("  Average Time:     %v\n", metrics.AverageTime)
			fmt.Printf("  Total Time:       %v\n", metrics.TotalTime)
		}
	}
}

func benchmarkRunner(runnerName string) {
	fmt.Printf("Benchmarking %s runner...\n\n", runnerName)

	// Create a more complex circuit for benchmarking
	b := builder.New(builder.Q(3), builder.C(3))
	b.H(0).H(1).H(2)
	b.CNOT(0, 1).CNOT(1, 2)
	b.Measure(0, 0).Measure(1, 1).Measure(2, 2)

	circuit, err := b.BuildCircuit()
	if err != nil {
		log.Fatalf("Failed to build circuit: %v", err)
	}

	runner, err := simulator.CreateRunner(runnerName)
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	shots := []int{100, 500, 1000, 5000}

	fmt.Println("Single-shot benchmarks:")
	for _, shotCount := range shots {
		start := time.Now()
		for i := 0; i < shotCount; i++ {
			_, err := runner.RunOnce(circuit)
			if err != nil {
				log.Fatalf("Run failed: %v", err)
			}
		}
		duration := time.Since(start)
		rate := float64(shotCount) / duration.Seconds()
		fmt.Printf("  %5d shots: %8v (%6.0f shots/sec)\n", shotCount, duration, rate)
	}

	// Test batch execution if supported
	if batchRunner, ok := runner.(simulator.BatchRunner); ok {
		fmt.Println("\nBatch execution benchmarks:")
		for _, shotCount := range shots {
			start := time.Now()
			_, err := batchRunner.RunBatch(circuit, shotCount)
			if err != nil {
				log.Fatalf("Batch run failed: %v", err)
			}
			duration := time.Since(start)
			rate := float64(shotCount) / duration.Seconds()
			fmt.Printf("  %5d shots: %8v (%6.0f shots/sec)\n", shotCount, duration, rate)
		}
	}

	// Test context-based execution if supported
	if contextRunner, ok := runner.(simulator.ContextualRunner); ok {
		fmt.Println("\nContext-based execution test:")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		count := 0
		start := time.Now()
		for {
			_, err := contextRunner.RunOnceWithContext(ctx, circuit)
			if err != nil {
				if err == context.DeadlineExceeded {
					break
				}
				log.Fatalf("Context run failed: %v", err)
			}
			count++
		}
		duration := time.Since(start)
		rate := float64(count) / duration.Seconds()
		fmt.Printf("  Completed %d shots in %v (%6.0f shots/sec) before timeout\n", count, duration, rate)
	}
}

func boolToStatus(b bool) string {
	if b {
		return "✅ Supported"
	}
	return "❌ Not Supported"
}
