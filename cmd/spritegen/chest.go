package main

import (
	"image"
	"image/png"
	"os"
)

// Treasure chest sprite sheet generator
// 32x32 pixels per frame, 6 frames: closed, wobble, opening1, opening2, open, open_glow

const (
	chestFrameWidth  = 32
	chestFrameHeight = 32
	chestNumFrames   = 6
)

// Chest frame indices
const (
	ChestFrameClosed = iota
	ChestFrameWobble1
	ChestFrameOpening1
	ChestFrameOpening2
	ChestFrameOpen
	ChestFrameOpenGlow
)

// Chest colors
var (
	chestWood     = C{0x8B, 0x45, 0x13, 0xFF} // Brown wood
	chestWoodDark = C{0x5D, 0x2E, 0x0C, 0xFF} // Dark wood shadow
	chestWoodLite = C{0xA8, 0x5C, 0x2B, 0xFF} // Light wood highlight
	chestMetal    = C{0xCD, 0xA5, 0x32, 0xFF} // Gold metal
	chestMetalLt  = C{0xFF, 0xD7, 0x00, 0xFF} // Gold highlight
	chestMetalDk  = C{0x8B, 0x73, 0x00, 0xFF} // Gold shadow
	chestLock     = C{0xB0, 0x80, 0x00, 0xFF} // Lock gold
	chestGlow     = C{0xFF, 0xF0, 0x80, 0xFF} // Inner glow
	chestSparkle  = C{0xFF, 0xFF, 0xFF, 0xFF} // White sparkle
)

func generateChest() {
	width := chestFrameWidth * chestNumFrames
	height := chestFrameHeight
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill transparent
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, X)
		}
	}

	// Generate each frame
	for frame := 0; frame < chestNumFrames; frame++ {
		drawChestFrame(img, frame)
	}

	// Save
	os.MkdirAll("assets/ui", 0755)
	f, _ := os.Create("assets/ui/chest.png")
	defer f.Close()
	png.Encode(f, img)
}

func drawChestFrame(img *image.RGBA, frame int) {
	ox := frame * chestFrameWidth
	oy := 0

	switch frame {
	case ChestFrameClosed:
		drawChestClosed(img, ox, oy, 0)
	case ChestFrameWobble1:
		drawChestClosed(img, ox, oy, 1) // Slight wobble
	case ChestFrameOpening1:
		drawChestOpening(img, ox, oy, 0.3)
	case ChestFrameOpening2:
		drawChestOpening(img, ox, oy, 0.6)
	case ChestFrameOpen:
		drawChestOpen(img, ox, oy, false)
	case ChestFrameOpenGlow:
		drawChestOpen(img, ox, oy, true)
	}
}

// drawChestClosed draws the closed treasure chest
func drawChestClosed(img *image.RGBA, ox, oy, wobble int) {
	// Chest dimensions: 24x20 centered in 32x32
	cx := ox + 4
	cy := oy + 10 + wobble

	// Base (bottom box) - 24x10
	for y := 10; y < 20; y++ {
		for x := 0; x < 24; x++ {
			c := chestWood
			if x < 2 || y > 17 {
				c = chestWoodDark
			} else if x > 21 || y < 12 {
				c = chestWoodLite
			}
			img.Set(cx+x, cy+y, c)
		}
	}

	// Lid (rounded top) - 24x8 with curve
	for y := 2; y < 10; y++ {
		// Width varies for rounded top
		inset := 0
		if y < 4 {
			inset = 4 - y
		}
		for x := inset; x < 24-inset; x++ {
			c := chestWood
			if x < inset+2 || y > 7 {
				c = chestWoodDark
			} else if x > 22-inset || y < 4 {
				c = chestWoodLite
			}
			img.Set(cx+x, cy+y, c)
		}
	}

	// Metal bands (horizontal stripes)
	for x := 1; x < 23; x++ {
		c := chestMetal
		if x < 3 {
			c = chestMetalDk
		} else if x > 20 {
			c = chestMetalLt
		}
		img.Set(cx+x, cy+6, c)  // Top band on lid
		img.Set(cx+x, cy+11, c) // Middle band
		img.Set(cx+x, cy+17, c) // Bottom band
	}

	// Lock (center)
	for dy := 0; dy < 4; dy++ {
		for dx := 0; dx < 4; dx++ {
			c := chestMetal
			if dx == 0 || dy == 3 {
				c = chestMetalDk
			} else if dx == 3 || dy == 0 {
				c = chestMetalLt
			}
			img.Set(cx+10+dx, cy+8+dy, c)
		}
	}
	// Lock keyhole
	img.Set(cx+11, cy+9, chestWoodDark)
	img.Set(cx+12, cy+9, chestWoodDark)
	img.Set(cx+11, cy+10, chestWoodDark)
	img.Set(cx+12, cy+10, chestWoodDark)

	// Corner rivets
	img.Set(cx+2, cy+4, chestMetalLt)
	img.Set(cx+21, cy+4, chestMetalLt)
	img.Set(cx+2, cy+18, chestMetalLt)
	img.Set(cx+21, cy+18, chestMetalLt)
}

