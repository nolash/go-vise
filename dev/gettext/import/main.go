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
	newline bool
	d string
	w map[string]io.WriteCloser
	mw map[string]io.WriteCloser
}

func newTranslator(ctx context.Context, rs resource.Resource, outPath string, newline bool) *translator {
	return &translator{
		ctx: ctx,
		rs: rs,
		d: outPath,
		newline: newline,
		w: make(map[string]io.WriteCloser),
		mw: make(map[string]io.WriteCloser),
	}
}

func(tr *translator) applyLanguage(node *debug.Node) {

}

func(tr *translator) ensureFileNameFor(ln lang.Language, domain string) (string, error) {
	fileName := domain + ".po"
	p := path.Join(tr.d, ln.Code)
	err := os.MkdirAll(p, 0700)
	if err != nil {
		return "", err
	}
	return path.Join(p, fileName), nil
}

func(tr *translator) Close() error {
	var s string
	var err error
	for _, v := range(tr.langs) {
		o, ok := tr.w[v.Code]
		if ok {
			err = o.Close()
			if err != nil {
				s += fmt.Sprintf("\ntemplate writer close error %s: %v", v.Code, err)
			}
		}
		o, ok = tr.mw[v.Code]
		if ok {
			err = o.Close()
			if err != nil {
				s += fmt.Sprintf("\nmenu writer close error %s: %v", v.Code, err)
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

	for k, w := range(tr.mw) {
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

// TODO: DRY; merge with menuFunc
func(tr *translator) nodeFunc(node *debug.Node) error {
	for k, w := range(tr.w) {
		var s string
		ln, err := lang.LanguageFromCode(k)
		ctx := context.WithValue(tr.ctx, "Language", ln)
		r, err := tr.rs.GetTemplate(ctx, node.Name)
		for i, v := range(strings.Split(r, "\n")) {
			if tr.newline {
				if i > 0 {
					s += "\\n\"\n"
				} else if len(s) > 0 {
					s += "\"\n"
				}
				s += fmt.Sprintf("\t\"%s", v)
			} else {
				s += fmt.Sprintf("\t\"%s\"\n", v)
			}
		}
		if tr.newline {
			if len(s) > 0 {
				s += "\"\n"
			}
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
	s := fmt.Sprintf(`msgid ""
msgstr ""
	"Content-Type: text/plain; charset=UTF-8\n"
	"Language: %s\n"

`, ln.Code)

	filePath, err := tr.ensureFileNameFor(ln, resource.TemplatePoDomain)
	if err != nil {
		return err
	}
	w, err := os.OpenFile(filePath, os.O_WRONLY | os.O_CREATE, 0644)
	w.Write([]byte(s))
	tr.w[ln.Code] = w

	filePath, err = tr.ensureFileNameFor(ln, resource.MenuPoDomain)
	if err != nil {
		return err
	}
	w, err = os.OpenFile(filePath, os.O_WRONLY | os.O_CREATE, 0644)
	w.Write([]byte(s))
	tr.mw[ln.Code] = w

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
	var newline bool
	var langs langVar

	flag.StringVar(&dir, "d", ".", "node resource dir to read from")
	flag.StringVar(&outDir, "o", "locale", "output directory")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.BoolVar(&newline, "newline", false, "insert newlines in multiline strings")
	flag.Var(&langs, "l", "process for language")
	flag.Parse()

	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	err := os.MkdirAll(outDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "output dir create error: %v", err)
		os.Exit(1)
	}

	ctx := context.Background()
	rsStore := fsdb.NewFsDb()
	err = rsStore.Connect(ctx, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resource db connect error: %v", err)
		os.Exit(1)
	}

	rs := resource.NewDbResource(rsStore)

	tr := newTranslator(ctx, rs, outDir, newline)
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
