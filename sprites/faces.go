package sprites

import (
	"image"
	"image/color"
)

// Claude's official color palette
var (
	PrimaryOrange   = color.RGBA{0xFF, 0x99, 0x33, 0xFF}
	ShadowOrange    = color.RGBA{0xCC, 0x66, 0x00, 0xFF}
	HighlightOrange = color.RGBA{0xFF, 0xBB, 0x77, 0xFF}
	Outline         = color.RGBA{0x22, 0x22, 0x22, 0xFF}
	EyeColor        = color.RGBA{0x22, 0x22, 0x22, 0xFF}
	MouthColor      = color.RGBA{0x44, 0x22, 0x00, 0xFF}
	White           = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	Clear           = color.RGBA{0x00, 0x00, 0x00, 0x00}
	TerminalGreen   = color.RGBA{0x00, 0xFF, 0x88, 0xFF}
	SparkYellow     = color.RGBA{0xFF, 0xF5, 0x96, 0xFF}
	SparkBright     = color.RGBA{0xFF, 0xFF, 0xC8, 0xFF}
	SparkDark       = color.RGBA{0xFF, 0xDC, 0x50, 0xFF}
)

// EyeState represents different eye expressions
type EyeState int

const (
	EyeOpen EyeState = iota
	EyeHalfClosed
	EyeClosed
	EyeSquint  // ">" shape
	EyeHappy   // "^" shape
	EyeX       // X_X hurt
)

// MouthState represents different mouth shapes
type MouthState int

const (
	MouthNone MouthState = iota
	MouthSmile
	MouthOpen    // "o" shape
	MouthWide    // Yawning/surprised
	MouthLine    // Neutral "-"
)

// GenerateEyeTexture creates a 3x4 eye texture for the given state
func GenerateEyeTexture(state EyeState) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 3, 4))

	// Fill transparent
	for y := 0; y < 4; y++ {
		for x := 0; x < 3; x++ {
			img.Set(x, y, Clear)
		}
	}

	switch state {
	case EyeOpen:
		// Full 3x4 dark rectangle
		for y := 0; y < 4; y++ {
			for x := 0; x < 3; x++ {
				img.Set(x, y, EyeColor)
			}
		}

	case EyeHalfClosed:
		// Bottom 2 rows only
		for y := 2; y < 4; y++ {
			for x := 0; x < 3; x++ {
				img.Set(x, y, EyeColor)
			}
		}

	case EyeClosed:
		// Single horizontal line in middle
		for x := 0; x < 3; x++ {
			img.Set(x, 2, EyeColor)
		}

	case EyeSquint:
		// ">" shape
		img.Set(0, 0, EyeColor)
		img.Set(1, 0, EyeColor)
		img.Set(1, 1, EyeColor)
		img.Set(2, 1, EyeColor)
		img.Set(1, 2, EyeColor)
		img.Set(2, 2, EyeColor)
		img.Set(0, 3, EyeColor)
		img.Set(1, 3, EyeColor)

	case EyeHappy:
		// "^" shape (upside down V)
		img.Set(0, 2, EyeColor)
		img.Set(1, 1, EyeColor)
		img.Set(2, 2, EyeColor)
		img.Set(0, 3, EyeColor)
		img.Set(1, 2, EyeColor)
		img.Set(2, 3, EyeColor)

	case EyeX:
		// X shape
		img.Set(0, 0, EyeColor)
		img.Set(2, 0, EyeColor)
		img.Set(1, 1, EyeColor)
		img.Set(1, 2, EyeColor)
		img.Set(0, 3, EyeColor)
		img.Set(2, 3, EyeColor)
	}

	return img
}

// GenerateMouthTexture creates a mouth texture for the given state
func GenerateMouthTexture(state MouthState) *image.RGBA {
	var width, height int

	switch state {
	case MouthNone:
		return nil
	case MouthSmile:
		width, height = 5, 2
	case MouthOpen:
		width, height = 3, 3
	case MouthWide:
		width, height = 5, 4
	case MouthLine:
		width, height = 3, 1
	default:
		return nil
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill transparent
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, Clear)
		}
	}

	switch state {
	case MouthSmile:
		// Curved smile
		img.Set(1, 0, MouthColor)
		img.Set(2, 0, MouthColor)
		img.Set(3, 0, MouthColor)
		img.Set(0, 1, MouthColor)
		img.Set(4, 1, MouthColor)

	case MouthOpen:
		// Hollow "o"
		img.Set(1, 0, MouthColor)
		img.Set(0, 1, MouthColor)
		img.Set(2, 1, MouthColor)
		img.Set(1, 2, MouthColor)

	case MouthWide:
		// Filled oval for yawn/surprise
		img.Set(1, 0, MouthColor)
		img.Set(2, 0, MouthColor)
		img.Set(3, 0, MouthColor)
		for x := 0; x < 5; x++ {
			img.Set(x, 1, MouthColor)
			img.Set(x, 2, MouthColor)
		}
		img.Set(1, 3, MouthColor)
		img.Set(2, 3, MouthColor)
		img.Set(3, 3, MouthColor)

	case MouthLine:
		// Simple horizontal line
		for x := 0; x < 3; x++ {
			img.Set(x, 0, MouthColor)
		}
	}

	return img
}

