package persist

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/vm"
	"git.defalsify.org/vise.git/db"
)

func TestSerializeState(t *testing.T) {
	st := state.NewState(12)
	st.Down("foo")
	st.Down("bar")
	st.Down("baz")
	st.Next()
	st.Next()

	b := vm.NewLine(nil, vm.LOAD, []string{"foo"}, []byte{42}, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	st.SetCode(b)

	ca := cache.NewCache().WithCacheSize(1024)
	ca.Add("inky", "pinky", 13)
	ca.Add("blinky", "clyde", 42)

	ctx := context.Background()
	store := db.NewMemDb()
	store.Connect(ctx, "")
	pr := NewPersister(store).WithContext(context.Background()).WithSession("xyzzy").WithContent(&st, ca)
	v, err := pr.Serialize()
	if err != nil {
		t.Error(err)
	}

	prnew := NewPersister(store).WithSession("xyzzy")
	err = prnew.Deserialize(v)
	if err != nil {
		t.Fatal(err)
	}
	stNew := prnew.GetState()
	stOld := pr.GetState()
	caNew := prnew.GetMemory()
	caOld := pr.GetMemory()

	if !reflect.DeepEqual(stNew.ExecPath, stOld.ExecPath) {
		t.Fatalf("expected %s, got %s", stNew.ExecPath, stOld.ExecPath)
	}
	if !bytes.Equal(stNew.Code, stOld.Code) {
		t.Fatalf("expected %x, got %x", stNew.Code, stOld.Code)
	}
	if stNew.BitSize != stOld.BitSize {
		t.Fatalf("expected %v, got %v", stNew.BitSize, stOld.BitSize)
	}
	if stNew.SizeIdx != stOld.SizeIdx {
		t.Fatalf("expected %v, got %v", stNew.SizeIdx, stOld.SizeIdx)
	}
	if !reflect.DeepEqual(caNew, caOld) {
		t.Fatalf("expected %v, got %v", caNew, caOld)
	}
}

func TestSaveLoad(t *testing.T) {
	st := state.NewState(12)
	st.Down("foo")
	st.Down("bar")
	st.Down("baz")
	st.Next()
	st.Next()

	b := vm.NewLine(nil, vm.LOAD, []string{"foo"}, []byte{42}, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	st.SetCode(b)

	ca := cache.NewCache().WithCacheSize(1024)
	ca.Add("inky", "pinky", 13)
	ca.Add("blinky", "clyde", 42)

	ctx := context.Background()
	store := db.NewMemDb()
	store.Connect(ctx, "")
	pr := NewPersister(store).WithContent(&st, ca)
	err := pr.Save("xyzzy")
	if err != nil {
		t.Fatal(err)
	}

	prnew := NewPersister(store)
	err = prnew.Load("xyzzy")
	if err != nil {
		t.Fatal(err)
	}
	stNew := prnew.GetState()
	stOld := pr.GetState()
	caNew := prnew.GetMemory()
	caOld := pr.GetMemory()

	if !reflect.DeepEqual(stNew.ExecPath, stOld.ExecPath) {
		t.Fatalf("expected %s, got %s", stNew.ExecPath, stOld.ExecPath)
	}
	if !bytes.Equal(stNew.Code, stOld.Code) {
		t.Fatalf("expected %x, got %x", stNew.Code, stOld.Code)
	}
	if stNew.BitSize != stOld.BitSize {
		t.Fatalf("expected %v, got %v", stNew.BitSize, stOld.BitSize)
	}
	if stNew.SizeIdx != stOld.SizeIdx {
		t.Fatalf("expected %v, got %v", stNew.SizeIdx, stOld.SizeIdx)
	}
	if !reflect.DeepEqual(caNew, caOld) {
		t.Fatalf("expected %v, got %v", caNew, caOld)
	}
}

func TestSaveLoadFlags(t *testing.T) {
	ctx := context.Background()
	st := state.NewState(2)
	st.SetFlag(8)
	ca := cache.NewCache()

	store := db.NewMemDb()
	store.Connect(ctx, "")
	pr := NewPersister(store).WithContent(&st, ca)
	err := pr.Save("xyzzy")
	if err != nil {
		t.Fatal(err)
	}

	prnew := NewPersister(store)
	
	err = prnew.Load("xyzzy")
	if err != nil {
		t.Fatal(err)
	}
	stnew := prnew.GetState()
	if !stnew.GetFlag(8) {
		t.Fatalf("expected flag 8 set")
	}
}
