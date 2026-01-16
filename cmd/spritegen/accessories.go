package main

import (
	"image/png"
	"os"
	"path/filepath"
)

func generateAccessories() {
	// Create accessories directories
	os.MkdirAll("assets/accessories/hats", 0755)
	os.MkdirAll("assets/accessories/faces", 0755)
	os.MkdirAll("assets/accessories/capes", 0755)
	os.MkdirAll("assets/accessories/items", 0755)
	os.MkdirAll("assets/effects", 0755)

	// Generate hats
	hats := map[string]func(){
		"wizard":     func() { saveHat("wizard", generateWizardHat()) },
		"party":      func() { saveHat("party", generatePartyHat()) },
		"crown":      func() { saveHat("crown", generateCrown()) },
		"tophat":     func() { saveHat("tophat", generateTopHat()) },
		"propeller":  func() { saveHat("propeller", generatePropellerHat()) },
		"headphones": func() { saveHat("headphones", generateHeadphones()) },
		"beret":      func() { saveHat("beret", generateBeret()) },
		"catears":    func() { saveHat("catears", generateCatEars()) },
		"pirate":     func() { saveHat("pirate", generatePirateHat()) },
		"viking":     func() { saveHat("viking", generateVikingHelmet()) },
		"chef":       func() { saveHat("chef", generateChefHat()) },
		"halo":       func() { saveHat("halo", generateHalo()) },
		"jester":     func() { saveHat("jester", generateJesterCap()) },
		"cowboy":     func() { saveHat("cowboy", generateCowboyHat()) },
		"fedora":     func() { saveHat("fedora", generateFedora()) },
	}
	for _, gen := range hats {
		gen()
	}

	// Generate face accessories
	faces := map[string]func(){
		"dealwithit":  func() { saveFace("dealwithit", generateDealWithIt()) },
		"mustache":    func() { saveFace("mustache", generateMustache()) },
		"monocle":     func() { saveFace("monocle", generateMonocle()) },
		"borat":       func() { saveFace("borat", generateBorat()) },
		"pipe":        func() { saveFace("pipe", generatePipe()) },
		"eyepatch":    func() { saveFace("eyepatch", generateEyepatch()) },
		"glasses3d":   func() { saveFace("glasses3d", generateGlasses3D()) },
		"groucho":     func() { saveFace("groucho", generateGroucho()) },
		"bandana":     func() { saveFace("bandana", generateBandana()) },
		"wizardbeard": func() { saveFace("wizardbeard", generateWizardBeard()) },
	}
	for _, gen := range faces {
		gen()
	}

	// Generate effects
	generateEffects()
}

func saveFace(name string, pixels [][]C) {
	path := filepath.Join("assets/accessories/faces", name+".png")
	savePixels(path, pixels)
}

func saveHat(name string, pixels [][]C) {
	path := filepath.Join("assets/accessories/hats", name+".png")
	savePixels(path, pixels)
}

func savePixels(path string, pixels [][]C) {
	if len(pixels) == 0 {
		return
	}
	height := len(pixels)
	width := len(pixels[0])

	img := createImage(width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, pixels[y][x])
		}
	}

	f, _ := os.Create(path)
	defer f.Close()
	png.Encode(f, img)
}

// Hat generators - Paul Robertson quality: detailed, personality, great shading
func generateWizardHat() [][]C {
	// Classic pointy wizard hat - simple cone shape, wide brim
	o := O // outline

	// Classic purple wizard colors
	m := C{80, 50, 120, 255}   // main purple
	s := C{55, 35, 90, 255}    // shadow
	sd := C{40, 25, 65, 255}   // deep shadow
	h := C{110, 75, 160, 255}  // highlight
	hb := C{140, 100, 190, 255} // bright highlight
	// Gold star/buckle accent
	g := C{255, 215, 0, 255}   // gold
	gd := C{200, 160, 0, 255}  // gold dark

	// Simple pointy cone, wide brim (wider than Claude's head ~18px)
	return [][]C{
		// Pointy tip
		{X, X, X, X, X, X, X, X, X, X, o, X, X, X, X, X, X, X, X, X, X},
		{X, X, X, X, X, X, X, X, X, o, hb, o, X, X, X, X, X, X, X, X, X},
		{X, X, X, X, X, X, X, X, X, o, h, h, o, X, X, X, X, X, X, X, X},
		{X, X, X, X, X, X, X, X, o, h, m, m, h, o, X, X, X, X, X, X, X},
		{X, X, X, X, X, X, X, X, o, m, m, m, m, o, X, X, X, X, X, X, X},
		{X, X, X, X, X, X, X, o, h, m, m, m, m, s, o, X, X, X, X, X, X},
		{X, X, X, X, X, X, X, o, m, m, m, m, m, s, o, X, X, X, X, X, X},
		{X, X, X, X, X, X, o, h, m, m, m, m, m, s, sd, o, X, X, X, X, X},
		{X, X, X, X, X, X, o, m, m, m, m, m, m, m, sd, o, X, X, X, X, X},
		{X, X, X, X, X, o, h, m, m, m, m, m, m, m, s, sd, o, X, X, X, X},
		{X, X, X, X, X, o, m, m, m, m, m, m, m, m, m, sd, o, X, X, X, X},
		{X, X, X, X, o, h, m, m, m, m, m, m, m, m, s, sd, o, X, X, X, X},
		{X, X, X, X, o, m, m, m, m, m, m, m, m, m, m, s, sd, o, X, X, X},
		// Gold buckle/band
		{X, X, X, o, gd, g, g, g, g, g, g, g, g, g, g, g, gd, o, X, X, X},
		// Wide brim
		{X, o, o, h, m, m, m, m, m, m, m, m, m, m, m, m, s, o, o, X, X},
		{o, hb, h, m, m, m, m, m, m, m, m, m, m, m, m, m, s, sd, o, X, X},
		{o, h, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, sd, sd, o, X},
		{X, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, X, X},
	}
}

