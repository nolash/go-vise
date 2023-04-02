package resource

import (
	"testing"
)

func TestNewFs(t *testing.T) {
	n := NewFsResource("./testdata")
	_ = n
}
