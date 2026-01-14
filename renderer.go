package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// AccessoryPrefs stores user's accessory preferences
type AccessoryPrefs struct {
	HatName  string `json:"hat"`
	FaceName string `json:"face"`
}

func getPrefsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude-quest-prefs.json")
}

const (
	spriteFrameWidth  = 32
	spriteFrameHeight = 32
	spriteMaxFrames   = 12
	claudeScale       = 2 // Draw Claude 2x bigger
)

// Particle represents a visual effect particle
type Particle struct {
	X, Y     float32
	VX, VY   float32
	Life     float32
	MaxLife  float32
	Color    rl.Color
	Size     float32
}

// Renderer handles all drawing operations
type Renderer struct {
	config      *Config
	background  rl.Texture2D
	spriteSheet rl.Texture2D
	particles   []Particle
	hasSprites  bool

	// Accessories
	hats        []rl.Texture2D
	hatNames    []string
	currentHat  int // -1 = no hat
	faces       []rl.Texture2D
	faceNames   []string
	currentFace int // -1 = no face accessory

	// UI state
	activeRow int // 0 = HAT, 1 = FACE
}

// NewRenderer creates a new renderer with loaded assets
func NewRenderer(config *Config) *Renderer {
	r := &Renderer{
		config:      config,
		particles:   make([]Particle, 0, 100),
		currentHat:  -1, // No hat by default
		currentFace: -1, // No face accessory by default
	}

	// Try to load sprite sheet
	if _, err := os.Stat("assets/claude/spritesheet.png"); err == nil {
		r.spriteSheet = rl.LoadTexture("assets/claude/spritesheet.png")
		r.hasSprites = true
		fmt.Println("Loaded sprite sheet")
	} else {
		fmt.Println("No sprite sheet found, using placeholder graphics")
	}

	// Load all accessories
	r.loadHats()
	r.loadFaces()

	// Load user preferences
	r.LoadPrefs()

	return r
}

// loadHats loads all hat textures from assets/accessories/hats
func (r *Renderer) loadHats() {
	hatFiles := []string{"wizard", "party", "crown", "tophat", "propeller"}

	for _, name := range hatFiles {
		path := fmt.Sprintf("assets/accessories/hats/%s.png", name)
		if _, err := os.Stat(path); err == nil {
			tex := rl.LoadTexture(path)
			r.hats = append(r.hats, tex)
			r.hatNames = append(r.hatNames, name)
			fmt.Printf("Loaded hat: %s\n", name)
		}
	}
}

// CycleHat cycles to the next hat (or no hat)
func (r *Renderer) CycleHat(direction int) {
	if len(r.hats) == 0 {
		return
	}
	// -1 = no hat, 0 to len-1 = hats
	r.currentHat += direction
	if r.currentHat >= len(r.hats) {
		r.currentHat = -1
	} else if r.currentHat < -1 {
		r.currentHat = len(r.hats) - 1
	}
}

// GetCurrentHatName returns the name of the current hat or empty string
func (r *Renderer) GetCurrentHatName() string {
	if r.currentHat < 0 || r.currentHat >= len(r.hatNames) {
		return ""
	}
	return r.hatNames[r.currentHat]
}

// loadFaces loads all face accessory textures
func (r *Renderer) loadFaces() {
	faceFiles := []string{"dealwithit", "mustache", "monocle", "borat"}

	for _, name := range faceFiles {
		path := fmt.Sprintf("assets/accessories/faces/%s.png", name)
		if _, err := os.Stat(path); err == nil {
			tex := rl.LoadTexture(path)
			r.faces = append(r.faces, tex)
			r.faceNames = append(r.faceNames, name)
			fmt.Printf("Loaded face: %s\n", name)
		}
	}
}

// CycleFace cycles to the next face accessory (or none)
func (r *Renderer) CycleFace(direction int) {
	if len(r.faces) == 0 {
		return
	}
	r.currentFace += direction
	if r.currentFace >= len(r.faces) {
		r.currentFace = -1
	} else if r.currentFace < -1 {
		r.currentFace = len(r.faces) - 1
	}
}

