package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Strip picker constants - calculated for proper centering
const (
	stripHeight  = 32                              // Height of the picker strip
	stripY       = screenHeight - stripHeight - 2  // Y position (166)
	stripX       = 4                               // X position
	stripW       = screenWidth - 8                 // Width (312)
	slotWidth    = 72                              // Width per slot (4*72=288)
	slotCount    = 4                               // Number of slots
	slotIconSize = 12                              // Size of item icons
)

// Calculated at init - center the slots horizontally
var slotStartX = stripX + (stripW-slotCount*slotWidth)/2 // = 4 + (312-288)/2 = 4 + 12 = 16

// Aura/Trail color swatches for picker display
var auraColors = map[string]rl.Color{
	"aura_pixel":    {R: 255, G: 255, B: 200, A: 255},
	"aura_flame":    {R: 255, G: 120, B: 40, A: 255},
	"aura_frost":    {R: 150, G: 200, B: 255, A: 255},
	"aura_electric": {R: 255, G: 240, B: 80, A: 255},  // Bright yellow lightning
	"aura_shadow":   {R: 80, G: 50, B: 120, A: 255},
	"aura_heart":    {R: 255, G: 100, B: 150, A: 255},
	"aura_code":     {R: 0, G: 255, B: 100, A: 255},
	"aura_rainbow":  {R: 255, G: 100, B: 100, A: 255},
}

var trailColors = map[string]rl.Color{
	"trail_sparkle": {R: 255, G: 255, B: 200, A: 255},
	"trail_flame":   {R: 255, G: 150, B: 50, A: 255},
	"trail_frost":   {R: 150, G: 200, B: 255, A: 255},
	"trail_hearts":  {R: 255, G: 100, B: 150, A: 255},
	"trail_pixel":   {R: 100, G: 255, B: 100, A: 255},
	"trail_rainbow": {R: 255, G: 200, B: 100, A: 255},
}

// ToggleModalPicker opens or closes the picker strip
func (r *Renderer) ToggleModalPicker() {
	r.pickerModal = !r.pickerModal
	if r.pickerModal {
		r.pickerPreviewHat = r.currentHat
		r.pickerPreviewFace = r.currentFace
		r.pickerPreviewAura = r.currentAura
		r.pickerPreviewTrail = r.currentTrail
		r.syncItemIndices()
	} else {
		r.currentHat = r.pickerPreviewHat
		r.currentFace = r.pickerPreviewFace
		r.currentAura = r.pickerPreviewAura
		r.currentTrail = r.pickerPreviewTrail
		r.SavePrefs()
	}
}

// IsModalPickerOpen returns whether the picker strip is open
func (r *Renderer) IsModalPickerOpen() bool {
	return r.pickerModal
}

// syncItemIndices sets the picker item indices to match current equipped items
func (r *Renderer) syncItemIndices() {
	slotTypes := []ItemSlot{SlotHat, SlotFace, SlotAura, SlotTrail}
	currentIdxs := []int{r.currentHat, r.currentFace, r.currentAura, r.currentTrail}
	nameArrays := [][]string{r.hatNames, r.faceNames, r.auraNames, r.trailNames}

	for slot := 0; slot < 4; slot++ {
		currentIdx := currentIdxs[slot]
		slotType := slotTypes[slot]
		names := nameArrays[slot]

		currentID := ""
		if currentIdx >= 0 && currentIdx < len(names) {
			currentID = names[currentIdx]
		}

		pos := 0
		if currentID == "" {
			r.pickerItemIndex[slot] = 0
			continue
		}

		idx := 1 // Start after "none"
		for _, item := range ItemRegistry {
			if item.Slot == slotType {
				isOwned := r.profile != nil && r.profile.IsOwned(item.ID)
				if isOwned {
					if item.ID == currentID {
						pos = idx
						break
					}
					idx++
				}
			}
		}
		r.pickerItemIndex[slot] = pos
	}
}

// ModalPickerNavigate handles navigation in the picker strip
func (r *Renderer) ModalPickerNavigate(dx, dy int) {
	if !r.pickerModal {
		return
	}

	if dx != 0 {
		r.pickerSlot += dx
		if r.pickerSlot < 0 {
			r.pickerSlot = 3
		} else if r.pickerSlot > 3 {
			r.pickerSlot = 0
		}
	}

	if dy != 0 {
		r.cycleSlotItem(r.pickerSlot, dy)
	}
}

