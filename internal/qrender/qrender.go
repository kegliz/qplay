package qrender

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"os"

	"kegnet.dev/qplay/internal/qprog"
)

type Renderer struct {
	imageWidth  int
	imageHeight int
	//lineHeight  int
	lineWidth   int
	lineSpacing int
	topY        int // Starting position for the first line and text
	lineOffsetX int // Intendation for the lines
	textOffsetX int // Intendation for the text
	fontSize    int
	gateSpace   int
	gateSize    int
	inputText   string
}

// NewDefaultQRenderer creates a new QRenderer with default values
func NewDefaultQRenderer() *Renderer {
	return &Renderer{
		imageWidth:  300,
		imageHeight: 300,
		//lineHeight:  10,
		lineWidth:   240,
		lineSpacing: 40,
		topY:        20,
		lineOffsetX: 30,
		textOffsetX: 5,
		fontSize:    20,
		gateSpace:   10,
		gateSize:    30,
		inputText:   "|0>",
	}
}

// RenderCircuit renders a circuit
func (qr Renderer) RenderCircuit(p *qprog.Program) *image.RGBA {
	if p.NumOfQubits > 0 {
		qr.imageHeight = qr.topY + p.NumOfQubits*qr.lineSpacing
	}

	img := image.NewRGBA(image.Rect(0, 0, qr.imageWidth, qr.imageHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	if p.NumOfQubits == 0 {
		return img
	}
	// Starting position for the first line and text
	yPosition := qr.topY

	// drawing the lines
	for i := 0; i < p.NumOfQubits; i++ {
		lineStart := image.Pt(qr.lineOffsetX, yPosition)
		lineEnd := image.Pt(qr.lineOffsetX+qr.lineWidth, yPosition)
		qr.drawLine(img, lineStart, lineEnd, color.Black)
		qr.drawText(img, image.Pt(qr.textOffsetX, yPosition+5), color.Black, qr.inputText)
		yPosition += qr.lineSpacing
	}

	//drawing the gates (H, X)
	for i, step := range p.Steps {
		for _, gate := range step.Gates {
			switch gate.Type {
			case qprog.HGate:
				qr.drawHGate(img, gate.Targets[0], i)
			case qprog.XGate:
				qr.drawXGate(img, gate.Targets[0], i)
			}
		}
	}
	return img
}

// SaveImage saves an image to a file
func SaveImage(img *image.RGBA, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("cannot create circuit.png: %v", err)
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		return fmt.Errorf("cannot encode png: %v", err)
	}
	return nil
}

// drawXGate draws a blue rectangle with an X in the center of it
func (qr Renderer) drawXGate(img *image.RGBA, target int, step int) {
	qr.drawOneQubitOneCharGate(img, target, step, "X")
}

// drawHGate draws a blue rectangle with an H in the center of it
func (qr Renderer) drawHGate(img *image.RGBA, target int, step int) {
	qr.drawOneQubitOneCharGate(img, target, step, "H")
}

// drawRectWithText draws a blue rectangle with a one-char-long text in the center of it
func (qr Renderer) drawOneQubitOneCharGate(img *image.RGBA, target int, step int, txt string) {
	blue := color.RGBA{0, 0, 255, 255}
	posX := qr.lineOffsetX + qr.gateSpace + step*(qr.gateSize+qr.gateSpace)
	posY := qr.topY + target*qr.lineSpacing - qr.gateSize/2
	r := image.Rect(posX, posY, posX+qr.gateSize, posY+qr.gateSize)
	draw.Draw(img, r, &image.Uniform{blue}, image.Point{}, draw.Src)
	xPos := (r.Min.X + r.Max.X) / 2
	yPos := (r.Min.Y + r.Max.Y) / 2

	qr.drawTextAroundCenter(img, xPos, yPos, color.White, txt)
}

// drawText draws a text on the image
func (qr Renderer) drawText(img *image.RGBA, p image.Point, col color.Color, txt string) {
	point := fixed.P(p.X, p.Y)
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(txt)
}
func (qr Renderer) drawTextAroundCenter(img *image.RGBA, xPos int, yPos int, col color.Color, txt string) {
	//point := fixed.P(p.X, p.Y)
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		//Dot:  point,
	}
	corrXPos := fixed.I(xPos) - d.MeasureString(txt)/2
	textBounds, _ := d.BoundString(txt)
	textHeight := textBounds.Max.Y - textBounds.Min.Y
	corrYPos := fixed.I(yPos + textHeight.Ceil()/2 - 1)

	d.Dot = fixed.Point26_6{
		X: corrXPos,
		Y: corrYPos,
	}
	d.DrawString(txt)
}

// drawLine draws a line on the image
func (qr Renderer) drawLine(img *image.RGBA, start, end image.Point, col color.Color) {
	for x := start.X; x < end.X; x++ {
		img.Set(x, start.Y, col)
	}
}
