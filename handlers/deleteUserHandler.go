package handlers

import (
	"github.com/NamedKitten/kittehimageboard/template"
	"github.com/rs/zerolog/log"
	"net/http"
)

type DeleteUserTemplate struct {
	templates.TemplateTemplate
}

func DeleteUserPageHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}

	templateInfo := DeleteUserTemplate{
		TemplateTemplate: templates.TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
		},
	}

	err := templates.RenderTemplate(w, "deleteUser.html", templateInfo)
	if err != nil {
		panic(err)
	}
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}

	if user.Owner {
		http.Redirect(w, r, "/", 302)
		return
	}
	log.Info().Str("username", user.Username).Msg("Account Deletion")

	DB.DeleteUser(user.Username)

	http.Redirect(w, r, "/", 302)
}