// GetCurrentFaceName returns the name of the current face accessory
func (r *Renderer) GetCurrentFaceName() string {
	if r.currentFace < 0 || r.currentFace >= len(r.faceNames) {
		return ""
	}
	return r.faceNames[r.currentFace]
}

// SwitchRow switches between HAT (0) and FACE (1) rows
func (r *Renderer) SwitchRow(direction int) {
	r.activeRow += direction
	if r.activeRow < 0 {
		r.activeRow = 1
	} else if r.activeRow > 1 {
		r.activeRow = 0
	}
}

// CycleActive cycles the currently active row's accessory
func (r *Renderer) CycleActive(direction int) {
	if r.activeRow == 0 {
		r.CycleHat(direction)
	} else {
		r.CycleFace(direction)
	}
	r.SavePrefs()
}

// SavePrefs saves current accessory choices to disk
func (r *Renderer) SavePrefs() {
	prefs := AccessoryPrefs{
		HatName:  r.GetCurrentHatName(),
		FaceName: r.GetCurrentFaceName(),
	}
	data, _ := json.Marshal(prefs)
	os.WriteFile(getPrefsPath(), data, 0644)
}

// LoadPrefs loads accessory choices from disk
func (r *Renderer) LoadPrefs() {
	data, err := os.ReadFile(getPrefsPath())
	if err != nil {
		return
	}
	var prefs AccessoryPrefs
	if json.Unmarshal(data, &prefs) != nil {
		return
	}
	// Find and set hat by name
	for i, name := range r.hatNames {
		if name == prefs.HatName {
			r.currentHat = i
			break
		}
	}
	// Find and set face by name
	for i, name := range r.faceNames {
		if name == prefs.FaceName {
			r.currentFace = i
			break
		}
	}
}

// Draw renders the current animation state
func (r *Renderer) Draw(state *AnimationState) {
	// Draw background
	r.drawBackground()

	// Update and draw particles
	r.updateParticles()
	r.drawParticles()

	// Draw Claude sprite
	r.drawClaude(state)

	// Draw face accessory (under hat)
	r.drawFace(state)

	// Draw hat on top of Claude
	r.drawHat(state)

	// Spawn new particles based on animation
	r.spawnParticles(state)

	// Draw debug info if enabled
	if r.config.Debug {
		r.drawDebug(state)
	}
}

func (r *Renderer) drawBackground() {
	// Draw a wizard's study background

	// Back wall
	rl.DrawRectangle(0, 0, screenWidth, 160, rl.Color{R: 35, G: 30, B: 45, A: 255})

	// Floor
	rl.DrawRectangle(0, 160, screenWidth, 40, rl.Color{R: 60, G: 50, B: 70, A: 255})

	// Floor boards
	for i := int32(0); i < screenWidth; i += 40 {
		rl.DrawLine(i, 160, i+20, 200, rl.Color{R: 50, G: 42, B: 60, A: 255})
	}

	// Bookshelf left
	rl.DrawRectangle(5, 40, 50, 120, rl.Color{R: 80, G: 55, B: 45, A: 255})
	rl.DrawRectangle(8, 45, 44, 25, rl.Color{R: 45, G: 35, B: 50, A: 255})
	rl.DrawRectangle(8, 75, 44, 25, rl.Color{R: 45, G: 35, B: 50, A: 255})
	rl.DrawRectangle(8, 105, 44, 25, rl.Color{R: 45, G: 35, B: 50, A: 255})
	// Books
	colors := []rl.Color{
		{R: 180, G: 80, B: 80, A: 255},
		{R: 80, G: 120, B: 180, A: 255},
		{R: 80, G: 160, B: 100, A: 255},
		{R: 200, G: 180, B: 100, A: 255},
	}
	for row := 0; row < 3; row++ {
		for book := 0; book < 6; book++ {
			bx := int32(10 + book*7)
			by := int32(48 + row*30)
			bh := int32(18 + (book*3)%8)
			rl.DrawRectangle(bx, by+(22-bh), 5, bh, colors[(book+row)%4])
		}
	}

	// Desk right
	rl.DrawRectangle(240, 100, 70, 60, rl.Color{R: 90, G: 60, B: 50, A: 255})
	rl.DrawRectangle(242, 95, 66, 8, rl.Color{R: 100, G: 70, B: 55, A: 255})
	// Scroll on desk
	rl.DrawRectangle(250, 98, 30, 5, rl.Color{R: 230, G: 220, B: 190, A: 255})

	// Window
	rl.DrawRectangle(130, 30, 60, 80, rl.Color{R: 25, G: 35, B: 60, A: 255})
	rl.DrawRectangle(132, 32, 56, 76, rl.Color{R: 40, G: 50, B: 80, A: 255})
	// Window frame
	rl.DrawRectangle(158, 32, 4, 76, rl.Color{R: 70, G: 50, B: 45, A: 255})
	rl.DrawRectangle(132, 68, 56, 4, rl.Color{R: 70, G: 50, B: 45, A: 255})
	// Stars in window
	rl.DrawPixel(140, 45, rl.Color{R: 255, G: 255, B: 200, A: 255})
	rl.DrawPixel(175, 55, rl.Color{R: 255, G: 255, B: 200, A: 255})
	rl.DrawPixel(150, 85, rl.Color{R: 255, G: 255, B: 200, A: 255})
	rl.DrawPixel(170, 95, rl.Color{R: 200, G: 200, B: 255, A: 255})

	// Candle on desk
	rl.DrawRectangle(290, 85, 6, 12, rl.Color{R: 230, G: 220, B: 180, A: 255})
	// Flame (animated via frame count)
	flicker := int32(rl.GetFrameTime()*1000) % 3
	rl.DrawRectangle(291+flicker%2, 78, 4, 7, rl.Color{R: 255, G: 200, B: 100, A: 255})
	rl.DrawRectangle(292, 80, 2, 4, rl.Color{R: 255, G: 255, B: 200, A: 255})
}

