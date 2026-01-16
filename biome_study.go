package main

import rl "github.com/gen2brain/raylib-go/raylib"

// ============================================================================
// BIOME 4: WIZARD'S LIBRARY - Endless corridor of books, windows, magic
// ============================================================================
func (r *Renderer) drawBiomeWizardLibrary() {
	scroll := r.scrollOffset
	time := r.biomeTimer

	// === BACK WALL - Rich purple with subtle texture ===
	for y := int32(0); y < 160; y++ {
		t := float32(y) / 160.0
		c := rl.Color{
			R: uint8(30 + t*15),
			G: uint8(25 + t*12),
			B: uint8(42 + t*18),
			A: 255,
		}
		rl.DrawLine(0, y, screenWidth, y, c)
	}

	// Wall texture - subtle brick pattern (very slow parallax)
	brickScroll := int32(scroll * 0.05)
	wallBrick := rl.Color{R: 38, G: 32, B: 50, A: 80}
	for y := int32(10); y < 150; y += 16 {
		offset := int32(0)
		if (y/16)%2 == 1 {
			offset = 20
		}
		for base := int32(-60); base < screenWidth+60; base += 40 {
			x := base - brickScroll%40 + offset
			rl.DrawRectangle(x, y, 38, 14, wallBrick)
		}
	}

	// === DISTANT WINDOWS (slow parallax) ===
	windowScroll := int32(scroll * 0.1)
	for base := int32(-320); base < screenWidth+320; base += 320 {
		wx := base - windowScroll%320
		r.drawLibraryWindow(wx+160, 15, time)
	}

	// === MID LAYER - Hanging chandeliers (same parallax as wall, offset from windows) ===
	chandelierScroll := int32(scroll * 0.05)
	for base := int32(-320); base < screenWidth+320; base += 320 {
		cx := base - chandelierScroll%320
		r.drawChandelier(cx, 25, time) // Windows at +160, chandeliers at 0 (between windows)
	}

	// === BOOKSHELVES (medium parallax) ===
	shelfScroll := int32(scroll * 0.4)
	for base := int32(-220); base < screenWidth+220; base += 220 {
		sx := base - shelfScroll%220
		r.drawTallBookshelf(sx+80, 25, time)
	}

	// === DESKS WITH ITEMS (faster parallax) ===
	deskScroll := int32(scroll * 0.6)
	for base := int32(-350); base < screenWidth+350; base += 350 {
		dx := base - deskScroll%350
		r.drawWizardDesk(dx+175, 90, time)
	}

	// === FLOOR - Warm wooden planks ===
	rl.DrawRectangle(0, 160, screenWidth, 40, rl.Color{R: 55, G: 42, B: 50, A: 255})

	// Floor planks with parallax
	floorScroll := int32(scroll * 1.0)
	plankColors := []rl.Color{
		{R: 65, G: 48, B: 55, A: 255},
		{R: 58, G: 44, B: 52, A: 255},
		{R: 62, G: 46, B: 54, A: 255},
	}
	for base := int32(-50); base < screenWidth+50; base += 42 {
		px := base - floorScroll%42
		i := (base / 42) % 3
		if i < 0 {
			i += 3
		}
		rl.DrawRectangle(px, 160, 40, 40, plankColors[i])
		rl.DrawLine(px, 160, px, 200, rl.Color{R: 45, G: 35, B: 42, A: 255})
		// Wood grain
		rl.DrawLine(px+10, 165, px+12, 195, rl.Color{R: 50, G: 38, B: 45, A: 100})
		rl.DrawLine(px+25, 168, px+28, 198, rl.Color{R: 50, G: 38, B: 45, A: 100})
	}

	// === RUGS (foreground parallax) ===
	rugScroll := int32(scroll * 0.8)
	for base := int32(-400); base < screenWidth+400; base += 400 {
		rx := base - rugScroll%400
		r.drawLibraryRug(rx+200, 175)
	}

	// === CARPET RUNNER (ground level) ===
	rl.DrawRectangle(0, 168, screenWidth, 18, rl.Color{R: 80, G: 35, B: 45, A: 255})
	// Carpet pattern
	for base := int32(-30); base < screenWidth+30; base += 30 {
		px := base - floorScroll%30
		rl.DrawRectangle(px+5, 172, 10, 2, rl.Color{R: 120, G: 50, B: 60, A: 255})
		rl.DrawRectangle(px+18, 175, 8, 2, rl.Color{R: 100, G: 45, B: 55, A: 255})
	}

	// === FLOATING DUST MOTES ===
	r.drawLibraryDust(scroll, time)

	// === MAGICAL ORBS floating through ===
	r.drawFloatingOrbs(scroll, time)
}