func generatePartyHat() [][]C {
	o := O
	// Rainbow stripes!
	r := C{255, 80, 100, 255}   // red
	rd := C{200, 50, 70, 255}   // red dark
	y := C{255, 230, 80, 255}   // yellow
	yd := C{220, 190, 50, 255}  // yellow dark
	g := C{80, 220, 120, 255}   // green
	gd := C{50, 180, 90, 255}   // green dark
	b := C{100, 150, 255, 255}  // blue
	bd := C{70, 110, 220, 255}  // blue dark
	// Pompom with sparkle
	pm := C{255, 100, 150, 255} // pompom main
	ps := C{200, 60, 110, 255}  // pompom shadow
	ph := C{255, 180, 200, 255} // pompom highlight
	pw := C{255, 230, 240, 255} // pompom white sparkle

	return [][]C{
		{X, X, X, X, X, X, X, pw, X, X, X, X, X, X, X, X},
		{X, X, X, X, X, X, ph, pm, ph, X, X, X, X, X, X, X},
		{X, X, X, X, X, ph, pm, pm, pm, ph, X, X, X, X, X, X},
		{X, X, X, X, X, ps, pm, pm, pm, ps, X, X, X, X, X, X},
		{X, X, X, X, X, X, o, o, o, X, X, X, X, X, X, X},
		{X, X, X, X, X, o, rd, r, r, o, X, X, X, X, X, X},
		{X, X, X, X, o, rd, r, r, r, r, o, X, X, X, X, X},
		{X, X, X, o, yd, y, y, y, y, y, y, o, X, X, X, X},
		{X, X, o, yd, y, y, y, y, y, y, y, y, o, X, X, X},
		{X, o, gd, gd, g, g, g, g, g, g, g, g, g, o, X, X},
		{o, bd, bd, b, b, b, b, b, b, b, b, b, b, b, o, X},
		{o, rd, r, r, r, r, r, r, r, r, r, r, r, r, r, o},
		{o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o},
	}
}

func generateCrown() [][]C {
	o := O
	g := C{255, 215, 80, 255}   // gold
	gd := C{190, 150, 40, 255}  // gold dark
	gdd := C{140, 100, 20, 255} // gold deep shadow
	gb := C{255, 240, 150, 255} // gold bright
	gw := C{255, 255, 220, 255} // gold white sparkle
	// Jewels
	r := C{220, 40, 60, 255}    // ruby
	rd := C{160, 20, 40, 255}   // ruby dark
	rh := C{255, 100, 120, 255} // ruby highlight
	e := C{40, 200, 100, 255}   // emerald
	ed := C{20, 140, 60, 255}   // emerald dark
	eh := C{120, 255, 160, 255} // emerald highlight
	sb := C{100, 180, 255, 255} // sapphire
	sbd := C{60, 120, 200, 255} // sapphire dark
	sh := C{160, 210, 255, 255} // sapphire highlight

	return [][]C{
		{X, X, o, X, X, X, X, o, X, X, X, X, o, X, X, X, X, o, X, X},
		{X, X, gw, X, X, X, X, gw, X, X, X, X, gw, X, X, X, X, gw, X, X},
		{X, o, gb, o, X, X, o, gb, o, X, X, o, gb, o, X, X, o, gb, o, X},
		{X, o, g, gb, o, o, g, g, gb, o, o, g, g, gb, o, o, g, g, o, X},
		{o, gd, g, g, gb, gb, g, g, g, gb, gb, g, g, g, gb, gb, g, g, gd, o},
		{o, gd, g, g, rh, r, g, g, g, eh, e, g, g, g, sh, sb, g, g, gd, o},
		{o, gd, g, g, r, rd, g, g, g, e, ed, g, g, g, sb, sbd, g, g, gd, o},
		{o, gdd, gd, gd, gd, gd, gd, gd, gd, gd, gd, gd, gd, gd, gd, gd, gd, gd, gdd, o},
		{o, gdd, gd, g, g, g, g, g, g, g, g, g, g, g, g, g, g, gd, gdd, o},
		{o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o},
	}
}

