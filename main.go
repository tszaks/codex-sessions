package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"
)

const watchInterval = 2 * time.Second

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) > 0 && args[0] == "resume" {
		return runResume(args[1:])
	}
	return runSessions(args, stdout)
}

func runSessions(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("codex-sessions", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	includeAll := fs.Bool("all", false, "include inactive Codex threads too")
	jsonOutput := fs.Bool("json", false, "output JSON instead of a table")
	includeDetails := fs.Bool("details", false, "show recent action details")
	watch := fs.Bool("watch", false, "refresh every 2 seconds")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}
	if *watch && *jsonOutput {
		return errors.New("--watch cannot be used with --json")
	}

	opts := SessionCollectOptions{
		IncludeAll:     *includeAll,
		IncludeDetails: *includeDetails,
	}
	ctx := context.Background()
	if *watch {
		return watchSessions(ctx, stdout, opts)
	}
	return renderOnce(ctx, stdout, opts, *jsonOutput)
}

func runResume(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: codex-sessions resume <thread-id-or-prefix>")
	}

	session, err := resolveSession(context.Background(), args[0])
	if err != nil {
		return err
	}

	cmd := exec.Command("codex", "resume", session.ThreadID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func watchSessions(ctx context.Context, stdout io.Writer, opts SessionCollectOptions) error {
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	for {
		if file, ok := stdout.(*os.File); ok {
			fmt.Fprint(file, "\033[H\033[2J")
		}

		if err := renderOnce(ctx, stdout, opts, false); err != nil {
			return err
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return nil
		}
	}
}

func renderOnce(ctx context.Context, stdout io.Writer, opts SessionCollectOptions, jsonOutput bool) error {
	snapshot, err := CollectSessions(ctx, opts)
	if err != nil {
		return err
	}

	if jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(snapshot)
	}

	renderTable(stdout, snapshot, opts)
	return nil
}

func renderTable(stdout io.Writer, snapshot *SessionSnapshot, opts SessionCollectOptions) {
	activeCount, inactiveCount := countSessions(snapshot.Sessions)

	fmt.Fprintln(stdout)
	if opts.IncludeAll {
		fmt.Fprintf(stdout, "%d active, %d inactive Codex sessions\n", activeCount, inactiveCount)
	} else {
		fmt.Fprintf(stdout, "%d active Codex sessions\n", activeCount)
	}
	fmt.Fprintf(stdout, "updated %s\n\n", snapshot.GeneratedAt.Local().Format(time.Kitchen))

	if len(snapshot.Sessions) == 0 {
		if opts.IncludeAll {
			fmt.Fprintln(stdout, "No Codex sessions found.")
		} else {
			fmt.Fprintln(stdout, "No active Codex sessions found.")
		}
		fmt.Fprintln(stdout)
		return
	}

	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	if opts.IncludeAll {
		fmt.Fprintln(tw, "PID\tTTY\tAGE\tSTATUS\tTHREAD\tPROJECT\tTOPIC\tLAST ACTIVE")
	} else {
		fmt.Fprintln(tw, "PID\tTTY\tAGE\tTHREAD\tPROJECT\tTOPIC\tLAST ACTIVE")
	}

	for _, session := range snapshot.Sessions {
		pid := "-"
		if session.PID > 0 {
			pid = fmt.Sprintf("%d", session.PID)
		}

		project := truncate(displayPath(session.EffectiveWorkdir), 24)
		topic := truncate(strings.Join(strings.Fields(session.Title), " "), 32)
		lastActive := "-"
		if !session.LastActiveAt.IsZero() {
			lastActive = timeAgo(session.LastActiveAt.Local())
		}

		if opts.IncludeAll {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				pid,
				fallback(session.TTY, "-"),
				formatShortDuration(session.AgeSeconds),
				session.Status,
				shortThreadID(session.ThreadID),
				project,
				topic,
				lastActive,
			)
		} else {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				pid,
				fallback(session.TTY, "-"),
				formatShortDuration(session.AgeSeconds),
				shortThreadID(session.ThreadID),
				project,
				topic,
				lastActive,
			)
		}

		if opts.IncludeDetails && session.RecentAction != "" {
			fmt.Fprintf(tw, "\t\t\t\t\t\trecent: %s\t\n", truncate(session.RecentAction, 64))
		}
	}
	_ = tw.Flush()
	fmt.Fprintln(stdout)
}

func countSessions(sessions []SessionSummary) (active int, inactive int) {
	for _, session := range sessions {
		if session.Status == "active" {
			active++
		} else {
			inactive++
		}
	}
	return active, inactive
}

func resolveSession(ctx context.Context, query string) (SessionSummary, error) {
	snapshot, err := CollectSessions(ctx, SessionCollectOptions{IncludeAll: true})
	if err != nil {
		return SessionSummary{}, err
	}

	var matches []SessionSummary
	for _, session := range snapshot.Sessions {
		if session.ThreadID == query || strings.HasPrefix(session.ThreadID, query) {
			matches = append(matches, session)
		}
	}

	switch len(matches) {
	case 0:
		return SessionSummary{}, fmt.Errorf("no Codex session found matching %q", query)
	case 1:
		return matches[0], nil
	default:
		ids := make([]string, 0, len(matches))
		for _, match := range matches {
			ids = append(ids, shortThreadID(match.ThreadID))
		}
		return SessionSummary{}, fmt.Errorf("multiple Codex sessions match %q: %s", query, strings.Join(ids, ", "))
	}
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func truncate(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	if limit <= 3 {
		return string(runes[:limit])
	}
	return string(runes[:limit-3]) + "..."
}
