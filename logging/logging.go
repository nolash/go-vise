package logging

import (
	"context"
	"io"
)

const (
	// No logging
	LVL_NONE = iota
	// Error log level - unexpected states that need immediate developer attention.
	LVL_ERROR
	// Warning log level - unexpected states that may need developer attention.
	LVL_WARN
	// Info log level - expected states that are of general interest to the end-user.
	LVL_INFO
	// Debug log level - expected states that are of general interest to the developer.
	LVL_DEBUG
	// Trace log level - everything else.
	LVL_TRACE
)

var (
	levelStr = map[int]string{
		LVL_ERROR: "E",
		LVL_WARN:  "W",
		LVL_INFO:  "I",
		LVL_DEBUG: "D",
		LVL_TRACE: "T",
	}
)

// AsString returns the string representation used in logging output for the given log level.
func AsString(level int) string {
	return levelStr[level]
}

type Logger interface {
	// Writef logs a line to the given writer with the given loglevel.
	Writef(w io.Writer, level int, msg string, args ...any)
	// WriteCtxf logs a line with context to the given writer with the given loglevel.
	WriteCtxf(ctx context.Context, w io.Writer, level int, msg string, args ...any)
	// Printf logs a line to the default writer with the given loglevel.
	Printf(level int, msg string, args ...any)
	// Printf logs a line with context to the default writer with the given loglevel.
	PrintCtxf(ctx context.Context, level int, msg string, args ...any)
	// Tracef logs a line to the default writer the TRACE loglevel.
	Tracef(msg string, args ...any)
	// TraceCtxf logs a line with context to the default writer the TRACE loglevel.
	TraceCtxf(ctx context.Context, msg string, args ...any)
	// Debugf logs a line to the default writer the DEBUG loglevel.
	Debugf(msg string, args ...any)
	// DebugCtxf logs a line with context to the default writer the DEBUG loglevel.
	DebugCtxf(ctx context.Context, msg string, args ...any)
	// Infof logs a line to the default writer the INFO loglevel.
	Infof(msg string, args ...any)
	// InfoCtxf logs a line with context to the default writer the INFO loglevel.
	InfoCtxf(ctx context.Context, msg string, args ...any)
	// Warnf logs a line to the default writer the WARN loglevel.
	Warnf(msg string, args ...any)
	// WarnCtxf logs a line with context to the default writer the WARN loglevel.
	WarnCtxf(ctx context.Context, msg string, args ...any)
	// Errorf logs a line to the default writer the ERROR loglevel.
	Errorf(msg string, args ...any)
	// ErrorCtxf logs a line with context to the default writer the ERROR loglevel.
	ErrorCtxf(ctx context.Context, msg string, args ...any)
}
