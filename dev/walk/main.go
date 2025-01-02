package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"git.defalsify.org/vise.git/debug"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/logging"
	fsdb "git.defalsify.org/vise.git/db/fs"

)

var (
	logg = logging.NewVanilla()
)

type translator struct {
	langs []lang.Language
	ctx context.Context
	rs resource.Resource
	fileBase string
	d string
	w map[string]io.WriteCloser
}

func newTranslator(ctx context.Context, rs resource.Resource, outPath string) *translator {
	return &translator{
		ctx: ctx,
		rs: rs,
		d: outPath,
		w: make(map[string]io.WriteCloser),
		fileBase: "visedata",
	}
}

func(tr *translator) applyLanguage(node *debug.Node) {

}

func(tr *translator) fileNameFor(ln lang.Language) string {
	fileName := tr.fileBase + "." + ln.Code + ".po"
	return path.Join(tr.d, fileName)
}

func(tr *translator) Close() error {
	var s string
	var err error
	for k, v := range(tr.langs) {
		o, ok := tr.w[v.Code]
		if ok {
			err = o.Close()
			if err != nil {
				s += fmt.Sprintf("\nclose error %s: %v", k, err)
			}
		}
	}
	if len(s) > 0 {
		err = fmt.Errorf("%s", s)
	}
	return err
}

func(tr *translator) process(s string) error {
	return nil
}

func(tr *translator) menuFunc(sym string) error {
	var v string

	for k, w := range(tr.w) {
		var s string
		ln, err := lang.LanguageFromCode(k)
		ctx := context.WithValue(tr.ctx, "Language", ln)
		r, err := tr.rs.GetMenu(ctx, sym)
		for _, v = range(strings.Split(r, "\n")) {
			s += fmt.Sprintf("\t\"%s\"\n", v)
		}
		s = fmt.Sprintf(`msgid ""
	"%s"
msgstr ""
%s

`, sym, s)
		if err == nil {
			logg.DebugCtxf(tr.ctx, "menu translation found", "node", sym)
			_, err = w.Write([]byte(s))
			if err != nil {
				return err
			}
		} else {
			logg.DebugCtxf(tr.ctx, "no menuitem translation found", "node", sym)
		}
	}
	return nil
}

func(tr *translator) nodeFunc(node *debug.Node) error {
	var v string

	for k, w := range(tr.w) {
		var s string
		ln, err := lang.LanguageFromCode(k)
		ctx := context.WithValue(tr.ctx, "Language", ln)
		r, err := tr.rs.GetTemplate(ctx, node.Name)
		for _, v = range(strings.Split(r, "\n")) {
			s += fmt.Sprintf("\t\"%s\"\n", v)
		}
		s = fmt.Sprintf(`msgid ""
	"%s"
msgstr ""
%s

`, node.Name, s)
		if err == nil {
			_, err = w.Write([]byte(s))
			if err != nil {
				return err
			}
		} else {
			logg.DebugCtxf(tr.ctx, "no template found", "node", node.Name)
		}
	}
	return nil
}

func(tr *translator) AddLang(ln lang.Language) error {
	filePath := tr.fileNameFor(ln)
	w, err := os.OpenFile(filePath, os.O_WRONLY | os.O_CREATE, 0644)
	s := fmt.Sprintf(`msgid ""
msgstr ""
	"Content-Type: text/plain; charset=UTF-8\n"
	"Language: %s\n"

`, ln.Code)
	w.Write([]byte(s))
	tr.w[ln.Code] = w
	return err
}

type langVar struct {
	v []lang.Language
}

func(lv *langVar) Set(s string) error {
	v, err := lang.LanguageFromCode(s)
	if err != nil {
		return err
	}
	lv.v = append(lv.v, v)
	return err
}

func(lv *langVar) String() string {
	var s []string
	for _, v := range(lv.v) {
		s = append(s, v.Code)
	}
	return strings.Join(s, ",")
}

func(lv *langVar) Langs() []lang.Language {
	return lv.v
}

func main() {
	var dir string
	var outDir string
	var root string
	var langs langVar

	flag.StringVar(&dir, "d", ".", "resource dir to read from")
	flag.StringVar(&outDir, "o", ".", "output directory")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.Var(&langs, "l", "process for language")
	flag.Parse()

	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	ctx := context.Background()
	rsStore := fsdb.NewFsDb()
	err := rsStore.Connect(ctx, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resource db connect error: %v", err)
		os.Exit(1)
	}

	rs := resource.NewDbResource(rsStore)

	tr := newTranslator(ctx, rs, outDir)
	defer tr.Close()
	for _, ln := range(langs.Langs()) {
		logg.DebugCtxf(ctx, "lang", "lang", ln)
		err = tr.AddLang(ln)
		if err != nil {
			fmt.Fprintf(os.Stderr, "add language failed for %s: %v", ln.Code, err)
			os.Exit(1)
		}
	}

	nm := debug.NewNodeMap(root)
	err = nm.Run(ctx, rs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "node tree process fail: %v", err)
		os.Exit(1)
	}
	
	for k, v := range(debug.NodeIndex) {
		err = tr.nodeFunc(&v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "translate process error for node %s: %v", k, err)
			os.Exit(1)
		}
	}

	for k, _ := range(debug.MenuIndex) {
		logg.Tracef("processing menu", "sym", k)
		err = tr.menuFunc(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "translate process error for menu %s: %v", k, err)
			os.Exit(1)
		}
	}

}
