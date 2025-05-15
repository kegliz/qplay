package simulator

import (
	"fmt"
	"sync"

	"github.com/kegliz/qplay/qc/circuit"
	"github.com/rs/zerolog/log"
)

// RunParallelChan executes the circuit and returns a histogram mapping classical
// bit‑strings (little‑endian) to counts.
func (s *Simulator) RunParallelChan(c circuit.Circuit) (map[string]int, error) {

	// shots and workers are now initialized in New
	s.log.Info().
		Int("shots", s.Shots).
		Int("workers", s.Workers).
		Int("qubits", c.Qubits()).
		Int("clbits", c.Clbits()).
		Int("depth", c.Depth()).
		Msg("itsu: Starting RunParallelChan")

	hist := make(map[string]int)
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	errChan := make(chan error, s.Workers) // Channel to collect the first error from each worker

	// fan‑out jobs
	jobs := make(chan struct{}, s.Shots)
	for range s.Shots {
		jobs <- struct{}{}
	}
	close(jobs)

	for wid := range s.Workers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			var workerErr error // Track first error for this worker

			for range jobs {
				// Skip further processing if this worker already encountered an error
				if workerErr != nil {
					continue
				}

				key, err := s.runner.RunOnce(c) // Run the circuit once

				if err != nil {
					// Record the first error encountered by this worker
					workerErr = fmt.Errorf("worker %d failed: %w", id, err)
					log.Error().Err(workerErr).Int("worker_id", id).Msg("itsu: Shot failed")
					continue // Continue to allow other workers to finish
				}

				mu.Lock()
				hist[key]++
				mu.Unlock()
			}

			// Report the first error encountered by this worker, if any
			if workerErr != nil {
				// Use non-blocking send in case multiple workers error out
				select {
				case errChan <- workerErr:
				default:
					// Log if error couldn't be sent (e.g., channel full)
					s.log.Warn().Err(workerErr).Int("worker_id", id).Msg("itsu: Worker failed to send error (channel full?)")
				}
			}
		}(wid)
	}

	s.log.Debug().Msg("itsu: Waiting for workers to finish...")
	wg.Wait()
	s.log.Info().Msg("itsu: Workers finished.")
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
			s.log.Warn().Err(err).Int("error_count", errCount).Msg("itsu: Additional error reported")
		}
	}

	if errCount > 0 {
		s.log.Warn().Err(firstErr).Int("error_count", errCount).Msgf("itsu: Run finished with %d error(s)", errCount)
	} else {
		s.log.Info().Int("shots", s.Shots).Msg("itsu: RunParallelChan finished successfully")
	}

	return hist, firstErr
}
