package slogging

import (
	"context"
	"io"
	"log/slog"
	"os"

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
		// TODO: Could this be a functional option?
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
	s.slogger.Log(nil, LevelTrace, msg, args...)
}

func (s *Slog) TraceCtxf(ctx context.Context, msg string, args ...any) {
	s.slogger.LogAttrs(ctx, LevelTrace, msg, s.extractContextKeys(ctx, args...)...)
}

func (s *Slog) Debugf(msg string, args ...any) {
	s.slogger.Debug(msg, args...)
}

func (s *Slog) DebugCtxf(ctx context.Context, msg string, args ...any) {
	s.slogger.LogAttrs(ctx, slog.LevelDebug, msg, s.extractContextKeys(ctx, args...)...)
}

func (s *Slog) Infof(msg string, args ...any) {
	s.slogger.Info(msg, args...)
}

func (s *Slog) InfoCtxf(ctx context.Context, msg string, args ...any) {
	s.slogger.LogAttrs(ctx, slog.LevelInfo, msg, s.extractContextKeys(ctx, args...)...)
}

func (s *Slog) Warnf(msg string, args ...any) {
	s.slogger.Warn(msg, args...)
}

func (s *Slog) WarnCtxf(ctx context.Context, msg string, args ...any) {
	s.slogger.LogAttrs(ctx, slog.LevelWarn, msg, s.extractContextKeys(ctx, args...)...)
}

func (s *Slog) Errorf(msg string, args ...any) {
	s.slogger.Error(msg, args...)
}

func (s *Slog) ErrorCtxf(ctx context.Context, msg string, args ...any) {
	s.slogger.LogAttrs(ctx, slog.LevelError, msg, s.extractContextKeys(ctx, args...)...)
}

func (s *Slog) extractContextKeys(ctx context.Context, args ...any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(s.ctxKeys)+len(args))

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
