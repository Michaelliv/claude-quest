package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Provider identifies the source of a normalized activity event.
type Provider string

const (
	ProviderClaude Provider = "claude"
	ProviderCodex  Provider = "codex"
)

// NormalizedEventType is intentionally limited to metadata-safe activities.
type NormalizedEventType string

const (
	NormalizedSessionStarted NormalizedEventType = "session_started"
	NormalizedToolCompleted  NormalizedEventType = "tool_completed"
	NormalizedFileChanged    NormalizedEventType = "file_changed"
	NormalizedTestCompleted  NormalizedEventType = "test_completed"
	NormalizedGitMilestone   NormalizedEventType = "git_milestone"
	NormalizedSessionEnded   NormalizedEventType = "session_ended"
)

type EventStatus string

const (
	EventStatusSuccess EventStatus = "success"
	EventStatusFailure EventStatus = "failure"
	EventStatusUnknown EventStatus = "unknown"
)

type TestResult string

const (
	TestResultNone   TestResult = "none"
	TestResultPassed TestResult = "passed"
	TestResultFailed TestResult = "failed"
)

type GitMilestone string

const (
	GitMilestoneNone   GitMilestone = "none"
	GitMilestoneCommit GitMilestone = "commit"
	GitMilestonePR     GitMilestone = "pr"
	GitMilestoneMerge  GitMilestone = "merge"
)

// NormalizedEvent never contains prompts, commands, paths, tool payloads, or credentials.
type NormalizedEvent struct {
	Provider        Provider            `json:"provider"`
	SessionID       string              `json:"session_id"`
	ProjectID       string              `json:"project_id"`
	EventType       NormalizedEventType `json:"event_type"`
	Timestamp       time.Time           `json:"timestamp"`
	Status          EventStatus         `json:"status"`
	DurationMS      int64               `json:"duration_ms,omitempty"`
	FileChangeCount int                 `json:"file_change_count,omitempty"`
	TestResult      TestResult          `json:"test_result,omitempty"`
	GitMilestone    GitMilestone        `json:"git_milestone,omitempty"`
}

// CodexProviderConfig contains only non-secret local behavior settings.
type CodexProviderConfig struct {
	Enabled         bool     `json:"enabled"`
	KillSwitch      bool     `json:"kill_switch"`
	AllowedProjects []string `json:"allowed_projects"`
	RetentionDays   int      `json:"retention_days"`
	MaxBytes        int64    `json:"max_bytes"`
	StateDir        string   `json:"-"`
}

func DefaultCodexProviderConfig() CodexProviderConfig {
	return CodexProviderConfig{
		Enabled:       false,
		KillSwitch:    false,
		RetentionDays: 14,
		MaxBytes:      25 * 1024 * 1024,
	}
}

func codexProviderConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude-quest", "codex-provider.json"), nil
}

func LoadCodexProviderConfig() (CodexProviderConfig, error) {
	config := DefaultCodexProviderConfig()
	path, err := codexProviderConfigPath()
	if err != nil {
		return config, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return config, nil
	}
	if err != nil {
		return config, err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultCodexProviderConfig(), err
	}
	if config.RetentionDays <= 0 {
		config.RetentionDays = DefaultCodexProviderConfig().RetentionDays
	}
	if config.MaxBytes <= 0 {
		config.MaxBytes = DefaultCodexProviderConfig().MaxBytes
	}
	return config, nil
}

func canonicalProjectPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return filepath.Clean(resolved), nil
	}
	return filepath.Clean(abs), nil
}

func IsAllowedProject(config CodexProviderConfig, project string) bool {
	if !config.Enabled || config.KillSwitch {
		return false
	}
	canonicalProject, err := canonicalProjectPath(project)
	if err != nil {
		return false
	}
	for _, allowed := range config.AllowedProjects {
		canonicalAllowed, err := canonicalProjectPath(allowed)
		if err == nil && canonicalProject == canonicalAllowed {
			return true
		}
	}
	return false
}

