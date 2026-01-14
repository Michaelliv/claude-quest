package main

import (
	"fmt"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 320
	screenHeight = 200
	windowScale  = 2 // Initial scale (640x400 ~ terminal size)
	windowTitle  = "Claude Quest"
)

// getScaledDestRect calculates destination rectangle that maintains aspect ratio and centers content
func getScaledDestRect() rl.Rectangle {
	windowW := float32(rl.GetScreenWidth())
	windowH := float32(rl.GetScreenHeight())

	// Calculate scale to fit while maintaining aspect ratio
	scaleX := windowW / float32(screenWidth)
	scaleY := windowH / float32(screenHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Calculate centered position
	scaledW := float32(screenWidth) * scale
	scaledH := float32(screenHeight) * scale
	offsetX := (windowW - scaledW) / 2
	offsetY := (windowH - scaledH) / 2

	return rl.Rectangle{X: offsetX, Y: offsetY, Width: scaledW, Height: scaledH}
}

func printUsage() {
	fmt.Println(`Claude Quest - RPG Animation Viewer for Claude Code

Usage:
  cq                    Watch the current directory's latest conversation
  cq watch [dir]        Watch a specific directory's conversation
  cq replay <file>      Replay an existing conversation JSONL file
  cq demo               Cycle through all animations (demo mode)

Options:
  -s, --speed <ms>      Replay speed in milliseconds (default: 200)
  -h, --help            Show this help message

Examples:
  cq                                    # Watch current project
  cq watch ~/Projects/myapp             # Watch specific project
  cq replay ~/.claude/projects/-Users-me-Projects-myapp/abc123.jsonl
  cq demo                               # See all animations`)
}

var animationNames = []string{
	"Idle", "Enter", "Casting (Read)", "Attack (Bash)",
	"Writing (Edit)", "Victory", "Hurt (Error)", "Thinking",
}

func runDemo() {
	fmt.Println("Demo mode - cycling through all animations")

	// Enable resizable window
	rl.SetConfigFlags(rl.FlagWindowResizable)

	// Initialize raylib window
	rl.InitWindow(screenWidth*windowScale, screenHeight*windowScale, windowTitle+" - Demo")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	target := rl.LoadRenderTexture(screenWidth, screenHeight)
	defer rl.UnloadRenderTexture(target)

	config := LoadConfig("config.json")
	config.Debug = true
	renderer := NewRenderer(config)
	animations := NewAnimationSystem()

	currentAnim := 0
	animTimer := float32(0)
	animDuration := float32(2.0) // Show each animation for 2 seconds

	// Start with first animation
	animations.HandleEvent(Event{Type: EventType(currentAnim)})

	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()
		animTimer += dt

		// Switch animation every 2 seconds
		if animTimer >= animDuration {
			animTimer = 0
			currentAnim = (currentAnim + 1) % 8

			// Map demo index to event type
			eventTypes := []EventType{
				EventIdle, EventSystemInit, EventReading, EventBash,
				EventWriting, EventSuccess, EventError, EventThinking,
			}
			animations.HandleEvent(Event{Type: eventTypes[currentAnim]})
		}

		animations.Update(dt)

		// Handle keyboard input for accessories
		// Up/Down = switch row, Left/Right = cycle value
		if rl.IsKeyPressed(rl.KeyUp) {
			renderer.SwitchRow(-1)
		}
		if rl.IsKeyPressed(rl.KeyDown) {
			renderer.SwitchRow(1)
		}
		if rl.IsKeyPressed(rl.KeyLeft) {
			renderer.CycleActive(-1)
		}
		if rl.IsKeyPressed(rl.KeyRight) {
			renderer.CycleActive(1)
		}

		// Render
		rl.BeginTextureMode(target)
		rl.ClearBackground(rl.Color{R: 24, G: 20, B: 37, A: 255})
		renderer.Draw(animations.GetState())
		renderer.DrawAccessoryPicker()
		rl.EndTextureMode()

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		sourceRec := rl.Rectangle{X: 0, Y: float32(screenHeight), Width: float32(screenWidth), Height: -float32(screenHeight)}
		destRec := getScaledDestRect()
		rl.DrawTexturePro(target.Texture, sourceRec, destRec, rl.Vector2{}, 0, rl.White)
		rl.EndDrawing()
	}
}

