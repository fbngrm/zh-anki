package template

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/fbngrm/zh-anki/pkg/hash"
)

type Processor struct {
	funcMap  template.FuncMap
	tmplPath string
}

func NewProcessor(deckname, path string, tags []string) *Processor {
	return &Processor{
		funcMap: template.FuncMap{
			"audio": func(query string) string {
				return "[sound:" + hash.Sha1(query) + ".mp3]"
			},
			"removeSpaces": func(s string) string {
				return strings.ReplaceAll(s, " ", "")
			},
			"deckName": func() string {
				return deckname
			},
			"tags": func() string {
				return strings.Join(tags, ", ")
			},
			"join": func(s []string) string {
				return strings.Join(s, " | ")
			},
			"joinWord": func(s []string) string {
				return strings.Join(s, "")
			},
		},
		tmplPath: path + "/*.tmpl",
	}
}

func (p *Processor) Fill(a any, templateName string) (string, error) {
	tmpl, err := template.New(templateName).Funcs(p.funcMap).ParseGlob(p.tmplPath)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, a)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
