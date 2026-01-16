package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// simpleSinF returns the sine of x (in radians)
func simpleSinF(x float64) float64 {
	return math.Sin(x)
}

// simpleCosF returns the cosine of x (in radians)
func simpleCosF(x float64) float64 {
	return math.Cos(x)
}

// drawShippedBanner draws the epic rainbow "SHIPPED!" banner flying across the screen
func (r *Renderer) drawShippedBanner(state *GameState) {
	// Rainbow colors (brighter versions of indigo/violet)
	rainbow := []rl.Color{
		{255, 80, 80, 255},   // Red (slightly softened)
		{255, 160, 50, 255},  // Orange
		{255, 255, 80, 255},  // Yellow
		{80, 255, 120, 255},  // Green
		{80, 180, 255, 255},  // Blue
		{140, 100, 255, 255}, // Indigo (brighter)
		{200, 120, 255, 255}, // Violet (brighter)
	}

	text := "SHIPPED!"

	// Helper to calculate position along the arc at a given time
	getArcPos := func(t float32) (int32, int32) {
		progress := t / 3.0
		px := int32(-120 + progress*450)
		arcHeight := float32(60)
		arcProgress := (progress - 0.5) * 2
		py := int32(100 - arcHeight*(1-arcProgress*arcProgress))
		return px, py
	}

	// Draw smear trail - many smaller SHIPPED! following behind
	// Longer tail with 32 trails
	numTrails := 32
	for trail := numTrails - 1; trail >= 0; trail-- {
		// Each trail is slightly behind in time
		trailTime := state.ShippedTimer - float32(trail)*0.03
		if trailTime < 0 {
			continue
		}

		trailX, trailY := getArcPos(trailTime)

		// Size decreases for trails further back
		scale := 1.0 - float32(trail)*0.025
		if scale < 0.25 {
			scale = 0.25
		}
		fontSize := int32(24 * scale)

		// Alpha decreases for trails further back
		alpha := uint8(255 - trail*7)
		if alpha < 40 {
			alpha = 40
		}

		// Rainbow alternates - colors cycle through based on time
		colorIdx := (trail + int(state.ShippedTimer*8)) % len(rainbow)
		c := rainbow[colorIdx]
		c.A = alpha

		// Draw the trail text (single color per trail)
		rl.DrawText(text, trailX, trailY, fontSize, c)
	}

	// Get main text position
	x, y := getArcPos(state.ShippedTimer)
	fontSize := int32(24)

	// Main text - white with black outline (pops against rainbow)
	black := rl.Color{0, 0, 0, 255}
	white := rl.Color{255, 255, 255, 255}

	// Black outline
	for dx := int32(-2); dx <= 2; dx++ {
		for dy := int32(-2); dy <= 2; dy++ {
			if dx != 0 || dy != 0 {
				rl.DrawText(text, x+dx, y+dy, fontSize, black)
			}
		}
	}
	// White fill
	rl.DrawText(text, x, y, fontSize, white)
}