// --- WIZARD'S LIBRARY ELEMENTS ---

func (r *Renderer) drawLibraryWindow(x, y int32, time float32) {
	// Window frame (ornate)
	rl.DrawRectangle(x-45, y, 90, 105, rl.Color{R: 65, G: 45, B: 40, A: 255})
	rl.DrawRectangle(x-42, y+3, 84, 99, rl.Color{R: 75, G: 52, B: 45, A: 255})

	// Night sky through window
	for wy := int32(y + 5); wy < y+100; wy++ {
		t := float32(wy-y-5) / 95.0
		c := rl.Color{
			R: uint8(15 + t*10),
			G: uint8(20 + t*15),
			B: uint8(45 + t*20),
			A: 255,
		}
		rl.DrawLine(x-40, wy, x+40, wy, c)
	}

	// Moon with glow
	moonX, moonY := x+15, y+25
	rl.DrawCircle(moonX, moonY, 18, rl.Color{R: 60, G: 60, B: 100, A: 30})
	rl.DrawCircle(moonX, moonY, 14, rl.Color{R: 80, G: 80, B: 120, A: 40})
	rl.DrawCircle(moonX, moonY, 8, rl.Color{R: 240, G: 235, B: 220, A: 255})
	rl.DrawCircle(moonX-2, moonY-1, 7, rl.Color{R: 250, G: 248, B: 235, A: 255})

	// Twinkling stars
	starPositions := [][2]int32{{-30, 20}, {-15, 40}, {5, 15}, {-22, 65}, {25, 45}, {10, 75}, {-32, 85}, {30, 30}}
	for i, pos := range starPositions {
		twinkle := uint8(180 + 75*simpleSinF(float64(time)*2.0+float64(i)*0.8))
		rl.DrawPixel(x+pos[0], y+pos[1], rl.Color{R: twinkle, G: twinkle, B: 255, A: 255})
	}

	// Window dividers
	rl.DrawRectangle(x-2, y+5, 4, 95, rl.Color{R: 60, G: 42, B: 38, A: 255})
	rl.DrawRectangle(x-40, y+45, 80, 4, rl.Color{R: 60, G: 42, B: 38, A: 255})

	// Curtains with gentle sway
	curtainSway := int32(2 * simpleSinF(float64(time)*0.8))
	// Left curtain
	for cy := int32(y - 3); cy < y+105; cy++ {
		wave := int32(simpleSinF(float64(cy)*0.1+float64(time)*0.5) * 2)
		rl.DrawLine(x-55+wave+curtainSway, cy, x-42+wave+curtainSway, cy, rl.Color{R: 100, G: 40, B: 50, A: 255})
	}
	// Right curtain
	for cy := int32(y - 3); cy < y+105; cy++ {
		wave := int32(simpleSinF(float64(cy)*0.1+float64(time)*0.5+1) * 2)
		rl.DrawLine(x+42-wave-curtainSway, cy, x+55-wave-curtainSway, cy, rl.Color{R: 100, G: 40, B: 50, A: 255})
	}
}

func (r *Renderer) drawChandelier(x, y int32, time float32) {
	// Chain
	rl.DrawRectangle(x-1, y-10, 2, 15, rl.Color{R: 150, G: 130, B: 90, A: 255})

	// Base
	rl.DrawRectangle(x-12, y+3, 24, 3, rl.Color{R: 170, G: 150, B: 100, A: 255})

	// Candle holders and flames
	for i := int32(-1); i <= 1; i++ {
		cx := x + i*8
		// Holder
		rl.DrawRectangle(cx-1, y+5, 3, 6, rl.Color{R: 160, G: 140, B: 90, A: 255})
		// Candle
		rl.DrawRectangle(cx-1, y-2, 2, 7, rl.Color{R: 235, G: 225, B: 200, A: 255})

		// Flame with flicker
		flicker := simpleSinF(float64(time)*8.0 + float64(i)*2.0)
		flickerX := int32(flicker * 1)
		rl.DrawCircle(cx+flickerX, y-4, 4, rl.Color{R: 255, G: 200, B: 100, A: 40})
		rl.DrawRectangle(cx-1+flickerX, y-6, 2, 4, rl.Color{R: 255, G: 180, B: 80, A: 255})
		rl.DrawPixel(cx+flickerX, y-7, rl.Color{R: 255, G: 255, B: 200, A: 255})
	}
}

