package handlers

import (
	"net/http"

	"github.com/NamedKitten/kittehimageboard/i18n"
	templates "github.com/NamedKitten/kittehimageboard/template"
	"github.com/rs/zerolog/log"
)

type DeleteUserTemplate struct {
	templates.T
}

func DeleteUserPageHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}

	templateInfo := DeleteUserTemplate{
		T: templates.T{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
			Translator:   i18n.GetTranslator(r),
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

	err := DB.DeleteUser(user.Username)
	if err != nil {
		log.Error().Err(err).Msg("Delete User")
		renderError(w, "DELETE_USER_ERR", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", 302)
}
