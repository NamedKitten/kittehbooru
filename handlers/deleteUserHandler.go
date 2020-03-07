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
	ctx := r.Context()

	user, loggedIn := DB.CheckForLoggedInUser(ctx, r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
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
	ctx := r.Context()

	user, loggedIn := DB.CheckForLoggedInUser(ctx, r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if user.Owner {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	log.Info().Str("username", user.Username).Msg("Account Deletion")

	err := DB.DeleteUser(ctx, user.Username)
	if err != nil {
		log.Error().Err(err).Msg("Delete User")
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
