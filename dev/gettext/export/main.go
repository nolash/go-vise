package main

import (
	"context"
	"flag"
	"fmt"
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
	haveLang map[string]bool
	ctx context.Context
	rs *resource.PoResource
	outPath string
	madePath bool
}

func newTranslator(ctx context.Context, defaultLanguage lang.Language, inPath string, outPath string) *translator {
	tr := &translator{
		langs: []lang.Language{defaultLanguage},
		haveLang: make(map[string]bool),
		ctx: ctx,
		rs: resource.NewPoResource(defaultLanguage, inPath),
		outPath: outPath,
	}
	tr.haveLang[defaultLanguage.Code] = true
	return tr
}

func(tr *translator) AddLang(ln lang.Language) error {
	var ok bool
	_, ok = tr.haveLang[ln.Code]
	if !ok {
		tr.langs = append(tr.langs, ln)
		tr.rs = tr.rs.WithLanguage(ln)
		tr.haveLang[ln.Code] = true
	}
	return nil
}

func(tr *translator) nodeFunc(node *debug.Node) error {
	sym := node.Name
	for i, ln := range(tr.langs) {
		s, err := tr.rs.GetTemplate(tr.ctx, sym)
		if err != nil {
			logg.DebugCtxf(tr.ctx, "template not found", "sym", s)
			continue
		}
		if s != sym {
			if !tr.madePath {
				err := os.MkdirAll(tr.outPath, 0700)
				if err != nil {
					return err
				}
			}
			fb := sym
			if i > 0 {
				fb += "_" + ln.Code
			}
			fp := path.Join(tr.outPath, fb)
			w, err := os.OpenFile(fp, os.O_WRONLY | os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			c, err := w.Write([]byte(s))
			defer w.Close()
			if err != nil {
				return err
			}
			logg.DebugCtxf(tr.ctx, "wrote node", "sym", sym, "lang", ln.Code, "bytes", c)
		}
	}
	return nil
}

func(tr *translator) menuFunc(sym string) error {
	for i, ln := range(tr.langs) {
		s, err := tr.rs.GetTemplate(tr.ctx, sym)
		if err != nil {
			logg.DebugCtxf(tr.ctx, "template not found", "sym", s)
			continue
		}
		if s != sym {
			if !tr.madePath {
				err := os.MkdirAll(tr.outPath, 0700)
				if err != nil {
					return err
				}
			}
			// TODO: use menu sym generator func instead
			fb := sym + "_menu"
			if i > 0 {
				fb += "_" + ln.Code
			}
			// TODO: use lang filename generator func instead
			fp := path.Join(tr.outPath, fb)
			w, err := os.OpenFile(fp, os.O_WRONLY | os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			c, err := w.Write([]byte(s))
			defer w.Close()
			if err != nil {
				return err
			}
			logg.DebugCtxf(tr.ctx, "wrote menu", "sym", sym, "lang", ln.Code, "bytes", c)
		}
	}
}

func(tr *translator) Close() error {
	return nil
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
	var inDir string
	var outDir string
	var root string
	var langs langVar
	var defaultLanguage string

	flag.StringVar(&dir, "d", ".", "node resource dir to read from")
	flag.StringVar(&inDir, "i", "", "gettext dir")
	flag.StringVar(&outDir, "o", "locale", "output directory")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.Var(&langs, "l", "process for language")
	flag.StringVar(&defaultLanguage, "defaultlanguage", "eng", "default language to resolve for")
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

	ln, err := lang.LanguageFromCode(defaultLanguage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid default language: %v", err)
		os.Exit(1)
	}
	tr := newTranslator(ctx, ln, inDir, outDir)
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
