package main

import (
	"image"
	"image/png"
	"os"
)

// Mini Claude sprite sheet generator
// Mini Claude is 16x16 pixels (half size of main Claude)
// Animations: Spawn (jump out), Idle, Walk, Poof (disappear)

const (
	miniFrameWidth  = 16
	miniFrameHeight = 16
	miniNumAnims    = 4
	miniMaxFrames   = 12
)

// Mini animation frame counts
// Spawn, Idle, Walk, Poof
var miniFrameCounts = []int{8, 8, 8, 6}

func generateMiniClaude() {
	width := miniFrameWidth * miniMaxFrames
	height := miniFrameHeight * miniNumAnims
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill transparent
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, X)
		}
	}

	// Generate each animation
	for anim := 0; anim < miniNumAnims; anim++ {
		for frame := 0; frame < miniFrameCounts[anim]; frame++ {
			drawMiniFrame(img, anim, frame)
		}
	}

	// Save
	os.MkdirAll("assets/claude", 0755)
	f, _ := os.Create("assets/claude/mini_spritesheet.png")
	defer f.Close()
	png.Encode(f, img)
}

func drawMiniFrame(img *image.RGBA, anim, frame int) {
	offsetX := frame * miniFrameWidth
	offsetY := anim * miniFrameHeight

	switch anim {
	case 0: // Spawn - jump out arc
		drawMiniSpawn(img, offsetX, offsetY, frame)
	case 1: // Idle - cute bounce/vibe
		drawMiniIdle(img, offsetX, offsetY, frame)
	case 2: // Walk - tiny waddle
		drawMiniWalk(img, offsetX, offsetY, frame)
	case 3: // Poof - disappear in sparkles
		drawMiniPoof(img, offsetX, offsetY, frame)
	}
}

// Draw mini Claude blob - simplified version
// Body is ~9x5 pixels centered in 16x16 frame
func drawMiniBlob(img *image.RGBA, ox, oy, yOffset int, happy bool) {
	// Body: 9 wide, 5 tall, centered
	bodyTop := 5 + yOffset
	bodyBottom := 10 + yOffset
	bodyLeft := 3
	bodyRight := 12

	// Draw main body
	for y := bodyTop; y < bodyBottom; y++ {
		for x := bodyLeft; x < bodyRight; x++ {
			py := oy + y
			px := ox + x

			var c C
			if x < bodyLeft+1 {
				c = S // Left shadow
			} else if x >= bodyRight-1 {
				c = H // Right highlight
			} else if y < bodyTop+1 {
				c = H // Top highlight
			} else if y >= bodyBottom-1 {
				c = S // Bottom shadow
			} else {
				c = P // Main body
			}
			img.Set(px, py, c)
		}
	}

	// Eyes (2x2 dark squares)
	eyeY := oy + 6 + yOffset
	if happy {
		// Happy eyes ^_^
		img.Set(ox+5, eyeY+1, O)
		img.Set(ox+6, eyeY, O)
		img.Set(ox+9, eyeY, O)
		img.Set(ox+10, eyeY+1, O)
	} else {
		// Normal eyes
		img.Set(ox+5, eyeY, O)
		img.Set(ox+6, eyeY, O)
		img.Set(ox+5, eyeY+1, O)
		img.Set(ox+6, eyeY+1, O)

		img.Set(ox+9, eyeY, O)
		img.Set(ox+10, eyeY, O)
		img.Set(ox+9, eyeY+1, O)
		img.Set(ox+10, eyeY+1, O)
	}

	// Tiny arms (1x2 on sides)
	armY := oy + 7 + yOffset
	img.Set(ox+2, armY, P)
	img.Set(ox+2, armY+1, S)
	img.Set(ox+13, armY, P)
	img.Set(ox+13, armY+1, H)

	// Tiny legs (2 pairs, 1x2 each)
	legY := oy + 10 + yOffset
	img.Set(ox+4, legY, S)
	img.Set(ox+4, legY+1, S)
	img.Set(ox+6, legY, P)
	img.Set(ox+6, legY+1, P)
	img.Set(ox+9, legY, P)
	img.Set(ox+9, legY+1, P)
	img.Set(ox+11, legY, H)
	img.Set(ox+11, legY+1, H)
}

func drawMiniSpawn(img *image.RGBA, ox, oy, frame int) {
	// 8 frame spawn: jump up from bottom (0-3), arc down and land (4-7)
	// Simulates jumping out of big Claude

	switch frame {
	case 0: // Start - tiny, below frame
		// Just sparkle at spawn point
		img.Set(ox+8, oy+14, Y)
		img.Set(ox+7, oy+13, W)
		img.Set(ox+9, oy+13, W)

	case 1: // Emerging - small blob jumping up
		img.Set(ox+8, oy+12, Y)
		img.Set(ox+6, oy+11, W)
		img.Set(ox+10, oy+11, W)
		// Tiny body forming
		for x := 6; x < 10; x++ {
			img.Set(ox+x, oy+10, P)
		}

	case 2: // Rising - stretched upward
		// Stretched mini blob going up
		for y := 6; y < 12; y++ {
			for x := 6; x < 10; x++ {
				c := P
				if x == 6 {
					c = S
				} else if x == 9 {
					c = H
				}
				img.Set(ox+x, oy+y, c)
			}
		}
		// Sparkle trail
		img.Set(ox+7, oy+13, Y)
		img.Set(ox+8, oy+14, Y)

	case 3: // Peak of jump - highest point, squashed
		drawMiniBlob(img, ox, oy, -4, true) // High up, happy
		// Sparkles around
		img.Set(ox+2, oy+4, Y)
		img.Set(ox+13, oy+3, W)

	case 4: // Starting descent
		drawMiniBlob(img, ox, oy, -3, true)
		img.Set(ox+3, oy+5, Y)

	case 5: // Falling
		drawMiniBlob(img, ox, oy, -1, false)

	case 6: // About to land - stretched down
		drawMiniBlob(img, ox, oy, 1, false)

	case 7: // Landed - squash then settle
		drawMiniBlob(img, ox, oy, 2, true)
		// Landing sparkles
		img.Set(ox+2, oy+12, Y)
		img.Set(ox+13, oy+12, Y)
	}
}

