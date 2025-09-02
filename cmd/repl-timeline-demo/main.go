package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/eventbus"
	"github.com/go-go-golems/bobatea/pkg/repl"
	"github.com/go-go-golems/bobatea/pkg/timeline"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

// DemoEvaluator: simple + shell
// - If input starts with !, run shell after the ! and stream stdout/stderr
// - Else, emit a markdown result echoing the input
type DemoEvaluator struct{ CoalesceMs int }

func (d DemoEvaluator) EvaluateStream(ctx context.Context, code string, emit func(repl.Event)) error {
	code = strings.TrimSpace(code)
	if strings.HasPrefix(code, "!") {
		cmdStr := strings.TrimSpace(strings.TrimPrefix(code, "!"))
		if cmdStr == "" {
			return nil
		}
		zlog.Debug().Str("cmd", cmdStr).Msg("starting shell command")
		cmd := exec.CommandContext(ctx, "bash", "-c", cmdStr)
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		start := time.Now()
		if err := cmd.Start(); err != nil {
			emit(repl.Event{Kind: repl.EventStderr, Props: map[string]any{"append": err.Error() + "\n", "is_error": true}})
			return err
		}

		// Coalescers
		interval := d.CoalesceMs
		if interval <= 0 {
			interval = 60
		}
		coalesce := func(ch <-chan string, kind repl.EventKind, extra map[string]any) {
			ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
			defer ticker.Stop()
			buf := strings.Builder{}
			flush := func() {
				if buf.Len() == 0 {
					return
				}
				props := map[string]any{"append": buf.String()}
				for k, v := range extra {
					props[k] = v
				}
				emit(repl.Event{Kind: kind, Props: props})
				buf.Reset()
			}
			for {
				select {
				case line, ok := <-ch:
					if !ok {
						flush()
						return
					}
					buf.WriteString(line)
					buf.WriteByte('\n')
				case <-ticker.C:
					flush()
				}
			}
		}

		// Stream lines with coalescing to reduce UI churn
		outCh := make(chan string, 1024)
		errCh := make(chan string, 1024)
		var wg sync.WaitGroup
		wg.Add(4)
		zlog.Debug().Msg("streaming stdout and stderr")
		go func() { defer wg.Done(); streamLines(stdout, func(line string) { outCh <- line }) }()
		go func() { defer wg.Done(); streamLines(stderr, func(line string) { errCh <- line }) }()
		go func() { defer wg.Done(); coalesce(outCh, repl.EventStdout, map[string]any{}) }()
		go func() { defer wg.Done(); coalesce(errCh, repl.EventStderr, map[string]any{"is_error": true}) }()
		zlog.Debug().Msg("streaming stdout and stderr done")

		// Start coalescers and wait for process end
		zlog.Debug().Msg("waiting for command to finish")
		err := cmd.Wait()
		close(outCh)
		close(errCh)
		wg.Wait()
		dur := time.Since(start)
		exit := 0
		if err != nil {
			exit = 1
		}
		md := fmt.Sprintf("Command: `%s`\n\nExit: %d\nDuration: %s", cmdStr, exit, dur)
		emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": md}})
		zlog.Debug().Str("cmd", cmdStr).Int("exit", exit).Dur("duration", dur).Msg("command finished")

		return err
	}

	// Simple markdown echo
	md := fmt.Sprintf("You said:\n\n```\n%s\n```", code)
	emit(repl.Event{Kind: repl.EventResultMarkdown, Props: map[string]any{"markdown": md}})
	return nil
}

func (d DemoEvaluator) GetPrompt() string        { return "> " }
func (d DemoEvaluator) GetName() string          { return "Demo" }
func (d DemoEvaluator) SupportsMultiline() bool  { return true }
func (d DemoEvaluator) GetFileExtension() string { return ".txt" }

func streamLines(r any, onLine func(string)) {
	// Minimal scanner using bufio
	if rc, ok := r.(interface{ Read([]byte) (int, error) }); ok {
		buf := make([]byte, 8192)
		var acc strings.Builder
		for {
			n, err := rc.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				acc.WriteString(chunk)
				for {
					s := acc.String()
					idx := strings.IndexByte(s, '\n')
					if idx < 0 {
						break
					}
					line := s[:idx]
					onLine(line)
					acc.Reset()
					acc.WriteString(s[idx+1:])
				}
			}
			if err != nil {
				rem := acc.String()
				if rem != "" {
					onLine(rem)
				}
				return
			}
		}
	}
}

func main() {
	// logging: default to info, allow override via --log-level
	var lvlFlag string
	var logFile string
	var coalesceMs int
	flag.StringVar(&lvlFlag, "log-level", "error", "log level (trace, debug, info, warn, error)")
	flag.StringVar(&logFile, "log-file", "", "path to write logs (JSON)")
	flag.IntVar(&coalesceMs, "coalesce-ms", 60, "stdout/stderr coalescing interval (ms)")
	flag.Parse()

	if parsed, err := zerolog.ParseLevel(lvlFlag); err == nil {
		zerolog.SetGlobalLevel(parsed)
	} else {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644) // #nosec G302
		if err == nil {
			// setup zerolog console writer for human readable logs
			consoleWriter := zerolog.ConsoleWriter{
				Out:        f,
				TimeFormat: "15:04:05",
				NoColor:    true,
			}
			zlog.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
		}
	} else {
		// Avoid any console logging interfering with TUI
		zlog.Logger = zerolog.New(io.Discard)
	}

	// Add caller info
	zlog.Logger = zlog.Logger.With().Caller().Logger()

	evaluator := DemoEvaluator{CoalesceMs: coalesceMs}
	cfg := repl.Config{
		Title:                "Timeline REPL Demo",
		Placeholder:          "Type text, or !<shell command> (e.g., !echo hi)",
		Width:                100,
		StartMultiline:       false,
		EnableExternalEditor: false,
		EnableHistory:        true,
		MaxHistorySize:       1000,
	}
	// Build in-memory bus
	bus, err := eventbus.NewInMemoryBus()
	if err != nil {
		log.Fatal(err)
	}
	// Register transformer and UI forwarder
	repl.RegisterReplToTimelineTransformer(bus)

	// Construct model with publisher
	m := repl.NewModel(evaluator, cfg, bus.Publisher)
	p := tea.NewProgram(m, tea.WithAltScreen())
	timeline.RegisterUIForwarder(bus, p)

	// Run router + program together
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errs := make(chan error, 2)
	go func() { errs <- bus.Run(ctx) }()
	go func() {
		_, e := p.Run()
		cancel()
		errs <- e
	}()
	if e := <-errs; e != nil {
		log.Fatal(e)
	}
}
