package vm

import (
	"testing"

	"git.defalsify.org/vise.git/state"
)

func TestPhoneInput(t *testing.T) {
	err := ValidInput([]byte("+12345"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestMenuInputs(t *testing.T) {
	err := ValidInput([]byte("0"))
	if err != nil {
		t.Fatal(err)
	}

	err = ValidInput([]byte("99"))
	if err != nil {
		t.Fatal(err)
	}

	err = ValidInput([]byte("foo"))
	if err != nil {
		t.Fatal(err)
	}

	err = ValidInput([]byte("foo Bar"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestFalseInput(t *testing.T) {
	err := ValidInput([]byte{0x0a})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTargetInput(t *testing.T) {
	var err error
	st := state.NewState(1)
	_, err = CheckTarget([]byte(""), st)
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = CheckTarget([]byte("_"), st)
	if err == nil {
		t.Fatal("expected error")
	}
	st.Down("foo")
	v, err := CheckTarget([]byte("_"), st)
	if err != nil {
		t.Fatal(err)
	}
	if !v {
		t.Fatal("expected true")
	}
	v, err = CheckTarget([]byte("<"), st)
	if err != nil {
		t.Fatal(err)
	}
	if v {
		t.Fatal("expected false")
	}
	v, err = CheckTarget([]byte(">"), st)
	if err != nil {
		t.Fatal(err)
	}
	if !v {
		t.Fatal("expected true")
	}
	v, err = CheckTarget([]byte("%"), st)
	if err == nil {
		t.Fatal("expected error")
	}
	v, err = CheckTarget([]byte("foo"), st)
	if err != nil {
		t.Fatal(err)
	}
	if !v {
		t.Fatal("expected true")
	}
}
