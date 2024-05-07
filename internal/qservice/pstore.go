package qservice

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/kegliz/qplay/internal/qprog"
)

type (
	// ProgramStore is an interface for storing programs.
	ProgramStore interface {
		// SaveProgram saves a program and returns its id.
		SaveProgram(p *qprog.Program) (string, error)

		// GetProgram returns a program with the given id.
		GetProgram(id string) (*qprog.Program, error)
	}

	// programStore is an in-memory implementation of ProgramStore.
	programStore struct {
		programs map[string]*qprog.Program
		sync.RWMutex
	}
)

// NewProgramStore creates a new program store.
func NewProgramStore() ProgramStore {
	return &programStore{
		programs: make(map[string]*qprog.Program),
	}
}

// SaveProgram implements ProgramStore.
func (ps *programStore) SaveProgram(p *qprog.Program) (string, error) {
	err := p.Check()
	if err != nil {
		return "", fmt.Errorf("program check failed: %w", err)
	}
	id := uuid.New().String()
	ps.Lock()
	ps.programs[id] = p
	ps.Unlock()
	return id, nil
}

// GetProgram implements ProgramStore.
func (ps *programStore) GetProgram(id string) (*qprog.Program, error) {
	ps.RLock()
	p, ok := ps.programs[id]
	ps.RUnlock()
	if !ok {
		return nil, fmt.Errorf("program with id %s not found", id)
	}
	return p, nil
}
