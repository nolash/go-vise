package logging

import (
	"context"
	"io"
	"os"
)

const (
	LVL_NONE = iota
	LVL_ERROR
	LVL_WARN
	LVL_INFO
	LVL_DEBUG
	LVL_TRACE
)

var (
	levelStr = map[int]string{
		LVL_ERROR: "E",	
		LVL_WARN: "W",	
		LVL_INFO: "I",	
		LVL_DEBUG: "D",	
		LVL_TRACE: "T",	
	}
)

var (
	LogWriter = os.Stderr
)


func AsString(level int) string {
	return levelStr[level]	
}


type Logger interface {
	Writef(w io.Writer, level int, msg string, args ...any)
	WriteCtxf(ctx context.Context, w io.Writer, level int, msg string, args ...any)
	Printf(level int, msg string, args ...any)
	PrintCtxf(ctx context.Context, level int, msg string, args ...any)
	Tracef(msg string, args ...any)
	TraceCtxf(ctx context.Context, msg string, args ...any)
	Debugf(msg string, args ...any)
	DebugCtxf(ctx context.Context, msg string, args ...any)
	Infof(msg string, args ...any)
	InfoCtxf(ctx context.Context, msg string, args ...any)
	Warnf(msg string, args ...any)
	WarnCtxf(ctx context.Context, msg string, args ...any)
	Errorf(msg string, args ...any)
	ErrorCtxf(ctx context.Context, msg string, args ...any)
}

