package templates

import (
	tmplHTML "html/template"
	"strings"
	tmpl "text/template"

	"github.com/NamedKitten/kittehimageboard/database"
	"github.com/NamedKitten/kittehimageboard/i18n"
	"github.com/NamedKitten/kittehimageboard/types"
)

var DB *database.DB

// The base struct for all templating operations.
// Includes stuff for logged in user and more for headers.
type T struct {
	// LoggedIn specifies whether a user is logged in or not.
	LoggedIn bool
	// LoggedInUser is the user struct of a logged in user.
	// If no user is logged in, all fields will be blank.
	LoggedInUser types.User

	Translator *i18n.Translator
}

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
		"startsWith": func(thing, startsWith string) bool {
			return strings.HasPrefix(thing, startsWith)
		},
		"settings": func() database.Settings {
			return DB.Settings
		},
		"newStringInterfaceMap": func() map[string]interface{} {
			return make(map[string]interface{})
		},
		"addToStringInterfaceMap": func(d map[string]interface{}, s string, i interface{}) map[string]interface{} {
			d[s] = i
			return d
		},
		"contentURL": func() string {
			return DB.Settings.ContentURL
		},
		"thumbnailURL": func() string {
			return DB.Settings.ThumbnailURL
		},
	}

}
