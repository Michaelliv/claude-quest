# Claude Quest

**Your AI companion, brought to life.**

> *A pixel-art RPG companion that visualizes your Claude Code sessions in real-time. Watch Claude cast spells while reading files, attack bugs with bash commands, and celebrate victories when your code works.*

---

## What is this?

Claude Quest transforms your Claude Code terminal sessions into an animated adventure. Instead of just watching text scroll by, open Claude Quest in a window and see your AI assistant come to life as a cute, expressive pixel-art character.

**Claude reads a file?** Watch them cast a spell.
**Running a bash command?** They swing into action.
**Hit an error?** They take damage (but recover!).
**Task complete?** Victory dance!

It's the coding companion you didn't know you needed.

---

## Features

### Live Session Watching
Connect to your active Claude Code conversation and watch Claude react in real-time to every action.

### Two Modes

**Vibin' Mode** - A cozy wizard's study with flickering candles, twinkling stars through the window, bubbling potions, and floating dust motes. Perfect for focused coding sessions.

**Quest Mode** - Claude walks through four beautiful parallax biomes that cycle as you work:
- **Enchanted Forest** - Glowing mushrooms, fireflies, ancient magic trees
- **Mountain Journey** - Snow-capped peaks, wandering clouds, pine forests
- **Midnight Quest** - Starlit ruins, crystal caves, ethereal waterfalls
- **Kingdom Road** - Castles on hills, spinning windmills, cozy cottages

### Customization
Dress up Claude with hats and accessories. Express yourself!

### Demo Mode
Not using Claude Code? Run `cq demo` to see all animations and explore the biomes.

---

## Installation

### npm (easiest)
```bash
npm install -g claude-quest
```

### Homebrew (coming soon)
```bash
brew install claude-quest
```

### Direct Download
Grab the latest release for your platform from [GitHub Releases](https://github.com/Michaelliv/claude-quest/releases):
- macOS (Intel & Apple Silicon)
- Linux (x64)
- Windows (x64)

### From Source
```bash
git clone https://github.com/Michaelliv/claude-quest.git
cd claude-quest
go build -o cq .
```

---

## Usage

### Watch your current project
```bash
# In your project directory (where you're running Claude Code)
cq
```

### Watch a specific project
```bash
cq watch ~/Projects/my-app
```

### Replay a past conversation
```bash
cq replay ~/.claude/projects/.../conversation.jsonl
```

### Demo mode
```bash
cq demo
```

### Controls
| Key | Action |
|-----|--------|
| `W` | Toggle Walk/Vibe mode |
| `↑` `↓` | Switch picker row |
| `←` `→` | Cycle accessories |
| `Q` | Show quest text (demo) |
| `M` | Add mana (demo) |
| `1-4` | Think levels (demo) |

---

## How It Works

Claude Quest monitors Claude Code's conversation logs (JSONL files) and translates events into animations:

| Claude Action | Animation |
|--------------|-----------|
| Reading files | Casting spell |
| Bash commands | Attack |
| Writing/editing | Writing |
| Thinking | Contemplating |
| Extended thinking | Intense focus + particles |
| Success | Victory dance |
| Error | Taking damage |
| New task | Quest received |

The mana bar shows context window usage - watch it fill as your conversation grows!

---

## The Aesthetic

Inspired by the legendary pixel art of [Paul Robertson](http://probertson.tumblr.com/), Claude Quest embraces:

- **Expressive animation** - Every frame matters
- **Rich secondary motion** - Things bob, sway, and breathe
- **Atmospheric details** - Floating particles, flickering flames, twinkling stars
- **Personality** - Claude isn't just a sprite, they're a character

The 320x200 resolution pays homage to classic DOS-era RPGs while the 24fps animations keep everything buttery smooth.

---

## Requirements

- **Claude Code** - This is a companion app for [Claude Code](https://claude.ai/code)
- Works on macOS, Linux, and Windows
- No dependencies when installed via npm or binary releases

---

## Development

Built with:
- **Go** - Core application
- **raylib-go** - Graphics and windowing
- **Custom sprite system** - All animations generated procedurally

### Building from source
```bash
# Install dependencies (Linux only)
sudo apt-get install libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev libwayland-dev libxkbcommon-dev

# Build
go build -o cq .

# Regenerate sprites
go run ./cmd/spritegen/...
```

---

## FAQ

**Q: Does this actually connect to Claude?**
A: It reads Claude Code's local conversation logs. No API keys or network connections needed.

**Q: Can I use this without Claude Code?**
A: Yes! Run `cq demo` to enjoy the animations standalone.

**Q: Why pixel art?**
A: Because pixel art is beautiful, and coding should be fun.

**Q: Is this official?**
A: This is a community project, not affiliated with Anthropic.

---

## Contributing

Found a bug? Want to add a new hat? PRs welcome!

---

## License

MIT

---

<p align="center">
  <i>Made with pixels and passion</i>
</p>
