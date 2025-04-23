package drawutil

import (
	"image"
	"image/color"
	"image/draw"
)

func Line(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	// very small Bresenham
	dx, dy := abs(x2-x1), abs(y2-y1)
	sx, sy := sign(x2-x1), sign(y2-y1)
	err := dx - dy
	for {
		img.Set(x1, y1, col)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func GateBox(img *image.RGBA, x, y, w, h int, text string, fill, stroke color.Color) {
	// naïve rectangle + centered rune (ASCII)
	rect := image.Rect(x, y, x+w, y+h)
	draw.Draw(img, rect, &image.Uniform{fill}, image.Point{}, draw.Src)
	// border
	for i := 0; i < w; i++ {
		img.Set(x+i, y, stroke)
		img.Set(x+i, y+h-1, stroke)
	}
	for i := 0; i < h; i++ {
		img.Set(x, y+i, stroke)
		img.Set(x+w-1, y+i, stroke)
	}
	// very small text – just first rune
	if len(text) == 0 {
		return
	}
	r := rune(text[0])
	img.Set(x+w/2, y+h/2, stroke)
	_ = r // placeholder – swap in a real tiny‑font renderer if desired
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
func sign(a int) int {
	switch {
	case a < 0:
		return -1
	case a > 0:
		return 1
	default:
		return 0
	}
}
