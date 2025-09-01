package logutil

import (
    "io"
    "os"

    "github.com/rs/zerolog"
    zlog "github.com/rs/zerolog/log"
)

// InitTUILoggingToDiscard configures zerolog to a given level and discards all output,
// avoiding interference with TUI stdout.
func InitTUILoggingToDiscard(level zerolog.Level) {
    zerolog.SetGlobalLevel(level)
    zlog.Logger = zerolog.New(io.Discard)
}

// InitTUILoggingToFile writes logs to a file using a console writer without color.
// If file open fails, it falls back to discarding output.
func InitTUILoggingToFile(level zerolog.Level, path string) {
    zerolog.SetGlobalLevel(level)
    f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644) // #nosec G302
    if err != nil {
        zlog.Logger = zerolog.New(io.Discard)
        return
    }
    cw := zerolog.ConsoleWriter{Out: f, NoColor: true}
    zlog.Logger = zerolog.New(cw).With().Timestamp().Logger()
}

