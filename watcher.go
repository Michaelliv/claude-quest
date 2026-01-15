package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// EventType represents the type of Claude Code event
type EventType int

const (
	EventSystemInit EventType = iota
	EventThinking
	EventReading    // Glob, Read, Grep, WebFetch, WebSearch
	EventBash       // Bash tool
	EventWriting    // Edit, Write, NotebookEdit
	EventSuccess    // Successful result
	EventError      // Error result
	EventIdle       // No activity

	// New event types
	EventQuest      // User prompt (quest text)
	EventCompact    // Conversation compacted (sleep/rest)
	EventThinkHard  // Extended thinking requested
	EventSpawnAgent    // Task tool spawned an agent
	EventAgentComplete // Task tool completed (agent should poof)
	EventTodoUpdate    // TodoWrite tool used
	EventAskUser       // AskUserQuestion tool
	EventEnemyHit      // Enemy hit Claude (triggers hurt animation)
	EventVictoryPose   // Triumphant fist pump (for epic moments like git push)
	EventGitPush       // Git push detected - SHIPPED! rainbow effect
)

// TokenUsage tracks context window usage for mana bar
type TokenUsage struct {
	InputTokens         int `json:"input_tokens"`
	CacheReadTokens     int `json:"cache_read_input_tokens"`
	CacheCreationTokens int `json:"cache_creation_input_tokens"`
	OutputTokens        int `json:"output_tokens"`
}

// Total returns total tokens used
func (t *TokenUsage) Total() int {
	return t.InputTokens + t.CacheReadTokens + t.CacheCreationTokens
}

// TodoItem represents a single todo item
type TodoItem struct {
	Content    string `json:"content"`
	Status     string `json:"status"`
	ActiveForm string `json:"activeForm"`
}

// CompactInfo contains compaction metadata
type CompactInfo struct {
	Trigger   string `json:"trigger"`
	PreTokens int    `json:"preTokens"`
}

// Event represents a parsed Claude Code event
type Event struct {
	Type    EventType
	Details string

	// Extended data for game mechanics
	TokenUsage   *TokenUsage  // For mana bar
	TodoItems    []TodoItem   // For todo display
	CompactInfo  *CompactInfo // For compact/sleep
	ToolName     string       // Original tool name
	ToolUseID    string       // Tool use ID (for tracking Task completions)
	IsError      bool         // Whether this was an error
	ThinkLevel   ThinkLevel   // For think hard effects
	ThoughtText  string       // Claude's thinking content (for thought bubble)
}

// ClaudeMessage represents the structure of Claude Code JSONL format
type ClaudeMessage struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype,omitempty"`

	// For system messages
	CompactMetadata *CompactInfo `json:"compactMetadata,omitempty"`

	// For summary messages
	Summary string `json:"summary,omitempty"`

	// Message content
	Message struct {
		Role    string `json:"role"`
		Model   string `json:"model,omitempty"`
		Content json.RawMessage `json:"content"` // Can be string or array
		Usage   *TokenUsage     `json:"usage,omitempty"`
	} `json:"message,omitempty"`
}

// ContentItem represents a single content item in message.content array
type ContentItem struct {
	Type      string          `json:"type"`
	ID        string          `json:"id,omitempty"`         // For tool_use - the tool use ID
	Name      string          `json:"name,omitempty"`
	Text      string          `json:"text,omitempty"`
	Thinking  string          `json:"thinking,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"` // For tool_result (can be string or array)
	ToolUseID string          `json:"tool_use_id,omitempty"` // For tool_result - reference to tool_use ID
}

// TodoWriteInput represents the input for TodoWrite tool
type TodoWriteInput struct {
	Todos []TodoItem `json:"todos"`
}

// TaskInput represents the input for Task tool
type TaskInput struct {
	Description  string `json:"description"`
	SubagentType string `json:"subagent_type"`
	Prompt       string `json:"prompt"`
}

// WatchMode determines how the watcher operates
type WatchMode int

const (
	ModeLive   WatchMode = iota // Watch active conversation
	ModeReplay                  // Replay existing conversation
)

