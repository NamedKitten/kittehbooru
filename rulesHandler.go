package main

import (
	"net/http"
)

type RulesTemplateData struct {
	RulesLineCount int
	TemplateTemplate
}

func rulesHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	x := RulesTemplateData{
		len(DB.Settings.Rules),
		TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
		},
	}
	err := renderTemplate(w, "rules.html", x)
	if err != nil {
		panic(err)
	}
}
