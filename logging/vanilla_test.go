package logging

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestVanilla(t *testing.T) {
	logg := NewVanilla().WithDomain("test").WithLevel(LVL_WARN)
	w := bytes.NewBuffer(nil)
	logg.Writef(w, LVL_DEBUG, "message", "xyzzy", 666, "inky", "pinky")
	if len(w.Bytes()) > 0 {
		t.Errorf("expected nothing, got %s", w.Bytes())
	}
	logg = logg.WithLevel(LVL_DEBUG)
	logg.Writef(w, LVL_DEBUG, "message", "xyzzy", 666, "inky", "pinky")
	if len(w.Bytes()) == 0 {
		t.Errorf("expected output")
	}
}

func TestVanillaCtx(t *testing.T) {
	logg := NewVanilla().WithDomain("test").WithLevel(LVL_DEBUG).WithContextKey("foo")
	ctx := context.WithValue(context.Background(), "foo", "bar") 
	w := bytes.NewBuffer(nil)
	LogWriter = w

	logg.DebugCtxf(ctx, "message", "xyzzy", 666, "inky", "pinky")
	s := string(w.Bytes())
	if len(s) == 0 {
		t.Errorf("expected output")
	}
	if !strings.Contains(s, "foo=bar") {
		t.Errorf("expected 'foo=bar' in output, output was: %s", s)
	}
	if !strings.Contains(s, "test") {
		t.Errorf("expected 'test' in output, output was: %s", s)
	}
}
