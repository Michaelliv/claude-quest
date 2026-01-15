package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

// Type alias for cleaner pattern definitions
type C = color.RGBA

const (
	frameWidth  = 32
	frameHeight = 32
	numAnims    = 9
	maxFrames   = 24
)

// Animation frame counts (must match animations.go) - doubled for smoothness
// Idle, Enter, Casting, Attack, Writing, Victory, Hurt, Thinking, Walk
var frameCounts = []int{16, 20, 16, 16, 16, 20, 16, 12, 16}

// Claude's official color palette from Clawdachi
var (
	P = C{0xFF, 0x99, 0x33, 0xFF} // Primary Orange #FF9933
	S = C{0xCC, 0x66, 0x00, 0xFF} // Shadow Orange #CC6600
	H = C{0xFF, 0xBB, 0x77, 0xFF} // Highlight Orange #FFBB77
	O = C{0x22, 0x22, 0x22, 0xFF} // Outline/Eyes #222222
	M = C{0x44, 0x22, 0x00, 0xFF} // Mouth #442200
	W = C{0xFF, 0xFF, 0xFF, 0xFF} // White
	G = C{0x00, 0xFF, 0x88, 0xFF} // Terminal Green #00FF88
	Y = C{0xFF, 0xF5, 0x96, 0xFF} // Spark Yellow
	X = C{0x00, 0x00, 0x00, 0x00} // Transparent
)

func createImage(width, height int) *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, width, height))
}

func main() {
	// Create sprite sheet
	width := frameWidth * maxFrames
	height := frameHeight * numAnims
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill transparent
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, X)
		}
	}

	// Generate each animation
	for anim := 0; anim < numAnims; anim++ {
		for frame := 0; frame < frameCounts[anim]; frame++ {
			drawFrame(img, anim, frame)
		}
	}

	// Save
	os.MkdirAll("assets/claude", 0755)
	f, _ := os.Create("assets/claude/spritesheet.png")
	defer f.Close()
	png.Encode(f, img)

	// Generate accessories and effects
	generateAccessories()

	// Generate mini Claude for subagents
	generateMiniClaude()
}

func drawFrame(img *image.RGBA, anim, frame int) {
	offsetX := frame * frameWidth
	offsetY := anim * frameHeight

	switch anim {
	case 0: // Idle - smooth breathing with blink
		// 16 frames: breathe in (0-7), breathe out (8-15), blink at frame 12-14
		breathCurve := []int{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0}
		breathOffset := breathCurve[frame]
		drawClaudeBlob(img, offsetX, offsetY, breathOffset, false, false)
		// Blink
		if frame >= 12 && frame <= 14 {
			drawBlink(img, offsetX, offsetY+breathOffset, frame-12)
		}

	case 1: // Enter - pop in / materialize effect
		drawClaudeEnter(img, offsetX, offsetY, frame)

	case 2: // Casting - reading/searching (arms up, sparkles)
		drawClaudeCasting(img, offsetX, offsetY, frame)

	case 3: // Attack - bash command (punch motion)
		drawClaudeAttack(img, offsetX, offsetY, frame)

	case 4: // Writing - edit/write (typing motion)
		drawClaudeWriting(img, offsetX, offsetY, frame)

	case 5: // Victory - success (jumping)
		drawClaudeVictory(img, offsetX, offsetY, frame)

	case 6: // Hurt - error (knocked back)
		drawClaudeHurt(img, offsetX, offsetY, frame)

	case 7: // Thinking - processing
		drawClaudeThinking(img, offsetX, offsetY, frame)

	case 8: // Walk - infinite walking cycle
		drawClaudeWalk(img, offsetX, offsetY, frame)
	}
}

