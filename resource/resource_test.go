package resource

import (
	"context"
	"fmt"
)


type TestSizeResource struct {
	*MenuResource
}

func getTemplate(sym string) (string, error) {
	var tpl string
	switch sym {
	case "small":
		tpl = "one {{.foo}} two {{.bar}} three {{.baz}}"
	case "toobig":
		tpl = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus in mattis lorem. Aliquam erat volutpat. Ut vitae metus."
	case "pages":
		tpl = "one {{.foo}} two {{.bar}} three {{.baz}}\n{{.xyzzy}}"
	}
	return tpl, nil
}

func funcFor(sym string) (EntryFunc, error) {
	switch sym {
	case "foo":
		return get, nil
	case "bar":
		return get, nil
	case "baz":
		return get, nil
	case "xyzzy":
		return getXyzzy, nil
	}
	return nil, fmt.Errorf("unknown func: %s", sym)
}

func get(sym string, ctx context.Context) (string, error) {
	switch sym {
	case "foo":
		return "inky", nil
	case "bar":
		return "pinky", nil
	case "baz":
		return "blinky", nil
	}
	return "", fmt.Errorf("unknown sym: %s", sym)
}

func getXyzzy(sym string, ctx context.Context) (string, error) {
	return "inky pinky\nblinky clyde sue\ntinkywinky dipsy\nlala poo\none two three four five six seven\neight nine ten\neleven twelve", nil
}