func (r *Renderer) drawTallBookshelf(x, y int32, time float32) {
	height := int32(130)

	// Frame
	rl.DrawRectangle(x, y, 55, height, rl.Color{R: 70, G: 48, B: 40, A: 255})
	rl.DrawRectangle(x+3, y+3, 49, height-6, rl.Color{R: 60, G: 42, B: 35, A: 255})

	// Shelf dividers
	for sy := int32(y + 3); sy < y+height-10; sy += 32 {
		rl.DrawRectangle(x+3, sy, 49, 3, rl.Color{R: 75, G: 52, B: 42, A: 255})
	}

	// Books
	bookColors := []rl.Color{
		{R: 160, G: 60, B: 70, A: 255},
		{R: 70, G: 100, B: 160, A: 255},
		{R: 70, G: 140, B: 90, A: 255},
		{R: 180, G: 160, B: 80, A: 255},
		{R: 140, G: 80, B: 140, A: 255},
		{R: 80, G: 70, B: 60, A: 255},
		{R: 200, G: 100, B: 60, A: 255},
	}
	for shelf := int32(0); shelf < 4; shelf++ {
		shelfY := y + 6 + shelf*32
		bx := x + 5
		for book := int32(0); book < 7; book++ {
			bw := 5 + (book*shelf)%3
			bh := 22 + (book*3+shelf*2)%8
			// Ensure positive index for book colors (x can be negative during scroll)
			colorIdx := (int(book) + int(shelf)*2 + int(x/50)) % len(bookColors)
			if colorIdx < 0 {
				colorIdx += len(bookColors)
			}
			bc := bookColors[colorIdx]
			rl.DrawRectangle(bx, shelfY+28-bh, bw, bh, bc)
			rl.DrawLine(bx+bw/2, shelfY+30-bh, bx+bw/2, shelfY+26, rl.Color{R: bc.R - 30, G: bc.G - 30, B: bc.B - 30, A: 255})
			bx += bw + 1
			if bx > x+48 {
				break
			}
		}
	}

	// Occasional glowing orb on shelf (use abs to handle negative x)
	xAbs := x
	if xAbs < 0 {
		xAbs = -xAbs
	}
	if int(xAbs/80)%3 == 0 {
		orbY := y + 40
		orbGlow := uint8(150 + 50*simpleSinF(float64(time)*1.5+float64(x)*0.1))
		rl.DrawCircle(x+35, orbY, 5, rl.Color{R: 100, G: orbGlow, B: 200, A: 80})
		rl.DrawCircle(x+35, orbY, 3, rl.Color{R: 150, G: orbGlow, B: 230, A: 150})
		rl.DrawCircle(x+35, orbY, 2, rl.Color{R: 200, G: 220, B: 255, A: 255})
	}

	// Occasional skull
	if int(xAbs/120)%2 == 1 {
		rl.DrawCircle(x+45, y+height-15, 5, rl.Color{R: 230, G: 225, B: 215, A: 255})
		rl.DrawPixel(x+43, y+height-16, rl.Color{R: 30, G: 25, B: 35, A: 255})
		rl.DrawPixel(x+47, y+height-16, rl.Color{R: 30, G: 25, B: 35, A: 255})
	}
}

