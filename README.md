# Claude Quest

**For Claude Code users who want to actually see what their AI is doing.**

Instead of watching text scroll by in a terminal, Claude Quest shows you a pixel-art character that reacts to every action in real-time. Reading files? Casting a spell. Running bash? Swinging into battle. Error? Taking damage. Success? Victory dance.

```bash
npm install -g claude-quest
cq
```

That's it. Open it alongside Claude Code and watch your coding session come alive.

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

### Two Modes

**Vibin'** - A cozy wizard's study. Flickering candles, twinkling stars through the window, bubbling potions. For focused work.

**Quest!** - Claude walks through four parallax biomes that cycle every 20 seconds:
- Enchanted Forest (fireflies, glowing mushrooms)
- Mountain Journey (snow peaks, pine forests)
- Midnight Quest (starlit ruins, crystal caves)
- Kingdom Road (castles, windmills)

### The Mana Bar
Shows your context window usage. Watch it fill as your conversation grows. When Claude compacts the conversation, it resets. Satisfying.

### Customization
Hats and accessories. Wizard hat, crown, deal-with-it sunglasses. Because why not.

---

## Installation

### npm
```bash
npm install -g claude-quest
```

### Direct Download
[GitHub Releases](https://github.com/Michaelliv/claude-quest/releases) - macOS, Linux, Windows

### From Source
```bash
git clone https://github.com/Michaelliv/claude-quest.git
cd claude-quest
go build -o cq .
```

---

## Usage

```bash
cq              # Watch current project
cq demo         # See all animations (no Claude Code needed)
cq watch ~/dir  # Watch specific project
```

**Controls:** `W` toggle walk mode, `↑↓←→` customize accessories

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
Yes. `cq demo` runs standalone.

**Why does this exist?**
Because staring at terminal text for hours is less fun than watching a pixel wizard battle bugs.

---

## License

MIT

---

<p align="center">
<i>Turn your terminal into a quest.</i>
</p>
