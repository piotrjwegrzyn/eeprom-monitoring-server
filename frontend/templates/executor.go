package templates

import (
	"bytes"
	"path"
	"strings"
	"text/template"
)

const (
	PageIndex   = "index.html"
	PageSignIn  = "signin.html"
	PageNewEdit = "new.html"
)

type Executor struct {
	templates *template.Template
}

func (e *Executor) ExecuteSignIn(data SignIn) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, PageSignIn, data); err != nil {
		return nil, err
	}

	return &buf, nil
}

func (e *Executor) ExecuteIndex(data Index) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, PageIndex, data); err != nil {
		return nil, err
	}

	return &buf, nil
}

func (e *Executor) ExecuteNewEdit(data NewEdit) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, PageNewEdit, data); err != nil {
		return nil, err
	}

	return &buf, nil
}

func NewExecutor(dir string) (*Executor, error) {
	templates, err := template.New("ems").Funcs(template.FuncMap{
		"ToUpper": strings.ToUpper,
		"ToLower": strings.ToLower,
	}).ParseFiles(
		path.Join(dir, PageSignIn),
		path.Join(dir, PageIndex),
		path.Join(dir, PageNewEdit),
	)
	if err != nil {
		return nil, err
	}

	return &Executor{
		templates: templates,
	}, nil
}
