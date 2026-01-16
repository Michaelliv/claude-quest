package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Claude position constants (same as in renderer_claude.go)
const (
	claudeX = screenWidth/2 - spriteFrameWidth
	claudeY = screenHeight - spriteFrameHeight*claudeScale - 24
)

// ============================================================================
// AURA SYSTEM - Particle effects around Claude
// ============================================================================

// drawAura renders the currently selected aura around Claude
func (r *Renderer) drawAura(state *AnimationState) {
	if r.currentAura < 0 || r.currentAura >= len(r.auraNames) {
		return
	}

	auraName := r.auraNames[r.currentAura]
	time := float32(rl.GetTime())
	cx := float32(claudeX + spriteFrameWidth)  // Center X
	cy := float32(claudeY + spriteFrameHeight) // Center Y

	switch auraName {
	case "aura_pixel":
		r.drawAuraPixelDust(cx, cy, time)
	case "aura_flame":
		r.drawAuraFlame(cx, cy, time)
	case "aura_frost":
		r.drawAuraFrost(cx, cy, time)
	case "aura_electric":
		r.drawAuraElectric(cx, cy, time)
	case "aura_shadow":
		r.drawAuraShadow(cx, cy, time)
	case "aura_heart":
		r.drawAuraHeart(cx, cy, time)
	case "aura_code":
		r.drawAuraCode(cx, cy, time)
	case "aura_rainbow":
		r.drawAuraRainbow(cx, cy, time)
	}
}

// drawAuraPixelDust - Bright white/yellow sparkles around Claude
func (r *Renderer) drawAuraPixelDust(cx, cy, time float32) {
	// Outer glow ring
	for i := 0; i < 24; i++ {
		angle := float64(time*2.5) + float64(i)*0.26
		radius := 28.0 + 10.0*math.Sin(float64(time*3.0)+float64(i)*0.5)
		x := cx + float32(math.Cos(angle)*radius)
		y := cy + float32(math.Sin(angle)*radius) - 20
		alpha := uint8(200 + 55*math.Sin(float64(time*6.0)+float64(i)))
		// Bright sparkle with glow
		rl.DrawCircle(int32(x), int32(y), 2, rl.Color{R: 255, G: 255, B: 200, A: alpha / 2})
		rl.DrawCircle(int32(x), int32(y), 1, rl.Color{R: 255, G: 255, B: 255, A: alpha})
	}
	// Inner sparkle bursts
	for i := 0; i < 8; i++ {
		angle := float64(time*4.0) + float64(i)*0.785
		radius := 15.0 + 5.0*math.Sin(float64(time*5.0)+float64(i))
		x := cx + float32(math.Cos(angle)*radius)
		y := cy + float32(math.Sin(angle)*radius) - 20
		rl.DrawCircle(int32(x), int32(y), 2, rl.Color{R: 255, G: 255, B: 100, A: 255})
	}
}

