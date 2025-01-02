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
	writeDomains = []string{
		resource.PoDomain,
		resource.TemplateKeyPoDomain,
		resource.MenuKeyPoDomain,
	}
	writeDomainReady = make(map[string]bool)
)

type translator struct {
	langs []lang.Language
	ctx context.Context
	rs resource.Resource
	newline bool
	d string
}

func newTranslator(ctx context.Context, rs resource.Resource, outPath string, newline bool) *translator {
	return &translator{
		ctx: ctx,
		rs: rs,
		d: outPath,
		newline: newline,
	}
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

// skip default*.po for translations other than default
func(tr *translator) writersFor(ln lang.Language) ([]io.WriteCloser, error) {
	var r []io.WriteCloser
	_, ready := writeDomainReady[ln.Code]
	for _, v := range(writeDomains) {
		fp, err := tr.ensureFileNameFor(ln, v)
		if err != nil {
			return r, err
		}
		if !ready {
			w, err := os.OpenFile(fp, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644)
			if err != nil {
				return r, err
			}
			s := fmt.Sprintf(`msgid ""
msgstr ""
	"Content-Type: text/plain; charset=UTF-8\n"
	"Language: %s\n"

`, ln.Code)
			_, err = w.Write([]byte(s))
			if err != nil {
				return r, err
			}
			w.Close()
		}
		w, err := os.OpenFile(fp, os.O_WRONLY | os.O_APPEND, 0644)
		logg.DebugCtxf(tr.ctx, "writer", "fp", fp)
		if err != nil {
			return r, err
		}
		r = append(r, w)
	}
	writeDomainReady[ln.Code] = true
	return r, nil
}

func(tr *translator) writeTranslation(w io.Writer, sym string, msgid string, msgstr string) error {
	s := fmt.Sprintf(`#: vise_node.%s
msgid ""
%s
msgstr ""
%s

`, sym, msgid, msgstr)
	_, err := w.Write([]byte(s))
	if err != nil {
		return err
	}
	return nil
}

func(tr *translator) closeWriters(writers []io.WriteCloser) {
	for _, w := range(writers) {
		w.Close()
	}
}

// TODO: DRY; merge with menuFunc
func(tr *translator) nodeFunc(node *debug.Node) error {
	var def string
	for i, ln := range(tr.langs) {
		var s string
		ww, err := tr.writersFor(ln)
		defer tr.closeWriters(ww)
		if err != nil {
			return fmt.Errorf("failed writers for lang '%s': %v", ln.Code, err)
		}
		ctx := context.WithValue(tr.ctx, "Language", ln)
		r, err := tr.rs.GetTemplate(ctx, node.Name)
		if err == nil {
			logg.TraceCtxf(tr.ctx, "template found", "lang", ln, "node", node.Name)
			for i, v := range(strings.Split(r, "\n")) {
				if i > 0 {
					if tr.newline {
						s += fmt.Sprintf("\t\"\\n\"\n")
					}
				}
				s += fmt.Sprintf("\t\"%s\"\n", v)
			}
			if def == "" {
				def = fmt.Sprintf("\t\"%s\"\n", node.Name)
				err = tr.writeTranslation(ww[1], node.Name, def, s)
			}
			if i == 0 {
				def = s
			}
			err = tr.writeTranslation(ww[0], node.Name, def, s)
			if err != nil {
				return err
			}
		} else {
			logg.DebugCtxf(tr.ctx, "no template found", "node", node.Name, "lang", ln)
		}
	}
	return nil
}

// TODO: drop the multiline gen
func(tr *translator) menuFunc(sym string) error {
	var def string
	for i, ln := range(tr.langs) {
		var s string
		ww, err := tr.writersFor(ln)
		defer tr.closeWriters(ww)
		if err != nil {
			return fmt.Errorf("failed writers for lang '%s': %v", ln.Code, err)
		}
		ctx := context.WithValue(tr.ctx, "Language", ln)
		r, err := tr.rs.GetMenu(ctx, sym)
		if err == nil {
			logg.TraceCtxf(tr.ctx, "menu found", "lang", ln, "menu", sym)
			for i, v := range(strings.Split(r, "\n")) {
				if i > 0 {
					if tr.newline {
						s += fmt.Sprintf("\t\"\\n\"\n")
					}
				}
				s += fmt.Sprintf("\t\"%s\"\n", v)
			}
			if def == "" {
				def = fmt.Sprintf("\t\"%s\"\n", sym)
				err = tr.writeTranslation(ww[2], sym, def, s)
			}
			if i == 0 {
				def = s
			}
			err = tr.writeTranslation(ww[0], sym, def, s)
			if err != nil {
				return err
			}
		} else {
			logg.DebugCtxf(tr.ctx, "no menu found", "menu", sym, "lang", ln)
		}
	}
	return nil
}

func(tr *translator) AddLang(ln lang.Language) error {
	var err error
	tr.langs = append(tr.langs, ln)
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
