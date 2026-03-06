package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseElapsedTime(t *testing.T) {
	tests := []struct {
		raw  string
		want int64
	}{
		{"45", 45},
		{"05:24", 324},
		{"10:49:45", 38985},
		{"2-03:04:05", 183845},
	}

	for _, tt := range tests {
		got, err := parseElapsedTime(tt.raw)
		if err != nil {
			t.Fatalf("parseElapsedTime(%q) failed: %v", tt.raw, err)
		}
		if got != tt.want {
			t.Fatalf("parseElapsedTime(%q) = %d, want %d", tt.raw, got, tt.want)
		}
	}
}

func TestExtractWorkdirFromCommand(t *testing.T) {
	call := toolCall{
		Name: "exec_command",
		Args: map[string]any{
			"cmd": `cd "/Users/tyler/Projects/cli" && git status`,
		},
	}
	if got := extractWorkdir(call); got != "/Users/tyler/Projects/cli" {
		t.Fatalf("extractWorkdir() = %q", got)
	}
}

func TestResolveSessionNoMatch(t *testing.T) {
	_, err := resolveSession(context.Background(), "does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing session")
	}
}

func TestCollectSessionsWithFixture(t *testing.T) {
	if _, err := exec.LookPath(sqlite3Command); err != nil {
		t.Skipf("sqlite3 not available: %v", err)
	}

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, codexStateDBFile)
	createTestDB(t, dbPath, `
CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	cwd TEXT NOT NULL DEFAULT '',
	title TEXT NOT NULL DEFAULT '',
	first_user_message TEXT NOT NULL DEFAULT '',
	git_branch TEXT,
	git_origin_url TEXT,
	archived INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	ts INTEGER NOT NULL,
	target TEXT NOT NULL DEFAULT '',
	message TEXT,
	thread_id TEXT,
	process_uuid TEXT
);
INSERT INTO threads (id, created_at, updated_at, cwd, title, first_user_message, git_branch, git_origin_url, archived) VALUES
	('thread-active', 1772821500, 1772821950, '/Users/tyler', 'Active session title', 'active prompt', 'main', 'git@example.com:cli.git', 0),
	('thread-inactive', 1772821400, 1772821800, '/Users/tyler', 'Inactive session title', 'inactive prompt', 'feat/archive', 'git@example.com:archive.git', 0);
INSERT INTO logs (ts, target, message, thread_id, process_uuid) VALUES
	(1772821950, 'codex_core::stream_events_utils', 'ToolCall: exec_command {"cmd":"git status","workdir":"/Users/tyler/Projects/cli"}', 'thread-active', 'pid:4242:test-process'),
	(1772821800, 'codex_core::stream_events_utils', 'ToolCall: exec_command {"cmd":"cd /Users/tyler/Projects/archive && git status"}', 'thread-inactive', 'pid:9999:old-process');
`)

	originalHome := codexHomeDirFunc
	originalNow := nowFunc
	originalHostname := hostnameFunc
	originalList := listLiveCodexProcessesVar
	defer func() {
		codexHomeDirFunc = originalHome
		nowFunc = originalNow
		hostnameFunc = originalHostname
		listLiveCodexProcessesVar = originalList
	}()

	codexHomeDirFunc = func() (string, error) { return tempDir, nil }
	nowFunc = func() time.Time { return time.Unix(1772822000, 0).UTC() }
	hostnameFunc = func() (string, error) { return "test-host", nil }
	listLiveCodexProcessesVar = func(context.Context) ([]liveCodexProcess, error) {
		return []liveCodexProcess{{PID: 4242, TTY: "ttys001", AgeSeconds: 125}}, nil
	}

	snapshot, err := CollectSessions(context.Background(), SessionCollectOptions{IncludeAll: true, IncludeDetails: true})
	if err != nil {
		t.Fatalf("CollectSessions failed: %v", err)
	}
	if len(snapshot.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(snapshot.Sessions))
	}
	if snapshot.Sessions[0].ThreadID != "thread-active" {
		t.Fatalf("unexpected first session: %+v", snapshot.Sessions[0])
	}
	if !strings.Contains(snapshot.Sessions[0].EffectiveWorkdir, "/Users/tyler/Projects/cli") {
		t.Fatalf("unexpected workdir: %q", snapshot.Sessions[0].EffectiveWorkdir)
	}
}

func createTestDB(t *testing.T, dbPath, sql string) {
	t.Helper()
	cmd := exec.Command(sqlite3Command, dbPath, sql)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to create sqlite test db: %v\n%s", err, output)
	}
}
