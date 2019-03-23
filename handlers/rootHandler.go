package handlers

import (
	"github.com/NamedKitten/kittehimageboard/template"
	"net/http"
)

// RootTemplateData contains data to be used in the template.
// We require LoggedIn and User to display text and more buttons if
// a user is already logged in.
type RootTemplateData struct {
	templates.TemplateTemplate
}

// rootHandler is the root endpoint where a index page is served.
func RootHandler(w http.ResponseWriter, r *http.Request) {
	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", 302)
		return
	}
	user, loggedIn := DB.CheckForLoggedInUser(r)
	x := RootTemplateData{
		templates.TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
		},
	}
	err := templates.RenderTemplate(w, "index.html", x)
	if err != nil {
		panic(err)
	}
}
