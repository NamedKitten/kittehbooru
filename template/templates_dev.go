// +build development

package templates

import (
	"io"

	"github.com/NamedKitten/hot"
)

var templateEngine *hot.Template

func init() {
	config := &hot.Config{
		Watch:          true,
		BaseName:       "kittehbooru",
		Dir:            "frontend/templates",
		FilesExtension: []string{".html"},
		Funcs:          getTemplateFuncs(),
	}

	tpl, err := hot.New(config)
	if err != nil {
		panic(err)
	}
	templateEngine = tpl
}

func RenderTemplate(w io.Writer, name string, ctx interface{}) error {
	return templateEngine.Execute(w, name, ctx)
}
