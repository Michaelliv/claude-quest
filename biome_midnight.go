package main

import rl "github.com/gen2brain/raylib-go/raylib"

// ============================================================================
// BIOME 2: MIDNIGHT QUEST - Starry sky, glowing crystals, mysterious fog
// ============================================================================
func (r *Renderer) drawBiomeMidnightQuest() {
	scroll := r.scrollOffset
	time := r.biomeTimer

	// Deep night sky
	for y := int32(0); y < 100; y++ {
		t := float32(y) / 100.0
		c := rl.Color{
			R: uint8(10 + t*15),
			G: uint8(10 + t*20),
			B: uint8(25 + t*25),
			A: 255,
		}
		rl.DrawLine(0, y, screenWidth, y, c)
	}

	// Many twinkling stars
	starScroll := int32(scroll * 0.02)
	for i := 0; i < 30; i++ {
		sx := int32((i*47+13)%screenWidth) - starScroll%screenWidth
		if sx < 0 {
			sx += screenWidth
		}
		sy := int32((i*31 + 7) % 70)
		twinkle := uint8(150 + 105*simpleSinF(float64(time*2+float32(i)*0.5)))
		size := (i % 3)
		if size == 0 {
			rl.DrawPixel(sx, sy, rl.Color{R: twinkle, G: twinkle, B: 255, A: 255})
		} else {
			rl.DrawRectangle(sx, sy, int32(size), int32(size), rl.Color{R: twinkle, G: twinkle, B: 255, A: 255})
		}
	}

	// Large moon
	moonX := int32(80) - int32(scroll*0.01)%300
	rl.DrawCircle(moonX, 35, 18, rl.Color{R: 220, G: 220, B: 240, A: 255})
	rl.DrawCircle(moonX+4, 33, 15, rl.Color{R: 15, G: 15, B: 30, A: 255}) // Shadow

	// Distant crystal formations
	distScroll := int32(scroll * 0.15)
	for base := int32(-180); base < screenWidth+180; base += 180 {
		x := base - distScroll%180
		r.drawCrystalFormation(x+50, 95, 25, rl.Color{R: 60, G: 40, B: 80, A: 255}, false)
		r.drawCrystalFormation(x+130, 90, 30, rl.Color{R: 50, G: 50, B: 90, A: 255}, false)
	}

	// Mysterious fog
	fogAlpha := uint8(50 + 30*simpleSinF(float64(time*0.3)))
	rl.DrawRectangle(0, 80, screenWidth, 30, rl.Color{R: 40, G: 50, B: 70, A: fogAlpha})

	// Mid crystal spires (some glowing!)
	midScroll := int32(scroll * 0.4)
	for base := int32(-140); base < screenWidth+140; base += 140 {
		x := base - midScroll%140
		r.drawCrystalFormation(x+30, 130, 35, rl.Color{R: 70, G: 50, B: 100, A: 255}, true) // Glowing
		r.drawCrystalFormation(x+90, 125, 40, rl.Color{R: 60, G: 60, B: 110, A: 255}, false)
	}

	// More fog
	rl.DrawRectangle(0, 115, screenWidth, 25, rl.Color{R: 30, G: 40, B: 60, A: 35})

	// Foreground spooky trees
	fgScroll := int32(scroll * 0.7)
	for base := int32(-120); base < screenWidth+120; base += 120 {
		x := base - fgScroll%120
		r.drawSpookyTree(x+40, 158, 30)
		r.drawSpookyTree(x+90, 155, 25)
	}

	// Ground - dark mystical
	groundScroll := int32(scroll * 1.0)
	rl.DrawRectangle(0, 160, screenWidth, 40, rl.Color{R: 25, G: 25, B: 35, A: 255})

	// Glowing mushrooms and crystals on ground
	for base := int32(-70); base < screenWidth+70; base += 70 {
		gx := base - groundScroll%70
		r.drawGlowingMushroom(gx+20, 163, time)
		r.drawSmallCrystal(gx+50, 165, time)
	}

	// Path - ancient stones
	rl.DrawRectangle(0, 170, screenWidth, 16, rl.Color{R: 35, G: 35, B: 45, A: 255})
	for base := int32(-45); base < screenWidth+45; base += 45 {
		px := base - groundScroll%45
		rl.DrawRectangle(px+8, 173, 14, 7, rl.Color{R: 45, G: 45, B: 55, A: 255})
	}

	// Floating magic particles
	r.drawMagicParticles(scroll, time)
}

