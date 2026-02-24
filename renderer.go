package main

import (
	"fmt"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// AccessoryPrefs stores user's accessory preferences
type AccessoryPrefs struct {
	HatName   string `json:"hat"`
	FaceName  string `json:"face"`
	AuraName  string `json:"aura"`
	TrailName string `json:"trail"`
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

	// Unicode font for Korean text (--korean flag)
	gameFont    rl.Font
	hasGameFont bool

	// Treasure chest
	chestTexture    rl.Texture2D
	hasChestTexture bool

	// Profile for ownership checks
	profile *CareerProfile

	// Accessories
	hats        []rl.Texture2D
	hatNames    []string
	currentHat  int // -1 = no hat
	faces       []rl.Texture2D
	faceNames   []string
	currentFace int // -1 = no face accessory

	// Auras and trails (particle-based, no textures needed)
	auraNames    []string
	currentAura  int // -1 = no aura
	trailNames   []string
	currentTrail int // -1 = no trail

	// Trail particles (separate from main particles)
	trailParticles []Particle
	lastClaudeX    float32 // For spawning trail behind movement

	// UI state
	activeRow int // 0 = HAT, 1 = FACE, 2 = AURA, 3 = TRAIL

	// Parallax scrolling
	scrollOffset float32

	// Biome system
	currentBiome int
	biomeTimer   float32

	// Picker visibility
	pickerExpanded bool
	pickerAnim     float32 // 0.0 = collapsed, 1.0 = expanded

	// Modal picker state
	pickerModal       bool    // True when modal picker is open
	pickerModalAnim   float32 // 0.0 = closed, 1.0 = fully open
	pickerSlot        int     // Currently selected slot (0=HAT, 1=FACE, 2=AURA, 3=TRAIL)
	pickerItemIndex   [4]int  // Selected item index per slot (-1 = none)
	pickerScrollPos   [4]int  // Scroll position per slot (for when items overflow)
	pickerPreviewHat  int     // Preview hat while browsing (-1 = use current)
	pickerPreviewFace int     // Preview face while browsing
	pickerPreviewAura int     // Preview aura while browsing
	pickerPreviewTrail int    // Preview trail while browsing
}

// SetProfile sets the career profile for ownership checks
func (r *Renderer) SetProfile(profile *CareerProfile) {
	r.profile = profile
	// Validate currently equipped items - reset if not owned
	if profile != nil {
		if r.currentHat >= 0 && r.currentHat < len(r.hatNames) {
			if !profile.IsOwned(r.hatNames[r.currentHat]) {
				r.currentHat = -1 // Reset to no hat
			}
		}
		if r.currentFace >= 0 && r.currentFace < len(r.faceNames) {
			if !profile.IsOwned(r.faceNames[r.currentFace]) {
				r.currentFace = -1 // Reset to no face
			}
		}
		if r.currentAura >= 0 && r.currentAura < len(r.auraNames) {
			if !profile.IsOwned(r.auraNames[r.currentAura]) {
				r.currentAura = -1 // Reset to no aura
			}
		}
		if r.currentTrail >= 0 && r.currentTrail < len(r.trailNames) {
			if !profile.IsOwned(r.trailNames[r.currentTrail]) {
				r.currentTrail = -1 // Reset to no trail
			}
		}
	}
}

// CycleBiome changes the current biome by delta (use 1 for next, -1 for previous)
func (r *Renderer) CycleBiome(delta int) {
	r.currentBiome = (r.currentBiome + delta + 5) % 5
	r.biomeTimer = 0 // Reset biome timer
}

// SetBiome sets the biome directly (0-4)
func (r *Renderer) SetBiome(index int) {
	if index >= 0 && index < 5 {
		r.currentBiome = index
		r.biomeTimer = 0
	}
}

// NewRenderer creates a new renderer with loaded assets
func NewRenderer(config *Config) *Renderer {
	r := &Renderer{
		config:         config,
		particles:      make([]Particle, 0, 100),
		trailParticles: make([]Particle, 0, 50),
		currentHat:     -1, // No hat by default
		currentFace:    -1, // No face accessory by default
		currentAura:    -1, // No aura by default
		currentTrail:   -1, // No trail by default
		// Initialize aura and trail names (particle-based, no textures)
		auraNames:  []string{"aura_pixel", "aura_flame", "aura_frost", "aura_electric", "aura_shadow", "aura_heart", "aura_code", "aura_rainbow"},
		trailNames: []string{"trail_sparkle", "trail_flame", "trail_frost", "trail_hearts", "trail_pixel", "trail_rainbow"},
		// Modal picker defaults
		pickerPreviewHat:   -2, // -2 means "use current" (distinct from -1 = none selected)
		pickerPreviewFace:  -2,
		pickerPreviewAura:  -2,
		pickerPreviewTrail: -2,
		pickerItemIndex:    [4]int{-1, -1, -1, -1},
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

	// Try to load chest sprite
	chestPath := getAssetPath("ui/chest.png")
	if _, err := os.Stat(chestPath); err == nil {
		r.chestTexture = rl.LoadTexture(chestPath)
		r.hasChestTexture = true
		fmt.Println("Loaded chest sprite from:", chestPath)
	}

	// Load Korean font if --korean flag is set
	if config.Korean {
		fontPath := getAssetPath("fonts/neodgm.ttf")
		if _, err := os.Stat(fontPath); err == nil {
			codepoints := make([]rune, 0, 95+11172)
			for c := rune(0x20); c <= 0x7E; c++ {
				codepoints = append(codepoints, c)
			}
			for c := rune(0xAC00); c <= 0xD7A3; c++ {
				codepoints = append(codepoints, c)
			}
			r.gameFont = rl.LoadFontEx(fontPath, 16, codepoints)
			rl.SetTextureFilter(r.gameFont.Texture, rl.FilterPoint)
			r.hasGameFont = true
			fmt.Println("Loaded Korean font from:", fontPath)
		} else {
			fmt.Println("Warning: --korean flag set but font not found at:", fontPath)
		}
	}

	// Load all accessories
	r.loadHats()
	r.loadFaces()

	// Load user preferences
	r.LoadPrefs()

	return r
}

// loadHats loads all hat textures from assets/accessories/hats

// Draw renders the current animation state
func (r *Renderer) Draw(state *AnimationState) {
	// Draw background
	r.drawBackground()

	// Update and draw trail particles (behind Claude)
	r.updateTrailParticles()
	r.drawTrailParticles()
	r.spawnTrailParticles(state)

	// Update and draw particles
	r.updateParticles()
	r.drawParticles()

	// Draw aura behind Claude
	r.drawAura(state)

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
	// Always use parallax background (Quest mode)
	r.drawParallaxBackground()
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
	case 4:
		r.drawBiomeWizardLibrary()
	}
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
	if r.hasGameFont {
		rl.UnloadFont(r.gameFont)
	}
}
