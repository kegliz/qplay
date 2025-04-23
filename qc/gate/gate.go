package gate

import "strings"

// Gate is the *minimal* contract each quantum gate must fulfil.
// The interface is tiny on purpose so optimisers and simulators
// can depend on it without pulling in graphical or param APIs.
type Gate interface {
	Name() string       // canonical name e.g. "H", "CNOT"
	QubitSpan() int     // how many qubits it acts on
	DrawSymbol() string // single-char/fallback symbol used by renderers
	Targets() []int     // Relative indices of target qubits (within the span)
	Controls() []int    // Relative indices of control qubits (within the span)
}

// Factory returns an immutable gate by many common aliases.
//
//	g, _ := gate.Factory("cx")  // -> same instance as CNOT()
func Factory(name string) (Gate, error) {
	switch norm(name) {
	case "h":
		return H(), nil
	case "x":
		return X(), nil
	case "s":
		return S(), nil
	case "swap":
		return Swap(), nil
	case "cx", "cnot":
		return CNOT(), nil
	case "t", "toffoli", "ccx":
		return Toffoli(), nil
	case "fredkin", "cswap":
		return Fredkin(), nil
	case "m", "measure", "meas":
		return Measure(), nil
	}
	return nil, ErrUnknownGate{name}
}

// ErrUnknownGate is returned by Factory when the label isn't recognised.
type ErrUnknownGate struct{ Name string }

func (e ErrUnknownGate) Error() string { return "qcircuit: unknown gate " + e.Name }

// helpers --------------------------------------------------------------

func norm(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
