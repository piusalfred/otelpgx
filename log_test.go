package otelpgx

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/tracelog"
)

func TestLogger_determineLogLevel(t *testing.T) {
	levels := []slog.Level{
		LevelTrace,
		slog.LevelInfo,
		slog.LevelError,
		LevelNone,
		slog.LevelWarn,
		slog.LevelDebug,
	}

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     LevelTrace,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelLabel, exists := LevelNames[level]
				if !exists {
					levelLabel = level.String()
				}

				a.Value = slog.StringValue(levelLabel)
			}
			return a
		},
	}))

	logger := Logger{
		logger:    l,
		converter: defaultLogLevelConverter{},
	}

	response := logger.determineLogLevel(context.Background(), levels)

	if response != LevelTrace {
		t.Errorf("Expected %v, got %v", slog.LevelInfo, response)
	}
}

func TestNewTraceLogger(t *testing.T) {
	logger := slog.New(
		slog.NewTextHandler(
			os.Stderr,
			&slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelError,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.LevelKey {
						level := a.Value.Any().(slog.Level)
						levelLabel, exists := LevelNames[level]
						if !exists {
							levelLabel = level.String()
						}

						a.Value = slog.StringValue(levelLabel)
					}
					return a
				},
			},
		),
	)

	tests := []struct {
		name string
		opts []LoggerOption
		want tracelog.LogLevel
	}{
		{
			name: "no options",
			want: tracelog.LogLevelNone,
		},
		{
			name: "with logger options only",
			opts: []LoggerOption{
				WithLogger(logger),
			},
			want: tracelog.LogLevelError,
		},
		{
			name: "with log level set",
			opts: []LoggerOption{
				WithLogLevel(LevelTrace),
				WithLogger(logger),
			},
			want: tracelog.LogLevelTrace,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := NewTraceLogger(tt.opts...)
			if tl.LogLevel != tt.want {
				t.Errorf("NewTraceLogger() = %v, want %v", tl.LogLevel, tt.want)
			}
		})
	}
}
