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
	maxTokens    = 200000 // Opus 4.5 context window
)

// ThrownTool represents a tool name being thrown forward
type ThrownTool struct {
	Text    string
	X, Y    float32
	VX, VY  float32
	Life    float32
	MaxLife float32
	Color   uint32 // Packed RGBA
}

// GameState tracks UI state for quest text, mana bar, todos
type GameState struct {
	// Quest display
	QuestText   string
	QuestTimer  float32
	QuestFade   float32

	// Mana bar (context window)
	ManaTotal   int
	ManaMax     int
	ManaDisplay float32 // Smoothly animated value

	// Todos
	Todos       []TodoItem

	// Effects
	ThinkHardActive bool
	ThinkHardTimer  float32
	ThinkLevel      ThinkLevel
	CompactActive   bool
	CompactTimer    float32

	// Activity tracking - for walk/scroll during activity only
	LastActivityTime float32
	IsActive         bool

	// Thrown tools effect
	ThrownTools []ThrownTool
}

// NewGameState creates a new game state
func NewGameState() *GameState {
	return &GameState{
		ManaMax:     maxTokens,
		ManaDisplay: 0,
	}
}

// Update updates game state animations
func (g *GameState) Update(dt float32) {
	// Animate quest fade
	if g.QuestText != "" {
		g.QuestTimer += dt
		if g.QuestTimer < 0.5 {
			// Fade in
			g.QuestFade = g.QuestTimer / 0.5
		} else if g.QuestTimer < 8.0 {
			// Full display
			g.QuestFade = 1.0
		} else if g.QuestTimer < 9.0 {
			// Fade out
			g.QuestFade = 1.0 - (g.QuestTimer - 8.0)
		} else {
			// Clear
			g.QuestText = ""
			g.QuestTimer = 0
			g.QuestFade = 0
		}
	}

	// Smooth mana animation
	target := float32(g.ManaTotal)
	if g.ManaDisplay < target {
		g.ManaDisplay += (target - g.ManaDisplay) * dt * 3
	} else if g.ManaDisplay > target {
		g.ManaDisplay -= (g.ManaDisplay - target) * dt * 3
	}

	// Think hard effect timer
	if g.ThinkHardActive {
		g.ThinkHardTimer += dt
		if g.ThinkHardTimer > 3.0 {
			g.ThinkHardActive = false
			g.ThinkHardTimer = 0
		}
	}

	// Compact effect timer
	if g.CompactActive {
		g.CompactTimer += dt
		if g.CompactTimer > 2.0 {
			g.CompactActive = false
			g.CompactTimer = 0
		}
	}

	// Activity timeout - go inactive after 60 seconds of no events
	// (keeps walking during thinking pauses)
	if g.IsActive {
		g.LastActivityTime += dt
		if g.LastActivityTime > 60.0 {
			g.IsActive = false
		}
	}

	// Update thrown tools
	alive := g.ThrownTools[:0]
	for i := range g.ThrownTools {
		t := &g.ThrownTools[i]
		t.Life += dt
		if t.Life < t.MaxLife {
			// Move forward and arc with strong gravity
			t.X += t.VX * dt
			t.Y += t.VY * dt
			t.VY += 180 * dt // Stronger gravity for bigger arc
			alive = append(alive, *t)
		}
	}
	g.ThrownTools = alive
}

// ThrowTool creates a thrown tool effect with random direction
func (g *GameState) ThrowTool(toolName string, color uint32) {
	// Start from Claude's position with slight random offset
	startX := float32(screenWidth/2) + (randFloat()*20 - 10)
	startY := float32(screenHeight/2-10) + (randFloat()*10 - 5)

	// Random angle: spread in all upward directions
	// -160 to -20 degrees (full upper arc, both left and right)
	angle := (-20 - randFloat()*140) * 3.14159 / 180 // Convert to radians
	// Randomly flip to go forward (right) or backward (left)
	if randFloat() > 0.5 {
		angle = -angle // Flip to right side
	}
	speed := float32(110 + randFloat()*50) // 110-160 speed (faster for bigger arc)

	baseVX := speed * float32(simpleCosF(float64(angle)))
	baseVY := speed * float32(simpleSinF(float64(angle)))

	tool := ThrownTool{
		Text:    toolName,
		X:       startX,
		Y:       startY,
		VX:      baseVX,
		VY:      baseVY,
		Life:    0,
		MaxLife: 1.5,
		Color:   color,
	}
	g.ThrownTools = append(g.ThrownTools, tool)
}