// BlinkFrames returns the eye states for a blink animation
// Returns: open -> half -> closed -> half -> open
func BlinkFrames() []EyeState {
	return []EyeState{
		EyeOpen,
		EyeHalfClosed,
		EyeClosed,
		EyeHalfClosed,
		EyeOpen,
	}
}

// GenerateEffectTexture creates various effect sprites
type EffectType int

const (
	EffectHeart EffectType = iota
	EffectZzz
	EffectLightbulb
	EffectQuestion
	EffectSparkSmall
	EffectSparkMedium
	EffectSparkLarge
	EffectThoughtDot
)

func GenerateEffectTexture(effect EffectType) *image.RGBA {
	switch effect {
	case EffectHeart:
		return generateHeart()
	case EffectZzz:
		return generateZzz()
	case EffectLightbulb:
		return generateLightbulb()
	case EffectQuestion:
		return generateQuestion()
	case EffectSparkSmall:
		return generateSpark(2)
	case EffectSparkMedium:
		return generateSpark(3)
	case EffectSparkLarge:
		return generateSpark(4)
	case EffectThoughtDot:
		return generateThoughtDot()
	}
	return nil
}

func generateHeart() *image.RGBA {
	// 7x7 heart
	img := image.NewRGBA(image.Rect(0, 0, 7, 7))
	h := HighlightOrange
	p := PrimaryOrange
	s := ShadowOrange
	o := Outline

	pattern := [][]color.RGBA{
		{Clear, o, o, Clear, o, o, Clear},
		{o, h, p, o, h, p, o},
		{o, p, p, p, p, s, o},
		{o, p, p, p, p, s, o},
		{Clear, o, p, p, s, o, Clear},
		{Clear, Clear, o, p, o, Clear, Clear},
		{Clear, Clear, Clear, o, Clear, Clear, Clear},
	}

	for y, row := range pattern {
		for x, c := range row {
			img.Set(x, y, c)
		}
	}
	return img
}

func generateZzz() *image.RGBA {
	// Simple Z
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, Clear)
		}
	}

	// Z shape with outline
	z := PrimaryOrange
	img.Set(1, 0, z)
	img.Set(2, 0, z)
	img.Set(3, 0, z)
	img.Set(3, 1, z)
	img.Set(2, 2, z)
	img.Set(1, 3, z)
	img.Set(1, 4, z)
	img.Set(2, 4, z)
	img.Set(3, 4, z)

	return img
}

func generateLightbulb() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 7, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 7; x++ {
			img.Set(x, y, Clear)
		}
	}

	o := Outline
	g := SparkBright  // glow
	h := SparkYellow  // highlight
	m := SparkDark    // main
	b := color.RGBA{140, 140, 140, 255} // base

	pattern := [][]color.RGBA{
		{Clear, Clear, o, o, o, Clear, Clear},
		{Clear, o, g, g, h, o, Clear},
		{o, g, h, m, h, m, o},
		{o, h, h, m, m, m, o},
		{o, h, m, m, m, m, o},
		{Clear, o, m, m, m, o, Clear},
		{Clear, Clear, o, m, o, Clear, Clear},
		{Clear, Clear, o, b, o, Clear, Clear},
		{Clear, Clear, o, b, o, Clear, Clear},
		{Clear, Clear, Clear, o, Clear, Clear, Clear},
	}

	for y, row := range pattern {
		for x, c := range row {
			img.Set(x, y, c)
		}
	}
	return img
}

func generateQuestion() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 5, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, Clear)
		}
	}

	o := Outline
	w := White

	pattern := [][]color.RGBA{
		{Clear, o, o, o, Clear},
		{o, w, w, w, o},
		{o, Clear, o, w, o},
		{Clear, Clear, o, w, o},
		{Clear, o, w, o, Clear},
		{Clear, o, w, o, Clear},
		{Clear, Clear, Clear, Clear, Clear},
		{Clear, o, w, o, Clear},
	}

	for y, row := range pattern {
		for x, c := range row {
			img.Set(x, y, c)
		}
	}
	return img
}

func generateSpark(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, Clear)
		}
	}

	// Diamond/cross pattern
	mid := size / 2
	img.Set(mid, 0, SparkBright)
	img.Set(0, mid, SparkYellow)
	img.Set(size-1, mid, SparkYellow)
	img.Set(mid, size-1, SparkDark)
	if size > 2 {
		img.Set(mid, mid, White)
	}

	return img
}

func generateThoughtDot() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			img.Set(x, y, Clear)
		}
	}

	// Small white dot
	img.Set(1, 0, White)
	img.Set(0, 1, White)
	img.Set(1, 1, White)
	img.Set(2, 1, White)
	img.Set(1, 2, White)

	return img
}
