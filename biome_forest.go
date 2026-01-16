package main

import rl "github.com/gen2brain/raylib-go/raylib"

// ============================================================================
// BIOME 0: ENCHANTED FOREST - Magical trees, fireflies, mushrooms
// ============================================================================
func (r *Renderer) drawBiomeEnchantedForest() {
	scroll := r.scrollOffset
	time := r.biomeTimer

	// Sky gradient - mystical purple-green
	for y := int32(0); y < 100; y++ {
		t := float32(y) / 100.0
		c := rl.Color{
			R: uint8(15 + t*20),
			G: uint8(25 + t*30),
			B: uint8(35 + t*25),
			A: 255,
		}
		rl.DrawLine(0, y, screenWidth, y, c)
	}

	// Twinkling stars with glow
	starScroll := int32(scroll * 0.02)
	stars := []struct{ x, y int32 }{
		{40, 15}, {90, 25}, {150, 12}, {210, 30}, {270, 18}, {310, 35},
	}
	for i, star := range stars {
		sx := (star.x - starScroll + screenWidth) % screenWidth
		// Twinkle effect
		twinkle := uint8(180 + 75*simpleSinF(float64(time*3+float32(i))))
		rl.DrawPixel(sx, star.y, rl.Color{R: twinkle, G: twinkle, B: uint8(float32(twinkle) * 0.8), A: 255})
	}

	// Distant misty trees (very slow, faded)
	distantScroll := int32(scroll * 0.1)
	mistColor := rl.Color{R: 30, G: 45, B: 50, A: 255}
	for base := int32(-100); base < screenWidth+100; base += 100 {
		x := base - distantScroll%100
		r.drawMagicTree(x+30, 105, 35, mistColor, false)
		r.drawMagicTree(x+70, 100, 40, mistColor, false)
	}

	// Fog layer 1
	fogAlpha := uint8(40 + 20*simpleSinF(float64(time*0.5)))
	rl.DrawRectangle(0, 85, screenWidth, 25, rl.Color{R: 60, G: 80, B: 70, A: fogAlpha})

	// Mid trees (medium parallax)
	midScroll := int32(scroll * 0.4)
	midColor := rl.Color{R: 25, G: 55, B: 40, A: 255}
	midColor2 := rl.Color{R: 20, G: 45, B: 35, A: 255}
	for base := int32(-120); base < screenWidth+120; base += 120 {
		x := base - midScroll%120
		r.drawMagicTree(x+20, 135, 45, midColor, false)
		r.drawMagicTree(x+80, 130, 50, midColor2, true) // Glowing tree
	}

	// Fog layer 2
	rl.DrawRectangle(0, 120, screenWidth, 20, rl.Color{R: 50, G: 70, B: 60, A: 30})

	// Foreground trees with details
	fgScroll := int32(scroll * 0.8)
	fgColor := rl.Color{R: 20, G: 50, B: 35, A: 255}
	for base := int32(-150); base < screenWidth+150; base += 150 {
		x := base - fgScroll%150
		r.drawMagicTree(x+40, 158, 25, fgColor, false)
		r.drawMagicTree(x+110, 155, 30, fgColor, true)
		// Mushrooms at tree bases
		r.drawMushroom(x+35, 162, rl.Color{R: 200, G: 80, B: 80, A: 255})
		r.drawMushroom(x+120, 160, rl.Color{R: 80, G: 150, B: 200, A: 255})
	}

	// Ground - mossy forest floor
	groundScroll := int32(scroll * 1.0)
	rl.DrawRectangle(0, 160, screenWidth, 40, rl.Color{R: 25, G: 45, B: 30, A: 255})

	// Grass and ferns
	for base := int32(-60); base < screenWidth+60; base += 60 {
		gx := base - groundScroll%60
		r.drawFern(gx+10, 162, time)
		r.drawFern(gx+40, 164, time+1)
		r.drawGrass(gx+25, 163)
		r.drawGrass(gx+55, 161)
	}

	// Path - dirt trail
	rl.DrawRectangle(0, 168, screenWidth, 18, rl.Color{R: 50, G: 40, B: 30, A: 255})
	for base := int32(-50); base < screenWidth+50; base += 50 {
		px := base - groundScroll%50
		// Pebbles
		rl.DrawPixel(px+15, 175, rl.Color{R: 70, G: 60, B: 50, A: 255})
		rl.DrawPixel(px+30, 178, rl.Color{R: 65, G: 55, B: 45, A: 255})
		rl.DrawPixel(px+45, 173, rl.Color{R: 75, G: 65, B: 55, A: 255})
	}

	// Fireflies! (floating particles)
	r.drawFireflies(scroll, time)
}

