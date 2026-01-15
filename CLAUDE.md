# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
go build -o cq .              # Build binary for current platform
./cq demo                     # Interactive demo (no Claude Code needed)
./cq                          # Watch current project's Claude conversations
./cq watch ~/path/to/project  # Watch specific project
./cq replay <jsonl-file>      # Replay a conversation file
./cq doctor                   # Diagnostics - check paths and JSONL structure
```

Cross-platform builds use goreleaser (triggered by git tags `v*`):
```bash
goreleaser release --snapshot  # Build all platforms locally
```

## Architecture

Claude Quest is a pixel-art RPG companion that visualizes Claude Code operations in real-time. It watches JSONL conversation logs and animates a character reacting to different tool calls.

### Core Components

**watcher.go** - File watcher that monitors `~/.claude/projects/[encoded-path]/*.jsonl`. Parses JSON lines, extracts tool usage events, and emits typed `Event` structs through a channel. Supports live tailing and replay modes.

**main.go** - Entry point with CLI parsing (`demo`, `watch`, `replay`, `doctor` subcommands). Contains the main game loop, `GameState` (mana, todos, thrown tools, mini agents), and `MiniAgent` tracking for subagents.

**animations.go** - State machine managing 9 animation types: Idle, Enter, Casting, Attack, Writing, Victory, Hurt, Thinking, Walk. Each animation has frame timing and transition rules.

**renderer.go** - All drawing logic using Raylib. Handles sprite sheets (main 32x32 and mini 16x16), parallax backgrounds, UI elements (quest text, mana bar, collapsible picker), particle effects, and biome cycling. Renders at 320x200 native resolution (DOS-era aesthetic), scaled up to window size.

**config.go** - User preferences (accessories, volume, background) stored at `~/.claude-quest-prefs.json`.

**pixelart.go** - Color palette definitions and pixel manipulation utilities.

**sprites/** - Generated Go files containing procedural sprite data for faces and outfits.

**cmd/spritegen/** - Sprite sheet generator. Creates `spritesheet.png` (main Claude, 32x32) and `mini_spritesheet.png` (mini Claude for subagents, 16x16). Run with `go run ./cmd/spritegen/`.

### Event Flow

1. Watcher parses JSONL and emits `Event{Type, Details, TokenUsage, TodoItems, ThinkLevel}`
2. Main loop receives events and updates `GameState` and `AnimationSystem`
3. Renderer draws current state at 60 FPS (animations run at 24 FPS)

### Event Type Mappings

| Claude Tool | Animation |
|-------------|-----------|
| glob, read, grep, websearch, webfetch | Casting |
| bash, killshell | Attack |
| edit, write, notebookedit | Writing |
| success results | Victory |
| errors | Hurt |
| extended thinking | Thinking + particles |
| Task (subagent) | Mini Claude spawns and jumps out |

### Build Requirements

- **CGO_ENABLED=1** (Raylib needs C bindings)
- **Linux deps**: libgl1-mesa-dev, libxi-dev, libxcursor-dev, libxrandr-dev, libxinerama-dev, libxxf86vm-dev, libwayland-dev, libxkbcommon-dev

### Distribution

npm package (`claude-quest`) runs a postinstall script (`scripts/install.js`) that downloads pre-built binaries from GitHub releases. The `cq` and `claude-quest` commands are aliases to the same binary.

## Key Constants

- Screen: 320x200 pixels, scaled 2x for sprites
- Main Claude: 32x32 pixel frames, max 12 frames per animation
- Mini Claude: 16x16 pixel frames (for subagents)
- Mana bar: 200,000 tokens max (drains as context is used)
- Biome cycle: 20 seconds per biome in Quest mode

## Demo Controls

- `W` - Toggle walk mode (Vibin' / Quest!)
- `Tab` - Toggle accessory picker
- `↑↓` - Switch picker row (HAT/FACE/MODE)
- `←→` - Cycle values
- `A` - Spawn mini agent (demo only)
- `P` - Poof mini agent (demo only)