// CodexHookInput is a narrow, metadata-only view of a Hook payload. Unknown JSON fields are ignored.
type CodexHookInput struct {
	SessionID       string    `json:"session_id"`
	CWD             string    `json:"cwd"`
	EventName       string    `json:"hook_event_name"`
	ToolName        string    `json:"tool_name"`
	Status          string    `json:"status"`
	DurationMS      int64     `json:"duration_ms"`
	FileChangeCount int       `json:"file_change_count"`
	TestResult      string    `json:"test_result"`
	GitMilestone    string    `json:"git_milestone"`
	Timestamp       time.Time `json:"timestamp"`
}

func opaqueID(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:12])
}

func normalizedStatus(value string) EventStatus {
	switch strings.ToLower(value) {
	case "success", "passed", "ok":
		return EventStatusSuccess
	case "failure", "failed", "error":
		return EventStatusFailure
	default:
		return EventStatusUnknown
	}
}

func normalizedTestResult(value string) TestResult {
	switch strings.ToLower(value) {
	case "passed", "pass", "success":
		return TestResultPassed
	case "failed", "fail", "failure":
		return TestResultFailed
	default:
		return TestResultNone
	}
}

func normalizedGitMilestone(value string) GitMilestone {
	switch strings.ToLower(value) {
	case "commit":
		return GitMilestoneCommit
	case "pr":
		return GitMilestonePR
	case "merge":
		return GitMilestoneMerge
	default:
		return GitMilestoneNone
	}
}

func normalizedEventType(eventName, toolName string, testResult TestResult, milestone GitMilestone) NormalizedEventType {
	if milestone != GitMilestoneNone {
		return NormalizedGitMilestone
	}
	if testResult != TestResultNone {
		return NormalizedTestCompleted
	}
	if strings.EqualFold(toolName, "apply_patch") || strings.EqualFold(toolName, "Edit") || strings.EqualFold(toolName, "Write") {
		return NormalizedFileChanged
	}
	switch eventName {
	case "SessionStart":
		return NormalizedSessionStarted
	case "SessionEnd":
		return NormalizedSessionEnded
	default:
		return NormalizedToolCompleted
	}
}

// SanitizeCodexHook validates the project boundary and converts metadata to a safe event.
func SanitizeCodexHook(config CodexProviderConfig, input CodexHookInput) (NormalizedEvent, bool) {
	if !IsAllowedProject(config, input.CWD) || input.SessionID == "" {
		return NormalizedEvent{}, false
	}
	testResult := normalizedTestResult(input.TestResult)
	milestone := normalizedGitMilestone(input.GitMilestone)
	timestamp := input.Timestamp.UTC()
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}
	fileChanges := input.FileChangeCount
	if fileChanges < 0 {
		fileChanges = 0
	}
	if fileChanges == 0 && (strings.EqualFold(input.ToolName, "apply_patch") || strings.EqualFold(input.ToolName, "Edit") || strings.EqualFold(input.ToolName, "Write")) {
		fileChanges = 1
	}
	return NormalizedEvent{
		Provider:        ProviderCodex,
		SessionID:       opaqueID(input.SessionID),
		ProjectID:       opaqueID(mustCanonicalProjectPath(input.CWD)),
		EventType:       normalizedEventType(input.EventName, input.ToolName, testResult, milestone),
		Timestamp:       timestamp,
		Status:          normalizedStatus(input.Status),
		DurationMS:      maxInt64(input.DurationMS, 0),
		FileChangeCount: fileChanges,
		TestResult:      testResult,
		GitMilestone:    milestone,
	}, true
}

func mustCanonicalProjectPath(path string) string {
	canonical, err := canonicalProjectPath(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return canonical
}

func maxInt64(value, min int64) int64 {
	if value < min {
		return min
	}
	return value
}

// CalculateXP rewards completed, observable outcomes and intentionally ignores token usage.
func CalculateXP(event NormalizedEvent) int {
	xp := event.FileChangeCount * 3
	if event.EventType == NormalizedTestCompleted {
		xp += 5
	}
	switch event.TestResult {
	case TestResultPassed:
		xp += 5
	case TestResultFailed:
		xp -= 2
	}
	switch event.GitMilestone {
	case GitMilestoneCommit:
		xp += 8
	case GitMilestonePR:
		xp += 12
	case GitMilestoneMerge:
		xp += 20
	}
	if event.Status == EventStatusFailure {
		xp -= 2
	}
	return xp
}
