package itsu

import (
	"sync"

	"github.com/itsubaki/q"
	"github.com/kegliz/qplay/internal/logger"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/rs/zerolog"
)

// pool caches *q.Q; each holds a big state slice we want to reuse.
var pool = sync.Pool{New: func() any { return q.New() }}

type PooledItsuOneShotRunner struct {
	log logger.Logger
}

func NewPooledItsuOneShotRunner() *PooledItsuOneShotRunner {
	return &PooledItsuOneShotRunner{
		log: *logger.NewLogger(logger.LoggerOptions{
			Debug: false,
		}),
	}
}
func (s *PooledItsuOneShotRunner) SetVerbose(verbose bool) {
	if verbose {
		s.log.Logger = s.log.Logger.Level(zerolog.DebugLevel) // Log all messages if verbose
	} else {
		s.log.Logger = s.log.Logger.Level(zerolog.InfoLevel)
	}
}

func (s *PooledItsuOneShotRunner) RunOnce(c circuit.Circuit) (string, error) {
	sim := pool.Get().(*q.Q)
	defer pool.Put(sim)
	return runOnce(sim, c)
}
