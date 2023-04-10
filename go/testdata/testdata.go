package testdata

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/festive/vm"
)

type genFunc func() error

var (
	BaseDir = testdataloader.GetBasePath()
	DataDir = ""
	dirLock = false	
)

func out(sym string, b []byte, tpl string, data map[string]string) error {
	fp := path.Join(DataDir, sym)
	err := ioutil.WriteFile(fp, []byte(tpl), 0644)
	if err != nil {
		return err
	}

	fb := sym + ".bin"
	fp = path.Join(DataDir, fb)
	err = ioutil.WriteFile(fp, b, 0644)
	if err != nil {
		return err
	}

	if data == nil {
		return nil
	}

	for k, v := range data {
		fb := k + ".txt"
		fp = path.Join(DataDir, fb)
		err = ioutil.WriteFile(fp, []byte(v), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func root() error {
	b := []byte{}
	b = vm.NewLine(b, vm.MOUT, []string{"1", "do the foo"}, nil, nil)
	b = vm.NewLine(b, vm.MOUT, []string{"2", "go to the bar"}, nil, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"1", "foo"}, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"2", "bar"}, nil, nil)

	tpl := "hello world"

	return out("root", b, tpl, nil)
}

func foo() error {
	b := []byte{}
	b = vm.NewLine(b, vm.MOUT, []string{"0", "to foo"}, nil, nil)
	b = vm.NewLine(b, vm.MOUT, []string{"1", "go bar"}, nil, nil)
	b = vm.NewLine(b, vm.LOAD, []string{"inky"}, []byte{20}, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"0", "_"}, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"1", "baz"}, nil, nil)
	//b = vm.NewLine(b, vm.CATCH, []string{"_catch"}, []byte{1}, []uint8{1})

	data := make(map[string]string)
	data["inky"] = "one"

	tpl := `this is in foo

it has more lines`

	return out("foo", b, tpl, data)
}

func bar() error {
	b := []byte{}
	b = vm.NewLine(b, vm.LOAD, []string{"pinky"}, []byte{0}, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"*", "^"}, nil, nil)

	tpl := "this is bar - an end node"

	data := make(map[string]string)
	data["pinky"] = "two"

	return out("bar", b, tpl, data)
}

func baz() error {
	b := []byte{}
	b = vm.NewLine(b, vm.MAP, []string{"inky"}, nil, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)

	tpl := "this is baz which uses the var {{.inky}} in the template."

	return out("baz", b, tpl, nil)
}

func defaultCatch() error {
	b := []byte{}
	b = vm.NewLine(b, vm.MOUT, []string{"0", "back"}, nil, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"*", "_"}, nil, nil)

	tpl := "invalid input"

	return out("_catch", b, tpl, nil)
}

func generate() error {
	err := os.MkdirAll(DataDir, 0755)
	if err != nil {
		return err
	}

	fns := []genFunc{root, foo, bar, baz, defaultCatch}
	for _, fn := range fns {
		err = fn()
		if err != nil {
			return err
		}
	}
	return nil
}

func Generate() (string, error) {
	dir, err := ioutil.TempDir("", "festive_testdata_")
	if err != nil {
		return "", err
	}
	DataDir = dir
	dirLock = true
	err = generate()
	return dir, err
}

func GenerateTo(dir string) error {
	if dirLock {
		return fmt.Errorf("directory already overridden")
	}
	DataDir = dir
	dirLock = true
	return generate()
}
