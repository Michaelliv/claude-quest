package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeJSONL creates a JSONL file with the given lines and forces its mtime
// to the supplied value via os.Chtimes (newer than the previous mtime by the
// caller's choice).
func writeJSONL(t *testing.T, dir, name string, lines []string, mtime time.Time) string {
	t.Helper()
	p := filepath.Join(dir, name)
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	if err := os.Chtimes(p, mtime, mtime); err != nil {
		t.Fatalf("chtimes %s: %v", p, err)
	}
	return p
}

func newWatcherFor(dir string) *Watcher {
	w := NewWatcher()
	w.ProjectDir = dir
	return w
}

// uuidName returns a 36-character UUID-shaped filename stem with .jsonl suffix.
func uuidName(n int) string {
	return fmt.Sprintf("%08d-1111-2222-3333-%012d.jsonl", n, n)
}

func TestCheckForNewerFile_ActiveFileNotSwitchedAway(t *testing.T) {
	dir := t.TempDir()

	now := time.Now()
	active := writeJSONL(t, dir, uuidName(1), []string{`{"uuid":"u-active-1","type":"user"}`}, now)
	// A second file that's slightly OLDER than the active one — must not
	// "win" the comparison even though both exist.
	_ = writeJSONL(t, dir, uuidName(2), []string{`{"uuid":"u-other-1","type":"user"}`}, now.Add(-2*time.Second))

	w := newWatcherFor(dir)
	w.FilePath = active
	w.lastModTime = now
	if w.checkForNewerFile() {
		t.Fatalf("checkForNewerFile switched even though active file is newest")
	}
	if filepath.Base(w.FilePath) != filepath.Base(active) {
		t.Fatalf("FilePath changed unexpectedly: %s", w.FilePath)
	}
}

func TestCheckForNewerFile_StaleBaselineDoesNotTriggerSwitch(t *testing.T) {
	// Regression: previously, lastModTime froze at startup and any other
	// file whose mtime exceeded that frozen value would cause a switch —
	// even if the active file was *currently* even newer.
	dir := t.TempDir()

	now := time.Now()
	active := writeJSONL(t, dir, uuidName(1), []string{`{"uuid":"u-active-1"}`}, now)
	_ = writeJSONL(t, dir, uuidName(2), []string{`{"uuid":"u-other-1"}`}, now.Add(-1*time.Second))

	w := newWatcherFor(dir)
	w.FilePath = active
	// Pretend baseline is way in the past (simulating tailFile having not
	// refreshed it yet). The fix uses a fresh stat of FilePath instead.
	w.lastModTime = now.Add(-1 * time.Hour)
	if w.checkForNewerFile() {
		t.Fatalf("switched on stale baseline; fix should compare against fresh stat")
	}
}

func TestCheckForNewerFile_TouchedHistoricalFileIgnored(t *testing.T) {
	// Claude Code occasionally touches old session files. Even if such a
	// file becomes "newest" by mtime, we should not switch to it because
	// it isn't a live session — its mtime is older than activeSessionWindow.
	dir := t.TempDir()

	stale := time.Now().Add(-(activeSessionWindow + time.Minute))
	active := writeJSONL(t, dir, uuidName(1), []string{`{"uuid":"u-1"}`}, stale.Add(-time.Hour))
	// "Touched" historical file: UUID-shaped name, newer than active, but
	// mtime itself is well outside the active-session window.
	_ = writeJSONL(t, dir, uuidName(2), []string{`{"uuid":"u-2"}`}, stale)

	w := newWatcherFor(dir)
	w.FilePath = active
	w.lastModTime = stale.Add(-time.Hour)
	if w.checkForNewerFile() {
		t.Fatalf("switched to a historical file whose mtime was outside the active-session window")
	}
}

func TestCheckForNewerFile_NonUUIDFilenameIgnored(t *testing.T) {
	dir := t.TempDir()

	now := time.Now()
	active := writeJSONL(t, dir, uuidName(1), []string{`{"uuid":"u-1"}`}, now.Add(-1*time.Second))
	// A non-UUID-shaped filename — must not be considered a live session.
	_ = writeJSONL(t, dir, "scratch.jsonl", []string{`{"uuid":"u-2"}`}, now)

	w := newWatcherFor(dir)
	w.FilePath = active
	w.lastModTime = now.Add(-1 * time.Second)
	if w.checkForNewerFile() {
		t.Fatalf("switched to a non-UUID-shaped file")
	}
}