// Watcher monitors Claude Code conversations and emits events
type Watcher struct {
	Events      chan Event
	Mode        WatchMode
	FilePath    string        // Path to JSONL file
	ProjectDir  string        // Claude project directory (for checking new files)
	ReplaySpeed time.Duration // Delay between events in replay mode
	lastPos     int64         // Last read position for tailing
	lastModTime time.Time     // Last modification time of current file

	// State tracking
	LastTokenUsage   *TokenUsage
	CurrentTodos     []TodoItem
	ActiveTaskAgents map[string]string // tool_use_id -> agent type (for poof detection)
}

// NewWatcher creates a new event watcher
func NewWatcher() *Watcher {
	return &Watcher{
		Events:           make(chan Event, 100),
		ReplaySpeed:      200 * time.Millisecond, // Default replay speed
		ActiveTaskAgents: make(map[string]string),
	}
}

// FindProjectConversation finds the latest conversation file for a project directory
func (w *Watcher) FindProjectConversation(projectDir string) error {
	// Convert project path to Claude's encoded format
	// /Users/foo/project -> -Users-foo-project
	absPath, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	encoded := strings.ReplaceAll(absPath, "/", "-")
	claudeProjectDir := filepath.Join(os.Getenv("HOME"), ".claude", "projects", encoded)
	w.ProjectDir = claudeProjectDir

	// Check if project directory exists
	if _, err := os.Stat(claudeProjectDir); os.IsNotExist(err) {
		return fmt.Errorf("no Claude conversations found for %s\nlooked in: %s", projectDir, claudeProjectDir)
	}

	// Find the most recently modified .jsonl file (excluding agent- files)
	filePath, modTime, err := w.findNewestConversation()
	if err != nil {
		return err
	}

	w.FilePath = filePath
	w.lastModTime = modTime
	return nil
}

// findNewestConversation finds the most recently modified conversation file
func (w *Watcher) findNewestConversation() (string, time.Time, error) {
	entries, err := os.ReadDir(w.ProjectDir)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to read project directory: %w", err)
	}

	var jsonlFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jsonl") {
			// Skip agent files - they're subagent sessions
			if !strings.HasPrefix(entry.Name(), "agent-") {
				jsonlFiles = append(jsonlFiles, entry)
			}
		}
	}

	if len(jsonlFiles) == 0 {
		return "", time.Time{}, fmt.Errorf("no conversation files found in %s", w.ProjectDir)
	}

	// Sort by modification time, newest first
	sort.Slice(jsonlFiles, func(i, j int) bool {
		infoI, _ := jsonlFiles[i].Info()
		infoJ, _ := jsonlFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	info, _ := jsonlFiles[0].Info()
	return filepath.Join(w.ProjectDir, jsonlFiles[0].Name()), info.ModTime(), nil
}

// StartLive begins watching the conversation file for new events
func (w *Watcher) StartLive() error {
	if w.FilePath == "" {
		return fmt.Errorf("no file path set, call FindProjectConversation first")
	}

	w.Mode = ModeLive

	// Open file and seek to end (we only want new events)
	file, err := os.Open(w.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open conversation file: %w", err)
	}

	// Seek to end
	pos, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to seek to end: %w", err)
	}
	w.lastPos = pos
	file.Close()

	// Emit init event
	w.Events <- Event{Type: EventSystemInit, Details: "Watching: " + filepath.Base(w.FilePath)}

	go w.tailFile()
	return nil
}

// tailFile continuously watches for new lines in the file
func (w *Watcher) tailFile() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	checkCounter := 0
	const checkInterval = 20 // Check for newer files every 2 seconds (20 * 100ms)

	for range ticker.C {
		checkCounter++

		// Periodically check if a newer conversation file was created
		if checkCounter >= checkInterval {
			checkCounter = 0
			if w.checkForNewerFile() {
				continue // Switched to new file, restart loop
			}
		}

		file, err := os.Open(w.FilePath)
		if err != nil {
			continue
		}

		// Check if file has grown
		info, err := file.Stat()
		if err != nil {
			file.Close()
			continue
		}

		if info.Size() > w.lastPos {
			// Seek to last position and read new content
			file.Seek(w.lastPos, io.SeekStart)
			scanner := bufio.NewScanner(file)

			// Increase buffer for large lines
			buf := make([]byte, 0, 1024*1024)
			scanner.Buffer(buf, 10*1024*1024)

			for scanner.Scan() {
				line := scanner.Text()
				if events := w.parseLine(line); len(events) > 0 {
					for _, evt := range events {
						w.Events <- evt
					}
				}
			}

			w.lastPos = info.Size()
		}

		file.Close()
	}
}

