package main

import rl "github.com/gen2brain/raylib-go/raylib"

// ============================================================================
// SHARED DRAWING HELPERS
// ============================================================================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// drawMountain draws a simple triangle mountain
func (r *Renderer) drawMountain(x, baseY, width, height int32, color rl.Color) {
	for row := int32(0); row < height; row++ {
		// Width at this row
		w := width * (height - row) / height
		startX := x - w/2
		rl.DrawRectangle(startX, baseY-row, w, 1, color)
	}
}

// drawHill draws a rounded hill
func (r *Renderer) drawHill(x, baseY, width, height int32, color rl.Color) {
	for row := int32(0); row < height; row++ {
		// Parabolic shape
		t := float32(row) / float32(height)
		w := int32(float32(width) * (1 - t*t))
		startX := x - w/2
		rl.DrawRectangle(startX, baseY-row, w, 1, color)
	}
}

// drawTree draws a simple pixel tree
func (r *Renderer) drawTree(x, baseY, height int32, color rl.Color) {
	// Trunk
	trunkColor := rl.Color{R: 60, G: 45, B: 35, A: 255}
	rl.DrawRectangle(x, baseY-height/3, 2, height/3, trunkColor)

	// Foliage - triangle
	for row := int32(0); row < height*2/3; row++ {
		w := (height*2/3 - row) * 2 / 3
		if w < 1 {
			w = 1
		}
		rl.DrawRectangle(x+1-w/2, baseY-height/3-row, w, 1, color)
	}
}

// drawGrass draws small grass tufts
func (r *Renderer) drawGrass(x, y int32) {
	grassColor := rl.Color{R: 50, G: 80, B: 45, A: 255}
	rl.DrawPixel(x, y, grassColor)
	rl.DrawPixel(x+1, y-1, grassColor)
	rl.DrawPixel(x+2, y, grassColor)
}
