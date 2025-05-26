package gate

// ---------- immutable value objects ----------------------------------

// simple 1-qubit gate
type u1 struct{ name, symbol string }

func (g u1) Name() string       { return g.name }
func (g u1) QubitSpan() int     { return 1 }
func (g u1) DrawSymbol() string { return g.symbol }
func (g u1) Targets() []int     { return []int{0} } // Target is the only qubit
func (g u1) Controls() []int    { return []int{} }  // No controls

// 2-qubit gate with fixed ASCII symbol (CNOT, SWAP, CZ)
type u2 struct {
	name, symbol      string
	targets, controls []int
}

func (g u2) Name() string       { return g.name }
func (g u2) QubitSpan() int     { return 2 }
func (g u2) DrawSymbol() string { return g.symbol }
func (g u2) Targets() []int     { return g.targets }
func (g u2) Controls() []int    { return g.controls }

// 3-qubit gate (Toffoli, Fredkin)
type u3 struct {
	name, symbol      string
	targets, controls []int
}

func (g u3) Name() string       { return g.name }
func (g u3) QubitSpan() int     { return 3 }
func (g u3) DrawSymbol() string { return g.symbol }
func (g u3) Targets() []int     { return g.targets }
func (g u3) Controls() []int    { return g.controls }

// measurement (1-qubit but special semantic)
type meas struct{}

func (meas) Name() string       { return "MEASURE" }
func (meas) QubitSpan() int     { return 1 }
func (meas) DrawSymbol() string { return "M" }
func (meas) Targets() []int     { return []int{0} } // Target is the only qubit
func (meas) Controls() []int    { return []int{} }  // No controls

// ---------- constructors (singletons) --------------------------------

var (
	hGate  = &u1{"H", "H"}
	xGate  = &u1{"X", "X"}
	yGate  = &u1{"Y", "Y"}
	sGate  = &u1{"S", "S"}
	zGate  = &u1{"Z", "Z"}
	swapG  = &u2{"SWAP", "×", []int{0, 1}, []int{}}     // Targets 0, 1; No controls
	cnotG  = &u2{"CNOT", "⊕", []int{1}, []int{0}}       // Target 1; Control 0
	czGate = &u2{"CZ", "●", []int{1}, []int{0}}         // Target 1; Control 0 (Symbol represents control dot)
	toffG  = &u3{"TOFFOLI", "T", []int{2}, []int{0, 1}} // Target 2; Controls 0, 1
	fredG  = &u3{"FREDKIN", "F", []int{1, 2}, []int{0}} // Targets 1, 2; Control 0
	measG  = &meas{}
)

// Public accessors return the shared immutable value.
// (Reduces allocations and supports pointer equality tricks in passes.)
func H() Gate       { return hGate }
func X() Gate       { return xGate }
func Y() Gate       { return yGate }
func S() Gate       { return sGate }
func Z() Gate       { return zGate }
func Swap() Gate    { return swapG }
func CNOT() Gate    { return cnotG }
func CZ() Gate      { return czGate } // Added CZ accessor
func Toffoli() Gate { return toffG }
func Fredkin() Gate { return fredG }
func Measure() Gate { return measG }
