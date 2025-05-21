package simulator

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockOneShotRunner is a mock implementation of the OneShotRunner interface for testing.
type mockOneShotRunner struct {
	runOnceFunc func(c circuit.Circuit, callNum int) (string, error)
	callCount   atomic.Int32
	mu          sync.Mutex // To protect runOnceFunc if it's changed mid-test (though not typical)
}

func newMockOneShotRunner(fn func(c circuit.Circuit, callNum int) (string, error)) *mockOneShotRunner {
	return &mockOneShotRunner{
		runOnceFunc: fn,
	}
}

func (m *mockOneShotRunner) RunOnce(c circuit.Circuit) (string, error) {
	callNum := m.callCount.Add(1)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.runOnceFunc != nil {
		return m.runOnceFunc(c, int(callNum))
	}
	return "0", nil // Default success
}

func (m *mockOneShotRunner) CallCount() int {
	return int(m.callCount.Load())
}

func (m *mockOneShotRunner) Reset() {
	m.callCount.Store(0)
}

// helper to create a simple circuit
func newTestCircuit(t *testing.T) circuit.Circuit {
	b := builder.New(builder.Q(1), builder.C(1))
	b.H(0).Measure(0, 0)
	c, err := b.BuildCircuit()
	require.NoError(t, err)
	return c
}

func TestSimulator_RunSerial(t *testing.T) {
	testCirc := newTestCircuit(t)
	shots := 10

	t.Run("Success", func(t *testing.T) {
		mockRunner := newMockOneShotRunner(func(c circuit.Circuit, callNum int) (string, error) {
			if callNum%2 == 0 {
				return "0", nil
			}
			return "1", nil
		})
		sim := NewSimulator(SimulatorOptions{Shots: shots, Runner: mockRunner})

		hist, err := sim.RunSerial(testCirc)
		require.NoError(t, err)
		assert.Equal(t, shots, mockRunner.CallCount())
		assert.Equal(t, shots/2, hist["0"])
		assert.Equal(t, shots/2, hist["1"])
	})

	t.Run("Error", func(t *testing.T) {
		failAtShot := 3
		expectedErr := fmt.Errorf("mock error at shot %d", failAtShot)
		mockRunner := newMockOneShotRunner(func(c circuit.Circuit, callNum int) (string, error) {
			if callNum == failAtShot {
				return "", expectedErr
			}
			return "0", nil
		})
		sim := NewSimulator(SimulatorOptions{Shots: shots, Runner: mockRunner})

		hist, err := sim.RunSerial(testCirc)
		require.Error(t, err)
		assert.ErrorContains(t, err, expectedErr.Error())
		assert.Equal(t, failAtShot, mockRunner.CallCount())
		assert.Equal(t, failAtShot-1, hist["0"]) // Only successful shots before error
	})
}

func TestSimulator_RunParallelStatic(t *testing.T) {
	testCirc := newTestCircuit(t)
	shots := 20 // Choose shots > numCPU for meaningful parallelism
	numWorkers := runtime.NumCPU()
	if numWorkers == 1 {
		numWorkers = 2 // Ensure at least 2 workers for some tests if possible
	}
	if shots < numWorkers { // ensure shots are enough for workers
		shots = numWorkers * 2
	}

	t.Run("Success", func(t *testing.T) {
		mockRunner := newMockOneShotRunner(func(c circuit.Circuit, callNum int) (string, error) {
			if callNum%2 == 0 {
				return "0", nil
			}
			return "1", nil
		})
		sim := NewSimulator(SimulatorOptions{Shots: shots, Workers: numWorkers, Runner: mockRunner})

		hist, err := sim.RunParallelStatic(testCirc)
		require.NoError(t, err)
		assert.Equal(t, shots, mockRunner.CallCount())
		// Exact distribution depends on scheduling, but totals should be correct
		count0 := 0
		count1 := 0
		for i := 1; i <= shots; i++ {
			if i%2 == 0 {
				count0++
			} else {
				count1++
			}
		}
		assert.Equal(t, count0, hist["0"])
		assert.Equal(t, count1, hist["1"])
	})

	t.Run("SingleError", func(t *testing.T) {
		failAfterCalls := shots / 2 // An error occurs mid-way
		expectedErr := fmt.Errorf("mock error on call %d", failAfterCalls)
		var actualErr atomic.Value

		mockRunner := newMockOneShotRunner(func(c circuit.Circuit, callNum int) (string, error) {
			if callNum == failAfterCalls {
				// Store the first error that would be reported by a worker
				actualErr.CompareAndSwap(nil, expectedErr)
				return "", expectedErr
			}
			return "0", nil
		})
		sim := NewSimulator(SimulatorOptions{Shots: shots, Workers: numWorkers, Runner: mockRunner})

		hist, err := sim.RunParallelStatic(testCirc)
		require.Error(t, err)
		// The error returned by RunParallelStatic should be the one from the failing shot.
		assert.EqualError(t, err, actualErr.Load().(error).Error())

		// In RunParallelStatic, if a worker goroutine encounters an error, it returns.
		// Other goroutines might complete their assigned shots.
		// The total call count might be less than 'shots' if an error stops processing early.
		// The exact number of calls can be complex to predict without knowing partitioning.
		// However, it should be at least failAfterCalls.
		assert.GreaterOrEqual(t, mockRunner.CallCount(), failAfterCalls, "Should have at least processed up to the failing call")
		// Hist will contain results from successful shots before the first error was signaled.
		// The content of hist is harder to assert precisely due to parallel nature and early exit.
		// We mainly check that an error was propagated.
		// And that not all shots were necessarily completed.
		if mockRunner.CallCount() < shots {
			t.Logf("RunParallelStatic with error completed %d calls out of %d shots", mockRunner.CallCount(), shots)
		}
		// Check that the histogram contains some results if failAfterCalls > 0
		if failAfterCalls > 1 {
			assert.NotZero(t, hist["0"], "Histogram should have some entries for successful shots")
		}
	})
}

