package main

import (
	"image"
	"image/png"
	"os"
)

// Enemy sprite sheet generator
// Each enemy is 32x16 pixels (wider for text)
// Enemies: Bug, ERROR, LOW CONTEXT

const (
	enemyFrameWidth  = 32
	enemyFrameHeight = 16
	enemyNumTypes    = 3
	enemyMaxFrames   = 4 // Simple animation frames
)

// Enemy types
const (
	EnemyBug = iota
	EnemyError
	EnemyLowContext
)

func generateEnemies() {
	width := enemyFrameWidth * enemyMaxFrames
	height := enemyFrameHeight * enemyNumTypes
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill transparent
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, X)
		}
	}

	// Generate each enemy type
	for enemyType := 0; enemyType < enemyNumTypes; enemyType++ {
		for frame := 0; frame < enemyMaxFrames; frame++ {
			drawEnemyFrame(img, enemyType, frame)
		}
	}

	// Save
	os.MkdirAll("assets/enemies", 0755)
	f, _ := os.Create("assets/enemies/enemy_spritesheet.png")
	defer f.Close()
	png.Encode(f, img)
}

func drawEnemyFrame(img *image.RGBA, enemyType, frame int) {
	offsetX := frame * enemyFrameWidth
	offsetY := enemyType * enemyFrameHeight

	switch enemyType {
	case EnemyBug:
		drawBug(img, offsetX, offsetY, frame)
	case EnemyError:
		drawErrorText(img, offsetX, offsetY, frame)
	case EnemyLowContext:
		drawLowContextText(img, offsetX, offsetY, frame)
	}
}

// Bug colors
var (
	bugBody   = C{0x8B, 0x45, 0x13, 0xFF} // Brown
	bugShell  = C{0x22, 0x8B, 0x22, 0xFF} // Green
	bugShellH = C{0x32, 0xCD, 0x32, 0xFF} // Light green highlight
	bugLegs   = C{0x2F, 0x1F, 0x0F, 0xFF} // Dark brown
	bugEye    = C{0xFF, 0x00, 0x00, 0xFF} // Red angry eyes
	bugEyeW   = C{0xFF, 0xFF, 0xFF, 0xFF} // White
)

func drawBug(img *image.RGBA, ox, oy, frame int) {
	// Wobble animation
	wobble := []int{0, 1, 0, -1}[frame]
	legFrame := frame % 2

	// Center the bug in the frame
	cx := ox + 10
	cy := oy + 4 + wobble

	// Body (oval shape) - 12x8 pixels
	// Main shell
	for y := 0; y < 8; y++ {
		for x := 0; x < 12; x++ {
			// Create oval shape
			dx := float64(x) - 5.5
			dy := float64(y) - 3.5
			if dx*dx/36 + dy*dy/16 < 1 {
				c := bugShell
				if y < 3 {
					c = bugShellH // Highlight on top
				}
				img.Set(cx+x, cy+y, c)
			}
		}
	}

	// Shell line down middle
	for y := 1; y < 7; y++ {
		img.Set(cx+5, cy+y, bugBody)
		img.Set(cx+6, cy+y, bugBody)
	}

	// Head (front of bug)
	img.Set(cx+11, cy+3, bugBody)
	img.Set(cx+11, cy+4, bugBody)
	img.Set(cx+12, cy+3, bugBody)
	img.Set(cx+12, cy+4, bugBody)

	// Angry eyes
	img.Set(cx+12, cy+2, bugEyeW)
	img.Set(cx+13, cy+2, bugEye)
	img.Set(cx+12, cy+5, bugEyeW)
	img.Set(cx+13, cy+5, bugEye)

	// Antennae
	img.Set(cx+13, cy+1, bugLegs)
	img.Set(cx+14, cy+0, bugLegs)
	img.Set(cx+13, cy+6, bugLegs)
	img.Set(cx+14, cy+7, bugLegs)

	// Legs (3 on each side, animated)
	legOffsets := []int{0, 1}[legFrame]
	// Top legs
	img.Set(cx+2, cy-1+legOffsets, bugLegs)
	img.Set(cx+1, cy-2+legOffsets, bugLegs)
	img.Set(cx+5, cy-1+legOffsets, bugLegs)
	img.Set(cx+4, cy-2+legOffsets, bugLegs)
	img.Set(cx+8, cy-1+legOffsets, bugLegs)
	img.Set(cx+7, cy-2+legOffsets, bugLegs)
	// Bottom legs
	img.Set(cx+2, cy+8-legOffsets, bugLegs)
	img.Set(cx+1, cy+9-legOffsets, bugLegs)
	img.Set(cx+5, cy+8-legOffsets, bugLegs)
	img.Set(cx+4, cy+9-legOffsets, bugLegs)
	img.Set(cx+8, cy+8-legOffsets, bugLegs)
	img.Set(cx+7, cy+9-legOffsets, bugLegs)
}

// Error text colors
var (
	errorRed    = C{0xFF, 0x33, 0x33, 0xFF} // Bright red
	errorDark   = C{0xAA, 0x00, 0x00, 0xFF} // Dark red shadow
	errorOrange = C{0xFF, 0x66, 0x00, 0xFF} // Orange highlight
)

