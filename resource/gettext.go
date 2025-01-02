package resource

import (
	"context"

	gotext "gopkg.in/leonelquinteros/gotext.v1"

	"git.defalsify.org/vise.git/lang"
)

type PoResource struct {
	path string
	domain string
	defaultLanguage lang.Language
	tr map[string]*gotext.Locale
}

func NewPoResource(defaultLanguage lang.Language, path string, domain string) *PoResource {
	o := &PoResource {
		path: path,
		domain: domain,
		defaultLanguage: defaultLanguage,
		tr: make(map[string]*gotext.Locale),
	}
	return o.WithLanguage(defaultLanguage)
}

func(p *PoResource) WithLanguage(ln lang.Language) *PoResource {
	o := gotext.NewLocale(p.path, ln.Code)
	o.AddDomain(p.domain)
	p.tr[ln.Code] = o
	return p
}

func(p *PoResource) get(ctx context.Context, sym string) (string, error) {
	s := sym
	ln, ok := lang.LanguageFromContext(ctx)
	if !ok {
		ln = p.defaultLanguage
	}
	o, ok := p.tr[ln.Code]
	if ok {
		logg.TraceCtxf(ctx, "poresource get", "sym", sym, "ln", ln, "path", p.path, "o", o)
		s = o.Get(sym)
	}
	return s, nil
}

func(p *PoResource) GetMenu(ctx context.Context, sym string) (string, error) {
	return p.get(ctx, sym)
}

func(p *PoResource) GetTemplate(ctx context.Context, sym string) (string, error) {
	return p.get(ctx, sym)
}
