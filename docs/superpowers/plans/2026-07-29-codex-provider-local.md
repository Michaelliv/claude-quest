# Codex Local Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a default-deny, local-only Codex provider that drives existing Claude Quest animations and outcome-based XP.

**Architecture:** New provider files define a sanitized `NormalizedEvent`, default-deny local configuration, bounded local spool, adapter, and CLI commands. The existing Claude JSONL watcher remains unchanged; Codex hooks and session fallback feed only sanitized metadata into the common event mapping.

**Tech Stack:** Go 1.25 standard library, existing Raylib application, JSON configuration and JSONL spool.

## Global Constraints

- Never read Codex authentication, token, API-key, or environment-secret sources.
- Never persist prompts, commands, file contents, tool payloads, outputs, or transcript paths.
- Local file and local IPC only; no network client or listener.
- Codex collection is off by default and requires a canonical project allowlist.
- Preserve current Claude watcher behavior and JSONL parsing.

---

### Task 1: Provider contracts and policy

**Files:**
- Create: `provider.go`
- Create: `provider_test.go`

**Interfaces:**
- Produces `NormalizedEvent`, `CodexProviderConfig`, `LoadCodexProviderConfig`, `IsAllowedProject`, `CalculateXP`.

- [ ] **Step 1: Write failing tests** for default-deny configuration, canonical allowlist matching, raw-field sanitization, and literal outcome XP values.
- [ ] **Step 2: Run** `go test ./...` and verify the missing provider symbols fail compilation.
- [ ] **Step 3: Implement** the smallest provider contract and policy needed for the tested behavior.
- [ ] **Step 4: Run** `go test ./...` and verify the provider tests pass.

### Task 2: Local spool and adapters

**Files:**
- Create: `codex_provider.go`
- Create: `codex_provider_test.go`

**Interfaces:**
- Consumes `NormalizedEvent`, `CodexProviderConfig`.
- Produces `CodexAdapter`, `CodexSpool`, `CodexSessionWatcher`.

- [ ] **Step 1: Write failing tests** for allowlisted hook acceptance, default rejection, bounded retention, raw-field exclusion, and animation mapping.
- [ ] **Step 2: Run** `go test ./...` and verify the new behavior fails.
- [ ] **Step 3: Implement** owner-only spool writes, rotation, hook parsing, read-only fallback validation, and event mapping.
- [ ] **Step 4: Run** `go test ./...` and verify all tests pass.

### Task 3: CLI, documentation, and smoke test

**Files:**
- Modify: `main.go`
- Modify: `README.md`
- Modify: `CLAUDE.md`
- Create: `scripts/smoke-codex-provider.sh`

- [ ] **Step 1: Write failing CLI-focused test coverage** for command arguments or a command-independent integration seam.
- [ ] **Step 2: Run** `go test ./...` and confirm failure.
- [ ] **Step 3: Implement** `codex-hook` and `codex-watch` commands without changing existing commands; document trust, enablement, kill switch, allowlist, fallback limits, and retention.
- [ ] **Step 4: Run** `go test ./...`, `go vet ./...`, build, and the local smoke script.

### Task 4: Review and integration

**Files:**
- Review all task-owned files.

- [ ] **Step 1: Run** focused and full test, vet, build, and smoke commands.
- [ ] **Step 2: Perform** a security review for credential access, raw-data retention, path traversal, permissive defaults, and network use.
- [ ] **Step 3: Commit** the scoped changes, merge the local feature branch into local `main`, and install the verified local build over the approved global package.
