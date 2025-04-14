package qprog

type gateType string

const (
	HGate       gateType = "H"
	XGate       gateType = "X"
	CNotGate    gateType = "CNot"
	ToffoliGate gateType = "Toffoli"
	ZGate       gateType = "Z"
	CZGate      gateType = "CZ"
	Measurement gateType = "M"
)

// NewXGate returns a new XGate.
func NewXGate(target int) *Gate {
	return &Gate{
		Type:    XGate,
		Targets: []int{target},
	}
}

// NewHGate returns a new HGate.
func NewHGate(target int) *Gate {
	return &Gate{
		Type:    HGate,
		Targets: []int{target},
	}
}

// NewZGate returns a new ZGate.
func NewZGate(target int) *Gate {
	return &Gate{
		Type:    ZGate,
		Targets: []int{target},
	}
}

// NewMeasurement returns a new Measurement.
func NewMeasurement(target int) *Gate {
	return &Gate{
		Type:    Measurement,
		Targets: []int{target},
	}
}

// NewCNotGate returns a new CNotGate.
func NewCNotGate(control int, target int) *Gate {
	return &Gate{
		Type:     CNotGate,
		Targets:  []int{target},
		Controls: []int{control},
	}
}

// NewCZGate returns a new CZGate.
func NewCZGate(control int, target int) *Gate {
	return &Gate{
		Type:     CZGate,
		Targets:  []int{target},
		Controls: []int{control},
	}
}

// NewToffoliGate returns a new TofoliGate.
func NewToffoliGate(control0 int, control1 int, target int) *Gate {
	return &Gate{
		Type:     ToffoliGate,
		Targets:  []int{target},
		Controls: []int{control0, control1},
	}
}
