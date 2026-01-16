package main

import rl "github.com/gen2brain/raylib-go/raylib"

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
			// Explosive rise (reduced height to stay in frame)
			riseY := []int{1, -1, -2, -3}
			return 0, float32(riseY[f-4])
		} else if f < 14 {
			// Peak pose with subtle bob (peakY=-3)
			bob := []int{0, 1, 0, -1, 0, 1}
			return 0, float32(-3 + bob[f-8])
		} else if f < 18 {
			// Settle down
			settleY := []int{-2, -1, 0, 0}
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
	hatName := r.hatNames[r.currentHat]

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

	var hatX, hatY float32

	// Special positioning for headphones - wrap around head at ear level
	if hatName == "headphones" {
		// Headphones: band on top, cups at ear level (sprite y ~14-17)
		// Position so the top of headphones aligns with top of head
		// Scale wider to wrap around head properly
		hatW = hatW * 1.4
		spriteHeadY := float32(9) // Top of head
		hatX = claudeX + scaledW/2 - hatW/2 + headOffX*float32(claudeScale)
		hatY = claudeY + (spriteHeadY+headOffY)*float32(claudeScale)
	} else {
		// Standard hat positioning - sits ON TOP of head
		// In sprite space, Claude's body top is at y=12
		// Hat should sit ON TOP of the body, so just above y=12
		// Hat bottom edge should be around sprite y=10
		spriteHeadY := float32(10) // Where top of head is in 32x32 sprite

		// Convert to screen coords:
		// claudeY is the top-left of the 32x32 sprite frame (scaled)
		// Add spriteHeadY * scale to get head position
		// Add animation offset * scale
		hatX = claudeX + scaledW/2 - hatW/2 + headOffX*float32(claudeScale)
		hatY = claudeY + (spriteHeadY+headOffY)*float32(claudeScale) - hatH + 2*float32(claudeScale)
	}

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
	case "pipe":
		// Pipe stem comes from mouth (right side), bowl extends outward
		// Stem starts at sprite y row 7-9, position at mouth level y~18
		spriteY = 12
		spriteXOffset = 3
		centerVertically = false
	case "eyepatch":
		// Eyepatch over left eye (left eye is at sprite x=11-13, y=13-16)
		// Patch portion is at sprite columns 1-5, diagonal strap
		spriteY = 12
		spriteXOffset = -3
		centerVertically = false
	case "wizardbeard":
		// Beard hangs below face/chin
		spriteY = 18
		centerVertically = false
	case "bandana":
		// Bandana headband sits on forehead above eyes
		spriteY = 10
		centerVertically = false
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
