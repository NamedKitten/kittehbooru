package templates

import (
	"github.com/NamedKitten/kittehimageboard/database"
	tmplHTML "html/template"
	"io/ioutil"
	"strings"
	tmpl "text/template"

	"github.com/NamedKitten/kittehimageboard/types"
)

var DB *database.DBType

// The base struct for all templating operations.
// Includes stuff for logged in user and more for headers.
type TemplateTemplate struct {
	// LoggedIn specifies whether a user is logged in or not.
	LoggedIn bool
	// LoggedInUser is the user struct of a logged in user.
	// If no user is logged in, all fields will be blank.
	LoggedInUser types.User
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
		"readPostContent": func(post types.Post) tmplHTML.HTML {
			textBytes, _ := ioutil.ReadFile("content/" + post.Filename + "." + post.FileExtension)
			text := string(textBytes)
			escaped := tmplHTML.HTMLEscapeString(text)
			replaced := strings.Replace(escaped, "\r\n", "&#10", -1)
			replaced = strings.Replace(replaced, "\n", "&#10", -1)
			html := tmplHTML.HTML(replaced)
			return html
		},
		"startsWith": func(thing, startsWith string) bool {
			return strings.HasPrefix(thing, startsWith)
		},
		"settings": func() database.Settings {
			return DB.Settings
		},
	}

}