// drawThinkHardEffect renders firework-style particle effects
func (r *Renderer) drawThinkHardEffect(state *GameState) {
	// Get the burst text based on think level
	var burstText string
	var baseColor rl.Color
	var isUltra bool
	var particleIntensity float32

	switch state.ThinkLevel {
	case ThinkNormal:
		burstText = "THINK!"
		baseColor = rl.Color{R: 150, G: 180, B: 255, A: 255}
		particleIntensity = 0.15
	case ThinkHard:
		burstText = "THINK HARD!"
		baseColor = rl.Color{R: 255, G: 200, B: 80, A: 255}
		particleIntensity = 0.25
	case ThinkHarder:
		burstText = "THINK HARDER!"
		baseColor = rl.Color{R: 255, G: 150, B: 50, A: 255}
		particleIntensity = 0.35
	case ThinkUltra:
		burstText = "ULTRATHINK!"
		isUltra = true
		particleIntensity = 0.5
	default:
		burstText = "THINK!"
		baseColor = rl.Color{R: 200, G: 220, B: 255, A: 255}
		particleIntensity = 0.15
	}

	// Position varies based on timer - cycles through different spots
	positions := []struct{ x, y int32 }{
		{screenWidth/2 + 40, 70},  // Right of head
		{screenWidth/2 - 50, 65},  // Left of head
		{screenWidth/2 + 50, 55},  // Upper right
		{screenWidth/2 - 40, 80},  // Lower left
		{screenWidth/2, 50},       // Above head
	}
	posIdx := int(state.ThinkHardTimer*1.5) % len(positions)
	cx := positions[posIdx].x
	cy := positions[posIdx].y

	// Spawn firework particles
	r.spawnThinkParticles(float32(cx), float32(cy), particleIntensity, isUltra, state.ThinkHardTimer)

	// Measure text
	fontSize := int32(8)
	if isUltra {
		fontSize = 10
	}
	textWidth := rl.MeasureText(burstText, fontSize)

	// Subtle pulsing glow behind text
	pulse := float32(1.0) + float32(0.1*simpleSinF(float64(state.ThinkHardTimer*6)))
	glowW := int32(float32(textWidth+8) * pulse)
	glowH := int32(float32(fontSize+6) * pulse)
	glowX := cx - glowW/2
	glowY := cy - glowH/2

	// Draw glow (semi-transparent background)
	glowColor := baseColor
	if isUltra {
		hue := int(state.ThinkHardTimer*200) % 360
		glowColor = hsvToRGB(hue, 0.6, 1.0)
	}
	glowColor.A = 150
	rl.DrawRectangle(glowX, glowY, glowW, glowH, glowColor)

	// Draw text
	textX := cx - textWidth/2
	textY := cy - fontSize/2

	// Shadow for all
	shadowColor := rl.Color{R: 20, G: 15, B: 30, A: 180}
	rl.DrawText(burstText, textX+1, textY+1, fontSize, shadowColor)

	if isUltra {
		// Cycling bright color for ULTRATHINK
		hue := int(state.ThinkHardTimer*300) % 360
		textColor := hsvToRGB(hue, 1.0, 1.0)
		rl.DrawText(burstText, textX, textY, fontSize, textColor)
	} else {
		rl.DrawText(burstText, textX, textY, fontSize, baseColor)
	}
}

// spawnThinkParticles creates firework-style particles for thinking effect
func (r *Renderer) spawnThinkParticles(cx, cy, intensity float32, isUltra bool, timer float32) {
	// Spawn rate based on intensity
	if rand.Float32() > intensity {
		return
	}

	// Create a burst of particles
	numParticles := 2
	if isUltra {
		numParticles = 4
	}

	for i := 0; i < numParticles; i++ {
		// Random angle for burst direction
		angle := rand.Float32() * 6.28318 // 2*PI

		// Velocity based on angle
		speed := float64(20 + rand.Float32()*40)
		vx := speed * simpleCosF(float64(angle))
		vy := speed * simpleSinF(float64(angle))

		// Color
		var color rl.Color
		if isUltra {
			// Rainbow colors
			hue := (int(timer*200) + rand.Intn(120)) % 360
			color = hsvToRGB(hue, 1.0, 1.0)
		} else {
			// Warm colors - yellows, oranges, whites
			colors := []rl.Color{
				{R: 255, G: 255, B: 200, A: 255}, // White-yellow
				{R: 255, G: 220, B: 100, A: 255}, // Yellow
				{R: 255, G: 180, B: 80, A: 255},  // Orange
				{R: 255, G: 200, B: 150, A: 255}, // Light orange
			}
			color = colors[rand.Intn(len(colors))]
		}

		// Spawn particle near center with some spread
		px := cx + (rand.Float32()-0.5)*10
		py := cy + (rand.Float32()-0.5)*10

		r.particles = append(r.particles, Particle{
			X:       px,
			Y:       py,
			VX:      float32(vx),
			VY:      float32(vy),
			Life:    0.4 + rand.Float32()*0.4,
			MaxLife: 0.8,
			Color:   color,
			Size:    1 + rand.Float32()*2,
		})
	}

	// Extra trailing sparkles for ultra
	if isUltra && rand.Float32() < 0.3 {
		// Spawn a larger, slower sparkle
		hue := int(timer*300) % 360
		r.particles = append(r.particles, Particle{
			X:       cx + (rand.Float32()-0.5)*30,
			Y:       cy + (rand.Float32()-0.5)*20,
			VX:      (rand.Float32() - 0.5) * 10,
			VY:      -5 - rand.Float32()*10,
			Life:    0.8,
			MaxLife: 0.8,
			Color:   hsvToRGB(hue, 1.0, 1.0),
			Size:    3,
		})
	}
}

