package handlers

import (
	"net/http"
	"github.com/NamedKitten/kittehimageboard/template"
)

type RulesTemplateData struct {
	RulesLineCount int
	templates.TemplateTemplate
}

func RulesHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	x := RulesTemplateData{
		len(DB.Settings.Rules),
		templates.TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
		},
	}
	err := templates.RenderTemplate(w, "rules.html", x)
	if err != nil {
		panic(err)
	}
}
