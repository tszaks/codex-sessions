# codex-sessions

Small CLI for seeing which Codex sessions are live on your machine.

It reads local Codex state from `~/.codex/state_5.sqlite`, joins that with live `codex` processes from `ps`, and shows:

- which sessions are active
- what they are about
- which project they are effectively working in
- a short recent action preview

## Install

```bash
git clone https://github.com/tszaks/codex-sessions.git
cd codex-sessions
go build
```

Or run it directly:

```bash
go run . --details
```

## Usage

List active sessions:

```bash
codex-sessions
```

Show details:

```bash
codex-sessions --details
```

Get JSON output:

```bash
codex-sessions --json
```

Include inactive recent threads too:

```bash
codex-sessions --all
```

Live refresh:

```bash
codex-sessions --watch
```

Jump back into a session:

```bash
codex-sessions resume 019cc45a
```

## Notes

- macOS-first for now
- read-only by default
- `resume` shells out to `codex resume <thread-id>`
- if `~/.codex/state_5.sqlite` is missing, it still shows live starting-up sessions when possible
