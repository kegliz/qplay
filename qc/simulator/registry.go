package simulator

import (
	"fmt"
	"sync"
)

// RunnerFactory is a function that creates a new OneShotRunner instance.
type RunnerFactory func() OneShotRunner

// RunnerRegistry manages the registration and creation of quantum backend runners.
type RunnerRegistry struct {
	mu        sync.RWMutex
	factories map[string]RunnerFactory
}

// Global registry instance
var defaultRegistry = NewRunnerRegistry()

// NewRunnerRegistry creates a new runner registry.
func NewRunnerRegistry() *RunnerRegistry {
	return &RunnerRegistry{
		factories: make(map[string]RunnerFactory),
	}
}

// Register registers a runner factory with the given name.
// This function is thread-safe and can be called from init() functions.
func (r *RunnerRegistry) Register(name string, factory RunnerFactory) error {
	if name == "" {
		return fmt.Errorf("runner name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("runner factory cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("runner %q is already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// MustRegister is like Register but panics if the registration fails.
// This is typically used in init() functions where registration failures
// should be fatal.
func (r *RunnerRegistry) MustRegister(name string, factory RunnerFactory) {
	if err := r.Register(name, factory); err != nil {
		panic(fmt.Sprintf("failed to register runner %q: %v", name, err))
	}
}

// Create creates a new runner instance using the factory registered under the given name.
func (r *RunnerRegistry) Create(name string) (OneShotRunner, error) {
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown runner: %q", name)
	}

	runner := factory()
	if runner == nil {
		return nil, fmt.Errorf("runner factory for %q returned nil", name)
	}

	return runner, nil
}

// ListRunners returns a list of all registered runner names.
func (r *RunnerRegistry) ListRunners() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// Unregister removes a runner from the registry.
// This is primarily useful for testing.
func (r *RunnerRegistry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.factories[name]
	if exists {
		delete(r.factories, name)
	}
	return exists
}

// Package-level convenience functions that operate on the default registry

// RegisterRunner registers a runner factory with the default registry.
func RegisterRunner(name string, factory RunnerFactory) error {
	return defaultRegistry.Register(name, factory)
}

// MustRegisterRunner is like RegisterRunner but panics on failure.
func MustRegisterRunner(name string, factory RunnerFactory) {
	defaultRegistry.MustRegister(name, factory)
}

// CreateRunner creates a runner using the default registry.
func CreateRunner(name string) (OneShotRunner, error) {
	return defaultRegistry.Create(name)
}

// ListRunners returns all registered runner names from the default registry.
func ListRunners() []string {
	return defaultRegistry.ListRunners()
}

// GetDefaultRegistry returns the default runner registry.
// This is useful for advanced use cases or testing.
func GetDefaultRegistry() *RunnerRegistry {
	return defaultRegistry
}
