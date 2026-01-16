package main

import rl "github.com/gen2brain/raylib-go/raylib"

// ============================================================================
// BIOME 3: KINGDOM ROAD - Castle in distance, villages, farmland
// ============================================================================
func (r *Renderer) drawBiomeKingdomRoad() {
	scroll := r.scrollOffset
	time := r.biomeTimer

	// Warm sunset sky
	for y := int32(0); y < 100; y++ {
		t := float32(y) / 100.0
		c := rl.Color{
			R: uint8(80 + t*80),
			G: uint8(50 + t*60),
			B: uint8(60 + t*40),
			A: 255,
		}
		rl.DrawLine(0, y, screenWidth, y, c)
	}

	// Setting sun
	sunX := int32(280) - int32(scroll*0.01)%400
	rl.DrawCircle(sunX, 85, 15, rl.Color{R: 255, G: 200, B: 100, A: 255})
	rl.DrawCircle(sunX, 85, 20, rl.Color{R: 255, G: 180, B: 80, A: 60})

	// Clouds
	cloudScroll := int32(scroll * 0.05)
	for base := int32(-200); base < screenWidth+200; base += 200 {
		cx := base - cloudScroll%200
		r.drawCloud(cx+50, 25)
		r.drawCloud(cx+150, 35)
	}

	// Distant castle!
	castleX := int32(160) - int32(scroll*0.08)%600
	if castleX > -100 && castleX < screenWidth+100 {
		r.drawCastle(castleX, 60)
	}

	// Rolling hills with farms
	hillScroll := int32(scroll * 0.2)
	for base := int32(-200); base < screenWidth+200; base += 200 {
		x := base - hillScroll%200
		r.drawFarmHill(x+50, 110, 80, 40, rl.Color{R: 70, G: 90, B: 50, A: 255})
		r.drawFarmHill(x+150, 105, 100, 50, rl.Color{R: 65, G: 85, B: 45, A: 255})
		// Windmill on some hills
		if base%400 == 0 {
			r.drawWindmill(x+80, 75, time)
		}
	}

	// Mid hills with houses
	midScroll := int32(scroll * 0.45)
	for base := int32(-160); base < screenWidth+160; base += 160 {
		x := base - midScroll%160
		r.drawFarmHill(x+30, 135, 60, 30, rl.Color{R: 55, G: 75, B: 40, A: 255})
		r.drawFarmHill(x+100, 130, 70, 35, rl.Color{R: 60, G: 80, B: 45, A: 255})
		// Cottages
		r.drawCottage(x+50, 118)
		r.drawCottage(x+120, 112)
	}

	// Foreground fences and crops
	fgScroll := int32(scroll * 0.75)
	for base := int32(-100); base < screenWidth+100; base += 100 {
		x := base - fgScroll%100
		r.drawFence(x, 158)
		r.drawWheatField(x+30, 155, time)
	}

	// Ground - fertile farmland
	groundScroll := int32(scroll * 1.0)
	rl.DrawRectangle(0, 160, screenWidth, 40, rl.Color{R: 65, G: 80, B: 45, A: 255})

	// Flowers and grass
	for base := int32(-50); base < screenWidth+50; base += 50 {
		gx := base - groundScroll%50
		r.drawFlower(gx+10, 163, rl.Color{R: 255, G: 200, B: 100, A: 255})
		r.drawFlower(gx+30, 165, rl.Color{R: 255, G: 150, B: 150, A: 255})
		r.drawGrass(gx+20, 162)
		r.drawGrass(gx+45, 164)
	}

	// Cobblestone road
	rl.DrawRectangle(0, 168, screenWidth, 18, rl.Color{R: 90, G: 80, B: 70, A: 255})
	for base := int32(-30); base < screenWidth+30; base += 30 {
		px := base - groundScroll%30
		rl.DrawRectangle(px+3, 170, 8, 6, rl.Color{R: 100, G: 90, B: 80, A: 255})
		rl.DrawRectangle(px+14, 172, 10, 7, rl.Color{R: 80, G: 70, B: 60, A: 255})
		rl.DrawRectangle(px+8, 178, 7, 5, rl.Color{R: 95, G: 85, B: 75, A: 255})
	}
}

// --- KINGDOM ROAD ELEMENTS ---

func (r *Renderer) drawCloud(x, y int32) {
	cloudColor := rl.Color{R: 255, G: 240, B: 230, A: 180}
	rl.DrawCircle(x, y, 8, cloudColor)
	rl.DrawCircle(x+10, y+2, 6, cloudColor)
	rl.DrawCircle(x-8, y+2, 5, cloudColor)
	rl.DrawCircle(x+5, y-3, 5, cloudColor)
}

