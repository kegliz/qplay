package simulator

import (
	"runtime"

	"github.com/kegliz/qplay/internal/logger"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/rs/zerolog"
)

// ItsuSimulatorOptions encapsulates the parameters for creating a Simulator.
type SimulatorOptions struct {
	Shots   int
	Workers int // number of concurrent workers (0 => NumCPU)
	Runner  OneShotRunner
}

// Simulator executes an immutable circuit for a given number of shots.
// It uses a pool of worker goroutines (Workers==0 → NumCPU) to run shots
// in parallel.  The implementation relies only on public symbols that
// exist in release v0.0.5 of github.com/itsubaki/q, so it compiles out‑of‑
// the box.
type Simulator struct {
	Shots   int
	Workers int // number of concurrent workers (0 => NumCPU)
	runner  OneShotRunner

	log logger.Logger
}

// NewSimulator creates a new Simulator
func NewSimulator(options SimulatorOptions) *Simulator {
	shots := options.Shots
	if shots <= 0 {
		shots = 1024 // Default shots
	}

	workers := options.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > shots { // Optimization: Don't start more workers than shots
		workers = shots
	}

	return &Simulator{Shots: shots, Workers: workers, runner: options.Runner,
		log: *logger.NewLogger(logger.LoggerOptions{
			Debug: false,
		})}
}

// SetVerbose make the simulator log all messages (debug level).
func (s *Simulator) SetVerbose(verbose bool) {
	if verbose {
		s.log.Logger = s.log.Logger.Level(zerolog.DebugLevel) // Log all messages if verbose
	} else {
		s.log.Logger = s.log.Logger.Level(zerolog.InfoLevel)
	}
}

// OneShotRunner is an interface for running a circuit once.
type OneShotRunner interface {
	// RunOnce executes the circuit for one shot.
	RunOnce(circuit.Circuit) (string, error)
}

// Run defaults to RunParallelStatic.
func (s *Simulator) Run(c circuit.Circuit) (map[string]int, error) {
	return s.RunParallelStatic(c)
}
