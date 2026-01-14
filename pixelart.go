package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

// PixelColor represents a single pixel with RGBA components
type PixelColor struct {
	R, G, B, A uint8
}

// Clear is a fully transparent pixel
var Clear = PixelColor{0, 0, 0, 0}

// NewPixel creates a new opaque pixel color
func NewPixel(r, g, b uint8) PixelColor {
	return PixelColor{r, g, b, 255}
}

// NewPixelAlpha creates a new pixel color with alpha
func NewPixelAlpha(r, g, b, a uint8) PixelColor {
	return PixelColor{r, g, b, a}
}

// ClaudePalette contains the official Claude color palette
var ClaudePalette = struct {
	PrimaryOrange   PixelColor
	ShadowOrange    PixelColor
	HighlightOrange PixelColor
	Outline         PixelColor
	EyeColor        PixelColor
	MouthColor      PixelColor
	White           PixelColor
	SparkYellow     PixelColor
	SparkBright     PixelColor
	SparkDark       PixelColor
	TerminalGreen   PixelColor
}{
	PrimaryOrange:   NewPixel(0xFF, 0x99, 0x33), // #FF9933
	ShadowOrange:    NewPixel(0xCC, 0x66, 0x00), // #CC6600
	HighlightOrange: NewPixel(0xFF, 0xBB, 0x77), // #FFBB77
	Outline:         NewPixel(0x22, 0x22, 0x22), // #222222
	EyeColor:        NewPixel(0x22, 0x22, 0x22), // #222222
	MouthColor:      NewPixel(0x44, 0x22, 0x00), // #442200
	White:           NewPixel(0xFF, 0xFF, 0xFF), // #FFFFFF
	SparkYellow:     NewPixel(0xFF, 0xF5, 0x96), // #FFF596
	SparkBright:     NewPixel(0xFF, 0xFF, 0xC8), // #FFFFC8
	SparkDark:       NewPixel(0xFF, 0xDC, 0x50), // #FFDC50
	TerminalGreen:   NewPixel(0x00, 0xFF, 0x88), // #00FF88
}

// Shorthand aliases
var (
	P = ClaudePalette.PrimaryOrange
	S = ClaudePalette.ShadowOrange
	H = ClaudePalette.HighlightOrange
	O = ClaudePalette.Outline
	E = ClaudePalette.EyeColor
	M = ClaudePalette.MouthColor
	W = ClaudePalette.White
	_ = Clear // X for clear in patterns
)

// ImageFromPixels creates an image from a 2D array of pixel colors
// pixels[row][col] where row 0 is the TOP of the image
func ImageFromPixels(pixels [][]PixelColor) *image.RGBA {
	if len(pixels) == 0 || len(pixels[0]) == 0 {
		return nil
	}

	height := len(pixels)
	width := len(pixels[0])

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := pixels[y][x]
			img.Set(x, y, color.RGBA{p.R, p.G, p.B, p.A})
		}
	}

	return img
}

// SavePNG saves an image to a PNG file
func SavePNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// CreateSpriteSheet creates a sprite sheet from multiple animation frames
// Each row is one animation, each column is one frame
// animations[animIndex][frameIndex] = 2D pixel array for that frame
func CreateSpriteSheet(animations [][][][]PixelColor, frameWidth, frameHeight, maxFrames int) *image.RGBA {
	numAnims := len(animations)
	width := frameWidth * maxFrames
	height := frameHeight * numAnims

	sheet := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with transparent
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sheet.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}

	// Draw each animation frame
	for animIdx, frames := range animations {
		for frameIdx, frame := range frames {
			offsetX := frameIdx * frameWidth
			offsetY := animIdx * frameHeight

			if len(frame) == 0 {
				continue
			}
			frameImg := ImageFromPixels(frame)

			// Copy frame to sheet
			for y := 0; y < frameHeight && y < frameImg.Bounds().Dy(); y++ {
				for x := 0; x < frameWidth && x < frameImg.Bounds().Dx(); x++ {
					sheet.Set(offsetX+x, offsetY+y, frameImg.At(x, y))
				}
			}
		}
	}

	return sheet
}
