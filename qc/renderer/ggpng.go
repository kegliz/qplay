package renderer

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"

	"github.com/fogleman/gg" // ✱ pure‑Go 2‑D vector lib :contentReference[oaicite:0]{index=0}
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/kegliz/qplay/qc/gate"
)

// ─── ggPNG renderer ──────────────────────────────────────────────────────
// GGPNG is a renderer that uses the gg library to create PNG images of quantum circuits.
// It draws the circuit operations and wires based on the provided circuit data.

type GGPNG struct{ Cell float64 }

// NewRenderer returns a renderer that emits lossless PNGs using gg.
func NewRenderer(cellPx int) GGPNG { return GGPNG{Cell: float64(cellPx)} }

func (r GGPNG) Render(c circuit.Circuit) (image.Image, error) {
	// Ensure minimum width for drawing wires even if circuit is empty (MaxStep = -1)
	steps := c.MaxStep() + 1
	if steps < 1 {
		steps = 1 // Minimum 1 step width to show wires
	}
	w := int(float64(steps) * r.Cell)
	h := int(float64(c.Qubits()) * r.Cell)

	// Handle edge cases
	if h <= 0 {
		h = int(r.Cell) // Minimum height if no qubits
	}

	dc := gg.NewContext(w, h)
	dc.SetRGB(1, 1, 1) // white background
	dc.Clear()

	// — wires
	dc.SetRGB(0, 0, 0)
	dc.SetLineWidth(1) // Set default line width
	for i := 0; i < c.Qubits(); i++ {
		y := r.y(i)
		dc.DrawLine(0, y, float64(w), y)
		dc.Stroke()
	}

	// Process operations using calculated TimeStep and Line
	for _, op := range c.Operations() {
		// Handle standard single-qubit box gates first
		switch op.G.Name() {
		case "H", "X", "Y", "Z", "S":
			r.drawBoxGate(dc, op)
			continue // Move to next operation
		}

		// Handle multi-qubit and special gates
		switch op.G.Name() {
		case "CNOT":
			r.drawCNOT(dc, op)
		case "CZ": // Added CZ case
			r.drawCZ(dc, op)
		case "FREDKIN":
			r.drawFredkin(dc, op)
		case "SWAP":
			// The DAG builder and circuit.FromDAG handle layout and order.
			// The isTopSwap check is no longer needed.
			r.drawSwap(dc, op) // Draw the full swap symbol (crosses and line)
		case "TOFFOLI":
			r.drawToffoli(dc, op)
		case "MEASURE":
			r.drawMeasurement(dc, op)
		default:
			// Attempt to draw any other unrecognized single-qubit gate as a box
			if g, ok := op.G.(gate.Gate); ok && g.QubitSpan() == 1 {
				fmt.Printf("Renderer warning: Drawing unknown gate '%s' as a default box.\n", g.Name())
				r.drawBoxGate(dc, op)
			} else {
				return nil, fmt.Errorf("renderer: unsupported or unknown gate type '%s'", op.G.Name())
			}
		}
	}

	return dc.Image(), nil
}

