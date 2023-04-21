package resource

import (
	"testing"
)

func TestNewFs(t *testing.T) {
	n := NewFsResource("./testdata")
	_ = n
}
//
//func TestResourceLanguage(t *testing.T) {
//	var err error
//	generateTestData(t)
//	ctx := context.TODO()
//	st := state.NewState(0)
//	rs := NewFsWrapper(dataDir, &st)
//	ca := cache.NewCache()
//
//	cfg := Config{
//		Root: "root",
//	}
//
//	en := NewEngine(cfg, &st, &rs, ca, ctx)
//	_, err = en.Init(ctx)
//	if err == nil {
//		t.Fatalf("expected error")
//	}
//	cfg = Config{
//		Root: "root",
//	}
//	en = NewEngine(cfg, &st, &rs, ca, ctx)
//	_, err = en.Init(ctx)
//	if err != nil {
//		t.Fatal(err)
//	}
//}