// Draw the base Claude blob - a cute rectangular crab-like creature
func drawClaudeBlob(img *image.RGBA, ox, oy, breathOffset int, armsUp, legsWide bool) {
	// Body is wider rectangle - more horizontal like the reference
	// Body: 18 pixels wide, 10 pixels tall

	bodyTop := 12 - breathOffset
	bodyBottom := 22
	bodyLeft := 7
	bodyRight := 25

	// Draw main body
	for y := bodyTop; y < bodyBottom; y++ {
		for x := bodyLeft; x < bodyRight; x++ {
			py := oy + y
			px := ox + x

			// Determine color based on position (shading)
			var c color.RGBA
			if x < bodyLeft+2 {
				c = S // Left shadow
			} else if x >= bodyRight-2 {
				c = H // Right highlight
			} else if y < bodyTop+2 {
				c = H // Top highlight
			} else if y >= bodyBottom-2 {
				c = S // Bottom shadow
			} else {
				c = P // Main body
			}
			img.Set(px, py, c)
		}
	}

	// Eyes (3x4 dark rectangles)
	eyeY := oy + 13 - breathOffset
	// Left eye at x=11
	for dy := 0; dy < 4; dy++ {
		for dx := 0; dx < 3; dx++ {
			img.Set(ox+11+dx, eyeY+dy, O)
		}
	}
	// Right eye at x=18
	for dy := 0; dy < 4; dy++ {
		for dx := 0; dx < 3; dx++ {
			img.Set(ox+18+dx, eyeY+dy, O)
		}
	}

	// Arms (small 3x3 stubs on sides, right at body edge)
	armY := oy + 15 - breathOffset // vertically centered on body
	if armsUp {
		armY = oy + 10 - breathOffset
	}

	// Left arm - just to the left of body (body starts at x=7)
	for dy := 0; dy < 3; dy++ {
		for dx := 0; dx < 3; dx++ {
			c := P
			if dx == 0 {
				c = S
			}
			img.Set(ox+4+dx, armY+dy, c) // x=4,5,6
		}
	}
	// Right arm - just to the right of body (body ends at x=25)
	for dy := 0; dy < 3; dy++ {
		for dx := 0; dx < 3; dx++ {
			c := P
			if dx == 2 {
				c = H
			}
			img.Set(ox+25+dx, armY+dy, c) // x=25,26,27
		}
	}

	// 4 Legs - 2 on left, 2 on right, with gap in middle
	legY := oy + 22
	spread := 0
	if legsWide {
		spread = 1
	}

	// Left side - 2 legs (at positions 8-9 and 11-12)
	for dy := 0; dy < 5; dy++ {
		img.Set(ox+8-spread, legY+dy, S)
		img.Set(ox+9-spread, legY+dy, P)
	}
	for dy := 0; dy < 5; dy++ {
		img.Set(ox+11, legY+dy, S)
		img.Set(ox+12, legY+dy, P)
	}
	// Right side - 2 legs (at positions 19-20 and 22-23)
	for dy := 0; dy < 5; dy++ {
		img.Set(ox+19, legY+dy, P)
		img.Set(ox+20, legY+dy, H)
	}
	for dy := 0; dy < 5; dy++ {
		img.Set(ox+22+spread, legY+dy, P)
		img.Set(ox+23+spread, legY+dy, H)
	}
}

func drawClaudeEnter(img *image.RGBA, ox, oy, frame int) {
	// 20 frame pop-in: sparkles (0-7), materialize (8-14), bounce settle (15-19)
	if frame < 8 {
		// Sparkles appearing - more frames, smoother build
		numSparkles := frame + 1
		sparkPositions := [][]int{
			{16, 16}, {14, 14}, {18, 18}, {12, 12},
			{20, 12}, {10, 16}, {22, 16}, {16, 20},
		}
		for i := 0; i < numSparkles && i < len(sparkPositions); i++ {
			p := sparkPositions[i]
			if (frame+i)%2 == 0 {
				img.Set(ox+p[0], oy+p[1], Y)
			} else {
				img.Set(ox+p[0], oy+p[1], W)
			}
		}
	} else if frame < 15 {
		// Materialize with squash-stretch
		progress := frame - 8 // 0-6
		squash := []int{3, 2, 1, 0, -1, 0, 0}[progress]
		drawBlobSquashed(img, ox, oy, 0, squash)
		// Fading sparkles
		if progress < 4 {
			img.Set(ox+6, oy+10, Y)
			img.Set(ox+26, oy+12, Y)
		}
	} else {
		// Settle bounce
		bounce := []int{-2, -1, 0, 0, 0}[frame-15]
		drawClaudeBlob(img, ox, oy+bounce, 0, false, false)
	}
}