func (r GGPNG) Save(path string, c circuit.Circuit) error { // Changed CircuitStruct to Circuit
	img, err := r.Render(c)
	if err != nil {
		return err
	}
	f, err := os.Create(path) // Check error on create
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// ─── helpers ──────────────────────────────────────────────────────────────

func (r GGPNG) x(step int) float64 { return float64(step)*r.Cell + r.Cell/2 }
func (r GGPNG) y(line int) float64 { return float64(line)*r.Cell + r.Cell/2 }

func (r GGPNG) drawBoxGate(dc *gg.Context, op circuit.Operation) {
	// Assumes op.Line is the target qubit for single-qubit gates
	if op.Line < 0 {
		return
	} // Skip if no line associated
	x, y := r.x(op.TimeStep), r.y(op.Line)
	size := r.Cell * .7
	dc.DrawRectangle(x-size/2, y-size/2, size, size)
	dc.SetRGB(1, 1, 1) // White fill
	dc.FillPreserve()
	dc.SetRGB(0, 0, 0) // Black stroke
	dc.SetLineWidth(1) // Ensure consistent line width
	dc.Stroke()
	// Use DrawSymbol() for the text
	dc.DrawStringAnchored(op.G.DrawSymbol(), x, y, 0.5, 0.5)
}

func (r GGPNG) drawToffoli(dc *gg.Context, op circuit.Operation) {
	if len(op.Qubits) != 3 {
		fmt.Printf("Renderer warning: TOFFOLI gate at step %d does not have 3 qubits: %v\n", op.TimeStep, op.Qubits)
		return
	} // Expect 3 qubits for Toffoli
	col := op.TimeStep
	// Qubits might not be contiguous, sort them to identify relative roles
	// Assuming standard Toffoli: controls on first two, target on last relative to op.Line
	// Let's rely on op.Line being the first control qubit index from the builder
	// And assume the builder added qubits in order [control1, control2, target]
	// This requires the builder to enforce this order or the gate definition to be clear.
	// Based on gate.Toffoli() Controls=[0,1], Target=[2] relative indices.
	// op.Qubits = [abs_q_ctrl1, abs_q_ctrl2, abs_q_target]
	// op.Line = abs_q_ctrl1 (minimum)

	// Use absolute qubit indices directly from op.Qubits
	// Assuming the builder added qubits in the order [control1, control2, target]
	// This assumption needs to be consistent with how the builder adds the gate.
	// Let's check the builder: b.add3(gate.Toffoli(), a, bq, t) -> []int{a, bq, t}
	// And gate.Toffoli() defines Controls=[0, 1], Targets=[2] relative to that slice.
	ctrl1Line := op.Qubits[0]  // Absolute line for control 1
	ctrl2Line := op.Qubits[1]  // Absolute line for control 2
	targetLine := op.Qubits[2] // Absolute line for target

	x := r.x(col)
	// Draw controls (●)
	dc.SetRGB(0, 0, 0)
	dc.DrawCircle(x, r.y(ctrl1Line), r.Cell*0.12)
	dc.Fill()
	dc.DrawCircle(x, r.y(ctrl2Line), r.Cell*0.12)
	dc.Fill()

	// Vertical line connecting the involved qubits
	minLine := min(ctrl1Line, ctrl2Line, targetLine)
	maxLine := max(ctrl1Line, ctrl2Line, targetLine)
	dc.DrawLine(x, r.y(minLine), x, r.y(maxLine))
	dc.Stroke()

	// Target ⊕
	targetY := r.y(targetLine)
	dc.DrawCircle(x, targetY, r.Cell*0.18)
	dc.Stroke()
	dc.DrawLine(x-r.Cell*0.18, targetY, x+r.Cell*0.18, targetY)
	dc.Stroke()
	dc.DrawLine(x, targetY-r.Cell*0.18, x, targetY+r.Cell*0.18)
	dc.Stroke()
}

func (r GGPNG) drawMeasurement(dc *gg.Context, op circuit.Operation) {
	// Assumes op.Line is the target qubit
	if op.Line < 0 {
		return
	}
	x, y := r.x(op.TimeStep), r.y(op.Line)
	rad := r.Cell * 0.25
	dc.SetRGB(0, 0, 0)
	dc.NewSubPath()
	dc.DrawArc(x, y, rad, math.Pi, 2*math.Pi)
	dc.ClosePath()
	dc.Stroke() // Draw the arc border
	// Draw the needle
	dc.MoveTo(x, y)
	dc.LineTo(x+rad*0.8, y-rad*0.8)
	dc.Stroke()
	// Draw the label "M"
	dc.DrawStringAnchored("M", x+rad*1.6, y-rad*0.4, 0.0, 0.5)
}

func (r GGPNG) drawCNOT(dc *gg.Context, op circuit.Operation) {
	if len(op.Qubits) != 2 {
		fmt.Printf("Renderer warning: CNOT gate at step %d does not have 2 qubits: %v\n", op.TimeStep, op.Qubits)
		return
	} // Expect 2 qubits
	col := op.TimeStep
	// Based on gate.CNOT() Controls=[0], Target=[1] relative indices.
	// op.Qubits = [abs_q_control, abs_q_target]
	// Builder: b.add2(gate.CNOT(), c, t) -> []int{c, t}
	controlLine := op.Qubits[0]
	targetLine := op.Qubits[1]

	x := r.x(col)
	// Control ●
	dc.SetRGB(0, 0, 0)
	dc.DrawCircle(x, r.y(controlLine), r.Cell*0.12)
	dc.Fill()

	// Vertical wire connecting control and target
	dc.DrawLine(x, r.y(controlLine), x, r.y(targetLine))
	dc.Stroke()

	// Target ⊕
	targetY := r.y(targetLine)
	dc.DrawCircle(x, targetY, r.Cell*0.18)
	dc.Stroke()
	dc.DrawLine(x-r.Cell*0.18, targetY, x+r.Cell*0.18, targetY)
	dc.Stroke()
	dc.DrawLine(x, targetY-r.Cell*0.18, x, targetY+r.Cell*0.18)
	dc.Stroke()
}

// drawCZ draws the Controlled-Z gate.
// It consists of a control dot and a target dot connected by a vertical line.
func (r GGPNG) drawCZ(dc *gg.Context, op circuit.Operation) {
	if len(op.Qubits) != 2 {
		fmt.Printf("Renderer warning: CZ gate at step %d does not have 2 qubits: %v\n", op.TimeStep, op.Qubits)
		return
	} // Expect 2 qubits
	col := op.TimeStep
	// Based on gate.CZ() Controls=[0], Target=[1] relative indices.
	// op.Qubits = [abs_q_control, abs_q_target]
	// Builder: b.add2(gate.CZ(), c, t) -> []int{c, t}
	controlLine := op.Qubits[0]
	targetLine := op.Qubits[1]

	x := r.x(col)
	yCtrl := r.y(controlLine)
	yTgt := r.y(targetLine)

	// Control dot ●
	dc.SetRGB(0, 0, 0)
	dc.DrawCircle(x, yCtrl, r.Cell*0.12)
	dc.Fill()

	// Target dot ●
	dc.DrawCircle(x, yTgt, r.Cell*0.12)
	dc.Fill()

	// Vertical wire connecting control and target
	dc.DrawLine(x, yCtrl, x, yTgt)
	dc.Stroke()
}

func (r GGPNG) drawSwap(dc *gg.Context, op circuit.Operation) {
	if len(op.Qubits) != 2 {
		fmt.Printf("Renderer warning: SWAP gate at step %d does not have 2 qubits: %v\n", op.TimeStep, op.Qubits)
		return
	} // Expect 2 qubits for SWAP
	col := op.TimeStep
	// op.Qubits = [abs_q1, abs_q2]
	// Builder: b.add2(gate.Swap(), q1, q2) -> []int{q1, q2}
	q1Line := op.Qubits[0]
	q2Line := op.Qubits[1]

	x := r.x(col)
	y1 := r.y(q1Line)
	y2 := r.y(q2Line)

	// Draw crosses at both qubit lines
	dc.SetRGB(0, 0, 0)
	r.drawSwapCross(dc, x, y1)
	r.drawSwapCross(dc, x, y2)

	// Draw the vertical connecting line
	dc.SetLineWidth(1) // Ensure consistent line width
	dc.DrawLine(x, y1, x, y2)
	dc.Stroke()
}

func (r GGPNG) drawSwapCross(dc *gg.Context, x, y float64) {
	d := r.Cell * 0.18
	dc.DrawLine(x-d, y-d, x+d, y+d)
	dc.Stroke()
	dc.DrawLine(x-d, y+d, x+d, y-d)
	dc.Stroke()
}

func (r GGPNG) drawFredkin(dc *gg.Context, op circuit.Operation) {
	if len(op.Qubits) != 3 {
		fmt.Printf("Renderer warning: FREDKIN gate at step %d does not have 3 qubits: %v\n", op.TimeStep, op.Qubits)
		return
	} // Expect 3 qubits
	col := op.TimeStep
	// Based on gate.Fredkin() Controls=[0], Targets=[1, 2] relative indices.
	// op.Qubits = [abs_q_control, abs_q_target1, abs_q_target2]
	// Builder: b.add3(gate.Fredkin(), ctrl, t1, t2) -> []int{ctrl, t1, t2}
	controlLine := op.Qubits[0]
	target1Line := op.Qubits[1]
	target2Line := op.Qubits[2]

	x := r.x(col)

	// Control ●
	dc.SetRGB(0, 0, 0)
	dc.DrawCircle(x, r.y(controlLine), r.Cell*0.12)
	dc.Fill()

	// Vertical wire connecting all involved qubits
	minLine := min(controlLine, target1Line, target2Line)
	maxLine := max(controlLine, target1Line, target2Line)
	dc.DrawLine(x, r.y(minLine), x, r.y(maxLine))
	dc.Stroke()

	// Swap crosses on target lines
	r.drawSwapCross(dc, x, r.y(target1Line))
	r.drawSwapCross(dc, x, r.y(target2Line))
}

// Helper min/max for multiple ints
func min(vars ...int) int {
	if len(vars) == 0 {
		panic("min: no arguments")
	}
	minimum := vars[0]
	for _, i := range vars[1:] {
		if i < minimum {
			minimum = i
		}
	}
	return minimum
}

func max(vars ...int) int {
	if len(vars) == 0 {
		panic("max: no arguments")
	}
	maximum := vars[0]
	for _, i := range vars[1:] {
		if i > maximum { // Removed comma
			maximum = i
		}
	}
	return maximum
}
