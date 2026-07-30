#!/usr/bin/env bash
set -euo pipefail

go_bin=${GO_BIN:-go}
temp_root=$(mktemp -d)
project="$temp_root/allowed"
config_dir="$temp_root/home/.claude-quest"
mkdir -p "$project" "$config_dir"

cat > "$config_dir/codex-provider.json" <<JSON
{
  "enabled": true,
  "kill_switch": false,
  "allowed_projects": ["$project"],
  "retention_days": 14,
  "max_bytes": 26214400
}
JSON

"$go_bin" build -o "$temp_root/cq" .
printf '{"session_id":"private-session","cwd":"%s","hook_event_name":"PostToolUse","tool_name":"apply_patch","status":"success"}\n' "$project" | HOME="$temp_root/home" "$temp_root/cq" codex-hook

event_file="$config_dir/codex-provider/events.jsonl"
test -f "$event_file"
test "$(stat -f '%Lp' "$event_file")" = "600"
rg -q '"provider":"codex"' "$event_file"
if rg -q 'private-session|tool_name|command|prompt' "$event_file"; then
  echo "smoke test failed: raw data was retained" >&2
  exit 1
fi

echo "Codex provider smoke test passed. Temporary evidence: $temp_root"
