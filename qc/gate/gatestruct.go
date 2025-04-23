package gate

// NOTE: This file defines GateStruct and associated constructors.
// This appears to be an older or alternative representation compared to the
// Gate interface and the singleton instances defined in builtin.go.
// Consider removing this file and circuit/circuitstruct.go if the DAG-based
// approach (using the Gate interface and circuit.Circuit) is the primary method.

// -----------------------
type (
	// GateType is the type of a quantum gate.
	gateType string
	// GateStruct is a quantum gate.
	// targets and controls are distinct qubit indices.
	GateStruct struct {
		Type     gateType `json:"name"`
		Targets  []int    `json:"targets"`
		Controls []int    `json:"controls"`
	}
)

const (
	HGate       gateType = "H"
	XGate       gateType = "X"
	CNotGate    gateType = "CNOT"
	ToffoliGate gateType = "Toffoli"
	ZGate       gateType = "Z"
	CZGate      gateType = "CZ"
	Measurement gateType = "M"
	SwapGate    gateType = "SWAP"
	FredkinGate gateType = "Fredkin"
)

// NewXGate returns a new XGate.
func NewXGate(target int) *GateStruct {
	return &GateStruct{
		Type:    XGate,
		Targets: []int{target},
	}
}

// NewHGate returns a new HGate.
func NewHGate(target int) *GateStruct {
	return &GateStruct{
		Type:    HGate,
		Targets: []int{target},
	}
}

// NewZGate returns a new ZGate.
func NewZGate(target int) *GateStruct {
	return &GateStruct{
		Type:    ZGate,
		Targets: []int{target},
	}
}

// NewMeasurement returns a new Measurement.
func NewMeasurement(target int) *GateStruct {
	return &GateStruct{
		Type:    Measurement,
		Targets: []int{target},
	}
}

// NewCNotGate returns a new CNotGate.
func NewCNotGate(control int, target int) *GateStruct {
	return &GateStruct{
		Type:     CNotGate,
		Targets:  []int{target},
		Controls: []int{control},
	}
}

// NewCZGate returns a new CZGate.
func NewCZGate(control int, target int) *GateStruct {
	return &GateStruct{
		Type:     CZGate,
		Targets:  []int{target},
		Controls: []int{control},
	}
}

// NewToffoliGate returns a new TofoliGate.
func NewToffoliGate(control0 int, control1 int, target int) *GateStruct {
	return &GateStruct{
		Type:     ToffoliGate,
		Targets:  []int{target},
		Controls: []int{control0, control1},
	}
}

// NewSwapGate returns a new SwapGate.
func NewSwapGate(target0 int, target1 int) *GateStruct {
	return &GateStruct{
		Type:    SwapGate,
		Targets: []int{target0, target1},
	}
}

// NewFredkinGate returns a new FredkinGate.
func NewFredkinGate(control int, target0 int, target1 int) *GateStruct {
	return &GateStruct{
		Type:     FredkinGate,
		Targets:  []int{target0, target1},
		Controls: []int{control},
	}
}