// drawClaudeWalk draws a 16-frame walk cycle with each leg moving independently
// Wave pattern flows through legs: outer-left -> inner-left -> inner-right -> outer-right
func drawClaudeWalk(img *image.RGBA, ox, oy, frame int) {
	// Subtle body bob from leg motion
	bobCurve := []int{0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0}
	bob := bobCurve[frame]

	// Draw body
	bodyTop := 12 + bob
	bodyBottom := 22 + bob

	for y := bodyTop; y < bodyBottom; y++ {
		for x := 7; x < 25; x++ {
			py := oy + y
			px := ox + x
			var c C
			if x < 9 {
				c = S
			} else if x >= 23 {
				c = H
			} else if y < bodyTop+2 {
				c = H
			} else if y >= bodyBottom-2 {
				c = S
			} else {
				c = P
			}
			img.Set(px, py, c)
		}
	}

	// Eyes
	eyeY := oy + 13 + bob
	for dy := 0; dy < 4; dy++ {
		for dx := 0; dx < 3; dx++ {
			img.Set(ox+11+dx, eyeY+dy, O)
			img.Set(ox+18+dx, eyeY+dy, O)
		}
	}

	// Arms - slight bob
	armY := oy + 15 + bob
	for dy := 0; dy < 3; dy++ {
		img.Set(ox+4, armY+dy, S)
		img.Set(ox+5, armY+dy, P)
		img.Set(ox+6, armY+dy, P)
		img.Set(ox+25, armY+dy, P)
		img.Set(ox+26, armY+dy, P)
		img.Set(ox+27, armY+dy, H)
	}

	// 4 legs - each has independent timing, wave flows through
	// Each leg cycle: plant (0-3) -> lift (4-5) -> swing (6-7) -> plant
	// Legs are offset by 4 frames each for wave effect
	legY := oy + 22 + bob

	// Leg motion: lift amount and forward/back position
	// 16 frame cycle per leg, but each leg starts at different phase
	legCycle := func(phase int) (lift, slide int) {
		p := phase % 16
		// Smoother curve: mostly planted, brief lift and swing
		lifts := []int{0, 0, 0, 0, 0, 0, 1, 2, 2, 1, 0, 0, 0, 0, 0, 0}
		slides := []int{1, 1, 1, 1, 0, 0, 0, -1, -1, 0, 0, 1, 1, 1, 1, 1}
		return lifts[p], slides[p]
	}

	// Each leg offset by 4 frames - creates wave from left to right
	l1Lift, l1Slide := legCycle(frame)      // outer left
	l2Lift, l2Slide := legCycle(frame + 4)  // inner left
	l3Lift, l3Slide := legCycle(frame + 8)  // inner right
	l4Lift, l4Slide := legCycle(frame + 12) // outer right

	// Draw legs
	// Leg 1 - outer left (x=8-9)
	for dy := 0; dy < 5; dy++ {
		img.Set(ox+8+l1Slide, legY+dy-l1Lift, S)
		img.Set(ox+9+l1Slide, legY+dy-l1Lift, P)
	}
	// Leg 2 - inner left (x=11-12)
	for dy := 0; dy < 5; dy++ {
		img.Set(ox+11+l2Slide, legY+dy-l2Lift, S)
		img.Set(ox+12+l2Slide, legY+dy-l2Lift, P)
	}
	// Leg 3 - inner right (x=19-20)
	for dy := 0; dy < 5; dy++ {
		img.Set(ox+19+l3Slide, legY+dy-l3Lift, P)
		img.Set(ox+20+l3Slide, legY+dy-l3Lift, H)
	}
	// Leg 4 - outer right (x=22-23)
	for dy := 0; dy < 5; dy++ {
		img.Set(ox+22+l4Slide, legY+dy-l4Lift, P)
		img.Set(ox+23+l4Slide, legY+dy-l4Lift, H)
	}
}

