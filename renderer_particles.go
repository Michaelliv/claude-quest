package main

import (
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (r *Renderer) spawnParticles(state *AnimationState) {
	cx := float32(screenWidth / 2)
	cy := float32(160 - spriteFrameHeight*claudeScale/2) // Center of Claude

	switch state.CurrentAnim {
	case AnimCasting:
		// Magic sparkles
		if rand.Float32() < 0.3 {
			r.particles = append(r.particles, Particle{
				X:       cx + (rand.Float32()-0.5)*40,
				Y:       cy - 20 + (rand.Float32()-0.5)*20,
				VX:      (rand.Float32() - 0.5) * 20,
				VY:      -rand.Float32() * 30,
				Life:    1.0,
				MaxLife: 1.0,
				Color:   rl.Color{R: 255, G: 220, B: 120, A: 255},
				Size:    2,
			})
		}

	case AnimAttack:
		// Impact particles
		if state.Frame >= 4 && state.Frame <= 6 && rand.Float32() < 0.5 {
			r.particles = append(r.particles, Particle{
				X:       cx + 20,
				Y:       cy,
				VX:      rand.Float32() * 40,
				VY:      (rand.Float32() - 0.5) * 30,
				Life:    0.5,
				MaxLife: 0.5,
				Color:   rl.Color{R: 255, G: 255, B: 200, A: 255},
				Size:    3,
			})
		}

	case AnimWriting:
		// Ink dots
		if rand.Float32() < 0.1 {
			r.particles = append(r.particles, Particle{
				X:       cx + 15,
				Y:       cy + 10,
				VX:      rand.Float32() * 5,
				VY:      rand.Float32() * 10,
				Life:    0.8,
				MaxLife: 0.8,
				Color:   rl.Color{R: 30, G: 30, B: 50, A: 255},
				Size:    1,
			})
		}

	case AnimVictory:
		// Celebration sparkles
		if rand.Float32() < 0.4 {
			r.particles = append(r.particles, Particle{
				X:       cx + (rand.Float32()-0.5)*60,
				Y:       cy - 30,
				VX:      (rand.Float32() - 0.5) * 20,
				VY:      -rand.Float32() * 40,
				Life:    1.2,
				MaxLife: 1.2,
				Color:   rl.Color{R: 255, G: 255, B: 100, A: 255},
				Size:    2,
			})
		}

	case AnimHurt:
		// Impact stars
		if state.Frame < 3 && rand.Float32() < 0.4 {
			r.particles = append(r.particles, Particle{
				X:       cx + 10,
				Y:       cy - 10,
				VX:      rand.Float32() * 30,
				VY:      (rand.Float32() - 0.5) * 20,
				Life:    0.4,
				MaxLife: 0.4,
				Color:   rl.Color{R: 255, G: 100, B: 100, A: 255},
				Size:    2,
			})
		}

	case AnimThinking:
		// Thought bubbles
		if state.Frame == 3 && rand.Float32() < 0.2 {
			r.particles = append(r.particles, Particle{
				X:       cx + 20,
				Y:       cy - 30,
				VX:      2,
				VY:      -10,
				Life:    2.0,
				MaxLife: 2.0,
				Color:   rl.Color{R: 200, G: 200, B: 220, A: 200},
				Size:    4,
			})
		}
	}
}

func (r *Renderer) updateParticles() {
	dt := rl.GetFrameTime()
	alive := r.particles[:0]

	for i := range r.particles {
		p := &r.particles[i]
		p.Life -= dt
		if p.Life > 0 {
			p.X += p.VX * dt
			p.Y += p.VY * dt
			p.VY += 50 * dt // gravity
			alive = append(alive, *p)
		}
	}
	r.particles = alive
}

func (r *Renderer) drawParticles() {
	for _, p := range r.particles {
		alpha := uint8(255 * (p.Life / p.MaxLife))
		color := p.Color
		color.A = alpha

		size := int32(p.Size)
		if p.Size > 2 {
			rl.DrawCircle(int32(p.X), int32(p.Y), p.Size, color)
		} else {
			rl.DrawRectangle(int32(p.X), int32(p.Y), size, size, color)
		}
	}
}
