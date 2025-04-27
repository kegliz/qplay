package itsu

import (
	"fmt"
	"os" // Import os package for zerolog output
	"runtime"
	"sync"
	"time" // Import time package

	"github.com/itsubaki/q"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/rs/zerolog"     // Import zerolog
	"github.com/rs/zerolog/log" // Import zerolog's global logger
)

// Simulator executes an immutable circuit for a given number of shots.
// It uses a pool of worker goroutines (Workers==0 → NumCPU) to run shots
// in parallel.  The implementation relies only on public symbols that
// exist in release v0.0.5 of github.com/itsubaki/q, so it compiles out‑of‑
// the box.
type Simulator struct {
	Shots   int
	Workers int  // number of concurrent workers (0 => NumCPU)
	Verbose bool // enable/disable logging
}

// New creates a new Simulator with logging disabled by default.
func New(shots int) *Simulator { return &Simulator{Shots: shots, Verbose: false} }

// SetVerbose enables or disables logging for the simulator.
func (s *Simulator) SetVerbose(verbose bool) {
	s.Verbose = verbose
}

// Run executes the circuit and returns a histogram mapping classical
// bit‑strings (little‑endian) to counts.
func (s *Simulator) Run(c circuit.Circuit) (map[string]int, error) {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs                                       // Compact timestamp
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}) // Human-friendly output

	if s.Verbose {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel) // Log only warnings and errors if not verbose
	}

	shots := s.Shots
	if shots <= 0 {
		shots = 1024
	}

	workers := s.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > shots { // Optimization: Don't start more workers than shots
		workers = shots
	}

	log.Info().
		Int("shots", shots).
		Int("workers", workers).
		Int("qubits", c.Qubits()).
		Int("clbits", c.Clbits()).
		Int("depth", c.Depth()).
		Msg("itsu: Starting Run")

	hist := make(map[string]int)
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	errChan := make(chan error, workers) // Channel to collect the first error from each worker

	// fan‑out jobs
	jobs := make(chan struct{}, shots)
	for range shots {
		jobs <- struct{}{}
	}
	close(jobs)

	shotCounter := &struct { // Use a thread-safe counter for progress logging
		sync.Mutex
		count int
	}{}

	for wid := range workers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// log.Debug().Int("worker_id", id).Msg("itsu: Worker starting") // Use Debug for finer-grained logs if needed

			var workerErr error // Track first error for this worker
			workerShotCount := 0

			for range jobs {
				// Skip further processing if this worker already encountered an error
				if workerErr != nil {
					continue
				}

				sim := q.New() // Create a new simulator instance FOR EACH SHOT

				// log.Debug().Int("worker_id", id).Int("shot", workerShotCount+1).Msg("itsu: Worker starting shot")
				start := time.Now() // Time each shot simulation

				key, err := runOnce(sim, c)
				duration := time.Since(start)
				// log.Debug().Int("worker_id", id).Int("shot", workerShotCount+1).Dur("duration", duration).Msg("itsu: Worker finished shot")

				if err != nil {
					// Record the first error encountered by this worker
					// Include shot number and duration for context
					workerErr = fmt.Errorf("worker %d: shot %d failed after %v: %w", id, workerShotCount+1, duration, err)
					log.Error().Err(workerErr).Int("worker_id", id).Int("shot", workerShotCount+1).Msg("itsu: Shot failed")
					continue // Continue to allow other workers to finish
				}

				mu.Lock()
				hist[key]++
				mu.Unlock()

				// Log progress periodically
				shotCounter.Lock()
				shotCounter.count++
				currentCount := shotCounter.count
				shotCounter.Unlock()
				if s.Verbose && (currentCount%100 == 0 || currentCount == shots) { // Log every 100 shots and on the last shot
					log.Info().Int("completed", currentCount).Int("total", shots).Msgf("itsu: Completed %d/%d shots", currentCount, shots)
				}

				workerShotCount++
			}

			// Report the first error encountered by this worker, if any
			if workerErr != nil {
				// Use non-blocking send in case multiple workers error out
				select {
				case errChan <- workerErr:
				default:
					// Log if error couldn't be sent (e.g., channel full)
					log.Warn().Err(workerErr).Int("worker_id", id).Msg("itsu: Worker failed to send error (channel full?)")
				}
			}
			// log.Debug().Int("worker_id", id).Int("processed_shots", workerShotCount).Msg("itsu: Worker finished")
		}(wid)
	}

	log.Info().Msg("itsu: Waiting for workers to finish...")
	wg.Wait()
	log.Info().Msg("itsu: Workers finished.")
	close(errChan) // Close channel after all workers are done

	// Check if any errors were reported
	var firstErr error
	errCount := 0
	for err := range errChan {
		errCount++
		if firstErr == nil {
			firstErr = err // Capture the very first error reported
		}
		// Log additional errors if desired (as Warn or Error level)
		if errCount > 1 {
			log.Warn().Err(err).Int("error_count", errCount).Msg("itsu: Additional error reported")
		}
	}

	if errCount > 0 {
		log.Warn().Err(firstErr).Int("error_count", errCount).Msgf("itsu: Run finished with %d error(s)", errCount)
	} else {
		log.Info().Int("shots", shots).Msg("itsu: Run finished successfully")
	}

	return hist, firstErr // Return histogram and the first error encountered
}

// runOnce plays the circuit exactly one time on the provided simulator,
// returning the measured classical bit‑string.
func runOnce(sim *q.Q, c circuit.Circuit) (string, error) {
	// Reset the simulator state to |0...0> for the number of qubits in the circuit.
	// Note: sim.ZeroWith might implicitly reset, but explicit Reset ensures clean state.
	// Depending on itsubaki/q implementation, Reset might be redundant if ZeroWith does it.
	// If ZeroWith allocates a *new* state vector internally, Reset might not be needed here,
	// but calling it before ZeroWith ensures the simulator object itself is in a known state.
	qs := sim.ZeroWith(c.Qubits())
	// Initialize classical bits to all '0's. Length should match circuit's classical bits.
	cbits := make([]byte, c.Clbits())
	for i := range cbits {
		cbits[i] = '0' // Explicitly initialize to '0'
	}

	for i, op := range c.Operations() { // Add index for logging context in errors
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
		case "S":
			sim.S(qs[op.Qubits[0]])
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