func (r *Renderer) drawClaude(state *AnimationState) {
	// Scaled dimensions
	scaledW := float32(spriteFrameWidth * claudeScale)
	scaledH := float32(spriteFrameHeight * claudeScale)

	// Position Claude in center of scene
	x := float32(screenWidth/2) - scaledW/2
	y := float32(160) - scaledH + 10 // Feet on floor

	if r.hasSprites {
		// Calculate source rectangle from sprite sheet
		frameX := float32(state.Frame * spriteFrameWidth)
		frameY := float32(int(state.CurrentAnim) * spriteFrameHeight)

		sourceRec := rl.Rectangle{
			X:      frameX,
			Y:      frameY,
			Width:  spriteFrameWidth,
			Height: spriteFrameHeight,
		}

		destRec := rl.Rectangle{
			X:      x,
			Y:      y,
			Width:  scaledW,
			Height: scaledH,
		}

		rl.DrawTexturePro(r.spriteSheet, sourceRec, destRec, rl.Vector2{}, 0, rl.White)
	} else {
		// Fallback placeholder
		r.drawPlaceholderClaude(int(x), int(y), state)
	}
}

// getHeadOffset returns the X,Y offset of Claude's head for the current animation frame
// These offsets EXACTLY match the sprite generator (cmd/spritegen/main.go)
func getHeadOffset(state *AnimationState) (float32, float32) {
	f := state.Frame

	switch state.CurrentAnim {
	case AnimIdle:
		// breathCurve from spritegen - subtracted from bodyTop, so positive = UP
		breathCurve := []int{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0}
		return 0, float32(-breathCurve[f%len(breathCurve)]) // Negate: breath up = hat up

	case AnimEnter:
		// Frames 0-7: sparkles only, no Claude visible
		if f < 8 {
			return 0, -100 // Off screen
		} else if f < 15 {
			// Frames 8-14: materializing (squash effect, roughly centered)
			return 0, 0
		} else {
			// Frames 15-19: bounce settle
			bounce := []int{-2, -1, 0, 0, 0}
			return 0, float32(bounce[f-15])
		}

	case AnimCasting:
		if f < 5 {
			// Wind up at oy+1
			return 0, 1
		} else if f < 13 {
			// Floating up
			floatY := []int{-1, -2, -2, -2, -1, -1, -2, -2}
			return 0, float32(floatY[f-5])
		} else {
			// Settle
			settleY := []int{-1, 0, 0}
			idx := f - 13
			if idx >= len(settleY) {
				idx = len(settleY) - 1
			}
			return 0, float32(settleY[idx])
		}

	case AnimAttack:
		if f < 3 {
			// oy+1+frame/2
			return 0, float32(1 + f/2)
		} else if f < 5 {
			// oy+3
			return 0, 3
		} else if f < 7 {
			// oy+2
			return 0, 2
		} else if f == 7 {
			// Smear frame
			return 0, 0
		} else if f < 10 {
			// ox+3, oy+1
			return 3, 1
		} else if f < 14 {
			// Follow-through
			bounceY := []int{-2, -1, 0, 1}
			xOff := 2 - (f-10)/2
			return float32(xOff), float32(bounceY[f-10])
		} else {
			// Recovery
			bounce := []int{-1, 0}
			idx := f - 14
			if idx >= len(bounce) {
				idx = len(bounce) - 1
			}
			return 0, float32(bounce[idx])
		}

	case AnimWriting:
		// bob is subtracted from bodyTop in spritegen, so positive = UP
		bobCurve := []int{0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0}
		return 0, float32(-bobCurve[f%len(bobCurve)]) // Negate: bob up = hat up

	case AnimVictory:
		if f < 4 {
			// Anticipation at oy+2
			return 0, 2
		} else if f < 9 {
			// Jump up
			jumpY := []int{0, 3, 6, 9, 11}
			return 0, float32(-jumpY[f-4])
		} else if f < 12 {
			// Peak with wiggle
			wiggle := []int{0, 1, 0}
			return float32(wiggle[f-9]), -12
		} else if f < 16 {
			// Fall down
			jumpY := []int{10, 7, 4, 1}
			return 0, float32(-jumpY[f-12])
		} else {
			// Land bounce
			bounceY := []int{2, 0, -1, 0}
			idx := f - 16
			if idx >= len(bounceY) {
				idx = len(bounceY) - 1
			}
			return 0, float32(bounceY[idx])
		}

	case AnimHurt:
		if f < 3 {
			// Impact at oy+2
			return 0, 2
		} else if f < 9 {
			// Knockback
			knockX := []int{2, 5, 7, 8, 7, 5}
			return float32(-knockX[f-3]), 0
		} else {
			// Recovery
			recoverX := []int{4, 3, 2, 1, 0, 0, 0}
			bounceY := []int{-1, 0, 1, 0, -1, 0, 0}
			idx := f - 9
			if idx >= len(recoverX) {
				idx = len(recoverX) - 1
			}
			return float32(-recoverX[idx]), float32(bounceY[idx])
		}

	case AnimThinking:
		sway := []int{0, 0, 0, 1, 1, 1, 0, 0, 0, -1, -1, -1}
		return float32(sway[f%len(sway)]), 0
	}

	return 0, 0
}

