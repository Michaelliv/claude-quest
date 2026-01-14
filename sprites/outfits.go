package sprites

import (
	"image"
	"image/color"
)

// OutfitType represents different wearable categories
type OutfitType int

const (
	OutfitHat OutfitType = iota
	OutfitCape
	OutfitHeldItem
)

// HatType represents different hat options
type HatType int

const (
	HatNone HatType = iota
	HatWizard
	HatParty
	HatCrown
	HatTopHat
	HatPropeller
	HatHalo
)

// GenerateHatTexture creates a hat sprite
func GenerateHatTexture(hat HatType) *image.RGBA {
	switch hat {
	case HatWizard:
		return generateWizardHat()
	case HatParty:
		return generatePartyHat()
	case HatCrown:
		return generateCrown()
	case HatTopHat:
		return generateTopHat()
	case HatPropeller:
		return generatePropellerHat()
	case HatHalo:
		return generateHalo()
	default:
		return nil
	}
}

func generateWizardHat() *image.RGBA {
	// 16x12 wizard hat
	img := image.NewRGBA(image.Rect(0, 0, 16, 12))
	clearImage(img)

	// Deep purple colors
	main := color.RGBA{80, 50, 120, 255}
	shadow := color.RGBA{50, 30, 80, 255}
	highlight := color.RGBA{120, 80, 160, 255}
	star := SparkYellow
	o := Outline

	// Pointy wizard hat shape
	// Tip
	img.Set(8, 0, o)
	img.Set(7, 1, o)
	img.Set(8, 1, highlight)
	img.Set(9, 1, o)

	// Upper cone
	for y := 2; y < 6; y++ {
		width := y + 1
		left := 8 - width/2
		for x := left; x < left+width; x++ {
			if x == left || x == left+width-1 {
				img.Set(x, y, o)
			} else if x < 8 {
				img.Set(x, y, shadow)
			} else {
				img.Set(x, y, main)
			}
		}
	}

	// Brim
	for x := 2; x < 14; x++ {
		img.Set(x, 10, o)
		if x > 2 && x < 13 {
			img.Set(x, 9, main)
			img.Set(x, 11, shadow)
		}
	}

	// Star decoration
	img.Set(10, 4, star)
	img.Set(9, 5, star)
	img.Set(10, 5, White)
	img.Set(11, 5, star)
	img.Set(10, 6, star)

	return img
}

func generatePartyHat() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 12, 10))
	clearImage(img)

	purple := color.RGBA{150, 80, 180, 255}
	purpleDark := color.RGBA{110, 50, 140, 255}
	gold := color.RGBA{255, 215, 80, 255}
	goldDark := color.RGBA{220, 170, 40, 255}
	pompom := color.RGBA{255, 230, 100, 255}
	o := Outline

	// Pompom at top
	img.Set(6, 0, pompom)
	img.Set(5, 1, pompom)
	img.Set(6, 1, pompom)
	img.Set(7, 1, pompom)

	// Striped cone
	for y := 2; y < 9; y++ {
		width := (y - 1) * 2
		if width < 2 {
			width = 2
		}
		left := 6 - width/2

		isGold := (y % 2) == 0

		for x := left; x < left+width && x < 12; x++ {
			if x == left || x == left+width-1 {
				img.Set(x, y, o)
			} else if isGold {
				if x < 6 {
					img.Set(x, y, goldDark)
				} else {
					img.Set(x, y, gold)
				}
			} else {
				if x < 6 {
					img.Set(x, y, purpleDark)
				} else {
					img.Set(x, y, purple)
				}
			}
		}
	}

	// Brim outline
	for x := 1; x < 11; x++ {
		img.Set(x, 9, o)
	}

	return img
}

func generateCrown() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 14, 8))
	clearImage(img)

	gold := color.RGBA{255, 215, 80, 255}
	goldDark := color.RGBA{200, 160, 40, 255}
	goldBright := color.RGBA{255, 240, 150, 255}
	gem := color.RGBA{220, 50, 50, 255}
	o := Outline

	// Crown spikes
	spikes := []int{2, 5, 7, 9, 12}
	for _, x := range spikes {
		img.Set(x, 0, o)
		img.Set(x, 1, goldBright)
	}

	// Crown body
	for y := 2; y < 7; y++ {
		for x := 1; x < 13; x++ {
			if x == 1 || x == 12 {
				img.Set(x, y, o)
			} else if y == 2 {
				img.Set(x, y, goldBright)
			} else if y >= 5 {
				img.Set(x, y, goldDark)
			} else {
				img.Set(x, y, gold)
			}
		}
	}

	// Bottom outline
	for x := 1; x < 13; x++ {
		img.Set(x, 7, o)
	}

	// Center gem
	img.Set(6, 3, gem)
	img.Set(7, 3, gem)
	img.Set(6, 4, gem)
	img.Set(7, 4, gem)

	return img
}