// checkForNewerFile checks if a newer conversation file exists and switches to it
func (w *Watcher) checkForNewerFile() bool {
	if w.ProjectDir == "" {
		return false
	}

	filePath, modTime, err := w.findNewestConversation()
	if err != nil {
		return false
	}

	// If we found a different file that's newer, switch to it
	if filePath != w.FilePath && modTime.After(w.lastModTime) {
		oldFile := filepath.Base(w.FilePath)
		newFile := filepath.Base(filePath)

		w.FilePath = filePath
		w.lastModTime = modTime
		w.lastPos = 0 // Start from beginning of new file

		// Notify about the switch
		w.Events <- Event{
			Type:    EventSystemInit,
			Details: fmt.Sprintf("Switched: %s", newFile),
		}

		// Log the switch
		fmt.Printf("Switched from %s to %s\n", oldFile, newFile)
		return true
	}

	return false
}

// StartReplay plays through an existing conversation file
func (w *Watcher) StartReplay(filePath string) error {
	w.Mode = ModeReplay
	w.FilePath = filePath

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open replay file: %w", err)
	}

	w.Events <- Event{Type: EventSystemInit, Details: "Replaying: " + filepath.Base(filePath)}

	go func() {
		defer file.Close()
		scanner := bufio.NewScanner(file)

		// Increase buffer size for large JSON lines
		buf := make([]byte, 0, 1024*1024)
		scanner.Buffer(buf, 10*1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if events := w.parseLine(line); len(events) > 0 {
				for _, evt := range events {
					w.Events <- evt
					time.Sleep(w.ReplaySpeed)
				}
			}
		}

		// Signal replay complete
		w.Events <- Event{Type: EventSuccess, Details: "Replay complete"}
	}()

	return nil
}

// parseLine parses a JSON line and returns events if applicable
func (w *Watcher) parseLine(line string) []Event {
	var msg ClaudeMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return nil
	}

	var events []Event

	switch msg.Type {
	case "system":
		events = w.parseSystemMessage(msg)

	case "assistant":
		events = w.parseAssistantMessage(msg)

	case "user":
		events = w.parseUserMessage(msg)

	case "result":
		events = w.parseResultMessage(msg)

	case "summary":
		// Conversation was summarized (after compact)
		if msg.Summary != "" {
			events = append(events, Event{
				Type:    EventIdle,
				Details: truncate(msg.Summary, 50),
			})
		}
	}

	return events
}

// parseSystemMessage handles system type messages
func (w *Watcher) parseSystemMessage(msg ClaudeMessage) []Event {
	switch msg.Subtype {
	case "compact_boundary":
		// Conversation was compacted - Claude should rest/sleep
		evt := Event{
			Type:    EventCompact,
			Details: "Conversation compacted",
		}
		if msg.CompactMetadata != nil {
			evt.CompactInfo = msg.CompactMetadata
			evt.Details = fmt.Sprintf("Compacted from %dk tokens", msg.CompactMetadata.PreTokens/1000)
		}
		return []Event{evt}

	case "local_command":
		// User ran a slash command - just acknowledge
		return nil

	default:
		// Session start or other system event
		return []Event{{Type: EventSystemInit, Details: "Session started"}}
	}
}

// parseAssistantMessage handles assistant type messages
func (w *Watcher) parseAssistantMessage(msg ClaudeMessage) []Event {
	var events []Event

	// Update token usage for mana bar
	if msg.Message.Usage != nil {
		w.LastTokenUsage = msg.Message.Usage
	}

	// Parse content array
	content := w.parseMessageContent(msg.Message.Content)

	for _, item := range content {
		switch item.Type {
		case "tool_use":
			evt := w.parseToolUse(item)
			if evt != nil {
				evt.TokenUsage = w.LastTokenUsage
				events = append(events, *evt)
			}

		case "thinking":
			// Extended thinking block
			thinkLen := len(item.Thinking)
			details := "Thinking..."
			if thinkLen > 500 {
				details = "Deep thinking..."
			}
			events = append(events, Event{
				Type:        EventThinking,
				Details:     details,
				TokenUsage:  w.LastTokenUsage,
				ThoughtText: item.Thinking,
			})

		case "text":
			if len(item.Text) > 0 {
				events = append(events, Event{
					Type:       EventThinking,
					Details:    truncate(item.Text, 40),
					TokenUsage: w.LastTokenUsage,
				})
			}
		}
	}

	return events
}

