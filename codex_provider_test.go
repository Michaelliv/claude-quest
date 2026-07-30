package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testCodexConfig(t *testing.T) (CodexProviderConfig, string) {
	t.Helper()
	project := t.TempDir()
	return CodexProviderConfig{
		Enabled:         true,
		AllowedProjects: []string{project},
		RetentionDays:   14,
		MaxBytes:        1024,
		StateDir:        t.TempDir(),
	}, project
}

func TestCodexSpoolStoresOnlyNormalizedMetadata(t *testing.T) {
	config, project := testCodexConfig(t)
	spool := NewCodexSpool(config)
	event, ok := SanitizeCodexHook(config, CodexHookInput{
		SessionID: "private-session", CWD: project, EventName: "PostToolUse", ToolName: "apply_patch", Status: "success",
	})
	if !ok {
		t.Fatal("expected test hook to be allowed")
	}
	if err := spool.Append(event); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	data, err := os.ReadFile(spool.CurrentPath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(data)
	for _, forbidden := range []string{"private-session", project, "tool_name", "command", "prompt"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("spool retained forbidden value %q: %s", forbidden, text)
		}
	}
	info, err := os.Stat(spool.CurrentPath())
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("spool permissions = %o, want 600", info.Mode().Perm())
	}
}

func TestCodexSpoolRotatesAtConfiguredCap(t *testing.T) {
	config, _ := testCodexConfig(t)
	config.MaxBytes = 250
	spool := NewCodexSpool(config)
	event := NormalizedEvent{Provider: ProviderCodex, SessionID: "safe", ProjectID: "safe", EventType: NormalizedToolCompleted, Timestamp: time.Now().UTC(), Status: EventStatusSuccess}
	if err := spool.Append(event); err != nil {
		t.Fatalf("first Append() error = %v", err)
	}
	if err := spool.Append(event); err != nil {
		t.Fatalf("second Append() error = %v", err)
	}
	entries, err := filepath.Glob(filepath.Join(config.StateDir, "events*.jsonl"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	var total int64
	for _, path := range entries {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat(%q) error = %v", path, err)
		}
		total += info.Size()
	}
	if total > config.MaxBytes {
		t.Fatalf("spool retained %d bytes, cap is %d", total, config.MaxBytes)
	}
}

func TestCodexAdapterMapsTestAndGitEventsToExistingAnimations(t *testing.T) {
	adapter := CodexAdapter{}
	passed := adapter.ToEvent(NormalizedEvent{Provider: ProviderCodex, EventType: NormalizedTestCompleted, Status: EventStatusSuccess, TestResult: TestResultPassed})
	if passed.Type != EventSuccess || passed.Details != "Codex: test passed" {
		t.Fatalf("passed test mapped to %+v", passed)
	}
	merged := adapter.ToEvent(NormalizedEvent{Provider: ProviderCodex, EventType: NormalizedGitMilestone, Status: EventStatusSuccess, GitMilestone: GitMilestoneMerge})
	if merged.Type != EventGitPush || merged.Details != "Codex: merged" {
		t.Fatalf("merged milestone mapped to %+v", merged)
	}
}

func TestParseFallbackLineRejectsRawAndForeignEvents(t *testing.T) {
	if _, ok := ParseCodexFallbackLine([]byte(`{"provider":"claude","event_type":"tool_completed"}`)); ok {
		t.Fatal("foreign provider event must be rejected")
	}
	if _, ok := ParseCodexFallbackLine([]byte(`{"provider":"codex","event_type":"tool_completed","command":"git status"}`)); ok {
		t.Fatal("unknown raw fields must be rejected")
	}
}

func TestProcessCodexHookWritesAllowedSanitizedEvent(t *testing.T) {
	config, project := testCodexConfig(t)
	spool := NewCodexSpool(config)
	input := bytes.NewBufferString(`{"session_id":"session-1","cwd":"` + project + `","hook_event_name":"PostToolUse","tool_name":"apply_patch","status":"success"}`)
	if err := ProcessCodexHook(input, config, spool); err != nil {
		t.Fatalf("ProcessCodexHook() error = %v", err)
	}
	if _, err := os.Stat(spool.CurrentPath()); err != nil {
		t.Fatalf("sanitized event was not written: %v", err)
	}
}