func generateTopHat() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 16, 12))
	clearImage(img)

	main := color.RGBA{30, 30, 35, 255}
	highlight := color.RGBA{60, 60, 70, 255}
	band := color.RGBA{180, 50, 50, 255}
	o := Outline

	// Top of hat
	for x := 4; x < 12; x++ {
		img.Set(x, 0, o)
	}

	// Main cylinder
	for y := 1; y < 8; y++ {
		for x := 4; x < 12; x++ {
			if x == 4 || x == 11 {
				img.Set(x, y, o)
			} else if y == 6 {
				img.Set(x, y, band) // Red band
			} else if x < 7 {
				img.Set(x, y, main)
			} else {
				img.Set(x, y, highlight)
			}
		}
	}

	// Brim
	for x := 1; x < 15; x++ {
		img.Set(x, 8, o)
		if x > 1 && x < 14 {
			img.Set(x, 9, main)
			img.Set(x, 10, main)
		}
		img.Set(x, 11, o)
	}

	return img
}

func generatePropellerHat() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 14, 10))
	clearImage(img)

	red := color.RGBA{230, 70, 70, 255}
	blue := color.RGBA{70, 130, 230, 255}
	yellow := color.RGBA{255, 220, 80, 255}
	main := color.RGBA{70, 130, 200, 255}
	o := Outline

	// Propeller
	img.Set(3, 0, red)
	img.Set(4, 0, red)
	img.Set(5, 1, o)
	img.Set(6, 1, yellow) // center
	img.Set(7, 1, yellow)
	img.Set(8, 1, o)
	img.Set(9, 0, blue)
	img.Set(10, 0, blue)

	// Beanie cap
	for y := 2; y < 8; y++ {
		width := 6 + (y - 2)
		left := 7 - width/2
		for x := left; x < left+width && x < 14; x++ {
			if x == left || x == left+width-1 {
				img.Set(x, y, o)
			} else {
				img.Set(x, y, main)
			}
		}
	}

	// Stripes
	for x := 4; x < 10; x++ {
		if x%2 == 0 {
			img.Set(x, 5, red)
		}
	}

	return img
}

func generateHalo() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 12, 4))
	clearImage(img)

	inner := color.RGBA{255, 255, 200, 255}
	outer := color.RGBA{255, 240, 150, 255}
	glow := color.RGBA{255, 255, 220, 200}

	// Oval halo shape
	// Top edge
	for x := 3; x < 9; x++ {
		img.Set(x, 0, glow)
	}

	// Middle with hole
	img.Set(2, 1, outer)
	img.Set(3, 1, inner)
	img.Set(8, 1, inner)
	img.Set(9, 1, outer)

	img.Set(2, 2, outer)
	img.Set(3, 2, inner)
	img.Set(8, 2, inner)
	img.Set(9, 2, outer)

	// Bottom edge
	for x := 3; x < 9; x++ {
		img.Set(x, 3, glow)
	}

	return img
}

// CapeType represents different cape options
type CapeType int

const (
	CapeNone CapeType = iota
	CapeRed
	CapeBlue
	CapePurple
	CapeRainbow
)

// GenerateCapeTexture creates a cape sprite
func GenerateCapeTexture(cape CapeType) *image.RGBA {
	if cape == CapeNone {
		return nil
	}

	img := image.NewRGBA(image.Rect(0, 0, 20, 16))
	clearImage(img)

	var main, dark, light color.RGBA

	switch cape {
	case CapeRed:
		main = color.RGBA{200, 50, 50, 255}
		dark = color.RGBA{140, 30, 30, 255}
		light = color.RGBA{240, 80, 80, 255}
	case CapeBlue:
		main = color.RGBA{50, 100, 200, 255}
		dark = color.RGBA{30, 60, 140, 255}
		light = color.RGBA{80, 140, 240, 255}
	case CapePurple:
		main = color.RGBA{120, 60, 180, 255}
		dark = color.RGBA{80, 40, 120, 255}
		light = color.RGBA{160, 100, 220, 255}
	case CapeRainbow:
		// Special rainbow gradient handled below
		main = color.RGBA{255, 100, 100, 255}
		dark = color.RGBA{100, 100, 255, 255}
		light = color.RGBA{255, 255, 100, 255}
	}

	o := Outline

	// Cape shape - flowing behind
	for y := 0; y < 16; y++ {
		// Width increases as we go down
		width := 8 + y/2
		left := 10 - width/2

		for x := left; x < left+width && x < 20; x++ {
			if x == left || x == left+width-1 {
				img.Set(x, y, o)
			} else if cape == CapeRainbow {
				// Rainbow gradient
				rainbow := []color.RGBA{
					{255, 100, 100, 255}, // red
					{255, 180, 100, 255}, // orange
					{255, 255, 100, 255}, // yellow
					{100, 255, 100, 255}, // green
					{100, 180, 255, 255}, // blue
					{180, 100, 255, 255}, // purple
				}
				img.Set(x, y, rainbow[y%6])
			} else if x < 10 {
				img.Set(x, y, dark)
			} else if x == 10 {
				img.Set(x, y, main)
			} else {
				img.Set(x, y, light)
			}
		}
	}

	return img
}

