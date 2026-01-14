# Claude Quest - Game Mechanics Specification

Based on JSONL research from multiple Claude Code sessions.

## Core Systems

### 1. Animation System (Existing ✓)
Already implemented with 8 animations:
- **AnimIdle** - Default breathing
- **AnimEnter** - Session start pop-in
- **AnimCasting** - Reading/searching (sparkles)
- **AnimAttack** - Bash commands (punch)
- **AnimWriting** - Edit/Write (typing)
- **AnimVictory** - Success (jump)
- **AnimHurt** - Error (knockback)
- **AnimThinking** - Processing (sway + thought bubble)

### 2. Quest System (NEW)
Display user prompts as "quests" at the top of the screen.

**Visual Design:**
```
┌────────────────────────────────────────┐
│ ⚔️ QUEST: help me implement this feature │
└────────────────────────────────────────┘
```

**Implementation:**
- Parse `type: "user"` messages where content is string or `content[0].type === "text"`
- Exclude tool results (has `type: "tool_result"`)
- Animate text scrolling in from right
- Truncate long quests with "..."
- Queue multiple quests if they come fast

**Data Source:**
```go
// In watcher.go parseLine():
case "user":
    for _, content := range msg.Message.Content {
        if content.Type == "text" && len(content.Text) > 0 {
            return &Event{Type: EventQuest, Details: content.Text}
        }
    }
    // Or if content is string directly
```

### 3. Mana Bar (Context Window) (NEW)
Visualize context window usage as a mana bar.

**Visual Design:**
```
MANA ████████████░░░░░░░ 150K/200K
```

**Data Source:**
```go
// From assistant messages:
type UsageInfo struct {
    InputTokens          int
    CacheReadTokens      int
    CacheCreationTokens  int
    OutputTokens         int
}

// Total context ≈ input + cache_read + cache_creation
// Max context (Opus 4.5) = ~200K tokens
```

**Behavior:**
- Blue fill for normal usage
- Yellow when > 75% full
- Red when > 90% full (approaching limit)
- Flash/pulse when nearing limit
- Refill animation on compact

### 4. Rest/Sleep Mode (Compact) (NEW)
When conversation is compacted, show Claude sleeping/resting.

**Detection:**
```go
// subtype == "compact_boundary"
type CompactMetadata struct {
    Trigger   string // "manual" or automatic
    PreTokens int    // Tokens before compact
}
```

**Animation:**
- New `AnimSleep` animation (Zzz bubbles, eyes closed)
- Or overlay "Zzz" particles on idle
- Mana bar refills smoothly
- Optional: Show dream thought bubble with summary text

### 5. Think Hard Effects (NEW)
Comic-style text bursts for intense thinking.

**Detection:**
Look for these patterns in user messages (case-insensitive):
- "think hard"
- "think harder"
- "ultrathink"
- "deep think"

Also detect long thinking blocks:
```go
// If thinking content > 1000 chars, likely extended thinking
```

**Visual Effects:**
- Comic burst text: "POW!", "BOOM!", "THINK!", "💭💥"
- Screen shake (subtle)
- Sparks/lightning around Claude
- Thought bubble grows larger

### 6. Agent Spawn Effects (NEW)
Visual feedback when Task tool spawns agents.

**Detection:**
```go
case toolName == "task":
    return &Event{Type: EventSpawnAgent, Details: input.Description}
```

**Visual:**
- Mini Claude appears next to main Claude
- Or: Portal/teleport effect
- Agent type shown as text badge ("Explore", "Plan", etc.)

### 7. Todo List Display (NEW)
Show TodoWrite updates as a quest log.

**Detection:**
```go
case toolName == "todowrite":
    // Parse input.todos array
```

**Visual Design:**
```
┌─ TASKS ──────────────┐
│ ✓ Create import cmd  │
│ ► Create export cmd  │
│ ○ Run tests          │
└──────────────────────┘
```
- ✓ = completed
- ► = in_progress
- ○ = pending

**Position:** Bottom-right or side panel

### 8. Tool Error Effects (NEW)
Enhanced error feedback.

**Detection:**
```go
// is_error: true
// Or content contains:
// - "Exit code 1"
// - "<tool_use_error>"
// - "Error:"
```

**Visual:**
- Red flash overlay
- Error text floats up
- More dramatic AnimHurt particles
- Optional: Shake screen

---

## Event Types (Expanded)