func generateTopHat() [][]C {
	o := O
	m := C{25, 25, 30, 255}    // main black
	md := C{15, 15, 20, 255}   // black dark
	h := C{50, 50, 60, 255}    // highlight
	hb := C{70, 70, 85, 255}   // bright highlight
	hw := C{100, 100, 120, 255} // white shine
	// Satin red band with shine
	r := C{180, 40, 50, 255}   // red
	rd := C{130, 25, 35, 255}  // red dark
	rh := C{220, 80, 90, 255}  // red highlight
	// Gold buckle
	g := C{255, 215, 80, 255}
	gd := C{190, 150, 40, 255}

	return [][]C{
		{X, X, X, X, X, o, o, o, o, o, o, o, o, o, X, X, X, X, X},
		{X, X, X, X, o, md, m, m, m, m, h, h, hb, hb, o, X, X, X, X},
		{X, X, X, X, o, md, m, m, m, m, h, h, hb, hw, o, X, X, X, X},
		{X, X, X, X, o, md, m, m, m, m, h, h, hb, hb, o, X, X, X, X},
		{X, X, X, X, o, md, m, m, m, m, h, h, hb, hb, o, X, X, X, X},
		{X, X, X, X, o, md, m, m, m, m, h, h, hb, hb, o, X, X, X, X},
		{X, X, X, X, o, md, m, m, m, m, h, h, hb, hb, o, X, X, X, X},
		{X, X, X, X, o, rd, r, r, gd, g, g, gd, rh, rh, o, X, X, X, X},
		{X, X, X, X, o, rd, r, r, gd, g, g, gd, rh, rh, o, X, X, X, X},
		{X, X, X, X, o, md, m, m, m, m, h, h, hb, hb, o, X, X, X, X},
		{X, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, X},
		{o, md, md, m, m, m, m, m, m, m, h, h, h, hb, hb, hb, hw, hb, o},
		{o, md, md, m, m, m, m, m, m, m, h, h, h, hb, hb, hb, hb, hb, o},
		{o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o},
	}
}

func generatePropellerHat() [][]C {
	o := O
	// Propeller blades with motion blur feel
	r := C{230, 70, 70, 255}    // red
	rd := C{180, 40, 50, 255}   // red dark
	rh := C{255, 120, 120, 255} // red highlight
	b := C{70, 130, 230, 255}   // blue
	bd := C{40, 90, 180, 255}   // blue dark
	bh := C{120, 170, 255, 255} // blue highlight
	// Center hub
	y := C{255, 220, 80, 255}   // yellow
	yd := C{200, 160, 40, 255}  // yellow dark
	yh := C{255, 250, 180, 255} // yellow highlight
	// Cap with beanie texture
	m := C{70, 130, 200, 255}   // main blue
	md := C{45, 95, 160, 255}   // dark
	mh := C{100, 160, 230, 255} // highlight
	mb := C{130, 180, 245, 255} // bright
	// Red stripe
	sr := C{230, 70, 80, 255}
	srd := C{180, 40, 50, 255}

	return [][]C{
		{X, X, rd, r, rh, X, X, X, X, X, X, X, bd, b, bh, X, X},
		{X, X, X, rd, r, rh, X, X, X, X, X, bd, b, bh, X, X, X},
		{X, X, X, X, rd, r, o, yd, y, yh, o, b, bh, X, X, X, X},
		{X, X, X, X, X, o, yd, y, y, y, yh, o, X, X, X, X, X},
		{X, X, X, X, X, o, yd, y, yh, y, yh, o, X, X, X, X, X},
		{X, X, X, X, X, X, o, o, o, o, o, X, X, X, X, X, X},
		{X, X, X, X, o, md, m, m, m, m, mh, mb, o, X, X, X, X},
		{X, X, X, o, md, m, m, m, m, m, m, mh, mb, o, X, X, X},
		{X, X, o, md, srd, sr, sr, sr, sr, sr, sr, sr, mh, mb, o, X, X},
		{X, o, md, m, m, m, m, m, m, m, m, m, m, mh, mb, o, X},
		{o, md, m, m, m, m, m, m, m, m, m, m, m, m, mh, mb, o},
		{o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o},
	}
}

func generateHalo() [][]C {
	// Divine radiant halo with glow
	w := C{255, 255, 255, 255}  // pure white
	i := C{255, 255, 220, 255}  // inner bright
	m := C{255, 250, 180, 255}  // middle gold
	o := C{255, 240, 140, 255}  // outer gold
	od := C{230, 210, 100, 255} // outer dark
	// Soft glow (semi-transparent)
	g1 := C{255, 255, 200, 180} // glow bright
	g2 := C{255, 250, 160, 120} // glow medium
	g3 := C{255, 245, 120, 60}  // glow soft

	return [][]C{
		{X, X, X, X, g3, g2, g1, g1, g1, g1, g2, g3, X, X, X, X},
		{X, X, g3, g2, g1, i, i, w, w, i, i, g1, g2, g3, X, X},
		{X, g3, g1, i, m, m, m, m, m, m, m, m, i, g1, g3, X},
		{X, g2, i, m, o, od, X, X, X, X, od, o, m, i, g2, X},
		{g3, g1, m, o, od, X, X, X, X, X, X, od, o, m, g1, g3},
		{g2, i, m, od, X, X, X, X, X, X, X, X, od, m, i, g2},
		{g2, i, m, od, X, X, X, X, X, X, X, X, od, m, i, g2},
		{g3, g1, m, o, od, X, X, X, X, X, X, od, o, m, g1, g3},
		{X, g2, i, m, o, od, X, X, X, X, od, o, m, i, g2, X},
		{X, g3, g1, i, m, m, m, m, m, m, m, m, i, g1, g3, X},
		{X, X, g3, g2, g1, i, i, i, i, i, i, g1, g2, g3, X, X},
		{X, X, X, X, g3, g2, g1, g1, g1, g1, g2, g3, X, X, X, X},
	}
}