// parseUserMessage handles user type messages
func (w *Watcher) parseUserMessage(msg ClaudeMessage) []Event {
	var events []Event

	content := w.parseMessageContent(msg.Message.Content)

	// Check if this is a tool result or a user prompt
	hasToolResult := false
	hasError := false

	for _, item := range content {
		if item.Type == "tool_result" {
			hasToolResult = true

			// Check if this is a Task completion (agent should poof)
			if item.ToolUseID != "" {
				if agentType, ok := w.ActiveTaskAgents[item.ToolUseID]; ok {
					delete(w.ActiveTaskAgents, item.ToolUseID)
					events = append(events, Event{
						Type:      EventAgentComplete,
						Details:   agentType,
						ToolUseID: item.ToolUseID,
					})
				}
			}

			if item.IsError {
				hasError = true
				// Extract error content (could be string or array)
				errorDetails := "Error"
				if len(item.Content) > 0 {
					errorDetails = truncate(string(item.Content), 40)
				}
				events = append(events, Event{
					Type:    EventError,
					Details: errorDetails,
					IsError: true,
				})
			}
		}
	}

	// If not a tool result, this is a user prompt (quest!)
	if !hasToolResult {
		text := w.extractUserPromptText(msg.Message.Content)
		if text != "" {
			// Check for think hard patterns
			thinkLevel := detectThinkLevel(text)
			if thinkLevel != ThinkNone {
				events = append(events, Event{
					Type:       EventThinkHard,
					Details:    truncate(text, 50),
					ThinkLevel: thinkLevel,
				})
			} else {
				events = append(events, Event{
					Type:    EventQuest,
					Details: truncate(text, 100),
				})
			}
		}
	}

	// Return error event if there was one
	if hasError && len(events) == 0 {
		events = append(events, Event{
			Type:    EventError,
			Details: "Tool error",
			IsError: true,
		})
	}

	return events
}

// parseResultMessage handles result type messages
func (w *Watcher) parseResultMessage(msg ClaudeMessage) []Event {
	switch msg.Subtype {
	case "success":
		return []Event{{Type: EventSuccess, Details: "Task completed!"}}
	case "error_max_turns", "error_during_execution":
		return []Event{{Type: EventError, Details: "Something went wrong", IsError: true}}
	}
	return nil
}

// parseMessageContent parses the message content which can be string or array
func (w *Watcher) parseMessageContent(raw json.RawMessage) []ContentItem {
	if len(raw) == 0 {
		return nil
	}

	// Try as array first
	var items []ContentItem
	if err := json.Unmarshal(raw, &items); err == nil {
		return items
	}

	// Try as string (user prompt)
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return []ContentItem{{Type: "text", Text: text}}
	}

	return nil
}

// extractUserPromptText extracts the user's text from content
func (w *Watcher) extractUserPromptText(raw json.RawMessage) string {
	// Try as string first
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}

	// Try as array
	var items []ContentItem
	if err := json.Unmarshal(raw, &items); err == nil {
		for _, item := range items {
			if item.Type == "text" && item.Text != "" {
				return item.Text
			}
		}
	}

	return ""
}

