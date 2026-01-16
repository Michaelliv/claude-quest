package main

import (
	"encoding/json"
	"fmt"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)


// loadHats loads all hat textures from assets/accessories/hats
func (r *Renderer) loadHats() {
	hatsDir := getAssetPath("accessories/hats")
	entries, err := os.ReadDir(hatsDir)
	if err != nil {
		fmt.Printf("Could not read hats directory: %v\n", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		if len(filename) < 5 || filename[len(filename)-4:] != ".png" {
			continue
		}
		name := filename[:len(filename)-4]
		path := fmt.Sprintf("%s/%s", hatsDir, filename)
		tex := rl.LoadTexture(path)
		r.hats = append(r.hats, tex)
		r.hatNames = append(r.hatNames, name)
		fmt.Printf("Loaded hat: %s\n", name)
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
	facesDir := getAssetPath("accessories/faces")
	entries, err := os.ReadDir(facesDir)
	if err != nil {
		fmt.Printf("Could not read faces directory: %v\n", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		if len(filename) < 5 || filename[len(filename)-4:] != ".png" {
			continue
		}
		name := filename[:len(filename)-4]
		path := fmt.Sprintf("%s/%s", facesDir, filename)
		tex := rl.LoadTexture(path)
		r.faces = append(r.faces, tex)
		r.faceNames = append(r.faceNames, name)
		fmt.Printf("Loaded face: %s\n", name)
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

// DrawAccessoryPicker draws the collapsible accessory picker UI (legacy, replaced by modal)
func (r *Renderer) DrawAccessoryPicker() {
	// Replaced by modal picker - kept for backwards compatibility
	r.DrawAccessoryPickerHint()
}

// DrawAccessoryPickerHint draws a small hint to open the accessory picker
func (r *Renderer) DrawAccessoryPickerHint() {
	// Don't show hint if modal is open
	if r.pickerModal {
		return
	}

	// Small hint in bottom-left corner
	panelX := int32(4)
	panelY := int32(screenHeight - 14)
	panelW := int32(50)
	panelH := int32(10)

	// Background
	panelBg := rl.Color{R: 20, G: 18, B: 30, A: 200}
	panelBorder := rl.Color{R: 60, G: 55, B: 80, A: 200}

	rl.DrawRectangle(panelX-1, panelY-1, panelW+2, panelH+2, panelBorder)
	rl.DrawRectangle(panelX, panelY, panelW, panelH, panelBg)

	// Hint text
	hintColor := rl.Color{R: 120, G: 115, B: 150, A: 255}
	rl.DrawText("Tab", panelX+4, panelY+1, 8, hintColor)

	// Show current equipped count
	ownedCount := 0
	if r.currentHat >= 0 {
		ownedCount++
	}
	if r.currentFace >= 0 {
		ownedCount++
	}
	if r.currentAura >= 0 {
		ownedCount++
	}
	if r.currentTrail >= 0 {
		ownedCount++
	}

	if ownedCount > 0 {
		countColor := rl.Color{R: 255, G: 200, B: 80, A: 255}
		countText := fmt.Sprintf("%d", ownedCount)
		rl.DrawText(countText, panelX+30, panelY+1, 8, countColor)
	}
}