func drawMiniIdle(img *image.RGBA, ox, oy, frame int) {
	// 8 frame idle: gentle bounce/vibe
	bounceCurve := []int{0, 0, -1, -1, 0, 0, 1, 1}
	bounce := bounceCurve[frame]

	// Blink on frame 6
	happy := frame == 6 || frame == 7

	drawMiniBlob(img, ox, oy, bounce, happy)
}

func drawMiniWalk(img *image.RGBA, ox, oy, frame int) {
	// 8 frame walk cycle
	bobCurve := []int{0, 0, 1, 1, 0, 0, 1, 1}
	bob := bobCurve[frame]

	// Draw body
	bodyTop := 5 + bob
	bodyBottom := 10 + bob

	for y := bodyTop; y < bodyBottom; y++ {
		for x := 3; x < 12; x++ {
			py := oy + y
			px := ox + x
			var c C
			if x < 4 {
				c = S
			} else if x >= 11 {
				c = H
			} else if y < bodyTop+1 {
				c = H
			} else if y >= bodyBottom-1 {
				c = S
			} else {
				c = P
			}
			img.Set(px, py, c)
		}
	}

	// Eyes
	eyeY := oy + 6 + bob
	img.Set(ox+5, eyeY, O)
	img.Set(ox+6, eyeY, O)
	img.Set(ox+5, eyeY+1, O)
	img.Set(ox+6, eyeY+1, O)
	img.Set(ox+9, eyeY, O)
	img.Set(ox+10, eyeY, O)
	img.Set(ox+9, eyeY+1, O)
	img.Set(ox+10, eyeY+1, O)

	// Arms
	armY := oy + 7 + bob
	img.Set(ox+2, armY, P)
	img.Set(ox+2, armY+1, S)
	img.Set(ox+13, armY, P)
	img.Set(ox+13, armY+1, H)

	// Animated legs - wave pattern
	legY := oy + 10 + bob
	legPhase := frame % 4
	legOffsets := [][]int{
		{0, 1, 0, -1}, // leg 1
		{1, 0, -1, 0}, // leg 2
		{0, -1, 0, 1}, // leg 3
		{-1, 0, 1, 0}, // leg 4
	}

	// Leg 1
	img.Set(ox+4+legOffsets[0][legPhase], legY, S)
	img.Set(ox+4+legOffsets[0][legPhase], legY+1, S)
	// Leg 2
	img.Set(ox+6+legOffsets[1][legPhase], legY, P)
	img.Set(ox+6+legOffsets[1][legPhase], legY+1, P)
	// Leg 3
	img.Set(ox+9+legOffsets[2][legPhase], legY, P)
	img.Set(ox+9+legOffsets[2][legPhase], legY+1, P)
	// Leg 4
	img.Set(ox+11+legOffsets[3][legPhase], legY, H)
	img.Set(ox+11+legOffsets[3][legPhase], legY+1, H)
}

func drawMiniPoof(img *image.RGBA, ox, oy, frame int) {
	// 6 frame poof: shrink and sparkle away
	switch frame {
	case 0: // Full size, starting to glow
		drawMiniBlob(img, ox, oy, 0, true)
		img.Set(ox+2, oy+4, Y)
		img.Set(ox+13, oy+4, Y)

	case 1: // Glowing more
		drawMiniBlob(img, ox, oy, 0, true)
		img.Set(ox+1, oy+3, Y)
		img.Set(ox+14, oy+3, Y)
		img.Set(ox+3, oy+12, W)
		img.Set(ox+12, oy+12, W)

	case 2: // Starting to shrink
		// Smaller body
		for y := 6; y < 10; y++ {
			for x := 5; x < 11; x++ {
				img.Set(ox+x, oy+y, P)
			}
		}
		// Sparkles expanding
		img.Set(ox+2, oy+5, Y)
		img.Set(ox+13, oy+5, Y)
		img.Set(ox+4, oy+3, W)
		img.Set(ox+11, oy+3, W)
		img.Set(ox+7, oy+12, Y)

	case 3: // Small blob
		for y := 7; y < 9; y++ {
			for x := 6; x < 10; x++ {
				img.Set(ox+x, oy+y, P)
			}
		}
		// More sparkles
		img.Set(ox+3, oy+4, Y)
		img.Set(ox+12, oy+4, W)
		img.Set(ox+5, oy+11, Y)
		img.Set(ox+10, oy+11, W)
		img.Set(ox+8, oy+2, Y)

	case 4: // Tiny dot
		img.Set(ox+7, oy+8, P)
		img.Set(ox+8, oy+8, P)
		// Expanding sparkle ring
		img.Set(ox+4, oy+5, Y)
		img.Set(ox+11, oy+5, Y)
		img.Set(ox+4, oy+11, W)
		img.Set(ox+11, oy+11, W)
		img.Set(ox+2, oy+8, Y)
		img.Set(ox+13, oy+8, Y)

	case 5: // Gone - just fading sparkles
		img.Set(ox+3, oy+4, Y)
		img.Set(ox+12, oy+4, W)
		img.Set(ox+3, oy+12, W)
		img.Set(ox+12, oy+12, Y)
		img.Set(ox+7, oy+7, Y)
	}
}
