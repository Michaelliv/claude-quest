package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncodeProjectPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "simple windows path",
			in:   `C:\Users\alice`,
			want: "C--Users-alice",
		},
		{
			name: "underscore in path segment",
			in:   `c:\Users\alice\dev\my_app\sub-dir`,
			want: "c--Users-alice-dev-my-app-sub-dir",
		},
		{
			name: "plus sign in path",
			in:   `C:\src\a+b`,
			want: "C--src-a-b",
		},
		{
			name: "multiple underscores and plus together",
			in:   `C:\src\foo_bar_baz+qux`,
			want: "C--src-foo-bar-baz-qux",
		},
		{
			name: "dot in path",
			in:   `/Users/foo/my.project`,
			want: "-Users-foo-my-project",
		},
		{
			name: "space in path",
			in:   `/Users/foo/my project`,
			want: "-Users-foo-my-project",
		},
		{
			name: "posix path",
			in:   `/home/user/repo`,
			want: "-home-user-repo",
		},
		{
			// Non-ASCII (e.g. CJK) codepoints are dashed individually, matching
			// Claude Code's `/[^a-zA-Z0-9]/g` regex over UTF-16 code units.
			name: "non-ascii characters become dashes",
			in:   `/home/user/文档/proj`,
			want: "-home-user----proj",
		},
		{
			// Misc punctuation — parentheses, brackets, ampersand, etc. all
			// collapse to dashes just like the JS regex.
			name: "punctuation collapses to dashes",
			in:   `/tmp/foo (bar) [baz] & qux`,
			want: "-tmp-foo--bar---baz----qux",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := encodeProjectPath(c.in)
			if got != c.want {
				t.Errorf("encodeProjectPath(%q) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

func TestEncodeProjectPathLongPathHashSuffix(t *testing.T) {
	// Mirror Claude Code's pM truncation: when the sanitized name exceeds 200
	// chars, the result is the first 200 chars + "-" + base36(abs(hash))
	// where hash is djb2-style over UTF-16 code units, 32-bit signed wraparound.
	long := "/" + strings.Repeat("a", 250)
	got := encodeProjectPath(long)

	// Sanitized form: leading '-' then 250 'a's -> 251 chars. Should truncate
	// to first 200, then '-', then a non-empty base36 hash.
	if !strings.HasPrefix(got, "-"+strings.Repeat("a", 199)) {
		t.Fatalf("unexpected prefix: %q", got[:min(len(got), 30)])
	}
	rest := got[200:]
	if !strings.HasPrefix(rest, "-") {
		t.Fatalf("expected '-' separator after first 200 chars, got %q", rest[:min(len(rest), 10)])
	}
	suffix := rest[1:]
	if suffix == "" {
		t.Fatal("expected non-empty hash suffix")
	}
	for _, r := range suffix {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'z')) {
			t.Fatalf("hash suffix should be base36 alphanumeric, got %q", suffix)
		}
	}

	// Distinct long inputs must produce distinct full keys (the whole point
	// of the hash suffix — protect against collisions on truncation).
	other := encodeProjectPath("/" + strings.Repeat("a", 249) + "b")
	if other == got {
		t.Fatalf("expected different keys for different long inputs, both = %q", got)
	}
}

func TestResolveProjectDir(t *testing.T) {
	root := t.TempDir()
	// Folder created with a lowercase drive letter, mimicking what Claude
	// Code stores when started from a session whose cwd reported the drive
	// in lowercase (e.g. some VSCode launches on Windows).
	actual := "c--Users-alice-dev-my-app"
	if err := os.MkdirAll(filepath.Join(root, actual), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Run("exact match returns exact path", func(t *testing.T) {
		got := resolveProjectDir(root, actual)
		want := filepath.Join(root, actual)
		if got != want {
			t.Errorf("got %q; want %q", got, want)
		}
	})

	t.Run("case-mismatched lookup resolves to an existing folder", func(t *testing.T) {
		// What filepath.Abs may produce in PowerShell (uppercase drive) when
		// the folder on disk was created with a lowercase drive letter.
		requested := "C--Users-alice-dev-my-app"
		got := resolveProjectDir(root, requested)
		// On case-insensitive filesystems (Windows NTFS, macOS APFS default)
		// the exact path stats successfully and is returned as-is. On
		// case-sensitive filesystems (Linux ext4) the fallback returns the
		// on-disk casing. Either is correct — both must open the same folder.
		if _, err := os.Stat(got); err != nil {
			t.Errorf("resolveProjectDir returned %q but it is not statable: %v", got, err)
		}
		// Also ensure that whichever casing was returned points to the same
		// directory we created.
		gotAbs, _ := filepath.EvalSymlinks(got)
		wantAbs, _ := filepath.EvalSymlinks(filepath.Join(root, actual))
		// On Windows EvalSymlinks normalises the case to the on-disk form.
		// On case-sensitive FS the strings already match (fallback path).
		if gotAbs == "" || wantAbs == "" || gotAbs != wantAbs {
			t.Errorf("returned path %q does not resolve to created folder %q", gotAbs, wantAbs)
		}
	})

	t.Run("no match returns exact (non-existent) path", func(t *testing.T) {
		requested := "C--does-not-exist"
		got := resolveProjectDir(root, requested)
		want := filepath.Join(root, requested)
		if got != want {
			t.Errorf("got %q; want %q (should fall through to non-existent path)", got, want)
		}
	})

	t.Run("missing projects root returns exact path", func(t *testing.T) {
		missing := filepath.Join(root, "does-not-exist")
		requested := "C--anything"
		got := resolveProjectDir(missing, requested)
		want := filepath.Join(missing, requested)
		if got != want {
			t.Errorf("got %q; want %q", got, want)
		}
	})
}