// drawAuraFlame - Blazing fire engulfing Claude
func (r *Renderer) drawAuraFlame(cx, cy, time float32) {
	baseY := cy + 10

	// Layer 1: Outer glow - warm ambient light
	glowPulse := float32(0.7 + 0.3*math.Sin(float64(time*4.0)))
	rl.DrawCircle(int32(cx), int32(cy-10), 35, rl.Color{R: 255, G: 100, B: 20, A: uint8(30 * glowPulse)})
	rl.DrawCircle(int32(cx), int32(cy-10), 25, rl.Color{R: 255, G: 150, B: 50, A: uint8(50 * glowPulse)})

	// Layer 2: Back flames (darker, behind Claude)
	for i := 0; i < 7; i++ {
		seed := float64(i) * 1.3
		tongueX := cx + float32(i-3)*9
		baseHeight := 35.0 + 10.0*math.Sin(float64(time*2.5)+seed)
		sway := float32(math.Sin(float64(time*3.5)+seed)) * 4

		for j := 0; j < int(baseHeight/3); j++ {
			progress := float32(j) / float32(baseHeight/3)
			// Organic narrowing - flames taper with slight bulges
			widthBase := 4.0 - progress*3.0
			widthWobble := float32(math.Sin(float64(time*6.0)+seed+float64(j)*0.3)) * 0.5
			width := float32(widthBase) + widthWobble
			if width < 0.5 {
				width = 0.5
			}

			// Cumulative sway - more at top
			xOff := sway * progress * progress
			flicker := float32(math.Sin(float64(time*10.0)+seed+float64(j)*0.4)) * (1 + progress*2)

			y := baseY - float32(j)*3
			x := tongueX + xOff + flicker

			// Darker back flames
			red := uint8(200 - progress*80)
			green := uint8(80 - progress*60)
			alpha := uint8((1.0 - progress*0.8) * 150)
			rl.DrawCircle(int32(x), int32(y), width, rl.Color{R: red, G: green, B: 10, A: alpha})
		}
	}

	// Layer 3: Main flame tongues (bright, in front)
	for i := 0; i < 9; i++ {
		seed := float64(i) * 0.9
		tongueX := cx + float32(i-4)*7
		// Varying heights with time - flames dance
		baseHeight := 40.0 + 15.0*math.Sin(float64(time*3.0)+seed) + 8.0*math.Sin(float64(time*5.0)+seed*2)
		sway := float32(math.Sin(float64(time*4.0)+seed)) * 5

		for j := 0; j < int(baseHeight/2.5); j++ {
			progress := float32(j) / float32(baseHeight/2.5)

			// Natural flame shape - wide at base, narrow at top with dancing
			widthBase := 5.0 - progress*4.0
			widthPulse := float32(math.Sin(float64(time*8.0)+seed+float64(j)*0.2)) * (0.5 + progress)
			width := float32(widthBase) + widthPulse
			if width < 0.3 {
				width = 0.3
			}

			// Sway increases exponentially toward tip
			xOff := sway * progress * progress * 1.5
			// High-frequency flicker
			flicker := float32(math.Sin(float64(time*12.0)+seed+float64(j)*0.5)) * (1 + progress*3)

			y := baseY - float32(j)*2.5
			x := tongueX + xOff + flicker

			// Color gradient: white core -> yellow -> orange -> red -> dark red tip
			var red, green, blue uint8
			if progress < 0.15 {
				// Hot white/blue core
				red, green, blue = 255, 255, 240
			} else if progress < 0.35 {
				// Bright yellow
				red, green, blue = 255, 240, 100
			} else if progress < 0.55 {
				// Orange
				p := (progress - 0.35) / 0.2
				red = 255
				green = uint8(240 - p*140)
				blue = uint8(100 - p*80)
			} else if progress < 0.8 {
				// Red
				p := (progress - 0.55) / 0.25
				red = uint8(255 - p*55)
				green = uint8(100 - p*70)
				blue = 20
			} else {
				// Dark red tip
				red, green, blue = 180, 30, 10
			}

			alpha := uint8((1.0 - progress*0.7) * 255)
			rl.DrawCircle(int32(x), int32(y), width, rl.Color{R: red, G: green, B: blue, A: alpha})

			// Inner bright core for lower parts
			if progress < 0.3 && width > 1.5 {
				rl.DrawCircle(int32(x), int32(y), width*0.5, rl.Color{R: 255, G: 255, B: 255, A: alpha / 2})
			}
		}
	}

	// Layer 4: Sparks and embers spiraling up
	for i := 0; i < 12; i++ {
		seed := float64(i) * 0.7
		phase := math.Mod(float64(time*1.5)+seed, 2.5)
		progress := float32(phase / 2.5)

		// Spiral path upward
		angle := phase * 3.0
		radius := 5.0 + float64(progress)*20
		sparkX := cx + float32(math.Sin(angle)*radius)
		sparkY := baseY - progress*60

		// Sparks get smaller and dimmer as they rise
		size := float32(2.0 - progress*1.5)
		if size < 0.5 {
			size = 0.5
		}
		alpha := uint8((1.0 - progress) * 255)

		// Color fades from yellow to orange to red
		var r, g uint8
		if progress < 0.3 {
			r, g = 255, 220
		} else if progress < 0.6 {
			r, g = 255, uint8(220-progress*200)
		} else {
			r, g = uint8(255-progress*100), 50
		}

		rl.DrawCircle(int32(sparkX), int32(sparkY), size, rl.Color{R: r, G: g, B: 30, A: alpha})
	}

	// Layer 5: Heat shimmer particles (subtle)
	for i := 0; i < 6; i++ {
		seed := float64(i) * 1.1
		phase := math.Mod(float64(time*0.8)+seed, 2.0)
		x := cx + float32(math.Sin(float64(time*2.0)+seed)*20)
		y := baseY - 50 - float32(phase*25)
		alpha := uint8(40 + 30*math.Sin(float64(time*5.0)+seed))
		rl.DrawCircle(int32(x), int32(y), 2, rl.Color{R: 255, G: 200, B: 150, A: alpha})
	}
}

