package handlers

import (
	"net/http"
	"time"

	"github.com/NamedKitten/kittehbooru/i18n"
	templates "github.com/NamedKitten/kittehbooru/template"
	"github.com/rs/zerolog/log"
)

/*
templates.T{Translator: i18n.GetTranslator(r),}
Translator: i18n.GetTranslator(r),
"github.com/NamedKitten/kittehbooru/i18n"

*/

// LoginPageHandler takes you to the login page or the root page if you are already logged in.
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", http.StatusFound)
		return
	}
	_, loggedIn := DB.CheckForLoggedInUser(ctx, r)
	if loggedIn {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	err := templates.RenderTemplate(w, "login.html", templates.T{Translator: i18n.GetTranslator(r)})
	if err != nil {
		renderError(w, "TEMPLATE_RENDER_ERROR", err, http.StatusBadRequest)
	}
}

// loginHandler is the API endpoint that handles checking if a login
// is correct and giving the user a session token.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Parse Form")
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	_, err = DB.User(ctx, username)
	if err != nil {
		renderError(w, "INCORRECT_USER_OR_PASSWORD", err, http.StatusBadRequest)
	} else {
		if DB.CheckPassword(ctx, username, password) {
			http.SetCookie(w, &http.Cookie{
				Name:    "sessionToken",
				Value:   DB.CreateSession(ctx, username),
				Expires: time.Now().Add(2 * time.Hour),
			})
			http.Redirect(w, r, "/", http.StatusFound)
			return
		} else {
			renderError(w, "INCORRECT_USER_OR_PASSWORD", err, http.StatusBadRequest)
		}
	}
	_ = username
	_ = password
	LoginPageHandler(w, r)
}
