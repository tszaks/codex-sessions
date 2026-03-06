# codex-sessions

`codex-sessions` is a small local CLI for answering one practical question:

Which Codex sessions are running on this machine right now?

It reads local Codex state from `~/.codex/state_5.sqlite`, joins that with live `codex` processes from `ps`, and shows:

- which sessions are active
- what each session is about
- which project or working directory it is actually touching
- a short recent action summary

It also includes a lightweight jump helper:

```bash
codex-sessions resume <thread-id-or-prefix>
```

## Why this exists

If you run multiple Codex sessions at once, it gets hard to answer simple questions:

- Which terminal is working on what?
- Which repo is each session actually in?
- Is this session still active or just sitting there?
- What was the last thing it did?

`codex-sessions` gives you a fast local view without adding a daemon, dashboard, or separate backend.

## How it works

The tool combines two local data sources:

1. `ps`
   Used to find live `codex` processes, terminal TTYs, and process age.
2. `~/.codex/state_5.sqlite`
   Used to map a live process back to the correct Codex thread, title, cwd, and recent tool activity.

To avoid common false matches, session resolution uses:

- the live process PID
- the process start window based on elapsed process age
- the root thread created for that terminal session, not a spawned subagent thread

## Install

### Homebrew

```bash
brew install tszaks/tap/codex-sessions
```

### Build from source

```bash
git clone https://github.com/tszaks/codex-sessions.git
cd codex-sessions
go build
```

That creates a local binary named `codex-sessions`.

### Run without building

```bash
go run . --details
```

## Commands

### Show active sessions

```bash
codex-sessions
```

### Show recent action details

```bash
codex-sessions --details
```

### Include inactive threads too

```bash
codex-sessions --all
```

### Output JSON

```bash
codex-sessions --json
```

### Refresh continuously

```bash
codex-sessions --watch
```

### Resume a session

```bash
codex-sessions resume 019cc45a
```

The `resume` command delegates to:

```bash
codex resume <full-thread-id>
```

## Example table output

```text
5 active Codex sessions
updated 2:56PM

PID    TTY      AGE     THREAD    PROJECT                   TOPIC                             LAST ACTIVE
82996  ttys006  1h43m   019cc45a  ~/Projects/codex-sessions build a tool to see other...     just now
34607  ttys010  13h44m  019cc1c6  ~/Projects/AskVero/ios    review ask-ai realtime...        3m ago
```

## Example JSON output

```json
{
  "generated_at": "2026-03-06T19:55:55Z",
  "host": "Mac",
  "sessions": [
    {
      "pid": 82996,
      "tty": "ttys006",
      "age_seconds": 6134,
      "thread_id": "019cc45a-f772-7fd0-b7c0-f762df930f57",
      "title": "build a tool to see other codex sessions",
      "session_cwd": "/Users/tyler",
      "effective_workdir": "/Users/tyler/Projects/codex-sessions",
      "last_active_at": "2026-03-06T19:55:54Z",
      "status": "active",
      "recent_action": "exec_command: go run . --details"
    }
  ]
}
```

## Output fields

Each session object can include:

- `pid`
- `tty`
- `age_seconds`
- `thread_id`
- `title`
- `first_user_message`
- `session_cwd`
- `effective_workdir`
- `last_active_at`
- `git_branch`
- `git_origin_url`
- `status`
- `recent_action`

## Design goals

- local-first
- read-only by default
- fast enough to run often
- simple enough to inspect and extend
- structured output that is useful for both humans and LLMs

## Current scope

- macOS-first
- expects local Codex state under `~/.codex`
- uses `sqlite3` and `ps`
- no background service
- no session mutation except `resume`

## Non-goals

- remote coordination between machines
- cross-user session sharing
- a full interactive dashboard
- parsing or exposing shell environment snapshots

## Notes for LLM and tool use

If you are calling this tool from another script, agent, or LLM wrapper:

- use `--json`
- treat `thread_id` as the stable session identifier
- prefer `effective_workdir` over `session_cwd` when both are present
- expect `status` to be either `active` or `inactive`
- use `resume <thread-id-or-prefix>` only when you want to hand control back to Codex

## Development

Run tests:

```bash
go test ./...
```

Run the tool locally:

```bash
go run . --details
```

## License

MIT