// drawAuraFrost - Swirling ice crystals and snowflakes
func (r *Renderer) drawAuraFrost(cx, cy, time float32) {
	// Falling snowflakes
	for i := 0; i < 20; i++ {
		baseX := cx + float32((i*13)%50-25)
		phase := float64(time*2.0) + float64(i)*0.5
		yOffset := float32(math.Mod(phase, 2.0) * 40)
		x := baseX + float32(math.Sin(phase*2.0)*8)
		y := cy - 50 + yOffset
		alpha := uint8(230 - yOffset*3)
		if alpha < 50 {
			alpha = 50
		}
		// Snowflake with cross pattern
		rl.DrawCircle(int32(x), int32(y), 2, rl.Color{R: 180, G: 220, B: 255, A: alpha})
		rl.DrawPixel(int32(x-2), int32(y), rl.Color{R: 220, G: 240, B: 255, A: alpha})
		rl.DrawPixel(int32(x+2), int32(y), rl.Color{R: 220, G: 240, B: 255, A: alpha})
		rl.DrawPixel(int32(x), int32(y-2), rl.Color{R: 220, G: 240, B: 255, A: alpha})
		rl.DrawPixel(int32(x), int32(y+2), rl.Color{R: 220, G: 240, B: 255, A: alpha})
	}
	// Icy glow ring
	for i := 0; i < 12; i++ {
		angle := float64(time*1.5) + float64(i)*0.52
		radius := 25.0 + 5.0*math.Sin(float64(time*2.0)+float64(i))
		x := cx + float32(math.Cos(angle)*radius)
		y := cy - 20 + float32(math.Sin(angle)*radius*0.6)
		rl.DrawCircle(int32(x), int32(y), 3, rl.Color{R: 150, G: 200, B: 255, A: 150})
	}
}

// drawAuraElectric - Gentle static electricity sparks
func (r *Renderer) drawAuraElectric(cx, cy, time float32) {
	// Soft glow around Claude
	glowAlpha := uint8(80 + 40*math.Sin(float64(time*2.0)))
	rl.DrawCircle(int32(cx), int32(cy-20), 20, rl.Color{R: 200, G: 220, B: 255, A: glowAlpha / 3})

	// A few gentle arcs
	for i := 0; i < 6; i++ {
		// Only show 2-3 at a time
		if int(time*3+float32(i)*1.7)%3 != 0 {
			continue
		}
		angle := float64(time*1.5) + float64(i)*1.05
		length := 18.0 + 8.0*math.Sin(float64(time*3.0)+float64(i))
		x1 := cx + float32(math.Cos(angle)*8)
		y1 := cy - 20 + float32(math.Sin(angle)*8)
		x2 := cx + float32(math.Cos(angle)*length)
		y2 := cy - 20 + float32(math.Sin(angle)*length)

		// Single gentle arc
		midX := (x1+x2)/2 + float32(math.Sin(float64(time*5+float32(i)))*3)
		midY := (y1+y2)/2 + float32(math.Cos(float64(time*5+float32(i)))*3)

		rl.DrawLine(int32(x1), int32(y1), int32(midX), int32(midY), rl.Color{R: 180, G: 200, B: 255, A: 200})
		rl.DrawLine(int32(midX), int32(midY), int32(x2), int32(y2), rl.Color{R: 220, G: 230, B: 255, A: 150})
	}

	// Small floating spark particles
	for i := 0; i < 10; i++ {
		angle := float64(time*2.0) + float64(i)*0.63
		radius := 15.0 + 10.0*math.Sin(float64(time*1.5)+float64(i)*0.8)
		x := cx + float32(math.Cos(angle)*radius)
		y := cy - 20 + float32(math.Sin(angle)*radius*0.7)
		alpha := uint8(150 + 80*math.Sin(float64(time*3.0)+float64(i)))
		rl.DrawCircle(int32(x), int32(y), 1, rl.Color{R: 200, G: 220, B: 255, A: alpha})
	}
}