```go
type EventType int

const (
    EventSystemInit  EventType = iota // Session start
    EventThinking                     // Thinking/processing
    EventReading                      // Read/Glob/Grep/Web
    EventBash                         // Bash command
    EventWriting                      // Edit/Write
    EventSuccess                      // Task complete
    EventError                        // Tool error
    EventIdle                         // No activity

    // NEW EVENTS
    EventQuest                        // User prompt
    EventCompact                      // Conversation compacted
    EventSpawnAgent                   // Task tool used
    EventTodoUpdate                   // TodoWrite used
    EventThinkHard                    // Extended thinking requested
    EventAskUser                      // AskUserQuestion used
    EventPlanMode                     // ExitPlanMode used
)
```

---

## UI Layout

```
┌────────────────────────────────────────────────────────────────┐
│  ⚔️ QUEST: implement user authentication                        │  <- Quest text
├────────────────────────────────────────────────────────────────┤
│                                                                │
│                                                                │
│                        [CLAUDE SPRITE]                         │
│                                                                │
│                                                                │
│     ┌─ TASKS ─────────┐                                        │
│     │ ► Research auth │                                        │
│     │ ○ Implement JWT │                                        │
│     └─────────────────┘                                        │
├────────────────────────────────────────────────────────────────┤
│  HAT  < wizard >     MANA ████████████░░░░░░░ 150K/200K       │
│  FACE < - >                                                    │
└────────────────────────────────────────────────────────────────┘
```

---

## Implementation Priority

### Phase 1: Core Feedback
1. Quest text display (user prompts)
2. Mana bar (context window)
3. Compact/sleep detection

### Phase 2: Enhanced Effects
4. Think hard effects
5. Error enhancement
6. Agent spawn visuals

### Phase 3: Information Display
7. Todo list display
8. Tool details (what file being read, etc.)

---

## Data Structures

### Expanded Event
```go
type Event struct {
    Type     EventType
    Details  string

    // NEW FIELDS
    TokenUsage   *TokenUsage   // For mana bar
    TodoList     []TodoItem    // For todo display
    AgentInfo    *AgentInfo    // For spawn effects
    ErrorInfo    *ErrorInfo    // For error details
}

type TokenUsage struct {
    Total    int
    Max      int  // ~200K for Opus
}

type TodoItem struct {
    Content    string
    Status     string // "completed", "in_progress", "pending"
    ActiveForm string
}

type AgentInfo struct {
    Type        string // "Explore", "Plan", etc.
    Description string
}

type ErrorInfo struct {
    Tool    string
    Message string
    Code    int
}
```

### Game State
```go
type GameState struct {
    CurrentQuest    string
    QuestTimer      float32  // For scroll animation

    ManaTotal       int
    ManaMax         int
    ManaPrevious    int      // For smooth animation

    TodoItems       []TodoItem

    IsCompacting    bool     // Sleep mode active
    CompactProgress float32  // Animation progress

    ThinkHardActive bool     // Comic effects
    ThinkHardTimer  float32

    ActiveAgents    []string // Spawned agent types
}
```

---

## Sample Watcher Updates

```go
func (w *Watcher) parseLine(line string) *Event {
    var msg ClaudeMessage
    if err := json.Unmarshal([]byte(line), &msg); err != nil {
        return nil
    }

    switch msg.Type {
    case "system":
        if msg.Subtype == "compact_boundary" {
            return &Event{
                Type:    EventCompact,
                Details: fmt.Sprintf("Compacted from %d tokens", msg.CompactMetadata.PreTokens),
            }
        }
        return &Event{Type: EventSystemInit}

    case "user":
        // Check for quest (user prompt)
        if text := extractUserPrompt(msg); text != "" {
            // Check for think hard
            if isThinkHard(text) {
                return &Event{Type: EventThinkHard, Details: text}
            }
            return &Event{Type: EventQuest, Details: text}
        }

    case "assistant":
        // Extract token usage for mana bar
        usage := extractTokenUsage(msg)

        for _, content := range msg.Message.Content {
            if content.Type == "tool_use" {
                evt := w.parseToolUse(content.Name, content.Input)
                evt.TokenUsage = usage
                return evt
            }
            if content.Type == "thinking" {
                return &Event{Type: EventThinking, TokenUsage: usage}
            }
        }

    case "summary":
        return &Event{Type: EventIdle, Details: msg.Summary}
    }

    return nil
}

func isThinkHard(text string) bool {
    lower := strings.ToLower(text)
    return strings.Contains(lower, "think hard") ||
           strings.Contains(lower, "ultrathink") ||
           strings.Contains(lower, "think deeper")
}
```
