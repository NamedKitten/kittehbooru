// +build development

package main

import (
	"github.com/NamedKitten/hot"
	"io"
)

var templateEngine *hot.Template

func init() {
	config := &hot.Config{
		Watch:          true,
		BaseName:       "kittehimageboard",
		Dir:            "templates",
		FilesExtension: []string{".html"},
		Funcs:          getTemplateFuncs(),
	}

	tpl, err := hot.New(config)
	if err != nil {
		panic(err)
	}
	templateEngine = tpl
}

func renderTemplate(w io.Writer, name string, ctx interface{}) error {
	return templateEngine.Execute(w, name, ctx)
}
