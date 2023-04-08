package resource

import (
	"context"
	"fmt"
)


type TestSizeResource struct {
	*MenuResource
}

func getTemplate(sym string, sizer *Sizer) (string, error) {
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
		return getFoo, nil
	case "bar":
		return getBar, nil
	case "baz":
		return getBaz, nil
	case "xyzzy":
		return getXyzzy, nil
	}
	return nil, fmt.Errorf("unknown func: %s", sym)
}

func getFoo(ctx context.Context) (string, error) {
	return "inky", nil
}

func getBar(ctx context.Context) (string, error) {
	return "pinky", nil
}

func getBaz(ctx context.Context) (string, error) {
	return "blinky", nil
}

func getXyzzy(ctx context.Context) (string, error) {
	return "inky pinky\nblinky clyde sue\ntinkywinky dipsy\nlala poo\none two three four five six seven\neight nine ten\neleven twelve", nil
}
