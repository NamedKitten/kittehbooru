package main

import (
	tmplHTML "html/template"
	"strings"
	tmpl "text/template"
)

func getTemplateFuncs() tmpl.FuncMap {
	return tmpl.FuncMap{
		"nlhtml": func(text string) tmplHTML.HTML {
			escaped := tmplHTML.HTMLEscapeString(text)
			replaced := strings.Replace(escaped, "\r\n", "&#10", -1)
			replaced = strings.Replace(replaced, "\n", "&#10", -1)
			html := tmplHTML.HTML(replaced)
			return html
		},
		"nl2br": func(text string) tmplHTML.HTML {
			escaped := tmplHTML.HTMLEscapeString(text)
			replaced := strings.Replace(escaped, "\r\n", "<br>", -1)
			replaced = strings.Replace(replaced, "\n", "<br>", -1)
			html := tmplHTML.HTML(replaced)
			return html
		},
		"settings": func() Settings {
			return DB.Settings
		},
	}

}