func drawClaudeCasting(img *image.RGBA, ox, oy, frame int) {
	// 16 frame casting: wind up (0-4), arms up magic (5-12), settle (13-15)
	switch {
	case frame < 5: // Wind up - squat then stretch up
		squash := []int{-1, -2, -1, 1, 2}[frame]
		drawBlobSquashed(img, ox, oy+1, 0, squash)

	case frame < 13: // Magic casting with arms up
		// Slight float
		floatY := []int{-1, -2, -2, -2, -1, -1, -2, -2}[frame-5]
		drawClaudeBlob(img, ox, oy+floatY, 0, true, false)

		// Rotating sparkle pattern
		sparkPhase := frame - 5
		sparkRadius := 10
		for i := 0; i < 6; i++ {
			angle := (sparkPhase*30 + i*60) % 360
			// Simple angle to x,y
			sx := ox + 16 + (sparkRadius * simpleCos(angle) / 100)
			sy := oy + 8 + floatY + (sparkRadius * simpleSin(angle) / 100)
			if (i+frame)%2 == 0 {
				img.Set(sx, sy, Y)
			} else {
				img.Set(sx, sy, W)
			}
		}

		// Central glow
		if frame%2 == 0 {
			img.Set(ox+16, oy+6+floatY, W)
		}

	default: // Settle back down
		settleY := []int{-1, 0, 0}[frame-13]
		drawClaudeBlob(img, ox, oy+settleY, 0, false, false)
	}
}

func simpleSin(deg int) int {
	// Simple sine lookup (returns -100 to 100)
	sins := []int{0, 50, 87, 100, 87, 50, 0, -50, -87, -100, -87, -50}
	return sins[(deg/30)%12]
}

func simpleCos(deg int) int {
	return simpleSin(deg + 90)
}