func (r *Renderer) drawHat(state *AnimationState) {
	if r.currentHat < 0 || r.currentHat >= len(r.hats) {
		return
	}

	hat := r.hats[r.currentHat]

	// Get animation-specific offset (in sprite pixels, before scaling)
	headOffX, headOffY := getHeadOffset(state)

	// Don't draw if off-screen (e.g., during Enter sparkles phase)
	if headOffY < -50 {
		return
	}

	// Claude's base position in SCREEN coords (same as drawClaude)
	scaledW := float32(spriteFrameWidth * claudeScale)
	scaledH := float32(spriteFrameHeight * claudeScale)
	claudeX := float32(screenWidth/2) - scaledW/2
	claudeY := float32(160) - scaledH + 10

	// Hat dimensions (scale with Claude)
	hatW := float32(hat.Width) * float32(claudeScale)
	hatH := float32(hat.Height) * float32(claudeScale)

	// In sprite space, Claude's body top is at y=12
	// Hat should sit ON TOP of the body, so just above y=12
	// Hat bottom edge should be around sprite y=10
	spriteHeadY := float32(10) // Where top of head is in 32x32 sprite

	// Convert to screen coords:
	// claudeY is the top-left of the 32x32 sprite frame (scaled)
	// Add spriteHeadY * scale to get head position
	// Add animation offset * scale
	hatX := claudeX + scaledW/2 - hatW/2 + headOffX*float32(claudeScale)
	hatY := claudeY + (spriteHeadY+headOffY)*float32(claudeScale) - hatH + 2*float32(claudeScale)

	sourceRec := rl.Rectangle{
		X:      0,
		Y:      0,
		Width:  float32(hat.Width),
		Height: float32(hat.Height),
	}

	destRec := rl.Rectangle{
		X:      hatX,
		Y:      hatY,
		Width:  hatW,
		Height: hatH,
	}

	rl.DrawTexturePro(hat, sourceRec, destRec, rl.Vector2{}, 0, rl.White)
}

