package simulator

import (
	"context"
	"testing"

	"maps"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFullFeaturedRunner implements all enhanced interfaces for testing
type mockFullFeaturedRunner struct {
	*mockOneShotRunner
	backendInfo BackendInfo
	config      map[string]interface{}
	metrics     ExecutionMetrics
}

func newMockFullFeaturedRunner() *mockFullFeaturedRunner {
	return &mockFullFeaturedRunner{
		mockOneShotRunner: newMockOneShotRunner(func(c circuit.Circuit, callNum int) (string, error) {
			return "0", nil
		}),
		backendInfo: BackendInfo{
			Name:        "Mock Runner",
			Version:     "v1.0.0",
			Description: "Mock runner for testing",
			Vendor:      "test",
			Capabilities: map[string]bool{
				"context_support": true,
				"batch_execution": true,
			},
		},
		config:  make(map[string]interface{}),
		metrics: ExecutionMetrics{},
	}
}

func (m *mockFullFeaturedRunner) GetBackendInfo() BackendInfo {
	return m.backendInfo
}

func (m *mockFullFeaturedRunner) RunOnceWithContext(ctx context.Context, c circuit.Circuit) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return m.RunOnce(c)
	}
}

func (m *mockFullFeaturedRunner) Configure(options map[string]any) error {
	maps.Copy(m.config, options)
	return nil
}

func (m *mockFullFeaturedRunner) SetVerbose(verbose bool) {
	m.config["verbose"] = verbose
}

func (m *mockFullFeaturedRunner) GetConfiguration() map[string]any {
	return m.config
}

func (m *mockFullFeaturedRunner) GetMetrics() ExecutionMetrics {
	return m.metrics
}

func (m *mockFullFeaturedRunner) ResetMetrics() {
	m.metrics = ExecutionMetrics{}
}

func (m *mockFullFeaturedRunner) ValidateCircuit(c circuit.Circuit) error {
	return nil
}

func (m *mockFullFeaturedRunner) GetSupportedGates() []string {
	return []string{"H", "X", "CNOT", "MEASURE"}
}

func (m *mockFullFeaturedRunner) RunBatch(c circuit.Circuit, shots int) ([]string, error) {
	results := make([]string, shots)
	for i := range shots {
		result, err := m.RunOnce(c)
		if err != nil {
			return results[:i], err
		}
		results[i] = result
	}
	return results, nil
}