func (r *Renderer) drawWizardDesk(x, y int32, time float32) {
	// Desk body
	rl.DrawRectangle(x-40, y, 80, 65, rl.Color{R: 75, G: 52, B: 42, A: 255})
	// Desk top
	rl.DrawRectangle(x-43, y-5, 86, 8, rl.Color{R: 85, G: 58, B: 48, A: 255})
	// Desk legs
	rl.DrawRectangle(x-37, y+60, 6, 5, rl.Color{R: 65, G: 45, B: 38, A: 255})
	rl.DrawRectangle(x+31, y+60, 6, 5, rl.Color{R: 65, G: 45, B: 38, A: 255})

	// Drawer
	rl.DrawRectangle(x-20, y+15, 40, 25, rl.Color{R: 65, G: 45, B: 38, A: 255})
	rl.DrawCircle(x, y+27, 2, rl.Color{R: 180, G: 160, B: 100, A: 255})

	// Open spellbook
	rl.DrawRectangle(x-20, y-13, 30, 8, rl.Color{R: 90, G: 60, B: 50, A: 255})
	rl.DrawRectangle(x-18, y-15, 12, 8, rl.Color{R: 230, G: 220, B: 190, A: 255})
	rl.DrawRectangle(x-4, y-15, 12, 8, rl.Color{R: 225, G: 215, B: 185, A: 255})
	// Text lines
	rl.DrawLine(x-16, y-13, x-8, y-13, rl.Color{R: 60, G: 50, B: 40, A: 200})
	rl.DrawLine(x-16, y-11, x-9, y-11, rl.Color{R: 60, G: 50, B: 40, A: 200})

	// Quill in inkwell
	rl.DrawRectangle(x+15, y-13, 6, 8, rl.Color{R: 40, G: 35, B: 50, A: 255})
	rl.DrawLine(x+18, y-13, x+25, y-25, rl.Color{R: 220, G: 200, B: 180, A: 255})
	rl.DrawLine(x+25, y-25, x+30, y-30, rl.Color{R: 180, G: 100, B: 80, A: 255})

	// Crystal ball
	crystalX, crystalY := x-30, y-13
	rl.DrawCircle(crystalX, crystalY, 6, rl.Color{R: 80, G: 100, B: 140, A: 200})
	rl.DrawCircle(crystalX-1, crystalY-1, 4, rl.Color{R: 100, G: 120, B: 160, A: 180})
	mistAngle := time * 2.0
	mx := int32(simpleSinF(float64(mistAngle)) * 2)
	my := int32(simpleSinF(float64(mistAngle)+1.5) * 2)
	rl.DrawPixel(crystalX+mx, crystalY+my, rl.Color{R: 180, G: 200, B: 255, A: 200})
	rl.DrawPixel(crystalX-2, crystalY-2, rl.Color{R: 255, G: 255, B: 255, A: 200})

	// Potions
	potionColors := []rl.Color{
		{R: 200, G: 80, B: 100, A: 255},
		{R: 80, G: 150, B: 200, A: 255},
	}
	for i, pc := range potionColors {
		px := x + 25 + int32(i)*12
		rl.DrawRectangle(px, y-20, 8, 12, rl.Color{R: 200, G: 200, B: 220, A: 100})
		bubbleOff := int32(simpleSinF(float64(time)*3.0+float64(i)*1.5) * 2)
		rl.DrawRectangle(px+1, y-16+bubbleOff, 6, 7-bubbleOff, pc)
		rl.DrawRectangle(px+2, y-22, 4, 3, rl.Color{R: 140, G: 100, B: 70, A: 255})
	}
}

func (r *Renderer) drawLibraryRug(x, y int32) {
	rugColors := []rl.Color{
		{R: 120, G: 45, B: 60, A: 255},
		{R: 140, G: 55, B: 70, A: 255},
		{R: 90, G: 35, B: 50, A: 255},
	}
	for i := 3; i >= 0; i-- {
		radius := float32(45 - i*10)
		rl.DrawEllipse(x, y, radius, radius*0.3, rugColors[i%3])
	}
	// Fringe
	for fx := int32(x - 42); fx < x+42; fx += 4 {
		rl.DrawLine(fx, y+7, fx+1, y+11, rl.Color{R: 100, G: 40, B: 55, A: 255})
	}
}

func (r *Renderer) drawLibraryDust(scroll float32, time float32) {
	for i := 0; i < 12; i++ {
		baseX := float64((i*61 + 23) % screenWidth)
		baseY := float64(40 + (i*41)%100)
		fx := baseX + 15*simpleSinF(float64(time)*0.3+float64(i)*0.7) - float64(int(scroll*0.2)%screenWidth)
		fy := baseY + 10*simpleSinF(float64(time)*0.5+float64(i)*1.1)

		// Wrap around
		for fx < 0 {
			fx += float64(screenWidth)
		}
		for fx > float64(screenWidth) {
			fx -= float64(screenWidth)
		}

		alpha := uint8(80 + 40*simpleSinF(float64(time)*0.8+float64(i)))
		rl.DrawPixel(int32(fx), int32(fy), rl.Color{R: 255, G: 240, B: 200, A: alpha})
	}
}

func (r *Renderer) drawFloatingOrbs(scroll float32, time float32) {
	for i := 0; i < 3; i++ {
		baseX := float64((i*97 + 50) % screenWidth)
		baseY := float64(60 + (i*37)%60)

		fx := baseX + 20*simpleSinF(float64(time)*0.4+float64(i)*1.2) - float64(int(scroll*0.15)%screenWidth)
		fy := baseY + 15*simpleSinF(float64(time)*0.6+float64(i)*0.9)

		// Wrap
		for fx < -20 {
			fx += float64(screenWidth + 40)
		}
		for fx > float64(screenWidth+20) {
			fx -= float64(screenWidth + 40)
		}

		// Pulsing glow
		pulse := uint8(100 + 80*simpleSinF(float64(time)*2+float64(i)*1.5))
		hue := (int(time*30) + i*60) % 360
		orbColor := hsvToRGB(hue, 0.6, 1.0)
		orbColor.A = pulse

		rl.DrawCircle(int32(fx), int32(fy), 4, rl.Color{R: orbColor.R, G: orbColor.G, B: orbColor.B, A: pulse / 3})
		rl.DrawCircle(int32(fx), int32(fy), 2, orbColor)
	}
}