func generateEffects() {
	// Heart
	savePixels("assets/effects/heart.png", generateHeartEffect())

	// Sparkles
	savePixels("assets/effects/spark_small.png", generateSparkEffect(2))
	savePixels("assets/effects/spark_medium.png", generateSparkEffect(3))
	savePixels("assets/effects/spark_large.png", generateSparkEffect(4))

	// Thought bubble
	savePixels("assets/effects/thought_dot.png", generateThoughtDot())
}

func generateHeartEffect() [][]C {
	h := H // highlight orange
	p := P // primary
	s := S // shadow
	o := O // outline

	return [][]C{
		{X, o, o, X, o, o, X},
		{o, h, p, o, h, p, o},
		{o, p, p, p, p, s, o},
		{o, p, p, p, p, s, o},
		{X, o, p, p, s, o, X},
		{X, X, o, p, o, X, X},
		{X, X, X, o, X, X, X},
	}
}

func generateSparkEffect(size int) [][]C {
	result := make([][]C, size)
	for i := range result {
		result[i] = make([]C, size)
		for j := range result[i] {
			result[i][j] = X
		}
	}

	mid := size / 2
	yb := C{0xFF, 0xFF, 0xC8, 0xFF} // bright
	ym := Y                          // medium
	yd := C{0xFF, 0xDC, 0x50, 0xFF} // dark

	result[0][mid] = yb
	result[mid][0] = ym
	result[mid][size-1] = ym
	result[size-1][mid] = yd
	if size > 2 {
		result[mid][mid] = W
	}

	return result
}

func generateThoughtDot() [][]C {
	return [][]C{
		{X, W, X},
		{W, W, W},
		{X, W, X},
	}
}

// ============ FACE ACCESSORIES ============
// Paul Robertson quality: detail, shine, personality!

func generateSunglasses() [][]C {
	// Aviator style with gradient lens and shine
	o := O                        // outline
	f := C{180, 160, 120, 255}    // gold frame
	fd := C{140, 120, 80, 255}    // frame dark
	fh := C{220, 200, 160, 255}   // frame highlight
	fw := C{255, 250, 220, 255}   // frame white shine
	// Gradient lens - darker at top
	l1 := C{40, 35, 50, 255}      // lens top (darkest)
	l2 := C{60, 55, 70, 255}      // lens mid-dark
	l3 := C{80, 75, 90, 255}      // lens mid
	l4 := C{100, 95, 110, 255}    // lens bottom (lightest)
	ls := C{140, 160, 180, 200}   // lens shine streak

	return [][]C{
		{X, X, X, o, o, o, o, o, o, X, X, X, o, o, o, o, o, o, X, X, X},
		{X, X, o, fd, f, fw, f, f, fd, o, o, o, fd, f, fw, f, f, fd, o, X, X},
		{X, o, fd, l1, l1, ls, l1, l1, l1, fd, o, fd, l1, l1, ls, l1, l1, l1, fd, o, X},
		{o, fd, l1, l2, l2, ls, l2, l2, l2, f, o, f, l2, l2, ls, l2, l2, l2, f, fd, o},
		{o, f, l2, l3, l3, l3, l3, l3, l3, fh, o, fh, l3, l3, l3, l3, l3, l3, fh, f, o},
		{o, f, l3, l4, l4, l4, l4, l4, l4, fh, X, fh, l4, l4, l4, l4, l4, l4, fh, f, o},
		{X, o, fd, f, f, fh, f, f, fd, o, X, X, o, fd, f, fh, f, f, fd, o, X},
		{X, X, o, o, o, o, o, o, o, X, X, X, X, o, o, o, o, o, o, X, X},
	}
}

func generateDealWithIt() [][]C {
	// Classic 8-bit pixel glasses - blocky on purpose but with style
	o := O
	b := C{5, 5, 8, 255}         // pure black lens
	f := C{15, 15, 18, 255}      // frame
	w := C{80, 80, 100, 255}     // reflection

	return [][]C{
		{X, o, o, o, o, o, o, o, X, X, X, o, o, o, o, o, o, o, X},
		{o, f, f, f, f, f, f, f, o, X, o, f, f, f, f, f, f, f, o},
		{o, f, b, b, b, b, b, f, o, o, o, f, b, b, b, b, b, f, o},
		{o, f, b, w, b, b, b, f, f, f, f, f, b, w, b, b, b, f, o},
		{o, f, b, b, b, b, b, f, o, X, o, f, b, b, b, b, b, f, o},
		{o, f, f, f, f, f, f, f, o, X, o, f, f, f, f, f, f, f, o},
		{X, o, o, o, o, o, o, o, X, X, X, o, o, o, o, o, o, o, X},
	}
}

func generateMustache() [][]C {
	// 70s/80s "porn stache" - smaller, thinner version
	m := C{50, 35, 25, 255}       // main brown
	md := C{30, 20, 12, 255}      // dark brown
	mh := C{80, 55, 40, 255}      // highlight
	o := O

	// Smaller, thinner - 12px wide, 3px tall
	return [][]C{
		{X, o, o, o, o, o, o, o, o, o, o, X},
		{o, md, m, mh, m, m, m, m, mh, m, md, o},
		{X, o, md, m, m, m, m, m, m, md, o, X},
	}
}