func main() {
	watcher := NewWatcher()
	var err error

	// Parse command line arguments
	args := os.Args[1:]

	if len(args) == 0 {
		// Default: watch current directory
		cwd, _ := os.Getwd()
		err = watcher.FindProjectConversation(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		err = watcher.StartLive()
	} else {
		switch args[0] {
		case "-h", "--help", "help":
			printUsage()
			os.Exit(0)

		case "demo":
			runDemo()
			os.Exit(0)

		case "watch":
			dir := "."
			if len(args) > 1 {
				dir = args[1]
			}
			err = watcher.FindProjectConversation(dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			err = watcher.StartLive()

		case "replay":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Error: replay requires a file path")
				printUsage()
				os.Exit(1)
			}
			filePath := args[1]

			// Check for speed flag
			for i := 2; i < len(args); i++ {
				if args[i] == "-s" || args[i] == "--speed" {
					if i+1 < len(args) {
						var speed int
						fmt.Sscanf(args[i+1], "%d", &speed)
						if speed > 0 {
							watcher.ReplaySpeed = time.Duration(speed) * time.Millisecond
						}
					} else {
						// -s without value means 2x speed
						watcher.ReplaySpeed = 100 * time.Millisecond
					}
				}
			}

			err = watcher.StartReplay(filePath)

		default:
			// Assume it's a directory path for watching
			err = watcher.FindProjectConversation(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			err = watcher.StartLive()
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Watching: %s\n", watcher.FilePath)

	// Enable resizable window
	rl.SetConfigFlags(rl.FlagWindowResizable)

	// Initialize raylib window
	rl.InitWindow(screenWidth*windowScale, screenHeight*windowScale, windowTitle)
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	// Create render texture for pixel-perfect scaling
	target := rl.LoadRenderTexture(screenWidth, screenHeight)
	defer rl.UnloadRenderTexture(target)

	// Initialize game systems
	config := LoadConfig("config.json")
	renderer := NewRenderer(config)
	animations := NewAnimationSystem()

	for !rl.WindowShouldClose() {
		// Process any pending events from the watcher
		select {
		case event := <-watcher.Events:
			animations.HandleEvent(event)
		default:
		}

		// Update animation state
		animations.Update(rl.GetFrameTime())

		// Handle keyboard input for accessories
		// Up/Down = switch row, Left/Right = cycle value
		if rl.IsKeyPressed(rl.KeyUp) {
			renderer.SwitchRow(-1)
		}
		if rl.IsKeyPressed(rl.KeyDown) {
			renderer.SwitchRow(1)
		}
		if rl.IsKeyPressed(rl.KeyLeft) {
			renderer.CycleActive(-1)
		}
		if rl.IsKeyPressed(rl.KeyRight) {
			renderer.CycleActive(1)
		}

		// Render to texture at native resolution
		rl.BeginTextureMode(target)
		rl.ClearBackground(rl.Color{R: 24, G: 20, B: 37, A: 255}) // Dark purple bg
		renderer.Draw(animations.GetState())
		renderer.DrawAccessoryPicker()
		rl.EndTextureMode()

		// Draw scaled texture to window
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// Flip texture vertically (raylib render textures are flipped)
		sourceRec := rl.Rectangle{X: 0, Y: float32(screenHeight), Width: float32(screenWidth), Height: -float32(screenHeight)}
		destRec := getScaledDestRect()
		rl.DrawTexturePro(target.Texture, sourceRec, destRec, rl.Vector2{}, 0, rl.White)

		rl.EndDrawing()
	}
}
