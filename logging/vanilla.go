package logging

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
)

var (
	// LogWriter is used as io.Writer the Vanilla Logger implementation.
	LogWriter io.Writer = os.Stderr
)

// Vanilla is a basic single-line structured output logger for terminal output.
type Vanilla struct {
	domain      string
	levelFilter int
	ctxkey      []string
}

// NewVanilla creates a new Vanilla logger.
func NewVanilla() Vanilla {
	return Vanilla{
		domain:      "main",
		levelFilter: LogLevel,
	}
}

// WithDomain sets the logging domain. It is prepended to the caller file/line information.
func (v Vanilla) WithDomain(domain string) Vanilla {
	v.domain = domain
	return v
}

// WithLevel overrides the globally set loglevel for the logger instance.
func (v Vanilla) WithLevel(level int) Vanilla {
	v.levelFilter = level
	return v
}

func (v Vanilla) WithContextKey(k string) Vanilla {
	v.ctxkey = append(v.ctxkey, k)
	return v
}

// Printf logs to the global writer.
func (v Vanilla) Printf(level int, msg string, args ...any) {
	v.Writef(LogWriter, level, msg, args...)
}

// compile log line from inputs and send to given writer.
func (v Vanilla) writef(ctx context.Context, w io.Writer, file string, line int, level int, msg string, args ...any) {
	var argsStr string
	if level > v.levelFilter {
		return
	}
	if ctx == nil {
		argsStr = v.argsToString(nil, args)
	} else {
		argsStr = v.argsToString(ctx, args)
	}

	if len(msg) > 0 {
		fmt.Fprintf(w, "[%s] %s:%s:%v %s\t%s\n", AsString(level), v.domain, file, line, msg, argsStr)
	} else {
		fmt.Fprintf(w, "[%s] %s:%s:%v %s\n", AsString(level), v.domain, file, line, argsStr)
	}
}

// Writef logs to the given writer.
func (v Vanilla) Writef(w io.Writer, level int, msg string, args ...any) {
	file, line := getCaller(2)
	v.writef(nil, w, file, line, level, msg, args)
}

// WriteCtxf logs with context to the given writer.
func (v Vanilla) WriteCtxf(ctx context.Context, w io.Writer, level int, msg string, args ...any) {
	file, line := getCaller(2)
	v.writef(ctx, w, file, line, level, msg, args...)
}

// get caller information and pass on to writef
func (v Vanilla) printf(level int, msg string, args ...any) {
	file, line := getCaller(3)
	v.writef(nil, LogWriter, file, line, level, msg, args...)
}

// get caller information and pass on to writef
func (v Vanilla) printCtxf(ctx context.Context, level int, msg string, args ...any) {
	file, line := getCaller(3)
	v.writef(ctx, LogWriter, file, line, level, msg, args...)
}

// PrintCtxf logs with context to the global writer.
func (v Vanilla) PrintCtxf(ctx context.Context, level int, msg string, args ...any) {
	v.printf(level, msg, args...)
}

// Tracef logs a line with level TRACE to the global writer.
func (v Vanilla) Tracef(msg string, args ...any) {
	v.printf(LVL_TRACE, msg, args...)
}

// Debugf logs a line with level DEBUG to the global writer.
func (v Vanilla) Debugf(msg string, args ...any) {
	v.printf(LVL_DEBUG, msg, args...)
}

// Infof logs a line with level INFO to the global writer.
func (v Vanilla) Infof(msg string, args ...any) {
	v.printf(LVL_INFO, msg, args...)
}

// Warnf logs a line with level WARN to the global writer.
func (v Vanilla) Warnf(msg string, args ...any) {
	v.printf(LVL_WARN, msg, args...)
}

// Errorf logs a line with level ERROR to the global writer.
func (v Vanilla) Errorf(msg string, args ...any) {
	v.printf(LVL_ERROR, msg, args...)
}

// TraceCtxf logs a line with context with level TRACE to the global writer.
func (v Vanilla) TraceCtxf(ctx context.Context, msg string, args ...any) {
	v.printCtxf(ctx, LVL_TRACE, msg, args...)
}

// DebugCtxf logs a line with context with level DEBUG to the global writer.
func (v Vanilla) DebugCtxf(ctx context.Context, msg string, args ...any) {
	v.printCtxf(ctx, LVL_DEBUG, msg, args...)
}

// InfoCtxf logs a line with context with level INFO to the global writer.
func (v Vanilla) InfoCtxf(ctx context.Context, msg string, args ...any) {
	v.printCtxf(ctx, LVL_INFO, msg, args...)
}

// WarnCtxf logs a line with context with level WARN to the global writer.
func (v Vanilla) WarnCtxf(ctx context.Context, msg string, args ...any) {
	v.printCtxf(ctx, LVL_WARN, msg, args...)
}

// ErrorCtxf logs a line with context with level ERROR to the global writer.
func (v Vanilla) ErrorCtxf(ctx context.Context, msg string, args ...any) {
	v.printCtxf(ctx, LVL_ERROR, msg, args...)
}

// return file basename and line for caller information.
func getCaller(depth int) (string, int) {
	var file string
	var line int
	_, file, line, _ = runtime.Caller(depth)
	baseFile := path.Base(file)
	return baseFile, line
}

// string representation of the given structured log args.
func (v Vanilla) argsToString(ctx context.Context, args []any) string {
	var s string

	if ctx != nil {
		for _, k := range v.ctxkey {
			v := ctx.Value(k)
			if v != nil {
				v, ok := v.(string)
				if ok {
					args = append(args, "x-"+k, v)
				}
			}
		}
	}
	c := len(args)
	var i int
	for i = 0; i < c; i += 2 {
		if len(s) > 0 {
			s += ", "
		}

		if i+1 < c {
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
