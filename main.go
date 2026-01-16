package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// MiniAnimType represents mini Claude animation types
type MiniAnimType int

const (
	MiniAnimSpawn MiniAnimType = iota
	MiniAnimIdle
	MiniAnimWalk
	MiniAnimPoof
)

// MiniAgent represents a mini Claude spawned for a subagent
type MiniAgent struct {
	ID        string       // Unique agent ID
	Name      string       // Agent type name to display
	X, Y      float32      // Position (landing spot)
	TargetX   float32      // Target X for landing
	Animation MiniAnimType // Current animation
	Frame     int          // Current frame
	Timer     float32      // Animation timer
	SpawnVY   float32      // Vertical velocity during spawn jump
}

// EnemyType represents different enemy sprites
type EnemyType int

const (
	EnemyBug EnemyType = iota
	EnemyError
	EnemyLowContext
)

// FlyingEnemy represents an enemy flying toward Claude
type FlyingEnemy struct {
	Type    EnemyType
	X, Y    float32 // Current position
	VX, VY  float32 // Velocity (VX negative = moving left, VY affected by gravity)
	Frame   int     // Animation frame
	Timer   float32 // Animation timer
	Hit     bool    // Has it hit Claude?
	Impact  float32 // Impact effect timer (> 0 means showing impact)
}

// FloatingXP represents a floating "+XP" indicator
type FloatingXP struct {
	Amount  int
	X, Y    float32
	Timer   float32
	MaxLife float32
}

// GameState tracks UI state for quest text, mana bar, todos
type GameState struct {
	// Quest display
	QuestText  string
	QuestTimer float32
	QuestFade  float32

	// Mana bar (context window)
	ManaTotal   int
	ManaMax     int
	ManaDisplay float32 // Smoothly animated value

	// Todos
	Todos []TodoItem

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

	// Mini agents (subagents displayed as mini Claudes)
	MiniAgents []MiniAgent

	// Flying enemies (attack Claude on errors)
	FlyingEnemies []FlyingEnemy
	PendingHurt   bool // Set when enemy hits, triggers hurt animation

	// Thought bubble display
	ThoughtText  string  // Current thought text to display
	ThoughtTimer float32 // Timer for thought display (6 seconds)
	ThoughtFade  float32 // Fade in/out animation

	// SHIPPED! rainbow effect (git push celebration)
	ShippedActive bool    // Whether the SHIPPED effect is playing
	ShippedTimer  float32 // Animation timer

	// Progression system
	Profile *CareerProfile // Persistent career data
	Session SessionStats   // Current session stats

	// Level up / chest state
	PendingLevelUp    bool            // True when level up occurred, triggers chest
	PendingBonusChest bool            // True when bonus chest triggered
	BonusChestReason  string          // Why bonus chest was triggered
	ActiveChest       *TreasureChest  // Currently active treasure chest (nil if none)

	// Floating XP indicators
	FloatingXPs []FloatingXP
}