func (r *Renderer) drawFace(state *AnimationState) {
	if r.currentFace < 0 || r.currentFace >= len(r.faces) {
		return
	}

	face := r.faces[r.currentFace]
	faceName := r.faceNames[r.currentFace]

	// Get animation-specific offset
	headOffX, headOffY := getHeadOffset(state)

	// Don't draw if off-screen
	if headOffY < -50 {
		return
	}

	// Claude's base position
	scaledW := float32(spriteFrameWidth * claudeScale)
	scaledH := float32(spriteFrameHeight * claudeScale)
	claudeX := float32(screenWidth/2) - scaledW/2
	claudeY := float32(160) - scaledH + 10

	// Face accessory dimensions
	faceW := float32(face.Width) * float32(claudeScale)
	faceH := float32(face.Height) * float32(claudeScale)

	// Position depends on accessory type
	// Claude's eyes are at ~y=13-17, body at y=12-22
	// Right eye at x=18-20, left eye at x=11-13 (sprite center is x=16)
	var spriteY float32
	var spriteXOffset float32 = 0 // offset from center
	var centerVertically bool = true
	switch faceName {
	case "dealwithit":
		// Goes ON the eyes (y=13-17)
		spriteY = 15
	case "monocle":
		// Monocle on RIGHT eye (x=18-20, y=13-16)
		// Right eye center is ~x=19, sprite center is x=16, so offset +3
		spriteY = 13
		spriteXOffset = 3
		centerVertically = false // position from top of sprite
	case "mustache":
		// Below eyes, on the "mouth" area
		spriteY = 19
	case "borat":
		// Mankini - straps at shoulder level (body top y=12), pouch at bottom
		// Sprite is 11px tall, center at y=18 puts straps at ~12-13, pouch at ~23
		spriteY = 18
	default:
		spriteY = 16
	}

	faceX := claudeX + scaledW/2 - faceW/2 + (headOffX+spriteXOffset)*float32(claudeScale)
	faceY := claudeY + (spriteY+headOffY)*float32(claudeScale)
	if centerVertically {
		faceY -= faceH / 2
	}

	sourceRec := rl.Rectangle{
		X:      0,
		Y:      0,
		Width:  float32(face.Width),
		Height: float32(face.Height),
	}

	destRec := rl.Rectangle{
		X:      faceX,
		Y:      faceY,
		Width:  faceW,
		Height: faceH,
	}

	rl.DrawTexturePro(face, sourceRec, destRec, rl.Vector2{}, 0, rl.White)
}

func (r *Renderer) drawPlaceholderClaude(x, y int, state *AnimationState) {
	// Simple placeholder when no sprites loaded
	color := rl.Color{R: 217, G: 119, B: 87, A: 255}

	bobOffset := 0
	if state.CurrentAnim == AnimIdle {
		bobOffset = int(state.Frame/10) % 2
	}

	// Body
	rl.DrawRectangle(int32(x+8), int32(y+20), 16, 24, color)
	// Head
	rl.DrawCircle(int32(x+16), int32(y+14+bobOffset), 10, color)
	// Eyes
	rl.DrawCircle(int32(x+13), int32(y+12+bobOffset), 2, rl.White)
	rl.DrawCircle(int32(x+19), int32(y+12+bobOffset), 2, rl.White)
}

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

