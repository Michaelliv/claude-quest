package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultCodexProviderConfigDeniesCollection(t *testing.T) {
	config := DefaultCodexProviderConfig()
	if config.Enabled {
		t.Fatal("default config must keep collection disabled")
	}
	if config.KillSwitch {
		t.Fatal("default config must not report an active collection path")
	}
	if IsAllowedProject(config, t.TempDir()) {
		t.Fatal("empty allowlist must deny every project")
	}
}

func TestIsAllowedProjectMatchesOnlyCanonicalAllowlistEntry(t *testing.T) {
	root := t.TempDir()
	allowed := filepath.Join(root, "allowed")
	disallowed := filepath.Join(root, "allowed-copy")
	config := CodexProviderConfig{Enabled: true, AllowedProjects: []string{allowed}}

	if !IsAllowedProject(config, filepath.Join(allowed, ".")) {
		t.Fatal("canonical spelling of an allowed project should be accepted")
	}
	if IsAllowedProject(config, disallowed) {
		t.Fatal("prefix-matching must not grant access to another project")
	}
}

func TestSanitizeCodexHookHashesIdentityAndDropsRawFields(t *testing.T) {
	project := t.TempDir()
	config := CodexProviderConfig{Enabled: true, AllowedProjects: []string{project}}
	event, ok := SanitizeCodexHook(config, CodexHookInput{
		SessionID:  "session-with-private-context",
		CWD:        project,
		EventName:  "PostToolUse",
		ToolName:   "apply_patch",
		Status:     "success",
		DurationMS: 42,
		Timestamp:  time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC),
	})
	if !ok {
		t.Fatal("allowed hook event should be accepted")
	}
	if event.SessionID == "session-with-private-context" || event.SessionID == "" {
		t.Fatalf("session ID must be hashed, got %q", event.SessionID)
	}
	if event.ProjectID == project || event.ProjectID == "" {
		t.Fatalf("project path must not be persisted, got %q", event.ProjectID)
	}
	if event.EventType != NormalizedFileChanged || event.Status != EventStatusSuccess || event.DurationMS != 42 {
		t.Fatalf("unexpected sanitized event: %+v", event)
	}
}

func TestCalculateXPUsesCompletedOutcomesInsteadOfTokenTotals(t *testing.T) {
	event := NormalizedEvent{
		Provider:        ProviderCodex,
		EventType:       NormalizedTestCompleted,
		Status:          EventStatusSuccess,
		FileChangeCount: 2,
		TestResult:      TestResultPassed,
		GitMilestone:    GitMilestoneCommit,
	}
	if got, want := CalculateXP(event), 24; got != want {
		t.Fatalf("CalculateXP() = %d, want %d", got, want)
	}
}
