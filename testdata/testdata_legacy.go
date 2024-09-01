package testdata

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/vm"
)

type genFunc func() error

var (
	BaseDir = testdataloader.GetBasePath()
	DataDir = ""
	dirLock = false	
)

func outLegacy(sym string, b []byte, tpl string, data map[string]string) error {
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
	b = vm.NewLine(b, vm.MOUT, []string{"do the foo", "1"}, nil, nil)
	b = vm.NewLine(b, vm.MOUT, []string{"go to the bar", "2"}, nil, nil)
	b = vm.NewLine(b, vm.MOUT, []string{"language template", "3"}, nil, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"foo", "1"}, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"bar", "2"}, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"lang", "3"}, nil, nil)

	tpl := "hello world"

	return out("root", b, tpl, nil)
}

func foo() error {
	b := []byte{}
	b = vm.NewLine(b, vm.MOUT, []string{"to foo", "0"}, nil, nil)
	b = vm.NewLine(b, vm.MOUT, []string{"go bar", "1"}, nil, nil)
	b = vm.NewLine(b, vm.MOUT, []string{"see long", "2"}, nil, nil)
	b = vm.NewLine(b, vm.LOAD, []string{"inky"}, []byte{20}, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"_", "0"}, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"baz", "1"}, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"long", "2"}, nil, nil)

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
	b = vm.NewLine(b, vm.INCMP, []string{"^", "*"}, nil, nil)

	tpl := "this is bar - any input will return to top"

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

func long() error {
	b := []byte{}
	b = vm.NewLine(b, vm.MOUT, []string{"back", "0"}, nil, nil)
	b = vm.NewLine(b, vm.MNEXT, []string{"nexxt", "00"}, nil, nil)
	b = vm.NewLine(b, vm.MPREV, []string{"prevvv", "11"}, nil, nil)
	b = vm.NewLine(b, vm.LOAD, []string{"longdata"}, []byte{0x00}, nil)
	b = vm.NewLine(b, vm.MAP, []string{"longdata"}, nil, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"_", "0"}, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{">", "00"}, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"<", "11"}, nil, nil)

	tpl := `data
{{.longdata}}`

	data := make(map[string]string)
	data["longdata"] = `INKY 12
PINKY 5555
BLINKY 3t7
CLYDE 11
TINKYWINKY 22
DIPSY 666
LALA 111
POO 222
`

	return out("long", b, tpl, data)
}

func defaultCatch() error {
	b := []byte{}
	b = vm.NewLine(b, vm.MOUT, []string{"back", "0"}, nil, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"_", "*"}, nil, nil)

	tpl := "invalid input"

	return out("_catch", b, tpl, nil)
}

func lang() error {
	b := []byte{}
	b = vm.NewLine(b, vm.MOUT, []string{"back", "0"}, nil, nil)
	b = vm.NewLine(b, vm.LOAD, []string{"inky"}, []byte{20}, nil)
	b = vm.NewLine(b, vm.MAP, []string{"inky"}, nil, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.INCMP, []string{"_", "*"}, nil, nil)

	tpl := "this changes with language {{.inky}}"

	err := out("lang", b, tpl, nil)
	if err != nil {
		return err
	}

	tpl = "dette endrer med språket {{.inky}}"
	fp := path.Join(DataDir, "lang_nor")
	err = os.WriteFile(fp, []byte(tpl), 0600)
	if err != nil {
		return err
	}

	menu := "tilbake"
	fp = path.Join(DataDir, "back_menu_nor")
	return os.WriteFile(fp, []byte(menu), 0600)
}

func generateLegacy() error {
	out = outLegacy
	err := os.MkdirAll(DataDir, 0755)
	if err != nil {
		return err
	}

	fns := []genFunc{root, foo, bar, baz, long, lang, defaultCatch}
	for _, fn := range fns {
		err = fn()
		if err != nil {
			return err
		}
	}
	return nil
}

// Generate outputs bytecode, templates and content symbols to a temporary directory.
//
// This directory can in turn be used as data source for the the resource.FsResource object.
func GenerateLegacy() (string, error) {
	dir, err := ioutil.TempDir("", "vise_testdata_")
	if err != nil {
		return "", err
	}
	DataDir = dir
	dirLock = true
	err = generateLegacy()
	return dir, err
}


// Generate outputs bytecode, templates and content symbols to a specified directory.
//
// The directory must exist, and must not have been used already in the same code execution.
//
// This directory can in turn be used as data source for the the resource.FsResource object.
func GenerateLegacyTo(dir string) error {
	if dirLock {
		return fmt.Errorf("directory already overridden")
	}
	DataDir = dir
	dirLock = true
	return generateLegacy()
}