func generateMonocle() [][]C {
	// Round monocle with thin gold frame, chain pointing straight DOWN
	g := C{255, 215, 80, 255}    // gold frame (thin 1px)
	gw := C{255, 255, 220, 255}  // gold sparkle
	l := C{200, 220, 240, 180}   // lens
	lh := C{230, 245, 255, 200}  // lens highlight
	lw := C{255, 255, 255, 220}  // lens white sparkle
	c := C{200, 180, 120, 255}   // chain

	// Round monocle with thin 1px gold frame
	return [][]C{
		// Row 0: top of circle
		{X, g, g, g, X},
		// Row 1: upper
		{g, l, lw, l, g},
		// Row 2: middle
		{g, l, lh, l, g},
		// Row 3: lower
		{g, l, l, l, g},
		// Row 4: bottom with chain
		{X, g, gw, g, X},
		// Row 5-8: chain straight down
		{X, X, c, X, X},
		{X, X, c, X, X},
		{X, X, c, X, X},
		{X, X, c, X, X},
	}
}

func generateBorat() [][]C {
	// The infamous mankini - lime green, U-shaped
	// Wider (20px), thinner straps (2px), taller to reach shoulders
	g := C{50, 205, 50, 255}     // lime green
	gd := C{30, 150, 30, 255}    // green dark
	gh := C{100, 240, 100, 255}  // green highlight
	gw := C{180, 255, 180, 255}  // green white shine
	o := O

	// Wide U-shape: thin straps on far edges, curve to small pouch at bottom
	return [][]C{
		// Row 0-3: Thin shoulder straps on outer edges going down
		{o, g, gh, X, X, X, X, X, X, X, X, X, X, X, X, X, X, gh, g, o},
		{o, g, gh, X, X, X, X, X, X, X, X, X, X, X, X, X, X, gh, g, o},
		{o, gd, g, X, X, X, X, X, X, X, X, X, X, X, X, X, X, g, gd, o},
		{o, gd, g, X, X, X, X, X, X, X, X, X, X, X, X, X, X, g, gd, o},
		// Row 4-5: Straps curve inward
		{X, o, gd, g, X, X, X, X, X, X, X, X, X, X, X, X, g, gd, o, X},
		{X, X, o, gd, g, gw, X, X, X, X, X, X, X, X, gw, g, gd, o, X, X},
		// Row 6-7: Straps meet at center
		{X, X, X, o, gd, g, gh, X, X, X, X, X, X, gh, g, gd, o, X, X, X},
		{X, X, X, X, o, gd, g, gw, X, X, X, X, gw, g, gd, o, X, X, X, X},
		// Row 8-9: Front pouch
		{X, X, X, X, X, o, gd, g, gh, gh, gh, gh, g, gd, o, X, X, X, X, X},
		{X, X, X, X, X, X, o, gd, g, gw, gw, g, gd, o, X, X, X, X, X, X},
		{X, X, X, X, X, X, X, o, gd, g, g, gd, o, X, X, X, X, X, X, X},
	}
}

// ============ NEW HAT GENERATORS ============
// All sprites use 16px width for consistency

func generateHeadphones() [][]C {
	o := O
	m := C{40, 40, 45, 255}
	md := C{25, 25, 30, 255}
	mh := C{60, 60, 70, 255}
	r := C{200, 50, 60, 255}
	c := C{35, 35, 40, 255}

	return [][]C{
		{X, X, X, o, o, o, o, o, o, o, o, o, o, X, X, X},
		{X, X, o, md, m, m, m, m, m, m, m, m, md, o, X, X},
		{X, o, md, m, mh, mh, mh, mh, mh, mh, mh, m, md, o, X, X},
		{o, c, o, X, X, X, X, X, X, X, X, X, X, o, c, o},
		{o, c, r, o, X, X, X, X, X, X, X, X, o, r, c, o},
		{o, c, r, o, X, X, X, X, X, X, X, X, o, r, c, o},
		{o, c, o, X, X, X, X, X, X, X, X, X, X, o, c, o},
		{X, o, o, X, X, X, X, X, X, X, X, X, X, o, o, X},
	}
}

func generateBeret() [][]C {
	o := O
	// Classic French artist beret - burgundy/wine red
	m := C{120, 40, 50, 255}    // main
	md := C{90, 25, 35, 255}    // dark
	mh := C{160, 60, 70, 255}   // highlight
	mb := C{190, 90, 100, 255}  // bright

	return [][]C{
		{X, X, X, X, X, X, o, o, o, X, X, X, X, X, X},
		{X, X, X, X, o, o, mb, mh, mh, o, o, X, X, X, X},
		{X, X, X, o, mh, mb, m, m, m, m, mh, o, X, X, X},
		{X, X, o, mh, m, m, m, m, m, m, m, m, o, X, X},
		{X, o, mh, m, m, m, m, m, m, m, m, m, md, o, X},
		{o, mh, m, m, m, m, m, m, m, m, m, m, m, md, o},
		{o, m, m, m, m, m, m, m, m, m, m, m, m, md, o},
		{o, o, o, o, o, o, o, o, o, o, o, o, o, o, o},
	}
}