// drawChestOpening draws the chest mid-opening
func drawChestOpening(img *image.RGBA, ox, oy int, openAmount float32) {
	cx := ox + 4
	cy := oy + 10

	// Base stays fixed
	for y := 10; y < 20; y++ {
		for x := 0; x < 24; x++ {
			c := chestWood
			if x < 2 || y > 17 {
				c = chestWoodDark
			} else if x > 21 || y < 12 {
				c = chestWoodLite
			}
			img.Set(cx+x, cy+y, c)
		}
	}

	// Lid tilts back - hinge is at back of chest (y=10)
	// As it opens, lid rotates back, getting shorter in front view
	lidHeight := int(8 - openAmount*4) // Gets shorter as it tilts back
	if lidHeight < 3 {
		lidHeight = 3
	}
	lidY := cy + 10 - lidHeight // Lid bottom connects to base top

	for y := 0; y < lidHeight; y++ {
		// Narrower at top for rounded look
		inset := 0
		if y < 2 {
			inset = 2 - y
		}
		for x := inset; x < 24-inset; x++ {
			c := chestWood
			if x < inset+2 {
				c = chestWoodDark
			} else if x > 22-inset {
				c = chestWoodLite
			} else if y == 0 {
				c = chestWoodLite // Top highlight
			}
			img.Set(cx+x, lidY+y, c)
		}
	}

	// Metal band on lid (if lid is tall enough)
	if lidHeight > 4 {
		bandY := lidY + lidHeight/2
		for x := 1; x < 23; x++ {
			c := chestMetal
			if x < 3 {
				c = chestMetalDk
			} else if x > 20 {
				c = chestMetalLt
			}
			img.Set(cx+x, bandY, c)
		}
	}

	// Glow from inside (visible gap between lid and base)
	glowHeight := int(openAmount * 3)
	for dy := 0; dy < glowHeight; dy++ {
		for x := 3; x < 21; x++ {
			img.Set(cx+x, cy+10-dy, chestGlow)
		}
	}

	// Metal bands on base
	for x := 1; x < 23; x++ {
		c := chestMetal
		if x < 3 {
			c = chestMetalDk
		} else if x > 20 {
			c = chestMetalLt
		}
		img.Set(cx+x, cy+11, c)
		img.Set(cx+x, cy+17, c)
	}

	// Corner rivets on base
	img.Set(cx+2, cy+18, chestMetalLt)
	img.Set(cx+21, cy+18, chestMetalLt)
}

// drawChestOpen draws the fully open chest
func drawChestOpen(img *image.RGBA, ox, oy int, glow bool) {
	cx := ox + 4
	cy := oy + 10

	// Base
	for y := 10; y < 20; y++ {
		for x := 0; x < 24; x++ {
			c := chestWood
			if x < 2 || y > 17 {
				c = chestWoodDark
			} else if x > 21 || y < 12 {
				c = chestWoodLite
			}
			img.Set(cx+x, cy+y, c)
		}
	}

	// Lid fully open (tilted back, connects at base top y=10)
	// Lid is now just 3px tall (viewed from front when tilted back 90 degrees)
	lidY := cy + 10 - 3 // Connects to base top
	for y := 0; y < 3; y++ {
		for x := 2; x < 22; x++ {
			c := chestWood
			if y == 2 {
				c = chestWoodDark // Bottom edge
			} else if y == 0 {
				c = chestWoodLite // Top edge
			}
			img.Set(cx+x, lidY+y, c)
		}
	}

	// Metal band on open lid
	for x := 3; x < 21; x++ {
		img.Set(cx+x, lidY+1, chestMetal)
	}

	// Inside the chest (visible opening with glow)
	for y := 0; y < 4; y++ {
		for x := 3; x < 21; x++ {
			if glow {
				img.Set(cx+x, cy+10+y, chestGlow)
			} else {
				img.Set(cx+x, cy+10+y, chestWoodDark)
			}
		}
	}

	// Metal bands on base
	for x := 1; x < 23; x++ {
		c := chestMetal
		if x < 3 {
			c = chestMetalDk
		} else if x > 20 {
			c = chestMetalLt
		}
		img.Set(cx+x, cy+14, c)
		img.Set(cx+x, cy+17, c)
	}

	// Sparkles when glowing
	if glow {
		img.Set(cx+6, cy+11, chestSparkle)
		img.Set(cx+12, cy+10, chestSparkle)
		img.Set(cx+17, cy+11, chestSparkle)
		img.Set(cx+10, cy+12, chestSparkle)
		img.Set(cx+14, cy+12, chestSparkle)

		// Rays of light above chest
		img.Set(cx+8, cy+7, chestGlow)
		img.Set(cx+12, cy+5, chestGlow)
		img.Set(cx+16, cy+7, chestGlow)
	}

	// Corner rivets
	img.Set(cx+2, cy+18, chestMetalLt)
	img.Set(cx+21, cy+18, chestMetalLt)
}
