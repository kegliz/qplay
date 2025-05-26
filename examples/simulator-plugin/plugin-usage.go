package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/simulator"

	// Import the itsu package to register the plugin
	_ "github.com/kegliz/qplay/qc/simulator/itsu"
)

func main() {
	fmt.Println("🔌 Quantum Backend Plugin Architecture Demo")
	fmt.Println("==========================================")

	// 1. List available runners
	fmt.Println("\n1. Available Runners:")
	runners := simulator.ListRunners()
	for _, name := range runners {
		fmt.Printf("   • %s\n", name)
	}

	// 2. Create a runner using the plugin system
	fmt.Println("\n2. Creating runner via plugin system:")
	runner, err := simulator.CreateRunner("itsu")
	if err != nil {
		panic(fmt.Sprintf("Failed to create runner: %v", err))
	}
	fmt.Printf("   ✅ Successfully created '%s' runner\n", "itsu")

	// 3. Check capabilities
	fmt.Println("\n3. Checking runner capabilities:")
	fmt.Printf("   • Context Support:    %t\n", simulator.SupportsContext(runner))
	fmt.Printf("   • Configuration:      %t\n", simulator.SupportsConfiguration(runner))
	fmt.Printf("   • Metrics Collection: %t\n", simulator.SupportsMetrics(runner))
	fmt.Printf("   • Circuit Validation: %t\n", simulator.SupportsValidation(runner))
	fmt.Printf("   • Batch Execution:    %t\n", simulator.SupportsBatch(runner))
	fmt.Printf("   • Backend Info:       %t\n", simulator.SupportsBackendInfo(runner))

	// 4. Get backend information
	if info := simulator.GetBackendInfo(runner); info != nil {
		fmt.Printf("\n4. Backend Information:\n")
		fmt.Printf("   • Name:        %s\n", info.Name)
		fmt.Printf("   • Version:     %s\n", info.Version)
		fmt.Printf("   • Description: %s\n", info.Description)
		fmt.Printf("   • Vendor:      %s\n", info.Vendor)
	}

	// 5. Configure the runner
	if configRunner, ok := runner.(simulator.ConfigurableRunner); ok {
		fmt.Println("\n5. Configuring runner:")
		err := configRunner.Configure(map[string]interface{}{
			"verbose": true,
			"timeout": 30,
		})
		if err != nil {
			fmt.Printf("   ❌ Configuration failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Configuration applied successfully\n")
			config := configRunner.GetConfiguration()
			fmt.Printf("   • Current config: %+v\n", config)
		}
	}

	// 6. Create a simple circuit
	fmt.Println("\n6. Creating and running a quantum circuit:")
	b := builder.New(builder.Q(2), builder.C(2))
	b.H(0).CNOT(0, 1).Measure(0, 0).Measure(1, 1)
	circuit, err := b.BuildCircuit()
	if err != nil {
		panic(fmt.Sprintf("Failed to build circuit: %v", err))
	}

	// 7. Test context-based execution
	if contextRunner, ok := runner.(simulator.ContextualRunner); ok {
		fmt.Println("\n7. Testing context-based execution:")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := contextRunner.RunOnceWithContext(ctx, circuit)
		if err != nil {
			fmt.Printf("   ❌ Execution failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Result: %s\n", result)
		}
	}

	// 8. Test batch execution
	if batchRunner, ok := runner.(simulator.BatchRunner); ok {
		fmt.Println("\n8. Testing batch execution:")
		results, err := batchRunner.RunBatch(circuit, 10)
		if err != nil {
			fmt.Printf("   ❌ Batch execution failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Batch results (%d shots): %v\n", len(results), results)
		}
	}

	// 9. Check metrics
	if metricsRunner, ok := runner.(simulator.MetricsCollector); ok {
		fmt.Println("\n9. Execution metrics:")
		metrics := metricsRunner.GetMetrics()
		fmt.Printf("   • Total Executions: %d\n", metrics.TotalExecutions)
		fmt.Printf("   • Successful Runs:  %d\n", metrics.SuccessfulRuns)
		fmt.Printf("   • Failed Runs:      %d\n", metrics.FailedRuns)
		fmt.Printf("   • Average Time:     %v\n", metrics.AverageTime)
		fmt.Printf("   • Total Time:       %v\n", metrics.TotalTime)
	}

	// 10. Create simulator using plugin system
	fmt.Println("\n10. Creating simulator with plugin runner:")
	sim, err := simulator.NewSimulatorWithRunner("itsu", simulator.SimulatorOptions{
		Shots:   1024,
		Workers: 8,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create simulator: %v", err))
	}

	results, err := sim.Run(circuit)
	if err != nil {
		panic(fmt.Sprintf("Simulation failed: %v", err))
	}

	fmt.Printf("   ✅ Simulation completed with %d unique outcomes\n", len(results))
	for state, count := range results {
		probability := float64(count) / 1024.0
		fmt.Printf("      |%s⟩: %4d shots (%.1f%%)\n", state, count, probability*100)
	}

	fmt.Println("\n🎉 Plugin architecture demonstration completed successfully!")
}
