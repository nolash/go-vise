package resource

//import (
//	"bytes"
//	"fmt"
//	"log"
//	"text/template"
//)


//// DefaultRenderTemplate is an adapter to implement the builtin golang text template renderer as resource.RenderTemplate.
//func DefaultRenderTemplate(r Resource, sym string, values map[string]string, idx uint16, sizer *Sizer) (string, error) {
//	v, err := r.GetTemplate(sym, nil)
//	if err != nil {
//		return "", err
//	}
//
//	if sizer != nil {
//		values, err = sizer.GetAt(values, idx)
//	} else if idx > 0 {
//		return "", fmt.Errorf("sizer needed for indexed render")
//	}
//	log.Printf("render for index: %v", idx)
//
//	if err != nil {
//		return "", err
//	}
//	
//	tp, err := template.New("tester").Option("missingkey=error").Parse(v)
//	if err != nil {
//		return "", err
//	}
//
//	b := bytes.NewBuffer([]byte{})
//	err = tp.Execute(b, values)
//	if err != nil {
//		return "", err
//	}
//	return b.String(), err
//}