// NewGameState creates a new game state
func NewGameState() *GameState {
	profile := LoadProfile()
	profile.SessionsStarted++
	profile.Save()

	return &GameState{
		ManaMax:     maxTokens,
		ManaDisplay: 0,
		Profile:     profile,
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

	// SHIPPED! rainbow effect animation
	if g.ShippedActive {
		g.ShippedTimer += dt
		// Banner flies in an arc - extended to 4.5s so tail can exit smoothly
		if g.ShippedTimer > 4.5 {
			g.ShippedActive = false
			g.ShippedTimer = 0
		}
	}

	// Thought bubble timer (12 second display with fade in/out)
	if g.ThoughtText != "" {
		g.ThoughtTimer += dt
		if g.ThoughtTimer < 0.4 {
			// Fade in
			g.ThoughtFade = g.ThoughtTimer / 0.4
		} else if g.ThoughtTimer < 11.6 {
			// Full display
			g.ThoughtFade = 1.0
		} else if g.ThoughtTimer < 12.0 {
			// Fade out
			g.ThoughtFade = 1.0 - (g.ThoughtTimer - 11.6) / 0.4
		} else {
			// Clear
			g.ThoughtText = ""
			g.ThoughtTimer = 0
			g.ThoughtFade = 0
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

	// Update mini agents
	g.updateMiniAgents(dt)

	// Update flying enemies
	g.updateFlyingEnemies(dt)

	// Update flow meter decay (only decays when no activity)
	if g.Profile != nil {
		if g.Session.FlowDecayTimer > 0 {
			g.Session.FlowDecayTimer += dt
		}
		if g.Session.FlowDecayTimer > 5.0 {
			g.Session.FlowMeter -= dt * 0.03
			if g.Session.FlowMeter < 0 {
				g.Session.FlowMeter = 0
			}
		}
	}

	// Update floating XP indicators
	g.updateFloatingXPs(dt)

	// Spawn treasure chest if pending and no active chest
	if g.ActiveChest == nil {
		if g.PendingLevelUp {
			g.ActiveChest = NewLevelUpChest(g.Profile)
			g.PendingLevelUp = false
		} else if g.PendingBonusChest {
			g.ActiveChest = NewBonusChest(g.Profile, g.BonusChestReason)
			g.PendingBonusChest = false
			g.BonusChestReason = ""
		}
	}

	// Update active chest
	if g.ActiveChest != nil {
		g.ActiveChest.Update(dt)

		// Handle chest completion
		if g.ActiveChest.IsDone() {
			if g.ActiveChest.ClaimedItem != nil {
				g.Profile.ClaimItem(g.ActiveChest.ClaimedItem.ID)
				g.Profile.Save()
			} else if !g.ActiveChest.HasItems() {
				// Empty pool - grant bonus XP instead
				g.Profile.AddXP(500)
				g.Profile.Save()
				g.SpawnFloatingXP(500)
			}
			g.ActiveChest = nil
		}
	}
}

// SpawnFloatingXP creates a new floating XP indicator above Claude's head
func (g *GameState) SpawnFloatingXP(amount int) {
	// Spawn position: above Claude's head with some randomness
	baseX := float32(screenWidth/2 - 10)
	baseY := float32(70) // Above Claude
	offsetX := float32((len(g.FloatingXPs) % 3) - 1) * 15 // Spread out if multiple

	g.FloatingXPs = append(g.FloatingXPs, FloatingXP{
		Amount:  amount,
		X:       baseX + offsetX,
		Y:       baseY,
		Timer:   0,
		MaxLife: 1.5, // 1.5 seconds to float and fade
	})
}

// updateFloatingXPs updates floating XP indicator animations
func (g *GameState) updateFloatingXPs(dt float32) {
	alive := g.FloatingXPs[:0]
	for i := range g.FloatingXPs {
		xp := &g.FloatingXPs[i]
		xp.Timer += dt
		xp.Y -= dt * 25 // Float upward

		if xp.Timer < xp.MaxLife {
			alive = append(alive, *xp)
		}
	}
	g.FloatingXPs = alive
}

// Mini Claude animation frame counts: Spawn=8, Idle=8, Walk=8, Poof=6
var miniFrameCounts = []int{8, 8, 8, 6}

// updateMiniAgents updates all mini agent animations
func (g *GameState) updateMiniAgents(dt float32) {
	frameDuration := float32(0.08) // Slightly slower animation for mini

	aliveAgents := g.MiniAgents[:0]
	for i := range g.MiniAgents {
		m := &g.MiniAgents[i]
		m.Timer += dt

		// Advance frame
		if m.Timer >= frameDuration {
			m.Timer -= frameDuration
			m.Frame++

			// Check if animation completed
			maxFrames := miniFrameCounts[int(m.Animation)]
			if m.Frame >= maxFrames {
				m.Frame = 0

				switch m.Animation {
				case MiniAnimSpawn:
					// Spawn complete - go to idle or walk randomly
					if randFloat() > 0.5 {
						m.Animation = MiniAnimIdle
					} else {
						m.Animation = MiniAnimWalk
					}
				case MiniAnimPoof:
					// Poof complete - remove agent
					continue // Don't add to alive list
				case MiniAnimIdle, MiniAnimWalk:
					// Loop - occasionally switch between idle/walk
					if randFloat() > 0.9 {
						if m.Animation == MiniAnimIdle {
							m.Animation = MiniAnimWalk
						} else {
							m.Animation = MiniAnimIdle
						}
					}
				}
			}
		}

		// Update position during spawn animation (jumping arc)
		if m.Animation == MiniAnimSpawn {
			// Move toward target X
			dx := m.TargetX - m.X
			if dx > 0.5 || dx < -0.5 {
				m.X += dx * dt * 3
			}
			// Arc motion - use frame to determine Y offset
			// Frames 0-3: going up, 4-7: coming down
			if m.Frame < 4 {
				m.Y = 165 - float32(m.Frame)*8 // Rise
			} else {
				m.Y = 165 - float32(7-m.Frame)*8 // Fall
			}
		}

		aliveAgents = append(aliveAgents, *m)
	}
	g.MiniAgents = aliveAgents
}

// Enemy animation frame counts: Bug=4, Error=4, LowContext=4
const enemyFrameCount = 4

// updateFlyingEnemies updates all flying enemies
func (g *GameState) updateFlyingEnemies(dt float32) {
	frameDuration := float32(0.1)  // Animation speed
	claudeX := float32(screenWidth / 2)
	claudeY := float32(screenHeight/2 + 10) // Claude's center
	gravity := float32(120)        // Gravity strength

	aliveEnemies := g.FlyingEnemies[:0]
	for i := range g.FlyingEnemies {
		e := &g.FlyingEnemies[i]

		// Animate
		e.Timer += dt
		if e.Timer >= frameDuration {
			e.Timer -= frameDuration
			e.Frame = (e.Frame + 1) % enemyFrameCount
		}

		// Update impact effect
		if e.Impact > 0 {
			e.Impact -= dt
			if e.Impact <= 0 {
				// Impact done, trigger hurt
				g.PendingHurt = true
			}
		}

		// Move with gravity arc
		if !e.Hit {
			e.X += e.VX * dt
			e.Y += e.VY * dt
			e.VY += gravity * dt // Apply gravity

			// Check if hit Claude (within hitbox)
			dx := e.X - claudeX
			dy := e.Y - claudeY
			if dx < 30 && dx > -30 && dy < 30 && dy > -30 {
				e.Hit = true
				e.Impact = 0.3 // Show impact for 0.3 seconds
				e.VX = 0
				e.VY = 0
			}

			// Remove if off screen bottom
			if e.Y > screenHeight+50 {
				continue
			}
		} else {
			// After hit, fade out
			if e.Impact <= 0 {
				continue // Remove after impact done
			}
		}

		aliveEnemies = append(aliveEnemies, *e)
	}
	g.FlyingEnemies = aliveEnemies
}

// SpawnEnemy creates a flying enemy that attacks Claude
func (g *GameState) SpawnEnemy(enemyType EnemyType) {
	// Start from right side of screen at varied heights
	startX := float32(screenWidth + 30)

	// Random starting height: top, middle, or bottom third
	heightZone := randFloat()
	var startY float32
	var initialVY float32

	if heightZone < 0.33 {
		// High throw - starts high, arcs down
		startY = 20 + randFloat()*40
		initialVY = 20 + randFloat()*30
	} else if heightZone < 0.66 {
		// Middle throw - starts mid, slight arc
		startY = 60 + randFloat()*40
		initialVY = -20 + randFloat()*40
	} else {
		// Low throw - starts low, arcs up then down
		startY = 120 + randFloat()*40
		initialVY = -60 - randFloat()*40
	}

	// Horizontal speed toward Claude
	vx := float32(-140 - randFloat()*60) // Speed: 140-200 pixels/sec

	enemy := FlyingEnemy{
		Type:   enemyType,
		X:      startX,
		Y:      startY,
		VX:     vx,
		VY:     initialVY,
		Frame:  0,
		Timer:  0,
		Hit:    false,
		Impact: 0,
	}
	g.FlyingEnemies = append(g.FlyingEnemies, enemy)
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

// SpawnMiniAgent creates a new mini Claude for a subagent
func (g *GameState) SpawnMiniAgent(agentType string) {
	// Generate unique ID
	id := fmt.Sprintf("agent-%d", len(g.MiniAgents)+1)

	// Start position: at big Claude's feet
	startX := float32(screenWidth / 2)
	startY := float32(165) // Ground level

	// Target position: random spot to left or right of big Claude
	// Spread out based on how many agents already exist
	offset := float32(40 + len(g.MiniAgents)*25) // 40-90+ pixels away
	if randFloat() > 0.5 {
		offset = -offset // Go left
	}
	targetX := startX + offset

	// Clamp to screen bounds (leave some margin)
	if targetX < 30 {
		targetX = 30
	} else if targetX > screenWidth-30 {
		targetX = screenWidth - 30
	}

	mini := MiniAgent{
		ID:        id,
		Name:      agentType,
		X:         startX,
		Y:         startY,
		TargetX:   targetX,
		Animation: MiniAnimSpawn,
		Frame:     0,
		Timer:     0,
	}
	g.MiniAgents = append(g.MiniAgents, mini)
}

// PoofMiniAgent triggers the poof animation for an agent by ID
func (g *GameState) PoofMiniAgent(agentID string) {
	for i := range g.MiniAgents {
		if g.MiniAgents[i].ID == agentID || agentID == "" {
			// If no specific ID, poof the oldest agent
			g.MiniAgents[i].Animation = MiniAnimPoof
			g.MiniAgents[i].Frame = 0
			g.MiniAgents[i].Timer = 0
			if agentID != "" {
				return
			}
		}
	}
	// If no ID given and agents exist, poof the first one
	if agentID == "" && len(g.MiniAgents) > 0 {
		g.MiniAgents[0].Animation = MiniAnimPoof
		g.MiniAgents[0].Frame = 0
		g.MiniAgents[0].Timer = 0
	}
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

		// Update flow meter on activity
		if g.Profile != nil {
			g.Session.FlowDecayTimer = 0.001 // Start decay timer (small non-zero to indicate active)
			g.Session.FlowMeter += 0.05
			if g.Session.FlowMeter >= 1.0 {
				g.Session.FlowMeter = 1.0
				if !g.Session.FlowPeakReached {
					g.Session.FlowPeakReached = true
					// Grant XP for flow peak
					if g.Profile.RecordFlowPeak() {
						g.PendingLevelUp = true
					}
					g.Profile.Save()
				}
			}
			g.Session.TotalToolCalls++
		}
	}

	// Update mana from token usage
	if event.TokenUsage != nil {
		g.ManaTotal = event.TokenUsage.Total()
		if g.Profile != nil {
			g.Profile.RecordTokens(event.TokenUsage.Total())
		}
	}

	// Track progression based on event type
	if g.Profile != nil {
		leveledUp := false

		switch event.Type {
		case EventReading:
			g.Session.Reads++
			leveledUp = g.Profile.RecordRead()
			g.SpawnFloatingXP(XPRead)

		case EventWriting:
			g.Session.Writes++
			leveledUp = g.Profile.RecordWrite()
			g.SpawnFloatingXP(XPWrite)

		case EventBash:
			success := !event.IsError
			g.Session.RecordBashResult(success)
			leveledUp = g.Profile.RecordBash(success, g.Session.CurrentBashStreak)
			if success {
				xp := XPBashSuccess
				if g.Session.CurrentBashStreak > 1 {
					xp += XPStreakBonus
				}
				g.SpawnFloatingXP(xp)
			} else {
				g.SpawnFloatingXP(XPBashFail)
			}

		case EventThinkHard:
			leveledUp = g.Profile.RecordThinking(event.ThinkLevel)
			xp := XPThinkNormal
			switch event.ThinkLevel {
			case ThinkHard:
				xp = XPThinkHard
			case ThinkHarder:
				xp = XPThinkHard + XPThinkBonus
			case ThinkUltra:
				xp = XPThinkHard + XPThinkBonus*2
			}
			g.SpawnFloatingXP(xp)

		case EventAgentComplete:
			leveledUp = g.Profile.RecordAgentComplete()
			g.SpawnFloatingXP(XPAgentComplete)

		case EventTodoUpdate:
			// Count newly completed todos
			if event.TodoItems != nil {
				for _, todo := range event.TodoItems {
					if todo.Status == "completed" {
						// Check if this is a new completion
						wasCompleted := false
						for _, oldTodo := range g.Todos {
							if oldTodo.Content == todo.Content && oldTodo.Status == "completed" {
								wasCompleted = true
								break
							}
						}
						if !wasCompleted {
							g.Session.TodosCompleted++
							if g.Profile.RecordTodoComplete() {
								leveledUp = true
							}
							g.SpawnFloatingXP(XPTodoComplete)
						}
					}
				}
			}
		}

		if leveledUp {
			g.PendingLevelUp = true
		}

		// Check for bonus chest triggers
		if !g.Session.BonusChestAwarded {
			if triggered, reason := g.Session.CheckBonusChest(); triggered {
				g.PendingBonusChest = true
				g.BonusChestReason = reason
				g.Profile.BonusChestsFound++
			}
		}

		// Save profile after changes
		g.Profile.Save()
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

	case EventThinking:
		// Display thought in thought bubble if we have content
		if event.ThoughtText != "" {
			g.ThoughtText = event.ThoughtText
			g.ThoughtTimer = 0
			g.ThoughtFade = 0
		}

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

	case EventSpawnAgent:
		// Extract agent type from details (format: "Agent: typename")
		agentType := event.Details
		if len(agentType) > 7 && agentType[:7] == "Agent: " {
			agentType = agentType[7:]
		}
		g.SpawnMiniAgent(agentType)

	case EventAgentComplete:
		// Agent finished - poof the oldest mini agent
		g.PoofMiniAgent("")

	case EventError:
		// Spawn a bug or ERROR enemy
		if randFloat() > 0.5 {
			g.SpawnEnemy(EnemyBug)
		} else {
			g.SpawnEnemy(EnemyError)
		}

	case EventGitPush:
		// SHIPPED! - trigger epic rainbow banner effect
		g.ShippedActive = true
		g.ShippedTimer = 0
	}

	// Check for low context - spawn LOW CTX enemy when below 20%
	if g.ManaTotal > 0 {
		usedRatio := float32(g.ManaTotal) / float32(g.ManaMax)
		if usedRatio > 0.8 && randFloat() > 0.9 {
			// Only spawn occasionally when context is very low
			g.SpawnEnemy(EnemyLowContext)
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
  cq studio             Studio mode - asset dev environment (requires -tags debug build)
  cq doctor             Check if Claude Quest can run properly

Options:
  -s, --speed <ms>      Replay speed in milliseconds (default: 200)
  -h, --help            Show this help message

Examples:
  cq                                    # Watch current project
  cq watch ~/Projects/myapp             # Watch specific project
  cq replay ~/.claude/projects/-Users-me-Projects-myapp/abc123.jsonl
  go build -tags debug && ./cq studio   # Studio mode for asset development`)
}

// runDoctor checks if all requirements for Claude Quest are met
func runDoctor() {
	fmt.Println("Claude Quest Doctor")
	fmt.Println("===================")
	fmt.Println()

	allGood := true
	home := os.Getenv("HOME")

	// Check Claude Code installation
	fmt.Println("Claude Code:")

	claudeDir := filepath.Join(home, ".claude")
	if _, err := os.Stat(claudeDir); err == nil {
		fmt.Println("  [OK] ~/.claude/ exists")
	} else {
		fmt.Println("  [!!] ~/.claude/ not found - is Claude Code installed?")
		allGood = false
	}

	projectsDir := filepath.Join(home, ".claude", "projects")
	if _, err := os.Stat(projectsDir); err == nil {
		fmt.Println("  [OK] ~/.claude/projects/ exists")

		// Count project directories
		entries, _ := os.ReadDir(projectsDir)
		projectCount := 0
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), "-") {
				projectCount++
			}
		}
		fmt.Printf("  [OK] Found %d project(s)\n", projectCount)
	} else {
		fmt.Println("  [!!] ~/.claude/projects/ not found")
		allGood = false
	}

	// Check current project
	fmt.Println()
	fmt.Println("Current Project:")

	cwd, _ := os.Getwd()
	absPath, _ := filepath.Abs(cwd)
	encoded := strings.ReplaceAll(absPath, "/", "-")
	projectDir := filepath.Join(projectsDir, encoded)

	fmt.Printf("  Path: %s\n", cwd)
	fmt.Printf("  Encoded: %s\n", encoded)

	if _, err := os.Stat(projectDir); err == nil {
		fmt.Println("  [OK] Project directory exists")

		// Count JSONL files
		entries, _ := os.ReadDir(projectDir)
		jsonlCount := 0
		var latestFile string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") && !strings.HasPrefix(e.Name(), "agent-") {
				jsonlCount++
				latestFile = e.Name()
			}
		}
		if jsonlCount > 0 {
			fmt.Printf("  [OK] Found %d conversation file(s)\n", jsonlCount)
			fmt.Printf("  [OK] Latest: %s\n", latestFile)

			// Validate JSONL structure
			latestPath := filepath.Join(projectDir, latestFile)
			if valid, details := validateJSONL(latestPath); valid {
				fmt.Println("  [OK] JSONL structure valid")
				fmt.Printf("       %s\n", details)
			} else {
				fmt.Println("  [!!] JSONL structure invalid")
				fmt.Printf("       %s\n", details)
				allGood = false
			}
		} else {
			fmt.Println("  [!!] No conversation files found")
			allGood = false
		}
	} else {
		fmt.Println("  [--] No conversations for this project yet")
		fmt.Println("       (Run Claude Code here first)")
	}

	// Check assets
	fmt.Println()
	fmt.Println("Assets:")

	assetChecks := []struct {
		name string
		path string
	}{
		{"Sprite sheet", "claude/spritesheet.png"},
		{"Wizard hat", "accessories/hats/wizard.png"},
		{"Party hat", "accessories/hats/party.png"},
		{"Deal-with-it glasses", "accessories/faces/dealwithit.png"},
	}

	for _, check := range assetChecks {
		assetPath := getAssetPathForDoctor(check.path)
		if _, err := os.Stat(assetPath); err == nil {
			fmt.Printf("  [OK] %s\n", check.name)
		} else {
			fmt.Printf("  [!!] %s not found\n", check.name)
			fmt.Printf("       Looked in: %s\n", assetPath)
			allGood = false
		}
	}

	// Check optional user prefs
	fmt.Println()
	fmt.Println("User Config:")

	prefsPath := filepath.Join(home, ".claude-quest-prefs.json")
	if _, err := os.Stat(prefsPath); err == nil {
		fmt.Println("  [OK] Preferences file exists")
	} else {
		fmt.Println("  [--] No preferences file (will use defaults)")
	}

	// Summary
	fmt.Println()
	fmt.Println("===================")
	if allGood {
		fmt.Println("All checks passed! Claude Quest should work.")
	} else {
		fmt.Println("Some issues found. See [!!] items above.")
	}
}

