//go:build debug

package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Studio mode - clean asset development environment
func runStudio() {
	fmt.Println("Claude Quest Studio")
	fmt.Println("Asset development environment with hot reload")

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(screenWidth*windowScale, screenHeight*windowScale, "Claude Quest Studio")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	target := rl.LoadRenderTexture(screenWidth, screenHeight)
	defer rl.UnloadRenderTexture(target)

	config := LoadConfig("config.json")
	renderer := NewRenderer(config)
	animations := NewAnimationSystem()

	// Hot reloader
	hotReloader, _ := NewHotReloader(renderer)
	if hotReloader != nil {
		hotReloader.Start()
		defer hotReloader.Stop()
	}

	// Studio state
	paused := false
	speed := float32(1.0)
	currentAnim := 0
	showHelp := true
	biomeNames := []string{"Forest", "Mountain", "Midnight", "Kingdom", "Library"}

	// Picker state: 0=none, 1=animation, 2=biome, 3=hat, 4=face, 5=aura, 6=trail
	pickerMode := 0
	pickerIndex := 0

	// Start with Idle, enable loop mode for studio
	animations.SetAnimation(AnimIdle)
	animations.SetLoopMode(true)

	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()

		// Update animations (unless paused)
		if !paused {
			animations.Update(dt * speed)
		}

		renderer.UpdateScrollOnly(dt)
		renderer.UpdatePickerAnim(dt)

		// Hot reload
		if hotReloader != nil {
			hotReloader.ProcessReloads()
		}

		// === CONTROLS ===

		// Picker mode controls
		if pickerMode > 0 {
			// Get current list length
			listLen := 0
			switch pickerMode {
			case 1:
				listLen = len(animations.GetAllAnimations())
			case 2:
				listLen = len(biomeNames)
			case 3:
				listLen = len(renderer.hatNames) + 1 // +1 for "None"
			case 4:
				listLen = len(renderer.faceNames) + 1
			case 5:
				listLen = len(renderer.auraNames) + 1
			case 6:
				listLen = len(renderer.trailNames) + 1
			}

			// Up/Down: navigate
			if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressed(rl.KeyW) {
				pickerIndex--
				if pickerIndex < 0 {
					pickerIndex = listLen - 1
				}
			}
			if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyS) {
				pickerIndex++
				if pickerIndex >= listLen {
					pickerIndex = 0
				}
			}

			// Enter: select and close
			if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
				switch pickerMode {
				case 1: // Animation
					allAnims := animations.GetAllAnimations()
					currentAnim = pickerIndex
					animations.SetAnimation(allAnims[pickerIndex])
				case 2: // Biome
					renderer.SetBiome(pickerIndex)
				case 3: // Hat
					renderer.currentHat = pickerIndex - 1 // -1 = none
					renderer.SavePrefs()
				case 4: // Face
					renderer.currentFace = pickerIndex - 1
					renderer.SavePrefs()
				case 5: // Aura
					renderer.currentAura = pickerIndex - 1
					renderer.SavePrefs()
				case 6: // Trail
					renderer.currentTrail = pickerIndex - 1
					renderer.SavePrefs()
				}
				pickerMode = 0
			}

			// Escape: close without selecting
			if rl.IsKeyPressed(rl.KeyEscape) {
				pickerMode = 0
			}

			// Tab: cycle to next cosmetic picker
			if rl.IsKeyPressed(rl.KeyTab) && pickerMode >= 3 {
				pickerMode++
				if pickerMode > 6 {
					pickerMode = 3
				}
				pickerIndex = 0
			}
		} else {
			// Normal controls (when no picker open)

			// Space: pause/play
			if rl.IsKeyPressed(rl.KeySpace) {
				paused = !paused
			}

			// Arrow keys or ,/. : step frames (when paused)
			if paused {
				if rl.IsKeyPressed(rl.KeyComma) || rl.IsKeyPressed(rl.KeyLeft) {
					animations.StepFrame(-1)
				}
				if rl.IsKeyPressed(rl.KeyPeriod) || rl.IsKeyPressed(rl.KeyRight) {
					animations.StepFrame(1)
				}
			}

			// -/= : speed
			if rl.IsKeyPressed(rl.KeyMinus) {
				speed /= 2
				if speed < 0.125 {
					speed = 0.125
				}
			}
			if rl.IsKeyPressed(rl.KeyEqual) {
				speed *= 2
				if speed > 4 {
					speed = 4
				}
			}

			// A: open animation picker
			if rl.IsKeyPressed(rl.KeyA) {
				pickerMode = 1
				pickerIndex = currentAnim
			}

			// B: open biome picker
			if rl.IsKeyPressed(rl.KeyB) {
				pickerMode = 2
				pickerIndex = renderer.currentBiome
			}

			// Tab or C: open cosmetics picker (starts with hats)
			if rl.IsKeyPressed(rl.KeyTab) || rl.IsKeyPressed(rl.KeyC) {
				pickerMode = 3
				pickerIndex = renderer.currentHat + 1 // +1 because 0 is "None"
			}

			// R: reload textures
			if rl.IsKeyPressed(rl.KeyR) && hotReloader != nil {
				hotReloader.ForceReloadAll()
			}

			// G: regenerate sprites
			if rl.IsKeyPressed(rl.KeyG) && hotReloader != nil {
				hotReloader.QueueRegenerate()
			}

			// H: toggle help
			if rl.IsKeyPressed(rl.KeyH) {
				showHelp = !showHelp
			}
		}

		// === RENDER ===
		rl.BeginTextureMode(target)
		rl.ClearBackground(rl.Color{R: 24, G: 20, B: 37, A: 255})

		// Draw scene (no game UI, no old picker)
		renderer.Draw(animations.GetState())

		// Draw studio UI
		drawStudioUI(animations, renderer, paused, speed, biomeNames, showHelp, pickerMode, pickerIndex)

		rl.EndTextureMode()

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		sourceRec := rl.Rectangle{X: 0, Y: float32(screenHeight), Width: float32(screenWidth), Height: -float32(screenHeight)}
		destRec := getScaledDestRect()
		rl.DrawTexturePro(target.Texture, sourceRec, destRec, rl.Vector2{}, 0, rl.White)
		rl.EndDrawing()
	}
}

