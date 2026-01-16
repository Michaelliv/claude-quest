//go:build debug

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// HotReloader watches asset files and triggers reloads
type HotReloader struct {
	watcher       *fsnotify.Watcher
	renderer      *Renderer
	reloadQueue   chan string
	regenerating  bool
	mu            sync.Mutex
	lastRegenTime time.Time
}

// NewHotReloader creates a new hot reloader for the given renderer
func NewHotReloader(renderer *Renderer) (*HotReloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	hr := &HotReloader{
		watcher:     watcher,
		renderer:    renderer,
		reloadQueue: make(chan string, 100),
	}

	return hr, nil
}

// Start begins watching for file changes
func (hr *HotReloader) Start() error {
	// Watch assets directory recursively
	assetsPath := "assets"
	err := filepath.Walk(assetsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			if watchErr := hr.watcher.Add(path); watchErr != nil {
				fmt.Printf("Warning: couldn't watch %s: %v\n", path, watchErr)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Warning: couldn't walk assets directory: %v\n", err)
	}

	// Watch spritegen directory for source changes
	spritegenPath := "cmd/spritegen"
	if _, err := os.Stat(spritegenPath); err == nil {
		err := filepath.Walk(spritegenPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				if watchErr := hr.watcher.Add(path); watchErr != nil {
					fmt.Printf("Warning: couldn't watch %s: %v\n", path, watchErr)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Warning: couldn't walk spritegen directory: %v\n", err)
		}
	}

	// Start the watcher goroutine
	go hr.watchLoop()

	fmt.Println("Hot reload enabled - watching assets/ and cmd/spritegen/")
	fmt.Println("  Press R to force reload all textures")
	fmt.Println("  Press G to regenerate sprites")

	return nil
}

// watchLoop handles file system events
func (hr *HotReloader) watchLoop() {
	// Debounce timer to avoid multiple reloads for the same file
	debounce := make(map[string]time.Time)
	debounceInterval := 100 * time.Millisecond

	for {
		select {
		case event, ok := <-hr.watcher.Events:
			if !ok {
				return
			}

			// Only care about writes and creates
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// Debounce
			if lastTime, exists := debounce[event.Name]; exists {
				if time.Since(lastTime) < debounceInterval {
					continue
				}
			}
			debounce[event.Name] = time.Now()

			// Check what kind of file changed
			ext := strings.ToLower(filepath.Ext(event.Name))

			if ext == ".png" {
				// PNG file changed - queue for reload
				fmt.Printf("Asset changed: %s\n", event.Name)
				hr.reloadQueue <- event.Name
			} else if ext == ".go" && strings.Contains(event.Name, "spritegen") {
				// Spritegen source changed - regenerate
				fmt.Printf("Spritegen source changed: %s\n", event.Name)
				hr.QueueRegenerate()
			}

		case err, ok := <-hr.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

// ProcessReloads should be called from the main thread to process pending reloads
// Returns true if any textures were reloaded
func (hr *HotReloader) ProcessReloads() bool {
	reloaded := false

	// Process all pending reloads (non-blocking)
	for {
		select {
		case path := <-hr.reloadQueue:
			hr.reloadTexture(path)
			reloaded = true
		default:
			return reloaded
		}
	}
}

// reloadTexture reloads a specific texture based on its path
func (hr *HotReloader) reloadTexture(path string) {
	// Normalize the path
	path = filepath.Clean(path)

	// Determine which texture to reload based on path
	if strings.Contains(path, "spritesheet.png") && !strings.Contains(path, "mini") && !strings.Contains(path, "enemy") {
		// Main sprite sheet
		if hr.renderer.spriteSheet.ID != 0 {
			rl.UnloadTexture(hr.renderer.spriteSheet)
		}
		hr.renderer.spriteSheet = rl.LoadTexture(path)
		hr.renderer.hasSprites = true
		fmt.Printf("Reloaded: main spritesheet\n")

	} else if strings.Contains(path, "mini_spritesheet.png") {
		// Mini sprite sheet
		if hr.renderer.miniSpriteSheet.ID != 0 {
			rl.UnloadTexture(hr.renderer.miniSpriteSheet)
		}
		hr.renderer.miniSpriteSheet = rl.LoadTexture(path)
		hr.renderer.hasMiniSprites = true
		fmt.Printf("Reloaded: mini spritesheet\n")

	} else if strings.Contains(path, "enemy_spritesheet.png") {
		// Enemy sprite sheet
		if hr.renderer.enemySpriteSheet.ID != 0 {
			rl.UnloadTexture(hr.renderer.enemySpriteSheet)
		}
		hr.renderer.enemySpriteSheet = rl.LoadTexture(path)
		hr.renderer.hasEnemySprites = true
		fmt.Printf("Reloaded: enemy spritesheet\n")

	} else if strings.Contains(path, "chest.png") {
		// Chest texture
		if hr.renderer.chestTexture.ID != 0 {
			rl.UnloadTexture(hr.renderer.chestTexture)
		}
		hr.renderer.chestTexture = rl.LoadTexture(path)
		hr.renderer.hasChestTexture = true
		fmt.Printf("Reloaded: chest texture\n")

	} else if strings.Contains(path, "accessories/hats/") {
		// Hat texture - find and reload specific hat
		hr.reloadHat(path)

	} else if strings.Contains(path, "accessories/faces/") {
		// Face texture - find and reload specific face
		hr.reloadFace(path)

	} else {
		fmt.Printf("Unknown asset type, skipping: %s\n", path)
	}
}

// reloadHat reloads a specific hat texture
func (hr *HotReloader) reloadHat(path string) {
	name := strings.TrimSuffix(filepath.Base(path), ".png")
	for i, hatName := range hr.renderer.hatNames {
		if hatName == name {
			if hr.renderer.hats[i].ID != 0 {
				rl.UnloadTexture(hr.renderer.hats[i])
			}
			hr.renderer.hats[i] = rl.LoadTexture(path)
			fmt.Printf("Reloaded: hat '%s'\n", name)
			return
		}
	}
	// New hat - add it
	hr.renderer.hatNames = append(hr.renderer.hatNames, name)
	hr.renderer.hats = append(hr.renderer.hats, rl.LoadTexture(path))
	fmt.Printf("Added new hat: '%s'\n", name)
}

// reloadFace reloads a specific face texture
func (hr *HotReloader) reloadFace(path string) {
	name := strings.TrimSuffix(filepath.Base(path), ".png")
	for i, faceName := range hr.renderer.faceNames {
		if faceName == name {
			if hr.renderer.faces[i].ID != 0 {
				rl.UnloadTexture(hr.renderer.faces[i])
			}
			hr.renderer.faces[i] = rl.LoadTexture(path)
			fmt.Printf("Reloaded: face '%s'\n", name)
			return
		}
	}
	// New face - add it
	hr.renderer.faceNames = append(hr.renderer.faceNames, name)
	hr.renderer.faces = append(hr.renderer.faces, rl.LoadTexture(path))
	fmt.Printf("Added new face: '%s'\n", name)
}

// QueueRegenerate queues a sprite regeneration (debounced)
func (hr *HotReloader) QueueRegenerate() {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	// Debounce regeneration requests
	if time.Since(hr.lastRegenTime) < 500*time.Millisecond {
		return
	}
	hr.lastRegenTime = time.Now()

	if hr.regenerating {
		return
	}

	hr.regenerating = true
	go hr.doRegenerate()
}

// doRegenerate runs the sprite generator
func (hr *HotReloader) doRegenerate() {
	defer func() {
		hr.mu.Lock()
		hr.regenerating = false
		hr.mu.Unlock()
	}()

	fmt.Println("Regenerating sprites...")

	cmd := exec.Command("go", "run", "./cmd/spritegen/")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Sprite generation failed: %v\n%s\n", err, output)
		return
	}

	fmt.Println("Sprites regenerated successfully!")

	// Queue reload of all sprite sheets
	hr.reloadQueue <- getAssetPath("claude/spritesheet.png")
	hr.reloadQueue <- getAssetPath("claude/mini_spritesheet.png")
	hr.reloadQueue <- getAssetPath("enemies/enemy_spritesheet.png")
	hr.reloadQueue <- getAssetPath("ui/chest.png")
}

// ForceReloadAll reloads all textures
func (hr *HotReloader) ForceReloadAll() {
	fmt.Println("Force reloading all textures...")

	// Queue all known textures
	hr.reloadQueue <- getAssetPath("claude/spritesheet.png")
	hr.reloadQueue <- getAssetPath("claude/mini_spritesheet.png")
	hr.reloadQueue <- getAssetPath("enemies/enemy_spritesheet.png")
	hr.reloadQueue <- getAssetPath("ui/chest.png")

	// Queue all hats
	for _, name := range hr.renderer.hatNames {
		hr.reloadQueue <- getAssetPath("accessories/hats/" + name + ".png")
	}

	// Queue all faces
	for _, name := range hr.renderer.faceNames {
		hr.reloadQueue <- getAssetPath("accessories/faces/" + name + ".png")
	}
}

// Stop stops the hot reloader
func (hr *HotReloader) Stop() {
	if hr.watcher != nil {
		hr.watcher.Close()
	}
}
