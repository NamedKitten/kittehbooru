package handlers

import (
	"net/http"

	templates "github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/i18n"

)

type RulesTemplateData struct {
	RulesLineCount int
	templates.T
}

func RulesHandler(w http.ResponseWriter, r *http.Request) {
	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", 302)
		return
	}
	user, loggedIn := DB.CheckForLoggedInUser(r)
	x := RulesTemplateData{
		len(DB.Settings.Rules),
		templates.T{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
			Translator: i18n.GetTranslator(r),
		},
	}
	err := templates.RenderTemplate(w, "rules.html", x)
	if err != nil {
		panic(err)
	}
}
