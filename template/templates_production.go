// +build !development

package templates

import (
	"io"
	tmpl "text/template"
)

var templateEngine *tmpl.Template

func init() {
	templateEngine, _ = tmpl.New("").Funcs(getTemplateFuncs()).ParseGlob(
		"templates/*.html",
	)
}

func RenderTemplate(w io.Writer, name string, ctx interface{}) error {
	return templateEngine.ExecuteTemplate(w, name, ctx)
}
