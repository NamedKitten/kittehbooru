package handlers

import (
	"net/http"
	"time"

	templates "github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/i18n"
	"github.com/rs/zerolog/log"
)
/* 
templates.T{Translator: i18n.GetTranslator(r),}
Translator: i18n.GetTranslator(r),
"github.com/NamedKitten/kittehimageboard/i18n"

*/

// LoginPageHandler takes you to the login page or the root page if you are already logged in.
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", 302)
		return
	}
	_, loggedIn := DB.CheckForLoggedInUser(r)
	if loggedIn {
		http.Redirect(w, r, "/", 302)
		return
	}
	err := templates.RenderTemplate(w, "login.html", templates.T{Translator: i18n.GetTranslator(r),})
	if err != nil {
		panic(err)
	}
}

// loginHandler is the API endpoint that handles checking if a login
// is correct and giving the user a session token.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Parse Form")
		renderError(w, "PARSE_FORM_ERR", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	_, exist := DB.User(username)
	if !exist {
		log.Info().Str("username", username).Msg("Login Not Found")
	} else {
		if DB.CheckPassword(username, password) {
			log.Info().Str("username", username).Msg("Login")

			http.SetCookie(w, &http.Cookie{
				Name:    "sessionToken",
				Value:   DB.Sessions.CreateSession(username),
				Expires: time.Now().Add(2 * time.Hour),
			})
			http.Redirect(w, r, "/", 302)
			return
		} else {
			log.Info().Str("username", username).Msg("Invalid Password")
		}
	}
	_ = username
	_ = password
	LoginPageHandler(w, r)
}