func TestSimulator_RunParallelChan(t *testing.T) {
	testCirc := newTestCircuit(t)
	shots := 20
	numWorkers := runtime.NumCPU()
	if numWorkers == 1 {
		numWorkers = 2
	}
	if shots < numWorkers {
		shots = numWorkers * 2
	}

	t.Run("Success", func(t *testing.T) {
		mockRunner := newMockOneShotRunner(func(c circuit.Circuit, callNum int) (string, error) {
			if callNum%2 == 0 {
				return "0", nil
			}
			return "1", nil
		})
		sim := NewSimulator(SimulatorOptions{Shots: shots, Workers: numWorkers, Runner: mockRunner})

		hist, err := sim.RunParallelChan(testCirc)
		require.NoError(t, err)
		assert.Equal(t, shots, mockRunner.CallCount())
		count0 := 0
		count1 := 0
		for i := 1; i <= shots; i++ {
			if i%2 == 0 {
				count0++
			} else {
				count1++
			}
		}
		assert.Equal(t, count0, hist["0"])
		assert.Equal(t, count1, hist["1"])
	})

	t.Run("SingleError", func(t *testing.T) {
		failAtCall := shots / 2                         // Error occurs part-way through
		expectedErr := fmt.Errorf("mock runonce error") // This is the raw error from the runner

		mockRunner := newMockOneShotRunner(func(c circuit.Circuit, callNum int) (string, error) {
			if callNum == failAtCall {
				return "", expectedErr // mockRunner returns the direct error
			}
			return "0", nil
		})
		sim := NewSimulator(SimulatorOptions{Shots: shots, Workers: numWorkers, Runner: mockRunner})
		// sim.SetVerbose(true) // Enable logging to see worker errors if any

		hist, err := sim.RunParallelChan(testCirc)
		require.Error(t, err, "RunParallelChan should return an error")
		// RunParallelChan wraps the error, so check if the original error is contained.
		assert.ErrorContains(t, err, expectedErr.Error(), "The returned error should contain the original mock error")

		// In RunParallelChan, workers attempt to complete all shots from the job channel.
		// If one shot fails, its error is reported. Other shots should still complete.
		// The histogram should contain results from all successful shots.
		successfulCalls := 0
		for i := 1; i <= shots; i++ {
			if i != failAtCall {
				successfulCalls++
			}
		}

		// Verify histogram content
		if successfulCalls > 0 {
			// Check if hist["0"] exists and matches successfulCalls
			// This assumes all successful calls return "0" as per the mock.
			val, ok := hist["0"]
			assert.True(t, ok, "hist[\"0\"] should exist if there were successful calls")
			assert.Equal(t, successfulCalls, val, "Histogram count for '0' should match the number of successful shots")
		} else {
			assert.Empty(t, hist, "Histogram should be empty if all shots (or the first shot) failed")
		}

		// Call count: All shots should be attempted because jobs are distributed.
		// The failing worker will process one job that errors. Other workers continue.
		assert.Equal(t, shots, mockRunner.CallCount(), "All shots should have been attempted by the runners")

		t.Logf("RunParallelChan with error completed %d calls out of %d shots. Hist: %v, Err: %v", mockRunner.CallCount(), shots, hist, err)
	})
}
