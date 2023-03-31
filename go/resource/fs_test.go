package resource

import (
	"context"
	"testing"
)

func TestNewFs(t *testing.T) {
	n := NewFsResource("./testdata", context.TODO())
	_ = n
}
