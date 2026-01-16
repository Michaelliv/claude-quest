# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
go build -o cq .              # Build binary for current platform
./cq                          # Watch current project's Claude conversations
./cq watch ~/path/to/project  # Watch specific project
./cq replay <jsonl-file>      # Replay a conversation file
./cq doctor                   # Diagnostics - check paths and JSONL structure

# Studio mode - asset development with hot reload (not included in release)
go build -tags debug -o cq . && ./cq studio
```

Cross-platform builds use goreleaser (triggered by git tags `v*`):
```bash
goreleaser release --snapshot  # Build all platforms locally
```

## Build Requirements

- **CGO_ENABLED=1** (Raylib needs C bindings)
- **Linux deps**: libgl1-mesa-dev, libxi-dev, libxcursor-dev, libxrandr-dev, libxinerama-dev, libxxf86vm-dev, libwayland-dev, libxkbcommon-dev

## Architecture

Claude Quest is a pixel-art RPG companion that visualizes Claude Code operations in real-time. It watches JSONL conversation logs and animates a character reacting to different tool calls.

### Core Components

**main.go** - Entry point with CLI parsing (`watch`, `replay`, `doctor`, `studio` subcommands). Contains the main game loop, `GameState` (mana, XP, todos, thrown tools, mini agents, flying enemies), and event handling logic.

**studio.go** - Studio mode for asset development (only included with `-tags debug` build). Clean UI with animation stepper, biome selector, accessory picker, and on-screen control hints.

**hot_reload.go** - Hot reload system (only included with `-tags debug` build). Uses fsnotify to watch `assets/` and `cmd/spritegen/` for changes. Auto-reloads textures and regenerates sprites on file changes.

**watcher.go** - File watcher that monitors `~/.claude/projects/[encoded-path]/*.jsonl`. Parses JSON lines, extracts tool usage events, and emits typed `Event` structs through a channel. Supports live tailing and replay modes.

**animations.go** - State machine managing 10 animation types: Idle, Enter, Casting, Attack, Writing, Victory, Hurt, Thinking, Walk, VictoryPose. Each animation has frame timing and transition rules. Always in "Quest mode" - walks when active, idles when nothing happening.

### Renderer Architecture (Split into modules)

**renderer.go** - Core renderer struct, initialization, and main Draw loop. Orchestrates background and sprite drawing.

**renderer_claude.go** - Claude sprite drawing: `drawClaude`, `getHeadOffset` (animation-specific head positioning for accessories), `drawHat`, `drawFace`, `drawPlaceholderClaude`.

**renderer_particles.go** - Particle system: `spawnParticles`, `updateParticles`, `drawParticles`. Creates visual effects for casting, attacks, etc.

**renderer_ui.go** - Game UI elements: quest text, mana bar, XP bar, level display, flow meter, thought bubbles, floating XP indicators, SHIPPED! banner, mini agents, flying enemies, treasure chest ceremony.

**renderer_effects.go** - Visual effects: `drawShippedBanner` (rainbow SHIPPED!), `drawThinkHardEffect` (firework particles for thinking levels), `drawCompactEffect` (Zzz sleep effect), `hsvToRGB` color utility.

**renderer_picker.go** - Accessory picker UI with HAT/FACE rows. Handles cycling, preferences save/load, scroll updates.

**renderer_helpers.go** - Shared drawing utilities: `min`, `drawMountain`, `drawHill`, `drawTree`, `drawGrass`.

### Biome System (5 parallax backgrounds)

All biomes use parallax scrolling with multiple depth layers:

- **biome_forest.go** - Enchanted Forest: magical trees, fireflies, mushrooms, fog layers
- **biome_mountain.go** - Mountain Journey: snow peaks, waterfalls, ruins, sun glow
- **biome_midnight.go** - Midnight Quest: starry sky, glowing crystals, spooky trees, magic particles
- **biome_kingdom.go** - Kingdom Road: castle, windmills, cottages, farmland, sunset
- **biome_study.go** - Wizard's Library: endless corridor with bookshelves, windows, chandeliers, desks, floating orbs

Biomes cycle every 20 seconds. Each biome draws elements at different scroll speeds (0.05x for distant, 1.0x for ground) to create depth.

### Progression System

**progression.go** - Career profile with XP, levels, and unlockable accessories. Tracks owned items and level-up rewards.

**treasure_chest.go** - Treasure chest ceremony for level-up rewards. State machine: Closed → Wobble → Opening → Revealing → Choosing → Claiming → Done.

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
| git push | SHIPPED! rainbow banner |

### Sprite Generation

**cmd/spritegen/** - Sprite sheet generator. Creates:
- `assets/claude/spritesheet.png` (main Claude, 32x32 frames)
- `assets/claude/mini_spritesheet.png` (mini Claude for subagents, 16x16 frames)
- `assets/enemies/enemy_spritesheet.png` (bugs, errors, low context enemies)
- `assets/ui/chest.png` (treasure chest states)

Run with `go run ./cmd/spritegen/`.

## Key Constants

- Screen: 320x200 pixels, scaled 2x for sprites
- Main Claude: 32x32 pixel frames, max 12 frames per animation
- Mini Claude: 16x16 pixel frames (for subagents)
- Mana bar: 200,000 tokens max (drains as context is used)
- Biome cycle: 20 seconds per biome

## Studio Mode

Studio mode is a clean asset development environment. Run with:
```bash
go build -tags debug -o cq . && ./cq studio
```

### On-Screen Controls

All controls are shown at the bottom of the screen. Press `H` to toggle help.

**Playback:**
| Key | Action |
|-----|--------|
| `Space` | Pause/play |
| `<` `>` (or `,` `.`) | Step frame back/forward (when paused) |
| `-` `+` (or `-` `=`) | Speed down/up (0.125x to 4x) |

**Pickers:**
| Key | Action |
|-----|--------|
| `A` | Open animation picker |
| `B` | Open biome picker |
| `C` or `Tab` | Open cosmetics picker (Hat/Face/Aura/Trail) |

**Inside Picker:**
| Key | Action |
|-----|--------|
| `Up/Down` or `W/S` | Navigate items |
| `Enter` or `Space` | Select and close |
| `Esc` | Close without selecting |
| `Tab` | Cycle cosmetic type (when in cosmetics picker) |

**Hot Reload:**
| Key | Action |
|-----|--------|
| `R` | Force reload all textures |
| `G` | Regenerate sprites (runs spritegen) |

Note: In studio mode, animations loop continuously and biomes don't auto-cycle (unlike normal mode).

### Hot Reload

Studio mode watches files and auto-reloads:
- `assets/**/*.png` changes → textures reload instantly
- `cmd/spritegen/*.go` changes → sprites regenerate automatically

## Distribution

npm package (`claude-quest`) runs a postinstall script (`scripts/install.js`) that downloads pre-built binaries from GitHub releases. The `cq` and `claude-quest` commands are aliases to the same binary.

## Important Notes

- Parallax backgrounds always scroll; Claude walks when active, stands idle when nothing happening
- Negative x values in biome rendering can cause issues - always handle modulo of negative numbers
- Accessories follow Claude's head using `getHeadOffset()` which returns frame-specific x,y offsets matching the sprite generator curves