func drawStudioUI(anim *AnimationSystem, r *Renderer, paused bool, speed float32, biomeNames []string, showHelp bool, pickerMode, pickerIndex int) {
	state := anim.GetState()
	frameCount := anim.GetAnimationLength()

	// Colors
	dim := rl.Color{R: 100, G: 95, B: 120, A: 200}
	bright := rl.Color{R: 255, G: 200, B: 80, A: 255}
	keyBg := rl.Color{R: 50, G: 45, B: 70, A: 255}
	keyFg := rl.Color{R: 255, G: 255, B: 255, A: 255}
	paused_c := rl.Color{R: 255, G: 120, B: 120, A: 255}
	playing := rl.Color{R: 120, G: 255, B: 120, A: 255}
	selected := rl.Color{R: 255, G: 153, B: 51, A: 255}

	// === TOP BAR: Animation + Biome ===
	rl.DrawRectangle(0, 0, screenWidth, 14, rl.Color{R: 15, G: 12, B: 25, A: 220})

	// Left: Animation info
	animName := state.CurrentAnim.String()
	frameText := fmt.Sprintf("%d/%d", state.Frame+1, frameCount)
	rl.DrawText(animName, 4, 3, 8, bright)
	rl.DrawText(frameText, 4+int32(rl.MeasureText(animName, 8))+8, 3, 8, dim)

	// Progress bar
	barX := int32(90)
	barW := int32(40)
	progress := float32(state.Frame+1) / float32(frameCount)
	rl.DrawRectangle(barX, 5, barW, 4, rl.Color{R: 40, G: 35, B: 60, A: 255})
	rl.DrawRectangle(barX, 5, int32(float32(barW)*progress), 4, selected)

	// Center: Status
	var statusText string
	var statusColor rl.Color
	if paused {
		statusText = "PAUSED"
		statusColor = paused_c
	} else {
		statusText = fmt.Sprintf("%.2fx", speed)
		statusColor = playing
	}
	statusW := rl.MeasureText(statusText, 8)
	rl.DrawText(statusText, (screenWidth-statusW)/2, 3, 8, statusColor)

	// Right: Biome
	biomeName := biomeNames[r.currentBiome]
	biomeW := rl.MeasureText(biomeName, 8)
	rl.DrawText(biomeName, screenWidth-biomeW-4, 3, 8, bright)

	// === PICKER (if open) ===
	if pickerMode > 0 {
		var items []string
		var title string
		var hint string

		switch pickerMode {
		case 1:
			title = "Animation"
			items = []string{"Idle", "Enter", "Casting", "Attack", "Writing", "Victory", "Hurt", "Thinking", "Walk", "VictoryPose"}
			hint = "Up/Down  Enter  Esc"
		case 2:
			title = "Biome"
			items = biomeNames
			hint = "Up/Down  Enter  Esc"
		case 3:
			title = "Hat (Tab:next)"
			items = append([]string{"None"}, r.hatNames...)
			hint = "Up/Down  Enter  Tab:next  Esc"
		case 4:
			title = "Face (Tab:next)"
			items = append([]string{"None"}, r.faceNames...)
			hint = "Up/Down  Enter  Tab:next  Esc"
		case 5:
			title = "Aura (Tab:next)"
			items = append([]string{"None"}, r.auraNames...)
			hint = "Up/Down  Enter  Tab:next  Esc"
		case 6:
			title = "Trail (Tab:next)"
			items = append([]string{"None"}, r.trailNames...)
			hint = "Up/Down  Enter  Tab:next  Esc"
		}

		// Picker panel (centered, with scroll if needed)
		panelW := int32(110)
		maxVisible := 12
		visibleItems := len(items)
		if visibleItems > maxVisible {
			visibleItems = maxVisible
		}
		panelH := int32(visibleItems*10 + 20)
		panelX := (screenWidth - panelW) / 2
		panelY := (screenHeight - panelH) / 2

		// Calculate scroll offset
		scrollOffset := 0
		if len(items) > maxVisible {
			if pickerIndex > maxVisible/2 {
				scrollOffset = pickerIndex - maxVisible/2
			}
			if scrollOffset > len(items)-maxVisible {
				scrollOffset = len(items) - maxVisible
			}
		}

		// Background
		rl.DrawRectangle(panelX-2, panelY-2, panelW+4, panelH+4, rl.Color{R: 80, G: 75, B: 100, A: 255})
		rl.DrawRectangle(panelX, panelY, panelW, panelH, rl.Color{R: 25, G: 22, B: 35, A: 250})

		// Title
		rl.DrawText(title, panelX+4, panelY+4, 8, bright)
		rl.DrawLine(panelX+4, panelY+14, panelX+panelW-4, panelY+14, dim)

		// Items (with scroll)
		for i := 0; i < visibleItems && scrollOffset+i < len(items); i++ {
			itemIdx := scrollOffset + i
			item := items[itemIdx]
			y := panelY + 18 + int32(i)*10

			if itemIdx == pickerIndex {
				rl.DrawRectangle(panelX+2, y-1, panelW-4, 10, rl.Color{R: 60, G: 55, B: 80, A: 255})
				rl.DrawText(item, panelX+6, y, 8, selected)
			} else {
				rl.DrawText(item, panelX+6, y, 8, dim)
			}
		}

		// Scroll indicators
		if scrollOffset > 0 {
			rl.DrawText("...", panelX+panelW-16, panelY+18, 8, dim)
		}
		if scrollOffset+maxVisible < len(items) {
			rl.DrawText("...", panelX+panelW-16, panelY+panelH-12, 8, dim)
		}

		// Hint at bottom
		hintW := rl.MeasureText(hint, 8)
		rl.DrawText(hint, (screenWidth-hintW)/2, panelY+panelH+4, 8, dim)
	}

	// === BOTTOM: Help (if enabled and no picker) ===
	if showHelp && pickerMode == 0 {
		rl.DrawRectangle(0, screenHeight-22, screenWidth, 22, rl.Color{R: 15, G: 12, B: 25, A: 220})

		// Row 1: Playback + Pickers
		y1 := int32(screenHeight - 20)
		x := int32(4)
		x = drawKey(x, y1, "SPC", "play", keyBg, keyFg, dim)
		x = drawKey(x, y1, "< >", "frame", keyBg, keyFg, dim)
		x = drawKey(x, y1, "- +", "speed", keyBg, keyFg, dim)

		// Row 2: Pickers + reload
		y2 := int32(screenHeight - 10)
		x = int32(4)
		x = drawKey(x, y2, "A", "anim", keyBg, keyFg, dim)
		x = drawKey(x, y2, "B", "biome", keyBg, keyFg, dim)
		x = drawKey(x, y2, "C", "cosmetic", keyBg, keyFg, dim)
		x = drawKey(x, y2, "R", "reload", keyBg, keyFg, dim)
		x = drawKey(x, y2, "G", "gen", keyBg, keyFg, dim)
	} else if pickerMode == 0 {
		rl.DrawText("H help", screenWidth-36, screenHeight-10, 8, dim)
	}
}

// drawKey draws a key with background box and label, returns next x position
func drawKey(x, y int32, keyStr, label string, keyBg, keyFg, labelColor rl.Color) int32 {
	keyW := rl.MeasureText(keyStr, 8)
	pad := int32(2)

	// Key background
	rl.DrawRectangle(x, y, keyW+pad*2, 9, keyBg)
	// Key text
	rl.DrawText(keyStr, x+pad, y+1, 8, keyFg)
	// Label
	rl.DrawText(label, x+keyW+pad*2+3, y+1, 8, labelColor)
	labelW := rl.MeasureText(label, 8)

	return x + keyW + pad*2 + labelW + 8
}