func drawClaudeAttack(img *image.RGBA, ox, oy, frame int) {
	// 16 frame attack with smooth anticipation, smear, impact, follow-through
	// 0-4: Anticipation (smooth squat, arm pull back)
	// 5-6: Wind-up peak
	// 7: Smear frame
	// 8-9: Impact
	// 10-13: Follow-through bounce
	// 14-15: Recovery

	switch {
	case frame < 3: // Gradual anticipation
		squash := []int{-1, -2, -2}[frame]
		drawBlobSquashed(img, ox, oy+1+frame/2, 1, squash)
		// Arm pulling back gradually
		armX := 4 - frame
		for dy := 0; dy < 3; dy++ {
			img.Set(ox+armX, oy+16+dy, S)
			img.Set(ox+armX+1, oy+16+dy, P)
			img.Set(ox+armX+2, oy+16+dy, P)
		}

	case frame < 5: // Deep coil
		drawBlobSquashed(img, ox, oy+3, 2, -3)
		for dy := 0; dy < 3; dy++ {
			img.Set(ox+0, oy+15+dy, S)
			img.Set(ox+1, oy+15+dy, P)
			img.Set(ox+2, oy+15+dy, P)
		}

	case frame < 7: // Wind-up with squint
		drawBlobSquashed(img, ox, oy+2, 1, -2)
		drawEyesSquint(img, ox, oy+2)

	case frame == 7: // SMEAR
		drawBlobSmearHorizontal(img, ox, oy)
		for dy := 0; dy < 2; dy++ {
			for dx := 0; dx < 15; dx++ {
				c := P
				if dx > 10 {
					c = H
				}
				img.Set(ox+14+dx, oy+15+dy, c)
			}
		}

	case frame < 10: // Impact frames
		drawBlobSquashed(img, ox+3, oy+1, -1, 1)
		// Extended arm
		armLen := 8
		for dy := 0; dy < 3; dy++ {
			for dx := 0; dx < armLen; dx++ {
				c := P
				if dx > armLen-3 {
					c = H
				}
				img.Set(ox+24+dx, oy+15+dy, c)
			}
		}
		if frame == 8 {
			drawImpactBurst(img, ox+31, oy+14)
		} else {
			// Fading burst
			img.Set(ox+31, oy+14, Y)
			img.Set(ox+30, oy+12, Y)
			img.Set(ox+30, oy+16, Y)
		}

	case frame < 14: // Follow-through bounce
		bounceY := []int{-2, -1, 0, 1}[frame-10]
		stretchY := []int{1, 0, 0, -1}[frame-10]
		drawBlobSquashed(img, ox+2-(frame-10)/2, oy+bounceY, 0, stretchY)
		// Arm retracting
		armLen := 6 - (frame - 10)
		for dy := 0; dy < 3; dy++ {
			for dx := 0; dx < armLen; dx++ {
				img.Set(ox+25+dx, oy+15+bounceY+dy, P)
			}
		}
		// Lingering sparks
		if frame < 12 {
			img.Set(ox+29, oy+12, Y)
		}

	default: // Recovery
		bounce := []int{-1, 0}[frame-14]
		drawClaudeBlob(img, ox, oy+bounce, 0, false, false)
	}
}

// Helper: Draw squashed/stretched blob
func drawBlobSquashed(img *image.RGBA, ox, oy, squashX, squashY int) {
	// Body dimensions adjusted by squash
	bodyTop := 12 + squashY
	bodyBottom := 22 - squashY
	bodyLeft := 7 - squashX
	bodyRight := 25 + squashX

	for y := bodyTop; y < bodyBottom; y++ {
		for x := bodyLeft; x < bodyRight; x++ {
			py := oy + y
			px := ox + x
			var c C
			if x < bodyLeft+2 {
				c = S
			} else if x >= bodyRight-2 {
				c = H
			} else if y < bodyTop+2 {
				c = H
			} else if y >= bodyBottom-2 {
				c = S
			} else {
				c = P
			}
			img.Set(px, py, c)
		}
	}

	// Eyes
	eyeY := oy + 14 + squashY
	for dy := 0; dy < 3; dy++ {
		for dx := 0; dx < 3; dx++ {
			img.Set(ox+11-squashX+dx, eyeY+dy, O)
			img.Set(ox+18+squashX+dx, eyeY+dy, O)
		}
	}

	// Legs
	legY := oy + 22 - squashY
	for dy := 0; dy < 4; dy++ {
		img.Set(ox+8-squashX, legY+dy, S)
		img.Set(ox+9-squashX, legY+dy, P)
		img.Set(ox+11, legY+dy, S)
		img.Set(ox+12, legY+dy, P)
		img.Set(ox+19, legY+dy, P)
		img.Set(ox+20, legY+dy, H)
		img.Set(ox+22+squashX, legY+dy, P)
		img.Set(ox+23+squashX, legY+dy, H)
	}

	// Arms (default position)
	armY := oy + 15 + squashY/2
	for dy := 0; dy < 3; dy++ {
		img.Set(ox+4-squashX, armY+dy, S)
		img.Set(ox+5-squashX, armY+dy, P)
		img.Set(ox+6-squashX, armY+dy, P)
		img.Set(ox+25+squashX, armY+dy, P)
		img.Set(ox+26+squashX, armY+dy, P)
		img.Set(ox+27+squashX, armY+dy, H)
	}
}

