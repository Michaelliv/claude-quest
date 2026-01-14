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
	EventReading  // Glob, Read tools
	EventBash     // Bash tool
	EventWriting  // Edit, Write tools
	EventSuccess  // Successful result
	EventError    // Error result
	EventIdle     // No activity
)

// Event represents a parsed Claude Code event
type Event struct {
	Type    EventType
	Details string
}

// ClaudeMessage represents the structure of Claude Code JSONL format
type ClaudeMessage struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype,omitempty"`
	Message struct {
		Role    string `json:"role"`
		Content []struct {
			Type  string `json:"type"`
			Name  string `json:"name,omitempty"`
			Text  string `json:"text,omitempty"`
			Input any    `json:"input,omitempty"`
		} `json:"content"`
	} `json:"message,omitempty"`
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
	ReplaySpeed time.Duration // Delay between events in replay mode
	lastPos     int64         // Last read position for tailing
}

// NewWatcher creates a new event watcher
func NewWatcher() *Watcher {
	return &Watcher{
		Events:      make(chan Event, 100),
		ReplaySpeed: 200 * time.Millisecond, // Default replay speed
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

	// Check if project directory exists
	if _, err := os.Stat(claudeProjectDir); os.IsNotExist(err) {
		return fmt.Errorf("no Claude conversations found for %s\nlooked in: %s", projectDir, claudeProjectDir)
	}

	// Find the most recently modified .jsonl file
	entries, err := os.ReadDir(claudeProjectDir)
	if err != nil {
		return fmt.Errorf("failed to read project directory: %w", err)
	}

	var jsonlFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jsonl") {
			jsonlFiles = append(jsonlFiles, entry)
		}
	}

	if len(jsonlFiles) == 0 {
		return fmt.Errorf("no conversation files found in %s", claudeProjectDir)
	}

	// Sort by modification time, newest first
	sort.Slice(jsonlFiles, func(i, j int) bool {
		infoI, _ := jsonlFiles[i].Info()
		infoJ, _ := jsonlFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	w.FilePath = filepath.Join(claudeProjectDir, jsonlFiles[0].Name())
	return nil
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

	for range ticker.C {
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

			for scanner.Scan() {
				line := scanner.Text()
				if event := w.parseLine(line); event != nil {
					w.Events <- *event
				}
			}

			w.lastPos = info.Size()
		}

		file.Close()
	}
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
			if event := w.parseLine(line); event != nil {
				w.Events <- *event
				time.Sleep(w.ReplaySpeed)
			}
		}

		// Signal replay complete
		w.Events <- Event{Type: EventSuccess, Details: "Replay complete"}
	}()

	return nil
}

// parseLine parses a JSON line and returns an event if applicable
func (w *Watcher) parseLine(line string) *Event {
	var msg ClaudeMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return nil
	}

	switch msg.Type {
	case "system":
		return &Event{Type: EventSystemInit, Details: "Session started"}

	case "assistant":
		// Check for tool use in content
		for _, content := range msg.Message.Content {
			if content.Type == "tool_use" {
				return w.parseToolUse(content.Name)
			}
			if content.Type == "thinking" {
				return &Event{Type: EventThinking, Details: "Thinking..."}
			}
			if content.Type == "text" && len(content.Text) > 0 {
				return &Event{Type: EventThinking, Details: truncate(content.Text, 30)}
			}
		}

	case "user":
		// Tool results come as user messages
		for _, content := range msg.Message.Content {
			if content.Type == "tool_result" {
				return nil // Tool result received, animation handled by tool_use
			}
		}

	case "result":
		switch msg.Subtype {
		case "success":
			return &Event{Type: EventSuccess, Details: "Task completed!"}
		case "error_max_turns", "error_during_execution":
			return &Event{Type: EventError, Details: "Something went wrong"}
		}
	}

	return nil
}

// parseToolUse maps tool names to event types
func (w *Watcher) parseToolUse(toolName string) *Event {
	toolName = strings.ToLower(toolName)

	switch {
	case toolName == "glob" || toolName == "read" || toolName == "grep":
		return &Event{Type: EventReading, Details: "Reading " + toolName}
	case toolName == "bash":
		return &Event{Type: EventBash, Details: "Executing command"}
	case toolName == "edit" || toolName == "write" || toolName == "notebookedit":
		return &Event{Type: EventWriting, Details: "Writing code"}
	case toolName == "task":
		return &Event{Type: EventThinking, Details: "Spawning agent"}
	default:
		return &Event{Type: EventThinking, Details: "Using " + toolName}
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