// cycleSlotItem cycles the item in a slot
func (r *Renderer) cycleSlotItem(slot, direction int) {
	slotTypes := []ItemSlot{SlotHat, SlotFace, SlotAura, SlotTrail}
	slotType := slotTypes[slot]

	type itemEntry struct {
		idx   int
		id    string
		owned bool
		level int
	}
	var items []itemEntry

	items = append(items, itemEntry{idx: -1, id: "", owned: true, level: 0})

	var names []string
	switch slot {
	case 0:
		names = r.hatNames
	case 1:
		names = r.faceNames
	case 2:
		names = r.auraNames
	case 3:
		names = r.trailNames
	}

	for _, item := range ItemRegistry {
		if item.Slot == slotType {
			isOwned := r.profile != nil && r.profile.IsOwned(item.ID)
			if isOwned {
				idx := -1
				for i, name := range names {
					if name == item.ID {
						idx = i
						break
					}
				}
				if idx >= 0 {
					items = append(items, itemEntry{idx: idx, id: item.ID, owned: true, level: item.MinLevel})
				}
			}
		}
	}

	for _, item := range ItemRegistry {
		if item.Slot == slotType {
			isOwned := r.profile != nil && r.profile.IsOwned(item.ID)
			if !isOwned {
				idx := -1
				for i, name := range names {
					if name == item.ID {
						idx = i
						break
					}
				}
				items = append(items, itemEntry{idx: idx, id: item.ID, owned: false, level: item.MinLevel})
			}
		}
	}

	// Use pickerItemIndex to track position (not preview index, which doesn't update for locked items)
	currentPos := r.pickerItemIndex[slot]
	if currentPos < 0 || currentPos >= len(items) {
		currentPos = 0
	}

	newPos := currentPos + direction
	if newPos < 0 {
		newPos = len(items) - 1
	} else if newPos >= len(items) {
		newPos = 0
	}

	newItem := items[newPos]
	if newItem.owned {
		r.setPreviewIndex(slot, newItem.idx)
	}
	r.pickerItemIndex[slot] = newPos
}

func (r *Renderer) getPreviewIndex(slot int) int {
	switch slot {
	case 0:
		return r.pickerPreviewHat
	case 1:
		return r.pickerPreviewFace
	case 2:
		return r.pickerPreviewAura
	case 3:
		return r.pickerPreviewTrail
	}
	return -1
}

func (r *Renderer) setPreviewIndex(slot, idx int) {
	switch slot {
	case 0:
		r.pickerPreviewHat = idx
	case 1:
		r.pickerPreviewFace = idx
	case 2:
		r.pickerPreviewAura = idx
	case 3:
		r.pickerPreviewTrail = idx
	}
}

// UpdateModalPickerAnim updates the picker strip animation
func (r *Renderer) UpdateModalPickerAnim(dt float32) {
	speed := float32(10.0)
	if r.pickerModal {
		r.pickerModalAnim += dt * speed
		if r.pickerModalAnim > 1.0 {
			r.pickerModalAnim = 1.0
		}
	} else {
		r.pickerModalAnim -= dt * speed
		if r.pickerModalAnim < 0.0 {
			r.pickerModalAnim = 0.0
		}
	}
}

// DrawModalPicker renders the bottom strip picker
func (r *Renderer) DrawModalPicker() {
	if r.pickerModalAnim <= 0 {
		return
	}

	t := r.pickerModalAnim
	eased := t * (2 - t)

	// Strip slides up from bottom
	yOffset := (1 - eased) * float32(stripHeight+10)
	baseY := float32(stripY) + yOffset

	// Strip background
	bgColor := rl.Color{R: 20, G: 18, B: 35, A: uint8(eased * 220)}
	borderColor := rl.Color{R: 60, G: 55, B: 85, A: uint8(eased * 255)}

	rl.DrawRectangle(stripX-1, int32(baseY)-1, stripW+2, stripHeight+2, borderColor)
	rl.DrawRectangle(stripX, int32(baseY), stripW, stripHeight, bgColor)

	// Draw four slots centered
	slotNames := []string{"HAT", "FACE", "AURA", "TRAIL"}
	for slot := 0; slot < slotCount; slot++ {
		slotX := int32(slotStartX) + int32(slot*slotWidth)
		r.drawPickerSlot(slotX, int32(baseY), slot, slotNames[slot], eased)
	}
}

