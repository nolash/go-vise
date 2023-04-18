package logging

import (
	"bytes"
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