func generateCatEars() [][]C {
	o := O
	// Cute cat ears - pink inside
	m := C{60, 55, 65, 255}     // main gray
	md := C{40, 35, 45, 255}    // dark
	mh := C{85, 80, 95, 255}    // highlight
	// Pink inner ear
	p := C{255, 150, 170, 255}
	pd := C{220, 110, 130, 255}

	return [][]C{
		{X, o, o, X, X, X, X, X, X, X, X, X, o, o, X},
		{o, mh, m, o, X, X, X, X, X, X, X, o, mh, m, o},
		{o, mh, p, md, o, X, X, X, X, X, o, mh, p, md, o},
		{o, m, pd, p, md, o, X, X, X, o, m, pd, p, md, o},
		{o, m, m, pd, m, md, o, o, o, m, m, pd, m, md, o},
		{X, o, m, m, m, m, md, md, md, m, m, m, m, o, X},
		{X, X, o, o, o, o, o, o, o, o, o, o, o, X, X},
	}
}

func generatePirateHat() [][]C {
	o := O
	// Tricorn pirate hat with skull
	m := C{30, 25, 35, 255}     // main black
	md := C{20, 15, 25, 255}    // dark
	mh := C{50, 45, 60, 255}    // highlight
	// Gold trim
	g := C{255, 215, 80, 255}
	gd := C{200, 160, 40, 255}
	// Skull
	w := C{240, 235, 225, 255}
	wd := C{200, 195, 185, 255}

	return [][]C{
		{X, X, X, X, X, X, o, o, o, o, o, X, X, X, X, X, X},
		{X, X, X, o, o, o, mh, m, m, m, mh, o, o, o, X, X, X},
		{X, X, o, mh, m, m, m, m, m, m, m, m, m, mh, o, X, X},
		{X, o, mh, m, m, w, w, m, m, m, w, w, m, m, mh, o, X},
		{o, gd, g, m, m, wd, wd, m, w, m, wd, wd, m, m, g, gd, o},
		{o, mh, m, m, m, m, w, w, w, w, w, m, m, m, m, mh, o},
		{o, m, m, m, m, m, m, w, w, w, m, m, m, m, m, m, o},
		{X, o, md, m, m, m, m, m, m, m, m, m, m, m, md, o, X},
		{X, X, o, o, md, m, m, m, m, m, m, m, md, o, o, X, X},
		{o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o},
	}
}

func generateVikingHelmet() [][]C {
	o := O
	// Viking helmet with horns
	m := C{140, 130, 115, 255}  // main metal
	md := C{100, 90, 75, 255}   // dark
	mh := C{180, 170, 155, 255} // highlight
	mb := C{210, 200, 185, 255} // bright
	// Horn
	h := C{230, 220, 190, 255}
	hd := C{180, 165, 130, 255}
	hb := C{250, 245, 230, 255}

	return [][]C{
		{X, hb, h, o, X, X, X, X, X, X, X, X, X, o, hb, h, X},
		{o, h, hd, o, X, X, X, X, X, X, X, X, X, o, h, hd, o},
		{o, hd, h, o, X, X, X, X, X, X, X, X, X, o, hd, h, o},
		{X, o, h, hd, o, o, o, o, o, o, o, o, o, hd, h, o, X},
		{X, X, o, hd, mb, mh, mh, mh, mh, mh, mh, mh, mb, hd, o, X, X},
		{X, X, X, o, mh, m, m, m, m, m, m, m, mh, o, X, X, X},
		{X, X, X, o, m, m, m, m, m, m, m, m, m, o, X, X, X},
		{X, X, X, o, md, m, m, m, m, m, m, m, md, o, X, X, X},
		{X, X, X, o, o, o, o, o, o, o, o, o, o, o, X, X, X},
	}
}

func generateChefHat() [][]C {
	o := O
	// Tall white chef toque with vertical pleats
	w := C{255, 255, 255, 255}   // white
	wd := C{230, 230, 235, 255}  // white dark
	wdd := C{200, 200, 210, 255} // white darker
	wh := C{255, 255, 255, 255}  // highlight
	pl := C{220, 220, 225, 255}  // pleat line (subtle vertical crease)

	return [][]C{
		{X, X, X, X, o, o, o, o, o, o, o, X, X, X, X},
		{X, X, o, o, w, wh, pl, wh, pl, wh, w, o, o, X, X},
		{X, o, w, wh, w, pl, w, w, w, pl, wh, w, w, o, X},
		{X, o, w, w, pl, w, w, pl, w, w, pl, w, w, o, X},
		{X, o, wd, w, pl, w, w, pl, w, w, pl, w, wd, o, X},
		{X, o, wd, w, pl, w, w, pl, w, w, pl, w, wd, o, X},
		{X, o, wdd, wd, pl, w, w, pl, w, w, pl, wd, wdd, o, X},
		{X, o, wdd, wd, pl, w, w, pl, w, w, pl, wd, wdd, o, X},
		{o, wdd, wdd, wd, wd, wd, wd, wd, wd, wd, wd, wd, wdd, wdd, o},
		{o, o, o, o, o, o, o, o, o, o, o, o, o, o, o},
	}
}

