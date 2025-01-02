package resource

import (
	"context"

	gotext "gopkg.in/leonelquinteros/gotext.v1"

	"git.defalsify.org/vise.git/lang"
)

const (
	TemplatePoDomain = "template"
	MenuPoDomain = "menu"
	TemplateKeyPoDomain = "default"
	MenuKeyPoDomain = "default_menu"
)

type PoResource struct {
	*MenuResource
	path string
	defaultLanguage lang.Language
	tr map[string]*gotext.Locale
}

func NewPoResource(defaultLanguage lang.Language, path string) *PoResource {
	o := &PoResource {
		MenuResource: NewMenuResource(),
		path: path,
		defaultLanguage: defaultLanguage,
		tr: make(map[string]*gotext.Locale),
	}
	return o.WithLanguage(defaultLanguage)
}

func(p *PoResource) WithLanguage(ln lang.Language) *PoResource {
	o := gotext.NewLocale(p.path, ln.Code)
	o.AddDomain(MenuPoDomain)
	o.AddDomain(TemplatePoDomain)
	if ln.Code == p.defaultLanguage.Code {
		o.AddDomain(TemplateKeyPoDomain)
		o.AddDomain(MenuKeyPoDomain)
	}
	p.tr[ln.Code] = o
	return p
}

func(p *PoResource) get(ctx context.Context, sym string, domain string) (string, error) {
	s := sym
	ln, ok := lang.LanguageFromContext(ctx)
	if !ok {
		ln = p.defaultLanguage
	}
	o, ok := p.tr[p.defaultLanguage.Code]
	if ok {
		keyDomain := TemplateKeyPoDomain
		if domain == MenuPoDomain {
			keyDomain = MenuKeyPoDomain
		}
		s = o.GetD(keyDomain, sym)
		o, ok := p.tr[ln.Code]
		if ok {
			s = o.GetD(domain, s)
		}
	}
	return s, nil
}

func(p *PoResource) GetMenu(ctx context.Context, sym string) (string, error) {
	return p.get(ctx, sym, MenuPoDomain)
}

func(p *PoResource) GetTemplate(ctx context.Context, sym string) (string, error) {
	return p.get(ctx, sym, TemplatePoDomain)
}
