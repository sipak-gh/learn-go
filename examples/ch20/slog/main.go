package main

import (
	"log/slog"
	"os"
	"time"
)

func main() {
	// Remove the time so the output is the same on every run.
	noTime := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey && len(groups) == 0 {
			return slog.Attr{}
		}
		return a
	}

	text := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: noTime}))
	text.Info("file opened", "path", "notes.md", "size", 1024)
	text.Debug("not shown: below the default level, Info")
	text.Warn("slow file", "path", "big log.txt", "took", 1500*time.Millisecond)
	text.Error("open failed", "err", os.ErrNotExist)

	// With returns a logger that adds attributes to every record.
	reqLog := text.With("request", 42)
	reqLog.Info("started")

	// Groups nest attributes under a name.
	text.Info("done", slog.Group("stats", "files", 3, "matches", 7))

	// The JSON handler writes one JSON object per record.
	json := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		ReplaceAttr: noTime,
	}))
	json.Debug("hello", "user", "ana", slog.Int("tries", 3))
	json.Info("done", slog.Group("stats", "files", 3, "matches", 7))
}
