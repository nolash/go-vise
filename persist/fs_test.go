package persist

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/vm"
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

	pr := NewFsPersister(".").WithContent(&st, ca)
	v, err := pr.Serialize()
	if err != nil {
		t.Error(err)
	}

	prnew := NewFsPersister(".")
	err = prnew.Deserialize(v)
	if err != nil {
		t.Fatal(err)
	}	
	if !reflect.DeepEqual(prnew.State.ExecPath, pr.State.ExecPath) {
		t.Fatalf("expected %s, got %s", prnew.State.ExecPath, pr.State.ExecPath)
	}
	if !bytes.Equal(prnew.State.Code, pr.State.Code) {
		t.Fatalf("expected %x, got %x", prnew.State.Code, pr.State.Code)
	}
	if prnew.State.BitSize != pr.State.BitSize {
		t.Fatalf("expected %v, got %v", prnew.State.BitSize, pr.State.BitSize)
	}
	if prnew.State.SizeIdx != pr.State.SizeIdx {
		t.Fatalf("expected %v, got %v", prnew.State.SizeIdx, pr.State.SizeIdx)
	}
	if !reflect.DeepEqual(prnew.Memory, pr.Memory) {
		t.Fatalf("expected %v, got %v", prnew.Memory, pr.Memory)
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

	dir, err := ioutil.TempDir("", "vise_persist")
	if err != nil {
		t.Error(err)
	}
	pr := NewFsPersister(dir).WithContent(&st, ca)
	err = pr.Save("xyzzy")
	if err != nil {
		t.Error(err)
	}

	prnew := NewFsPersister(dir)
	err = prnew.Load("xyzzy")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(prnew.State.ExecPath, pr.State.ExecPath) {
		t.Fatalf("expected %s, got %s", prnew.State.ExecPath, pr.State.ExecPath)
	}
	if !bytes.Equal(prnew.State.Code, pr.State.Code) {
		t.Fatalf("expected %x, got %x", prnew.State.Code, pr.State.Code)
	}
	if prnew.State.BitSize != pr.State.BitSize {
		t.Fatalf("expected %v, got %v", prnew.State.BitSize, pr.State.BitSize)
	}
	if prnew.State.SizeIdx != pr.State.SizeIdx {
		t.Fatalf("expected %v, got %v", prnew.State.SizeIdx, pr.State.SizeIdx)
	}
	if !reflect.DeepEqual(prnew.Memory, pr.Memory) {
		t.Fatalf("expected %v, got %v", prnew.Memory, pr.Memory)
	}
}
