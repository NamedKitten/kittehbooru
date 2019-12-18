package handlers

import (
	"github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/utils"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

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
	err := templates.RenderTemplate(w, "login.html", nil)
	if err != nil {
		panic(err)
	}
}

// loginHandler is the API endpoint that handles checking if a login
// is correct and giving the user a session token.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	if userID, ok := DB.UsernameToID[username]; ok {
		if utils.CheckPassword(DB.Passwords[userID], password) {
			log.Info().Str("username", username).Msg("Login")

			http.SetCookie(w, &http.Cookie{
				Name:    "sessionToken",
				Value:   DB.CreateSession(userID),
				Expires: time.Now().Add(2 * time.Hour),
			})
			http.Redirect(w, r, "/", 302)
			return
		} else {
			log.Info().Str("username", username).Msg("Invalid Password")
		}

	} else {
		log.Info().Str("username", username).Msg("Login Not Found")
	}
	_ = username
	_ = password
	LoginPageHandler(w, r)
}