func TestUUIDSet_AddIsDedup(t *testing.T) {
	s := newUUIDSet(16)
	if !s.add("a") {
		t.Fatalf("first add of 'a' should return true (newly added)")
	}
	if s.add("a") {
		t.Fatalf("second add of 'a' should return false (already present)")
	}
	if !s.add("b") {
		t.Fatalf("first add of 'b' should return true")
	}
}

func TestUUIDSet_EvictsOldestAtCapacity(t *testing.T) {
	const cap = 4
	s := newUUIDSet(cap)
	for i := 0; i < cap; i++ {
		if !s.add(fmt.Sprintf("u%d", i)) {
			t.Fatalf("fresh add returned false at i=%d", i)
		}
	}
	// All cap distinct keys present — re-adding "u0" should still be a dup.
	if s.add("u0") {
		t.Fatalf("u0 should still be present before any evictions")
	}
	// Adding a new key forces eviction of the oldest ("u0").
	if !s.add("u4") {
		t.Fatalf("u4 should be a fresh add")
	}
	// Now "u0" must be evictable — re-adding returns true.
	if !s.add("u0") {
		t.Fatalf("u0 should have been evicted and now count as newly added")
	}
}

func TestUUIDSet_ResetClearsAll(t *testing.T) {
	s := newUUIDSet(8)
	for i := 0; i < 5; i++ {
		s.add(fmt.Sprintf("u%d", i))
	}
	s.reset()
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("u%d", i)
		if !s.add(key) {
			t.Fatalf("after reset, %q should be a fresh add", key)
		}
	}
}

func TestParseLine_DedupesByUUID(t *testing.T) {
	w := NewWatcher()
	line := `{"type":"assistant","uuid":"abc-123","message":{"role":"assistant","content":[{"type":"text","text":"hello"}]}}`

	first := w.parseLine(line)
	if len(first) == 0 {
		t.Fatalf("first parseLine should emit at least one event")
	}
	second := w.parseLine(line)
	if len(second) != 0 {
		t.Fatalf("second parseLine should be deduped to zero events, got %d", len(second))
	}
}

func TestParseLine_CompactResetsDedupSet(t *testing.T) {
	// After a compact boundary, pre-compact UUIDs are no longer referenced
	// by anything we'll process. The dedup set should reset so memory is
	// freed early, while still preserving the compact line's own UUID so
	// re-reading that exact line cannot re-emit the compact event.
	w := NewWatcher()

	pre := `{"type":"assistant","uuid":"pre-1","message":{"role":"assistant","content":[{"type":"text","text":"hi"}]}}`
	compact := `{"type":"system","subtype":"compact_boundary","uuid":"compact-1","compactMetadata":{"trigger":"auto","preTokens":120000}}`

	if got := w.parseLine(pre); len(got) == 0 {
		t.Fatalf("expected pre-compact event")
	}
	if got := w.parseLine(compact); len(got) == 0 || got[0].Type != EventCompact {
		t.Fatalf("expected compact event")
	}

	// Pre-compact UUID should now count as fresh again — reset cleared it.
	if got := w.parseLine(pre); len(got) == 0 {
		t.Fatalf("pre-compact UUID should be acceptable after compact reset, got nothing")
	}

	// But the compact line itself was re-added after reset, so re-reading
	// the compact line is still deduped.
	if got := w.parseLine(compact); len(got) != 0 {
		t.Fatalf("compact line should remain deduped after reset, got %d events", len(got))
	}
}

func TestCheckForNewerFile_RealSessionSwitchSeeksToEOF(t *testing.T) {
	dir := t.TempDir()

	now := time.Now()
	active := writeJSONL(t, dir, uuidName(1), []string{`{"uuid":"u-old-1"}`}, now.Add(-5*time.Second))

	// New live session: UUID-shaped, mtime is now, contains ten lines of
	// historical content that MUST NOT be replayed after the switch.
	historicalLines := make([]string, 10)
	for i := range historicalLines {
		historicalLines[i] = fmt.Sprintf(`{"uuid":"u-new-%d","type":"user"}`, i)
	}
	newSession := writeJSONL(t, dir, uuidName(2), historicalLines, now)

	w := newWatcherFor(dir)
	w.FilePath = active
	w.lastModTime = now.Add(-5 * time.Second)
	w.lastPos = 999 // a meaningless previous position

	if !w.checkForNewerFile() {
		t.Fatalf("expected switch to genuinely newer session file")
	}
	if w.FilePath != newSession {
		t.Fatalf("FilePath = %s; want %s", w.FilePath, newSession)
	}
	info, err := os.Stat(newSession)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if w.lastPos != info.Size() {
		t.Fatalf("lastPos = %d; want EOF (%d) so historical lines are skipped",
			w.lastPos, info.Size())
	}
}
