# Claude Quest

**For Claude Code users who want to actually see what their AI is doing.**

Instead of watching text scroll by in a terminal, Claude Quest shows you a pixel-art character that reacts to every action in real-time. Reading files? Casting a spell. Running bash? Swinging into battle. Error? Taking damage. Success? Victory dance.

<p align="center">
  <img src="assets/screenshot.png" alt="Claude Quest screenshot" width="600">
</p>

```bash
npm install -g claude-quest
```

Then, in a **new terminal tab** in the same directory where you run Claude Code:

```bash
cq
```

That's it. Keep it running alongside Claude Code and watch your coding session come alive.

---

## Who This Is For

- **Long session coders** - Makes 4-hour Claude Code sessions feel like a co-op adventure
- **Streamers & content creators** - Your viewers see exactly what Claude is doing, beautifully
- **Pixel art lovers** - Paul Robertson-inspired animations that are genuinely gorgeous
- **Anyone who thinks coding should be more fun** - Because it should

## Who This Isn't For

- People who find visual feedback distracting
- Minimalists who want nothing but terminal

---

## What You Get

### Real-Time Visualization
Every Claude Code action becomes an animation:

| Claude Action | What You See |
|--------------|-----------|
| Reading files | Casting spell |
| Bash commands | Attack animation |
| Writing code | Scribbling away |
| Thinking | Contemplating |
| Extended thinking | Intense focus + particle effects |
| Success | Victory dance |
| Error | Taking damage (recovers!) |
| Git push | SHIPPED! rainbow banner |

### Five Biomes

Claude walks through beautiful parallax backgrounds that cycle every 20 seconds:

- **Enchanted Forest** - Magical trees, fireflies, glowing mushrooms
- **Mountain Journey** - Snow peaks, waterfalls, ancient ruins
- **Midnight Quest** - Starry sky, glowing crystals, spooky trees
- **Kingdom Road** - Castle, windmills, cottages, sunset
- **Wizard's Library** - Endless corridor with bookshelves, floating orbs

### The Mana Bar
Shows your remaining context window. Starts full at 200k tokens and drains as your conversation grows. When Claude compacts, it refills. Satisfying.

### Customization
Unlock cosmetics as you level up by using Claude Code:
- **Hats** - Wizard hat, crown, viking helmet, and more
- **Faces** - Sunglasses, monocle, mustaches
- **Auras** - Flame, frost, electric, rainbow particle effects
- **Trails** - Sparkles, fire, hearts that follow Claude when walking

---

## Installation

### npm
```bash
npm install -g claude-quest
```

### Direct Download
[GitHub Releases](https://github.com/Michaelliv/claude-quest/releases) - macOS, Linux, Windows

---

## Usage

**Important:** Run `cq` in a new terminal tab, in the same directory where you're running Claude Code.

```bash
cq                    # Watch current project
cq watch ~/dir        # Watch specific project
cq replay <file.jsonl> # Replay an existing conversation
cq doctor             # Check if Claude Quest can run properly
```

**Controls:** Press `Tab` to open the accessory picker. Use `←→` to switch between slots (Hat/Face/Aura/Trail), `↑↓` to cycle items. Press `Tab` or `Esc` to close.

---

## How It Works

Reads Claude Code's local conversation logs (JSONL files). No API keys. No network calls. Just file watching.

---

## The Craft

Inspired by [Paul Robertson's](http://probertson.tumblr.com/) legendary pixel art:
- 320x200 resolution (DOS-era homage)
- 24fps hand-crafted animations
- Secondary motion on everything (bob, sway, breathe)
- Atmospheric details (particles, flames, stars)

---

## FAQ

**Is this official?**
Community project. Not affiliated with Anthropic.

**Works without Claude Code?**
You can replay saved conversations with `cq replay <file.jsonl>`, but live mode requires an active Claude Code session.

**Why does this exist?**
Because staring at terminal text for hours is less fun than watching a pixel wizard battle bugs.

---

## Development

### Building from Source

```bash
git clone https://github.com/Michaelliv/claude-quest.git
cd claude-quest
go build -o cq .
```

Requires Go 1.21+ and CGO (Raylib needs C bindings).

### Studio Mode

Studio mode is a development environment for working on sprites and animations. Not included in release builds.

```bash
go build -tags debug -o cq . && ./cq studio
```

**Controls:**
- `Space` - Pause/play animation
- `< >` or arrow keys - Step frame (when paused)
- `- +` - Speed down/up (0.125x to 4x)
- `A` - Animation picker
- `B` - Biome picker
- `C` or `Tab` - Cosmetics picker (Tab cycles Hat → Face → Aura → Trail)
- `R` - Force reload all textures
- `G` - Regenerate sprites from `cmd/spritegen`
- `H` - Toggle help overlay

**Inside pickers:** `↑↓` to navigate, `Enter` to select, `Esc` to cancel

See [CLAUDE.md](CLAUDE.md) for full architecture details.

---

## License

MIT

---

<p align="center">
<i>Turn your terminal into a quest.</i>
</p>
