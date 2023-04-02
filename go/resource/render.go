package resource

import (
	"bytes"
	"text/template"
)

func DefaultRenderTemplate(r Resource, sym string, values map[string]string) (string, error) {
	v, err := r.GetTemplate(sym)
	if err != nil {
		return "", err
	}
	tp, err := template.New("tester").Option("missingkey=error").Parse(v)
	if err != nil {
		return "", err
	}

	b := bytes.NewBuffer([]byte{})
	err = tp.Execute(b, values)
	if err != nil {
		return "", err
	}
	return b.String(), err
}