func drawEyesSquint(img *image.RGBA, ox, oy int) {
	// Determined/angry squint eyes "> <"
	eyeY := oy + 14
	// Left eye ">"
	img.Set(ox+11, eyeY, O)
	img.Set(ox+12, eyeY+1, O)
	img.Set(ox+11, eyeY+2, O)
	// Right eye "<"
	img.Set(ox+20, eyeY, O)
	img.Set(ox+19, eyeY+1, O)
	img.Set(ox+20, eyeY+2, O)
}

func drawBlobSmearHorizontal(img *image.RGBA, ox, oy int) {
	// Motion blur - stretched body
	bodyTop := 13
	bodyBottom := 21
	bodyLeft := 5  // stretched back
	bodyRight := 26 // normal right

	for y := bodyTop; y < bodyBottom; y++ {
		for x := bodyLeft; x < bodyRight; x++ {
			py := oy + y
			px := ox + x
			c := P
			if x < 10 {
				c = S // motion trail darker
			} else if x > 22 {
				c = H
			}
			img.Set(px, py, c)
		}
	}

	// Blurred eyes
	img.Set(ox+14, oy+15, O)
	img.Set(ox+15, oy+15, O)
	img.Set(ox+19, oy+15, O)
	img.Set(ox+20, oy+15, O)

	// No legs visible during smear - too fast!
}

func drawImpactBurst(img *image.RGBA, cx, cy int) {
	// Classic impact star burst
	img.Set(cx, cy, W)
	img.Set(cx+1, cy, W)
	img.Set(cx-1, cy, W)
	img.Set(cx, cy-1, W)
	img.Set(cx, cy+1, W)
	// Outer sparks
	img.Set(cx+3, cy-2, Y)
	img.Set(cx+3, cy+2, Y)
	img.Set(cx-2, cy-2, Y)
	img.Set(cx-2, cy+2, Y)
	img.Set(cx+4, cy, Y)
	img.Set(cx, cy-3, Y)
	img.Set(cx, cy+3, Y)
}

func drawBlink(img *image.RGBA, ox, oy, blinkFrame int) {
	// Overwrite eyes with blink state
	// blinkFrame: 0=closing, 1=closed, 2=opening
	eyeY := oy + 14

	// Clear existing eyes first by redrawing body color over them
	for dy := 0; dy < 4; dy++ {
		for dx := 0; dx < 3; dx++ {
			img.Set(ox+11+dx, eyeY+dy, P)
			img.Set(ox+18+dx, eyeY+dy, P)
		}
	}

	switch blinkFrame {
	case 0: // Half closed
		for dx := 0; dx < 3; dx++ {
			img.Set(ox+11+dx, eyeY+2, O)
			img.Set(ox+11+dx, eyeY+3, O)
			img.Set(ox+18+dx, eyeY+2, O)
			img.Set(ox+18+dx, eyeY+3, O)
		}
	case 1: // Fully closed - just a line
		for dx := 0; dx < 3; dx++ {
			img.Set(ox+11+dx, eyeY+2, O)
			img.Set(ox+18+dx, eyeY+2, O)
		}
	case 2: // Half open
		for dx := 0; dx < 3; dx++ {
			img.Set(ox+11+dx, eyeY+1, O)
			img.Set(ox+11+dx, eyeY+2, O)
			img.Set(ox+11+dx, eyeY+3, O)
			img.Set(ox+18+dx, eyeY+1, O)
			img.Set(ox+18+dx, eyeY+2, O)
			img.Set(ox+18+dx, eyeY+3, O)
		}
	}
}

