package main

import (
	"net/http"
)

// RootTemplateData contains data to be used in the template.
// We require LoggedIn and User to display text and more buttons if
// a user is already logged in.
type RootTemplateData struct {
	TemplateTemplate
}

// rootHandler is the root endpoint where a index page is served.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	x := RootTemplateData{
		TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
		},
	}
	err := renderTemplate(w, "index.html", x)
	if err != nil {
		panic(err)
	}
}