// parseToolUse handles tool_use content items
func (w *Watcher) parseToolUse(item ContentItem) *Event {
	toolName := strings.ToLower(item.Name)

	switch {
	// Reading tools
	case toolName == "glob" || toolName == "read" || toolName == "grep":
		return &Event{Type: EventReading, Details: "Reading files", ToolName: item.Name}

	case toolName == "websearch" || toolName == "webfetch":
		return &Event{Type: EventReading, Details: "Searching web", ToolName: item.Name}

	// Bash execution
	case toolName == "bash":
		// Check if this is a git push command
		var bashInput struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal(item.Input, &bashInput); err == nil {
			cmd := strings.ToLower(bashInput.Command)
			if strings.Contains(cmd, "git push") || strings.Contains(cmd, "git push") {
				return &Event{Type: EventGitPush, Details: "SHIPPED!", ToolName: item.Name}
			}
		}
		return &Event{Type: EventBash, Details: "Running command", ToolName: item.Name}

	case toolName == "killshell":
		return &Event{Type: EventBash, Details: "Stopping process", ToolName: item.Name}

	// Writing tools
	case toolName == "edit" || toolName == "write" || toolName == "notebookedit":
		return &Event{Type: EventWriting, Details: "Writing code", ToolName: item.Name}

	// Task/Agent spawning
	case toolName == "task":
		evt := &Event{Type: EventSpawnAgent, Details: "Spawning agent", ToolName: item.Name, ToolUseID: item.ID}
		agentType := "Agent"
		// Try to extract agent type from input
		var taskInput TaskInput
		if err := json.Unmarshal(item.Input, &taskInput); err == nil {
			if taskInput.SubagentType != "" {
				agentType = taskInput.SubagentType
				evt.Details = fmt.Sprintf("Agent: %s", taskInput.SubagentType)
			} else if taskInput.Description != "" {
				evt.Details = truncate(taskInput.Description, 30)
			}
		}
		// Track this Task for poof detection when it completes
		if item.ID != "" {
			w.ActiveTaskAgents[item.ID] = agentType
		}
		return evt

	case toolName == "taskoutput":
		return &Event{Type: EventThinking, Details: "Waiting for agent", ToolName: item.Name}

	// Todo management
	case toolName == "todowrite":
		evt := &Event{Type: EventTodoUpdate, Details: "Updating tasks", ToolName: item.Name}
		// Parse todos from input
		var todoInput TodoWriteInput
		if err := json.Unmarshal(item.Input, &todoInput); err == nil {
			evt.TodoItems = todoInput.Todos
			w.CurrentTodos = todoInput.Todos

			// Count status
			inProgress := 0
			completed := 0
			for _, t := range todoInput.Todos {
				if t.Status == "in_progress" {
					inProgress++
				} else if t.Status == "completed" {
					completed++
				}
			}
			evt.Details = fmt.Sprintf("Tasks: %d/%d done", completed, len(todoInput.Todos))
		}
		return evt

	// User interaction
	case toolName == "askuserquestion":
		return &Event{Type: EventAskUser, Details: "Asking question", ToolName: item.Name}

	case toolName == "exitplanmode":
		return &Event{Type: EventThinking, Details: "Plan ready", ToolName: item.Name}

	// Skills
	case toolName == "skill":
		return &Event{Type: EventThinking, Details: "Running skill", ToolName: item.Name}

	default:
		return &Event{Type: EventThinking, Details: "Using " + item.Name, ToolName: item.Name}
	}
}

// ThinkLevel represents intensity of thinking request
type ThinkLevel int

const (
	ThinkNone    ThinkLevel = iota
	ThinkNormal             // "think"
	ThinkHard               // "think hard"
	ThinkHarder             // "think harder"
	ThinkUltra              // "ultrathink"
)

// detectThinkLevel checks user message for thinking intensity
func detectThinkLevel(text string) ThinkLevel {
	lower := strings.ToLower(text)

	// Check in order of specificity (most specific first)
	if strings.Contains(lower, "ultrathink") {
		return ThinkUltra
	}
	if strings.Contains(lower, "think harder") {
		return ThinkHarder
	}
	if strings.Contains(lower, "think hard") || strings.Contains(lower, "think deeply") ||
		strings.Contains(lower, "think carefully") || strings.Contains(lower, "deep think") {
		return ThinkHard
	}
	if strings.Contains(lower, "really think") {
		return ThinkNormal
	}

	return ThinkNone
}

// isThinkHard checks if user message requests extended thinking (any level)
func isThinkHard(text string) bool {
	return detectThinkLevel(text) != ThinkNone
}

func truncate(s string, maxLen int) string {
	// Remove newlines for cleaner display
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)

	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
