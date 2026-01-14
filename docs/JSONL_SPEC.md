# Claude Code JSONL Format Specification

Research completed from multiple Claude Code sessions across different projects.

## Message Types

### 1. `system` - System Events
Session initialization and control events.

```json
{
  "type": "system",
  "subtype": "local_command|compact_boundary|turn_duration",
  "content": "...",
  "sessionId": "uuid",
  "version": "2.0.76",
  "gitBranch": "develop",
  "cwd": "/path/to/project"
}
```

**Subtypes:**
- `local_command` - User ran a slash command (/hooks, /clear, etc.)
- `compact_boundary` - Conversation was compacted
  - Has `compactMetadata`: `{ trigger: "manual", preTokens: 177269 }`
- `turn_duration` - Turn timing info

### 2. `user` - User Input
User prompts and tool results.

**Prompt (string):**
```json
{
  "type": "user",
  "message": {
    "role": "user",
    "content": "help me implement this feature"
  }
}
```

**Prompt (array with text):**
```json
{
  "type": "user",
  "message": {
    "role": "user",
    "content": [{ "type": "text", "text": "..." }]
  }
}
```

**Tool Result:**
```json
{
  "type": "user",
  "message": {
    "role": "user",
    "content": [{
      "type": "tool_result",
      "tool_use_id": "toolu_...",
      "content": "output here",
      "is_error": false
    }]
  },
  "toolUseResult": {
    "stdout": "...",
    "stderr": "...",
    "interrupted": false
  }
}
```

### 3. `assistant` - Claude Response
Claude's thinking and actions.

```json
{
  "type": "assistant",
  "message": {
    "model": "claude-opus-4-5-20251101",
    "role": "assistant",
    "content": [
      { "type": "thinking", "thinking": "..." },
      { "type": "text", "text": "..." },
      { "type": "tool_use", "id": "toolu_...", "name": "Read", "input": {...} }
    ],
    "usage": {
      "input_tokens": 10,
      "cache_read_input_tokens": 19254,
      "cache_creation_input_tokens": 283,
      "output_tokens": 198
    },
    "stop_reason": "end_turn|tool_use|null"
  }
}
```

### 4. `summary` - Conversation Summary
After compact, contains conversation summary.

```json
{
  "type": "summary",
  "summary": "AI-native secrets manager built with Bun",
  "leafUuid": "uuid"
}
```

### 5. `queue-operation` - Message Queue
Tracks user message queue operations.

```json
{
  "type": "queue-operation",
  "operation": "enqueue|dequeue|remove",
  "content": "message text",
  "sessionId": "uuid",
  "timestamp": "ISO8601"
}
```

### 6. `file-history-snapshot` - File Tracking
Periodic snapshots of file changes.

```json
{
  "type": "file-history-snapshot",
  "fileHistorySnapshot": [...]
}
```

### 7. `result` - Turn Result
End of an agentic turn (defined in code, not always present).

```json
{
  "type": "result",
  "subtype": "success|error_max_turns|error_during_execution"
}
```

---

## Tools Reference

### Reading/Searching
| Tool | Description | Input Fields |
|------|-------------|--------------|
| `Read` | Read file contents | `file_path`, `offset?`, `limit?` |
| `Glob` | Find files by pattern | `pattern`, `path?` |
| `Grep` | Search file contents | `pattern`, `path?`, `glob?`, `output_mode?` |

### Writing
| Tool | Description | Input Fields |
|------|-------------|--------------|
| `Edit` | Edit file with replacement | `file_path`, `old_string`, `new_string`, `replace_all?` |
| `Write` | Write new file | `file_path`, `content` |
| `NotebookEdit` | Edit Jupyter notebook | `notebook_path`, `cell_id`, `new_source`, `edit_mode?` |

### Execution
| Tool | Description | Input Fields |
|------|-------------|--------------|
| `Bash` | Run shell command | `command`, `description?`, `timeout?`, `run_in_background?` |
| `Task` | Spawn subagent | `prompt`, `description`, `subagent_type`, `model?` |
| `TaskOutput` | Get task output | `task_id`, `block?`, `timeout?` |
| `KillShell` | Kill background shell | `shell_id` |

### Web
| Tool | Description | Input Fields |
|------|-------------|--------------|
| `WebSearch` | Search the web | `query`, `allowed_domains?`, `blocked_domains?` |
| `WebFetch` | Fetch URL content | `url`, `prompt` |