// drawPickerSlot draws a single slot in the picker strip
func (r *Renderer) drawPickerSlot(x, y int32, slot int, label string, alpha float32) {
	isActive := r.pickerSlot == slot
	a := uint8(alpha * 255)

	// Cell dimensions: 72-4=68 wide, 32-4=28 tall
	cellW := int32(slotWidth - 4)  // 68
	cellH := int32(stripHeight - 4) // 28
	cellX := x + 2
	cellY := y + 2

	if isActive {
		highlightColor := rl.Color{R: 255, G: 200, B: 80, A: uint8(alpha * 50)}
		rl.DrawRectangle(cellX, cellY, cellW, cellH, highlightColor)
		borderColor := rl.Color{R: 255, G: 200, B: 80, A: a}
		rl.DrawRectangleLines(cellX, cellY, cellW, cellH, borderColor)
	}

	// Vertical layout in 28px cell:
	// - Label (6px) at y+5
	// - Gap (2px)
	// - Icon (12px) at y+13
	// Total: 5 + 6 + 2 + 12 + 3 = 28 (3px bottom margin)

	// Label - centered horizontally
	labelColor := rl.Color{R: 90, G: 85, B: 115, A: a}
	if isActive {
		labelColor = rl.Color{R: 180, G: 170, B: 200, A: a}
	}
	labelW := rl.MeasureText(label, 6)
	labelX := cellX + (cellW-labelW)/2
	labelY := cellY + 4
	rl.DrawText(label, labelX, labelY, 6, labelColor)

	// Get item info
	itemInfo := r.getSlotItemInfo(slot)

	// Icon area starts at cellY + 12, height 12
	iconAreaY := cellY + 12

	if itemInfo.id == "" {
		// "None" - centered dash
		dashW := int32(16)
		dashH := int32(2)
		dashX := cellX + (cellW-dashW)/2
		dashY := iconAreaY + (slotIconSize-dashH)/2
		dashColor := rl.Color{R: 70, G: 65, B: 90, A: a}
		rl.DrawRectangle(dashX, dashY, dashW, dashH, dashColor)
	} else if !itemInfo.owned {
		// Locked - lock icon + "Lv##" centered together
		// Lock: 6x8, text: ~20px for "Lv##", gap: 2px
		// Total width ≈ 6 + 2 + 20 = 28
		lockW := int32(6)
		lockH := int32(8)
		lvlText := fmt.Sprintf("Lv%d", itemInfo.level)
		textW := rl.MeasureText(lvlText, 6)
		gap := int32(3)
		totalW := lockW + gap + textW

		startX := cellX + (cellW-totalW)/2
		lockY := iconAreaY + (slotIconSize-lockH)/2

		// Draw lock
		lockColor := rl.Color{R: 90, G: 75, B: 65, A: a}
		// Lock shackle (top part)
		rl.DrawRectangle(startX+1, lockY, 4, 3, lockColor)
		// Lock body
		rl.DrawRectangle(startX, lockY+2, lockW, 5, lockColor)

		// Draw level text
		lvlColor := rl.Color{R: 150, G: 120, B: 70, A: a}
		textY := iconAreaY + (slotIconSize-6)/2
		rl.DrawText(lvlText, startX+lockW+gap, textY, 6, lvlColor)
	} else {
		// Owned item - draw icon centered
		iconX := cellX + (cellW-slotIconSize)/2
		iconY := iconAreaY
		r.drawSlotItemIcon(iconX, iconY, slot, itemInfo.id, a)
	}
}

// slotItemInfo holds info about the currently displayed item in a slot
type slotItemInfo struct {
	id    string
	owned bool
	level int
}

// getSlotItemInfo gets info about what's currently shown in a slot
func (r *Renderer) getSlotItemInfo(slot int) slotItemInfo {
	slotTypes := []ItemSlot{SlotHat, SlotFace, SlotAura, SlotTrail}
	slotType := slotTypes[slot]

	var names []string
	switch slot {
	case 0:
		names = r.hatNames
	case 1:
		names = r.faceNames
	case 2:
		names = r.auraNames
	case 3:
		names = r.trailNames
	}

	var items []slotItemInfo
	items = append(items, slotItemInfo{id: "", owned: true, level: 0})

	for _, item := range ItemRegistry {
		if item.Slot == slotType {
			isOwned := r.profile != nil && r.profile.IsOwned(item.ID)
			if isOwned {
				found := false
				for _, name := range names {
					if name == item.ID {
						found = true
						break
					}
				}
				if found {
					items = append(items, slotItemInfo{id: item.ID, owned: true, level: item.MinLevel})
				}
			}
		}
	}

	for _, item := range ItemRegistry {
		if item.Slot == slotType {
			isOwned := r.profile != nil && r.profile.IsOwned(item.ID)
			if !isOwned {
				items = append(items, slotItemInfo{id: item.ID, owned: false, level: item.MinLevel})
			}
		}
	}

	pos := r.pickerItemIndex[slot]
	if pos < 0 || pos >= len(items) {
		pos = 0
	}
	return items[pos]
}