func drawClaudeWriting(img *image.RGBA, ox, oy, frame int) {
	// 16 frame typing with smooth bob and rapid arm movement
	bobCurve := []int{0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0}
	bob := bobCurve[frame]
	drawClaudeBlob(img, ox, oy, bob, false, false)

	// Typing arm animation - rapid back and forth
	armPhase := frame % 4
	armOffset := []int{0, 1, 0, -1}[armPhase]

	// Overwrite right arm with typing motion
	armY := oy + 15 + bob
	for dy := 0; dy < 3; dy++ {
		img.Set(ox+25+armOffset, armY+dy, P)
		img.Set(ox+26+armOffset, armY+dy, P)
		img.Set(ox+27+armOffset, armY+dy, H)
	}

	// Terminal green sparkles (typing effect)
	if frame%3 == 0 {
		img.Set(ox+28+armOffset, oy+14, G)
	}
	if frame%4 == 1 {
		img.Set(ox+29, oy+16, G)
	}
}

func drawClaudeVictory(img *image.RGBA, ox, oy, frame int) {
	// 20 frame victory: anticipation (0-3), jump up (4-8), peak (9-11), fall (12-15), land bounce (16-19)
	switch {
	case frame < 4: // Anticipation squat
		squash := []int{0, -1, -2, -2}[frame]
		drawBlobSquashed(img, ox, oy+2, 0, squash)

	case frame < 9: // Jump up - stretch vertically
		jumpY := []int{0, 3, 6, 9, 11}[frame-4]
		stretch := []int{2, 2, 1, 1, 0}[frame-4]
		drawBlobSquashed(img, ox, oy-jumpY, 0, stretch)
		// Arms up
		armY := oy - jumpY + 10
		for dy := 0; dy < 3; dy++ {
			img.Set(ox+4, armY+dy, P)
			img.Set(ox+5, armY+dy, P)
			img.Set(ox+26, armY+dy, P)
			img.Set(ox+27, armY+dy, P)
		}

	case frame < 12: // Peak - happy wiggle
		jumpY := 12
		wiggle := []int{0, 1, 0}[frame-9]
		drawBlobSquashed(img, ox+wiggle, oy-jumpY, 0, 0)
		// Happy eyes ^_^
		drawHappyEyes(img, ox+wiggle, oy-jumpY)
		// Arms up wiggling
		armY := oy - jumpY + 10
		for dy := 0; dy < 3; dy++ {
			img.Set(ox+4-wiggle, armY+dy, P)
			img.Set(ox+27+wiggle, armY+dy, P)
		}
		// Sparkles!
		img.Set(ox+2, oy-jumpY+6, Y)
		img.Set(ox+29, oy-jumpY+6, Y)
		if frame == 10 {
			img.Set(ox+16, oy-jumpY-2, W)
		}

	case frame < 16: // Fall down
		jumpY := []int{10, 7, 4, 1}[frame-12]
		stretch := []int{1, 1, 0, -1}[frame-12]
		drawBlobSquashed(img, ox, oy-jumpY, 0, stretch)

	default: // Land and bounce settle
		bounceY := []int{2, 0, -1, 0}[frame-16]
		squash := []int{-2, -1, 1, 0}[frame-16]
		drawBlobSquashed(img, ox, oy+bounceY, 0, squash)
		// Happy expression lingers
		if frame < 18 {
			drawHappyEyes(img, ox, oy+bounceY)
		}
	}
}

func drawHappyEyes(img *image.RGBA, ox, oy int) {
	// ^_^ eyes
	eyeY := oy + 14
	// Overwrite with happy arcs
	for dx := 0; dx < 3; dx++ {
		for dy := 0; dy < 4; dy++ {
			img.Set(ox+11+dx, eyeY+dy, P) // clear left
			img.Set(ox+18+dx, eyeY+dy, P) // clear right
		}
	}
	// Left eye ^
	img.Set(ox+11, eyeY+2, O)
	img.Set(ox+12, eyeY+1, O)
	img.Set(ox+13, eyeY+2, O)
	// Right eye ^
	img.Set(ox+18, eyeY+2, O)
	img.Set(ox+19, eyeY+1, O)
	img.Set(ox+20, eyeY+2, O)
}