### Planning & Interaction
| Tool | Description | Input Fields |
|------|-------------|--------------|
| `TodoWrite` | Update task list | `todos[]` with `content`, `status`, `activeForm` |
| `AskUserQuestion` | Ask user questions | `questions[]` with options |
| `ExitPlanMode` | Exit planning mode | `plan`, `allowedPrompts?` |

### Skills
| Tool | Description | Input Fields |
|------|-------------|--------------|
| `Skill` | Invoke a skill | `skill`, `args?` |

---

## Agent Sessions

Agent (subagent) sessions have special fields:

```json
{
  "agentId": "aca1fa4",
  "isSidechain": true,
  "sessionId": "parent-session-uuid"
}
```

Agent files are named `agent-{agentId}.jsonl`.

---

## Models

| Model | Context | Used For |
|-------|---------|----------|
| `claude-opus-4-5-20251101` | Main | Primary conversation |
| `claude-sonnet-4-5-20250929` | Main | Alternative model |
| `claude-haiku-4-5-20251001` | Agent | Subagent tasks |
| `<synthetic>` | - | Synthetic messages |

---

## Key Fields for Game Mechanics

### Context/Mana Bar
```javascript
// From assistant message usage:
const totalTokens = usage.input_tokens +
                    usage.cache_read_input_tokens +
                    usage.cache_creation_input_tokens;
// Or from compact_boundary:
const preCompactTokens = compactMetadata.preTokens; // e.g., 177269
```

### Compact/Rest Detection
```javascript
// subtype === "compact_boundary" indicates conversation was compacted
// trigger: "manual" (user initiated) or automatic (context limit)
```

### Tool Categories for Animation
```javascript
const READING_TOOLS = ['Read', 'Glob', 'Grep', 'WebFetch', 'WebSearch'];
const WRITING_TOOLS = ['Edit', 'Write', 'NotebookEdit'];
const BASH_TOOLS = ['Bash', 'KillShell'];
const THINKING_TOOLS = ['Task', 'TaskOutput', 'AskUserQuestion', 'ExitPlanMode', 'TodoWrite', 'Skill'];
```

### User Input Detection
```javascript
// User prompt (quest text):
if (message.content is string || message.content[0].type === "text")

// Tool result (not a quest):
if (message.content[].type === "tool_result")
```

### Think Hard Detection
Look for patterns in user messages:
- "think hard", "think harder", "ultrathink"
- Extended thinking blocks are longer (check `thinking` field length)

### Error Detection
```javascript
// Tool error:
if (tool_result.is_error === true)

// Or content contains error markers:
if (content.includes("Exit code 1") || content.includes("<tool_use_error>"))
```

---

## Event Mapping for Game

| JSONL Event | Game Event | Animation |
|-------------|------------|-----------|
| `type: "system"` | EventSystemInit | AnimEnter |
| Tool: Read/Glob/Grep/Web* | EventReading | AnimCasting |
| Tool: Bash | EventBash | AnimAttack |
| Tool: Edit/Write/NotebookEdit | EventWriting | AnimWriting |
| Tool: Task | EventThinking | AnimThinking (spawn agent) |
| Tool: TodoWrite | EventThinking | AnimThinking (todo update) |
| `type: "assistant"` + `thinking` | EventThinking | AnimThinking |
| `result.subtype: "success"` | EventSuccess | AnimVictory |
| `is_error: true` | EventError | AnimHurt |
| `compact_boundary` | EventIdle | AnimIdle (sleep/rest) |
| User prompt (string) | - | Quest text display |

---

## Sample Tool Inputs

### Task (Agent Spawn)
```json
{
  "description": "Research Claude Agent SDK",
  "prompt": "Search for information about...",
  "subagent_type": "claude-code-guide"
}
```

### TodoWrite
```json
{
  "todos": [
    {"content": "Create import command", "status": "in_progress", "activeForm": "Creating import command"},
    {"content": "Create export command", "status": "pending", "activeForm": "Creating export command"}
  ]
}
```

### AskUserQuestion
```json
{
  "questions": [{
    "question": "Should rube shell out to the `claude` CLI or use the Claude API?",
    "header": "Claude integration",
    "options": [
      {"label": "Shell to claude CLI", "description": "Simpler..."},
      {"label": "Claude API via SDK", "description": "More control..."}
    ],
    "multiSelect": false
  }]
}
```

### ExitPlanMode
```json
{
  "plan": "# Implementation Plan\n\n## Phase 1...",
  "allowedPrompts": [
    {"tool": "Bash", "prompt": "run tests"},
    {"tool": "Bash", "prompt": "install dependencies"}
  ]
}
```
