# Codex Local Provider Design

## Goal

Add Codex activity to Claude Quest without changing Claude event compatibility or sending, retaining, or displaying sensitive source data.

## Scope and safety boundary

- All collection and storage is local-only; no network client is introduced.
- The collector never reads `~/.codex/auth.json`, environment values, tokens, API keys, or other credential stores.
- Raw prompts, commands, tool inputs and outputs, file bodies, and transcript paths are discarded before an event is persisted or rendered.
- Collection is disabled by default and stops immediately when the local kill switch is active.
- Only canonical project paths named in the local allowlist are eligible.

## Architecture

`ProviderWatcher` emits a shared `NormalizedEvent`. The existing Claude watcher continues to emit its existing `Event` values unchanged. `CodexAdapter` converts an already-sanitized local hook spool record into `NormalizedEvent`; `CodexSessionWatcher` is a read-only JSONL fallback that accepts only pre-sanitized event envelopes.

The runtime maps normalized Codex events to the existing animation `Event` type. Provider-tagged session metrics track XP separately for Claude and Codex while the existing visual behavior remains compatible.

## Local data flow

1. A project-local Codex Hook calls the installed `cq codex-hook` subcommand with one JSON object on standard input.
2. The subcommand checks the feature flag, kill switch, and canonical-path allowlist before parsing any event metadata.
3. It emits only a sanitized event: provider, event type, timestamp, status, duration, file-change count, test result, and git milestone.
4. The event is written to a local append-only spool with owner-only permissions and bounded rotation (14 days or 25 MiB total).
5. `cq codex-watch <project>` tails the spool and converts the records through `CodexAdapter` to existing game events.
6. If hooks are unavailable, `cq codex-watch <project> --fallback-sessions` tails only configured Codex session JSONL files read-only and accepts only sanitized envelopes; arbitrary transcript parsing is intentionally excluded.

## Configuration

The non-secret configuration file is `~/.claude-quest/codex-provider.json` and uses mode `0600` when created. It contains an explicit allowlist, `enabled`, `kill_switch`, retention days, and maximum bytes. The default configuration has `enabled: false`, `kill_switch: false`, no allowlisted projects, 14 days retention, and a 25 MiB cap.

## XP policy

XP is calculated from observable completion outcomes, not token totals: file change +3 each, completed test +5, passed test +5, commit +8, PR +12, merge +20, and failed event -2. It is reported by provider and as a combined total.

## Verification

Unit tests cover default-deny configuration, allowlist matching, sanitization, retention rotation, XP calculation, adapter mapping, and rejection of raw/untrusted fallback input. A smoke test creates a temporary allowlisted project, sends a sanitized hook record through the CLI, and confirms a local spool record contains no forbidden raw fields.