// drawCompactEffect renders the rest/sleep effect after compact
func (r *Renderer) drawCompactEffect(state *GameState) {
	// Draw "Zzz" floating up
	progress := state.CompactTimer / 2.0 // 0 to 1 over 2 seconds

	// Zzz text position - floats up and fades
	zx := int32(screenWidth/2 + 25)
	zy := int32(70 - int32(progress*20))

	alpha := uint8((1.0 - progress) * 255)
	zColor := rl.Color{R: 180, G: 180, B: 220, A: alpha}

	// Draw multiple Z's at different sizes
	rl.DrawText("z", zx, zy, 8, zColor)
	rl.DrawText("z", zx+8, zy-6, 10, zColor)
	rl.DrawText("Z", zx+18, zy-14, 12, zColor)

	// Draw "REST" text
	if progress < 0.5 {
		restAlpha := uint8((0.5 - progress) * 2 * 200)
		restColor := rl.Color{R: 100, G: 180, B: 100, A: restAlpha}
		restText := "MANA RESTORED"
		restWidth := rl.MeasureText(restText, 8)
		rl.DrawText(restText, (screenWidth-restWidth)/2, 45, 8, restColor)
	}
}

// hsvToRGB converts HSV to RGB color
func hsvToRGB(h int, s, v float64) rl.Color {
	h = h % 360
	c := v * s
	x := c * (1 - absF(float64(h%120)/60.0-1))
	m := v - c

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return rl.Color{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

// absF returns the absolute value of a float64
func absF(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// spawnChestParticles creates sparkle particles for treasure chest opening
func (r *Renderer) spawnChestParticles(cx, cy float32, intense bool) {
	// Spawn rate
	spawnChance := float32(0.3)
	if intense {
		spawnChance = 0.6
	}

	if rand.Float32() > spawnChance {
		return
	}

	// Number of particles per spawn
	numParticles := 1
	if intense {
		numParticles = 2
	}

	for i := 0; i < numParticles; i++ {
		// Random angle for burst direction (upward bias)
		angle := rand.Float32()*3.14159 + 3.14159/2 // PI/2 to 3PI/2 (upward half)

		// Velocity
		speed := float64(15 + rand.Float32()*25)
		vx := speed * simpleCosF(float64(angle))
		vy := speed * simpleSinF(float64(angle))

		// Gold/yellow sparkle colors
		colors := []rl.Color{
			{R: 255, G: 215, B: 0, A: 255},   // Gold
			{R: 255, G: 255, B: 150, A: 255}, // Light yellow
			{R: 255, G: 200, B: 100, A: 255}, // Orange-gold
			{R: 255, G: 255, B: 255, A: 255}, // White sparkle
		}
		color := colors[rand.Intn(len(colors))]

		// Spawn from chest area
		px := cx + (rand.Float32()-0.5)*40
		py := cy + (rand.Float32()-0.5)*20

		r.particles = append(r.particles, Particle{
			X:       px,
			Y:       py,
			VX:      float32(vx),
			VY:      float32(vy),
			Life:    0.5 + rand.Float32()*0.5,
			MaxLife: 1.0,
			Color:   color,
			Size:    1 + rand.Float32()*2,
		})
	}
}
