package handlers

import (
	"net/http"

	templates "github.com/NamedKitten/kittehimageboard/template"
)

// RootTemplateData contains data to be used in the template.
// We require LoggedIn and User to display text and more buttons if
// a user is already logged in.
type RootTemplateData struct {
	templates.T
}

// rootHandler is the root endpoint where a index page is served.
func RootHandler(w http.ResponseWriter, r *http.Request) {
	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", 302)
		return
	}
	user, loggedIn := DB.CheckForLoggedInUser(r)
	x := RootTemplateData{
		templates.T{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
		},
	}
	err := templates.RenderTemplate(w, "index.html", x)
	if err != nil {
		panic(err)
	}
}
