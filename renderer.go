package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

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

// getAssetPath returns the path to an asset file, checking both relative to
// the executable (for npm installs) and relative to cwd (for development)
func getAssetPath(relativePath string) string {
	// First try relative to executable (npm install location)
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		// Check if we're in a symlinked bin directory (npm global install)
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exeDir = filepath.Dir(resolved)
		}
		// npm installs put assets in ../assets relative to bin/cq
		npmAssetPath := filepath.Join(exeDir, "..", "assets", relativePath)
		if _, err := os.Stat(npmAssetPath); err == nil {
			return npmAssetPath
		}
		// Also check same directory as binary
		sameDirPath := filepath.Join(exeDir, "assets", relativePath)
		if _, err := os.Stat(sameDirPath); err == nil {
			return sameDirPath
		}
	}
	// Fall back to relative path (development)
	return filepath.Join("assets", relativePath)
}

const (
	spriteFrameWidth  = 32
	spriteFrameHeight = 32
	spriteMaxFrames   = 12
	claudeScale       = 2 // Draw Claude 2x bigger

	// Mini Claude sprites
	miniFrameWidth  = 16
	miniFrameHeight = 16
	miniMaxFrames   = 12

	// Enemy sprites
	enemyFrameWidth  = 32
	enemyFrameHeight = 16
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
	config           *Config
	background       rl.Texture2D
	spriteSheet      rl.Texture2D
	miniSpriteSheet  rl.Texture2D
	enemySpriteSheet rl.Texture2D
	particles        []Particle
	hasSprites       bool
	hasMiniSprites   bool
	hasEnemySprites  bool

	// Accessories
	hats        []rl.Texture2D
	hatNames    []string
	currentHat  int // -1 = no hat
	faces       []rl.Texture2D
	faceNames   []string
	currentFace int // -1 = no face accessory

	// UI state
	activeRow int // 0 = HAT, 1 = FACE

	// Parallax scrolling
	scrollOffset float32
	walkMode     bool

	// Biome system
	currentBiome int
	biomeTimer   float32

	// Picker visibility
	pickerExpanded bool
	pickerAnim     float32 // 0.0 = collapsed, 1.0 = expanded
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
	spritePath := getAssetPath("claude/spritesheet.png")
	if _, err := os.Stat(spritePath); err == nil {
		r.spriteSheet = rl.LoadTexture(spritePath)
		r.hasSprites = true
		fmt.Println("Loaded sprite sheet from:", spritePath)
	} else {
		fmt.Println("No sprite sheet found, using placeholder graphics")
	}

	// Try to load mini sprite sheet
	miniSpritePath := getAssetPath("claude/mini_spritesheet.png")
	if _, err := os.Stat(miniSpritePath); err == nil {
		r.miniSpriteSheet = rl.LoadTexture(miniSpritePath)
		r.hasMiniSprites = true
		fmt.Println("Loaded mini sprite sheet from:", miniSpritePath)
	}

	// Try to load enemy sprite sheet
	enemySpritePath := getAssetPath("enemies/enemy_spritesheet.png")
	if _, err := os.Stat(enemySpritePath); err == nil {
		r.enemySpriteSheet = rl.LoadTexture(enemySpritePath)
		r.hasEnemySprites = true
		fmt.Println("Loaded enemy sprite sheet from:", enemySpritePath)
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
		path := getAssetPath(fmt.Sprintf("accessories/hats/%s.png", name))
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
		path := getAssetPath(fmt.Sprintf("accessories/faces/%s.png", name))
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

// SwitchRow switches between HAT (0), FACE (1), and BIOME (2) rows
func (r *Renderer) SwitchRow(direction int) {
	r.activeRow += direction
	if r.activeRow < 0 {
		r.activeRow = 2
	} else if r.activeRow > 2 {
		r.activeRow = 0
	}
}

// CycleActive cycles the currently active row's accessory
func (r *Renderer) CycleActive(direction int) {
	switch r.activeRow {
	case 0:
		r.CycleHat(direction)
	case 1:
		r.CycleFace(direction)
	case 2:
		r.ToggleWalkMode()
	}
	r.SavePrefs()
}

// Walk mode names for the picker (funny names)
var modeNames = []string{"Vibin", "Quest!"}

// ToggleWalkMode toggles between idle and walk mode
func (r *Renderer) ToggleWalkMode() {
	r.walkMode = !r.walkMode
}

// IsWalkMode returns whether walk mode is active
func (r *Renderer) IsWalkMode() bool {
	return r.walkMode
}

// TogglePicker expands/collapses the accessory picker
func (r *Renderer) TogglePicker() {
	r.pickerExpanded = !r.pickerExpanded
}

// IsPickerExpanded returns whether the picker is visible
func (r *Renderer) IsPickerExpanded() bool {
	return r.pickerExpanded
}

// UpdatePickerAnim animates the picker expand/collapse
func (r *Renderer) UpdatePickerAnim(dt float32) {
	speed := float32(6.0) // Animation speed
	if r.pickerExpanded {
		r.pickerAnim += dt * speed
		if r.pickerAnim > 1.0 {
			r.pickerAnim = 1.0
		}
	} else {
		r.pickerAnim -= dt * speed
		if r.pickerAnim < 0.0 {
			r.pickerAnim = 0.0
		}
	}
}

// GetCurrentModeName returns a fun name for the current mode
func (r *Renderer) GetCurrentModeName() string {
	if r.walkMode {
		return modeNames[1] // "Quest!"
	}
	return modeNames[0] // "Vibin"
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

// SetWalkMode enables/disables the infinite walking parallax mode
func (r *Renderer) SetWalkMode(enabled bool) {
	r.walkMode = enabled
}

// UpdateScroll advances the parallax scroll (call each frame)
func (r *Renderer) UpdateScroll(dt float32) {
	if r.walkMode {
		r.scrollOffset += dt * 25 // Walk speed (slightly slower for atmosphere)
		// Wrap around to prevent overflow
		if r.scrollOffset > 10000 {
			r.scrollOffset -= 10000
		}

		// Biome transitions every ~20 seconds of walking
		r.biomeTimer += dt
		if r.biomeTimer > 20 {
			r.biomeTimer = 0
			r.currentBiome = (r.currentBiome + 1) % 4
		}
	}
}

func (r *Renderer) drawBackground() {
	if r.walkMode {
		r.drawParallaxBackground()
	} else {
		r.drawStudyBackground()
	}
}

// drawParallaxBackground renders the infinite scrolling landscape with biomes
func (r *Renderer) drawParallaxBackground() {
	switch r.currentBiome {
	case 0:
		r.drawBiomeEnchantedForest()
	case 1:
		r.drawBiomeMountainJourney()
	case 2:
		r.drawBiomeMidnightQuest()
	case 3:
		r.drawBiomeKingdomRoad()
	}
}

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
	sunX := int32(200) - int32(scroll*0.01)%screenWidth
	for i := int32(19); i >= 0; i-- {
		alpha := uint8(60 - i*3)
		rl.DrawCircle(sunX, 95, float32(i+5), rl.Color{R: 255, G: 200, B: 150, A: alpha})
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
		sy := int32((i*31+7) % 70)
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

// ============================================================================
// DRAWING HELPERS FOR ALL BIOMES
// ============================================================================

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

// drawStudyBackground draws the original wizard's study background
func (r *Renderer) drawStudyBackground() {
	time := float32(rl.GetTime())

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

	// Wall texture - subtle brick pattern
	wallBrick := rl.Color{R: 38, G: 32, B: 50, A: 80}
	for y := int32(10); y < 150; y += 16 {
		offset := int32(0)
		if (y/16)%2 == 1 {
			offset = 20
		}
		for x := int32(-20) + offset; x < screenWidth+20; x += 40 {
			rl.DrawRectangle(x, y, 38, 14, wallBrick)
		}
	}

	// === COZY RUG - Circular pattern under Claude ===
	rugCenter := int32(screenWidth / 2)
	rugColors := []rl.Color{
		{R: 120, G: 45, B: 60, A: 255},
		{R: 140, G: 55, B: 70, A: 255},
		{R: 90, G: 35, B: 50, A: 255},
	}
	for i := 3; i >= 0; i-- {
		radius := float32(45 - i*10)
		rl.DrawEllipse(rugCenter, 175, radius, radius*0.3, rugColors[i%3])
	}
	// Rug fringe
	for x := int32(rugCenter - 42); x < rugCenter+42; x += 4 {
		rl.DrawLine(x, 182, x+1, 186, rl.Color{R: 100, G: 40, B: 55, A: 255})
	}

	// === FLOOR - Warm wooden planks ===
	rl.DrawRectangle(0, 160, screenWidth, 40, rl.Color{R: 55, G: 42, B: 50, A: 255})
	plankColors := []rl.Color{
		{R: 65, G: 48, B: 55, A: 255},
		{R: 58, G: 44, B: 52, A: 255},
		{R: 62, G: 46, B: 54, A: 255},
	}
	for i := int32(0); i < 8; i++ {
		px := i * 42
		rl.DrawRectangle(px, 160, 40, 40, plankColors[int(i)%3])
		rl.DrawLine(px, 160, px, 200, rl.Color{R: 45, G: 35, B: 42, A: 255})
		// Wood grain
		rl.DrawLine(px+10, 165, px+12, 195, rl.Color{R: 50, G: 38, B: 45, A: 100})
		rl.DrawLine(px+25, 168, px+28, 198, rl.Color{R: 50, G: 38, B: 45, A: 100})
	}

	// === GRAND WINDOW - Arched with moon and stars ===
	// Window frame (ornate)
	rl.DrawRectangle(115, 15, 90, 105, rl.Color{R: 65, G: 45, B: 40, A: 255})
	rl.DrawRectangle(118, 18, 84, 99, rl.Color{R: 75, G: 52, B: 45, A: 255})

	// Night sky through window
	for y := int32(20); y < 115; y++ {
		t := float32(y-20) / 95.0
		c := rl.Color{
			R: uint8(15 + t*10),
			G: uint8(20 + t*15),
			B: uint8(45 + t*20),
			A: 255,
		}
		rl.DrawLine(120, y, 200, y, c)
	}

	// Moon with glow
	moonX, moonY := int32(175), int32(40)
	// Glow layers
	rl.DrawCircle(moonX, moonY, 18, rl.Color{R: 60, G: 60, B: 100, A: 30})
	rl.DrawCircle(moonX, moonY, 14, rl.Color{R: 80, G: 80, B: 120, A: 40})
	rl.DrawCircle(moonX, moonY, 10, rl.Color{R: 100, G: 100, B: 140, A: 50})
	// Moon
	rl.DrawCircle(moonX, moonY, 8, rl.Color{R: 240, G: 235, B: 220, A: 255})
	rl.DrawCircle(moonX-2, moonY-1, 7, rl.Color{R: 250, G: 248, B: 235, A: 255})
	// Craters
	rl.DrawCircle(moonX+2, moonY+2, 2, rl.Color{R: 220, G: 215, B: 200, A: 255})
	rl.DrawCircle(moonX-3, moonY+1, 1, rl.Color{R: 225, G: 220, B: 205, A: 255})

	// Twinkling stars
	starPositions := [][2]int32{{130, 35}, {145, 55}, {155, 30}, {138, 80}, {185, 60}, {170, 90}, {128, 100}, {190, 45}}
	for i, pos := range starPositions {
		twinkle := uint8(180 + 75*simpleSinF(float64(time)*2.0+float64(i)*0.8))
		rl.DrawPixel(pos[0], pos[1], rl.Color{R: twinkle, G: twinkle, B: 255, A: 255})
		if i%3 == 0 {
			rl.DrawPixel(pos[0]+1, pos[1], rl.Color{R: twinkle / 2, G: twinkle / 2, B: 200, A: 150})
		}
	}

	// Window dividers
	rl.DrawRectangle(158, 20, 4, 95, rl.Color{R: 60, G: 42, B: 38, A: 255})
	rl.DrawRectangle(120, 60, 80, 4, rl.Color{R: 60, G: 42, B: 38, A: 255})

	// Curtains with gentle sway
	curtainSway := int32(2 * simpleSinF(float64(time)*0.8))
	// Left curtain
	for y := int32(12); y < 120; y++ {
		wave := int32(simpleSinF(float64(y)*0.1+float64(time)*0.5) * 2)
		rl.DrawLine(105+wave+curtainSway, y, 118+wave+curtainSway, y, rl.Color{R: 100, G: 40, B: 50, A: 255})
	}
	// Right curtain
	for y := int32(12); y < 120; y++ {
		wave := int32(simpleSinF(float64(y)*0.1+float64(time)*0.5+1) * 2)
		rl.DrawLine(202-wave-curtainSway, y, 215-wave-curtainSway, y, rl.Color{R: 100, G: 40, B: 50, A: 255})
	}

	// === TALL BOOKSHELF LEFT ===
	// Frame
	rl.DrawRectangle(2, 25, 55, 135, rl.Color{R: 70, G: 48, B: 40, A: 255})
	rl.DrawRectangle(5, 28, 49, 129, rl.Color{R: 60, G: 42, B: 35, A: 255})

	// Shelf dividers
	for y := int32(28); y < 155; y += 32 {
		rl.DrawRectangle(5, y, 49, 3, rl.Color{R: 75, G: 52, B: 42, A: 255})
	}

	// Books with varied sizes and colors
	bookColors := []rl.Color{
		{R: 160, G: 60, B: 70, A: 255},   // Red
		{R: 70, G: 100, B: 160, A: 255},  // Blue
		{R: 70, G: 140, B: 90, A: 255},   // Green
		{R: 180, G: 160, B: 80, A: 255},  // Gold
		{R: 140, G: 80, B: 140, A: 255},  // Purple
		{R: 80, G: 70, B: 60, A: 255},    // Brown
		{R: 200, G: 100, B: 60, A: 255},  // Orange
	}
	for shelf := 0; shelf < 4; shelf++ {
		shelfY := int32(31 + shelf*32)
		bx := int32(7)
		for book := 0; book < 7; book++ {
			bw := int32(5 + (book*shelf)%3)
			bh := int32(22 + (book*3+shelf*2)%8)
			bc := bookColors[(book+shelf*2)%len(bookColors)]
			// Book body
			rl.DrawRectangle(bx, shelfY+28-bh, bw, bh, bc)
			// Spine detail
			rl.DrawLine(bx+bw/2, shelfY+30-bh, bx+bw/2, shelfY+26, rl.Color{R: bc.R - 30, G: bc.G - 30, B: bc.B - 30, A: 255})
			bx += bw + 1
			if bx > 50 {
				break
			}
		}
	}

	// Glowing orb on shelf
	orbY := int32(45)
	orbGlow := uint8(150 + 50*simpleSinF(float64(time)*1.5))
	rl.DrawCircle(35, orbY, 6, rl.Color{R: 100, G: orbGlow, B: 200, A: 80})
	rl.DrawCircle(35, orbY, 4, rl.Color{R: 150, G: orbGlow, B: 230, A: 150})
	rl.DrawCircle(35, orbY, 2, rl.Color{R: 200, G: 220, B: 255, A: 255})

	// === DESK WITH ITEMS - Right side ===
	// Desk body
	rl.DrawRectangle(235, 95, 80, 65, rl.Color{R: 75, G: 52, B: 42, A: 255})
	// Desk top
	rl.DrawRectangle(232, 90, 86, 8, rl.Color{R: 85, G: 58, B: 48, A: 255})
	// Desk legs
	rl.DrawRectangle(238, 155, 6, 5, rl.Color{R: 65, G: 45, B: 38, A: 255})
	rl.DrawRectangle(306, 155, 6, 5, rl.Color{R: 65, G: 45, B: 38, A: 255})

	// Drawer
	rl.DrawRectangle(255, 110, 40, 25, rl.Color{R: 65, G: 45, B: 38, A: 255})
	rl.DrawCircle(275, 122, 2, rl.Color{R: 180, G: 160, B: 100, A: 255})

	// Open spellbook
	rl.DrawRectangle(245, 82, 30, 8, rl.Color{R: 90, G: 60, B: 50, A: 255})
	rl.DrawRectangle(247, 80, 12, 8, rl.Color{R: 230, G: 220, B: 190, A: 255})
	rl.DrawRectangle(261, 80, 12, 8, rl.Color{R: 225, G: 215, B: 185, A: 255})
	// Text lines
	rl.DrawLine(249, 82, 257, 82, rl.Color{R: 60, G: 50, B: 40, A: 200})
	rl.DrawLine(249, 84, 256, 84, rl.Color{R: 60, G: 50, B: 40, A: 200})
	rl.DrawLine(263, 82, 270, 82, rl.Color{R: 60, G: 50, B: 40, A: 200})
	rl.DrawLine(263, 84, 269, 84, rl.Color{R: 60, G: 50, B: 40, A: 200})

	// Quill in inkwell
	rl.DrawRectangle(280, 82, 6, 8, rl.Color{R: 40, G: 35, B: 50, A: 255})
	rl.DrawLine(283, 82, 290, 70, rl.Color{R: 220, G: 200, B: 180, A: 255})
	rl.DrawLine(290, 70, 295, 65, rl.Color{R: 180, G: 100, B: 80, A: 255})

	// === CANDELABRA - Triple candle ===
	candleX := int32(300)
	// Base
	rl.DrawRectangle(candleX-8, 88, 16, 4, rl.Color{R: 180, G: 160, B: 100, A: 255})
	rl.DrawRectangle(candleX-2, 84, 4, 4, rl.Color{R: 170, G: 150, B: 90, A: 255})

	// Three candles
	candleOffsets := []int32{-6, 0, 6}
	for i, off := range candleOffsets {
		cx := candleX + off
		// Candle body
		rl.DrawRectangle(cx-2, 72, 4, 12, rl.Color{R: 235, G: 225, B: 200, A: 255})

		// Animated flame
		flicker := simpleSinF(float64(time)*8.0 + float64(i)*2.0)
		flickerX := int32(flicker * 1.5)
		flickerH := int32(6 + flicker*2)

		// Outer glow
		rl.DrawCircle(cx+flickerX, 68, 8, rl.Color{R: 255, G: 200, B: 100, A: 30})
		rl.DrawCircle(cx+flickerX, 68, 5, rl.Color{R: 255, G: 180, B: 80, A: 50})

		// Flame
		rl.DrawRectangle(cx-1+flickerX, 72-flickerH, 3, flickerH, rl.Color{R: 255, G: 180, B: 80, A: 255})
		rl.DrawRectangle(cx+flickerX, 72-flickerH+1, 2, flickerH-2, rl.Color{R: 255, G: 220, B: 150, A: 255})
		rl.DrawPixel(cx+flickerX, 72-flickerH+2, rl.Color{R: 255, G: 255, B: 220, A: 255})
	}

	// === POTION SHELF - Above desk ===
	rl.DrawRectangle(250, 40, 60, 6, rl.Color{R: 70, G: 48, B: 40, A: 255})
	// Shelf bracket
	rl.DrawTriangle(
		rl.Vector2{X: 252, Y: 46},
		rl.Vector2{X: 252, Y: 55},
		rl.Vector2{X: 260, Y: 46},
		rl.Color{R: 65, G: 45, B: 38, A: 255},
	)
	rl.DrawTriangle(
		rl.Vector2{X: 308, Y: 46},
		rl.Vector2{X: 308, Y: 55},
		rl.Vector2{X: 300, Y: 46},
		rl.Color{R: 65, G: 45, B: 38, A: 255},
	)

	// Potions with bubbling animation
	potionColors := []rl.Color{
		{R: 200, G: 80, B: 100, A: 255},  // Red health
		{R: 80, G: 150, B: 200, A: 255},  // Blue mana
		{R: 100, G: 200, B: 120, A: 255}, // Green poison
	}
	for i, pc := range potionColors {
		px := int32(260 + i*20)
		// Bottle
		rl.DrawRectangle(px, 28, 8, 12, rl.Color{R: 200, G: 200, B: 220, A: 100})
		// Liquid
		bubbleOff := int32(simpleSinF(float64(time)*3.0+float64(i)*1.5) * 2)
		rl.DrawRectangle(px+1, 32+bubbleOff, 6, 7-bubbleOff, pc)
		// Cork
		rl.DrawRectangle(px+2, 26, 4, 3, rl.Color{R: 140, G: 100, B: 70, A: 255})
		// Shine
		rl.DrawPixel(px+2, 30, rl.Color{R: 255, G: 255, B: 255, A: 150})
	}

	// === FLOATING DUST MOTES in candlelight ===
	for i := 0; i < 15; i++ {
		baseX := float64(220 + (i*41)%100)
		baseY := float64(50 + (i*37)%100)
		fx := baseX + 15*simpleSinF(float64(time)*0.3+float64(i)*0.7)
		fy := baseY + 10*simpleSinF(float64(time)*0.5+float64(i)*1.1)
		alpha := uint8(80 + 40*simpleSinF(float64(time)*0.8+float64(i)))
		rl.DrawPixel(int32(fx), int32(fy), rl.Color{R: 255, G: 240, B: 200, A: alpha})
	}

	// === SMALL DETAILS ===
	// Skull on bookshelf (spooky but cute)
	rl.DrawCircle(45, 140, 5, rl.Color{R: 230, G: 225, B: 215, A: 255})
	rl.DrawPixel(43, 139, rl.Color{R: 30, G: 25, B: 35, A: 255})
	rl.DrawPixel(47, 139, rl.Color{R: 30, G: 25, B: 35, A: 255})
	rl.DrawPixel(45, 142, rl.Color{R: 30, G: 25, B: 35, A: 200})

	// Crystal ball on desk corner
	crystalX, crystalY := int32(240), int32(82)
	rl.DrawCircle(crystalX, crystalY, 6, rl.Color{R: 80, G: 100, B: 140, A: 200})
	rl.DrawCircle(crystalX-1, crystalY-1, 4, rl.Color{R: 100, G: 120, B: 160, A: 180})
	// Swirling mist inside
	mistAngle := time * 2.0
	mx := int32(simpleSinF(float64(mistAngle)) * 2)
	my := int32(simpleSinF(float64(mistAngle)+1.5) * 2)
	rl.DrawPixel(crystalX+mx, crystalY+my, rl.Color{R: 180, G: 200, B: 255, A: 200})
	// Shine
	rl.DrawPixel(crystalX-2, crystalY-2, rl.Color{R: 255, G: 255, B: 255, A: 200})
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
		// Updated 16-frame thinking with sway and bob
		swayCurve := []int{0, 0, 1, 1, 1, 1, 0, 0, 0, 0, -1, -1, -1, -1, 0, 0}
		bobCurve := []int{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0}
		return float32(swayCurve[f%len(swayCurve)]), float32(-bobCurve[f%len(bobCurve)])

	case AnimWalk:
		// bob is added to body position in spritegen (bob=1 means body down)
		bobCurve := []int{0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0}
		return 0, float32(bobCurve[f%len(bobCurve)])

	case AnimVictoryPose:
		// Victory pose fist pump - matches spritegen exactly
		if f < 4 {
			// Anticipation - coil down (oy+frame)
			return 0, float32(f)
		} else if f < 8 {
			// Explosive rise
			riseY := []int{2, -2, -6, -8}
			return 0, float32(riseY[f-4])
		} else if f < 14 {
			// Peak pose with subtle bob
			bob := []int{0, 1, 0, -1, 0, 1}
			return 0, float32(-8 + bob[f-8])
		} else if f < 18 {
			// Settle down
			settleY := []int{-6, -4, -2, 0}
			return 0, float32(settleY[f-14])
		} else {
			// Final stance
			bounceY := []int{1, 0}
			idx := f - 18
			if idx >= len(bounceY) {
				idx = len(bounceY) - 1
			}
			return 0, float32(bounceY[idx])
		}
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
	// Layout constants (smaller text/spacing)
	rowH := int32(9)
	padding := int32(2)
	labelW := int32(22)
	arrowW := int32(5)
	valueW := int32(45)
	gap := int32(2)

	// Calculate full panel size (3 rows)
	panelW := padding + labelW + gap + arrowW + valueW + arrowW + padding

	// Collapsed bar height (same as mana bar: 10px)
	collapsedH := int32(10)

	// Full panel height = collapsed header + 3 rows
	fullPanelH := collapsedH + rowH + gap + rowH + gap + rowH + padding

	// Animate between collapsed and full panel
	// Use easeOutQuad for smooth animation
	t := r.pickerAnim
	eased := t * (2 - t) // easeOutQuad

	// Calculate animated height
	panelH := int32(float32(collapsedH) + eased*float32(fullPanelH-collapsedH))

	// Position in bottom-left corner (same as mana bar: screenHeight - height - 4)
	panelX := int32(4)
	panelY := screenHeight - panelH - 4

	// Colors
	panelBg := rl.Color{R: 20, G: 18, B: 30, A: 230}
	panelBorder := rl.Color{R: 60, G: 55, B: 80, A: 255}
	labelDim := rl.Color{R: 80, G: 75, B: 100, A: 255}
	labelActive := rl.Color{R: 180, G: 175, B: 200, A: 255}
	valueDim := rl.Color{R: 120, G: 115, B: 140, A: 255}
	valueActive := rl.Color{R: 255, G: 200, B: 80, A: 255}
	arrowDim := rl.Color{R: 80, G: 75, B: 100, A: 255}
	arrowActive := rl.Color{R: 180, G: 175, B: 200, A: 255}

	// Always draw panel background
	rl.DrawRectangle(panelX-1, panelY-1, panelW+2, panelH+2, panelBorder)
	rl.DrawRectangle(panelX, panelY, panelW, panelH, panelBg)

	// Draw "Tab ^" hint (changes to "Tab v" when expanded)
	hintAlpha := uint8(200)
	hintColor := rl.Color{R: 100, G: 95, B: 120, A: hintAlpha}
	if r.pickerExpanded {
		rl.DrawText("Tab v", panelX+padding, panelY+1, 8, hintColor)
	} else {
		rl.DrawText("Tab ^", panelX+padding, panelY+1, 8, hintColor)
	}

	// Don't draw panel contents if mostly collapsed
	if r.pickerAnim < 0.1 {
		return
	}

	// Fade in content alpha
	contentAlpha := uint8(eased * 255)
	labelDim.A = contentAlpha
	labelActive.A = contentAlpha
	valueDim.A = contentAlpha
	valueActive.A = contentAlpha
	arrowDim.A = contentAlpha
	arrowActive.A = contentAlpha

	// === ROW 1: HAT === (starts below the "Tab" hint)
	rowY := panelY + collapsedH
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
	rowY = panelY + collapsedH + rowH + gap
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

	// === ROW 3: MODE ===
	rowY = panelY + collapsedH + rowH + gap + rowH + gap
	x = panelX + padding

	// Determine colors based on active row
	modeLabelColor := labelDim
	modeArrowColor := arrowDim
	if r.activeRow == 2 {
		modeLabelColor = labelActive
		modeArrowColor = arrowActive
	}

	// Label
	rl.DrawText("MODE", x, rowY+2, 8, modeLabelColor)
	x += labelW + gap

	// Left arrow
	rl.DrawText("<", x, rowY+2, 8, modeArrowColor)
	x += arrowW

	// Value
	modeName := r.GetCurrentModeName()
	modeColor := valueDim
	if r.activeRow == 2 {
		modeColor = valueActive
	}
	modeTextW := rl.MeasureText(modeName, 8)
	modeTextX := x + (valueW-modeTextW)/2
	rl.DrawText(modeName, modeTextX, rowY+2, 8, modeColor)
	x += valueW

	// Right arrow
	rl.DrawText(">", x, rowY+2, 8, modeArrowColor)
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

// DrawGameUI renders the game UI elements (quest text, mana bar, etc.)
func (r *Renderer) DrawGameUI(state *GameState) {
	// Draw quest text at top
	r.drawQuestText(state)

	// Draw mana bar at bottom
	r.drawManaBar(state)

	// Draw thought bubble (above Claude)
	if state.ThoughtText != "" && state.ThoughtFade > 0 {
		r.drawThoughtBubble(state)
	}

	// Draw SHIPPED! rainbow banner (git push celebration)
	if state.ShippedActive {
		r.drawShippedBanner(state)
	}

	// Draw thrown tools
	r.drawThrownTools(state)

	// Draw flying enemies (before mini agents so they appear behind)
	r.drawFlyingEnemies(state)

	// Draw mini agents (subagent mini Claudes)
	r.drawMiniAgents(state)

	// Draw think hard effects
	if state.ThinkHardActive {
		r.drawThinkHardEffect(state)
	}

	// Draw compact/rest effects
	if state.CompactActive {
		r.drawCompactEffect(state)
	}
}

// drawThrownTools renders tool names flying through the air
func (r *Renderer) drawThrownTools(state *GameState) {
	for _, tool := range state.ThrownTools {
		// Fade out near end of life
		alpha := float32(1.0)
		if tool.Life > tool.MaxLife*0.7 {
			alpha = 1.0 - (tool.Life-tool.MaxLife*0.7)/(tool.MaxLife*0.3)
		}

		// Unpack color
		cr := uint8((tool.Color >> 24) & 0xFF)
		cg := uint8((tool.Color >> 16) & 0xFF)
		cb := uint8((tool.Color >> 8) & 0xFF)
		color := rl.Color{R: cr, G: cg, B: cb, A: uint8(alpha * 255)}
		shadowColor := rl.Color{R: 0, G: 0, B: 0, A: uint8(alpha * 180)}

		// Slight rotation based on velocity (tilted in direction of travel)
		// For now just draw with shadow for depth
		x := int32(tool.X)
		y := int32(tool.Y)
		fontSize := int32(8)

		// Shadow
		rl.DrawText(tool.Text, x+1, y+1, fontSize, shadowColor)
		// Main text
		rl.DrawText(tool.Text, x, y, fontSize, color)

		// Small trail particles
		if tool.Life < tool.MaxLife*0.5 {
			trailColor := color
			trailColor.A = uint8(alpha * 100)
			rl.DrawRectangle(x-3, y+3, 2, 2, trailColor)
			rl.DrawRectangle(x-6, y+4, 2, 2, trailColor)
		}
	}
}

// drawFlyingEnemies renders enemies flying toward Claude
func (r *Renderer) drawFlyingEnemies(state *GameState) {
	for _, enemy := range state.FlyingEnemies {
		// Show impact effect
		if enemy.Hit && enemy.Impact > 0 {
			r.drawImpactEffect(enemy.X, enemy.Y, enemy.Impact, enemy.Type)
			continue // Don't draw enemy sprite during impact
		}

		alpha := uint8(255)

		if r.hasEnemySprites {
			// Calculate source rectangle from enemy sprite sheet
			frameX := float32(enemy.Frame * enemyFrameWidth)
			frameY := float32(int(enemy.Type) * enemyFrameHeight)

			sourceRec := rl.Rectangle{
				X:      frameX,
				Y:      frameY,
				Width:  enemyFrameWidth,
				Height: enemyFrameHeight,
			}

			// Position - center the sprite on the enemy position
			destRec := rl.Rectangle{
				X:      enemy.X - float32(enemyFrameWidth/2),
				Y:      enemy.Y - float32(enemyFrameHeight/2),
				Width:  enemyFrameWidth,
				Height: enemyFrameHeight,
			}

			tint := rl.Color{R: 255, G: 255, B: 255, A: alpha}
			rl.DrawTexturePro(r.enemySpriteSheet, sourceRec, destRec, rl.Vector2{}, 0, tint)
		} else {
			// Fallback: draw colored rectangle
			var color rl.Color
			switch enemy.Type {
			case EnemyBug:
				color = rl.Color{R: 34, G: 139, B: 34, A: alpha} // Green
			case EnemyError:
				color = rl.Color{R: 255, G: 51, B: 51, A: alpha} // Red
			case EnemyLowContext:
				color = rl.Color{R: 255, G: 204, B: 0, A: alpha} // Yellow
			}
			rl.DrawRectangle(int32(enemy.X)-16, int32(enemy.Y)-8, 32, 16, color)
		}
	}
}

// drawImpactEffect renders the impact burst when an enemy hits Claude
func (r *Renderer) drawImpactEffect(x, y, timer float32, enemyType EnemyType) {
	// Impact expands outward
	progress := 1.0 - (timer / 0.3) // 0 to 1 as timer goes from 0.3 to 0
	size := 10 + progress*20        // Grows from 10 to 30

	// Color based on enemy type
	var color rl.Color
	switch enemyType {
	case EnemyBug:
		color = rl.Color{R: 100, G: 200, B: 100, A: uint8(200 * (1 - progress))} // Green
	case EnemyError:
		color = rl.Color{R: 255, G: 100, B: 100, A: uint8(200 * (1 - progress))} // Red
	case EnemyLowContext:
		color = rl.Color{R: 255, G: 220, B: 100, A: uint8(200 * (1 - progress))} // Yellow
	}

	// Draw expanding ring
	cx := int32(x)
	cy := int32(y)
	halfSize := int32(size / 2)

	// Outer ring
	rl.DrawRectangleLines(cx-halfSize, cy-halfSize, int32(size), int32(size), color)

	// Inner burst lines (star pattern)
	for i := 0; i < 8; i++ {
		angle := float32(i) * 3.14159 / 4
		dx := int32(simpleCosF(float64(angle)) * float64(size/2))
		dy := int32(simpleSinF(float64(angle)) * float64(size/2))
		rl.DrawLine(cx, cy, cx+dx, cy+dy, color)
	}

	// Center flash
	flashAlpha := uint8(255 * (1 - progress))
	flashColor := rl.Color{R: 255, G: 255, B: 255, A: flashAlpha}
	rl.DrawRectangle(cx-3, cy-3, 6, 6, flashColor)
}

// drawMiniAgents renders all active mini Claudes (subagents)
func (r *Renderer) drawMiniAgents(state *GameState) {
	if !r.hasMiniSprites {
		// Fallback: draw colored rectangles with names
		for _, agent := range state.MiniAgents {
			x := int32(agent.X) - 8
			y := int32(agent.Y) - 16

			// Simple colored box
			bodyColor := rl.Color{R: 218, G: 165, B: 140, A: 255} // Peach
			rl.DrawRectangle(x, y, 16, 16, bodyColor)

			// Draw name below
			nameColor := rl.Color{R: 200, G: 180, B: 160, A: 255}
			textWidth := rl.MeasureText(agent.Name, 6)
			nameX := x + 8 - textWidth/2
			rl.DrawText(agent.Name, nameX, y+18, 6, nameColor)
		}
		return
	}

	// Draw each mini agent
	for _, agent := range state.MiniAgents {
		// Calculate source rectangle from mini sprite sheet
		frameX := float32(agent.Frame * miniFrameWidth)
		frameY := float32(int(agent.Animation) * miniFrameHeight)

		sourceRec := rl.Rectangle{
			X:      frameX,
			Y:      frameY,
			Width:  miniFrameWidth,
			Height: miniFrameHeight,
		}

		// Position - center the sprite on the agent position
		destRec := rl.Rectangle{
			X:      agent.X - float32(miniFrameWidth/2),
			Y:      agent.Y - float32(miniFrameHeight),
			Width:  miniFrameWidth,
			Height: miniFrameHeight,
		}

		rl.DrawTexturePro(r.miniSpriteSheet, sourceRec, destRec, rl.Vector2{}, 0, rl.White)

		// Draw agent name below the sprite
		nameColor := rl.Color{R: 200, G: 180, B: 160, A: 255}
		shadowColor := rl.Color{R: 0, G: 0, B: 0, A: 150}
		textWidth := rl.MeasureText(agent.Name, 6)
		nameX := int32(agent.X) - textWidth/2
		nameY := int32(agent.Y) + 2

		// Shadow
		rl.DrawText(agent.Name, nameX+1, nameY+1, 6, shadowColor)
		// Name text
		rl.DrawText(agent.Name, nameX, nameY, 6, nameColor)
	}
}

// drawQuestText renders the current quest/user prompt
func (r *Renderer) drawQuestText(state *GameState) {
	if state.QuestText == "" || state.QuestFade <= 0 {
		return
	}

	// Colors with alpha based on fade
	alpha := uint8(state.QuestFade * 255)
	panelBg := rl.Color{R: 15, G: 12, B: 25, A: uint8(float32(alpha) * 0.85)}
	borderColor := rl.Color{R: 80, G: 65, B: 110, A: alpha}
	textColor := rl.Color{R: 220, G: 210, B: 190, A: alpha}
	labelColor := rl.Color{R: 160, G: 120, B: 60, A: alpha}

	// Panel dimensions - full width, multi-line support
	padding := int32(3)
	panelX := int32(5)
	panelY := int32(3)
	panelWidth := int32(screenWidth - 10)

	// Word wrap the text
	maxLineWidth := panelWidth - padding*2 - 2
	lines := wordWrap(state.QuestText, 6, maxLineWidth)
	if len(lines) > 3 {
		lines = lines[:3] // Max 3 lines
		lines[2] = lines[2] + "..."
	}

	lineHeight := int32(8)
	panelHeight := int32(len(lines))*lineHeight + padding*2

	// Draw panel background
	rl.DrawRectangle(panelX-1, panelY-1, panelWidth+2, panelHeight+2, borderColor)
	rl.DrawRectangle(panelX, panelY, panelWidth, panelHeight, panelBg)

	// Draw each line
	for i, line := range lines {
		y := panelY + padding + int32(i)*lineHeight
		if i == 0 {
			// First line with label
			rl.DrawText(">", panelX+padding, y, 6, labelColor)
			rl.DrawText(line, panelX+padding+8, y, 6, textColor)
		} else {
			rl.DrawText(line, panelX+padding+8, y, 6, textColor)
		}
	}
}

// drawThoughtBubble renders Claude's current thought in a cloud-like bubble
func (r *Renderer) drawThoughtBubble(state *GameState) {
	if state.ThoughtText == "" || state.ThoughtFade <= 0 {
		return
	}

	alpha := uint8(state.ThoughtFade * 200) // Slightly transparent

	// Truncate thought text if too long (show first 150 chars)
	thoughtText := state.ThoughtText
	if len(thoughtText) > 150 {
		thoughtText = thoughtText[:147] + "..."
	}

	// Remove newlines for cleaner display
	thoughtText = strings.ReplaceAll(thoughtText, "\n", " ")
	thoughtText = strings.ReplaceAll(thoughtText, "  ", " ")

	// Colors
	bubbleBg := rl.Color{R: 250, G: 248, B: 245, A: alpha}
	bubbleBorder := rl.Color{R: 180, G: 175, B: 165, A: alpha}
	textColor := rl.Color{R: 60, G: 55, B: 50, A: alpha}
	shadowColor := rl.Color{R: 0, G: 0, B: 0, A: uint8(float32(alpha) * 0.3)}

	// Bubble dimensions - positioned above Claude
	padding := int32(4)
	fontSize := int32(5)
	maxBubbleWidth := int32(180)

	// Word wrap the text
	lines := wordWrap(thoughtText, fontSize, maxBubbleWidth-padding*2)
	if len(lines) > 4 {
		lines = lines[:4]
		lines[3] = lines[3][:min(len(lines[3]), 20)] + "..."
	}

	lineHeight := int32(8) // More space between lines
	bubbleHeight := int32(len(lines))*lineHeight + padding*2 + 2 // Extra vertical padding

	// Calculate text width for bubble sizing
	maxTextWidth := int32(0)
	for _, line := range lines {
		w := int32(rl.MeasureText(line, fontSize))
		if w > maxTextWidth {
			maxTextWidth = w
		}
	}
	bubbleWidth := maxTextWidth + padding*2 + 4
	if bubbleWidth > maxBubbleWidth {
		bubbleWidth = maxBubbleWidth
	}
	if bubbleWidth < 40 {
		bubbleWidth = 40
	}

	// Position: above and to the right of Claude's head
	claudeX := float32(screenWidth / 2)
	bubbleX := int32(claudeX) - bubbleWidth/2 + 20
	bubbleY := int32(50) // Above Claude

	// Keep bubble on screen
	if bubbleX < 5 {
		bubbleX = 5
	}
	if bubbleX+bubbleWidth > screenWidth-5 {
		bubbleX = screenWidth - 5 - bubbleWidth
	}

	// Draw shadow
	rl.DrawRectangleRounded(
		rl.Rectangle{X: float32(bubbleX + 2), Y: float32(bubbleY + 2), Width: float32(bubbleWidth), Height: float32(bubbleHeight)},
		0.3, 8, shadowColor,
	)

	// Draw main bubble with rounded corners
	rl.DrawRectangleRounded(
		rl.Rectangle{X: float32(bubbleX), Y: float32(bubbleY), Width: float32(bubbleWidth), Height: float32(bubbleHeight)},
		0.3, 8, bubbleBg,
	)
	rl.DrawRectangleRoundedLines(
		rl.Rectangle{X: float32(bubbleX), Y: float32(bubbleY), Width: float32(bubbleWidth), Height: float32(bubbleHeight)},
		0.3, 8, bubbleBorder,
	)

	// Draw thought bubble "tail" - small circles leading toward Claude's head
	// Claude position: top-left (128, 106), size 64x64, head at ~Y=120
	claudeHeadX := int32(screenWidth / 2) // 160
	claudeHeadY := int32(120)             // Top of head area

	// Start tail from bottom of bubble, slightly left of center
	tailStartX := bubbleX + bubbleWidth/2
	tailStartY := bubbleY + bubbleHeight

	// Three circles in a curved path toward Claude's head
	// Calculate direction vector
	dx := float32(claudeHeadX - tailStartX)
	dy := float32(claudeHeadY - tailStartY)

	// Draw circles along the path, getting smaller as they approach Claude
	rl.DrawCircle(tailStartX+int32(dx*0.2), tailStartY+int32(dy*0.25), 4, bubbleBg)
	rl.DrawCircleLines(tailStartX+int32(dx*0.2), tailStartY+int32(dy*0.25), 4, bubbleBorder)
	rl.DrawCircle(tailStartX+int32(dx*0.45), tailStartY+int32(dy*0.5), 3, bubbleBg)
	rl.DrawCircleLines(tailStartX+int32(dx*0.45), tailStartY+int32(dy*0.5), 3, bubbleBorder)
	rl.DrawCircle(tailStartX+int32(dx*0.7), tailStartY+int32(dy*0.75), 2, bubbleBg)
	rl.DrawCircleLines(tailStartX+int32(dx*0.7), tailStartY+int32(dy*0.75), 2, bubbleBorder)

	// Draw text with better vertical centering
	for i, line := range lines {
		y := bubbleY + padding + 2 + int32(i)*lineHeight
		rl.DrawText(line, bubbleX+padding+2, y, fontSize, textColor)
	}
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

// wordWrap splits text into lines that fit within maxWidth
func wordWrap(text string, fontSize int32, maxWidth int32) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if rl.MeasureText(testLine, fontSize) <= maxWidth {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// drawManaBar renders the context window usage as a mana bar
func (r *Renderer) drawManaBar(state *GameState) {
	// Position to the right of the picker (aligned at same height)
	barHeight := int32(10)
	barX := int32(120) // Leave room for MANA label at barX-30 = 90
	barY := int32(screenHeight - barHeight - 4)
	barWidth := int32(screenWidth - barX - 5)

	// Background
	bgColor := rl.Color{R: 20, G: 18, B: 30, A: 230}
	borderColor := rl.Color{R: 60, G: 55, B: 80, A: 255}
	rl.DrawRectangle(barX-1, barY-1, barWidth+2, barHeight+2, borderColor)
	rl.DrawRectangle(barX, barY, barWidth, barHeight, bgColor)

	// Calculate fill (mana drains as tokens are used)
	usedRatio := state.ManaDisplay / float32(state.ManaMax)
	if usedRatio > 1 {
		usedRatio = 1
	}
	remainingRatio := 1.0 - usedRatio
	fillWidth := int32(float32(barWidth-2) * remainingRatio)

	// Color based on remaining mana
	var fillColor rl.Color
	if remainingRatio > 0.5 {
		// Blue - plenty left
		fillColor = rl.Color{R: 80, G: 120, B: 200, A: 255}
	} else if remainingRatio > 0.25 {
		// Yellow - caution
		fillColor = rl.Color{R: 200, G: 180, B: 80, A: 255}
	} else if remainingRatio > 0.1 {
		// Orange - warning
		fillColor = rl.Color{R: 220, G: 140, B: 60, A: 255}
	} else {
		// Red - danger, almost empty
		fillColor = rl.Color{R: 200, G: 80, B: 80, A: 255}
	}

	// Draw fill
	if fillWidth > 0 {
		rl.DrawRectangle(barX+1, barY+1, fillWidth, barHeight-2, fillColor)
	}

	// Draw label
	labelColor := rl.Color{R: 120, G: 115, B: 140, A: 255}
	rl.DrawText("MANA", barX-30, barY, 8, labelColor)

	// Draw remaining token count
	remaining := state.ManaMax - int(state.ManaDisplay)
	if remaining < 0 {
		remaining = 0
	}
	tokenText := fmt.Sprintf("%dk", remaining/1000)
	textWidth := rl.MeasureText(tokenText, 8)
	rl.DrawText(tokenText, barX+barWidth-textWidth-2, barY+1, 8, rl.Color{R: 160, G: 155, B: 180, A: 255})
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

// hsvToRGB converts HSV to RGB color
func hsvToRGB(h int, s, v float64) rl.Color {
	h = h % 360
	c := v * s
	x := c * (1 - abs(float64(h%120)/60.0-1))
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

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func simpleSinF(rad float64) float64 {
	// Use standard math would be better but keeping simple
	// Approximate with lookup
	deg := int(rad * 180 / 3.14159)
	sins := []float64{0, 0.5, 0.87, 1, 0.87, 0.5, 0, -0.5, -0.87, -1, -0.87, -0.5}
	idx := ((deg % 360) + 360) % 360
	return sins[(idx/30)%12]
}

func simpleCosF(rad float64) float64 {
	return simpleSinF(rad + 1.5708) // +90 degrees
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

// Unload frees all loaded textures
func (r *Renderer) Unload() {
	if r.background.ID != 0 {
		rl.UnloadTexture(r.background)
	}
	if r.spriteSheet.ID != 0 {
		rl.UnloadTexture(r.spriteSheet)
	}
}