// drawAuraShadow - Swirling dark purple/black wisps
func (r *Renderer) drawAuraShadow(cx, cy, time float32) {
	// Outer shadow ring
	for i := 0; i < 16; i++ {
		angle := float64(time*1.2) + float64(i)*0.39
		radius := 30.0 + 12.0*math.Sin(float64(time*1.5)+float64(i)*0.7)
		x := cx + float32(math.Cos(angle)*radius)
		y := cy - 15 + float32(math.Sin(angle)*radius*0.5)
		alpha := uint8(180 + 70*math.Sin(float64(time*2.0)+float64(i)))
		// Layered dark wisps
		rl.DrawCircle(int32(x), int32(y), 5, rl.Color{R: 40, G: 20, B: 60, A: alpha / 2})
		rl.DrawCircle(int32(x), int32(y), 3, rl.Color{R: 60, G: 30, B: 90, A: alpha})
		rl.DrawCircle(int32(x), int32(y), 1, rl.Color{R: 100, G: 50, B: 140, A: alpha})
	}
	// Inner void particles
	for i := 0; i < 8; i++ {
		angle := float64(-time*2.0) + float64(i)*0.785
		radius := 12.0 + 4.0*math.Sin(float64(time*3.0)+float64(i))
		x := cx + float32(math.Cos(angle)*radius)
		y := cy - 20 + float32(math.Sin(angle)*radius*0.6)
		rl.DrawCircle(int32(x), int32(y), 2, rl.Color{R: 20, G: 10, B: 30, A: 200})
	}
}

// drawAuraHeart - Floating pink/red hearts
func (r *Renderer) drawAuraHeart(cx, cy, time float32) {
	for i := 0; i < 14; i++ {
		baseX := cx + float32((i*11)%40-20)
		phase := float64(time*2.5) + float64(i)*0.6
		yOffset := float32(math.Mod(phase, 2.5) * 30)
		x := baseX + float32(math.Sin(phase*2.0)*8)
		y := cy - yOffset
		alpha := uint8(255 - yOffset*5)
		if alpha < 40 {
			alpha = 40
		}
		// Bigger heart shape
		c := rl.Color{R: 255, G: 80, B: 130, A: alpha}
		cLight := rl.Color{R: 255, G: 150, B: 180, A: alpha}
		// Heart made of circles
		rl.DrawCircle(int32(x-2), int32(y), 2, c)
		rl.DrawCircle(int32(x+2), int32(y), 2, c)
		rl.DrawCircle(int32(x), int32(y+2), 2, c)
		rl.DrawPixel(int32(x-3), int32(y+1), c)
		rl.DrawPixel(int32(x+3), int32(y+1), c)
		rl.DrawPixel(int32(x), int32(y+4), c)
		// Highlight
		rl.DrawPixel(int32(x-2), int32(y-1), cLight)
	}
}

// drawAuraCode - Matrix-style falling green code
func (r *Renderer) drawAuraCode(cx, cy, time float32) {
	codeChars := []rune{'0', '1', '{', '}', '<', '>', '/', '*', '#', '@', '$'}
	// Multiple columns of falling code
	for i := 0; i < 16; i++ {
		baseX := cx + float32((i*9)%60-30)
		phase := float64(time*4.0) + float64(i)*0.4
		yOffset := float32(math.Mod(phase, 2.0) * 45)
		y := cy - 55 + yOffset
		alpha := uint8(255 - yOffset*3)
		if alpha < 50 {
			alpha = 50
		}
		charIdx := (i + int(time*8)) % len(codeChars)
		char := string(codeChars[charIdx])
		// Bright green with glow
		rl.DrawText(char, int32(baseX), int32(y), 10, rl.Color{R: 0, G: 255, B: 80, A: alpha})
		// Trail
		if yOffset > 10 {
			charIdx2 := (i + int(time*8) + 3) % len(codeChars)
			rl.DrawText(string(codeChars[charIdx2]), int32(baseX), int32(y-10), 10, rl.Color{R: 0, G: 200, B: 60, A: alpha / 2})
		}
	}
}

// drawAuraRainbow - Vibrant cycling rainbow ring
func (r *Renderer) drawAuraRainbow(cx, cy, time float32) {
	// Outer rainbow ring
	for i := 0; i < 32; i++ {
		angle := float64(time*3.0) + float64(i)*0.196
		radius := 30.0 + 8.0*math.Sin(float64(time*4.0)+float64(i)*0.3)
		x := cx + float32(math.Cos(angle)*radius)
		y := cy - 20 + float32(math.Sin(angle)*radius*0.6)
		hue := int(math.Mod(float64(i)*11.25+float64(time*200), 360))
		color := hsvToRGB(hue, 1.0, 1.0)
		color.A = 255
		rl.DrawCircle(int32(x), int32(y), 3, color)
		color.A = 150
		rl.DrawCircle(int32(x), int32(y), 5, color)
	}
	// Inner glow
	for i := 0; i < 12; i++ {
		angle := float64(-time*2.0) + float64(i)*0.52
		radius := 15.0
		x := cx + float32(math.Cos(angle)*radius)
		y := cy - 20 + float32(math.Sin(angle)*radius*0.6)
		hue := int(math.Mod(float64(i)*30+float64(time*300), 360))
		color := hsvToRGB(hue, 1.0, 1.0)
		color.A = 200
		rl.DrawCircle(int32(x), int32(y), 2, color)
	}
}

