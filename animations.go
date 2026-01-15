package main

// AnimationType represents the current animation being played
type AnimationType int

const (
	AnimIdle AnimationType = iota
	AnimEnter
	AnimCasting   // Reading/searching
	AnimAttack    // Bash commands
	AnimWriting   // Edit/Write
	AnimVictory   // Success
	AnimHurt      // Error
	AnimThinking  // Processing
	AnimWalk      // Walking
)

func (a AnimationType) String() string {
	names := []string{
		"Idle",
		"Enter",
		"Casting",
		"Attack",
		"Writing",
		"Victory",
		"Hurt",
		"Thinking",
		"Walk",
	}
	if int(a) < len(names) {
		return names[a]
	}
	return "Unknown"
}

// AnimationState holds the current state of Claude's animation
type AnimationState struct {
	CurrentAnim AnimationType
	Frame       int
	Timer       float32
	Queue       []AnimationType // Queued animations to play
}

// AnimationSystem manages Claude's animation state machine
type AnimationSystem struct {
	state         *AnimationState
	frameDuration float32 // Seconds per frame
	animLengths   map[AnimationType]int
	walkMode      bool // When true and active, default to walk instead of idle
	isActive      bool // When true, there's recent activity (events coming in)
}

// NewAnimationSystem creates a new animation system
func NewAnimationSystem() *AnimationSystem {
	return &AnimationSystem{
		state: &AnimationState{
			CurrentAnim: AnimIdle,
			Frame:       0,
			Timer:       0,
			Queue:       make([]AnimationType, 0),
		},
		frameDuration: 0.042, // 24 FPS for smooth animation
		animLengths: map[AnimationType]int{
			AnimIdle:     16,
			AnimEnter:    20,
			AnimCasting:  16,
			AnimAttack:   16,
			AnimWriting:  16,
			AnimVictory:  20,
			AnimHurt:     16,
			AnimThinking: 12,
			AnimWalk:     16,
		},
	}
}

// HandleEvent processes a Claude Code event and triggers appropriate animation
func (a *AnimationSystem) HandleEvent(event Event) {
	var newAnim AnimationType

	switch event.Type {
	case EventSystemInit:
		newAnim = AnimEnter
	case EventReading:
		newAnim = AnimCasting
	case EventBash:
		newAnim = AnimAttack
	case EventWriting:
		newAnim = AnimWriting
	case EventSuccess:
		newAnim = AnimVictory
	case EventError:
		newAnim = AnimHurt
	case EventThinking:
		newAnim = AnimThinking
	case EventIdle:
		newAnim = AnimIdle

	// New event types
	case EventQuest:
		// Quest received - no animation change, just display quest text
		return
	case EventCompact:
		// Conversation compacted - go to idle (rest/sleep)
		newAnim = AnimIdle
	case EventThinkHard:
		// Extended thinking - use thinking animation (effects handled by renderer)
		newAnim = AnimThinking
	case EventSpawnAgent:
		// Agent spawned - use casting animation (summoning)
		newAnim = AnimCasting
	case EventTodoUpdate:
		// Todo update - use writing animation briefly
		newAnim = AnimWriting
	case EventAskUser:
		// Asking user - use thinking animation
		newAnim = AnimThinking

	default:
		return
	}

	// Queue the animation
	a.queueAnimation(newAnim)
}

// queueAnimation adds an animation to the queue or plays immediately
func (a *AnimationSystem) queueAnimation(anim AnimationType) {
	// If idle, play immediately
	if a.state.CurrentAnim == AnimIdle {
		a.state.CurrentAnim = anim
		a.state.Frame = 0
		a.state.Timer = 0
	} else {
		// Otherwise queue it
		a.state.Queue = append(a.state.Queue, anim)
	}
}

// Update advances the animation state
func (a *AnimationSystem) Update(deltaTime float32) {
	a.state.Timer += deltaTime

	// Advance frame based on timer
	if a.state.Timer >= a.frameDuration {
		a.state.Timer -= a.frameDuration
		a.state.Frame++

		// Check if animation completed
		animLen := a.animLengths[a.state.CurrentAnim]
		if a.state.Frame >= animLen {
			a.onAnimationComplete()
		}
	}
}

// onAnimationComplete handles animation end
func (a *AnimationSystem) onAnimationComplete() {
	// Check queue for next animation
	if len(a.state.Queue) > 0 {
		a.state.CurrentAnim = a.state.Queue[0]
		a.state.Queue = a.state.Queue[1:]
		a.state.Frame = 0
	} else {
		// Return to walk if in active walk mode, otherwise idle
		if a.walkMode && a.isActive {
			a.state.CurrentAnim = AnimWalk
		} else {
			a.state.CurrentAnim = AnimIdle
		}
		a.state.Frame = 0
	}
}

// SetWalkMode enables/disables walk mode (Quest mode)
func (a *AnimationSystem) SetWalkMode(enabled bool) {
	a.walkMode = enabled
}

// SetActive sets whether there's current activity (events coming in)
func (a *AnimationSystem) SetActive(active bool) {
	wasActive := a.isActive
	a.isActive = active

	// If becoming inactive while walking, switch to idle
	if wasActive && !active && a.state.CurrentAnim == AnimWalk {
		a.state.CurrentAnim = AnimIdle
		a.state.Frame = 0
	}
	// If becoming active in walk mode while idle, switch to walk
	if !wasActive && active && a.walkMode && a.state.CurrentAnim == AnimIdle {
		a.state.CurrentAnim = AnimWalk
		a.state.Frame = 0
	}
}

// GetState returns the current animation state for rendering
func (a *AnimationSystem) GetState() *AnimationState {
	return a.state
}
