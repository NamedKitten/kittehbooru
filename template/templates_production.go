// +build !development

package templates

import (
	"io"
	tmpl "text/template"
)

var templateEngine *tmpl.Template

func init() {
	var err error
	templateEngine, err = tmpl.New("").Funcs(getTemplateFuncs()).ParseGlob(
		"frontend/templates/*.html",
	)
	if err != nil {
		panic(err)
	}
}

func RenderTemplate(w io.Writer, name string, ctx interface{}) error {
	return templateEngine.ExecuteTemplate(w, name, ctx)
}
