package simulator

import (
	"runtime"
	"sync"

	"github.com/kegliz/qplay/qc/circuit"
)

// runParallelStatic  (static partition) – workers get equal shot counts, no channels.
func (s *Simulator) RunParallelStatic(c circuit.Circuit) (map[string]int, error) {
	shots := s.Shots
	if shots <= 0 {
		shots = 1024
	}
	workers := s.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > shots {
		workers = shots
	}

	per := shots / workers
	extra := shots % workers // first <extra> workers get +1

	s.log.Info().
		Int("shots", shots).
		Int("workers", workers).
		Int("qubits", c.Qubits()).
		Int("clbits", c.Clbits()).
		Int("depth", c.Depth()).
		Msg("itsu: Starting RunParallelStatic")

	hist := make(map[string]int, shots)
	var mu sync.Mutex
	errChan := make(chan error, 1)

	wg := sync.WaitGroup{}
	for w := range workers {
		cnt := per
		if w < extra {
			cnt++
		}
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for range n {
				key, err := s.runner.RunOnce(c) // Run the circuit once

				if err != nil {
					select { // capture first error
					case errChan <- err:
					default:
					}
					return
				}
				mu.Lock()
				hist[key]++
				mu.Unlock()
			}
		}(cnt)
	}

	wg.Wait()
	close(errChan)

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
		s.log.Info().Int("shots", shots).Msg("itsu: Run finished successfully")
	}

	return hist, firstErr
}