func (r *Renderer) drawCastle(x, y int32) {
	stoneColor := rl.Color{R: 90, G: 85, B: 100, A: 255}
	roofColor := rl.Color{R: 70, G: 50, B: 60, A: 255}
	// Main keep
	rl.DrawRectangle(x-15, y, 30, 40, stoneColor)
	// Towers
	rl.DrawRectangle(x-25, y-10, 12, 50, stoneColor)
	rl.DrawRectangle(x+13, y-10, 12, 50, stoneColor)
	// Tower roofs (pointed)
	for i := int32(0); i < 10; i++ {
		w := 12 - i
		rl.DrawRectangle(x-25+(12-w)/2, y-10-i, w, 1, roofColor)
		rl.DrawRectangle(x+13+(12-w)/2, y-10-i, w, 1, roofColor)
	}
	// Windows
	rl.DrawRectangle(x-5, y+10, 3, 5, rl.Color{R: 255, G: 220, B: 150, A: 255})
	rl.DrawRectangle(x+2, y+10, 3, 5, rl.Color{R: 255, G: 220, B: 150, A: 255})
	// Flag
	rl.DrawRectangle(x, y-15, 1, 10, rl.Color{R: 60, G: 50, B: 45, A: 255})
	rl.DrawRectangle(x+1, y-15, 6, 4, rl.Color{R: 200, G: 50, B: 50, A: 255})
}

func (r *Renderer) drawFarmHill(x, baseY, width, height int32, color rl.Color) {
	for row := int32(0); row < height; row++ {
		t := float32(row) / float32(height)
		w := int32(float32(width) * (1 - t*t))
		rl.DrawRectangle(x-w/2, baseY-row, w, 1, color)
	}
}

func (r *Renderer) drawWindmill(x, y int32, time float32) {
	// Tower
	rl.DrawRectangle(x-4, y, 8, 20, rl.Color{R: 180, G: 170, B: 150, A: 255})
	// Roof
	for i := int32(0); i < 6; i++ {
		rl.DrawRectangle(x-5+i/2, y-i, 10-i, 1, rl.Color{R: 120, G: 80, B: 60, A: 255})
	}
	// Rotating blades
	angle := time * 2
	for i := 0; i < 4; i++ {
		a := angle + float32(i)*1.57
		bx := x + int32(10*simpleCosF(float64(a)))
		by := y - 3 + int32(10*simpleSinF(float64(a)))
		rl.DrawLine(x, y-3, bx, by, rl.Color{R: 100, G: 80, B: 60, A: 255})
	}
}

func (r *Renderer) drawCottage(x, y int32) {
	// Walls
	rl.DrawRectangle(x-6, y, 12, 10, rl.Color{R: 180, G: 160, B: 140, A: 255})
	// Roof
	for i := int32(0); i < 6; i++ {
		rl.DrawRectangle(x-8+i, y-i, 16-i*2, 1, rl.Color{R: 140, G: 90, B: 70, A: 255})
	}
	// Door
	rl.DrawRectangle(x-1, y+4, 3, 6, rl.Color{R: 80, G: 60, B: 50, A: 255})
	// Window
	rl.DrawRectangle(x+3, y+3, 2, 2, rl.Color{R: 255, G: 230, B: 150, A: 255})
}

func (r *Renderer) drawFence(x, y int32) {
	fenceColor := rl.Color{R: 120, G: 100, B: 80, A: 255}
	// Posts
	for i := int32(0); i < 5; i++ {
		rl.DrawRectangle(x+i*8, y-6, 2, 8, fenceColor)
	}
	// Rails
	rl.DrawRectangle(x, y-5, 34, 1, fenceColor)
	rl.DrawRectangle(x, y-2, 34, 1, fenceColor)
}

func (r *Renderer) drawWheatField(x, y int32, time float32) {
	wheatColor := rl.Color{R: 220, G: 180, B: 80, A: 255}
	for i := int32(0); i < 8; i++ {
		sway := int32(simpleSinF(float64(time*2+float32(i+x)*0.3)) * 1)
		rl.DrawRectangle(x+i*3+sway, y-4, 1, 5, wheatColor)
		rl.DrawPixel(x+i*3+sway, y-5, rl.Color{R: 240, G: 200, B: 100, A: 255})
	}
}

func (r *Renderer) drawFlower(x, y int32, color rl.Color) {
	// Stem
	rl.DrawPixel(x, y, rl.Color{R: 50, G: 100, B: 50, A: 255})
	rl.DrawPixel(x, y-1, rl.Color{R: 50, G: 100, B: 50, A: 255})
	// Petals
	rl.DrawPixel(x, y-2, color)
	rl.DrawPixel(x-1, y-2, color)
	rl.DrawPixel(x+1, y-2, color)
	rl.DrawPixel(x, y-3, color)
}