func drawClaudeHurt(img *image.RGBA, ox, oy, frame int) {
	// 16 frame hurt: impact squash (0-2), knockback (3-8), bounce recover (9-15)
	switch {
	case frame < 3: // Impact squash
		squash := []int{-3, -2, -1}[frame]
		drawBlobSquashed(img, ox, oy+2, -1, squash)
		// Impact burst
		if frame == 0 {
			drawImpactBurst(img, ox+28, oy+14)
		}

	case frame < 9: // Knockback with stretch
		knockX := []int{2, 5, 7, 8, 7, 5}[frame-3]
		stretch := []int{2, 2, 1, 0, 0, -1}[frame-3]
		drawBlobSquashed(img, ox-knockX, oy, -1, stretch)
		// X_X eyes
		drawXEyes(img, ox-knockX, oy)
		// Flying stars
		if frame < 7 {
			img.Set(ox+28-knockX/2, oy+10, Y)
			img.Set(ox+30-knockX/3, oy+14, W)
		}

	default: // Recovery bounce
		recoverX := []int{4, 3, 2, 1, 0, 0, 0}[frame-9]
		bounceY := []int{-1, 0, 1, 0, -1, 0, 0}[frame-9]
		drawBlobSquashed(img, ox-recoverX, oy+bounceY, 0, 0)
		// X eyes fade to normal
		if frame < 13 {
			drawXEyes(img, ox-recoverX, oy+bounceY)
		}
	}
}

func drawXEyes(img *image.RGBA, ox, oy int) {
	eyeY := oy + 14
	// Clear and draw X pattern
	for dx := 0; dx < 3; dx++ {
		for dy := 0; dy < 4; dy++ {
			img.Set(ox+11+dx, eyeY+dy, P)
			img.Set(ox+18+dx, eyeY+dy, P)
		}
	}
	// Left X
	img.Set(ox+11, eyeY, O)
	img.Set(ox+13, eyeY, O)
	img.Set(ox+12, eyeY+1, O)
	img.Set(ox+11, eyeY+3, O)
	img.Set(ox+13, eyeY+3, O)
	// Right X
	img.Set(ox+18, eyeY, O)
	img.Set(ox+20, eyeY, O)
	img.Set(ox+19, eyeY+1, O)
	img.Set(ox+18, eyeY+3, O)
	img.Set(ox+20, eyeY+3, O)
}

func drawClaudeThinking(img *image.RGBA, ox, oy, frame int) {
	// 12 frame thinking: subtle sway with growing thought bubble
	sway := []int{0, 0, 0, 1, 1, 1, 0, 0, 0, -1, -1, -1}[frame]
	drawClaudeBlob(img, ox+sway, oy, 0, false, false)

	// Eyes look up when thinking
	eyeY := oy + 13
	if frame > 2 {
		// Shift pupils up
		for dx := 0; dx < 3; dx++ {
			img.Set(ox+11+sway+dx, eyeY+3, P) // clear bottom
			img.Set(ox+18+sway+dx, eyeY+3, P)
		}
	}

	// Thought bubble grows
	dotY := oy + 6
	bubbleSize := frame / 3 // 0, 0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3

	if bubbleSize >= 1 {
		img.Set(ox+24, dotY+4, W)
	}
	if bubbleSize >= 2 {
		img.Set(ox+26, dotY+2, W)
		img.Set(ox+27, dotY+2, W)
	}
	if bubbleSize >= 3 {
		// Full thought bubble
		for dx := 0; dx < 4; dx++ {
			img.Set(ox+28+dx, dotY-2, W)
			img.Set(ox+28+dx, dotY-1, W)
			img.Set(ox+28+dx, dotY, W)
		}
		// Question mark or lightbulb?
		if frame%4 < 2 {
			img.Set(ox+30, dotY-1, Y) // lightbulb hint
		}
	}
}