func drawErrorText(img *image.RGBA, ox, oy, frame int) {
	// Shake animation
	shake := []int{0, 1, -1, 0}[frame]

	cx := ox + 2
	cy := oy + 4 + shake

	// Draw "ERROR" in pixel font (5 letters, ~5px each = 25px + spacing)
	// E
	drawPixelLetter(img, cx, cy, 'E', errorRed, errorDark)
	// R
	drawPixelLetter(img, cx+6, cy, 'R', errorRed, errorDark)
	// R
	drawPixelLetter(img, cx+12, cy, 'R', errorRed, errorDark)
	// O
	drawPixelLetter(img, cx+18, cy, 'O', errorRed, errorDark)
	// R
	drawPixelLetter(img, cx+24, cy, 'R', errorRed, errorDark)

	// Angry eyebrows above (on frames 1 and 3)
	if frame == 1 || frame == 3 {
		img.Set(cx+8, cy-2, errorOrange)
		img.Set(cx+9, cy-1, errorOrange)
		img.Set(cx+18, cy-2, errorOrange)
		img.Set(cx+17, cy-1, errorOrange)
	}
}

var (
	lowCtxYellow = C{0xFF, 0xCC, 0x00, 0xFF} // Warning yellow
	lowCtxOrange = C{0xFF, 0x99, 0x00, 0xFF} // Orange
	lowCtxDark   = C{0xAA, 0x66, 0x00, 0xFF} // Dark shadow
)

func drawLowContextText(img *image.RGBA, ox, oy, frame int) {
	// Pulse animation (scale effect via color)
	pulse := frame % 2
	mainColor := lowCtxYellow
	if pulse == 1 {
		mainColor = lowCtxOrange
	}

	cx := ox + 1
	cy := oy + 5

	// Draw "LOW" smaller
	drawSmallLetter(img, cx, cy, 'L', mainColor, lowCtxDark)
	drawSmallLetter(img, cx+4, cy, 'O', mainColor, lowCtxDark)
	drawSmallLetter(img, cx+8, cy, 'W', mainColor, lowCtxDark)

	// Draw "CTX" below or next to it
	drawSmallLetter(img, cx+14, cy, 'C', mainColor, lowCtxDark)
	drawSmallLetter(img, cx+18, cy, 'T', mainColor, lowCtxDark)
	drawSmallLetter(img, cx+22, cy, 'X', mainColor, lowCtxDark)

	// Warning triangle above
	if frame < 2 {
		// Small triangle
		img.Set(cx+13, cy-4, mainColor)
		img.Set(cx+12, cy-3, mainColor)
		img.Set(cx+13, cy-3, lowCtxDark)
		img.Set(cx+14, cy-3, mainColor)
		img.Set(cx+11, cy-2, mainColor)
		img.Set(cx+12, cy-2, mainColor)
		img.Set(cx+13, cy-2, mainColor)
		img.Set(cx+14, cy-2, mainColor)
		img.Set(cx+15, cy-2, mainColor)
	}
}

// Simple 5x7 pixel font for main letters
func drawPixelLetter(img *image.RGBA, ox, oy int, letter rune, main, shadow C) {
	// Draw shadow first (offset by 1,1)
	drawLetterPixels(img, ox+1, oy+1, letter, shadow)
	// Draw main
	drawLetterPixels(img, ox, oy, letter, main)
}

func drawLetterPixels(img *image.RGBA, ox, oy int, letter rune, c C) {
	// 5x7 pixel font patterns
	patterns := map[rune][]string{
		'E': {
			"#####",
			"#    ",
			"#    ",
			"#### ",
			"#    ",
			"#    ",
			"#####",
		},
		'R': {
			"#### ",
			"#   #",
			"#   #",
			"#### ",
			"#  # ",
			"#   #",
			"#   #",
		},
		'O': {
			" ### ",
			"#   #",
			"#   #",
			"#   #",
			"#   #",
			"#   #",
			" ### ",
		},
		'L': {
			"#    ",
			"#    ",
			"#    ",
			"#    ",
			"#    ",
			"#    ",
			"#####",
		},
		'W': {
			"#   #",
			"#   #",
			"#   #",
			"# # #",
			"# # #",
			"## ##",
			"#   #",
		},
		'C': {
			" ####",
			"#    ",
			"#    ",
			"#    ",
			"#    ",
			"#    ",
			" ####",
		},
		'T': {
			"#####",
			"  #  ",
			"  #  ",
			"  #  ",
			"  #  ",
			"  #  ",
			"  #  ",
		},
		'X': {
			"#   #",
			"#   #",
			" # # ",
			"  #  ",
			" # # ",
			"#   #",
			"#   #",
		},
	}

	pattern, ok := patterns[letter]
	if !ok {
		return
	}

	for y, row := range pattern {
		for x, ch := range row {
			if ch == '#' {
				img.Set(ox+x, oy+y, c)
			}
		}
	}
}

// Smaller 3x5 pixel font
func drawSmallLetter(img *image.RGBA, ox, oy int, letter rune, main, shadow C) {
	drawSmallLetterPixels(img, ox+1, oy+1, letter, shadow)
	drawSmallLetterPixels(img, ox, oy, letter, main)
}

func drawSmallLetterPixels(img *image.RGBA, ox, oy int, letter rune, c C) {
	patterns := map[rune][]string{
		'L': {
			"#  ",
			"#  ",
			"#  ",
			"#  ",
			"###",
		},
		'O': {
			"###",
			"# #",
			"# #",
			"# #",
			"###",
		},
		'W': {
			"# #",
			"# #",
			"###",
			"###",
			"# #",
		},
		'C': {
			"###",
			"#  ",
			"#  ",
			"#  ",
			"###",
		},
		'T': {
			"###",
			" # ",
			" # ",
			" # ",
			" # ",
		},
		'X': {
			"# #",
			"# #",
			" # ",
			"# #",
			"# #",
		},
	}

	pattern, ok := patterns[letter]
	if !ok {
		return
	}

	for y, row := range pattern {
		for x, ch := range row {
			if ch == '#' {
				img.Set(ox+x, oy+y, c)
			}
		}
	}
}