// --- ENCHANTED FOREST ELEMENTS ---

func (r *Renderer) drawMagicTree(x, baseY, height int32, color rl.Color, glowing bool) {
	// Trunk
	trunkW := height / 8
	if trunkW < 2 {
		trunkW = 2
	}
	trunkH := height / 3
	trunkColor := rl.Color{R: 45, G: 35, B: 30, A: 255}
	rl.DrawRectangle(x-trunkW/2, baseY-trunkH, trunkW, trunkH, trunkColor)

	// Foliage - layered circles for organic look
	foliageY := baseY - trunkH
	layers := height * 2 / 3 / 8
	if layers < 2 {
		layers = 2
	}
	for i := int32(0); i < layers; i++ {
		layerY := foliageY - i*6
		layerW := height/2 - i*3
		if layerW < 4 {
			layerW = 4
		}
		layerColor := color
		if i > 0 {
			layerColor.R = uint8(min(255, int(color.R)+int(i)*5))
			layerColor.G = uint8(min(255, int(color.G)+int(i)*8))
		}
		// Draw as overlapping ovals
		for dy := int32(0); dy < 6; dy++ {
			w := layerW * (6 - dy) / 6
			rl.DrawRectangle(x-w/2, layerY-dy, w, 1, layerColor)
		}
	}

	// Glow effect
	if glowing {
		glowColor := rl.Color{R: 100, G: 255, B: 150, A: 40}
		rl.DrawCircle(x, foliageY-height/4, float32(height/3), glowColor)
	}
}

func (r *Renderer) drawMushroom(x, y int32, capColor rl.Color) {
	// Stem
	rl.DrawRectangle(x, y-3, 2, 4, rl.Color{R: 220, G: 210, B: 190, A: 255})
	// Cap
	rl.DrawRectangle(x-2, y-5, 6, 3, capColor)
	// Spots
	rl.DrawPixel(x-1, y-4, rl.Color{R: 255, G: 255, B: 255, A: 255})
	rl.DrawPixel(x+2, y-4, rl.Color{R: 255, G: 255, B: 255, A: 255})
}

func (r *Renderer) drawFern(x, y int32, time float32) {
	sway := int32(simpleSinF(float64(time*2+float32(x)*0.1)) * 1)
	fernColor := rl.Color{R: 40, G: 90, B: 50, A: 255}
	// Fronds
	rl.DrawPixel(x+sway, y-3, fernColor)
	rl.DrawPixel(x-1+sway, y-2, fernColor)
	rl.DrawPixel(x+1+sway, y-2, fernColor)
	rl.DrawPixel(x+sway, y-1, fernColor)
	rl.DrawPixel(x-2+sway, y-1, fernColor)
	rl.DrawPixel(x+2+sway, y-1, fernColor)
}

func (r *Renderer) drawFireflies(scroll float32, time float32) {
	for i := 0; i < 8; i++ {
		// Pseudo-random but deterministic positions
		baseX := float64((i*73 + 17) % screenWidth)
		baseY := float64(100 + (i*31)%60)
		// Float around
		fx := baseX + 10*simpleSinF(float64(time)*0.8+float64(i)*1.5)
		fy := baseY + 5*simpleSinF(float64(time)*1.2+float64(i)*0.7)
		// Pulse glow
		alpha := uint8(150 + 105*simpleSinF(float64(time)*3+float64(i)*2))
		// Draw with glow
		rl.DrawPixel(int32(fx), int32(fy), rl.Color{R: 200, G: 255, B: 150, A: alpha})
		rl.DrawPixel(int32(fx)+1, int32(fy), rl.Color{R: 200, G: 255, B: 150, A: alpha / 2})
		rl.DrawPixel(int32(fx), int32(fy)+1, rl.Color{R: 200, G: 255, B: 150, A: alpha / 2})
	}
}