// --- MIDNIGHT QUEST ELEMENTS ---

func (r *Renderer) drawCrystalFormation(x, baseY, height int32, color rl.Color, glowing bool) {
	// Main crystal
	for row := int32(0); row < height; row++ {
		w := (height - row) / 4
		if w < 1 {
			w = 1
		}
		c := color
		// Lighter at top
		c.R = uint8(min(255, int(color.R)+int(row)))
		c.G = uint8(min(255, int(color.G)+int(row)))
		c.B = uint8(min(255, int(color.B)+int(row)*2))
		rl.DrawRectangle(x-w/2, baseY-row, w, 1, c)
	}
	// Side crystals
	sideH := height * 2 / 3
	for row := int32(0); row < sideH; row++ {
		w := (sideH - row) / 5
		if w < 1 {
			w = 1
		}
		rl.DrawRectangle(x-height/4-w/2, baseY-row, w, 1, color)
		rl.DrawRectangle(x+height/4-w/2, baseY-row, w, 1, color)
	}
	if glowing {
		glowColor := rl.Color{R: 150, G: 100, B: 255, A: 50}
		rl.DrawCircle(x, baseY-height/2, float32(height/2), glowColor)
	}
}

func (r *Renderer) drawSpookyTree(x, baseY, height int32) {
	// Gnarled trunk
	trunkColor := rl.Color{R: 30, G: 25, B: 35, A: 255}
	rl.DrawRectangle(x-2, baseY-height/2, 4, height/2, trunkColor)
	// Twisted branches (no leaves)
	rl.DrawLine(x, baseY-height/2, x-8, baseY-height/2-10, trunkColor)
	rl.DrawLine(x, baseY-height/2, x+6, baseY-height/2-8, trunkColor)
	rl.DrawLine(x-8, baseY-height/2-10, x-12, baseY-height/2-15, trunkColor)
	rl.DrawLine(x+6, baseY-height/2-8, x+10, baseY-height/2-12, trunkColor)
}

func (r *Renderer) drawGlowingMushroom(x, y int32, time float32) {
	pulse := uint8(150 + 100*simpleSinF(float64(time*2+float32(x)*0.1)))
	// Stem
	rl.DrawRectangle(x, y-2, 2, 3, rl.Color{R: 100, G: 80, B: 120, A: 255})
	// Glowing cap
	rl.DrawRectangle(x-1, y-4, 4, 2, rl.Color{R: pulse, G: pulse / 2, B: 255, A: 255})
	// Glow
	rl.DrawPixel(x, y-5, rl.Color{R: pulse, G: pulse, B: 255, A: 100})
}

func (r *Renderer) drawSmallCrystal(x, y int32, time float32) {
	pulse := uint8(180 + 75*simpleSinF(float64(time*3+float32(x)*0.2)))
	rl.DrawPixel(x, y-2, rl.Color{R: pulse / 2, G: pulse, B: 255, A: 255})
	rl.DrawPixel(x, y-1, rl.Color{R: pulse / 2, G: pulse, B: 255, A: 255})
	rl.DrawPixel(x, y, rl.Color{R: 100, G: 100, B: 150, A: 255})
}

func (r *Renderer) drawMagicParticles(scroll float32, time float32) {
	for i := 0; i < 12; i++ {
		baseX := float64((i*61 + 23) % screenWidth)
		baseY := float64(80 + (i*41)%80)
		// Float upward and sway
		fx := baseX + 8*simpleSinF(float64(time)+float64(i))
		fy := baseY - float64(int(time*15+float32(i)*10)%80)
		// Pulse
		alpha := uint8(100 + 100*simpleSinF(float64(time)*2+float64(i)*1.5))
		hue := int(time*50+float32(i)*30) % 360
		color := hsvToRGB(hue, 0.7, 1.0)
		color.A = alpha
		rl.DrawPixel(int32(fx), int32(fy), color)
	}
}