// HeldItemType represents items Claude can hold
type HeldItemType int

const (
	HeldNone HeldItemType = iota
	HeldWand
	HeldSword
	HeldBook
	HeldCoffee
	HeldKeyboard
)

// GenerateHeldItemTexture creates a held item sprite
func GenerateHeldItemTexture(item HeldItemType) *image.RGBA {
	switch item {
	case HeldWand:
		return generateWand()
	case HeldSword:
		return generateSword()
	case HeldBook:
		return generateBook()
	case HeldCoffee:
		return generateCoffee()
	case HeldKeyboard:
		return generateKeyboard()
	default:
		return nil
	}
}

func generateWand() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 4, 12))
	clearImage(img)

	wood := color.RGBA{120, 80, 50, 255}
	woodDark := color.RGBA{80, 50, 30, 255}
	star := SparkYellow
	o := Outline

	// Star tip
	img.Set(2, 0, star)
	img.Set(1, 1, star)
	img.Set(2, 1, White)
	img.Set(3, 1, star)
	img.Set(2, 2, star)

	// Wand shaft
	for y := 3; y < 12; y++ {
		img.Set(1, y, woodDark)
		img.Set(2, y, wood)
	}

	return img
}

func generateSword() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 5, 14))
	clearImage(img)

	blade := color.RGBA{200, 200, 210, 255}
	bladeBright := color.RGBA{240, 240, 250, 255}
	hilt := color.RGBA{160, 140, 60, 255}
	grip := color.RGBA{80, 50, 30, 255}
	o := Outline

	// Blade tip
	img.Set(2, 0, bladeBright)
	img.Set(1, 1, blade)
	img.Set(2, 1, bladeBright)
	img.Set(3, 1, blade)

	// Blade body
	for y := 2; y < 9; y++ {
		img.Set(1, y, blade)
		img.Set(2, y, bladeBright)
		img.Set(3, y, blade)
	}

	// Hilt
	for x := 0; x < 5; x++ {
		img.Set(x, 9, hilt)
	}

	// Grip
	for y := 10; y < 14; y++ {
		img.Set(2, y, grip)
	}

	return img
}

func generateBook() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 8, 10))
	clearImage(img)

	cover := color.RGBA{140, 60, 60, 255}
	coverDark := color.RGBA{100, 40, 40, 255}
	pages := color.RGBA{240, 235, 220, 255}
	pagesDark := color.RGBA{200, 195, 180, 255}
	gold := color.RGBA{220, 180, 60, 255}

	// Spine
	for y := 0; y < 10; y++ {
		img.Set(0, y, coverDark)
	}

	// Back cover
	for y := 0; y < 10; y++ {
		for x := 1; x < 3; x++ {
			img.Set(x, y, cover)
		}
	}

	// Pages
	for y := 1; y < 9; y++ {
		for x := 3; x < 6; x++ {
			if y == 1 || y == 8 {
				img.Set(x, y, pagesDark)
			} else {
				img.Set(x, y, pages)
			}
		}
	}

	// Front cover
	for y := 0; y < 10; y++ {
		for x := 6; x < 8; x++ {
			img.Set(x, y, cover)
		}
	}

	// Gold decoration
	img.Set(7, 4, gold)
	img.Set(7, 5, gold)

	return img
}

func generateCoffee() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 6, 8))
	clearImage(img)

	cup := color.RGBA{240, 240, 245, 255}
	cupDark := color.RGBA{200, 200, 210, 255}
	coffee := color.RGBA{80, 50, 30, 255}
	steam := color.RGBA{220, 220, 230, 200}

	// Steam
	img.Set(2, 0, steam)
	img.Set(4, 0, steam)
	img.Set(3, 1, steam)

	// Cup body
	for y := 2; y < 8; y++ {
		for x := 1; x < 5; x++ {
			if x == 1 {
				img.Set(x, y, cupDark)
			} else {
				img.Set(x, y, cup)
			}
		}
	}

	// Coffee inside
	for y := 3; y < 5; y++ {
		for x := 2; x < 4; x++ {
			img.Set(x, y, coffee)
		}
	}

	// Handle
	img.Set(5, 4, cupDark)
	img.Set(5, 5, cupDark)

	return img
}

func generateKeyboard() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 12, 5))
	clearImage(img)

	base := color.RGBA{60, 60, 70, 255}
	key := color.RGBA{90, 90, 100, 255}
	keyBright := color.RGBA{120, 120, 130, 255}

	// Base
	for y := 0; y < 5; y++ {
		for x := 0; x < 12; x++ {
			img.Set(x, y, base)
		}
	}

	// Keys
	for y := 1; y < 4; y++ {
		for x := 1; x < 11; x++ {
			if x%2 == 1 {
				img.Set(x, y, keyBright)
			} else {
				img.Set(x, y, key)
			}
		}
	}

	return img
}

func clearImage(img *image.RGBA) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, Clear)
		}
	}
}
