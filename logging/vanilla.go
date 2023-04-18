package logging

import (
	"context"
	"fmt"
	"io"
	"path"
	"runtime"
)

type Vanilla struct {
	domain string
	levelFilter int
}

func NewVanilla() Vanilla {
	return Vanilla{
		domain: "main",
		levelFilter: LogLevel,
	}
}

func(v Vanilla) WithDomain(domain string) Vanilla {
	v.domain = domain
	return v
}

func(v Vanilla) WithLevel(level int) Vanilla {
	v.levelFilter = level
	return v
}

func(v Vanilla) Printf(level int, msg string, args ...any) {
	v.Writef(LogWriter, level, msg, args...)
}

func(v Vanilla) writef(w io.Writer, file string, line int, level int, msg string, args ...any) {
	if level > v.levelFilter {
		return
	}
	argsStr := argsToString(args)
	if len(msg) > 0 {
		fmt.Fprintf(w, "[%s] %s:%s:%v %s\t%s\n", v.AsString(level), v.domain, file, line, msg, argsStr)
	} else {
		fmt.Fprintf(w, "[%s] %s:%s:%v %s\n", v.AsString(level), v.domain, file, line, argsStr)
	}
}

func(v Vanilla) Writef(w io.Writer, level int, msg string, args ...any) {
	file, line := getCaller(2)
	v.writef(w, file, line, level, msg, args)
}

func argsToString(args []any) string {
	var s string
	c := len(args)
	var i int
	for i = 0; i < c; i += 2 {
		if len(s) > 0 {
			s += ", "
		}

		if i + 1 < c {
			var argByte []byte
			var ok bool
			argByte, ok = args[i+1].([]byte)
			if ok {
				s += fmt.Sprintf("%s=%x", args[i], argByte)
			} else {
				s += fmt.Sprintf("%s=%v", args[i], args[i+1])
			}
		} else {
			s += fmt.Sprintf("%s=??", args[i])
		}
	}
	return s
}

func(v Vanilla) WriteCtxf(ctx context.Context, w io.Writer, level int, msg string, args ...any) {
	v.Writef(w, level, msg, args...)
}

func(v Vanilla) printf(level int, msg string, args ...any) {
	file, line := getCaller(3)
	v.writef(LogWriter, file, line, level, msg, args...)
}

func(v Vanilla) printCtxf(ctx context.Context, level int, msg string, args ...any) {
	file, line := getCaller(3)
	v.writef(LogWriter, file, line, level, msg, args...)
}

func(v Vanilla) PrintCtxf(ctx context.Context, level int, msg string, args ...any) {
	v.printf(level, msg, args...)
}

func(v Vanilla) AsString(level int) string {
	return levelStr[level]	
}

func(v Vanilla) Tracef(msg string, args ...any) {
	v.printf(LVL_TRACE, msg, args...)
}

func(v Vanilla) Debugf(msg string, args ...any) {
	v.printf(LVL_DEBUG, msg, args...)
}

func(v Vanilla) Infof(msg string, args ...any) {
	v.printf(LVL_INFO, msg, args...)
}

func(v Vanilla) Warnf(msg string, args ...any) {
	v.printf(LVL_WARN, msg, args...)
}

func(v Vanilla) Errorf(msg string, args ...any) {
	v.printf(LVL_ERROR, msg, args...)
}

func(v Vanilla) TraceCtxf(ctx context.Context, msg string, args ...any) {
	v.printCtxf(ctx, LVL_TRACE, msg, args...)
}

func(v Vanilla) DebugCtxf(ctx context.Context, msg string, args ...any) {
	v.printCtxf(ctx, LVL_DEBUG, msg, args...)
}

func(v Vanilla) InfoCtxf(ctx context.Context, msg string, args ...any) {
	v.printCtxf(ctx, LVL_INFO, msg, args...)
}

func(v Vanilla) WarnCtxf(ctx context.Context, msg string, args ...any) {
	v.printCtxf(ctx, LVL_WARN, msg, args...)
}

func(v Vanilla) ErrorCtxf(ctx context.Context, msg string, args ...any) {
	v.printCtxf(ctx, LVL_ERROR, msg, args...)
}

func getCaller(depth int) (string, int) {
	var file string
	var line int
	_, file, line,_ = runtime.Caller(depth)
	baseFile := path.Base(file)
	return baseFile, line
}
