package qservice

import (
	"image"

	"github.com/kegliz/qplay/internal/qprog"
	"github.com/kegliz/qplay/internal/qrender"
	"github.com/kegliz/qplay/internal/server/logger"
)

type (
	ProgramValue struct {
		Program qprog.Program `json:"program"`
	}
	ProgramIDValue struct {
		ID string `json:"id"`
	}

	RenderResult struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
		Image   string `json:"image"`
	}

	// ServiceOptions are options for constructing a service
	ServiceOptions struct {
		Logger *logger.Logger
		Store  ProgramStore
	}

	Service interface {
		RenderCircuit(log *logger.Logger, id string) (*image.RGBA, error)
		SaveProgram(log *logger.Logger, pv *ProgramValue) (string, error)
	}

	service struct {
		store ProgramStore

		logger *logger.Logger
		qr     *qrender.Renderer
	}
)

// NewService creates a new service.
func NewService(opts ServiceOptions) Service {
	if opts.Logger == nil {
		opts.Logger = logger.NewLogger(logger.LoggerOptions{
			Debug: true,
		})
	}
	if opts.Store == nil {
		opts.Store = NewProgramStore()
	}
	s := service{
		logger: opts.Logger,
		store:  opts.Store,
		qr:     qrender.NewDefaultQRenderer(),
	}
	// some initialization for testing
	// program with 2 qubit - 1 step with 1 gate
	p4 := qprog.Program{
		NumOfQubits: 3,
		Steps: []qprog.Step{
			{
				Gates: []qprog.Gate{
					{Type: qprog.HGate, Targets: []int{0}},
				},
			},
			{
				Gates: []qprog.Gate{
					{Type: qprog.XGate, Targets: []int{1}},
					{Type: qprog.HGate, Targets: []int{2}},
				},
			},
		},
	}
	s.store.(*programStore).programs["test"] = &p4
	return &s
}

// RenderCircuit implements Service.
func (s *service) RenderCircuit(l *logger.Logger, id string) (*image.RGBA, error) {
	l.Debug().Msgf("Rendering circuit with id: " + id + " ...")
	p, err := s.store.GetProgram(id)
	if err != nil {
		return nil, err
	}
	img := s.qr.RenderCircuit(p)
	return img, nil
}

// SaveProgram implements Service.
func (s *service) SaveProgram(l *logger.Logger, pv *ProgramValue) (string, error) {
	l.Debug().Msg("Saving program... ")
	p := qprog.NewProgram(pv.Program.NumOfQubits)

	id, err := s.store.SaveProgram(p)

	return id, err
}
