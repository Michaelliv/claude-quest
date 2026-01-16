package main

import (
	"encoding/json"
	"fmt"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)


// loadHats loads all hat textures from assets/accessories/hats
func (r *Renderer) loadHats() {
	hatFiles := []string{
		"wizard", "party", "headphones", "beret", "tophat",
		"catears", "crown", "propeller", "pirate", "viking",
		"chef", "halo", "jester", "cowboy", "fedora",
	}

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

// CycleHat cycles to the next owned hat (or no hat)
func (r *Renderer) CycleHat(direction int) {
	if len(r.hats) == 0 {
		return
	}
	// Try up to len+1 times to find an owned hat (or -1 for none)
	for i := 0; i <= len(r.hats); i++ {
		r.currentHat += direction
		if r.currentHat >= len(r.hats) {
			r.currentHat = -1
		} else if r.currentHat < -1 {
			r.currentHat = len(r.hats) - 1
		}
		// -1 (no hat) is always valid
		if r.currentHat == -1 {
			return
		}
		// Check if this hat is owned
		if r.profile != nil && r.currentHat >= 0 && r.currentHat < len(r.hatNames) {
			if r.profile.IsOwned(r.hatNames[r.currentHat]) {
				return
			}
		} else if r.profile == nil {
			// No profile = all items available (backwards compat)
			return
		}
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
	faceFiles := []string{
		"mustache", "dealwithit", "monocle", "pipe", "borat",
		"eyepatch", "glasses3d", "groucho", "bandana", "wizardbeard",
	}

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

// CycleFace cycles to the next owned face accessory (or none)
func (r *Renderer) CycleFace(direction int) {
	if len(r.faces) == 0 {
		return
	}
	// Try up to len+1 times to find an owned face (or -1 for none)
	for i := 0; i <= len(r.faces); i++ {
		r.currentFace += direction
		if r.currentFace >= len(r.faces) {
			r.currentFace = -1
		} else if r.currentFace < -1 {
			r.currentFace = len(r.faces) - 1
		}
		// -1 (no face) is always valid
		if r.currentFace == -1 {
			return
		}
		// Check if this face is owned
		if r.profile != nil && r.currentFace >= 0 && r.currentFace < len(r.faceNames) {
			if r.profile.IsOwned(r.faceNames[r.currentFace]) {
				return
			}
		} else if r.profile == nil {
			// No profile = all items available (backwards compat)
			return
		}
	}
}

// GetCurrentFaceName returns the name of the current face accessory
func (r *Renderer) GetCurrentFaceName() string {
	if r.currentFace < 0 || r.currentFace >= len(r.faceNames) {
		return ""
	}
	return r.faceNames[r.currentFace]
}

// SwitchRow switches between HAT (0), FACE (1), AURA (2), TRAIL (3) rows
func (r *Renderer) SwitchRow(direction int) {
	r.activeRow += direction
	if r.activeRow < 0 {
		r.activeRow = 3
	} else if r.activeRow > 3 {
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
		r.CycleAura(direction)
	case 3:
		r.CycleTrail(direction)
	}
	r.SavePrefs()
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


// SavePrefs saves current accessory choices to disk
func (r *Renderer) SavePrefs() {
	prefs := AccessoryPrefs{
		HatName:   r.GetCurrentHatName(),
		FaceName:  r.GetCurrentFaceName(),
		AuraName:  r.GetCurrentAuraName(),
		TrailName: r.GetCurrentTrailName(),
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
	// Find and set aura by name
	for i, name := range r.auraNames {
		if name == prefs.AuraName {
			r.currentAura = i
			break
		}
	}
	// Find and set trail by name
	for i, name := range r.trailNames {
		if name == prefs.TrailName {
			r.currentTrail = i
			break
		}
	}
}

// UpdateScroll advances the parallax scroll (call each frame)
func (r *Renderer) UpdateScroll(dt float32) {
	r.scrollOffset += dt * 25 // Walk speed
	// Wrap around to prevent overflow
	if r.scrollOffset > 10000 {
		r.scrollOffset -= 10000
	}

	// Biome transitions every ~20 seconds of walking
	r.biomeTimer += dt
	if r.biomeTimer > 20 {
		r.biomeTimer = 0
		r.currentBiome = (r.currentBiome + 1) % 5
	}
}

// UpdateScrollOnly advances scroll without biome cycling (for studio mode)
func (r *Renderer) UpdateScrollOnly(dt float32) {
	r.scrollOffset += dt * 25
	if r.scrollOffset > 10000 {
		r.scrollOffset -= 10000
	}
}

// DrawAccessoryPicker draws the collapsible accessory picker UI
func (r *Renderer) DrawAccessoryPicker() {
	// Layout constants (smaller text/spacing)
	rowH := int32(9)
	padding := int32(2)
	labelW := int32(22)
	arrowW := int32(5)
	valueW := int32(55) // Wider to fit aura names
	gap := int32(2)

	// Calculate full panel size (4 rows: HAT, FACE, AURA, TRAIL)
	panelW := padding + labelW + gap + arrowW + valueW + arrowW + padding

	// Collapsed bar height (same as mana bar: 10px)
	collapsedH := int32(10)

	// Full panel height = collapsed header + 4 rows
	fullPanelH := collapsedH + (rowH+gap)*4 + padding

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

	// Helper to draw a row
	drawRow := func(rowIdx int, label, value string) {
		rowY := panelY + collapsedH + int32(rowIdx)*(rowH+gap)
		x := panelX + padding

		// Determine colors based on active row
		lblColor := labelDim
		arrColor := arrowDim
		valColor := valueDim
		if r.activeRow == rowIdx {
			lblColor = labelActive
			arrColor = arrowActive
			if value != "-" {
				valColor = valueActive
			}
		}

		// Label
		rl.DrawText(label, x, rowY+2, 8, lblColor)
		x += labelW + gap

		// Left arrow
		rl.DrawText("<", x, rowY+2, 8, arrColor)
		x += arrowW

		// Value (centered)
		textW := rl.MeasureText(value, 8)
		textX := x + (valueW-textW)/2
		rl.DrawText(value, textX, rowY+2, 8, valColor)
		x += valueW

		// Right arrow
		rl.DrawText(">", x, rowY+2, 8, arrColor)
	}

	// === ROW 0: HAT ===
	hatName := r.GetCurrentHatName()
	if hatName == "" {
		hatName = "-"
	}
	drawRow(0, "HAT", hatName)

	// === ROW 1: FACE ===
	faceName := r.GetCurrentFaceName()
	if faceName == "" {
		faceName = "-"
	}
	drawRow(1, "FACE", faceName)

	// === ROW 2: AURA ===
	auraName := r.GetCurrentAuraName()
	if auraName == "" {
		auraName = "-"
	} else {
		// Strip "aura_" prefix for display
		if len(auraName) > 5 {
			auraName = auraName[5:]
		}
	}
	drawRow(2, "AURA", auraName)

	// === ROW 3: TRAIL ===
	trailName := r.GetCurrentTrailName()
	if trailName == "" {
		trailName = "-"
	} else {
		// Strip "trail_" prefix for display
		if len(trailName) > 6 {
			trailName = trailName[6:]
		}
	}
	drawRow(3, "TRAL", trailName)
}
