package itsu

import (
	"fmt"

	"github.com/itsubaki/q"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/rs/zerolog/log" // Import zerolog's global logger for consistency
)

// RunSerial executes the circuit serially (one shot after another) and returns
// a histogram mapping classical bit-strings (little-endian) to counts.
// This method provides a simpler, non-concurrent alternative to Run.
func (s *Simulator) RunSerial(c circuit.Circuit) (map[string]int, error) {
	// Logging setup is assumed to be done by the caller or a higher-level function
	// if consistent logging across Run and RunSerial is needed.
	// Alternatively, duplicate the zerolog setup from Run here if RunSerial
	// can be called independently without Run's setup.

	shots := s.Shots
	if shots <= 0 {
		shots = 1024 // Default shots
	}

	log.Info().
		Int("shots", shots).
		Int("qubits", c.Qubits()).
		Int("clbits", c.Clbits()).
		Int("depth", c.Depth()).
		Msg("itsu: Starting RunSerial")

	hist := make(map[string]int)

	for i := range shots {
		sim := q.New() // Create a new simulator instance for each shot
		key, err := runOnce(sim, c)
		if err != nil {
			err = fmt.Errorf("shot %d failed: %w", i+1, err)
			log.Error().Err(err).Int("shot", i+1).Msg("itsu: Serial shot failed")
			return hist, err
		}
		hist[key]++

		// Optional: Add progress logging similar to the parallel version
		if s.Verbose && ((i+1)%100 == 0 || (i+1) == shots) {
			log.Info().Int("completed", i+1).Int("total", shots).Msgf("itsu: Completed %d/%d shots (serial)", i+1, shots)
		}
	}

	log.Info().Int("shots", shots).Msg("itsu: RunSerial finished successfully")
	return hist, nil // Return the final histogram and nil error
}