func generateJesterCap() [][]C {
	o := O
	// Three-pointed jester cap with bells
	// Purple and gold
	p := C{120, 60, 160, 255}   // purple
	pd := C{80, 40, 110, 255}   // purple dark
	ph := C{160, 100, 200, 255} // purple highlight
	// Gold bells
	g := C{255, 215, 80, 255}
	gd := C{200, 160, 40, 255}
	gw := C{255, 245, 180, 255}

	return [][]C{
		{X, gw, g, o, X, X, X, X, X, X, X, o, gw, g, o, X, X},
		{X, o, gd, o, X, X, X, gw, g, o, X, o, gd, g, o, X, X},
		{X, o, ph, pd, o, X, o, gd, g, o, o, ph, p, pd, o, X, X},
		{o, ph, p, p, pd, o, o, ph, o, o, ph, p, p, pd, o, X, X},
		{o, ph, p, p, p, pd, ph, p, pd, ph, p, p, p, p, pd, o, X},
		{o, p, p, p, p, p, p, p, p, p, p, p, p, p, p, p, o},
		{o, pd, p, p, p, p, p, p, p, p, p, p, p, p, p, pd, o},
		{o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o},
	}
}

func generateCowboyHat() [][]C {
	o := O
	// Classic brown cowboy hat
	m := C{140, 90, 50, 255}    // main brown
	md := C{100, 60, 30, 255}   // dark
	mh := C{180, 130, 80, 255}  // highlight
	mb := C{210, 160, 110, 255} // bright
	// Hat band
	b := C{60, 40, 25, 255}
	bh := C{90, 60, 40, 255}

	return [][]C{
		{X, X, X, X, X, o, o, o, o, o, o, o, X, X, X, X, X},
		{X, X, X, X, o, mb, mh, mh, mh, mh, mh, mb, o, X, X, X, X},
		{X, X, X, o, mh, m, m, m, m, m, m, m, mh, o, X, X, X},
		{X, X, X, o, m, m, m, m, m, m, m, m, m, o, X, X, X},
		{X, X, X, o, b, b, bh, bh, bh, bh, bh, b, b, o, X, X, X},
		{X, X, X, o, md, m, m, m, m, m, m, m, md, o, X, X, X},
		{X, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, X},
		{o, mb, mh, m, m, m, m, m, m, m, m, m, m, mh, mb, mh, o},
		{o, mh, m, md, md, md, md, md, md, md, md, md, m, m, mh, m, o},
		{X, o, o, o, o, o, o, o, o, o, o, o, o, o, o, o, X},
	}
}

func generateFedora() [][]C {
	o := O
	// Classic gray fedora
	m := C{80, 75, 85, 255}     // main gray
	md := C{55, 50, 60, 255}    // dark
	mh := C{110, 105, 120, 255} // highlight
	mb := C{140, 135, 150, 255} // bright
	// Black band
	b := C{25, 25, 30, 255}
	bh := C{45, 45, 55, 255}

	return [][]C{
		{X, X, X, X, X, o, o, o, o, o, o, X, X, X, X, X},
		{X, X, X, o, o, mb, mh, mh, mh, mb, o, o, X, X, X, X},
		{X, X, o, mh, m, m, m, m, m, m, m, mh, o, X, X, X},
		{X, X, o, m, m, m, m, m, m, m, m, m, o, X, X, X},
		{X, X, o, b, b, bh, b, b, b, bh, b, b, o, X, X, X},
		{X, X, o, md, m, m, m, m, m, m, m, md, o, X, X, X},
		{X, o, o, o, o, o, o, o, o, o, o, o, o, o, X, X},
		{o, mb, mh, m, m, m, m, m, m, m, m, m, mh, mb, o, X},
		{o, mh, m, md, md, md, md, md, md, md, md, m, mh, o, X, X},
		{X, o, o, o, o, o, o, o, o, o, o, o, o, o, X, X},
	}
}

// ============ NEW FACE GENERATORS ============

func generatePipe() [][]C {
	o := O
	// Classic smoking pipe - horizontal stem, bowl at end
	w := C{100, 65, 40, 255}    // wood brown
	wd := C{70, 45, 25, 255}    // wood dark
	wh := C{140, 100, 65, 255}  // wood highlight
	wb := C{60, 35, 20, 255}    // bowl interior (dark)
	// Smoke wisps
	s1 := C{220, 220, 230, 160}
	s2 := C{200, 200, 215, 100}

	return [][]C{
		// Smoke rising from bowl
		{X, X, X, X, X, X, X, X, s2, X, X},
		{X, X, X, X, X, X, X, s1, s2, s1, X},
		{X, X, X, X, X, X, X, s2, s1, X, X},
		// Bowl (top opening)
		{X, X, X, X, X, X, o, wb, wb, o, X},
		{X, X, X, X, X, o, wd, wb, wb, wh, o},
		{X, X, X, X, X, o, wd, w, w, wh, o},
		// Bowl curves into stem
		{X, X, X, X, X, X, o, wd, w, o, X},
		// Horizontal stem to mouth
		{o, o, o, o, o, o, o, w, o, X, X},
		{o, wh, w, w, w, w, wh, o, X, X, X},
		{X, o, o, o, o, o, o, X, X, X, X},
	}
}

func generateEyepatch() [][]C {
	o := O
	// Small eyepatch with diagonal \ strap
	b := C{20, 15, 25, 255}  // black patch
	bd := C{10, 8, 15, 255}  // black dark
	s := C{50, 40, 30, 255}  // brown strap

	return [][]C{
		// Diagonal strap going \ down-right
		{s, X, X, X, X, X, X},
		{X, s, o, o, o, X, X},
		{X, o, bd, b, bd, o, X},
		{X, o, b, b, b, o, s},
		{X, o, bd, b, bd, o, X},
		{X, X, o, o, o, X, X},
	}
}

