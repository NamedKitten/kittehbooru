// +build !development

package main

import (
	"fmt"
	"io"
	tmpl "text/template"
)

var templateEngine *tmpl.Template

func init() {
	fmt.Println("prod")
	templateEngine, _ = tmpl.New("").Funcs(getTemplateFuncs()).ParseFiles(
		"templates/search.html",
		"templates/view.html",
		"templates/login.html",
		"templates/register.html",
		"templates/header.html",
		"templates/delete.html",

		"templates/index.html",
		"templates/upload.html",
		"templates/setup.html",
	)
}

func renderTemplate(w io.Writer, name string, ctx interface{}) error {
	return templateEngine.ExecuteTemplate(w, name, ctx)
}