// ============================================================================
// TRAIL SYSTEM - Particles behind Claude when walking
// ============================================================================

// spawnTrailParticles spawns trail particles when Claude is walking
func (r *Renderer) spawnTrailParticles(state *AnimationState) {
	if r.currentTrail < 0 || r.currentTrail >= len(r.trailNames) {
		return
	}

	// Only spawn when walking
	if state.CurrentAnim != AnimWalk {
		return
	}

	// Spawn every few frames
	if rand.Float32() > 0.3 {
		return
	}

	trailName := r.trailNames[r.currentTrail]
	// Spawn position behind Claude
	spawnX := float32(claudeX + spriteFrameWidth - 5)
	spawnY := float32(claudeY + spriteFrameHeight*2 - 8)

	switch trailName {
	case "trail_sparkle":
		r.spawnTrailSparkle(spawnX, spawnY)
	case "trail_flame":
		r.spawnTrailFlame(spawnX, spawnY)
	case "trail_frost":
		r.spawnTrailFrost(spawnX, spawnY)
	case "trail_hearts":
		r.spawnTrailHearts(spawnX, spawnY)
	case "trail_pixel":
		r.spawnTrailPixel(spawnX, spawnY)
	case "trail_rainbow":
		r.spawnTrailRainbow(spawnX, spawnY)
	}
}

func (r *Renderer) spawnTrailSparkle(x, y float32) {
	p := Particle{
		X:       x + float32(rand.Intn(10)-5),
		Y:       y + float32(rand.Intn(6)-3),
		VX:      float32(rand.Float32()*2 - 1),
		VY:      float32(rand.Float32()*-2 - 1),
		Life:    0.8,
		MaxLife: 0.8,
		Color:   rl.Color{R: 255, G: 255, B: 200, A: 255},
		Size:    1,
	}
	r.trailParticles = append(r.trailParticles, p)
}

func (r *Renderer) spawnTrailFlame(x, y float32) {
	p := Particle{
		X:       x + float32(rand.Intn(8)-4),
		Y:       y + float32(rand.Intn(4)),
		VX:      float32(rand.Float32()*1 - 0.5),
		VY:      float32(rand.Float32()*-3 - 1),
		Life:    0.6,
		MaxLife: 0.6,
		Color:   rl.Color{R: 255, G: uint8(150 + rand.Intn(100)), B: 50, A: 255},
		Size:    2,
	}
	r.trailParticles = append(r.trailParticles, p)
}

func (r *Renderer) spawnTrailFrost(x, y float32) {
	p := Particle{
		X:       x + float32(rand.Intn(12)-6),
		Y:       y + float32(rand.Intn(4)),
		VX:      float32(rand.Float32()*2 - 1),
		VY:      float32(rand.Float32()*0.5 + 0.2),
		Life:    1.0,
		MaxLife: 1.0,
		Color:   rl.Color{R: 150, G: 200, B: 255, A: 200},
		Size:    1,
	}
	r.trailParticles = append(r.trailParticles, p)
}

func (r *Renderer) spawnTrailHearts(x, y float32) {
	if rand.Float32() > 0.5 {
		return // Less frequent
	}
	p := Particle{
		X:       x + float32(rand.Intn(10)-5),
		Y:       y,
		VX:      float32(rand.Float32()*1 - 0.5),
		VY:      float32(rand.Float32()*-2 - 0.5),
		Life:    1.2,
		MaxLife: 1.2,
		Color:   rl.Color{R: 255, G: 100, B: 150, A: 255},
		Size:    2,
	}
	r.trailParticles = append(r.trailParticles, p)
}

func (r *Renderer) spawnTrailPixel(x, y float32) {
	colors := []rl.Color{
		{R: 255, G: 100, B: 100, A: 255},
		{R: 100, G: 255, B: 100, A: 255},
		{R: 100, G: 100, B: 255, A: 255},
		{R: 255, G: 255, B: 100, A: 255},
	}
	p := Particle{
		X:       x + float32(rand.Intn(8)-4),
		Y:       y + float32(rand.Intn(4)),
		VX:      float32(rand.Float32()*2 - 1),
		VY:      float32(rand.Float32()*-1 - 0.5),
		Life:    0.7,
		MaxLife: 0.7,
		Color:   colors[rand.Intn(len(colors))],
		Size:    1,
	}
	r.trailParticles = append(r.trailParticles, p)
}

