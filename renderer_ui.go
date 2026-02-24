package main

import (
	"fmt"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// DrawGameUI renders the game UI elements (quest text, mana bar, etc.)
func (r *Renderer) DrawGameUI(state *GameState) {
	// Draw level display first (so other UI can render on top)
	r.drawLevelDisplay(state)

	// Draw mana bar at bottom
	r.drawManaBar(state)

	// Draw XP bar (above mana bar)
	r.drawXPBar(state)

	// Draw flow meter (right side)
	r.drawFlowMeter(state)

	// Draw quest text at top (on top of level indicator)
	r.drawQuestText(state)

	// Draw thought bubble (above Claude)
	if state.ThoughtText != "" && state.ThoughtFade > 0 {
		r.drawThoughtBubble(state)
	}

	// Draw floating XP indicators (above thought bubble)
	r.drawFloatingXPs(state)

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
	questFontSize := int32(6)
	if r.hasGameFont {
		questFontSize = 16 // neodgm native pixel size
	}
	lines := r.wordWrapText(state.QuestText, questFontSize, maxLineWidth)
	if len(lines) > 3 {
		lines = lines[:3] // Max 3 lines
		lines[2] = lines[2] + "..."
	}

	lineHeight := int32(questFontSize + 2)
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
			r.drawText(line, panelX+padding+8, y, questFontSize, textColor)
		} else {
			r.drawText(line, panelX+padding+8, y, questFontSize, textColor)
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
	if r.hasGameFont {
		fontSize = 16 // neodgm native pixel size
	}
	maxBubbleWidth := int32(180)

	// Word wrap the text
	lines := r.wordWrapText(thoughtText, fontSize, maxBubbleWidth-padding*2)
	if len(lines) > 4 {
		lines = lines[:4]
		lines[3] = lines[3][:min(len(lines[3]), 20)] + "..."
	}

	lineHeight := int32(fontSize + 3)
	bubbleHeight := int32(len(lines))*lineHeight + padding*2 + 2 // Extra vertical padding

	// Calculate text width for bubble sizing
	maxTextWidth := int32(0)
	for _, line := range lines {
		w := r.measureText(line, fontSize)
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
		r.drawText(line, bubbleX+padding+2, y, fontSize, textColor)
	}
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

// drawXPBar renders the XP progress bar (above mana bar)
func (r *Renderer) drawXPBar(state *GameState) {
	if state.Profile == nil {
		return
	}

	// Position above the mana bar
	barHeight := int32(6)
	barX := int32(120)
	barY := int32(screenHeight - 10 - 4 - barHeight - 2) // Above mana bar
	barWidth := int32(screenWidth - barX - 5)

	// Background
	bgColor := rl.Color{R: 20, G: 18, B: 30, A: 200}
	borderColor := rl.Color{R: 50, G: 45, B: 70, A: 255}
	rl.DrawRectangle(barX-1, barY-1, barWidth+2, barHeight+2, borderColor)
	rl.DrawRectangle(barX, barY, barWidth, barHeight, bgColor)

	// Calculate fill based on XP progress to next level
	progress := state.Profile.XPProgress()
	fillWidth := int32(float32(barWidth-2) * progress)

	// Gold/yellow color for XP
	fillColor := rl.Color{R: 255, G: 200, B: 80, A: 255}

	// Draw fill
	if fillWidth > 0 {
		rl.DrawRectangle(barX+1, barY+1, fillWidth, barHeight-2, fillColor)
	}

	// Draw XP label
	labelColor := rl.Color{R: 120, G: 115, B: 140, A: 255}
	rl.DrawText("XP", barX-16, barY-1, 6, labelColor)
}

// drawLevelDisplay renders the current level
func (r *Renderer) drawLevelDisplay(state *GameState) {
	if state.Profile == nil {
		return
	}

	// Position in top left corner
	levelText := fmt.Sprintf("Lv.%d", state.Profile.Level)

	// Colors
	levelColor := rl.Color{R: 255, G: 200, B: 80, A: 255} // Gold
	shadowColor := rl.Color{R: 0, G: 0, B: 0, A: 150}

	x := int32(4)
	y := int32(4)

	// Shadow
	rl.DrawText(levelText, x+1, y+1, 8, shadowColor)
	// Main text
	rl.DrawText(levelText, x, y, 8, levelColor)
}

// drawFlowMeter renders the flow meter (vertical bar on right side)
func (r *Renderer) drawFlowMeter(state *GameState) {
	// Position on right side
	barWidth := int32(8)
	barHeight := int32(60)
	barX := int32(screenWidth - barWidth - 4)
	barY := int32(screenHeight/2 - barHeight/2)

	// Background
	bgColor := rl.Color{R: 20, G: 18, B: 30, A: 200}
	borderColor := rl.Color{R: 50, G: 45, B: 70, A: 255}
	rl.DrawRectangle(barX-1, barY-1, barWidth+2, barHeight+2, borderColor)
	rl.DrawRectangle(barX, barY, barWidth, barHeight, bgColor)

	// Calculate fill (bottom to top)
	fillHeight := int32(float32(barHeight-2) * state.Session.FlowMeter)

	// Color transitions: blue -> purple -> gold as it fills
	var fillColor rl.Color
	if state.Session.FlowMeter < 0.5 {
		// Blue to purple
		t := state.Session.FlowMeter * 2
		fillColor = rl.Color{
			R: uint8(80 + t*60),
			G: uint8(120 - t*40),
			B: uint8(200 + t*55),
			A: 255,
		}
	} else {
		// Purple to gold
		t := (state.Session.FlowMeter - 0.5) * 2
		fillColor = rl.Color{
			R: uint8(140 + t*115),
			G: uint8(80 + t*120),
			B: uint8(255 - t*175),
			A: 255,
		}
	}

	// Draw fill from bottom
	if fillHeight > 0 {
		fillY := barY + barHeight - 1 - fillHeight
		rl.DrawRectangle(barX+1, fillY, barWidth-2, fillHeight, fillColor)
	}

	// Label
	labelColor := rl.Color{R: 100, G: 95, B: 120, A: 255}
	rl.DrawText("F", barX+1, barY-10, 8, labelColor)

	// Peak indicator
	if state.Session.FlowPeakReached {
		peakColor := rl.Color{R: 255, G: 200, B: 80, A: 200}
		rl.DrawText("*", barX+barWidth+1, barY-2, 8, peakColor)
	}
}

// measureText measures text width using gameFont if available, otherwise default font
func (r *Renderer) measureText(text string, fontSize int32) int32 {
	if r.hasGameFont {
		return int32(rl.MeasureTextEx(r.gameFont, text, float32(fontSize), 0).X)
	}
	return rl.MeasureText(text, fontSize)
}

// drawText draws text using gameFont if available, otherwise default font
func (r *Renderer) drawText(text string, x, y, fontSize int32, color rl.Color) {
	if r.hasGameFont {
		rl.DrawTextEx(r.gameFont, text, rl.Vector2{X: float32(x), Y: float32(y)}, float32(fontSize), 0, color)
	} else {
		rl.DrawText(text, x, y, fontSize, color)
	}
}

// wordWrapText splits text into lines that fit within maxWidth
func (r *Renderer) wordWrapText(text string, fontSize int32, maxWidth int32) []string {
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

		if r.measureText(testLine, fontSize) <= maxWidth {
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

// drawFloatingXPs renders floating "+XP" indicators
func (r *Renderer) drawFloatingXPs(state *GameState) {
	for _, xp := range state.FloatingXPs {
		// Calculate alpha (fade out over time)
		progress := xp.Timer / xp.MaxLife
		alpha := uint8(255 * (1 - progress))

		// Gold color for XP
		xpColor := rl.Color{R: 255, G: 200, B: 80, A: alpha}
		shadowColor := rl.Color{R: 0, G: 0, B: 0, A: alpha / 2}

		// Format text
		text := fmt.Sprintf("+%dXP", xp.Amount)

		x := int32(xp.X)
		y := int32(xp.Y)

		// Shadow
		rl.DrawText(text, x+1, y+1, 8, shadowColor)
		// Main text
		rl.DrawText(text, x, y, 8, xpColor)
	}
}

// DrawTreasureChest renders the treasure chest ceremony overlay
func (r *Renderer) DrawTreasureChest(state *GameState) {
	chest := state.ActiveChest
	if chest == nil {
		return
	}

	// Semi-transparent overlay
	overlayAlpha := uint8(180)
	if chest.State == ChestClosed {
		// Fade in overlay
		overlayAlpha = uint8(chest.Timer / 0.5 * 180)
		if overlayAlpha > 180 {
			overlayAlpha = 180
		}
	}
	rl.DrawRectangle(0, 0, screenWidth, screenHeight, rl.Color{R: 10, G: 8, B: 20, A: overlayAlpha})

	// Chest position (center of screen) - chest is 64px wide when scaled 2x
	chestX := int32(screenWidth/2 - 32)
	chestY := int32(screenHeight/2 + 10) // Lower to make room for items above

	// Apply wobble offset
	wobbleOffset := int32(chest.GetWobbleOffset())

	// Draw chest sprite
	if r.hasChestTexture {
		// Determine which frame to show
		frameIdx := 0 // Closed
		switch chest.State {
		case ChestClosed:
			frameIdx = 0
		case ChestWobble:
			frameIdx = 1
		case ChestOpening:
			if chest.GetOpenProgress() < 0.5 {
				frameIdx = 2
			} else {
				frameIdx = 3
			}
		default:
			// Open states
			if chest.State >= ChestRevealing {
				frameIdx = 4 // Open
				// Use glowing frame when revealing/choosing
				if chest.State == ChestRevealing || chest.State == ChestChoosing {
					frameIdx = 5 // Open with glow
				}
			}
		}

		// Source rectangle from sprite sheet (32x32 per frame, 6 frames horizontal)
		srcRect := rl.Rectangle{
			X:      float32(frameIdx * 32),
			Y:      0,
			Width:  32,
			Height: 32,
		}

		// Destination (scale 2x)
		destRect := rl.Rectangle{
			X:      float32(chestX + wobbleOffset),
			Y:      float32(chestY),
			Width:  64,
			Height: 64,
		}

		rl.DrawTexturePro(r.chestTexture, srcRect, destRect, rl.Vector2{}, 0, rl.White)
	} else {
		// Fallback: draw a simple rectangle
		rl.DrawRectangle(chestX+wobbleOffset, chestY, 64, 48, rl.Color{R: 139, G: 69, B: 19, A: 255})
	}

	// Spawn sparkle particles when chest is opening or revealing
	if chest.State == ChestOpening || chest.State == ChestRevealing {
		r.spawnChestParticles(float32(chestX+32), float32(chestY+20), chest.State == ChestRevealing)
	}

	// Draw items when revealed
	if chest.State >= ChestRevealing && chest.HasItems() {
		revealProgress := chest.GetRevealProgress()
		itemY := int32(45) // Fixed position near top

		// Calculate item positions (spread horizontally)
		numItems := len(chest.Items)
		boxW := int32(58)
		itemSpacing := boxW + 6
		totalWidth := int32(numItems)*itemSpacing - 6
		startX := (screenWidth - totalWidth) / 2

		for i, item := range chest.Items {
			itemX := startX + int32(i)*itemSpacing

			// Item box
			boxColor := rl.Color{R: 40, G: 35, B: 60, A: 230}
			borderColor := rl.Color{R: 80, G: 70, B: 110, A: 255}

			// Highlight selected item in choosing state
			if chest.State == ChestChoosing && i == chest.SelectedIdx {
				boxColor = rl.Color{R: 60, G: 50, B: 90, A: 255}
				borderColor = rl.Color{R: 255, G: 200, B: 80, A: 255} // Gold highlight
			}

			// Fade in based on reveal progress
			boxAlpha := uint8(revealProgress * 255)
			boxColor.A = boxAlpha
			borderColor.A = boxAlpha

			// Draw box
			boxH := int32(32)
			rl.DrawRectangle(itemX-1, itemY-1, boxW+2, boxH+2, borderColor)
			rl.DrawRectangle(itemX, itemY, boxW, boxH, boxColor)

			// Item name (truncate if needed)
			textColor := rl.Color{R: 220, G: 215, B: 240, A: boxAlpha}
			nameText := item.Name
			// Measure and truncate to fit
			for len(nameText) > 0 && rl.MeasureText(nameText, 6) > boxW-4 {
				nameText = nameText[:len(nameText)-1]
			}
			textW := rl.MeasureText(nameText, 6)
			rl.DrawText(nameText, itemX+(boxW-textW)/2, itemY+6, 6, textColor)

			// Slot type (smaller, below name)
			slotText := string(item.Slot)
			slotColor := rl.Color{R: 150, G: 145, B: 170, A: boxAlpha}
			slotW := rl.MeasureText(slotText, 6)
			rl.DrawText(slotText, itemX+(boxW-slotW)/2, itemY+18, 6, slotColor)
		}
	}

	// Draw "LEVEL UP!" or "BONUS!" text at top (always on top)
	var titleText string
	var titleColor rl.Color
	if chest.Type == ChestTypeLevelUp {
		titleText = "LEVEL UP!"
		titleColor = rl.Color{R: 255, G: 215, B: 0, A: 255} // Gold
	} else {
		titleText = "BONUS!"
		titleColor = rl.Color{R: 100, G: 255, B: 150, A: 255} // Green
	}

	// Pulsing title
	pulse := float32(1.0 + 0.15*simpleSinF(float64(chest.Timer*6)))
	titleFontSize := int32(14 * pulse)
	titleWidth := rl.MeasureText(titleText, titleFontSize)
	titleX := (screenWidth - titleWidth) / 2
	titleY := int32(15)

	// Shadow
	rl.DrawText(titleText, titleX+1, titleY+1, titleFontSize, rl.Color{R: 0, G: 0, B: 0, A: 200})
	rl.DrawText(titleText, titleX, titleY, titleFontSize, titleColor)

	// Instructions when choosing
	if chest.State == ChestChoosing {
		instructText := "< > select   ENTER claim"
		instructW := rl.MeasureText(instructText, 6)
		instructColor := rl.Color{R: 150, G: 145, B: 170, A: 200}
		rl.DrawText(instructText, (screenWidth-instructW)/2, chestY+70, 6, instructColor)
	}

	// "NEW ITEM!" celebration when claiming
	if chest.State == ChestClaiming && chest.ClaimedItem != nil {
		claimProgress := chest.GetClaimProgress()
		celebAlpha := uint8((1.0 - claimProgress) * 255)

		celebText := fmt.Sprintf("NEW: %s!", chest.ClaimedItem.Name)
		celebW := rl.MeasureText(celebText, 10)
		celebX := (screenWidth - celebW) / 2
		celebY := int32(90 - claimProgress*20)

		celebColor := rl.Color{R: 255, G: 215, B: 0, A: celebAlpha}
		shadowColor := rl.Color{R: 0, G: 0, B: 0, A: celebAlpha / 2}

		rl.DrawText(celebText, celebX+1, celebY+1, 10, shadowColor)
		rl.DrawText(celebText, celebX, celebY, 10, celebColor)
	}

	// Handle empty pool case
	if chest.State >= ChestRevealing && !chest.HasItems() {
		noItemsText := "All items unlocked!"
		noItemsW := rl.MeasureText(noItemsText, 8)
		rl.DrawText(noItemsText, (screenWidth-noItemsW)/2, 50, 8, rl.Color{R: 180, G: 175, B: 200, A: 255})

		bonusText := "+500 XP Bonus!"
		bonusW := rl.MeasureText(bonusText, 10)
		bonusColor := rl.Color{R: 255, G: 200, B: 80, A: 255}
		rl.DrawText(bonusText, (screenWidth-bonusW)/2, 65, 10, bonusColor)
	}
}