// getAssetPathForDoctor is a copy of getAssetPath for the doctor command
// (avoids raylib initialization issues)
func getAssetPathForDoctor(relativePath string) string {
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exeDir = filepath.Dir(resolved)
		}
		npmAssetPath := filepath.Join(exeDir, "..", "assets", relativePath)
		if _, err := os.Stat(npmAssetPath); err == nil {
			return npmAssetPath
		}
		sameDirPath := filepath.Join(exeDir, "assets", relativePath)
		if _, err := os.Stat(sameDirPath); err == nil {
			return sameDirPath
		}
	}
	return filepath.Join("assets", relativePath)
}

// validateJSONL checks if a JSONL file has the structure Claude Quest requires
func validateJSONL(path string) (bool, string) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Sprintf("cannot open: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	// Requirements we check for
	var (
		hasValidJSON     bool
		hasTypeField     bool
		hasMessageRole   bool
		hasContentArray  bool
		hasToolUseType   bool
		hasToolUseName   bool
		linesChecked     int
	)

	for scanner.Scan() {
		linesChecked++
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg struct {
			Type    string `json:"type"`
			Message struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}

		if json.Unmarshal([]byte(line), &msg) == nil {
			hasValidJSON = true

			if msg.Type != "" {
				hasTypeField = true
			}
			if msg.Message.Role != "" {
				hasMessageRole = true
			}

			// Check content structure
			if msg.Message.Content != nil {
				var content []struct {
					Type string `json:"type"`
					Name string `json:"name,omitempty"`
				}
				if json.Unmarshal(msg.Message.Content, &content) == nil && len(content) > 0 {
					hasContentArray = true
					for _, c := range content {
						if c.Type == "tool_use" {
							hasToolUseType = true
							if c.Name != "" {
								hasToolUseName = true
							}
						}
					}
				}
			}
		}

		// Stop once we've validated all requirements or checked enough lines
		if hasValidJSON && hasTypeField && hasMessageRole && hasContentArray && hasToolUseName {
			break
		}
		if linesChecked >= 100 {
			break
		}
	}

	// Check requirements
	var missing []string

	if !hasValidJSON {
		missing = append(missing, "valid JSON")
	}
	if !hasTypeField {
		missing = append(missing, "type field")
	}
	if !hasMessageRole {
		missing = append(missing, "message.role")
	}
	if !hasContentArray {
		missing = append(missing, "message.content array")
	}
	if !hasToolUseType {
		missing = append(missing, "tool_use content type")
	}
	if !hasToolUseName {
		missing = append(missing, "tool_use.name field")
	}

	if len(missing) > 0 {
		return false, "missing: " + strings.Join(missing, ", ")
	}

	return true, "all required fields present"
}

