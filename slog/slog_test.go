package slogging

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lmittmann/tint"
)

func newTestHandler(w io.Writer, o SlogOpts) slog.Handler {
	return slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource: o.IncludeSource,
		Level:     o.LogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				switch a.Value.Any().(slog.Level) {
				case LevelTrace:
					return slog.String(slog.LevelKey, "TRACE")
				}
			}
			return a
		},
	})
}

func TestNewSlogOutput(t *testing.T) {
	var buf bytes.Buffer

	logger := NewSlog(SlogOpts{
		Handler: newTestHandler(&buf, SlogOpts{
			LogLevel:      LevelTrace,
			IncludeSource: true,
		}),
	})

	logger.Tracef("trace test")
	logger.Debugf("debug test")
	logger.Infof("info test")
	logger.Warnf("warn test")
	logger.Errorf("error test")

	logOutput := buf.String()
	t.Logf("Log output:\n %s", logOutput)
	if !strings.Contains(logOutput, "TRACE") || !strings.Contains(logOutput, "trace test") {
		t.Errorf("expected TRACE message in log output: %s", logOutput)
	}
	if !strings.Contains(logOutput, "DEBUG") || !strings.Contains(logOutput, "debug test") {
		t.Errorf("expected DEBUG message in log output: %s", logOutput)
	}
	if !strings.Contains(logOutput, "INFO") || !strings.Contains(logOutput, "info test") {
		t.Errorf("expected INFO message in log output: %s", logOutput)
	}
	if !strings.Contains(logOutput, "WARN") || !strings.Contains(logOutput, "warn test") {
		t.Errorf("expected WARN message in log output: %s", logOutput)
	}
	if !strings.Contains(logOutput, "ERROR") || !strings.Contains(logOutput, "error test") {
		t.Errorf("expected ERROR message in log output: %s", logOutput)
	}
}

func TestNewSlogOutputWithCtx(t *testing.T) {
	var buf bytes.Buffer

	logger := NewSlog(SlogOpts{
		Handler: newTestHandler(&buf, SlogOpts{
			LogLevel:      LevelTrace,
			IncludeSource: true,
		}),
		CtxKeys: []string{"y"},
	})

	ctx := context.Background()
	ctx = context.WithValue(ctx, "x", "1334")
	ctx = context.WithValue(ctx, "y", "why")
	ctx = context.WithValue(ctx, "z", "0.0.0.0")

	logger.InfoCtxf(ctx, "info test with ctx", "a", "apples")
	logOutput := buf.String()
	t.Logf("Log output:\n %s", logOutput)
	if !strings.Contains(logOutput, "apples") || !strings.Contains(logOutput, "why") {
		t.Errorf("expected a and y attributes in log output: %s", logOutput)
	}
}

func TestIntegrationWithColourfulHandler(t *testing.T) {
	logger := NewSlog(SlogOpts{
		Handler: tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
			AddSource:  true,
		}),
		CtxKeys: []string{"y", "z"},
	})

	ctx := context.Background()
	ctx = context.WithValue(ctx, "x", "1334")
	ctx = context.WithValue(ctx, "y", "why")
	ctx = context.WithValue(ctx, "z", "0.0.0.0")

	logger.WarnCtxf(ctx, "warn test with ctx", "b", "bananas")
}
