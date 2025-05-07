package slogging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"

	"git.defalsify.org/vise.git/logging"
)

const (
	LevelTrace slog.Level = slog.Level(-8)
)

var _ logging.Logger = (*Slog)(nil)

type (
	Slog struct {
		slogger *slog.Logger
		ctxKeys []string
	}

	SlogOpts struct {
		// Component enriches each log line with a componenent key/value.
		// Useful for aggregating/filtering with your log collector.
		Component string
		// Handler allows overriding of the defult Logfmt handler.
		Handler slog.Handler
		// Minimal level to log. Defaults to Info.
		// No effect when passing a custom handler.
		LogLevel slog.Level
		// Add source location to each log line. Defaults to false.
		// No effect when passing a custom handler.
		IncludeSource bool
		// CtxKeys are the known keys to be used for logging context values.
		CtxKeys []string
	}
)

// NewSlog creates a new Slog logger instance.
func NewSlog(o SlogOpts) *Slog {
	if o.Handler == nil {
		o.Handler = buildDefaultHandler(os.Stderr, o.LogLevel, o.IncludeSource)
	}
	if o.Component == "" {
		o.Component = "vise"
	}
	return &Slog{
		slogger: slog.New(o.Handler).With("component", o.Component),
		ctxKeys: o.CtxKeys,
	}
}

func buildDefaultHandler(w io.Writer, level slog.Level, includeSource bool) slog.Handler {
	return slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource: includeSource,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				switch a.Value.Any().(slog.Level) {
				// stdlib slog does not support TRACE level, so we map it to a custom string.
				case LevelTrace:
					return slog.String(slog.LevelKey, "TRACE")
				}
			}
			return a
		},
	})
}

func (s *Slog) logWithCaller(ctx context.Context, level slog.Level, msg string, args ...any) {
	if !s.slogger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])

	record := slog.NewRecord(time.Now(), level, msg, pcs[0])

	if len(args) > 0 {
		for i := 0; i < len(args)-1; i += 2 {
			key, ok := args[i].(string)
			if !ok {
				continue
			}
			record.AddAttrs(slog.Any(key, args[i+1]))
		}
	}

	_ = s.slogger.Handler().Handle(ctx, record)
}

func (s *Slog) logWithCallerCtx(ctx context.Context, level slog.Level, msg string, args ...any) {
	if !s.slogger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])

	record := slog.NewRecord(time.Now(), level, msg, pcs[0])

	attrs := s.extractContextKeys(ctx, args...)
	record.AddAttrs(attrs...)

	_ = s.slogger.Handler().Handle(ctx, record)
}

func (s *Slog) Writef(w io.Writer, level int, msg string, args ...any) {
	s.slogger.Warn("Writef not implemented")
}

func (s *Slog) WriteCtxf(ctx context.Context, w io.Writer, level int, msg string, args ...any) {
	s.slogger.Warn("WriteCtxf not implemented")
}

func (s *Slog) Printf(level int, msg string, args ...any) {
	s.slogger.Warn("Printf not implemented")
}

func (s *Slog) PrintCtxf(ctx context.Context, level int, msg string, args ...any) {
	s.slogger.Warn("PrintCtxf not implemented")
}

func (s *Slog) Tracef(msg string, args ...any) {
	s.logWithCaller(context.Background(), LevelTrace, msg, args...)
}

func (s *Slog) TraceCtxf(ctx context.Context, msg string, args ...any) {
	s.logWithCallerCtx(ctx, LevelTrace, msg, args...)
}

func (s *Slog) Debugf(msg string, args ...any) {
	s.logWithCaller(context.Background(), slog.LevelDebug, msg, args...)
}

func (s *Slog) DebugCtxf(ctx context.Context, msg string, args ...any) {
	s.logWithCallerCtx(ctx, slog.LevelDebug, msg, args...)
}

func (s *Slog) Infof(msg string, args ...any) {
	s.logWithCaller(context.Background(), slog.LevelInfo, msg, args...)
}

func (s *Slog) InfoCtxf(ctx context.Context, msg string, args ...any) {
	s.logWithCallerCtx(ctx, slog.LevelInfo, msg, args...)
}

func (s *Slog) Warnf(msg string, args ...any) {
	s.logWithCaller(context.Background(), slog.LevelWarn, msg, args...)
}

func (s *Slog) WarnCtxf(ctx context.Context, msg string, args ...any) {
	s.logWithCallerCtx(ctx, slog.LevelWarn, msg, args...)
}

func (s *Slog) Errorf(msg string, args ...any) {
	s.logWithCaller(context.Background(), slog.LevelError, msg, args...)
}

func (s *Slog) ErrorCtxf(ctx context.Context, msg string, args ...any) {
	s.logWithCallerCtx(ctx, slog.LevelError, msg, args...)
}

func (s *Slog) extractContextKeys(ctx context.Context, args ...any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(s.ctxKeys)+len(args)/2)

	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		attrs = append(attrs, slog.Any(key, args[i+1]))
	}

	for _, key := range s.ctxKeys {
		if val, ok := ctx.Value(key).(string); ok {
			attrs = append(attrs, slog.String(key, val))
		}
	}

	return attrs
}
