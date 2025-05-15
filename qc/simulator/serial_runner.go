package simulator

import (
	"fmt"

	"github.com/kegliz/qplay/qc/circuit"
)

// RunSerial executes the circuit serially (one shot after another) and returns
// a histogram mapping classical bit-strings (little-endian) to counts.
// This method provides a simpler, non-concurrent alternative to Run.
func (s *Simulator) RunSerial(c circuit.Circuit) (map[string]int, error) {

	s.log.Info().
		Int("shots", s.Shots).
		Int("qubits", c.Qubits()).
		Int("clbits", c.Clbits()).
		Int("depth", c.Depth()).
		Msg("itsu: Starting RunSerial")

	hist := make(map[string]int)

	for i := range s.Shots {
		key, err := s.runner.RunOnce(c) // Run the circuit once
		if err != nil {
			err = fmt.Errorf("shot %d failed: %w", i+1, err)
			s.log.Error().Err(err).Int("shot", i+1).Msg("itsu: Serial shot failed")
			return hist, err
		}
		hist[key]++
	}

	s.log.Info().Int("shots", s.Shots).Msg("itsu: RunSerial finished successfully")
	return hist, nil
}
