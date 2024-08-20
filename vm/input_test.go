package vm

import (
	"testing"
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