var animationNames = []string{
	"Idle", "Enter", "Casting (Read)", "Attack (Bash)",
	"Writing (Edit)", "Victory", "Hurt (Error)", "Thinking",
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

		case "studio":
			runStudio()
			os.Exit(0)

		case "doctor":
			runDoctor()
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
	renderer.SetProfile(gameState.Profile)

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

		// Check if an enemy hit Claude - trigger hurt animation
		if gameState.PendingHurt {
			gameState.PendingHurt = false
			animations.HandleEvent(Event{Type: EventEnemyHit})
		}

		// Only scroll when there's activity (events coming in)
		if gameState.IsActive {
			renderer.UpdateScroll(dt)
		}

		// Update picker animation
		renderer.UpdatePickerAnim(dt)

		// Handle keyboard input
		// Chest input takes priority when chest is active
		if gameState.ActiveChest != nil && gameState.ActiveChest.IsInteractive() {
			if rl.IsKeyPressed(rl.KeyLeft) {
				gameState.ActiveChest.SelectPrev()
			}
			if rl.IsKeyPressed(rl.KeyRight) {
				gameState.ActiveChest.SelectNext()
			}
			if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
				gameState.ActiveChest.ConfirmSelection()
			}
		} else if gameState.ActiveChest != nil {
			// Skip chest animation with any key
			if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
				gameState.ActiveChest.SkipToReveal()
			}
		} else {
			// Normal input: Up/Down = switch row, Left/Right = cycle value
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
			// Toggle picker visibility with Tab
			if rl.IsKeyPressed(rl.KeyTab) {
				renderer.TogglePicker()
			}
		}

		// Render to texture at native resolution
		rl.BeginTextureMode(target)
		rl.ClearBackground(rl.Color{R: 24, G: 20, B: 37, A: 255}) // Dark purple bg
		renderer.Draw(animations.GetState())
		renderer.DrawGameUI(gameState)
		renderer.DrawAccessoryPicker()
		renderer.DrawTreasureChest(gameState)
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