// DrawAccessoryPicker draws a clean, centered accessory picker panel
func (r *Renderer) DrawAccessoryPicker() {
	// Colors
	panelBg := rl.Color{R: 20, G: 18, B: 30, A: 230}
	panelBorder := rl.Color{R: 60, G: 55, B: 80, A: 255}
	labelDim := rl.Color{R: 80, G: 75, B: 100, A: 255}
	labelActive := rl.Color{R: 180, G: 175, B: 200, A: 255}
	valueDim := rl.Color{R: 120, G: 115, B: 140, A: 255}
	valueActive := rl.Color{R: 255, G: 200, B: 80, A: 255}
	arrowDim := rl.Color{R: 80, G: 75, B: 100, A: 255}
	arrowActive := rl.Color{R: 180, G: 175, B: 200, A: 255}

	// Layout constants - smaller
	rowH := int32(10)
	padding := int32(3)
	labelW := int32(24)
	arrowW := int32(5)
	valueW := int32(50)
	gap := int32(3)

	// Calculate panel size
	panelW := padding + labelW + gap + arrowW + valueW + arrowW + padding
	panelH := padding + rowH + gap + rowH + padding

	// Position in bottom-left corner
	panelX := int32(4)
	panelY := screenHeight - panelH - 4

	// Draw panel
	rl.DrawRectangle(panelX-1, panelY-1, panelW+2, panelH+2, panelBorder)
	rl.DrawRectangle(panelX, panelY, panelW, panelH, panelBg)

	// === ROW 1: HAT ===
	rowY := panelY + padding
	x := panelX + padding

	// Determine colors based on active row
	hatLabelColor := labelDim
	hatArrowColor := arrowDim
	if r.activeRow == 0 {
		hatLabelColor = labelActive
		hatArrowColor = arrowActive
	}

	// Label
	rl.DrawText("HAT", x, rowY+2, 8, hatLabelColor)
	x += labelW + gap

	// Left arrow
	rl.DrawText("<", x, rowY+2, 8, hatArrowColor)
	x += arrowW

	// Value
	hatName := r.GetCurrentHatName()
	hatColor := valueDim
	if hatName == "" {
		hatName = "-"
	} else if r.activeRow == 0 {
		hatColor = valueActive
	}
	hatTextW := rl.MeasureText(hatName, 8)
	hatTextX := x + (valueW-hatTextW)/2
	rl.DrawText(hatName, hatTextX, rowY+2, 8, hatColor)
	x += valueW

	// Right arrow
	rl.DrawText(">", x, rowY+2, 8, hatArrowColor)

	// === ROW 2: FACE ===
	rowY = panelY + padding + rowH + gap
	x = panelX + padding

	// Determine colors based on active row
	faceLabelColor := labelDim
	faceArrowColor := arrowDim
	if r.activeRow == 1 {
		faceLabelColor = labelActive
		faceArrowColor = arrowActive
	}

	// Label
	rl.DrawText("FACE", x, rowY+2, 8, faceLabelColor)
	x += labelW + gap

	// Left arrow
	rl.DrawText("<", x, rowY+2, 8, faceArrowColor)
	x += arrowW

	// Value
	faceName := r.GetCurrentFaceName()
	faceColor := valueDim
	if faceName == "" {
		faceName = "-"
	} else if r.activeRow == 1 {
		faceColor = valueActive
	}
	faceTextW := rl.MeasureText(faceName, 8)
	faceTextX := x + (valueW-faceTextW)/2
	rl.DrawText(faceName, faceTextX, rowY+2, 8, faceColor)
	x += valueW

	// Right arrow
	rl.DrawText(">", x, rowY+2, 8, faceArrowColor)
}

func (r *Renderer) drawDebug(state *AnimationState) {
	// Animation name
	rl.DrawText(state.CurrentAnim.String(), 5, 5, 8, rl.Green)

	// Frame counter
	frameText := fmt.Sprintf("Frame: %d", state.Frame)
	rl.DrawText(frameText, 5, 15, 8, rl.Green)

	// Particle count
	particleText := fmt.Sprintf("Particles: %d", len(r.particles))
	rl.DrawText(particleText, 5, 25, 8, rl.Green)

	// FPS
	fpsText := fmt.Sprintf("FPS: %d", rl.GetFPS())
	rl.DrawText(fpsText, 5, 35, 8, rl.Green)
}

// Unload frees all loaded textures
func (r *Renderer) Unload() {
	if r.background.ID != 0 {
		rl.UnloadTexture(r.background)
	}
	if r.spriteSheet.ID != 0 {
		rl.UnloadTexture(r.spriteSheet)
	}
}
