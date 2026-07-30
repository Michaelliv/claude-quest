package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// CodexSpool persists only sanitized normalized events in a local owner-only directory.
type CodexSpool struct {
	config CodexProviderConfig
}

func NewCodexSpool(config CodexProviderConfig) *CodexSpool {
	return &CodexSpool{config: config}
}

func (s *CodexSpool) directory() (string, error) {
	if s.config.StateDir != "" {
		return s.config.StateDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude-quest", "codex-provider"), nil
}

func (s *CodexSpool) CurrentPath() string {
	dir, err := s.directory()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "events.jsonl")
}

func (s *CodexSpool) Ensure() error {
	dir, err := s.directory()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(dir, "events.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	return file.Chmod(0600)
}

// ProcessCodexHook decodes only the narrow hook metadata schema and rejects all disallowed events.
func ProcessCodexHook(input io.Reader, config CodexProviderConfig, spool *CodexSpool) error {
	decoder := json.NewDecoder(io.LimitReader(input, 1024*1024))
	var hook CodexHookInput
	if err := decoder.Decode(&hook); err != nil {
		return fmt.Errorf("invalid Codex hook metadata: %w", err)
	}
	event, ok := SanitizeCodexHook(config, hook)
	if !ok {
		return fmt.Errorf("Codex event rejected by local provider policy")
	}
	return spool.Append(event)
}

func (s *CodexSpool) Append(event NormalizedEvent) error {
	if event.Provider != ProviderCodex {
		return fmt.Errorf("refusing non-Codex event")
	}
	dir, err := s.directory()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return err
	}
	line, err := json.Marshal(event)
	if err != nil {
		return err
	}
	line = append(line, '\n')
	if err := s.rotateIfNeeded(dir, int64(len(line))); err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(dir, "events.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := file.Chmod(0600); err != nil {
		return err
	}
	if _, err = file.Write(line); err != nil {
		return err
	}
	return s.prune(dir)
}

func (s *CodexSpool) rotateIfNeeded(dir string, incomingBytes int64) error {
	current := filepath.Join(dir, "events.jsonl")
	if info, err := os.Stat(current); err == nil && info.Size()+incomingBytes > s.config.MaxBytes {
		archive := filepath.Join(dir, fmt.Sprintf("events-%s.jsonl", time.Now().UTC().Format("20060102T150405.000000000Z")))
		if err := os.Rename(current, archive); err != nil {
			return err
		}
	}
	return s.prune(dir)
}

func (s *CodexSpool) prune(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	type candidate struct {
		path string
		info os.FileInfo
	}
	var files []candidate
	var total int64
	cutoff := time.Now().AddDate(0, 0, -s.config.RetentionDays)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "events") || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		path := filepath.Join(dir, entry.Name())
		if info.ModTime().Before(cutoff) && entry.Name() != "events.jsonl" {
			if err := os.Remove(path); err != nil {
				return err
			}
			continue
		}
		files = append(files, candidate{path: path, info: info})
		total += info.Size()
	}
	sort.Slice(files, func(i, j int) bool { return files[i].info.ModTime().Before(files[j].info.ModTime()) })
	for _, file := range files {
		if total <= s.config.MaxBytes || filepath.Base(file.path) == "events.jsonl" {
			continue
		}
		if err := os.Remove(file.path); err != nil {
			return err
		}
		total -= file.info.Size()
	}
	return nil
}

// CodexAdapter maps safe provider events onto the stable existing renderer events.
type CodexAdapter struct{}

func (CodexAdapter) ToEvent(event NormalizedEvent) Event {
	base := Event{Provider: ProviderCodex, XPDelta: CalculateXP(event)}
	switch {
	case event.TestResult == TestResultPassed:
		base.Type, base.Details = EventSuccess, "Codex: test passed"
	case event.TestResult == TestResultFailed || event.Status == EventStatusFailure:
		base.Type, base.Details, base.IsError = EventError, "Codex: task failed", true
	case event.GitMilestone == GitMilestoneMerge:
		base.Type, base.Details = EventGitPush, "Codex: merged"
	case event.GitMilestone == GitMilestoneCommit || event.GitMilestone == GitMilestonePR:
		base.Type, base.Details = EventVictoryPose, "Codex: shipped"
	case event.EventType == NormalizedFileChanged:
		base.Type, base.Details = EventWriting, "Codex: writing"
	case event.EventType == NormalizedSessionStarted:
		base.Type, base.Details = EventSystemInit, "Codex session started"
	case event.EventType == NormalizedSessionEnded:
		base.Type, base.Details = EventSuccess, "Codex session completed"
	default:
		base.Type, base.Details = EventBash, "Codex: working"
	}
	return base
}

// ParseCodexFallbackLine accepts only an exact normalized Codex record. It rejects transcript payload fields.
func ParseCodexFallbackLine(line []byte) (NormalizedEvent, bool) {
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.DisallowUnknownFields()
	var event NormalizedEvent
	if err := decoder.Decode(&event); err != nil {
		return NormalizedEvent{}, false
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return NormalizedEvent{}, false
	}
	if event.Provider != ProviderCodex || event.EventType == "" || event.SessionID == "" || event.ProjectID == "" {
		return NormalizedEvent{}, false
	}
	return event, true
}

// CodexSessionWatcher tails a configured session JSONL file read-only and emits only valid normalized records.
// Native Codex transcripts are intentionally not parsed because their wire format is not a stable interface.
type CodexSessionWatcher struct {
	Events  chan NormalizedEvent
	file    string
	lastPos int64
}

func NewCodexSessionWatcher(file string) *CodexSessionWatcher {
	return &CodexSessionWatcher{Events: make(chan NormalizedEvent, 100), file: file}
}

func (w *CodexSessionWatcher) Start() error {
	file, err := os.Open(w.file)
	if err != nil {
		return err
	}
	defer file.Close()
	position, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	w.lastPos = position
	go w.tail()
	return nil
}

func (w *CodexSessionWatcher) tail() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		file, err := os.Open(w.file)
		if err != nil {
			continue
		}
		info, err := file.Stat()
		if err != nil || info.Size() <= w.lastPos {
			file.Close()
			continue
		}
		_, _ = file.Seek(w.lastPos, io.SeekStart)
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			if event, ok := ParseCodexFallbackLine(scanner.Bytes()); ok {
				w.Events <- event
			}
		}
		w.lastPos = info.Size()
		file.Close()
	}
}
