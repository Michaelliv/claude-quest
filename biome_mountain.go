package main

import rl "github.com/gen2brain/raylib-go/raylib"

// ============================================================================
// BIOME 1: MOUNTAIN JOURNEY - Epic peaks, waterfalls, ancient ruins
// ============================================================================
func (r *Renderer) drawBiomeMountainJourney() {
	scroll := r.scrollOffset
	time := r.biomeTimer

	// Sky gradient - cold blue to warm horizon
	for y := int32(0); y < 100; y++ {
		t := float32(y) / 100.0
		c := rl.Color{
			R: uint8(40 + t*60),
			G: uint8(50 + t*50),
			B: uint8(80 + t*30),
			A: 255,
		}
		rl.DrawLine(0, y, screenWidth, y, c)
	}

	// Sun glow on horizon (draw outermost first, then inner - back to front)
	// Keep sun above the mountain line (y=30 is safe since tallest mountain peaks around y=30)
	sunX := int32(250) - int32(scroll*0.01)%screenWidth
	for i := int32(12); i >= 0; i-- {
		alpha := uint8(50 - i*3)
		rl.DrawCircle(sunX, 45, float32(i+4), rl.Color{R: 255, G: 200, B: 150, A: alpha})
	}

	// Distant snow-capped mountains
	distScroll := int32(scroll * 0.08)
	for base := int32(-250); base < screenWidth+250; base += 250 {
		x := base - distScroll%250
		r.drawSnowMountain(x+50, 100, 100, 70)
		r.drawSnowMountain(x+150, 100, 80, 55)
		r.drawSnowMountain(x+220, 100, 60, 45)
	}

	// Mid mountains with ruins
	midScroll := int32(scroll * 0.25)
	for base := int32(-200); base < screenWidth+200; base += 200 {
		x := base - midScroll%200
		r.drawRockyMountain(x+40, 125, 70, 50)
		r.drawRockyMountain(x+130, 120, 90, 60)
		// Ancient ruins on some peaks
		if base%400 == 0 {
			r.drawRuins(x+130, 65)
		}
	}

	// Waterfall (on certain sections)
	waterfallX := int32(160) - midScroll%400
	if waterfallX > -30 && waterfallX < screenWidth+30 {
		r.drawWaterfall(waterfallX, 70, 60, time)
	}

	// Rocky hills
	hillScroll := int32(scroll * 0.5)
	for base := int32(-150); base < screenWidth+150; base += 150 {
		x := base - hillScroll%150
		r.drawRockyHill(x+30, 145, 50, 30)
		r.drawRockyHill(x+100, 140, 60, 35)
	}

	// Ground - rocky terrain
	groundScroll := int32(scroll * 1.0)
	rl.DrawRectangle(0, 160, screenWidth, 40, rl.Color{R: 60, G: 55, B: 50, A: 255})

	// Rocks and boulders
	for base := int32(-80); base < screenWidth+80; base += 80 {
		gx := base - groundScroll%80
		r.drawBoulder(gx+20, 165, 8)
		r.drawBoulder(gx+60, 168, 5)
		// Small plants
		r.drawAlpinePlant(gx+40, 163)
	}

	// Mountain path - stone
	rl.DrawRectangle(0, 170, screenWidth, 16, rl.Color{R: 75, G: 70, B: 65, A: 255})
	for base := int32(-40); base < screenWidth+40; base += 40 {
		px := base - groundScroll%40
		rl.DrawRectangle(px+5, 172, 12, 8, rl.Color{R: 85, G: 80, B: 75, A: 255})
		rl.DrawRectangle(px+22, 174, 10, 6, rl.Color{R: 70, G: 65, B: 60, A: 255})
	}
}

// --- MOUNTAIN JOURNEY ELEMENTS ---

func (r *Renderer) drawSnowMountain(x, baseY, width, height int32) {
	// Rock base
	rockColor := rl.Color{R: 60, G: 55, B: 70, A: 255}
	snowLine := height * 2 / 3
	for row := int32(0); row < height; row++ {
		w := width * (height - row) / height
		color := rockColor
		if row > snowLine {
			// Snow cap
			color = rl.Color{R: 240, G: 245, B: 255, A: 255}
		} else if row > snowLine-5 {
			// Snow transition
			color = rl.Color{R: 180, G: 190, B: 210, A: 255}
		}
		rl.DrawRectangle(x-w/2, baseY-row, w, 1, color)
	}
}

func (r *Renderer) drawRockyMountain(x, baseY, width, height int32) {
	baseColor := rl.Color{R: 70, G: 65, B: 60, A: 255}
	for row := int32(0); row < height; row++ {
		w := width * (height - row) / height
		// Add some texture variation
		c := baseColor
		if (row+x)%7 == 0 {
			c.R += 10
			c.G += 10
			c.B += 10
		}
		rl.DrawRectangle(x-w/2, baseY-row, w, 1, c)
	}
}

func (r *Renderer) drawRuins(x, y int32) {
	stoneColor := rl.Color{R: 100, G: 95, B: 90, A: 255}
	// Broken pillars
	rl.DrawRectangle(x, y, 4, 15, stoneColor)
	rl.DrawRectangle(x+12, y+5, 3, 10, stoneColor)
	// Archway remains
	rl.DrawRectangle(x+3, y-2, 10, 2, stoneColor)
}

func (r *Renderer) drawWaterfall(x, y, height int32, time float32) {
	// Water stream
	for row := int32(0); row < height; row++ {
		// Animated water
		offset := int32(time*10+float32(row)*0.5) % 3
		alpha := uint8(150 + (row%3)*30)
		rl.DrawRectangle(x+offset, y+row, 3, 2, rl.Color{R: 150, G: 200, B: 255, A: alpha})
	}
	// Splash at bottom
	splashY := y + height
	rl.DrawPixel(x-2, splashY, rl.Color{R: 200, G: 230, B: 255, A: 150})
	rl.DrawPixel(x+5, splashY, rl.Color{R: 200, G: 230, B: 255, A: 150})
}

func (r *Renderer) drawRockyHill(x, baseY, width, height int32) {
	color := rl.Color{R: 65, G: 60, B: 55, A: 255}
	for row := int32(0); row < height; row++ {
		t := float32(row) / float32(height)
		w := int32(float32(width) * (1 - t*t))
		rl.DrawRectangle(x-w/2, baseY-row, w, 1, color)
	}
}

func (r *Renderer) drawBoulder(x, y, size int32) {
	color := rl.Color{R: 80, G: 75, B: 70, A: 255}
	highlight := rl.Color{R: 100, G: 95, B: 90, A: 255}
	rl.DrawRectangle(x, y, size, size-1, color)
	rl.DrawRectangle(x, y, size-1, 1, highlight)
}

func (r *Renderer) drawAlpinePlant(x, y int32) {
	color := rl.Color{R: 60, G: 100, B: 60, A: 255}
	rl.DrawPixel(x, y-2, color)
	rl.DrawPixel(x-1, y-1, color)
	rl.DrawPixel(x+1, y-1, color)
	rl.DrawPixel(x, y, color)
}