func TestRunnerRegistry(t *testing.T) {
	// Create a separate registry for testing to avoid conflicts
	registry := NewRunnerRegistry()

	t.Run("Register and Create", func(t *testing.T) {
		err := registry.Register("test-runner", func() OneShotRunner {
			return newMockOneShotRunner(func(c circuit.Circuit, callNum int) (string, error) {
				return "test", nil
			})
		})
		require.NoError(t, err)

		runner, err := registry.Create("test-runner")
		require.NoError(t, err)
		assert.NotNil(t, runner)

		// Test the runner works
		testCirc := createSimpleTestCircuit(t)
		result, err := runner.RunOnce(testCirc)
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("Duplicate Registration", func(t *testing.T) {
		factory := func() OneShotRunner { return newMockOneShotRunner(nil) }

		err := registry.Register("duplicate", factory)
		require.NoError(t, err)

		err = registry.Register("duplicate", factory)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("Unknown Runner", func(t *testing.T) {
		runner, err := registry.Create("unknown-runner")
		assert.Error(t, err)
		assert.Nil(t, runner)
		assert.Contains(t, err.Error(), "unknown runner")
	})

	t.Run("List Runners", func(t *testing.T) {
		registry.Register("runner1", func() OneShotRunner { return newMockOneShotRunner(nil) })
		registry.Register("runner2", func() OneShotRunner { return newMockOneShotRunner(nil) })

		runners := registry.ListRunners()
		assert.Contains(t, runners, "runner1")
		assert.Contains(t, runners, "runner2")
		assert.GreaterOrEqual(t, len(runners), 2)
	})

	t.Run("Unregister", func(t *testing.T) {
		registry.Register("to-remove", func() OneShotRunner { return newMockOneShotRunner(nil) })

		removed := registry.Unregister("to-remove")
		assert.True(t, removed)

		_, err := registry.Create("to-remove")
		assert.Error(t, err)

		removed = registry.Unregister("non-existent")
		assert.False(t, removed)
	})

	t.Run("MustRegister Panic", func(t *testing.T) {
		assert.Panics(t, func() {
			registry.MustRegister("", func() OneShotRunner { return newMockOneShotRunner(nil) })
		})
	})
}

func TestEnhancedInterfaces(t *testing.T) {
	runner := newMockFullFeaturedRunner()

	t.Run("BackendProvider", func(t *testing.T) {
		assert.True(t, SupportsBackendInfo(runner))

		info := GetBackendInfo(runner)
		require.NotNil(t, info)
		assert.Equal(t, "Mock Runner", info.Name)
		assert.Equal(t, "v1.0.0", info.Version)
		assert.True(t, info.Capabilities["context_support"])
	})

	t.Run("ConfigurableRunner", func(t *testing.T) {
		assert.True(t, SupportsConfiguration(runner))

		err := runner.Configure(map[string]interface{}{
			"verbose": true,
			"timeout": 30,
		})
		require.NoError(t, err)

		config := runner.GetConfiguration()
		assert.Equal(t, true, config["verbose"])
		assert.Equal(t, 30, config["timeout"])
	})

	t.Run("MetricsCollector", func(t *testing.T) {
		assert.True(t, SupportsMetrics(runner))

		metrics := runner.GetMetrics()
		assert.Equal(t, int64(0), metrics.TotalExecutions)

		runner.ResetMetrics()
		metrics = runner.GetMetrics()
		assert.Equal(t, int64(0), metrics.TotalExecutions)
	})

	t.Run("ValidatingRunner", func(t *testing.T) {
		assert.True(t, SupportsValidation(runner))

		testCirc := createSimpleTestCircuit(t)
		err := runner.ValidateCircuit(testCirc)
		assert.NoError(t, err)

		gates := runner.GetSupportedGates()
		assert.Contains(t, gates, "H")
		assert.Contains(t, gates, "MEASURE")
	})

	t.Run("ContextualRunner", func(t *testing.T) {
		assert.True(t, SupportsContext(runner))

		testCirc := createSimpleTestCircuit(t)

		// Test normal execution
		ctx := context.Background()
		result, err := runner.RunOnceWithContext(ctx, testCirc)
		require.NoError(t, err)
		assert.Equal(t, "0", result)

		// Test cancellation
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err = runner.RunOnceWithContext(ctx, testCirc)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("BatchRunner", func(t *testing.T) {
		assert.True(t, SupportsBatch(runner))

		testCirc := createSimpleTestCircuit(t)
		results, err := runner.RunBatch(testCirc, 5)
		require.NoError(t, err)
		assert.Len(t, results, 5)
		for _, result := range results {
			assert.Equal(t, "0", result)
		}
	})
}

func TestSimulatorWithPlugins(t *testing.T) {
	// Create a test registry to avoid conflicts with default
	testRegistry := NewRunnerRegistry()
	testRegistry.Register("test-plugin", func() OneShotRunner {
		return newMockOneShotRunner(func(c circuit.Circuit, callNum int) (string, error) {
			return "plugin-result", nil
		})
	})

	t.Run("NewSimulatorWithRunner", func(t *testing.T) {
		// This test would work with the default registry if itsu is registered
		// For now, we'll test the error case
		_, err := NewSimulatorWithRunner("non-existent-runner", SimulatorOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create runner")
	})

	t.Run("NewSimulatorWithDefaults", func(t *testing.T) {
		// This test would work with the default registry if itsu is registered
		// For now, we'll test the error case
		_, err := NewSimulatorWithDefaults("non-existent-runner")
		assert.Error(t, err)
	})
}

func TestCapabilityChecking(t *testing.T) {
	// Test with basic runner (only implements OneShotRunner)
	basicRunner := newMockOneShotRunner(nil)
	assert.False(t, SupportsContext(basicRunner))
	assert.False(t, SupportsConfiguration(basicRunner))
	assert.False(t, SupportsMetrics(basicRunner))
	assert.False(t, SupportsValidation(basicRunner))
	assert.False(t, SupportsBatch(basicRunner))
	assert.Nil(t, GetBackendInfo(basicRunner))

	// Test with full-featured runner
	fullRunner := newMockFullFeaturedRunner()
	assert.True(t, SupportsContext(fullRunner))
	assert.True(t, SupportsConfiguration(fullRunner))
	assert.True(t, SupportsMetrics(fullRunner))
	assert.True(t, SupportsValidation(fullRunner))
	assert.True(t, SupportsBatch(fullRunner))
	assert.NotNil(t, GetBackendInfo(fullRunner))
}

func createSimpleTestCircuit(t *testing.T) circuit.Circuit {
	b := builder.New(builder.Q(1), builder.C(1))
	b.H(0).Measure(0, 0)
	c, err := b.BuildCircuit()
	require.NoError(t, err)
	return c
}

func TestDefaultRegistry(t *testing.T) {
	t.Run("Package Level Functions", func(t *testing.T) {
		// Test that package-level functions work
		runners := ListRunners()
		assert.IsType(t, []string{}, runners)

		// If itsu is registered, we should be able to create it
		// This test will pass once itsu registration is working
		defaultReg := GetDefaultRegistry()
		assert.NotNil(t, defaultReg)
	})
}
