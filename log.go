package otelpgx

import (
	"context"
	"log/slog"
	"os"
	"slices"

	"github.com/jackc/pgx/v5/tracelog"
)

const (
	LevelTrace slog.Level = -8
	LevelNone  slog.Level = 12
)

var LevelNames = map[slog.Level]string{
	LevelTrace: "TRACE",
	LevelNone:  "NONE",
}

type (
	Logger struct {
		logger     *slog.Logger
		converter  LogLevelConverter
		level      slog.Level
		isLevelSet bool
	}

	LogLevelConverter interface {
		ToTraceLogLevel(leveler slog.Leveler) tracelog.LogLevel
		ToSlogLevel(level tracelog.LogLevel) slog.Leveler
	}

	defaultLogLevelConverter struct{}

	LoggerOption func(*Logger)
)

// WithLogLevelConverter sets the log level converter.
func WithLogLevelConverter(c LogLevelConverter) LoggerOption {
	return func(l *Logger) {
		l.converter = c

		// If the log level has not been set, set it to the default level.
		if !l.isLevelSet {
			l.level = l.determineLogLevel(context.Background(),
				[]slog.Level{
					LevelTrace,
					slog.LevelDebug,
					slog.LevelInfo,
					slog.LevelWarn,
					slog.LevelError,
					LevelNone,
				})
		}
	}
}

// WithLogLevel sets the log level.
func WithLogLevel(level slog.Level) LoggerOption {
	return func(l *Logger) {
		l.level = level
		l.isLevelSet = true
	}
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) LoggerOption {
	return func(l *Logger) {
		l.logger = logger
		// if log level is not set by WithLogLevel, determine it automatically
		if !l.isLevelSet {
			l.level = l.determineLogLevel(context.Background(),
				[]slog.Level{
					LevelTrace,
					slog.LevelDebug,
					slog.LevelInfo,
					slog.LevelWarn,
					slog.LevelError,
					LevelNone,
				})
		}
	}
}

func (c defaultLogLevelConverter) ToTraceLogLevel(leveler slog.Leveler) tracelog.LogLevel {
	switch leveler.Level() {
	case slog.LevelDebug:
		return tracelog.LogLevelDebug
	case slog.LevelInfo:
		return tracelog.LogLevelInfo
	case slog.LevelWarn:
		return tracelog.LogLevelWarn
	case slog.LevelError:
		return tracelog.LogLevelError
	case LevelTrace:
		return tracelog.LogLevelTrace
	case LevelNone:
		return tracelog.LogLevelNone

	default:
		return tracelog.LogLevelNone
	}
}

func (c defaultLogLevelConverter) ToSlogLevel(level tracelog.LogLevel) slog.Leveler {
	switch level {
	case tracelog.LogLevelDebug:
		return slog.LevelDebug
	case tracelog.LogLevelInfo:
		return slog.LevelInfo
	case tracelog.LogLevelWarn:
		return slog.LevelWarn
	case tracelog.LogLevelError:
		return slog.LevelError
	case tracelog.LogLevelTrace:
		return LevelTrace
	case tracelog.LogLevelNone:
		return LevelNone
	default:
		return LevelNone
	}
}

// determineLogLevel inspects the logger to determine the log level.
func (l Logger) determineLogLevel(ctx context.Context, levels []slog.Level) slog.Level {
	if l.logger == nil {
		return LevelNone
	}

	// Find the last enabled level, we start by the most verbose level until we find the first enabled level.
	// sort levels in descending order
	slices.Sort(levels)

	lastEnabled := LevelNone

	for _, level := range levels {
		if l.logger.Enabled(ctx, level) {
			lastEnabled = level
			break
		}
	}

	return lastEnabled
}

func (l Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	ll := l.converter.ToSlogLevel(level).Level()

	attrs := make([]slog.Attr, 0, len(data))

	for k, v := range data {
		attrs = append(attrs, slog.Any(k, v))
	}

	l.logger.LogAttrs(ctx, ll, msg, attrs...)
}

// NewTraceLogger creates a new trace logger.
func NewTraceLogger(opts ...LoggerOption) *tracelog.TraceLog {
	ll := newLogger(opts...)

	return &tracelog.TraceLog{
		Logger:   ll,
		LogLevel: ll.converter.ToTraceLogLevel(ll.level),
	}
}

// newLogger creates a new logger.
func newLogger(opts ...LoggerOption) Logger {
	o := &slog.HandlerOptions{
		Level: LevelNone,
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
	}

	handler := slog.NewTextHandler(os.Stdout, o)
	logger := Logger{
		logger:    slog.New(handler),
		converter: defaultLogLevelConverter{},
		level:     LevelNone,
	}

	for _, opt := range opts {
		opt(&logger)
	}

	return logger
}