// drawSlotItemIcon draws the icon for an item
func (r *Renderer) drawSlotItemIcon(x, y int32, slot int, itemID string, alpha uint8) {
	switch slot {
	case 0: // HAT
		for i, name := range r.hatNames {
			if name == itemID && i < len(r.hats) {
				hat := r.hats[i]
				srcRec := rl.Rectangle{X: 0, Y: 0, Width: float32(hat.Width), Height: float32(hat.Height)}
				dstRec := rl.Rectangle{X: float32(x), Y: float32(y), Width: slotIconSize, Height: slotIconSize}
				rl.DrawTexturePro(hat, srcRec, dstRec, rl.Vector2{}, 0, rl.Color{R: 255, G: 255, B: 255, A: alpha})
				return
			}
		}
	case 1: // FACE
		for i, name := range r.faceNames {
			if name == itemID && i < len(r.faces) {
				face := r.faces[i]
				srcRec := rl.Rectangle{X: 0, Y: 0, Width: float32(face.Width), Height: float32(face.Height)}
				dstRec := rl.Rectangle{X: float32(x), Y: float32(y), Width: slotIconSize, Height: slotIconSize}
				rl.DrawTexturePro(face, srcRec, dstRec, rl.Vector2{}, 0, rl.Color{R: 255, G: 255, B: 255, A: alpha})
				return
			}
		}
	case 2: // AURA - color swatch with glow effect
		if color, ok := auraColors[itemID]; ok {
			// Outer glow
			glowColor := color
			glowColor.A = alpha / 3
			rl.DrawRectangle(x-1, y-1, slotIconSize+2, slotIconSize+2, glowColor)
			// Main swatch
			color.A = alpha
			rl.DrawRectangle(x, y, slotIconSize, slotIconSize, color)
			// Inner highlight
			innerColor := rl.Color{R: 255, G: 255, B: 255, A: alpha / 3}
			rl.DrawRectangle(x+2, y+2, slotIconSize-4, slotIconSize-4, innerColor)
		}
	case 3: // TRAIL - dust particles flowing left (opposite to Claude's walk direction)
		if color, ok := trailColors[itemID]; ok {
			cy := y + slotIconSize/2
			// Particles flow from right to left (Claude walks right, dust trails behind)
			// Rightmost = newest/brightest, leftmost = oldest/faintest
			color.A = alpha / 4
			rl.DrawCircle(x+1, cy-1, 1, color) // Faint, dispersed
			color.A = alpha / 3
			rl.DrawCircle(x+3, cy+1, 1, color) // Fading
			color.A = alpha / 2
			rl.DrawCircle(x+5, cy-1, 1, color) // Mid
			color.A = alpha * 2 / 3
			rl.DrawCircle(x+7, cy, 2, color) // Brighter
			color.A = alpha
			rl.DrawCircle(x+10, cy, 2, color) // Brightest (just left Claude)
		}
	}
}

// GetPreviewHat returns the hat to display (preview if picker open, current otherwise)
func (r *Renderer) GetPreviewHat() int {
	if r.pickerModal {
		return r.pickerPreviewHat
	}
	return r.currentHat
}

// GetPreviewFace returns the face to display
func (r *Renderer) GetPreviewFace() int {
	if r.pickerModal {
		return r.pickerPreviewFace
	}
	return r.currentFace
}

// GetPreviewAura returns the aura to display
func (r *Renderer) GetPreviewAura() int {
	if r.pickerModal {
		return r.pickerPreviewAura
	}
	return r.currentAura
}

// GetPreviewTrail returns the trail to display
func (r *Renderer) GetPreviewTrail() int {
	if r.pickerModal {
		return r.pickerPreviewTrail
	}
	return r.currentTrail
}