func (r *Renderer) spawnTrailRainbow(x, y float32) {
	hue := int(math.Mod(float64(rl.GetTime()*180), 360))
	color := hsvToRGB(hue, 1.0, 1.0)
	p := Particle{
		X:       x + float32(rand.Intn(10)-5),
		Y:       y + float32(rand.Intn(4)),
		VX:      float32(rand.Float32()*2 - 1),
		VY:      float32(rand.Float32()*-2 - 0.5),
		Life:    0.9,
		MaxLife: 0.9,
		Color:   color,
		Size:    2,
	}
	r.trailParticles = append(r.trailParticles, p)
}

// updateTrailParticles updates trail particle positions and lifetimes
func (r *Renderer) updateTrailParticles() {
	dt := rl.GetFrameTime()
	alive := r.trailParticles[:0]
	for i := range r.trailParticles {
		p := &r.trailParticles[i]
		p.Life -= dt
		if p.Life > 0 {
			p.X += p.VX * dt * 30
			p.Y += p.VY * dt * 30
			alive = append(alive, *p)
		}
	}
	r.trailParticles = alive
}

// drawTrailParticles renders all trail particles
func (r *Renderer) drawTrailParticles() {
	for i := range r.trailParticles {
		p := &r.trailParticles[i]
		alpha := uint8(float32(p.Color.A) * (p.Life / p.MaxLife))
		color := rl.Color{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
		if p.Size <= 1 {
			rl.DrawPixel(int32(p.X), int32(p.Y), color)
		} else {
			rl.DrawCircle(int32(p.X), int32(p.Y), p.Size, color)
		}
	}
}

// ============================================================================
// AURA/TRAIL CYCLING FUNCTIONS
// ============================================================================

// CycleAura cycles to the next owned aura (or no aura)
func (r *Renderer) CycleAura(direction int) {
	if len(r.auraNames) == 0 {
		return
	}
	// Try up to len+1 times to find an owned aura (or -1 for none)
	for i := 0; i <= len(r.auraNames); i++ {
		r.currentAura += direction
		if r.currentAura >= len(r.auraNames) {
			r.currentAura = -1
		} else if r.currentAura < -1 {
			r.currentAura = len(r.auraNames) - 1
		}
		// -1 (no aura) is always valid
		if r.currentAura == -1 {
			return
		}
		// Check if this aura is owned
		if r.profile != nil && r.currentAura >= 0 && r.currentAura < len(r.auraNames) {
			if r.profile.IsOwned(r.auraNames[r.currentAura]) {
				return
			}
		} else if r.profile == nil {
			// No profile = all items available (demo mode)
			return
		}
	}
}

// GetCurrentAuraName returns the name of the current aura or empty string
func (r *Renderer) GetCurrentAuraName() string {
	if r.currentAura < 0 || r.currentAura >= len(r.auraNames) {
		return ""
	}
	return r.auraNames[r.currentAura]
}

// CycleTrail cycles to the next owned trail (or no trail)
func (r *Renderer) CycleTrail(direction int) {
	if len(r.trailNames) == 0 {
		return
	}
	// Try up to len+1 times to find an owned trail (or -1 for none)
	for i := 0; i <= len(r.trailNames); i++ {
		r.currentTrail += direction
		if r.currentTrail >= len(r.trailNames) {
			r.currentTrail = -1
		} else if r.currentTrail < -1 {
			r.currentTrail = len(r.trailNames) - 1
		}
		// -1 (no trail) is always valid
		if r.currentTrail == -1 {
			return
		}
		// Check if this trail is owned
		if r.profile != nil && r.currentTrail >= 0 && r.currentTrail < len(r.trailNames) {
			if r.profile.IsOwned(r.trailNames[r.currentTrail]) {
				return
			}
		} else if r.profile == nil {
			// No profile = all items available (demo mode)
			return
		}
	}
}

// GetCurrentTrailName returns the name of the current trail or empty string
func (r *Renderer) GetCurrentTrailName() string {
	if r.currentTrail < 0 || r.currentTrail >= len(r.trailNames) {
		return ""
	}
	return r.trailNames[r.currentTrail]
}
