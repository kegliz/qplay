package renderer

import (
	"image"
	"image/color"

	"github.com/kegliz/qplay/qc/circuit"
)

// Renderer turns a circuit into an immutable image.
// Strategy pattern lets us supply many renderers (PNG, SVG, ASCII…).
type Renderer interface {
	Render(c circuit.Circuit) (image.Image, error)
}

// Defaultsize & look‑n‑feel knobs
var (
	WireColor  = color.Black
	GateFill   = color.White
	GateStroke = color.Black
)