// Simple random float 0-1
var randSeed uint32 = 12345

func randFloat() float32 {
	randSeed = randSeed*1103515245 + 12345
	return float32(randSeed&0x7FFFFFFF) / float32(0x7FFFFFFF)
}

// Tool colors (packed RGBA)
const (
	colorBash    = 0xFF6B6BFF // Red - attack
	colorRead    = 0x6BB5FFFF // Blue - magic
	colorWrite   = 0x6BFF6BFF // Green - creation
	colorWeb     = 0xFFD93DFF // Yellow - search
	colorAgent   = 0xDA6BFFFF // Purple - summon
	colorDefault = 0xAAAAAAFF // Gray
)

// HandleEvent updates game state based on events
func (g *GameState) HandleEvent(event Event) {
	// Mark activity for any real event (not idle)
	if event.Type != EventIdle {
		g.LastActivityTime = 0
		g.IsActive = true
	}

	// Update mana from token usage
	if event.TokenUsage != nil {
		g.ManaTotal = event.TokenUsage.Total()
	}

	// Throw tool name for tool events
	if event.ToolName != "" {
		var color uint32
		switch event.Type {
		case EventBash:
			color = colorBash
		case EventReading:
			if event.ToolName == "WebSearch" || event.ToolName == "WebFetch" {
				color = colorWeb
			} else {
				color = colorRead
			}
		case EventWriting:
			color = colorWrite
		case EventSpawnAgent:
			color = colorAgent
		default:
			color = colorDefault
		}
		g.ThrowTool(event.ToolName, color)
	}

	// Handle specific event types
	switch event.Type {
	case EventQuest:
		g.QuestText = event.Details
		g.QuestTimer = 0
		g.QuestFade = 0

	case EventThinkHard:
		g.QuestText = event.Details
		g.QuestTimer = 0
		g.ThinkHardActive = true
		g.ThinkHardTimer = 0
		g.ThinkLevel = event.ThinkLevel
		if g.ThinkLevel == ThinkNone {
			g.ThinkLevel = ThinkHard // Default if not specified
		}

	case EventCompact:
		g.CompactActive = true
		g.CompactTimer = 0
		// Reset mana after compact
		g.ManaTotal = 0

	case EventTodoUpdate:
		if event.TodoItems != nil {
			g.Todos = event.TodoItems
		}
	}
}

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
	fmt.Println("Keys: Q=quest, M=mana, C=compact, W=walk mode")
	fmt.Println("Think: 1=think, 2=think hard, 3=think harder, 4=ULTRATHINK")

	// Enable resizable window
	rl.SetConfigFlags(rl.FlagWindowResizable)

	// Initialize raylib window
	rl.InitWindow(screenWidth*windowScale, screenHeight*windowScale, windowTitle+" - Demo")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	target := rl.LoadRenderTexture(screenWidth, screenHeight)
	defer rl.UnloadRenderTexture(target)

	config := LoadConfig("config.json")
	config.Debug = false // Disable debug info in demo
	renderer := NewRenderer(config)
	animations := NewAnimationSystem()
	gameState := NewGameState()

	currentAnim := 0
	animTimer := float32(0)
	animDuration := float32(2.0) // Show each animation for 2 seconds

	// Demo quests
	demoQuests := []string{
		"help me implement user authentication",
		"fix the bug in the login form",
		"add dark mode support",
		"optimize database queries",
	}
	questIndex := 0

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
			event := Event{Type: eventTypes[currentAnim]}

			// Add token usage to some events
			if currentAnim > 0 {
				event.TokenUsage = &TokenUsage{
					InputTokens:         10000 + currentAnim*15000,
					CacheReadTokens:     20000 + currentAnim*10000,
					CacheCreationTokens: 5000,
				}
			}

			animations.HandleEvent(event)
			gameState.HandleEvent(event)
		}

		animations.Update(dt)
		gameState.Update(dt)
		renderer.UpdateScroll(dt)

		// Handle keyboard input for accessories
		if rl.IsKeyPressed(rl.KeyUp) {
			renderer.SwitchRow(-1)
		}
		if rl.IsKeyPressed(rl.KeyDown) {
			renderer.SwitchRow(1)
		}
		// Toggle walk mode
		if rl.IsKeyPressed(rl.KeyW) {
			renderer.ToggleWalkMode()
			animations.SetWalkMode(renderer.IsWalkMode())
		}
		if rl.IsKeyPressed(rl.KeyLeft) {
			renderer.CycleActive(-1)
			animations.SetWalkMode(renderer.IsWalkMode()) // Sync walk mode
		}
		if rl.IsKeyPressed(rl.KeyRight) {
			renderer.CycleActive(1)
			animations.SetWalkMode(renderer.IsWalkMode()) // Sync walk mode
		}

		// Demo triggers
		if rl.IsKeyPressed(rl.KeyQ) {
			// Show quest
			gameState.HandleEvent(Event{
				Type:    EventQuest,
				Details: demoQuests[questIndex%len(demoQuests)],
			})
			questIndex++
		}
		if rl.IsKeyPressed(rl.KeyM) {
			// Increase mana
			gameState.ManaTotal += 25000
			if gameState.ManaTotal > gameState.ManaMax {
				gameState.ManaTotal = 25000
			}
		}
		// Think levels: 1, 2, 3, 4
		if rl.IsKeyPressed(rl.KeyOne) {
			gameState.HandleEvent(Event{Type: EventThinkHard, Details: "really think about this", ThinkLevel: ThinkNormal})
			animations.HandleEvent(Event{Type: EventThinkHard})
		}
		if rl.IsKeyPressed(rl.KeyTwo) {
			gameState.HandleEvent(Event{Type: EventThinkHard, Details: "think hard about this problem", ThinkLevel: ThinkHard})
			animations.HandleEvent(Event{Type: EventThinkHard})
		}
		if rl.IsKeyPressed(rl.KeyThree) {
			gameState.HandleEvent(Event{Type: EventThinkHard, Details: "think harder! this is complex", ThinkLevel: ThinkHarder})
			animations.HandleEvent(Event{Type: EventThinkHard})
		}
		if rl.IsKeyPressed(rl.KeyFour) {
			gameState.HandleEvent(Event{Type: EventThinkHard, Details: "ULTRATHINK mode activated!", ThinkLevel: ThinkUltra})
			animations.HandleEvent(Event{Type: EventThinkHard})
		}
		if rl.IsKeyPressed(rl.KeyC) {
			// Compact effect
			gameState.HandleEvent(Event{Type: EventCompact})
			animations.HandleEvent(Event{Type: EventCompact})
		}

		// Render
		rl.BeginTextureMode(target)
		rl.ClearBackground(rl.Color{R: 24, G: 20, B: 37, A: 255})
		renderer.Draw(animations.GetState())
		renderer.DrawGameUI(gameState)
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
	gameState := NewGameState()

	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()

		// Process any pending events from the watcher
		select {
		case event := <-watcher.Events:
			animations.HandleEvent(event)
			gameState.HandleEvent(event)
		default:
		}

		// Update systems
		animations.Update(dt)
		gameState.Update(dt)

		// Sync activity state to animation system
		animations.SetActive(gameState.IsActive)

		// Only scroll when there's activity (events coming in)
		if gameState.IsActive {
			renderer.UpdateScroll(dt)
		}

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
			animations.SetWalkMode(renderer.IsWalkMode()) // Sync walk mode
		}
		if rl.IsKeyPressed(rl.KeyRight) {
			renderer.CycleActive(1)
			animations.SetWalkMode(renderer.IsWalkMode()) // Sync walk mode
		}
		// Toggle walk mode with W
		if rl.IsKeyPressed(rl.KeyW) {
			renderer.ToggleWalkMode()
			animations.SetWalkMode(renderer.IsWalkMode())
		}

		// Render to texture at native resolution
		rl.BeginTextureMode(target)
		rl.ClearBackground(rl.Color{R: 24, G: 20, B: 37, A: 255}) // Dark purple bg
		renderer.Draw(animations.GetState())
		renderer.DrawGameUI(gameState)
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