func generateGlasses3D() [][]C {
	o := O
	// Retro red/cyan 3D glasses
	r := C{220, 50, 50, 200}    // red lens
	rd := C{180, 30, 30, 200}   // red dark
	c := C{50, 200, 220, 200}   // cyan lens
	cd := C{30, 160, 180, 200}  // cyan dark
	// White frame
	f := C{240, 240, 245, 255}
	fd := C{200, 200, 210, 255}

	return [][]C{
		{X, o, o, o, o, o, o, o, o, X, X, o, o, o, o, o, o, o, o, X},
		{o, f, f, f, f, f, f, f, f, o, o, f, f, f, f, f, f, f, f, o},
		{o, f, rd, r, r, r, r, f, f, o, o, f, cd, c, c, c, c, f, f, o},
		{o, f, r, r, r, r, r, r, f, f, f, f, c, c, c, c, c, c, f, o},
		{o, f, r, r, r, r, r, r, f, o, o, f, c, c, c, c, c, c, f, o},
		{o, f, rd, r, r, r, r, f, f, o, o, f, cd, c, c, c, c, f, f, o},
		{o, fd, f, f, f, f, f, f, fd, o, o, fd, f, f, f, f, f, f, fd, o},
		{X, o, o, o, o, o, o, o, o, X, X, o, o, o, o, o, o, o, o, X},
	}
}

func generateGroucho() [][]C {
	o := O
	// Groucho Marx glasses with eyebrows, nose, and mustache
	b := C{30, 25, 20, 255}     // black
	// Skin tone nose
	n := C{230, 190, 160, 255}
	nd := C{200, 160, 130, 255}
	// Lens
	l := C{180, 200, 220, 180}

	return [][]C{
		// Bushy eyebrows
		{o, b, b, b, b, o, X, X, X, X, X, o, b, b, b, b, o},
		{b, b, b, b, b, b, o, X, X, X, o, b, b, b, b, b, b},
		// Glasses frame
		{o, o, o, o, o, o, o, X, X, X, o, o, o, o, o, o, o},
		{o, l, l, l, l, o, X, X, X, X, X, o, l, l, l, l, o},
		{o, l, l, l, l, o, X, o, o, o, X, o, l, l, l, l, o},
		{o, o, o, o, o, o, X, o, n, o, X, o, o, o, o, o, o},
		// Big nose
		{X, X, X, X, X, X, o, nd, n, nd, o, X, X, X, X, X, X},
		{X, X, X, X, X, X, X, o, n, o, X, X, X, X, X, X, X},
		// Mustache
		{X, X, o, b, b, b, b, b, b, b, b, b, b, b, o, X, X},
		{X, o, b, b, b, b, b, b, b, b, b, b, b, b, b, o, X},
	}
}

func generateBandana() [][]C {
	o := O
	// Bandana headband with knot and trailing tails
	r := C{180, 40, 50, 255}   // red fabric
	rd := C{140, 25, 35, 255}  // red dark
	rh := C{220, 70, 80, 255}  // red highlight

	return [][]C{
		// Headband wraps around forehead with knot on side
		{X, o, o, o, o, o, o, o, o, o, o, o, o, X, X},
		{o, rh, r, r, r, r, r, r, r, r, r, r, rd, o, X},
		{o, r, rd, r, r, r, r, r, r, r, r, rd, r, o, X},
		{X, o, o, o, o, o, o, o, o, o, o, o, rd, rh, o},
		// Trailing tails from knot
		{X, X, X, X, X, X, X, X, X, X, X, X, o, r, rd},
		{X, X, X, X, X, X, X, X, X, X, X, X, X, o, r},
		{X, X, X, X, X, X, X, X, X, X, X, X, X, X, o},
	}
}

func generateWizardBeard() [][]C {
	o := O
	// Long flowing wizard beard - white/gray
	w := C{240, 240, 245, 255}  // white
	wd := C{210, 210, 220, 255} // white dark
	wdd := C{180, 180, 195, 255} // darker
	wh := C{255, 255, 255, 255} // highlight

	return [][]C{
		{X, X, X, o, o, o, o, o, o, o, o, o, X, X, X},
		{X, X, o, wh, w, w, w, w, w, w, w, wh, o, X, X},
		{X, o, wh, w, w, w, w, w, w, w, w, w, wh, o, X},
		{o, wh, w, w, w, w, w, w, w, w, w, w, w, wh, o},
		{o, w, w, w, w, w, w, w, w, w, w, w, w, w, o},
		{o, w, wd, w, w, w, w, w, w, w, w, w, wd, w, o},
		{X, o, wd, w, w, w, w, w, w, w, w, w, wd, o, X},
		{X, o, wdd, wd, w, w, w, w, w, w, w, wd, wdd, o, X},
		{X, X, o, wdd, wd, w, w, w, w, w, wd, wdd, o, X, X},
		{X, X, X, o, wdd, wd, w, w, w, wd, wdd, o, X, X, X},
		{X, X, X, X, o, wdd, wd, w, wd, wdd, o, X, X, X, X},
		{X, X, X, X, X, o, wdd, wd, wdd, o, X, X, X, X, X},
		{X, X, X, X, X, X, o, o, o, X, X, X, X, X, X},
	}
}
